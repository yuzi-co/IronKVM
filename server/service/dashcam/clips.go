package dashcam

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Dir holds the clips. /data is the card's exFAT partition: a clip is written
// once, when it is saved, and nothing is written there while the dashcam only
// keeps the stream in memory.
const Dir = "/data/dashcam"

// nameLayout is the time at the front of a clip's name. It sorts in time
// order, and exFAT takes every character of it.
const nameLayout = "20060102-150405"

// clipName is what a clip's name may be: the time, the triggers joined by
// "+", an optional part number, and ".mp4". The routes take nothing else, so
// a name never reaches the file system with a separator or "..".
var clipName = regexp.MustCompile(`^[0-9]{8}-[0-9]{6}-[a-z0-9+-]{1,96}\.mp4$`)

// ValidName reports whether name is one a clip may have.
func ValidName(name string) bool {
	return clipName.MatchString(name)
}

// Clip is a saved clip, for the page.
type Clip struct {
	Name string    `json:"name"`
	Size int64     `json:"size"`
	Time time.Time `json:"time"`
	// Triggers are what saved it, from its name.
	Triggers []string `json:"triggers"`
}

// triggersOf reads the triggers back from a clip's name.
func triggersOf(name string) []string {
	rest := strings.TrimSuffix(name, ".mp4")
	if len(rest) <= len(nameLayout)+1 {
		return nil
	}
	rest = rest[len(nameLayout)+1:]
	if i := strings.LastIndex(rest, "-part"); i >= 0 {
		rest = rest[:i]
	}
	return strings.Split(rest, "+")
}

// listClips lists the clips in dir, newest first. A missing directory has
// none.
func listClips(dir string) ([]Clip, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return []Clip{}, nil
	}
	if err != nil {
		return nil, err
	}
	clips := make([]Clip, 0, len(entries))
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !ValidName(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		clips = append(clips, Clip{
			Name:     entry.Name(),
			Size:     info.Size(),
			Time:     info.ModTime(),
			Triggers: triggersOf(entry.Name()),
		})
	}
	// The name starts with the time the clip was asked for, which the
	// modification time of a file copied in does not keep.
	sort.Slice(clips, func(i, j int) bool { return clips[i].Name > clips[j].Name })
	return clips, nil
}

// prune deletes the oldest clips until the rest fit in limit bytes. The
// clips named in keep, the ones just written, stay whatever their size.
func prune(dir string, limit int64, keep []string) ([]string, error) {
	clips, err := listClips(dir)
	if err != nil {
		return nil, err
	}
	var total int64
	for _, c := range clips {
		total += c.Size
	}
	var removed []string
	for i := len(clips) - 1; i >= 0 && total > limit; i-- {
		c := clips[i]
		if slices.Contains(keep, c.Name) {
			continue
		}
		if err := os.Remove(filepath.Join(dir, c.Name)); err != nil {
			log.Warnf("dashcam: delete %s: %s", c.Name, err)
			continue
		}
		total -= c.Size
		removed = append(removed, c.Name)
	}
	return removed, nil
}

// removeTemporary deletes what a write cut short left behind.
func removeTemporary(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") && strings.HasSuffix(entry.Name(), ".tmp") {
			_ = os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
}

// clipBase is a clip's name before the part number and ".mp4".
func clipBase(at time.Time, triggers []string) string {
	return at.Format(nameLayout) + "-" + strings.Join(triggers, "+")
}

// writeClip writes the samples as one file per segment and answers their
// names. Each file is written beside its final name and renamed into place,
// so the list never shows half a clip.
func writeClip(dir string, base string, samples []Sample) ([]string, error) {
	segs := split(samples)
	if len(segs) == 0 {
		return nil, errNoParameterSets
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	var names []string
	for i, seg := range segs {
		name := base + ".mp4"
		if len(segs) > 1 {
			name = fmt.Sprintf("%s-part%d.mp4", base, i+1)
		}
		name = unusedName(dir, name)
		if err := writeFile(dir, name, seg); err != nil {
			return names, err
		}
		names = append(names, name)
	}
	return names, nil
}

// unusedName adds a number to a name that is taken. Two clips start in the
// same second only when the second is asked for by hand right after a clip
// ended.
func unusedName(dir, name string) string {
	stem := strings.TrimSuffix(name, ".mp4")
	for n := 2; ; n++ {
		if _, err := os.Lstat(filepath.Join(dir, name)); errors.Is(err, fs.ErrNotExist) {
			return name
		}
		name = fmt.Sprintf("%s-%d.mp4", stem, n)
	}
}

func writeFile(dir, name string, seg segment) (err error) {
	tmp := filepath.Join(dir, "."+name+".tmp")
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
		}
	}()

	// One large buffer keeps the card's writes few and whole.
	w := bufio.NewWriterSize(f, 256<<10)
	if err = writeMP4(w, seg); err != nil {
		return err
	}
	if err = w.Flush(); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, name))
}
