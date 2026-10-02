// Package gpiocdev reaches a GPIO line by the name the device tree gives it,
// through the GPIO character device (/dev/gpiochipN, uAPI v2).
//
// The vendor 5.10 kernel exports lines through /sys/class/gpio by a global
// number, and that is the path the server has always used. The mainline kernel
// has no sysfs GPIO interface at all (CONFIG_GPIO_SYSFS is off), and its
// global numbers would differ anyway. What it has is a board device tree that
// names each control line in gpio-line-names, with the same names the pin map
// files under ironkvm-dist's devices/<device>/pins use as function names. So a
// caller that cannot find the sysfs file asks this package for the line by
// name instead.
//
// Only the few ioctls the server needs are implemented: find a line by name,
// request it as an input or an output, read it and set it. Nothing outside
// golang.org/x/sys is needed, and nothing here runs on a kernel where the
// sysfs path exists, because the caller tries that first.
package gpiocdev

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// ChipGlob is where the GPIO character devices are found. Tests point it at a
// directory that holds none.
var ChipGlob = "/dev/gpiochip*"

// Consumer is the label the kernel shows for a line this process holds.
const Consumer = "nanokvm-server"

// ErrNotFound means no GPIO chip on this kernel has a line by that name. On
// the vendor kernel, whose device tree names no lines, that is every name.
var ErrNotFound = errors.New("no gpio line by that name")

// Direction is how a line is requested.
type Direction int

const (
	Input Direction = iota
	Output
)

// Line is a line this process holds. It stays held, at its last value, until
// Close.
type Line struct {
	mu   sync.Mutex
	name string
	fd   *os.File
}

// Name is the line's device tree name.
func (l *Line) Name() string { return l.name }

// Find returns the chip device and the offset of the line called name. The
// chips are searched in name order, and the first match wins; a board device
// tree gives each name to one line.
func Find(name string) (chip string, offset uint32, err error) {
	if name == "" {
		return "", 0, ErrNotFound
	}
	chips, _ := filepath.Glob(ChipGlob)
	sort.Strings(chips)
	for _, c := range chips {
		off, ok, ferr := findOnChip(c, name)
		if ferr != nil {
			continue
		}
		if ok {
			return c, off, nil
		}
	}
	return "", 0, fmt.Errorf("%w: %q", ErrNotFound, name)
}

// Request finds the line called name and holds it in the given direction. An
// output starts at initial, so a line that drives a button is never pulsed by
// the request itself.
func Request(name string, dir Direction, initial int) (*Line, error) {
	chip, offset, err := Find(name)
	if err != nil {
		return nil, err
	}
	fd, err := requestLine(chip, offset, dir, initial)
	if err != nil {
		return nil, fmt.Errorf("request gpio line %q on %s: %w", name, chip, err)
	}
	return &Line{name: name, fd: fd}, nil
}

// Get reads the line's physical level, 0 or 1.
func (l *Line) Get() (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.fd == nil {
		return 0, os.ErrClosed
	}
	return getValue(l.fd)
}

// Set drives an output line to the physical level v, 0 or 1.
func (l *Line) Set(v int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.fd == nil {
		return os.ErrClosed
	}
	return setValue(l.fd, v)
}

// Close releases the line.
func (l *Line) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.fd == nil {
		return nil
	}
	err := l.fd.Close()
	l.fd = nil
	return err
}

// cString turns a fixed, NUL-padded kernel string into a Go string.
func cString(b []byte) string {
	if i := strings.IndexByte(string(b), 0); i >= 0 {
		return string(b[:i])
	}
	return string(b)
}
