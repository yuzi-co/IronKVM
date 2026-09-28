package hid

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

const s03usbdev = "../../../kvmapp/system/init.d/S03usbdev"

// gadgetReportDesc points the configfs descriptor read at a scratch file
// holding descriptor, or at a missing file when descriptor is nil.
func gadgetReportDesc(t *testing.T, descriptor []byte) {
	t.Helper()

	restore := absoluteReportDescPath
	t.Cleanup(func() { absoluteReportDescPath = restore })

	path := filepath.Join(t.TempDir(), "report_desc")
	if descriptor != nil {
		if err := os.WriteFile(path, descriptor, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	absoluteReportDescPath = path
}

// touchScriptDescriptors returns the three absolute descriptors S03usbdev
// writes, in the order it writes them: touch, extended keys, plain.
func touchScriptDescriptors(t *testing.T) (touch, extkeys, plain []byte) {
	t.Helper()

	descriptors := absoluteDescriptors(t, s03usbdev, readScript(t, s03usbdev))
	if len(descriptors) != 3 {
		t.Fatalf("%s writes %d absolute descriptors, want 3", s03usbdev, len(descriptors))
	}
	return descriptors[0], descriptors[1], descriptors[2]
}

func TestTouchReportsEncodeOneContactEach(t *testing.T) {
	reports, err := touchReports([]TouchContact{
		{ID: 0, Down: true, X: 0x1234, Y: 0x5678},
		{ID: 1, Down: false, X: 0x7fff, Y: 0},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Tip Switch is bit 0 of the second byte and Contact ID the seven bits
	// above it. Only the first report says how many contacts the frame has.
	want := [][]byte{
		{TouchReportID, 0x01, 0x34, 0x12, 0x78, 0x56, 2},
		{TouchReportID, 0x02, 0xff, 0x7f, 0x00, 0x00, 0},
	}
	if len(reports) != len(want) {
		t.Fatalf("got %d reports, want %d", len(reports), len(want))
	}
	for i := range want {
		if !bytes.Equal(reports[i], want[i]) {
			t.Errorf("report %d = % x, want % x", i, reports[i], want[i])
		}
		if len(reports[i]) != TouchReportLen {
			t.Errorf("report %d is %d bytes, want %d", i, len(reports[i]), TouchReportLen)
		}
	}
}

func TestTouchReportsCarryTheHighContactID(t *testing.T) {
	reports, err := touchReports([]TouchContact{{ID: 0x7f, Down: true, X: 1, Y: 2}})
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{TouchReportID, 0xff, 0x01, 0x00, 0x02, 0x00, 1}
	if len(reports) != 1 || !bytes.Equal(reports[0], want) {
		t.Fatalf("got % x, want one report % x", reports, want)
	}
}

func TestValidateTouchFrameRefusesWhatTheDescriptorCannotSay(t *testing.T) {
	for name, frame := range map[string][]TouchContact{
		"no contacts":        nil,
		"three contacts":     {{ID: 0}, {ID: 1}, {ID: 2}},
		"contact ID too big": {{ID: 0x80}},
		"X too big":          {{X: 0x8000}},
		"Y too big":          {{Y: 0x8000}},
		"duplicate IDs":      {{ID: 1}, {ID: 1}},
	} {
		if err := ValidateTouchFrame(frame); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if err := ValidateTouchFrame([]TouchContact{{ID: 0, X: 0x7fff, Y: 0x7fff}, {ID: 0x7f}}); err != nil {
		t.Errorf("a frame at the limits was refused: %s", err)
	}
}

func TestHeldTouchesFollowTheFrames(t *testing.T) {
	a := TouchContact{ID: 0, Down: true, X: 10, Y: 10}
	b := TouchContact{ID: 1, Down: true, X: 20, Y: 20}

	held := heldTouches(nil, []TouchContact{a, b})
	if !slices.Equal(held, []TouchContact{a, b}) {
		t.Fatalf("held after two downs = %v", held)
	}

	// b lifts, a moves.
	moved := TouchContact{ID: 0, Down: true, X: 11, Y: 12}
	held = heldTouches(held, []TouchContact{moved, {ID: 1, Down: false, X: 20, Y: 20}})
	if !slices.Equal(held, []TouchContact{moved}) {
		t.Fatalf("held after a lift = %v", held)
	}

	// A frame that leaves a held contact out does not prove it lifted.
	c := TouchContact{ID: 5, Down: true}
	held = heldTouches(held, []TouchContact{c})
	if !slices.Equal(held, []TouchContact{moved, c}) {
		t.Fatalf("held after a frame without contact 0 = %v", held)
	}
}

func TestTouchLiftFramesStayWithinTheContactLimit(t *testing.T) {
	held := []TouchContact{
		{ID: 0, Down: true, X: 1}, {ID: 1, Down: true, X: 2}, {ID: 2, Down: true, X: 3},
	}
	frames := touchLiftFrames(held)
	if len(frames) != 2 || len(frames[0]) != 2 || len(frames[1]) != 1 {
		t.Fatalf("lift frames = %v", frames)
	}
	for _, frame := range frames {
		if err := ValidateTouchFrame(frame); err != nil {
			t.Errorf("a lift frame is invalid: %s", err)
		}
		for _, contact := range frame {
			if contact.Down {
				t.Errorf("lift frame keeps contact %d down", contact.ID)
			}
		}
	}
	if frames[1][0].X != 3 {
		t.Errorf("contact 2 lifted at X %d, want where it was held", frames[1][0].X)
	}
	if touchLiftFrames(nil) != nil {
		t.Error("nothing held still produced a lift frame")
	}
}

// The touch descriptor is the extended-keys descriptor with the touch screen
// appended. The pointer, the keys and their report IDs come first and
// unchanged, which is what keeps every pointer report byte for byte the same
// as with /boot/usb.extkeys.
func TestTheTouchDescriptorExtendsTheExtendedKeysOne(t *testing.T) {
	touch, extkeys, plain := touchScriptDescriptors(t)

	if !bytes.HasPrefix(touch, extkeys) {
		t.Fatalf("the touch descriptor does not start with the extended-keys descriptor")
	}
	if containsBytes(plain, []byte{0x85}) {
		t.Errorf("the default descriptor declares a report ID")
	}
	script := readScript(t, s03usbdev)
	if !strings.Contains(script, "if [ -e /boot/usb.touch ]") {
		t.Errorf("the touch descriptor is not behind /boot/usb.touch")
	}
	if strings.Index(script, "/boot/usb.touch ]") > strings.Index(script, "/boot/usb.extkeys ]") {
		t.Errorf("/boot/usb.extkeys is tested before /boot/usb.touch, so both markers give no touch")
	}
}

// reportBits totals the bits of every field of one kind under one report ID.
func reportBits(fields []descriptorField, main byte, id byte) int {
	bits := 0
	for _, field := range fields {
		if field.main == main && field.reportID == id {
			bits += field.size * field.count
		}
	}
	return bits
}

func TestTheTouchDescriptorDeclaresTheTouchScreen(t *testing.T) {
	touch, _, _ := touchScriptDescriptors(t)

	fields, err := parseReportDescriptor(touch)
	if err != nil {
		t.Fatal(err)
	}

	// Every input report fits the report_length the script gives, and the
	// pointer's is exactly the report this package prefixes.
	for _, id := range []byte{AbsolutePointerReportID, ConsumerReportID, SystemReportID, TouchReportID} {
		bits := reportBits(fields, mainInput, id)
		if bits%8 != 0 || bits == 0 || bits/8+1 > AbsoluteMouseReportLenWithID {
			t.Errorf("input report %d is %d bits, which does not fit report_length %d", id, bits, AbsoluteMouseReportLenWithID)
		}
	}
	if got := reportBits(fields, mainInput, AbsolutePointerReportID); got != AbsoluteMouseReportLen*8 {
		t.Errorf("the pointer report is %d bits, want %d", got, AbsoluteMouseReportLen*8)
	}
	if got := reportBits(fields, mainInput, TouchReportID); got != (TouchReportLen-1)*8 {
		t.Errorf("the touch report is %d bits, want %d", got, (TouchReportLen-1)*8)
	}

	// The touch input report, field by field, in the order touchReports
	// writes it.
	type want struct {
		usages []uint32
		size   int
		count  int
		finger bool
	}
	wants := []want{
		{[]uint32{0x000d0042}, 1, 1, true},              // Tip Switch
		{[]uint32{0x000d0051}, 7, 1, true},              // Contact Identifier
		{[]uint32{0x00010030, 0x00010031}, 16, 2, true}, // X, Y
		{[]uint32{0x000d0054}, 8, 1, false},             // Contact Count
	}
	var got []descriptorField
	for _, field := range fields {
		if field.main == mainInput && field.reportID == TouchReportID {
			got = append(got, field)
		}
	}
	if len(got) != len(wants) {
		t.Fatalf("the touch report has %d fields, want %d", len(got), len(wants))
	}
	for i, w := range wants {
		f := got[i]
		if !slices.Equal(f.usages, w.usages) || f.size != w.size || f.count != w.count {
			t.Errorf("touch field %d: usages %x size %d count %d, want %x size %d count %d",
				i, f.usages, f.size, f.count, w.usages, w.size, w.count)
		}
		if f.application != usageTouchScreen {
			t.Errorf("touch field %d is in application %x, want the touch screen", i, f.application)
		}
		inFinger := slices.Contains(f.collections, 0x000d0022)
		if inFinger != w.finger {
			t.Errorf("touch field %d: inside the finger collection %t, want %t", i, inFinger, w.finger)
		}
	}

	// Contact Count Maximum, and the certification blob that puts Linux's
	// hid-multitouch in the class that keeps the mouse collection.
	features := map[uint32]descriptorField{}
	for _, field := range fields {
		if field.main == mainFeature && len(field.usages) == 1 {
			features[field.usages[0]] = field
		}
	}
	if f, ok := features[0x000d0055]; !ok || f.reportID != TouchMaxCountReportID || f.application != usageTouchScreen {
		t.Errorf("no Contact Count Maximum feature under ID %d in the touch screen: %+v", TouchMaxCountReportID, f)
	}
	if f, ok := features[0xff0000c5]; !ok || f.size != 8 || f.count != 256 || f.reportID != TouchBlobReportID {
		t.Errorf("no 256-byte certification blob under ID %d: %+v", TouchBlobReportID, f)
	}

	if !touchDeclared(touch) {
		t.Error("touchDeclared does not recognise the script's own descriptor")
	}
}

func TestOnlyTheTouchDescriptorDeclaresTouch(t *testing.T) {
	_, extkeys, plain := touchScriptDescriptors(t)
	if touchDeclared(extkeys) {
		t.Error("the extended-keys descriptor reads as touch")
	}
	if touchDeclared(plain) {
		t.Error("the default descriptor reads as touch")
	}
	for _, descriptor := range absoluteDescriptors(t, "../../../kvmapp/system/init.d/S03usbhid", readScript(t, "../../../kvmapp/system/init.d/S03usbhid")) {
		if touchDeclared(descriptor) {
			t.Error("the hid-only descriptor reads as touch")
		}
	}
	if touchDeclared([]byte{0x05, 0x0d, 0x09, 0x04, 0xa1, 0x01}) {
		t.Error("a truncated descriptor reads as touch")
	}
}

func TestParseReportDescriptorRefusesBrokenInput(t *testing.T) {
	for name, descriptor := range map[string][]byte{
		"cut inside an item":     {0x26, 0xff},
		"collection left open":   {0xa1, 0x01},
		"end without collection": {0xc0},
		"long item":              {0xfe, 0x00, 0x00},
		"pop without push":       {0xb4},
	} {
		if _, err := parseReportDescriptor(descriptor); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestTouchAvailableReadsTheGadget(t *testing.T) {
	touch, extkeys, _ := touchScriptDescriptors(t)

	for _, tc := range []struct {
		name       string
		length     string
		descriptor []byte
		want       bool
	}{
		{"touch descriptor", "7\n", touch, true},
		{"extended keys only", "7\n", extkeys, false},
		{"no descriptor file", "7\n", nil, false},
		{"touch descriptor but no IDs", "6\n", touch, false},
	} {
		gadgetReportLength(t, tc.length)
		gadgetReportDesc(t, tc.descriptor)
		if got := (&Hid{}).TouchAvailable(); got != tc.want {
			t.Errorf("%s: TouchAvailable %t, want %t", tc.name, got, tc.want)
		}
	}
}

func TestOpeningThePointerReadsTouch(t *testing.T) {
	touch, extkeys, _ := touchScriptDescriptors(t)
	gadgetReportLength(t, "7\n")
	gadgetReportDesc(t, touch)

	node := filepath.Join(t.TempDir(), "hidg2")
	if err := os.WriteFile(node, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	h := &Hid{}
	device := h.absoluteMouseDevice(node)
	if err := h.openDeviceNoLock(device); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { h.closeDeviceNoLock(device) })
	if !h.absTouch {
		t.Fatal("opening on a touch gadget left touch off")
	}

	// stop_start with the marker gone keeps report_length 7 and the node, so
	// only the descriptor tells.
	if err := os.WriteFile(absoluteReportDescPath, extkeys, 0o644); err != nil {
		t.Fatal(err)
	}
	if !h.RefreshAbsoluteReportID() {
		t.Fatal("a descriptor that lost touch was not noticed")
	}
	if err := h.openDeviceNoLock(device); err != nil {
		t.Fatal(err)
	}
	if h.absTouch {
		t.Fatal("touch still on after the reopen")
	}
}

func TestTouchFramesAreWrittenOnlyWhenTheGadgetHasTouch(t *testing.T) {
	got := recordWrites(t)
	h := &Hid{absReportID: AbsolutePointerReportID}
	openDevice(t, h.absoluteMouseDevice(HID2))

	frame := []TouchContact{{ID: 0, Down: true, X: 1, Y: 2}, {ID: 1, Down: true, X: 3, Y: 4}}
	if err := h.WriteTouchFrame(frame); !errors.Is(err, errTouchUnavailable) {
		t.Fatalf("a frame without touch returned %v, want errTouchUnavailable", err)
	}
	if len(*got) != 0 {
		t.Fatalf("a refused frame wrote %x", *got)
	}

	h.absTouch = true
	if err := h.WriteTouchFrame(frame); err != nil {
		t.Fatal(err)
	}
	want, _ := touchReports(frame)
	if len(*got) != 2 || !bytes.Equal((*got)[0], want[0]) || !bytes.Equal((*got)[1], want[1]) {
		t.Fatalf("wrote %x, want %x", *got, want)
	}

	// The pointer is still prefixed as before.
	*got = nil
	if err := h.WriteAbsoluteMouseReport([]byte{1, 2, 3, 4, 5, 6}); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 || !bytes.Equal((*got)[0], []byte{AbsolutePointerReportID, 1, 2, 3, 4, 5, 6}) {
		t.Fatalf("the pointer wrote %x on a touch gadget", *got)
	}
}

// runMouseQueue runs the mouse worker over the given events and waits for it.
func runMouseQueue(t *testing.T, h *Hid, events ...QueuedReport) {
	t.Helper()

	queue := make(chan QueuedReport, len(events))
	for _, event := range events {
		queue <- event
	}
	close(queue)

	done := make(chan struct{})
	go func() {
		h.mouseReports(queue, "unused-relative", "unused-absolute")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("mouse worker did not stop")
	}
}

func touchHid(t *testing.T) *Hid {
	t.Helper()

	h := &Hid{absReportID: AbsolutePointerReportID, absTouch: true}
	openDevice(t, h.absoluteMouseDevice(HID2))
	return h
}

// A finger still down when the websocket goes away would stay on the host's
// screen, the touch equivalent of a stuck button.
func TestClosingTheQueueLiftsHeldContacts(t *testing.T) {
	got := recordWrites(t)
	h := touchHid(t)

	runMouseQueue(t, h, QueuedReport{Touch: []TouchContact{{ID: 0, Down: true, X: 0x100, Y: 0x200}}})

	want := [][]byte{
		{TouchReportID, 0x01, 0x00, 0x01, 0x00, 0x02, 1},
		{TouchReportID, 0x00, 0x00, 0x01, 0x00, 0x02, 1},
	}
	if !slices.EqualFunc(*got, want, bytes.Equal) {
		t.Fatalf("wrote %x, want %x", *got, want)
	}
}

func TestLiftedContactsAreNotLiftedAgain(t *testing.T) {
	got := recordWrites(t)
	h := touchHid(t)

	runMouseQueue(t, h,
		QueuedReport{Touch: []TouchContact{{ID: 0, Down: true}}},
		QueuedReport{Touch: []TouchContact{{ID: 0, Down: false}}},
	)
	if len(*got) != 2 {
		t.Fatalf("wrote %d reports, want the down and the up only: %x", len(*got), *got)
	}
}

func TestAPointerReportLiftsTheFingersFirst(t *testing.T) {
	got := recordWrites(t)
	h := touchHid(t)

	pointer := []byte{0, 0x10, 0, 0x20, 0, 0}
	runMouseQueue(t, h,
		QueuedReport{Touch: []TouchContact{{ID: 1, Down: true, X: 5, Y: 6}}},
		QueuedReport{Data: pointer},
	)

	want := [][]byte{
		{TouchReportID, 0x03, 5, 0, 6, 0, 1},
		{TouchReportID, 0x02, 5, 0, 6, 0, 1},
		append([]byte{AbsolutePointerReportID}, pointer...),
	}
	if !slices.EqualFunc(*got, want, bytes.Equal) {
		t.Fatalf("wrote %x, want %x", *got, want)
	}
}

func TestATouchFrameReleasesHeldButtonsFirst(t *testing.T) {
	got := recordWrites(t)
	h := touchHid(t)

	down := []byte{1, 0x10, 0, 0x20, 0, 0}
	runMouseQueue(t, h,
		QueuedReport{Data: down},
		QueuedReport{Touch: []TouchContact{{ID: 0, Down: false}}},
	)

	if len(*got) != 3 {
		t.Fatalf("wrote %x, want the press, its release and the frame", *got)
	}
	release := append([]byte{AbsolutePointerReportID}, absoluteMouseReleaseReport(down)...)
	if !bytes.Equal((*got)[1], release) {
		t.Fatalf("second write %x, want the button release %x", (*got)[1], release)
	}
	if (*got)[2][0] != TouchReportID {
		t.Fatalf("third write %x is not the touch frame", (*got)[2])
	}
}

func TestAFrameOnAGadgetWithoutTouchCompletesAsFailed(t *testing.T) {
	got := recordWrites(t)
	h := &Hid{absReportID: AbsolutePointerReportID}
	openDevice(t, h.absoluteMouseDevice(HID2))

	var results []bool
	runMouseQueue(t, h, QueuedReport{
		Touch:    []TouchContact{{ID: 0, Down: true}},
		Complete: func(ok bool) { results = append(results, ok) },
	})

	if !slices.Equal(results, []bool{false}) {
		t.Fatalf("completions %v, want one failure", results)
	}
	if len(*got) != 0 {
		t.Fatalf("a gadget without touch was written %x", *got)
	}
}
