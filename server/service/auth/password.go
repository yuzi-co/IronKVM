package auth

import (
	"NanoKVM-Server/proto"
	"NanoKVM-Server/utils"
	"errors"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// setRootPassword is a variable so tests can avoid shelling out to passwd(1).
var setRootPassword = changeRootPassword

// saveIdentity is a variable for the same reason.
var saveIdentity = writeBackIdentity

// identityScript carries the board's identity between slots. It is absent on an
// upstream image, and on a board that has one this is the only thing that makes
// a password change outlive the slot it was set on.
const identityScript = "/etc/init.d/S02identity"

func (s *Service) ChangePassword(c *gin.Context) {
	var req proto.ChangePasswordReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid parameters")
		return
	}

	// Require the current password, otherwise a stolen session - or a request
	// forged by another site - is enough to take over the device permanently.
	// A device that still has no account file uses the documented admin/admin
	// default, so the check adds nothing and would block the initial setup.
	if isAccountConfigured() && !CompareAccount(req.Username, req.OldPassword) {
		time.Sleep(2 * time.Second)
		rsp.ErrRsp(c, -6, "invalid current password")
		return
	}

	password, err := utils.DecodeDecrypt(req.Password)
	if err != nil || password == "" {
		rsp.ErrRsp(c, -2, "invalid password")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		rsp.ErrRsp(c, -3, "failed to hash password")
		return
	}

	if err = SetAccount(req.Username, string(hashedPassword)); err != nil {
		rsp.ErrRsp(c, -4, "failed to save password")
		return
	}

	// change root password
	err = setRootPassword(password)
	if err != nil {
		_ = DelAccount()
		rsp.ErrRsp(c, -5, "failed to change password")
		return
	}

	// Persist it beyond this slot. /etc/shadow lives in the slot's own root
	// filesystem and S02identity restores it from /data at every boot, so
	// without this a password change is undone by the next reboot and lost
	// outright by a slot switch.
	//
	// A failure here does not fail the request. The password is already
	// written, so refusing cannot undo it, and the error path above deletes the
	// account file on its way out, which would drop the web UI back to the
	// admin/admin default. A copy that is one boot stale is the better of the
	// two, and it is recoverable by changing the password again.
	if err = saveIdentity(); err != nil {
		log.Warnf("password changed but the identity write-back failed: %s", err)
	}

	rsp.OkRsp(c)
	log.Debugf("change password success, username: %s", req.Username)
}

func (s *Service) IsPasswordUpdated(c *gin.Context) {
	var rsp proto.Response

	if _, err := os.Stat(AccountFile); err != nil {
		rsp.OkRspWithData(c, &proto.IsPasswordUpdatedRsp{
			IsUpdated: false,
		})
		return
	}

	account, err := GetAccount()
	if err != nil || account == nil {
		rsp.ErrRsp(c, -1, "failed to get password")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte("admin"))

	rsp.OkRspWithData(c, &proto.IsPasswordUpdatedRsp{
		// If the hash is not valid, still assume it's not updated
		// The error we want to see is password and hash not matching
		IsUpdated: errors.Is(err, bcrypt.ErrMismatchedHashAndPassword),
	})
}

// writeBackIdentity copies the credentials the board just changed to /data, so
// the next boot of any slot restores them rather than the ones it replaced.
//
// A board without the script is an upstream image with no slot layout, where
// there is nothing to write back to and nothing to lose. That is not an error.
func writeBackIdentity() error {
	if _, err := os.Stat(identityScript); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	return exec.Command(identityScript, "save").Run()
}

func changeRootPassword(password string) error {
	err := passwd(password)
	if err != nil {
		log.Errorf("failed to change root password: %s", err)
		return err
	}

	log.Debugf("change root password successful.")
	return nil
}

func passwd(password string) error {
	cmd := exec.Command("passwd", "root")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer func() {
		_ = stdin.Close()
	}()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Start(); err != nil {
		return err
	}

	if _, err = io.WriteString(stdin, password+"\n"); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	if _, err = io.WriteString(stdin, password+"\n"); err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}
