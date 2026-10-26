package vm

import (
	"context"
	"testing"
)

// useIdleHdmi stubs the HDMI switch and settings so a lease can be taken and
// released without libkvm or the config file.
func useIdleHdmi(t *testing.T) {
	t.Helper()
	useHdmiDemand(t)

	originalSet, originalDisabled, originalTimeout := setHDMI, isHdmiDisabled, getHdmiIdleTimeout
	setHDMI = func(bool) {}
	isHdmiDisabled = func() bool { return false }
	getHdmiIdleTimeout = func() int { return 0 }
	t.Cleanup(func() {
		setHDMI, isHdmiDisabled, getHdmiIdleTimeout = originalSet, originalDisabled, originalTimeout
	})
}

func TestReadLeaseWithoutViewerDiscardsQueuedFrame(t *testing.T) {
	useIdleHdmi(t)

	release, claimFresh, err := AcquireHdmiCaptureLeaseForRead(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if claimFresh == nil || !claimFresh() {
		t.Fatal("a read with no viewer keeps the frame queued before it")
	}
}

func TestReadLeaseWithViewerKeepsFrame(t *testing.T) {
	useIdleHdmi(t)
	UpdateHdmiViewerSnapshot("direct", 1, 1)

	release, claimFresh, err := AcquireHdmiCaptureLeaseForRead(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if claimFresh != nil && claimFresh() {
		t.Fatal("a read discards a frame while a stream drains the queue")
	}
}
