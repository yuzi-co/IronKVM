// Package ventoy serves a Ventoy disk built from the images on /data without
// copying them.
//
// The disk is a device-mapper device: a generated head file holds the MBR,
// Ventoy's boot code and the metadata of an exFAT partition that lists the
// selected images, each image is mapped in place through a read-only loop
// device, and Ventoy's VTOYEFI partition follows. The mass storage function
// serves the device read-only on the disk drive. Once the table is loaded
// the kernel does all the reading, so a server restart does not disturb a
// host that is booting from it. See
// docs/superpowers/specs/2026-09-28-ventoy-design.md.
package ventoy

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/service/storage"
	"NanoKVM-Server/utils"
)

const (
	// DeviceName is the device-mapper name, and DevicePath its node.
	DeviceName = "ventoy"
	DevicePath = "/dev/mapper/ventoy"
	// deviceUUID marks the device as this server's.
	deviceUUID = "IRONKVM-VENTOY"

	headName  = "head.bin"
	stateName = "state.json"
)

var (
	errNoKernel = errors.New("this kernel has no device-mapper, which the Ventoy disk needs: install an image with CONFIG_BLK_DEV_DM")
	errNotReady = errors.New("install Ventoy first")
	errNoImages = errors.New("put at least one image on the Ventoy disk first")
	errInDrive  = errors.New("eject the Ventoy disk first: its images change only while it is in no drive")
	errNoData   = errors.New("the Ventoy files live on /data, and need an IronKVM image with /data mounted")
	errNotImage = errors.New("not an image in the image directory")
)

// The board's side, as variables so the tests can replace it.
var (
	kernelSupport = kernelHasDM
	assembleDisk  = assemble
	removeDisk    = disassemble
	diskExists    = deviceExists
	diskSize      = deviceSize

	insertDevice   = storage.InsertDevice
	ejectDrive     = storage.EjectDrive
	deviceDrive    = storage.DeviceDrive
	registerDevice = storage.RegisterDevice
)

// state is what state.json keeps: the selection and the disk's identity.
type state struct {
	Images    []string `json:"images"`
	UUID      string   `json:"uuid"`
	Signature string   `json:"signature"`
	Serial    uint32   `json:"serial"`
}

// Deps is what the service needs from the rest of the server.
type Deps struct {
	// ImageDir is the image directory. Only images in it can be on the disk.
	ImageDir string
	// OnData says whether /data is there to keep the release on.
	OnData func() bool
}

// Status is what the web UI shows.
type Status struct {
	Kernel    bool     `json:"kernel"`
	OnData    bool     `json:"onData"`
	Installed bool     `json:"installed"`
	Version   string   `json:"version"`
	Images    []string `json:"images"`
	Missing   []string `json:"missing"`
	InDrive   bool     `json:"inDrive"`
	Device    string   `json:"device"`
	Size      int64    `json:"size"`
}

// Service owns the Ventoy disk.
type Service struct {
	deps Deps

	// mu keeps the operations apart. An install takes minutes, and a build
	// replaces the head the device reads.
	mu sync.Mutex

	// servedMu guards served, the images of the device that exists. The
	// storage delete guard reads it with its own lock held, so nothing
	// holds servedMu while calling storage.
	servedMu sync.Mutex
	served   []string
}

func New(deps Deps) *Service {
	if deps.OnData == nil {
		deps.OnData = func() bool { return true }
	}
	return &Service{deps: deps}
}

// Start registers the disk with storage, and adopts a device left by an
// earlier run of the server: one in a drive keeps serving, and one in no
// drive is removed.
func (s *Service) Start() {
	registerDevice(storage.Device{
		Path:     DevicePath,
		Name:     "Ventoy",
		Images:   s.servedImages,
		Released: s.released,
	})

	s.mu.Lock()
	defer s.mu.Unlock()

	if !kernelSupport() || !diskExists(DeviceName) {
		return
	}
	drive, err := deviceDrive(DevicePath)
	if err != nil {
		log.Errorf("ventoy: read the drives: %s", err)
		return
	}
	if drive == "" {
		if err := removeDisk(DeviceName, DevicePath); err != nil {
			log.Errorf("ventoy: remove the idle disk: %s", err)
		}
		return
	}
	st, err := s.loadState()
	if err != nil {
		log.Errorf("ventoy: %s", err)
	}
	present, _ := s.split(st.Images)
	s.setServed(present)
	log.Infof("ventoy: the disk is in the %s drive, with %d images", drive, len(present))
}

func (s *Service) servedImages() []string {
	s.servedMu.Lock()
	defer s.servedMu.Unlock()
	return append([]string(nil), s.served...)
}

