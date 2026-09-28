package ventoy

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf16"
)

// The exFAT volume that is partition 1 of the Ventoy disk. Only its metadata
// is generated: the boot regions, the FAT, the allocation bitmap, the up-case
// table and the root directory. Each file takes a contiguous run of clusters
// after them, and the caller maps that run to the file itself. Everything
// here follows Microsoft's exFAT specification; the section numbers in the
// comments are that document's.

const (
	sectorSize  = 512
	sectorShift = 9

	// fatOffset is where the FAT starts, in sectors. The specification asks
	// for at least 24, past both boot regions; 128 keeps the FAT on a 64 KB
	// boundary.
	fatOffset = 128

	// minVolumeSectors is the smallest volume the specification allows,
	// 1 MB (section 3.1.6).
	minVolumeSectors = 1 << 20 >> sectorShift

	// maxNameLength is the longest file name, in UTF-16 code units
	// (section 7.7.3).
	maxNameLength = 255

	entrySize = 32
)

// Directory entry types (section 6.2.1).
const (
	entryBitmap = 0x81
	entryUpcase = 0x82
	entryLabel  = 0x83
	entryFile   = 0x85
	entryStream = 0xC0
	entryName   = 0xC1
)

// The FAT's special values (section 4.1).
const (
	fatMedia       = 0xFFFFFFF8
	fatEndOfChain  = 0xFFFFFFFF
	firstCluster   = 2
	maxClusterCnt  = 0xFFFFFFF5
	attrArchive    = 0x20
	flagAllocation = 0x01 // GeneralSecondaryFlags.AllocationPossible
	flagNoFatChain = 0x02 // GeneralSecondaryFlags.NoFatChain
)

// File is one file of the volume.
type File struct {
	Name    string
	Size    int64
	ModTime time.Time
}

type volumeOptions struct {
	// PartitionOffset is the volume's first sector on the disk.
	PartitionOffset uint64
	// EndAlign rounds the volume up so that the next sector on the disk is
	// a multiple of it. Ventoy starts partition 2 on a multiple of 8.
	EndAlign uint64
	Serial   uint32
	Label    string
}

// volume is a laid-out exFAT volume.
type volume struct {
	// Meta holds the volume's sectors from the first one up to the first
	// sector of the first file's clusters.
	Meta []byte
	// Sectors is the volume's length, VolumeLength.
	Sectors uint64
	// ClusterSectors is the number of sectors in a cluster.
	ClusterSectors uint64
	ClusterCount   uint32
	HeapOffset     uint64
	// Start is the first sector of each file's clusters and Allocated the
	// sectors they cover, both indexed as the files were given. An empty
	// file has neither.
	Start     []uint64
	Allocated []uint64
}

// clusterShift picks the cluster size for this much data, as Windows does
// for exFAT and Ventoy2Disk.sh does for its partition: 4 KB up to 256 MB,
// 32 KB up to 32 GB and 128 KB above.
func clusterShift(bytes uint64) uint {
	switch {
	case bytes <= 256<<20:
		return 3
	case bytes <= 32<<30:
		return 6
	default:
		return 8
	}
}

func ceilDiv(a, b uint64) uint64 { return (a + b - 1) / b }

func alignUp(a, b uint64) uint64 { return ceilDiv(a, b) * b }

// encodeName returns the name in UTF-16 or the reason exFAT cannot hold it
// (section 7.7.3).
func encodeName(name string) ([]uint16, error) {
	units := utf16.Encode([]rune(name))
	if len(units) == 0 || len(units) > maxNameLength {
		return nil, fmt.Errorf("name %q: exFAT names are 1 to %d UTF-16 code units", name, maxNameLength)
	}
	for _, u := range units {
		if u < 0x20 || strings.ContainsRune(`"*/:<>?\|`, rune(u)) {
			return nil, fmt.Errorf("name %q: exFAT does not allow %q", name, rune(u))
		}
	}
	if name == "." || name == ".." {
		return nil, fmt.Errorf("name %q is reserved", name)
	}
	return units, nil
}

