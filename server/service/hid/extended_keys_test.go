package hid

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

func TestExtendedKeyReports(t *testing.T) {
	for _, tc := range []struct {
		page           string
		usage          int
		press, release []byte
	}{
		{"consumer", 0xe9, []byte{ConsumerReportID, 0xe9, 0x00}, []byte{ConsumerReportID, 0, 0}},
		{"consumer", 0x223, []byte{ConsumerReportID, 0x23, 0x02}, []byte{ConsumerReportID, 0, 0}},
		{"consumer", 0x3ff, []byte{ConsumerReportID, 0xff, 0x03}, []byte{ConsumerReportID, 0, 0}},
		{"consumer", 1, []byte{ConsumerReportID, 0x01, 0x00}, []byte{ConsumerReportID, 0, 0}},
		{"system", 0x81, []byte{SystemReportID, 0x81}, []byte{SystemReportID, 0}},
		{"system", 0x82, []byte{SystemReportID, 0x82}, []byte{SystemReportID, 0}},
		{"system", 0xb7, []byte{SystemReportID, 0xb7}, []byte{SystemReportID, 0}},
	} {
		press, release, err := extendedKeyReports(tc.page, tc.usage)
		if err != nil {
			t.Errorf("%s 0x%x: %s", tc.page, tc.usage, err)
			continue
		}
		if !bytes.Equal(press, tc.press) || !bytes.Equal(release, tc.release) {
			t.Errorf("%s 0x%x: press %x release %x, want %x %x", tc.page, tc.usage, press, release, tc.press, tc.release)
		}
	}
}

func TestExtendedKeyReportsRejectBadInput(t *testing.T) {
	for _, tc := range []struct {
		page  string
		usage int
	}{
		{"consumer", 0},
		{"consumer", -1},
		{"consumer", 0x400},
		{"system", 0x80},
		{"system", 0xb8},
		{"system", 0},
		{"keyboard", 0x04},
		{"", 0xe9},
	} {
		if _, _, err := extendedKeyReports(tc.page, tc.usage); err == nil {
			t.Errorf("%q 0x%x was accepted", tc.page, tc.usage)
		}
	}
}

func TestKeyReportsCarryNoPointerID(t *testing.T) {
	gadgetReportLength(t, "7\n")
	got := recordWrites(t)
	h := &Hid{absReportID: AbsolutePointerReportID}
	openDevice(t, h.extendedKeyDevice())

	report := []byte{ConsumerReportID, 0xe9, 0x00}
	if err := h.WriteExtendedKeyReport(report); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 || !bytes.Equal((*got)[0], report) {
		t.Fatalf("wrote %x, want exactly %x", *got, report)
	}
}

func TestKeyReportsNeedAGadgetWithIDs(t *testing.T) {
	for _, value := range []string{"6\n", ""} {
		gadgetReportLength(t, value)
		got := recordWrites(t)
		h := &Hid{}
		openDevice(t, h.extendedKeyDevice())

		err := h.WriteExtendedKeyReport([]byte{ConsumerReportID, 0xe9, 0x00})
		if !errors.Is(err, errExtendedKeysUnavailable) {
			t.Errorf("report_length %q: err %v, want errExtendedKeysUnavailable", value, err)
		}
		if len(*got) != 0 {
			t.Errorf("report_length %q: wrote %x, want nothing", value, *got)
		}
		if h.ExtendedKeysAvailable() {
			t.Errorf("report_length %q: ExtendedKeysAvailable is true", value)
		}
	}
}

func TestKeyReportsRejectAnythingButTheKeyIDs(t *testing.T) {
	gadgetReportLength(t, "7\n")
	got := recordWrites(t)
	h := &Hid{}
	openDevice(t, h.extendedKeyDevice())

	for _, report := range [][]byte{
		{AbsolutePointerReportID, 0, 0, 0, 0, 0, 0},
		{ConsumerReportID, 0xe9},
		{SystemReportID, 0x82, 0x00},
		{},
	} {
		if err := h.WriteExtendedKeyReport(report); err == nil {
			t.Errorf("% x was accepted", report)
		}
	}
	if len(*got) != 0 {
		t.Fatalf("wrote %x, want nothing", *got)
	}
}