func (s *Service) setServed(images []string) {
	s.servedMu.Lock()
	defer s.servedMu.Unlock()
	s.served = images
}

// released runs when a drive lets go of the disk. The device is removed,
// so a later insert builds it again from the images as they are then.
func (s *Service) released() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.removeIdle(); err != nil {
		log.Errorf("ventoy: %s", err)
	}
}

// removeIdle removes the device when no drive holds it. The caller holds mu.
func (s *Service) removeIdle() error {
	if !kernelSupport() || !diskExists(DeviceName) {
		s.setServed(nil)
		return nil
	}
	drive, err := deviceDrive(DevicePath)
	if err != nil {
		return err
	}
	if drive != "" {
		return nil
	}
	if err := removeDisk(DeviceName, DevicePath); err != nil {
		return err
	}
	s.setServed(nil)
	return nil
}

// inDrive reports whether a drive holds the disk.
func inDrive() (bool, error) {
	drive, err := deviceDrive(DevicePath)
	return drive != "", err
}

// loadState reads state.json, and draws the disk's identity the first time.
func (s *Service) loadState() (state, error) {
	var st state
	data, err := os.ReadFile(inDir(stateName))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return st, err
	}
	if err == nil {
		if err := json.Unmarshal(data, &st); err != nil {
			return state{}, fmt.Errorf("read %s: %w", stateName, err)
		}
	}
	if len(st.UUID) != 32 || len(st.Signature) != 8 || st.Serial == 0 {
		var id [24]byte
		if _, err := rand.Read(id[:]); err != nil {
			return st, err
		}
		st.UUID = fmt.Sprintf("%x", id[:16])
		st.Signature = fmt.Sprintf("%x", id[16:20])
		st.Serial = uint32(id[20]) | uint32(id[21])<<8 | uint32(id[22])<<16 | uint32(id[23])<<24 | 1
	}
	return st, nil
}

func (s *Service) saveState(st state) error {
	if err := os.MkdirAll(Dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(inDir(stateName), data)
}

func writeAtomic(name string, data []byte) error {
	tmp := name + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, name)
}

func (st state) identity() (Identity, error) {
	var id Identity
	uuid, err := hex.DecodeString(st.UUID)
	if err != nil || len(uuid) != 16 {
		return id, fmt.Errorf("bad disk uuid %q in %s", st.UUID, stateName)
	}
	sig, err := hex.DecodeString(st.Signature)
	if err != nil || len(sig) != 4 {
		return id, fmt.Errorf("bad disk signature %q in %s", st.Signature, stateName)
	}
	copy(id.UUID[:], uuid)
	copy(id.Signature[:], sig)
	id.Serial = st.Serial
	return id, nil
}

// checkImage returns the clean path of an image in the image directory, or
// errNotImage.
func (s *Service) checkImage(file string) (string, error) {
	file = filepath.Clean(file)
	lower := strings.ToLower(file)
	if !utils.IsPathInside(s.deps.ImageDir, file) ||
		!(strings.HasSuffix(lower, ".iso") || strings.HasSuffix(lower, ".img")) {
		return "", fmt.Errorf("%w: %s", errNotImage, file)
	}
	return file, nil
}

// split separates the images that are regular files in the image directory,
// with every link followed, from those that are not, or are gone.
func (s *Service) split(images []string) (present, missing []string) {
	root, err := filepath.EvalSymlinks(s.deps.ImageDir)
	if err != nil {
		root = filepath.Clean(s.deps.ImageDir)
	}
	for _, image := range images {
		resolved, err := filepath.EvalSymlinks(image)
		if err == nil && utils.IsPathInside(root, resolved) {
			if fi, err := os.Stat(resolved); err == nil && fi.Mode().IsRegular() {
				present = append(present, image)
				continue
			}
		}
		missing = append(missing, image)
	}
	return present, missing
}

// Status reports the disk's state.
func (s *Service) status() Status {
	st := Status{
		Kernel:    kernelSupport(),
		OnData:    s.deps.OnData(),
		Installed: installed(),
		Version:   Version,
		Images:    []string{},
		Missing:   []string{},
		Device:    DevicePath,
	}
	if saved, err := s.loadState(); err == nil {
		st.Images = append(st.Images, saved.Images...)
		_, missing := s.split(saved.Images)
		st.Missing = append(st.Missing, missing...)
	}
	if st.Kernel && diskExists(DeviceName) {
		if in, err := inDrive(); err == nil && in {
			st.InDrive = true
			st.Size = diskSize(DeviceName)
		}
	}
	return st
}

