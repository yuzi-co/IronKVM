package hid

import (
	"errors"
	"fmt"
	"os"
)

// The touch screen shares the absolute pointer's endpoint under report ID 4.
// Only S03usbdev with /boot/usb.touch declares it; everywhere else a touch
// report would reach the host as a garbled pointer report, so nothing is
// written.
//
// A report carries one contact and is as long as the pointer report with its
// ID, so report_length stays 7 and a server that knows the IDs but not touch
// still drives the pointer correctly. A frame with two contacts is two
// reports: the first says how many contacts the frame has and the second says
// 0. Windows calls this hybrid mode, and Linux's hid-multitouch reads it too.
const (
	TouchReportID byte = 4

	// TouchMaxCountReportID is the Contact Count Maximum feature report, and
	// TouchBlobReportID the Windows certification blob. The host reads both
	// with GET_REPORT, which f_hid answers with zeros; nothing here writes
	// them.
	TouchMaxCountReportID byte = 5
	TouchBlobReportID     byte = 6

	// TouchReportLen is the report with its ID: the ID, Tip Switch and
	// Contact ID in one byte, a 16-bit X, a 16-bit Y, and Contact Count.
	TouchReportLen = 7

	// TouchMaxContacts is the Contact Count Maximum the descriptor declares.
	TouchMaxContacts = 2

	touchMaxContactID = 0x7f
	touchMaxCoord     = 0x7fff
)

// Usages the descriptor walk looks for, as page << 16 | usage.
const (
	usageTouchScreen = 0x000d0004
)

var errTouchUnavailable = errors.New("the USB gadget has no touch screen report")

// absoluteReportDescPath is where S03usbdev writes the absolute pointer's
// report descriptor. configfs reads it back as written. A variable so tests
// can point it at a scratch file.
var absoluteReportDescPath = "/sys/kernel/config/usb_gadget/g0/functions/hid.GS2/report_desc"

// TouchContact is one finger in a touch frame. X and Y are 0 to 32767, the
// same range as the absolute pointer.
type TouchContact struct {
	ID   byte
	Down bool
	X    uint16
	Y    uint16
}

// ValidateTouchFrame checks a frame before anything is queued or written. A
// frame has one or two contacts with distinct IDs.
func ValidateTouchFrame(contacts []TouchContact) error {
	if len(contacts) < 1 || len(contacts) > TouchMaxContacts {
		return fmt.Errorf("a touch frame has %d contacts, want 1 to %d", len(contacts), TouchMaxContacts)
	}
	for i, contact := range contacts {
		if contact.ID > touchMaxContactID {
			return fmt.Errorf("touch contact ID %d is above %d", contact.ID, touchMaxContactID)
		}
		if contact.X > touchMaxCoord || contact.Y > touchMaxCoord {
			return fmt.Errorf("touch contact %d at %d,%d is outside 0 to %d", contact.ID, contact.X, contact.Y, touchMaxCoord)
		}
		for _, other := range contacts[:i] {
			if other.ID == contact.ID {
				return fmt.Errorf("touch contact ID %d appears twice in one frame", contact.ID)
			}
		}
	}
	return nil
}

// touchReports encodes a frame as one report per contact, in the layout
// S03usbdev declares for report ID 4. Tip Switch is the lowest bit of the
// second byte and Contact ID the seven above it. Only the first report
// carries the contact count.
func touchReports(contacts []TouchContact) ([][]byte, error) {
	if err := ValidateTouchFrame(contacts); err != nil {
		return nil, err
	}

	reports := make([][]byte, 0, len(contacts))
	for i, contact := range contacts {
		flags := contact.ID << 1
		if contact.Down {
			flags |= 1
		}
		count := byte(0)
		if i == 0 {
			count = byte(len(contacts))
		}
		reports = append(reports, []byte{
			TouchReportID,
			flags,
			byte(contact.X), byte(contact.X >> 8),
			byte(contact.Y), byte(contact.Y >> 8),
			count,
		})
	}
	return reports, nil
}

