package settingsio

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"NanoKVM-Server/config"
)

// Device is the board whose settings are read and written. Root is prefixed to
// every path, so a test can run against a directory of its own; on the board
// it is empty.
type Device struct {
	Root string
}

func (d *Device) path(p string) string {
	return filepath.Join(d.Root, p)
}

// snapshot reads the board once per operation, so every section of one export
// or import sees the same server.yaml.
type snapshot struct {
	dev *Device

	conf       *config.Config
	confErr    error
	confLoaded bool
}

func (d *Device) snapshot() *snapshot {
	return &snapshot{dev: d}
}

// config returns server.yaml as the board keeps it, without the defaults the
// running server fills in.
func (s *snapshot) config() (*config.Config, error) {
	if !s.confLoaded {
		s.confLoaded = true
		data, err := os.ReadFile(s.dev.path(config.ConfigurationFile))
		if err != nil {
			s.confErr = err
		} else {
			var conf config.Config
			if err := yaml.Unmarshal(data, &conf); err != nil {
				s.confErr = errors.New("server.yaml is not valid YAML")
			} else {
				s.conf = &conf
			}
		}
	}
	return s.conf, s.confErr
}

// readFile returns a file's content, and false without an error when it is
// absent.
func (s *snapshot) readFile(p string) ([]byte, bool, error) {
	data, err := os.ReadFile(s.dev.path(p))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

func (s *snapshot) exists(p string) bool {
	_, err := os.Stat(s.dev.path(p))
	return err == nil
}

// fileChange is what an import does to one file: write data with mode, or
// remove the file.
type fileChange struct {
	data   []byte
	mode   os.FileMode
	remove bool
}

// plan collects the changes of an import before anything is written.
type plan struct {
	snap  *snapshot
	conf  *config.Config
	files map[string]fileChange
}

func newPlan(s *snapshot) *plan {
	return &plan{snap: s, files: map[string]fileChange{}}
}

// config returns the copy of server.yaml the import changes. The first call
// makes the copy.
func (p *plan) config() (*config.Config, error) {
	if p.conf == nil {
		conf, err := p.snap.config()
		if err != nil {
			return nil, err
		}
		copied := *conf
		p.conf = &copied
	}
	return p.conf, nil
}

func (p *plan) write(path string, data []byte, mode os.FileMode) {
	p.files[path] = fileChange{data: data, mode: mode}
}

func (p *plan) remove(path string) {
	p.files[path] = fileChange{remove: true}
}

// changes returns every file the plan writes, server.yaml included, keyed by
// its path on the board.
func (p *plan) changes() (map[string]fileChange, error) {
	out := make(map[string]fileChange, len(p.files)+1)
	for path, change := range p.files {
		out[path] = change
	}
	if p.conf != nil {
		data, err := yaml.Marshal(p.conf)
		if err != nil {
			return nil, err
		}
		mode := os.FileMode(0o644)
		if info, err := os.Stat(p.snap.dev.path(config.ConfigurationFile)); err == nil {
			mode = info.Mode().Perm()
		}
		out[config.ConfigurationFile] = fileChange{data: data, mode: mode}
	}
	return out, nil
}

// commit writes the changes so that either all of them land or none does, as
// far as a file system without transactions allows. Each new file is written
// to a temporary file in its own directory and synced first; only then are
// they renamed into place. A failed rename puts back the files already
// renamed, from copies read before the first one.
func (d *Device) commit(changes map[string]fileChange) error {
	paths := make([]string, 0, len(changes))
	for path := range changes {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	type staged struct {
		path      string
		tmp       string
		change    fileChange
		old       []byte
		oldMode   os.FileMode
		oldExists bool
	}
	steps := make([]staged, 0, len(paths))
	cleanup := func() {
		for _, st := range steps {
			if st.tmp != "" {
				_ = os.Remove(st.tmp)
			}
		}
	}

	for _, rel := range paths {
		st := staged{path: d.path(rel), change: changes[rel]}
		if info, err := os.Stat(st.path); err == nil {
			old, err := os.ReadFile(st.path)
			if err != nil {
				cleanup()
				return err
			}
			st.old, st.oldMode, st.oldExists = old, info.Mode().Perm(), true
		}
		if !st.change.remove {
			tmp, err := writeTemp(st.path, st.change.data, st.change.mode)
			if err != nil {
				cleanup()
				return err
			}
			st.tmp = tmp
		}
		steps = append(steps, st)
	}

	for i, st := range steps {
		var err error
		if st.change.remove {
			err = os.Remove(st.path)
			if errors.Is(err, fs.ErrNotExist) {
				err = nil
			}
		} else {
			err = rename(st.tmp, st.path)
			if err == nil {
				steps[i].tmp = ""
			}
		}
		if err != nil {
			for _, done := range steps[:i] {
				restore(done.path, done.old, done.oldMode, done.oldExists)
			}
			cleanup()
			return err
		}
	}

	dirs := map[string]bool{}
	for _, st := range steps {
		dirs[filepath.Dir(st.path)] = true
	}
	for dir := range dirs {
		syncDir(dir)
	}
	return nil
}

// rename moves a written file into place. Tests replace it to fail part way.
var rename = os.Rename

// writeTemp writes data beside path and syncs it, and returns the temporary
// file's name.
func writeTemp(path string, data []byte, mode os.FileMode) (string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(dir, ".settings-import-*")
	if err != nil {
		return "", err
	}
	if err = tmp.Chmod(mode); err == nil {
		_, err = tmp.Write(data)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}

// restore puts a file back as it was before the import, as well as it can.
func restore(path string, data []byte, mode os.FileMode, existed bool) {
	if !existed {
		_ = os.Remove(path)
		return
	}
	tmp, err := writeTemp(path, data, mode)
	if err != nil {
		return
	}
	if os.Rename(tmp, path) != nil {
		_ = os.Remove(tmp)
	}
}

func syncDir(dir string) {
	f, err := os.Open(dir)
	if err != nil {
		return
	}
	_ = f.Sync()
	_ = f.Close()
}