// setImages replaces the selection. It is refused while a drive holds the
// disk: the host would see its files change under it.
func (s *Service) setImages(images []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if in, err := inDrive(); err != nil {
		return err
	} else if in {
		return errInDrive
	}

	seen := map[string]bool{}
	clean := []string{}
	for _, image := range images {
		file, err := s.checkImage(image)
		if err != nil {
			return err
		}
		if !seen[file] {
			seen[file] = true
			clean = append(clean, file)
		}
	}
	sort.Strings(clean)

	st, err := s.loadState()
	if err != nil {
		return err
	}
	st.Images = clean
	if err := s.saveState(st); err != nil {
		return err
	}
	// A device built from the old set must not be inserted later.
	return s.removeIdle()
}

// insert builds the disk from the selection and puts it into the disk drive.
// It always builds afresh, so an image replaced since the last build is
// never served from a stale head.
func (s *Service) insert() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !kernelSupport() {
		return errNoKernel
	}
	if !installed() {
		return errNotReady
	}
	if in, err := inDrive(); err != nil {
		return err
	} else if in {
		return nil
	}

	st, err := s.loadState()
	if err != nil {
		return err
	}
	images, _ := s.split(st.Images)
	if len(images) == 0 {
		return errNoImages
	}
	// The identity drawn now is the one every later build uses.
	if err := s.saveState(st); err != nil {
		return err
	}

	if err := s.removeIdle(); err != nil {
		return err
	}
	disk, err := buildFromImages(images, st)
	if err != nil {
		return err
	}
	if err := writeAtomic(inDir(headName), disk.Head); err != nil {
		return err
	}

	src := sources{Head: inDir(headName), EFI: inDir(efiName), Files: images}
	if err := assembleDisk(DeviceName, deviceUUID, DevicePath, disk, src); err != nil {
		return err
	}
	s.setServed(images)

	if err := insertDevice(DevicePath); err != nil {
		if rmErr := removeDisk(DeviceName, DevicePath); rmErr != nil {
			log.Errorf("ventoy: remove the disk after a failed insert: %s", rmErr)
		}
		s.setServed(nil)
		return err
	}
	log.Infof("ventoy: inserted, %d images, %d sectors", len(images), disk.Sectors)
	return nil
}

// eject takes the disk out of the disk drive and removes the device.
func (s *Service) eject() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	drive, err := deviceDrive(DevicePath)
	if err != nil {
		return err
	}
	if drive != "" {
		if err := ejectDrive(drive); err != nil {
			return err
		}
	}
	return s.removeIdle()
}

// install fetches the release. It needs /data, and the disk out of the
// drives, as the files it replaces are the ones the device reads.
func (s *Service) install() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.deps.OnData() {
		return errNoData
	}
	if in, err := inDrive(); err != nil {
		return err
	} else if in {
		return errInDrive
	}
	if err := s.removeIdle(); err != nil {
		return err
	}
	return install()
}

func (s *Service) uninstall() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if in, err := inDrive(); err != nil {
		return err
	} else if in {
		return errInDrive
	}
	if err := s.removeIdle(); err != nil {
		return err
	}
	return uninstall()
}

// buildFromImages lays out the disk for the images, named by their base
// names, with the release's boot files and the saved identity.
func buildFromImages(images []string, st state) (*Disk, error) {
	id, err := st.identity()
	if err != nil {
		return nil, err
	}
	boot, err := os.ReadFile(inDir(bootName))
	if err != nil {
		return nil, err
	}
	core, err := os.ReadFile(inDir(coreName))
	if err != nil {
		return nil, err
	}

	names := uniqueNames(images)
	files := make([]File, len(images))
	for i, image := range images {
		fi, err := os.Stat(image)
		if err != nil {
			return nil, err
		}
		files[i] = File{Name: names[i], Size: fi.Size(), ModTime: fi.ModTime()}
	}
	tail := func(i int, buf []byte) error {
		f, err := os.Open(images[i])
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = f.ReadAt(buf, files[i].Size-int64(len(buf)))
		return err
	}
	return BuildDisk(files, tail, BootFiles{BootImg: boot, CoreImg: core}, id)
}

// uniqueNames gives each image its base name, and a second image with a
// name exFAT would take for the same one " (2)", " (3)" before its
// extension.
func uniqueNames(images []string) []string {
	names := make([]string, len(images))
	taken := map[string]bool{}
	for i, image := range images {
		base := filepath.Base(image)
		name := base
		ext := filepath.Ext(base)
		stem := strings.TrimSuffix(base, ext)
		for n := 2; taken[upcaseKey(name)]; n++ {
			name = fmt.Sprintf("%s (%d)%s", stem, n, ext)
		}
		taken[upcaseKey(name)] = true
		names[i] = name
	}
	return names
}
