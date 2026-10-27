package picoclaw

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"NanoKVM-Server/service/hid"

	log "github.com/sirupsen/logrus"
)

const picoclawMediaTempDirName = "picoclaw_media"

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
