package hid

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The report lengths in this package and the ones the init scripts write into
// configfs are the same numbers in two languages. f_hid copies report_length
// into the gadget when the function is created, and the host then reads exactly
// that many bytes. A server that sends a different number is not a mismatch the
// host reports: a short report is ignored, a long one is truncated, and the
// symptom is a mouse that half works.
//
// So the build holds them together, the same way endpoints_shell_test.go holds
// the endpoint budget to its table. Change one and this fails until the other
// follows.
//
// Both scripts are checked. S03usbdev composes the whole gadget and S03usbhid
// is the HID-only mode, and they carry their own copies of the descriptors:
// a change applied to one of them leaves the other mode broken.
var initScripts = []string{
	"../../../kvmapp/system/init.d/S03usbdev",
	"../../../kvmapp/system/init.d/S03usbhid",
}

// hidFunctions maps the configfs instance to the constant that must match it.
var hidFunctions = map[string]int{
	"hid.GS0": KeyboardReportLen,
	"hid.GS1": RelativeMouseReportLen,
	"hid.GS2": AbsoluteMouseReportLen,
}

func readScript(t *testing.T, path string) string {
	t.Helper()

	body, err := os.ReadFile(filepath.FromSlash(path))
	if err != nil {
		t.Fatalf("cannot read %s: %s", path, err)
	}
	return string(body)
}

// shellReportLengths pulls the values out of lines shaped like
//
//	echo 8 > functions/hid.GS0/report_length
func shellReportLengths(t *testing.T, script string) map[string]int {
	t.Helper()

	pattern := regexp.MustCompile(`echo\s+(\d+)\s*>\s*functions/(hid\.GS\d)/report_length`)
	found := map[string]int{}

	for _, match := range pattern.FindAllStringSubmatch(script, -1) {
		length, err := strconv.Atoi(match[1])
		if err != nil {
			t.Fatalf("unreadable report_length %q: %s", match[1], err)
		}
		if previous, seen := found[match[2]]; seen && previous != length {
			t.Fatalf("%s is given two different report lengths, %d and %d", match[2], previous, length)
		}
		found[match[2]] = length
	}

	return found
}

func TestShellAndGoAgreeOnEveryReportLength(t *testing.T) {
	for _, path := range initScripts {
		script := readScript(t, path)
		found := shellReportLengths(t, script)

		if len(found) != len(hidFunctions) {
			t.Fatalf("%s writes %d report lengths, want %d: %v", path, len(found), len(hidFunctions), found)
		}

		for function, want := range hidFunctions {
			got, ok := found[function]
			if !ok {
				t.Errorf("%s sets no report_length for %s", path, function)
				continue
			}
			if got != want {
				t.Errorf("%s gives %s a report_length of %d, but this package sends %d",
					path, function, got, want)
			}
		}
	}
}

// The descriptor has to describe as many bytes as report_length claims, or the
// host is told one thing and given another. This counts what the descriptor
// asks for rather than trusting the two numbers to have been changed together,
// which is the mistake that is easy to make and invisible until a device runs.
func TestTheRelativeDescriptorDescribesTheWholeReport(t *testing.T) {
	for _, path := range initScripts {
		script := readScript(t, path)

		descriptor := relativeDescriptor(t, path, script)
		bits := descriptorInputBits(t, descriptor)

		if bits%8 != 0 {
			t.Errorf("%s: the relative descriptor asks for %d bits, which is not whole bytes", path, bits)
			continue
		}
		if got := bits / 8; got != RelativeMouseReportLen {
			t.Errorf("%s: the relative descriptor describes %d bytes, but report_length and this package say %d",
				path, got, RelativeMouseReportLen)
		}
	}
}

// The horizontal wheel is AC Pan from the Consumer page, not a Generic Desktop
// usage. A descriptor that reached the right length with the wrong usage would
// pass the count above and produce a pointer that scrolls sideways when it
// should not, so the item itself is checked.
func TestTheRelativeDescriptorUsesACPan(t *testing.T) {
	for _, path := range initScripts {
		descriptor := relativeDescriptor(t, path, readScript(t, path))

		// 05 0c = Usage Page (Consumer), 0a 38 02 = Usage (AC Pan).
		if !containsBytes(descriptor, []byte{0x05, 0x0c, 0x0a, 0x38, 0x02}) {
			t.Errorf("%s: the relative descriptor has no Consumer/AC Pan item", path)
		}
	}
}

