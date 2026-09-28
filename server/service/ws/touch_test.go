package ws

import (
	"path/filepath"
	"slices"
	"testing"

	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/hid"
	"NanoKVM-Server/service/inputcontrol"
)

func TestDecodeTouchFrame(t *testing.T) {
	contacts, err := decodeTouchFrame([]byte{
		2,
		0x01, 0, 0x34, 0x12, 0x78, 0x56,
		0x00, 1, 0xff, 0x7f, 0x00, 0x00,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []hid.TouchContact{
		{ID: 0, Down: true, X: 0x1234, Y: 0x5678},
		{ID: 1, Down: false, X: 0x7fff, Y: 0},
	}
	if !slices.Equal(contacts, want) {
		t.Fatalf("decoded %+v, want %+v", contacts, want)
	}
}

func TestDecodeTouchFrameRefusesMalformedFrames(t *testing.T) {
	for name, payload := range map[string][]byte{
		"empty":             {},
		"no contacts":       {0},
		"three contacts":    append([]byte{3}, make([]byte, 18)...),
		"short":             {1, 0x01, 0, 0, 0, 0},
		"long":              {1, 0x01, 0, 0, 0, 0, 0, 0},
		"unknown flags":     {1, 0x03, 0, 0, 0, 0, 0},
		"contact ID 128":    {1, 0x01, 0x80, 0, 0, 0, 0},
		"X beyond 32767":    {1, 0x01, 0, 0x00, 0x80, 0, 0},
		"Y beyond 32767":    {1, 0x01, 0, 0, 0, 0x00, 0x80},
		"the same ID twice": {2, 0x01, 1, 0, 0, 0, 0, 0x01, 1, 0, 0, 0, 0},
	} {
		if contacts, err := decodeTouchFrame(payload); err == nil {
			t.Errorf("%s: decoded %+v", name, contacts)
		}
	}
}

// The message type is the byte the web client puts first. It is part of the
// protocol, so a renumbering must fail here and not in a browser.
func TestTouchEventIsMessageTypeThree(t *testing.T) {
	if TouchEvent != 3 {
		t.Fatalf("TouchEvent is %d, want 3", TouchEvent)
	}
}

func TestATouchFrameIsQueuedForTheMouseWorker(t *testing.T) {
	control := controlmode.NewManager(filepath.Join(t.TempDir(), "mode"), controlmode.ModeMCP)
	manual := inputcontrol.NewManualSession(control, &inputcontrol.Coordinator{})
	defer manual.Close()

	client := &Client{manual: manual, mouse: make(chan hid.QueuedReport, 1)}
	frame := []hid.TouchContact{{ID: 0, Down: true, X: 1, Y: 2}}
	client.queueManualTouch(frame)

	select {
	case queued := <-client.mouse:
		if !slices.Equal(queued.Touch, frame) || queued.Data != nil {
			t.Fatalf("queued Touch %+v Data %v, want the frame and no report", queued.Touch, queued.Data)
		}
		if queued.Complete == nil || queued.ResetAbsoluteMouse == nil {
			t.Fatal("the queued frame cannot complete or reset its lease")
		}
		queued.Complete(true)
	default:
		t.Fatal("nothing was queued")
	}
}
