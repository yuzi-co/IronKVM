package metrics

import (
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/sys/unix"
)

// fsStat is the part of statfs(2) the filesystem families report.
type fsStat struct {
	blockSize, blocks, free, avail, files, filesFree uint64
}

// statFS reads one mount point. A variable so tests answer for mounts that do
// not exist on the test host.
var statFS = func(path string) (fsStat, error) {
	var s unix.Statfs_t
	if err := unix.Statfs(path, &s); err != nil {
		return fsStat{}, err
	}

	return fsStat{
		blockSize: uint64(s.Bsize),
		blocks:    s.Blocks,
		free:      s.Bfree,
		avail:     s.Bavail,
		files:     s.Files,
		filesFree: s.Ffree,
	}, nil
}

// fsMountPointsExclude is node_exporter's default for
// --collector.filesystem.mount-points-exclude.
var fsMountPointsExclude = regexp.MustCompile(
	`^/(dev|proc|run/credentials/.+|sys|var/lib/docker/.+|var/lib/containers/storage/.+)($|/)`)

// fsTypesExclude is node_exporter's default for
// --collector.filesystem.fs-types-exclude, plus tmpfs and ramfs. Those two
// hold RAM, which node_memory_Shmem_bytes already reports, and the board has
// several of them.
var fsTypesExclude = regexp.MustCompile(
	`^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|` +
		`mqueue|nsfs|overlay|proc|procfs|pstore|ramfs|rpc_pipefs|securityfs|selinuxfs|squashfs|erofs|` +
		`sysfs|tmpfs|tracefs)$`)

// mountEscaper undoes the octal escapes /proc/mounts writes for the space, tab,
// newline and backslash in a path.
var mountEscaper = strings.NewReplacer(`\040`, " ", `\011`, "\t", `\012`, "\n", `\134`, `\`)

type mount struct {
	device, mountPoint, fsType string
	readOnly                   bool
}

// collectNodeFilesystem writes node_exporter's node_filesystem_* families for
// each real mount.
func collectNodeFilesystem(w *Writer) {
	type fs struct {
		mount
		stat fsStat
	}

	var filesystems []fs
	for _, m := range readMounts() {
		stat, err := statFS(m.mountPoint)
		if err != nil {
			continue
		}
		filesystems = append(filesystems, fs{m, stat})
	}

	families := []struct {
		name, help string
		value      func(fs) float64
	}{
		{"node_filesystem_size_bytes", "Filesystem size in bytes.",
			func(f fs) float64 { return float64(f.stat.blocks * f.stat.blockSize) }},
		{"node_filesystem_free_bytes", "Filesystem free space in bytes.",
			func(f fs) float64 { return float64(f.stat.free * f.stat.blockSize) }},
		{"node_filesystem_avail_bytes", "Filesystem space available to non-root users in bytes.",
			func(f fs) float64 { return float64(f.stat.avail * f.stat.blockSize) }},
		{"node_filesystem_files", "Filesystem total file nodes.",
			func(f fs) float64 { return float64(f.stat.files) }},
		{"node_filesystem_files_free", "Filesystem total free file nodes.",
			func(f fs) float64 { return float64(f.stat.filesFree) }},
		{"node_filesystem_readonly", "Filesystem read-only status.",
			func(f fs) float64 {
				if f.readOnly {
					return 1
				}
				return 0
			}},
	}
	for _, family := range families {
		for _, f := range filesystems {
			w.Gauge(family.name, family.help, family.value(f),
				L("device", f.device), L("fstype", f.fsType), L("mountpoint", f.mountPoint))
		}
	}
}

// readMounts returns the mounts the filesystem families report, from init's
// mount table as node_exporter reads it. A mount point mounted over keeps its
// place and takes the later, visible mount, so no series is written twice.
func readMounts() []mount {
	body, ok := readText(filepath.Join(procDir, "1", "mounts"))
	if !ok {
		if body, ok = readText(filepath.Join(procDir, "self", "mounts")); !ok {
			return nil
		}
	}

	var mounts []mount
	index := make(map[string]int)
	for _, line := range strings.Split(body, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		m := mount{
			device:     mountEscaper.Replace(fields[0]),
			mountPoint: mountEscaper.Replace(fields[1]),
			fsType:     fields[2],
		}
		for _, option := range strings.Split(fields[3], ",") {
			if option == "ro" {
				m.readOnly = true
			}
		}
		if fsMountPointsExclude.MatchString(m.mountPoint) || fsTypesExclude.MatchString(m.fsType) {
			continue
		}

		if i, seen := index[m.mountPoint]; seen {
			mounts[i] = m
			continue
		}
		index[m.mountPoint] = len(mounts)
		mounts = append(mounts, m)
	}

	return mounts
}
