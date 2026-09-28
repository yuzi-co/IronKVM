package netboot

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// safeImagePath is what an image's path under the image directory may hold to
// be offered in the menu. The path goes into an iPXE script, both as a menu
// entry and in a URL, and iPXE reads ${, spaces, # and the rest as syntax. An
// image with any other character is left out rather than escaped.
var safeImagePath = regexp.MustCompile(`^[A-Za-z0-9._-]+(/[A-Za-z0-9._-]+)*$`)

// offeredImage reports whether a path relative to the image directory is one
// the menu may name.
func offeredImage(rel string) bool {
	if !safeImagePath.MatchString(rel) || !strings.HasSuffix(strings.ToLower(rel), ".iso") {
		return false
	}
	for _, part := range strings.Split(rel, "/") {
		if part == "." || part == ".." {
			return false
		}
	}
	return true
}

// listImages walks the image directory the way the image list in the web UI
// does, and returns the ISOs the menu can offer, relative to it and sorted.
// Hidden files and directories are skipped: the download service keeps its
// partial files there as dot files.
func listImages(dir string) []string {
	var images []string

	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path != dir && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if offeredImage(rel) {
			images = append(images, rel)
		}
		return nil
	})

	sort.Strings(images)
	return images
}

// menu is the iPXE script the host gets from GET /menu.ipxe. base is the
// board's HTTP address on the link, http://<board>:8069.
//
// Each image is a sanboot of its URL, which iPXE reads by range requests and
// presents to the firmware as a disk, so an installer boots from it as from
// the virtual CD. netboot.xyz is its own binary, chained from the board; it
// needs the internet, which the link does not give, so the entry says what it
// needs. Nothing is chosen by default: the operator is at the KVM's console.
func menu(base string, images []string) string {
	var b strings.Builder

	b.WriteString("#!ipxe\n")
	b.WriteString("# The IronKVM network boot menu, written by the server for the host on the USB link.\n\n")
	fmt.Fprintf(&b, "set base %s\n\n", base)

	b.WriteString(":start\n")
	b.WriteString("menu IronKVM network boot\n")
	b.WriteString("item --gap -- Images on the KVM\n")
	if len(images) == 0 {
		b.WriteString("item --gap -- (no ISO images in the image directory)\n")
	}
	for i, image := range images {
		fmt.Fprintf(&b, "item iso%d %s\n", i, image)
	}
	b.WriteString("item --gap -- From the internet\n")
	b.WriteString("item netbootxyz netboot.xyz (needs internet on another network port of the host)\n")
	b.WriteString("item --gap -- Other\n")
	b.WriteString("item shell iPXE shell\n")
	b.WriteString("item exit Boot from the next device\n")
	b.WriteString("choose target || goto exit\n")
	b.WriteString("goto ${target}\n\n")

	for i, image := range images {
		fmt.Fprintf(&b, ":iso%d\n", i)
		fmt.Fprintf(&b, "sanboot --no-describe ${base}/iso/%s || goto failed\n\n", image)
	}

	b.WriteString(":netbootxyz\n")
	b.WriteString("echo netboot.xyz loads its menus from the internet. The USB link to the KVM\n")
	b.WriteString("echo does not reach the internet, so another network port of the host has to.\n")
	b.WriteString("iseq ${platform} efi && goto netbootxyz_efi ||\n")
	fmt.Fprintf(&b, "chain --autofree ${base}/boot/%s || goto failed\n", netbootXYZBIOS)
	b.WriteString(":netbootxyz_efi\n")
	b.WriteString("iseq ${buildarch} arm64 && goto netbootxyz_arm64 ||\n")
	fmt.Fprintf(&b, "chain --autofree ${base}/boot/%s || goto failed\n", netbootXYZEFIx64)
	b.WriteString(":netbootxyz_arm64\n")
	fmt.Fprintf(&b, "chain --autofree ${base}/boot/%s || goto failed\n\n", netbootXYZEFIarm64)

	b.WriteString(":shell\n")
	b.WriteString("shell\n")
	b.WriteString("goto start\n\n")

	b.WriteString(":exit\n")
	b.WriteString("exit\n\n")

	b.WriteString(":failed\n")
	b.WriteString("echo Booting failed.\n")
	b.WriteString("prompt Press any key to go back to the menu\n")
	b.WriteString("goto start\n")

	return b.String()
}