// upcaseMap is the up-case table, expanded to one entry per code unit.
var upcaseMap = func() []uint16 {
	m := make([]uint16, 1<<16)
	index := 0
	skip := false
	// The same reading as the Linux driver's: a unit equal to its own index
	// maps to itself, 0xFFFF starts a run of identities whose length follows,
	// and anything else is the up-case form of the unit at the index.
	for _, u := range upcaseTable {
		if index >= len(m) {
			break
		}
		switch {
		case skip:
			for i := 0; i < int(u) && index < len(m); i++ {
				m[index] = uint16(index)
				index++
			}
			skip = false
		case int(u) == index:
			m[index] = u
			index++
		case u == 0xFFFF:
			skip = true
		default:
			m[index] = u
			index++
		}
	}
	for ; index < len(m); index++ {
		m[index] = uint16(index)
	}
	return m
}()

// upcaseKey is the name as exFAT compares it: up-cased unit by unit.
func upcaseKey(name string) string {
	units := utf16.Encode([]rune(name))
	for i, u := range units {
		units[i] = upcaseMap[u]
	}
	return string(utf16.Decode(units))
}

// nameHash is the stream extension's NameHash (section 7.6.4), over the
// up-cased name.
func nameHash(units []uint16) uint16 {
	var hash uint16
	for _, u := range units {
		u = upcaseMap[u]
		for _, b := range []byte{byte(u), byte(u >> 8)} {
			hash = (hash&1)<<15 | hash>>1
			hash += uint16(b)
		}
	}
	return hash
}

// setChecksum is the file entry's SetChecksum (section 6.3.3), over the
// whole entry set except the checksum field itself.
func setChecksum(set []byte) uint16 {
	var sum uint16
	for i, b := range set {
		if i == 2 || i == 3 {
			continue
		}
		sum = (sum&1)<<15 | sum>>1
		sum += uint16(b)
	}
	return sum
}

// tableChecksum is the up-case table's TableChecksum (section 7.2.2).
func tableChecksum(data []byte) uint32 {
	var sum uint32
	for _, b := range data {
		sum = (sum&1)<<31 | sum>>1
		sum += uint32(b)
	}
	return sum
}

// bootChecksum is the boot checksum over the first 11 sectors of a boot
// region, which skips VolumeFlags and PercentInUse (section 3.4).
func bootChecksum(region []byte) uint32 {
	var sum uint32
	for i, b := range region[:11*sectorSize] {
		if i == 106 || i == 107 || i == 112 {
			continue
		}
		sum = (sum&1)<<31 | sum>>1
		sum += uint32(b)
	}
	return sum
}

// timestamp encodes t for a directory entry (section 7.4.8), in UTC, with
// its 10 ms increment.
func timestamp(t time.Time) (uint32, uint8) {
	t = t.UTC()
	switch {
	case t.Year() < 1980:
		t = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
	case t.Year() > 2107:
		t = time.Date(2107, 12, 31, 23, 59, 58, 0, time.UTC)
	}
	ts := uint32(t.Year()-1980)<<25 | uint32(t.Month())<<21 | uint32(t.Day())<<16 |
		uint32(t.Hour())<<11 | uint32(t.Minute())<<5 | uint32(t.Second()/2)
	return ts, uint8((t.Second()%2)*100 + t.Nanosecond()/10_000_000)
}

// utcOffsetValid marks a timestamp as UTC: OffsetValid set, offset zero.
const utcOffsetValid = 0x80

// fileEntrySet builds a file's directory entries: the file entry, the stream
// extension and the name entries (sections 7.4 to 7.7).
func fileEntrySet(units []uint16, f File, first uint32) []byte {
	nameEntries := int(ceilDiv(uint64(len(units)), 15))
	set := make([]byte, (2+nameEntries)*entrySize)

	file := set[0:entrySize]
	file[0] = entryFile
	file[1] = byte(1 + nameEntries)
	binary.LittleEndian.PutUint16(file[4:], attrArchive)
	ts, ms := timestamp(f.ModTime)
	binary.LittleEndian.PutUint32(file[8:], ts)
	binary.LittleEndian.PutUint32(file[12:], ts)
	binary.LittleEndian.PutUint32(file[16:], ts)
	file[20] = ms
	file[21] = ms
	file[22] = utcOffsetValid
	file[23] = utcOffsetValid
	file[24] = utcOffsetValid

	stream := set[entrySize : 2*entrySize]
	stream[0] = entryStream
	stream[1] = flagAllocation
	if f.Size > 0 {
		// The clusters are contiguous. The FAT holds their chain as well.
		stream[1] |= flagNoFatChain
	}
	stream[3] = byte(len(units))
	binary.LittleEndian.PutUint16(stream[4:], nameHash(units))
	binary.LittleEndian.PutUint64(stream[8:], uint64(f.Size))
	binary.LittleEndian.PutUint32(stream[20:], first)
	binary.LittleEndian.PutUint64(stream[24:], uint64(f.Size))

	for i := 0; i < nameEntries; i++ {
		name := set[(2+i)*entrySize : (3+i)*entrySize]
		name[0] = entryName
		for j := 0; j < 15 && i*15+j < len(units); j++ {
			binary.LittleEndian.PutUint16(name[2+2*j:], units[i*15+j])
		}
	}

	binary.LittleEndian.PutUint16(file[2:], setChecksum(set))
	return set
}

