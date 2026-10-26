//go:build !linux

package wireguard

import (
	"errors"
	"net"
	"net/netip"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// The board runs Linux; elsewhere the package builds for the tests and the
// tools, and every change fails.
type unsupported struct{}

func newSystem() system { return unsupported{} }

var errUnsupported = errors.New("WireGuard needs Linux")

func (unsupported) linkExists(string) (bool, error)          { return false, nil }
func (unsupported) addLink(string) error                     { return errUnsupported }
func (unsupported) delLink(string) error                     { return errUnsupported }
func (unsupported) configure(string, wgtypes.Config) error   { return errUnsupported }
func (unsupported) device(string) (*wgtypes.Device, error)   { return nil, errUnsupported }
func (unsupported) setLink(string, int, bool) error          { return errUnsupported }
func (unsupported) addAddr(string, netip.Prefix) error       { return errUnsupported }
func (unsupported) addRoute(string, netip.Prefix, int) error { return errUnsupported }
func (unsupported) addRule(rule) error                       { return errUnsupported }
func (unsupported) delRule(rule) error                       { return errUnsupported }
func (unsupported) setSrcValidMark() error                   { return errUnsupported }
func (unsupported) resolve(string) (*net.UDPAddr, error)     { return nil, errUnsupported }
