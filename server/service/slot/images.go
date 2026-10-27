package slot

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// maxImages bounds the list. A board keeps one or two slot images on /data;
// anything past this is a directory of something else.
const maxImages = 64

// Image is a slot image file on the board.
type Image struct {
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	// Signed is true when a signature lies beside the image or its
	// manifest. Whether the signature verifies is the tool's to say.
	Signed bool `json:"signed"`
	// Device and Version are what the manifest beside the image states.
	Device  string `json:"device,omitempty"`
	Version string `json:"version,omitempty"`
}

// listImages finds the slot images on ImageRoot and one directory below it.
//
// A virtual media disk image is also a .img on /data, so a plain .img counts
// only with a manifest or a signature beside it: the build writes both beside
// every slot image, and a disk image has neither. A .img.zst is always a slot
// image; nothing else on the board is kept that way.
func listImages() []Image {
	root, err := filepath.EvalSymlinks(ImageRoot)
	if err != nil {
		root = ImageRoot
	}
	dirs := []string{root}
	if entries, err := os.ReadDir(root); err == nil {
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") && e.Name() != "lost+found" {
				dirs = append(dirs, filepath.Join(root, e.Name()))
			}
		}
	}

	images := []Image{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if len(images) >= maxImages {
				break
			}
			if !e.Type().IsRegular() || !hasImageSuffix(e.Name()) {
				continue
			}
			path := filepath.Join(dir, e.Name())
			manifest := exists(path + ".manifest")
			signed := exists(path+".minisig") || exists(path+".manifest.minisig")
			if !strings.HasSuffix(path, ".zst") && !manifest && !signed {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			img := Image{Path: path, Size: info.Size(), Modified: info.ModTime().UTC(), Signed: signed}
			if manifest {
				img.Device, img.Version = readManifest(path + ".manifest")
			}
			images = append(images, img)
		}
	}
	sort.Slice(images, func(i, j int) bool { return images[i].Modified.After(images[j].Modified) })
	return images
}

func exists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Mode().IsRegular()
}

// readManifest reads the device and version lines of an image's manifest.
// Only a file that starts with the manifest's first line is read.
func readManifest(path string) (device, version string) {
	f, err := os.Open(path)
	if err != nil {
		return "", ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	if !sc.Scan() || strings.TrimSpace(sc.Text()) != "ironkvm-slot-manifest 1" {
		return "", ""
	}
	for lines := 0; sc.Scan() && lines < 64; lines++ {
		k, v, ok := strings.Cut(sc.Text(), " ")
		if !ok {
			continue
		}
		switch k {
		case "device":
			if device == "" {
				device = strings.TrimSpace(v)
			}
		case "version":
			if version == "" {
				version = strings.TrimSpace(v)
			}
		}
	}
	return device, version
}