// The first three bytes must stay buttons, X and Y. This interface claims the
// USB boot mouse protocol, and a host reading it in boot mode takes those three
// and nothing else. Appending the horizontal wheel keeps that true; inserting
// it would silently break every BIOS.
func TestTheRelativeDescriptorKeepsTheBootLayout(t *testing.T) {
	for _, path := range initScripts {
		descriptor := relativeDescriptor(t, path, readScript(t, path))

		// 09 30 09 31 = Usage (X), Usage (Y), which must appear before the
		// Consumer page is selected.
		xy := indexOfBytes(descriptor, []byte{0x09, 0x30, 0x09, 0x31})
		pan := indexOfBytes(descriptor, []byte{0x05, 0x0c})

		if xy < 0 {
			t.Errorf("%s: the relative descriptor has no X and Y usages", path)
			continue
		}
		if pan >= 0 && pan < xy {
			t.Errorf("%s: the horizontal wheel is declared before X and Y, which moves the boot bytes", path)
		}
	}
}

// relativeDescriptor decodes the escaped bytes the script echoes into
// hid.GS1/report_desc.
func relativeDescriptor(t *testing.T, path string, script string) []byte {
	t.Helper()

	pattern := regexp.MustCompile(`echo -ne (\S+) > functions/hid\.GS1/report_desc`)
	match := pattern.FindStringSubmatch(script)
	if match == nil {
		t.Fatalf("%s: no report_desc line for hid.GS1", path)
	}

	// The script is read by sh, which turns every '\\' into one '\'. What
	// echo -ne then sees is a run of \xNN escapes.
	escaped := strings.ReplaceAll(match[1], `\\`, `\`)

	descriptor, err := decodeHexEscapes(escaped)
	if err != nil {
		t.Fatalf("%s: %s", path, err)
	}
	return descriptor
}

func decodeHexEscapes(escaped string) ([]byte, error) {
	if !strings.HasPrefix(escaped, `\x`) {
		return nil, fmt.Errorf("the descriptor does not start with an escape: %q", escaped)
	}

	var out []byte
	for _, field := range strings.Split(escaped, `\x`) {
		if field == "" {
			continue
		}
		// \x5 and \x05 are the same byte, and the script uses both.
		value, err := strconv.ParseUint(field, 16, 8)
		if err != nil {
			return nil, fmt.Errorf("unreadable escape %q: %w", field, err)
		}
		out = append(out, byte(value))
	}
	return out, nil
}

// descriptorInputBits walks the short-item encoding and totals Report Size x
// Report Count over every Input item, which is the size of one report.
//
// Only what this descriptor uses is handled: one-byte short items, and the
// three-byte Usage that AC Pan needs. A longer item would be a descriptor this
// was not written for, so it fails rather than guessing.
func descriptorInputBits(t *testing.T, descriptor []byte) int {
	t.Helper()

	const (
		itemReportSize  = 0x74 // Global, tag 0x7
		itemReportCount = 0x94 // Global, tag 0x9
		itemInput       = 0x80 // Main, tag 0x8
	)

	var size, count, bits int

	for i := 0; i < len(descriptor); {
		prefix := descriptor[i]
		length := int(prefix & 0x03)
		if length == 3 {
			length = 4
		}
		if i+1+length > len(descriptor) {
			t.Fatalf("the descriptor ends inside an item at offset %d", i)
		}

		var value int
		for b := 0; b < length; b++ {
			value |= int(descriptor[i+1+b]) << (8 * b)
		}

		switch prefix &^ 0x03 {
		case itemReportSize:
			size = value
		case itemReportCount:
			count = value
		case itemInput:
			bits += size * count
		}

		i += 1 + length
	}

	return bits
}

func indexOfBytes(haystack []byte, needle []byte) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func containsBytes(haystack []byte, needle []byte) bool {
	return indexOfBytes(haystack, needle) >= 0
}
