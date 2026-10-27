package settingsio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// section is one entry of the allowlist: a named group of settings, how to
// read it from the board, how to check an imported value, and how to write it.
type section interface {
	name() string
	// current reads the value from the board.
	current(s *snapshot) (any, error)
	// decode decodes an imported value over the current one, so a field the
	// file does not carry keeps its value, and checks it on its own. The
	// warnings name what the import ignores.
	decode(s *snapshot, raw json.RawMessage) (value any, warnings []string, err error)
	// stage adds the writes for value, as decode returned it, to the plan.
	stage(p *plan, value any) error
}

// spec is a section whose value is T. T is the allowlist: only its fields
// are read, exported and written.
type spec[T any] struct {
	id string
	// get reads the value from the board.
	get func(s *snapshot) (T, error)
	// check validates v and may normalise it in place.
	check func(s *snapshot, v *T) error
	// put stages the writes for v.
	put func(p *plan, v T) error
}

func (sp spec[T]) name() string { return sp.id }

func (sp spec[T]) current(s *snapshot) (any, error) {
	return sp.get(s)
}

func (sp spec[T]) decode(s *snapshot, raw json.RawMessage) (any, []string, error) {
	var warnings []string

	value, err := sp.get(s)
	if err != nil {
		// A damaged file must not stop the import that would repair it. The
		// fields the file does not carry then start from zero.
		var zero T
		value = zero
		warnings = append(warnings, "the current value could not be read: "+err.Error())
	}

	unknown, err := unknownFields(raw, value)
	if err != nil {
		return nil, warnings, err
	}
	for _, field := range unknown {
		warnings = append(warnings, fmt.Sprintf("unknown field %q ignored", field))
	}

	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, warnings, fmt.Errorf("a value has the wrong type: %s", typeError(err))
	}
	if sp.check != nil {
		if err := sp.check(s, &value); err != nil {
			return nil, warnings, err
		}
	}
	return value, warnings, nil
}

func (sp spec[T]) stage(p *plan, value any) error {
	v, ok := value.(T)
	if !ok {
		return fmt.Errorf("section %s: unexpected value", sp.id)
	}
	return sp.put(p, v)
}

// unknownFields returns the keys of raw that v's type does not have, sorted.
// raw must be a JSON object.
func unknownFields(raw json.RawMessage, v any) ([]string, error) {
	var keys map[string]json.RawMessage
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&keys); err != nil || keys == nil {
		return nil, errors.New("the section is not a JSON object")
	}

	known := map[string]bool{}
	t := reflect.TypeOf(v)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t != nil && t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			tag := strings.Split(t.Field(i).Tag.Get("json"), ",")[0]
			if tag != "" && tag != "-" {
				known[tag] = true
			}
		}
	}

	var unknown []string
	for key := range keys {
		if !known[key] {
			unknown = append(unknown, key)
		}
	}
	sort.Strings(unknown)
	return unknown, nil
}

// typeError says which field had the wrong type without repeating the value.
func typeError(err error) string {
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) && typeErr.Field != "" {
		return fmt.Sprintf("%s should be %s", typeErr.Field, typeErr.Type)
	}
	return "not valid JSON"
}

// sections is the allowlist, in the order the preview lists it.
var sections = []section{
	loggingSection,
	loginSecuritySection,
	sessionsSection,
	powerLEDSection,
	redfishSection,
	ipmiSection,
	watchdogSection,
	nutSection,
	netbootSection,
	vncSection,
	webrtcSection,
	rtspSection,
	srtSection,
	picoclawSection,
	hostLogSection,
	webTitleSection,
	leaderKeySection,
	keyboardLayoutSection,
	shortcutsSection,
	macrosSection,
	mouseJigglerSection,
	keyJigglerSection,
	oledSection,
	hdmiSection,
	roomMicSection,
	notificationsSection,
}

func sectionByName(name string) (section, bool) {
	for _, sec := range sections {
		if sec.name() == name {
			return sec, true
		}
	}
	return nil, false
}

// SectionNames returns the names of the sections, in order.
func SectionNames() []string {
	names := make([]string, len(sections))
	for i, sec := range sections {
		names[i] = sec.name()
	}
	return names
}
