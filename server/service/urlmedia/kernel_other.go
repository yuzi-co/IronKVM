//go:build !linux

package urlmedia

import "net"

func kernelNBDAvailable() error { return errNoNBD }

func attachNBD(int64) (device, net.Conn, error) { return nil, nil, errNoNBD }
