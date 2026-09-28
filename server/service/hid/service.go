package hid

import (
	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/inputcontrol"
)

type Service struct {
	hid         *Hid
	control     *controlmode.Manager
	coordinator *inputcontrol.Coordinator
	paste       *pasteManager
}

func NewService() *Service {
	return &Service{
		hid:         GetHid(),
		control:     controlmode.GetManager(),
		coordinator: inputcontrol.GetCoordinator(),
		paste:       newPasteManager(),
	}
}

func (s *Service) newManualSession() *inputcontrol.ManualSession {
	return inputcontrol.NewManualSession(s.control, s.coordinator)
}
