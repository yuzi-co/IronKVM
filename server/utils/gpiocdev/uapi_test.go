//go:build linux

package gpiocdev

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"unsafe"
)

// The kernel reads these structures by size and offset. A field out of place
// requests the wrong line, so the sizes are held against
// include/uapi/linux/gpio.h, and so are the ioctl numbers the header derives.
func TestStructureSizesMatchTheKernelHeader(t *testing.T) {
	cases := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"gpiochip_info", unsafe.Sizeof(chipInfo{}), 68},
		{"gpio_v2_line_attribute", unsafe.Sizeof(lineAttribute{}), 16},
		{"gpio_v2_line_info", unsafe.Sizeof(lineInfo{}), 256},
		{"gpio_v2_line_config_attribute", unsafe.Sizeof(lineConfigAttribute{}), 24},
		{"gpio_v2_line_config", unsafe.Sizeof(lineConfig{}), 272},
		{"gpio_v2_line_request", unsafe.Sizeof(lineRequest{}), 592},
		{"gpio_v2_line_values", unsafe.Sizeof(lineValues{}), 16},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("sizeof(struct %s) = %d, want %d", c.name, c.got, c.want)
		}
	}
	if off := unsafe.Offsetof(lineRequest{}.Fd); off != 588 {
		t.Errorf("gpio_v2_line_request.fd at %d, want 588", off)
	}
	if off := unsafe.Offsetof(lineRequest{}.Config); off != 288 {
		t.Errorf("gpio_v2_line_request.config at %d, want 288", off)
	}
}

func TestIoctlNumbersMatchTheKernelHeader(t *testing.T) {
	cases := []struct {
		name      string
		got, want uintptr
	}{
		{"GPIO_GET_CHIPINFO_IOCTL", ioctlChipInfo, 0x8044B401},
		{"GPIO_V2_GET_LINEINFO_IOCTL", ioctlLineInfo, 0xC100B405},
		{"GPIO_V2_GET_LINE_IOCTL", ioctlLine, 0xC250B407},
		{"GPIO_V2_LINE_GET_VALUES_IOCTL", ioctlGetValues, 0xC010B40E},
		{"GPIO_V2_LINE_SET_VALUES_IOCTL", ioctlSetValues, 0xC010B40F},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %#x, want %#x", c.name, c.got, c.want)
		}
	}
}

// A kernel with no chips, or a name no chip carries, is ErrNotFound, which is
// what the caller reads as "use the sysfs path".
func TestFindWithoutChipsIsNotFound(t *testing.T) {
	old := ChipGlob
	ChipGlob = filepath.Join(t.TempDir(), "gpiochip*")
	t.Cleanup(func() { ChipGlob = old })

	if _, _, err := Find("power"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Find on no chips = %v, want ErrNotFound", err)
	}
	if _, err := Request("power", Output, 0); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Request on no chips = %v, want ErrNotFound", err)
	}
	if _, _, err := Find(""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Find of an empty name = %v, want ErrNotFound", err)
	}
}

// A file that is not a GPIO chip fails the chip-info ioctl and is skipped.
func TestFindSkipsAFileThatIsNotAChip(t *testing.T) {
	dir := t.TempDir()
	old := ChipGlob
	ChipGlob = filepath.Join(dir, "gpiochip*")
	t.Cleanup(func() { ChipGlob = old })
	if err := os.WriteFile(filepath.Join(dir, "gpiochip0"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Find("power"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Find over a plain file = %v, want ErrNotFound", err)
	}
}

func TestCString(t *testing.T) {
	var b [32]byte
	copy(b[:], "reset")
	if got := cString(b[:]); got != "reset" {
		t.Fatalf("cString = %q", got)
	}
}
