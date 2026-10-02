package picoclaw

import (
	"context"

	"NanoKVM-Server/service/vm"
)

// acquireCaptureLease holds HDMI capture for one screenshot. The task lease
// that keeps capture running for a whole turn is agent.TaskLeases.
func (s *Service) acquireCaptureLease(ctx context.Context) (func(), func() bool, error) {
	if s != nil && s.acquireHDMIForRead != nil {
		return s.acquireHDMIForRead(ctx)
	}
	return vm.AcquireHdmiCaptureLeaseForRead(ctx)
}