// touchFrameDown reports whether any contact in the frame is down.
func touchFrameDown(contacts []TouchContact) bool {
	for _, contact := range contacts {
		if contact.Down {
			return true
		}
	}
	return false
}

// heldTouches returns the contacts the host may still see down after a frame
// was sent, whether or not it arrived. A contact the frame lifts or leaves out
// might still be down if the frame was lost, so the ones held before stay
// unless the frame mentions them, and the frame's own down contacts are added
// at their new positions.
func heldTouches(before []TouchContact, frame []TouchContact) []TouchContact {
	var held []TouchContact
	for _, contact := range before {
		mentioned := false
		for _, next := range frame {
			if next.ID == contact.ID {
				mentioned = true
				break
			}
		}
		if !mentioned {
			held = append(held, contact)
		}
	}
	for _, contact := range frame {
		if contact.Down {
			held = append(held, contact)
		}
	}
	return held
}

// touchLiftFrames lifts every held contact where it was last seen, in as many
// frames as the contact limit needs.
func touchLiftFrames(held []TouchContact) [][]TouchContact {
	var frames [][]TouchContact
	for start := 0; start < len(held); start += TouchMaxContacts {
		end := min(start+TouchMaxContacts, len(held))
		frame := make([]TouchContact, 0, end-start)
		for _, contact := range held[start:end] {
			contact.Down = false
			frame = append(frame, contact)
		}
		frames = append(frames, frame)
	}
	return frames
}

// readAbsoluteTouch reports whether the gadget's absolute pointer descriptor
// declares the touch screen. The report length alone cannot tell: it is 7 with
// the extended keys whether or not touch is there.
func readAbsoluteTouch() bool {
	raw, err := os.ReadFile(absoluteReportDescPath)
	if err != nil {
		return false
	}
	return touchDeclared(raw)
}

// touchDeclared reports whether a report descriptor has a Touch Screen
// application collection whose input report under TouchReportID is exactly
// the report this package writes.
func touchDeclared(descriptor []byte) bool {
	fields, err := parseReportDescriptor(descriptor)
	if err != nil {
		return false
	}

	bits := 0
	for _, field := range fields {
		if field.main != mainInput || field.reportID != TouchReportID {
			continue
		}
		if field.application != usageTouchScreen {
			return false
		}
		bits += field.size * field.count
	}
	return bits == (TouchReportLen-1)*8
}

// TouchAvailable reports whether the gadget declares the touch screen. Like
// ExtendedKeysAvailable it reads configfs, and the write decides again under
// the lock.
func (h *Hid) TouchAvailable() bool {
	return readAbsoluteReportID() != 0 && readAbsoluteTouch()
}

// touchDevice is the absolute pointer's endpoint seen by the touch screen.
func (h *Hid) touchDevice() hidDevice {
	return h.touchDeviceAt(HID2)
}

func (h *Hid) touchDeviceAt(path string) hidDevice {
	return hidDevice{path: path, name: NameAbsoluteMouse, mu: &h.mouseMutex, file: &h.g2, health: &h.absHealth, needsID: true, needsTouch: true}
}

// WriteTouchFrame writes one frame, one report per contact. The mouse lock is
// held for the whole frame, so a pointer report cannot land between the two
// halves of a two-finger frame.
func (h *Hid) WriteTouchFrame(contacts []TouchContact) error {
	return h.writeTouchFrame(h.touchDevice(), contacts)
}

func (h *Hid) writeTouchFrame(device hidDevice, contacts []TouchContact) error {
	reports, err := touchReports(contacts)
	if err != nil {
		return err
	}

	device.mu.Lock()
	defer device.mu.Unlock()

	for _, report := range reports {
		err := h.writeHIDLocked(device, report)
		if errors.Is(err, errExtendedKeysUnavailable) || errors.Is(err, errTouchUnavailable) {
			// Nothing reached the endpoint, so it says nothing about its health.
			return err
		}
		if err := device.note(err); err != nil {
			return err
		}
	}
	return nil
}

