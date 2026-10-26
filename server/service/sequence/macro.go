package sequence

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/google/uuid"

	"NanoKVM-Server/proto"
)

// macroFile keeps the macros beside the keyboard shortcuts, so they follow
// the device rather than the browser. A variable so the tests can point it
// somewhere writable.
var macroFile = "/etc/kvm/macros.json"

const (
	// maxMacros bounds the file. At the step limit each macro is some 100 KB
	// of JSON, and the file is read whole.
	maxMacros  = 50
	maxNameLen = 64
)

var (
	macroMutex sync.RWMutex

	errMacroNotFound = errors.New("macro not found")
	errTooManyMacros = fmt.Errorf("at most %d macros can be saved", maxMacros)
)

type macroStore struct {
	Macros []proto.Macro `json:"macros"`
}

func loadMacros() (*macroStore, error) {
	data, err := os.ReadFile(macroFile)
	if errors.Is(err, os.ErrNotExist) || (err == nil && len(data) == 0) {
		return &macroStore{Macros: []proto.Macro{}}, nil
	}
	if err != nil {
		return nil, err
	}

	var store macroStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	if store.Macros == nil {
		store.Macros = []proto.Macro{}
	}
	return &store, nil
}

// saveMacros writes the file through a rename, so a power cut during the
// write leaves the old file rather than half of the new one.
func saveMacros(store *macroStore) error {
	data, err := json.Marshal(store)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(macroFile), ".macros-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), macroFile)
}

func listMacros() ([]proto.Macro, error) {
	macroMutex.RLock()
	defer macroMutex.RUnlock()

	store, err := loadMacros()
	if err != nil {
		return nil, err
	}
	return store.Macros, nil
}

func getMacro(id string) (proto.Macro, error) {
	macros, err := listMacros()
	if err != nil {
		return proto.Macro{}, err
	}
	for _, macro := range macros {
		if macro.ID == id {
			return macro, nil
		}
	}
	return proto.Macro{}, errMacroNotFound
}

// putMacro adds macro, or replaces the one with its ID. The steps are checked
// before anything is written, so the file never holds a macro that cannot be
// played.
func putMacro(macro proto.Macro) (proto.Macro, error) {
	macro.Name = truncate(strings.TrimSpace(macro.Name), maxNameLen)
	if macro.Name == "" {
		return proto.Macro{}, fmt.Errorf("the macro has no name")
	}
	if _, err := compileSteps(macro.Steps); err != nil {
		return proto.Macro{}, err
	}

	macroMutex.Lock()
	defer macroMutex.Unlock()

	store, err := loadMacros()
	if err != nil {
		return proto.Macro{}, err
	}

	if macro.ID != "" {
		for i := range store.Macros {
			if store.Macros[i].ID == macro.ID {
				store.Macros[i] = macro
				return macro, saveMacros(store)
			}
		}
		return proto.Macro{}, errMacroNotFound
	}

	if len(store.Macros) >= maxMacros {
		return proto.Macro{}, errTooManyMacros
	}
	macro.ID = uuid.New().String()
	store.Macros = append(store.Macros, macro)
	return macro, saveMacros(store)
}

func deleteMacro(id string) error {
	macroMutex.Lock()
	defer macroMutex.Unlock()

	store, err := loadMacros()
	if err != nil {
		return err
	}
	for i := range store.Macros {
		if store.Macros[i].ID == id {
			store.Macros = append(store.Macros[:i], store.Macros[i+1:]...)
			return saveMacros(store)
		}
	}
	return errMacroNotFound
}

// truncate cuts s to at most n runes.
func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

// ValidateMacros checks a whole list of macros the way putMacro checks one,
// for a list that arrives in one piece, such as a settings import. Every macro
// needs an ID, unique in the list, a name and steps that compile.
func ValidateMacros(macros []proto.Macro) error {
	if len(macros) > maxMacros {
		return errTooManyMacros
	}
	seen := make(map[string]bool, len(macros))
	for _, macro := range macros {
		if macro.ID == "" || len(macro.ID) > maxNameLen {
			return fmt.Errorf("a macro has no valid ID")
		}
		if seen[macro.ID] {
			return fmt.Errorf("two macros have the ID %q", macro.ID)
		}
		seen[macro.ID] = true
		name := strings.TrimSpace(macro.Name)
		if name == "" || utf8.RuneCountInString(name) > maxNameLen {
			return fmt.Errorf("the macro %q needs a name of 1 to %d characters", macro.ID, maxNameLen)
		}
		if _, err := compileSteps(macro.Steps); err != nil {
			return fmt.Errorf("the macro %q: %w", name, err)
		}
	}
	return nil
}
