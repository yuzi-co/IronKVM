package netboot

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestTheMenuOffersEveryImageAndNetbootXYZ(t *testing.T) {
	script := menu("http://172.31.255.1:8069", []string{"debian-13.iso", "tools/memtest.iso"})

	if !strings.HasPrefix(script, "#!ipxe\n") {
		t.Fatalf("not an iPXE script:\n%s", script)
	}
	for _, want := range []string{
		"set base http://172.31.255.1:8069\n",
		"item iso0 debian-13.iso\n",
		"item iso1 tools/memtest.iso\n",
		":iso0\nsanboot --no-describe ${base}/iso/debian-13.iso || goto failed\n",
		":iso1\nsanboot --no-describe ${base}/iso/tools/memtest.iso || goto failed\n",
		"item netbootxyz netboot.xyz (needs internet on another network port of the host)\n",
		"chain --autofree ${base}/boot/netboot.xyz.kpxe || goto failed\n",
		"chain --autofree ${base}/boot/netboot.xyz.efi || goto failed\n",
		"chain --autofree ${base}/boot/netboot.xyz-arm64.efi || goto failed\n",
		"item exit Boot from the next device\n",
		"choose target || goto exit\n",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("no %q in:\n%s", want, script)
		}
	}
	// The operator chooses; nothing boots by itself.
	if strings.Contains(script, "--default") || strings.Contains(script, "--timeout") {
		t.Errorf("the menu chooses on its own:\n%s", script)
	}
}

// Every goto names a label the script has.
func TestTheMenuJumpsOnlyToItsOwnLabels(t *testing.T) {
	script := menu("http://172.31.255.1:8069", []string{"a.iso", "b.iso"})

	labels := map[string]bool{}
	for _, line := range strings.Split(script, "\n") {
		if strings.HasPrefix(line, ":") {
			labels[line[1:]] = true
		}
	}
	for _, line := range strings.Split(script, "\n") {
		for _, word := range []string{"goto "} {
			i := strings.Index(line, word)
			if i < 0 {
				continue
			}
			target := strings.Fields(line[i+len(word):])[0]
			if target == "${target}" {
				continue
			}
			if !labels[target] {
				t.Errorf("%q jumps to %s, which is not a label", line, target)
			}
		}
	}
	for _, item := range []string{"iso0", "iso1", "netbootxyz", "shell", "exit"} {
		if !labels[item] {
			t.Errorf("the item %s has no label", item)
		}
	}
}

func TestTheMenuSaysSoWhenThereAreNoImages(t *testing.T) {
	script := menu("http://172.31.255.1:8069", nil)
	if !strings.Contains(script, "(no ISO images in the image directory)") {
		t.Errorf("no word about the empty image directory:\n%s", script)
	}
	if strings.Contains(script, "sanboot") {
		t.Errorf("a sanboot with no image:\n%s", script)
	}
}

// A path goes into the script as it is, so one that iPXE would read as syntax
// is not offered at all.
func TestOnlyPlainImagePathsAreOffered(t *testing.T) {
	for path, want := range map[string]bool{
		"debian.iso":             true,
		"a/b/ubuntu-24.04.iso":   true,
		"UPPER.ISO":              true,
		"with space.iso":         false,
		"${evil}.iso":            false,
		"a#b.iso":                false,
		"a&&b.iso":               false,
		"../escape.iso":          false,
		"a/./b.iso":              false,
		"disk.img":               false,
		"notes.txt":              false,
		"/absolute.iso":          false,
		"trailing/":              false,
		"a\nitem x y.iso":        false,
		"ünicode.iso":            false,
		"semi;colon.iso":         false,
		"quote\".iso":            false,
		"back\\slash.iso":        false,
		"x.iso/../../etc/passwd": false,
	} {
		if got := offeredImage(path); got != want {
			t.Errorf("offeredImage(%q) = %t, want %t", path, got, want)
		}
	}
}

func TestTheImageListWalksTheDirectoryAndSkipsHiddenFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{
		"b.iso", "a.ISO", "sub/c.iso", "disk.img", "with space.iso",
		".nanokvm-download-123", ".hidden/d.iso", "sub/.e.iso",
	} {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got := listImages(dir)
	want := []string{"a.ISO", "b.iso", "sub/c.iso"}
	if !slices.Equal(got, want) {
		t.Fatalf("images %q, want %q", got, want)
	}
}