// Main item kinds as parseReportDescriptor reports them: the item prefix with
// the size bits cleared.
const (
	mainInput      = 0x80
	mainOutput     = 0x90
	mainFeature    = 0xb0
	mainCollection = 0xa0
	mainEnd        = 0xc0
)

// descriptorField is one Input, Output or Feature item of a report
// descriptor, with the state it was declared under.
type descriptorField struct {
	main     byte
	reportID byte
	size     int
	count    int
	// usages are page << 16 | usage, in the order they were declared.
	usages []uint32
	// application is the usage of the enclosing application collection.
	application uint32
	// collections are the usages of every enclosing collection, outermost
	// first.
	collections []uint32
}

// parseReportDescriptor walks the short-item encoding of a HID report
// descriptor. It handles what the gadget scripts declare: short items, usage
// ranges, and push and pop. A long item is refused rather than guessed at.
func parseReportDescriptor(descriptor []byte) ([]descriptorField, error) {
	type globals struct {
		usagePage uint32
		reportID  byte
		size      int
		count     int
	}

	var (
		global      globals
		stack       []globals
		usages      []uint32
		collections []uint32
		kinds       []byte
		fields      []descriptorField
	)

	for i := 0; i < len(descriptor); {
		prefix := descriptor[i]
		if prefix == 0xfe {
			return nil, fmt.Errorf("long item at offset %d", i)
		}
		length := int(prefix & 0x03)
		if length == 3 {
			length = 4
		}
		if i+1+length > len(descriptor) {
			return nil, fmt.Errorf("the descriptor ends inside an item at offset %d", i)
		}

		var value uint32
		for b := 0; b < length; b++ {
			value |= uint32(descriptor[i+1+b]) << (8 * b)
		}

		usage := func() uint32 {
			if length == 4 {
				return value
			}
			return global.usagePage<<16 | value
		}

		tag := prefix &^ 0x03
		switch tag {
		// Global items.
		case 0x04:
			global.usagePage = value
		case 0x74:
			global.size = int(value)
		case 0x84:
			global.reportID = byte(value)
		case 0x94:
			global.count = int(value)
		case 0xa4:
			stack = append(stack, global)
		case 0xb4:
			if len(stack) == 0 {
				return nil, fmt.Errorf("pop without push at offset %d", i)
			}
			global = stack[len(stack)-1]
			stack = stack[:len(stack)-1]

		// Local items.
		case 0x08:
			usages = append(usages, usage())
		case 0x18, 0x28:
			// A usage range is recorded by its ends. Nothing here needs the
			// usages between them.
			usages = append(usages, usage())

		// Main items.
		case mainCollection:
			collection := uint32(0)
			if len(usages) > 0 {
				collection = usages[0]
			}
			collections = append(collections, collection)
			kinds = append(kinds, byte(value))
			usages = nil
		case mainEnd:
			if len(collections) == 0 {
				return nil, fmt.Errorf("end collection without a collection at offset %d", i)
			}
			collections = collections[:len(collections)-1]
			kinds = kinds[:len(kinds)-1]
			usages = nil
		case mainInput, mainOutput, mainFeature:
			application := uint32(0)
			for c := len(collections) - 1; c >= 0; c-- {
				if kinds[c] == 0x01 {
					application = collections[c]
					break
				}
			}
			fields = append(fields, descriptorField{
				main:        tag,
				reportID:    global.reportID,
				size:        global.size,
				count:       global.count,
				usages:      usages,
				application: application,
				collections: append([]uint32(nil), collections...),
			})
			usages = nil
		}

		i += 1 + length
	}

	if len(collections) != 0 {
		return nil, fmt.Errorf("%d collections are never closed", len(collections))
	}
	return fields, nil
}