// newVolume lays out an exFAT volume holding files, in the order given.
func newVolume(files []File, opts volumeOptions) (*volume, error) {
	if opts.EndAlign == 0 {
		opts.EndAlign = 1
	}
	label := utf16.Encode([]rune(opts.Label))
	if len(label) > 11 {
		return nil, fmt.Errorf("volume label %q is longer than 11 characters", opts.Label)
	}

	names := make([][]uint16, len(files))
	seen := map[string]bool{}
	var total uint64
	rootEntries := uint64(3) // label, bitmap, up-case
	for i, f := range files {
		units, err := encodeName(f.Name)
		if err != nil {
			return nil, err
		}
		if f.Size < 0 {
			return nil, fmt.Errorf("file %q has a negative size", f.Name)
		}
		key := upcaseKey(f.Name)
		if seen[key] {
			return nil, fmt.Errorf("two files are named %q", f.Name)
		}
		seen[key] = true
		names[i] = units
		total += uint64(f.Size)
		rootEntries += 2 + ceilDiv(uint64(len(units)), 15)
	}

	shift := clusterShift(total + 1<<20)
	spc := uint64(1) << shift
	clusterBytes := spc * sectorSize

	upcaseBytes := uint64(len(upcaseTable) * 2)
	upcaseClusters := ceilDiv(upcaseBytes, clusterBytes)
	// One entry more than the set needs, so the directory ends with an
	// end-of-directory entry.
	rootClusters := ceilDiv((rootEntries+1)*entrySize, clusterBytes)
	var fileClusters uint64
	for _, f := range files {
		fileClusters += ceilDiv(uint64(f.Size), clusterBytes)
	}

	// The bitmap's size depends on the cluster count, which counts the
	// bitmap, and a volume below the minimum gets free clusters, which
	// grow both. Iterate until the layout holds still.
	var (
		bitmapClusters uint64 = 1
		free           uint64
		used           uint64
		count          uint64
		fatLength      uint64
		heap           uint64
		sectors        uint64
	)
	for {
		used = bitmapClusters + upcaseClusters + rootClusters + fileClusters
		count = used + free
		if count > maxClusterCnt {
			return nil, errors.New("the files are too large for one exFAT volume")
		}
		fatLength = ceilDiv((count+2)*4, sectorSize)
		heap = alignUp(fatOffset+fatLength, spc)
		sectors = heap + count*spc
		sectors = alignUp(opts.PartitionOffset+sectors, opts.EndAlign) - opts.PartitionOffset

		need := ceilDiv(ceilDiv(count, 8), clusterBytes)
		if need != bitmapClusters {
			bitmapClusters = need
			continue
		}
		if sectors < minVolumeSectors {
			free += ceilDiv(minVolumeSectors-sectors, spc)
			continue
		}
		break
	}
	if (sectors-heap)/spc != count {
		return nil, fmt.Errorf("layout error: %d sectors hold %d clusters, not %d", sectors-heap, (sectors-heap)/spc, count)
	}

	v := &volume{
		Sectors:        sectors,
		ClusterSectors: spc,
		ClusterCount:   uint32(count),
		HeapOffset:     heap,
		Start:          make([]uint64, len(files)),
		Allocated:      make([]uint64, len(files)),
	}
	metaClusters := bitmapClusters + upcaseClusters + rootClusters
	meta := make([]byte, (heap+metaClusters*spc)*sectorSize)
	v.Meta = meta

	clusterOffset := func(c uint64) uint64 { return (heap + (c-firstCluster)*spc) * sectorSize }

	// The FAT: the two reserved entries, then one chain per allocation.
	fat := meta[fatOffset*sectorSize : (fatOffset+fatLength)*sectorSize]
	binary.LittleEndian.PutUint32(fat[0:], fatMedia)
	binary.LittleEndian.PutUint32(fat[4:], fatEndOfChain)
	next := uint64(firstCluster)
	allocate := func(clusters uint64) uint64 {
		if clusters == 0 {
			return 0
		}
		first := next
		for c := first; c < first+clusters-1; c++ {
			binary.LittleEndian.PutUint32(fat[c*4:], uint32(c+1))
		}
		binary.LittleEndian.PutUint32(fat[(first+clusters-1)*4:], fatEndOfChain)
		next += clusters
		return first
	}

	bitmapFirst := allocate(bitmapClusters)
	upcaseFirst := allocate(upcaseClusters)
	rootFirst := allocate(rootClusters)
	fileFirst := make([]uint64, len(files))
	for i, f := range files {
		clusters := ceilDiv(uint64(f.Size), clusterBytes)
		fileFirst[i] = allocate(clusters)
		if clusters > 0 {
			v.Start[i] = heap + (fileFirst[i]-firstCluster)*spc
			v.Allocated[i] = clusters * spc
		}
	}

	// The allocation bitmap: every cluster in use is one of the first ones.
	bitmap := meta[clusterOffset(bitmapFirst):]
	for c := uint64(0); c < used; c++ {
		bitmap[c/8] |= 1 << (c % 8)
	}

	upcase := meta[clusterOffset(upcaseFirst) : clusterOffset(upcaseFirst)+upcaseBytes]
	for i, u := range upcaseTable {
		binary.LittleEndian.PutUint16(upcase[2*i:], u)
	}

	// The root directory.
	root := meta[clusterOffset(rootFirst) : clusterOffset(rootFirst)+rootClusters*clusterBytes]
	pos := 0
	root[pos] = entryLabel
	root[pos+1] = byte(len(label))
	for i, u := range label {
		binary.LittleEndian.PutUint16(root[pos+2+2*i:], u)
	}
	pos += entrySize

	root[pos] = entryBitmap
	binary.LittleEndian.PutUint32(root[pos+20:], uint32(bitmapFirst))
	binary.LittleEndian.PutUint64(root[pos+24:], ceilDiv(count, 8))
	pos += entrySize

	root[pos] = entryUpcase
	binary.LittleEndian.PutUint32(root[pos+4:], tableChecksum(upcase))
	binary.LittleEndian.PutUint32(root[pos+20:], uint32(upcaseFirst))
	binary.LittleEndian.PutUint64(root[pos+24:], upcaseBytes)
	pos += entrySize

	for i, f := range files {
		pos += copy(root[pos:], fileEntrySet(names[i], f, uint32(fileFirst[i])))
	}

	// The main boot region, then its copy as the backup boot region.
	boot := meta[:12*sectorSize]
	copy(boot[0:3], []byte{0xEB, 0x76, 0x90})
	copy(boot[3:11], "EXFAT   ")
	binary.LittleEndian.PutUint64(boot[64:], opts.PartitionOffset)
	binary.LittleEndian.PutUint64(boot[72:], sectors)
	binary.LittleEndian.PutUint32(boot[80:], fatOffset)
	binary.LittleEndian.PutUint32(boot[84:], uint32(fatLength))
	binary.LittleEndian.PutUint32(boot[88:], uint32(heap))
	binary.LittleEndian.PutUint32(boot[92:], uint32(count))
	binary.LittleEndian.PutUint32(boot[96:], uint32(rootFirst))
	binary.LittleEndian.PutUint32(boot[100:], opts.Serial)
	binary.LittleEndian.PutUint16(boot[104:], 0x0100) // revision 1.00
	boot[108] = sectorShift
	boot[109] = byte(shift)
	boot[110] = 1    // NumberOfFats
	boot[111] = 0x80 // DriveSelect
	boot[112] = byte(used * 100 / count)
	// The boot code halts. Nothing boots from this volume's own sector.
	for i := 120; i < 510; i++ {
		boot[i] = 0xF4
	}
	boot[510], boot[511] = 0x55, 0xAA
	for s := 1; s <= 8; s++ {
		end := (s + 1) * sectorSize
		boot[end-2], boot[end-1] = 0x55, 0xAA
	}
	sum := bootChecksum(boot)
	for i := 11 * sectorSize; i < 12*sectorSize; i += 4 {
		binary.LittleEndian.PutUint32(boot[i:], sum)
	}
	copy(meta[12*sectorSize:24*sectorSize], boot)

	return v, nil
}
