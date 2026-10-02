package picoclaw

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/hid"

	log "github.com/sirupsen/logrus"
)

const picoclawMediaTempDirName = "picoclaw_media"

func ReleaseSession(sessionID string) {
	_, err := releaseOwnedSession(GetSessionLock(), sessionID, hid.ReleaseAllHIDState)
	if err != nil {
		log.Errorf("failed to release HID state for PicoClaw session %s: %v", sessionID, err)
	}
}

func (s *Service) releaseGatewaySession(sessionID string) {
	if s == nil {
		ReleaseSession(sessionID)
		return
	}
	s.ensureDependencies()
	// Only PicoClaw sends input while it is in control, so only what it left
	// held needs releasing. After a turn that pressed nothing this writes
	// nothing, and in particular never touches the relative mouse endpoint,
	// which times out whenever the host is not polling it.
	var releaseHID func() error
	if s.control.Current() == controlmode.ModePicoclaw {
		releaseHID = s.releaseHeldInput
	}
	if _, err := releaseOwnedSession(s.lock, sessionID, releaseHID); err != nil {
		log.Warnf("PicoClaw session %s ended with input still held on the host: %v", sessionID, err)
	}
}

func releaseOwnedSession(lock *SessionLock, sessionID string, releaseHID func() error) (bool, error) {
	if lock == nil || !lock.ReleaseOwned(sessionID) {
		return false, nil
	}
	if releaseHID == nil {
		return true, nil
	}
	return true, releaseHID()
}

// releaseHeldInput lets go of any key or mouse button PicoClaw's last report
// left down. The button release goes to where the pointer is, so it does not
// move the pointer.
func (s *Service) releaseHeldInput() error {
	keys, buttons := s.held.get()

	var errs []error
	if keys {
		if err := s.hid.WriteKeyboardReport(make([]byte, hid.KeyboardReportLen)); err != nil {
			errs = append(errs, fmt.Errorf("release keyboard: %w", err))
		} else {
			s.held.setKeys(false)
		}
	}
	if buttons {
		report := make([]byte, hid.AbsoluteMouseReportLen)
		if x, y, known := s.pointer.get(); known {
			report = absoluteMouseReport(x, y, 0x00, 0)
		}
		if err := s.hid.WriteAbsoluteMouseReport(report); err != nil {
			errs = append(errs, fmt.Errorf("release mouse buttons: %w", err))
		} else {
			s.held.setButtons(false)
		}
	}
	return errors.Join(errs...)
}

func picoclawMediaTempDir() string {
	return filepath.Join(os.TempDir(), picoclawMediaTempDirName)
}

func cleanupPicoclawMediaTempDir() {
	mediaDir := picoclawMediaTempDir()

	if _, err := os.Stat(mediaDir); err != nil {
		if !os.IsNotExist(err) {
			log.Warnf("failed to stat picoclaw media directory %s: %v", mediaDir, err)
		}
		return
	}

	if err := os.RemoveAll(mediaDir); err != nil {
		log.Warnf("failed to remove picoclaw media directory %s: %v", mediaDir, err)
		return
	}

	log.Infof("removed picoclaw media directory %s", mediaDir)
}
