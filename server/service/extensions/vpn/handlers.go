package vpn

import (
	"errors"
	"sync"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"

	"github.com/gin-gonic/gin"
)

// Refuse answers the request with the reason when the other VPN runs or
// starts at boot, and reports whether it did. A caller that acts on a false
// holds addon.LockVPN across both, or uses Guard.
func Refuse(c *gin.Context, d addon.Daemon) bool {
	err := addon.CheckExclusive(d.Name)
	if err == nil {
		return false
	}
	var rsp proto.Response
	rsp.ErrRsp(c, -1, err.Error())
	return true
}

// Boot handles POST boot for either add-on: {enabled} turns start at boot on
// or off. Turning it on is refused while the other VPN runs or starts at boot;
// turning it off never is. Both run under the VPN lock.
func Boot(c *gin.Context, d addon.Daemon, installed bool) {
	var req proto.VpnBootReq
	var rsp proto.Response

	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if req.Enabled && !installed {
		rsp.ErrRsp(c, -1, d.Title+" is not installed")
		return
	}

	defer addon.LockVPN()()
	if req.Enabled && Refuse(c, d) {
		return
	}
	if err := addon.SetBoot(d, req.Enabled); err != nil {
		rsp.ErrRsp(c, -1, Message("start at boot failed", err))
		return
	}
	rsp.OkRsp(c)
}

// Guard runs action under the VPN lock after the exclusivity check, so no
// request for the other VPN can act between the two. It answers the request
// with the refusal, or with what failed and why, and reports success.
func Guard(c *gin.Context, d addon.Daemon, what string, action func() error) bool {
	var rsp proto.Response
	err := addon.Exclusive(d.Name, action)
	var be *addon.BlockedError
	if errors.As(err, &be) {
		rsp.ErrRsp(c, -1, err.Error())
		return false
	}
	if err != nil {
		rsp.ErrRsp(c, -1, Message(what, err))
		return false
	}
	return true
}

// StartChecked starts a daemon after an install or an update, under the VPN
// lock and after the exclusivity check. The download before it ran without
// the lock, and the other VPN may have started meanwhile.
func StartChecked(d addon.Daemon, start func() error) error {
	return addon.Exclusive(d.Name, start)
}

// StopDaemon runs the boot script's stop. The scripts print FAIL both for a
// stop that left the daemon running and when nothing ran, so the failure
// counts only while the daemon is still there.
func StopDaemon(d addon.Daemon) error {
	err := Script(d.Script(), "stop", "")
	if err != nil && !addon.Running(d) {
		return nil
	}
	return err
}

// Busy marks an add-on that is installing, updating or uninstalling. Those
// share a workspace and the binary, so only one of them runs at a time.
type Busy struct {
	mu sync.Mutex
}

// Begin returns the function that ends this action, or answers the request
// with a busy message and returns nil when another one runs.
func (b *Busy) Begin(c *gin.Context, d addon.Daemon) func() {
	if !b.mu.TryLock() {
		var rsp proto.Response
		rsp.ErrRsp(c, -1, d.Title+" is busy with an install, update or uninstall. Try again when it has finished.")
		return nil
	}
	return b.mu.Unlock
}
