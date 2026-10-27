// Package slot puts the board's A/B slot tool behind the API: what each slot
// holds, an install with progress, and the trial, confirm and revert that
// decide which slot the board boots (ironkvm-dist#165).
//
// Every decision stays in the tool, /usr/bin/slot, which the distribution
// image carries. This package only runs it: it checks the request, starts one
// install at a time, and reads back the progress file the tool keeps. The tool
// is what refuses an image nobody signed, an image for another board and a
// write to the running slot, so a refusal here is a shortcut to the same
// answer and never a second definition of it.
package slot

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

// The paths are variables so the tests can point them at a stub tool and a
// directory of their own.
var (
	// Bin is the slot tool. A board that does not run the distribution image
	// has none, and the feature is then not offered (Available).
	Bin = "/usr/bin/slot"

	// ImageRoot is where an image already on the board may be installed
	// from: the data partition, which survives every slot write.
	ImageRoot = "/data"

	// ProgressFile, LogFile and JobFile describe the install that runs, or
	// ran last. They are in /tmp rather than in memory so that a server
	// restarted during an install still finds it: the install outlives the
	// server by design (see startInstall), and a reboot ends both.
	ProgressFile = "/tmp/ironkvm-slot-install.progress"
	LogFile      = "/tmp/ironkvm-slot-install.log"
	JobFile      = "/tmp/ironkvm-slot-install.job"
)

// How long the short commands may take. status reads a few small files.
// try can mount the slot to read its record, which is the slow one.
const (
	statusTimeout = 30 * time.Second
	tryTimeout    = 2 * time.Minute
	decideTimeout = 30 * time.Second
)

// Available reports whether the board has the slot tool. It is the probe of
// the "slot-install" feature.
func Available() bool {
	st, err := os.Stat(Bin)
	return err == nil && st.Mode().IsRegular() && st.Mode()&0o111 != 0
}

var errUnavailable = errors.New("this board has no slot tool; slot installs need the IronKVM distribution image")

// run runs the tool with args and returns what it printed. A failure carries
// the tool's own reason: its last "slot: " line, without the prefix.
func run(ctx context.Context, timeout time.Duration, args ...string) (string, error) {
	if !Available() {
		return "", errUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, Bin, args...)
	cmd.WaitDelay = time.Second
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return string(out), errors.New("slot " + args[0] + " did not finish in time")
		}
		if reason := lastReason(string(out)); reason != "" {
			return string(out), errors.New(reason)
		}
		return string(out), err
	}
	return string(out), nil
}

// lastReason is the last line the tool printed about itself, the one die()
// writes, without its "slot: " prefix.
func lastReason(out string) string {
	reason := ""
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); strings.HasPrefix(line, "slot: ") {
			reason = strings.TrimPrefix(line, "slot: ")
		}
	}
	return reason
}

// Info is one slot's record, as `slot status` prints it.
type Info struct {
	Name string `json:"name"`
	// State is installed, installing, interrupted, bad, or unknown for a slot
	// written before records existed.
	State     string `json:"state"`
	Version   string `json:"version,omitempty"`
	Sha256    string `json:"sha256,omitempty"`
	Installed string `json:"installed,omitempty"`
	// Trial is none, armed, booted, passed or failed.
	Trial string `json:"trial,omitempty"`
}

// Status is `slot status`, read into fields.
type Status struct {
	// Running is the slot this boot came from, and Trusted the one the board
	// boots when nothing else is asked for. They differ during a trial.
	Running string `json:"running"`
	Trusted string `json:"trusted"`
	// The lines the boot backend adds. On the SG2002 backend Trial is the
	// slot armed for the next boot or "none", Prev is the slot a revert
	// returns to or "none", and Recovery is "requested" or "no". A backend
	// without them leaves them empty.
	Trial    string `json:"trial,omitempty"`
	Prev     string `json:"prev,omitempty"`
	Recovery string `json:"recovery,omitempty"`
	// Bootloader and Kernel are there when the per-slot boot is supported.
	Bootloader string `json:"bootloader,omitempty"`
	Kernel     string `json:"kernel,omitempty"`
	Slots      []Info `json:"slots"`
	// Image is what the running slot's /etc/slot-manifest says of it.
	Image map[string]string `json:"image,omitempty"`
}

// Slot returns the record of the named slot.
func (s *Status) Slot(name string) (Info, bool) {
	for _, info := range s.Slots {
		if info.Name == name {
			return info, true
		}
	}
	return Info{}, false
}

// Installing reports the slot an install is writing, by the tool's records,
// which see an install started from a shell as well as one started here.
func (s *Status) Installing() string {
	for _, info := range s.Slots {
		if info.State == "installing" {
			return info.Name
		}
	}
	return ""
}

// imageFields are the lines of the running slot's manifest worth showing. The
// manifest also carries hashes and module lists that only a support bundle
// needs.
var imageFields = map[string]bool{"device": true, "version": true, "slotkernel": true, "flavour": true}

// parseStatus reads what `slot status` prints: "<word> <value>" lines, the
// slot-<name> lines with a state and key=value pairs, then a blank line and
// the running slot's manifest indented by two spaces.
func parseStatus(out string) Status {
	st := Status{Slots: []Info{}}
	inManifest := false
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			inManifest = true
			continue
		}
		fields := strings.Fields(line)
		if inManifest {
			if len(fields) >= 2 && imageFields[fields[0]] {
				if st.Image == nil {
					st.Image = map[string]string{}
				}
				if _, seen := st.Image[fields[0]]; !seen {
					st.Image[fields[0]] = strings.Join(fields[1:], " ")
				}
			}
			continue
		}
		if len(fields) < 2 {
			continue
		}
		value := strings.Join(fields[1:], " ")
		switch key := fields[0]; {
		case strings.HasPrefix(key, "slot-"):
			info := Info{Name: strings.TrimPrefix(key, "slot-"), State: fields[1]}
			for _, kv := range fields[2:] {
				k, v, ok := strings.Cut(kv, "=")
				if !ok {
					continue
				}
				switch k {
				case "version":
					info.Version = v
				case "sha256":
					info.Sha256 = v
				case "installed":
					info.Installed = v
				case "trial":
					info.Trial = v
				}
			}
			st.Slots = append(st.Slots, info)
		case key == "running":
			st.Running = value
		case key == "trusted":
			st.Trusted = value
		case key == "trial":
			st.Trial = value
		case key == "prev":
			st.Prev = value
		case key == "recovery":
			st.Recovery = value
		case key == "bootloader":
			st.Bootloader = value
		case key == "kernel":
			st.Kernel = value
		}
	}
	return st
}

// readStatus runs `slot status`. It also brings the records up to date: the
// tool turns an install whose process is gone into "interrupted" as it
// answers.
func readStatus(ctx context.Context) (Status, error) {
	out, err := run(ctx, statusTimeout, "status")
	if err != nil {
		return Status{}, err
	}
	return parseStatus(out), nil
}
