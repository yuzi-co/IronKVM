//go:build !linux

package gpiocdev

import (
	"errors"
	"os"
)

var errUnsupported = errors.New("gpio character devices exist only on linux")

func findOnChip(string, string) (uint32, bool, error) { return 0, false, errUnsupported }

func requestLine(string, uint32, Direction, int) (*os.File, error) {
	return nil, errUnsupported
}

func getValue(*os.File) (int, error) { return 0, errUnsupported }

func setValue(*os.File, int) error { return errUnsupported }