func TestPressExtendedKeyPressesThenReleases(t *testing.T) {
	var got [][]byte
	write := func(r []byte) error { got = append(got, r); return nil }
	press, release := []byte{ConsumerReportID, 0xe9, 0}, []byte{ConsumerReportID, 0, 0}

	if err := pressExtendedKey(context.Background(), write, press, release, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || !bytes.Equal(got[0], press) || !bytes.Equal(got[1], release) {
		t.Fatalf("wrote %x, want press then release", got)
	}
}

func TestPressExtendedKeyReleasesWhenTheRequestEnds(t *testing.T) {
	var got [][]byte
	write := func(r []byte) error { got = append(got, r); return nil }
	press, release := []byte{SystemReportID, 0x82}, []byte{SystemReportID, 0}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := pressExtendedKey(ctx, write, press, release, time.Hour)

	if err == nil {
		t.Fatal("a cancelled request reported success")
	}
	if len(got) != 2 || !bytes.Equal(got[1], release) {
		t.Fatalf("wrote %x, want the release after the press", got)
	}
}

func TestPressExtendedKeySkipsTheWaitWhenThePressFails(t *testing.T) {
	var got [][]byte
	fail := errors.New("stalled")
	write := func(r []byte) error {
		got = append(got, r)
		if len(got) == 1 {
			return fail
		}
		return nil
	}
	press, release := []byte{ConsumerReportID, 0xe9, 0}, []byte{ConsumerReportID, 0, 0}

	start := time.Now()
	err := pressExtendedKey(context.Background(), write, press, release, time.Hour)
	if !errors.Is(err, fail) {
		t.Fatalf("err %v, want the press error", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("waited out the hold after a failed press")
	}
	if len(got) != 2 || !bytes.Equal(got[1], release) {
		t.Fatalf("wrote %x, want a release after the failed press", got)
	}
}

// The configfs check runs before the lock, and a mode switch can rebuild the
// gadget without IDs between that check and the write. The write reopens the
// node and must then refuse: on a pointer without IDs, 02 e9 00 is a right
// click, and the release 02 00 00 holds the right button down.
func TestKeyReportsAreRefusedWhenTheOpenHandleHasNoIDs(t *testing.T) {
	gadgetReportLength(t, "7\n")
	got := recordWrites(t)
	h := &Hid{}
	openDevice(t, h.extendedKeyDevice())
	before := statusFor(t, h, NameAbsoluteMouse).State

	err := h.WriteExtendedKeyReport([]byte{ConsumerReportID, 0xe9, 0x00})
	if !errors.Is(err, errExtendedKeysUnavailable) {
		t.Fatalf("err %v, want errExtendedKeysUnavailable", err)
	}
	if len(*got) != 0 {
		t.Fatalf("wrote %x, want nothing", *got)
	}
	if after := statusFor(t, h, NameAbsoluteMouse).State; after != before {
		t.Fatalf("a refused key report moved the pointer's state from %s to %s", before, after)
	}
}

func TestPressExtendedKeyRetriesAFailedRelease(t *testing.T) {
	var got [][]byte
	write := func(r []byte) error {
		got = append(got, r)
		if len(got) == 2 {
			return errors.New("stalled")
		}
		return nil
	}
	press, release := []byte{ConsumerReportID, 0xe9, 0}, []byte{ConsumerReportID, 0, 0}

	if err := pressExtendedKey(context.Background(), write, press, release, time.Millisecond); err != nil {
		t.Fatalf("err %v after a release that went through on the retry", err)
	}
	if len(got) != 3 || !bytes.Equal(got[2], release) {
		t.Fatalf("wrote %x, want press, release, release", got)
	}
}
