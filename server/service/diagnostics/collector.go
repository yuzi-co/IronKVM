// Package diagnostics builds the support bundle: one tar.gz with the versions,
// the slot, the kernel log, the server log, the configuration without its
// secrets, and the state of the network, the USB gadget and the video
// pipeline.
//
// Two rules hold for everything here. Nothing reads a hardware register: no
// devmem, nothing under /sys/kernel/debug but the ION dump service/ion
// already reads, no /proc/cvitek (reading /proc/cvitek/vb blocks forever), no
// thermal sensor and no call into libkvm. And every source has a timeout and
// a size limit, so a bundle cannot hang the board or fill its memory.
package diagnostics

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"NanoKVM-Server/config"
)

// The limits of a bundle.
var (
	// SourceTimeout bounds one source; a command that takes longer is
	// killed, and a file read that takes longer is abandoned. A variable so
	// tests can shorten it.
	SourceTimeout = 5 * time.Second
	// BundleTimeout bounds the whole bundle.
	BundleTimeout = 30 * time.Second
)

const (
	// LogLimit is how much of the end of a log the bundle keeps.
	LogLimit = 256 << 10
	// SourceLimit bounds every other source.
	SourceLimit = 128 << 10
	// maxProcesses bounds the process list.
	maxProcesses = 400
)

// ServerLogFile is where S95nanokvm sends the server's output.
const ServerLogFile = "/tmp/nanokvm-server.log"

// ErrBusy is returned while another bundle is being built.
var ErrBusy = errors.New("a diagnostics bundle is already being built")

// Collector builds the bundle. Root is prefixed to every path it reads, so a
// test can run it against a directory of its own; on the board it is empty.
type Collector struct {
	Root string
	// Run runs a command and returns its standard output. Tests replace it.
	Run func(ctx context.Context, name string, args ...string) ([]byte, error)
	Now func() time.Time

	// Versions names the versions only the server knows, such as the
	// application version and the build stamp.
	Versions func() map[string]string
	// Settings returns the settings export, built from its allowlist.
	Settings func() (any, error)
	// Video returns the state of the video pipeline, without a call into
	// libkvm.
	Video func() any

	busy atomic.Bool
}

// NewCollector returns a collector for the board.
func NewCollector() *Collector {
	return &Collector{Run: runCommand, Now: time.Now}
}

func runCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.WaitDelay = time.Second
	return cmd.Output()
}

func (c *Collector) path(p string) string {
	return filepath.Join(c.Root, p)
}

func (c *Collector) readFile(p string) ([]byte, error) {
	return os.ReadFile(c.path(p))
}

// file is one entry of the bundle.
type file struct {
	name    string
	collect func(ctx context.Context) (string, error)
}

func (c *Collector) files() []file {
	return []file{
		{"versions.txt", c.versions},
		{"slot.txt", c.slot},
		{"dmesg.txt", c.dmesg},
		{"server-log.txt", c.serverLog},
		{"config.json", c.config},
		{"network.txt", c.network},
		{"usb-gadget.txt", c.usbGadget},
		{"video.json", c.video},
		{"system.txt", c.system},
	}
}

// Write builds the bundle into w. A source that fails or times out leaves a
// line in errors.txt and the bundle goes on.
func (c *Collector) Write(ctx context.Context, w io.Writer) error {
	if !c.busy.CompareAndSwap(false, true) {
		return ErrBusy
	}
	defer c.busy.Store(false)

	ctx, cancel := context.WithTimeout(ctx, BundleTimeout)
	defer cancel()

	now := c.Now()
	redactor := NewRedactor(c.knownSecrets())

	gz := gzip.NewWriter(w)
	tw := tar.NewWriter(gz)
	add := func(name, content string) error {
		data := []byte(redactor.Redact(content))
		if err := tw.WriteHeader(&tar.Header{
			Name:    "ironkvm-diagnostics/" + name,
			Mode:    0o644,
			Size:    int64(len(data)),
			ModTime: now,
		}); err != nil {
			return err
		}
		_, err := tw.Write(data)
		return err
	}

	if err := add("README.txt", readme(now)); err != nil {
		return err
	}
	var failures []string
	for _, f := range c.files() {
		content, err := collectWithTimeout(ctx, f.collect)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %s", f.name, err))
		}
		if err := add(f.name, content); err != nil {
			return err
		}
	}
	if len(failures) == 0 {
		failures = []string{"none"}
	}
	if err := add("errors.txt", strings.Join(failures, "\n")+"\n"); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

// collectWithTimeout runs one source and waits at most SourceTimeout for it.
// A source that does not return in time is abandoned: a read from a sysfs file
// cannot be interrupted, and waiting for it would hang the bundle.
func collectWithTimeout(ctx context.Context, collect func(context.Context) (string, error)) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, SourceTimeout)
	defer cancel()

	type result struct {
		text string
		err  error
	}
	done := make(chan result, 1)
	go func() {
		text, err := collect(ctx)
		done <- result{text, err}
	}()

	select {
	case r := <-done:
		return r.text, r.err
	case <-ctx.Done():
		return "timed out\n", errors.New("timed out")
	}
}

func readme(now time.Time) string {
	return fmt.Sprintf(`IronKVM diagnostics bundle
Made %s.

Every file was passed through a redactor before it was written:
  - every secret the board keeps (passwords, keys, tokens, the JWT secret,
    notification addresses and tokens) is replaced with %s;
  - so is every string that has the shape of a private key, a JSON web token,
    a bearer token, a password in a URL, or a password=, token= or secret=
    assignment.
config.json is built from an allowlist and holds no file from /etc/kvm, only
their names, sizes and modes.

Limits: each source %s, the bundle %s, each log its last %d KiB.
Sources that failed or timed out are in errors.txt.
`, now.UTC().Format(time.RFC3339), Redacted, SourceTimeout, BundleTimeout, LogLimit>>10)
}

// tail returns the end of s, at most limit bytes, cut at a line start.
func tail(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	s = s[len(s)-limit:]
	if i := strings.IndexByte(s, '\n'); i >= 0 && i < len(s)-1 {
		s = s[i+1:]
	}
	return fmt.Sprintf("[earlier output cut; last %d KiB kept]\n", limit>>10) + s
}

// head returns the start of s, at most limit bytes.
func head(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit] + fmt.Sprintf("\n[output cut at %d KiB]\n", limit>>10)
}

// readTail reads the last limit bytes of a file without reading the rest.
func (c *Collector) readTail(p string, limit int64) (string, error) {
	f, err := os.Open(c.path(p))
	if err != nil {
		return "", err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return "", err
	}
	offset := int64(0)
	if info.Size() > limit {
		offset = info.Size() - limit
	}
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return "", err
	}
	data, err := io.ReadAll(io.LimitReader(f, limit))
	if err != nil {
		return "", err
	}
	text := string(data)
	if offset > 0 {
		// The first line is a fragment; start at the next one.
		if i := strings.IndexByte(text, '\n'); i >= 0 && i < len(text)-1 {
			text = text[i+1:]
		}
		text = fmt.Sprintf("[earlier output cut; last %d KiB kept]\n", limit>>10) + text
	}
	return text, nil
}

// section is one titled part of a text file.
type section struct {
	b strings.Builder
}

func (s *section) add(title, body string) {
	fmt.Fprintf(&s.b, "=== %s\n%s", title, body)
	if !strings.HasSuffix(body, "\n") {
		s.b.WriteByte('\n')
	}
	s.b.WriteByte('\n')
}

// addFile adds a small file, or says why it could not.
func (c *Collector) addFile(s *section, p string) {
	data, err := c.readFile(p)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		s.add(p, "(absent)")
	case err != nil:
		s.add(p, "(error: "+err.Error()+")")
	default:
		s.add(p, head(string(data), SourceLimit))
	}
}

func (c *Collector) addCommand(ctx context.Context, s *section, name string, args ...string) {
	title := strings.TrimSpace(name + " " + strings.Join(args, " "))
	out, err := c.Run(ctx, name, args...)
	if err != nil {
		s.add(title, head(string(out), SourceLimit)+"(error: "+err.Error()+")")
		return
	}
	s.add(title, head(string(out), SourceLimit))
}

func (c *Collector) versions(context.Context) (string, error) {
	var s section
	if c.Versions != nil {
		keys := []string{}
		values := c.Versions()
		for k := range values {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		for _, k := range keys {
			fmt.Fprintf(&b, "%s: %s\n", k, values[k])
		}
		s.add("server", b.String())
	}
	for _, p := range []string{"/kvmapp/version", "/boot/ver", "/boot/ironkvm.ver", "/etc/kvm/hw", "/proc/version"} {
		c.addFile(&s, p)
	}
	return s.b.String(), nil
}

func (c *Collector) slot(context.Context) (string, error) {
	var s section
	for _, p := range []string{"/boot/slot", "/boot/slot.try", "/boot/slot.prev", "/boot/slot.good", "/proc/cmdline"} {
		c.addFile(&s, p)
	}
	mounts, err := c.readFile("/proc/mounts")
	if err != nil {
		s.add("root mount", "(error: "+err.Error()+")")
	} else {
		var root []string
		for _, line := range strings.Split(string(mounts), "\n") {
			if fields := strings.Fields(line); len(fields) > 1 && (fields[1] == "/" || fields[1] == "/mnt/slota") {
				root = append(root, line)
			}
		}
		s.add("root mount", strings.Join(root, "\n"))
	}
	return s.b.String(), nil
}

func (c *Collector) dmesg(ctx context.Context) (string, error) {
	out, err := c.Run(ctx, "dmesg")
	text := tail(string(out), LogLimit)
	if err != nil {
		return text + "\n(error: " + err.Error() + ")\n", err
	}
	return text, nil
}

func (c *Collector) serverLog(context.Context) (string, error) {
	text, err := c.readTail(ServerLogFile, LogLimit)
	if err != nil {
		return "(error: " + err.Error() + ")\n", err
	}
	return text, nil
}

// networkConfig names the fields of server.yaml that say how the board is
// reached. None of them is a secret; the export leaves them out because an
// import must not change them, and support needs to read them.
type networkConfig struct {
	Proto           string   `json:"proto"`
	Host            string   `json:"host"`
	HTTPPort        int      `json:"httpPort"`
	HTTPSPort       int      `json:"httpsPort"`
	HTTP2           bool     `json:"http2"`
	Authentication  string   `json:"authentication"`
	LoggerFile      string   `json:"loggerFile"`
	AllowedNetworks []string `json:"allowedNetworks"`
	ProxySet        bool     `json:"proxySet"`
	TurnSet         bool     `json:"turnSet"`
}

// etcEntry describes a file in /etc/kvm without its content.
type etcEntry struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Mode string `json:"mode"`
}

func (c *Collector) config(context.Context) (string, error) {
	out := map[string]any{}
	var errs []string

	if c.Settings != nil {
		settings, err := c.Settings()
		if err != nil {
			errs = append(errs, "settings: "+err.Error())
		} else {
			out["settings"] = settings
		}
	}

	if data, err := c.readFile(config.ConfigurationFile); err != nil {
		errs = append(errs, "server.yaml: "+err.Error())
	} else {
		var conf config.Config
		if err := yaml.Unmarshal(data, &conf); err != nil {
			errs = append(errs, "server.yaml is not valid YAML")
		} else {
			out["server"] = networkConfig{
				Proto:           conf.Proto,
				Host:            conf.Host,
				HTTPPort:        conf.Port.Http,
				HTTPSPort:       conf.Port.Https,
				HTTP2:           conf.HTTP2,
				Authentication:  conf.Authentication,
				LoggerFile:      conf.Logger.File,
				AllowedNetworks: conf.Security.AllowedNetworks,
				ProxySet:        conf.Proxy != "",
				TurnSet:         conf.Turn.TurnAddr != "",
			}
		}
	}

	var entries []etcEntry
	root := c.path("/etc/kvm")
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || p == root {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		if strings.Count(rel, string(filepath.Separator)) > 1 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		entries = append(entries, etcEntry{Name: filepath.ToSlash(rel), Size: info.Size(), Mode: info.Mode().String()})
		return nil
	})
	out["etcKvm"] = entries

	if len(errs) > 0 {
		out["errors"] = errs
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

func (c *Collector) network(ctx context.Context) (string, error) {
	var s section
	c.addCommand(ctx, &s, "ip", "addr")
	c.addCommand(ctx, &s, "ip", "route")
	c.addCommand(ctx, &s, "ip", "-6", "route")
	c.addFile(&s, "/etc/resolv.conf")
	c.addFile(&s, "/proc/net/dev")
	return s.b.String(), nil
}

// usbGadget describes the configfs gadgets and the device controllers. It
// reads names, links and the few attributes the kernel keeps in memory.
func (c *Collector) usbGadget(context.Context) (string, error) {
	var s section

	gadgets, err := os.ReadDir(c.path("/sys/kernel/config/usb_gadget"))
	if err != nil {
		s.add("/sys/kernel/config/usb_gadget", "(error: "+err.Error()+")")
	}
	for _, g := range gadgets {
		base := "/sys/kernel/config/usb_gadget/" + g.Name()
		var b strings.Builder
		for _, attr := range []string{"UDC", "idVendor", "idProduct", "bcdDevice"} {
			data, err := c.readFile(base + "/" + attr)
			if err == nil {
				fmt.Fprintf(&b, "%s: %s\n", attr, strings.TrimSpace(string(data)))
			}
		}
		if functions, err := os.ReadDir(c.path(base + "/functions")); err == nil {
			names := make([]string, 0, len(functions))
			for _, f := range functions {
				names = append(names, f.Name())
			}
			fmt.Fprintf(&b, "functions: %s\n", strings.Join(names, " "))
		}
		if configs, err := os.ReadDir(c.path(base + "/configs")); err == nil {
			for _, cfg := range configs {
				links, err := os.ReadDir(c.path(base + "/configs/" + cfg.Name()))
				if err != nil {
					continue
				}
				var linked []string
				for _, l := range links {
					if l.Type()&fs.ModeSymlink != 0 {
						target, _ := os.Readlink(c.path(base + "/configs/" + cfg.Name() + "/" + l.Name()))
						linked = append(linked, l.Name()+" -> "+filepath.Base(target))
					}
				}
				fmt.Fprintf(&b, "config %s: %s\n", cfg.Name(), strings.Join(linked, ", "))
			}
		}
		s.add("gadget "+g.Name(), b.String())
	}

	udcs, err := os.ReadDir(c.path("/sys/class/udc"))
	if err != nil {
		s.add("/sys/class/udc", "(error: "+err.Error()+")")
	}
	for _, u := range udcs {
		var b strings.Builder
		for _, attr := range []string{"state", "current_speed"} {
			data, err := c.readFile("/sys/class/udc/" + u.Name() + "/" + attr)
			if err == nil {
				fmt.Fprintf(&b, "%s: %s\n", attr, strings.TrimSpace(string(data)))
			}
		}
		s.add("udc "+u.Name(), b.String())
	}
	return s.b.String(), nil
}

func (c *Collector) video(context.Context) (string, error) {
	if c.Video == nil {
		return "{}\n", nil
	}
	data, err := json.MarshalIndent(c.Video(), "", "  ")
	if err != nil {
		return "", err
	}
	return head(string(data), SourceLimit) + "\n", nil
}

func (c *Collector) system(context.Context) (string, error) {
	var s section
	for _, p := range []string{"/proc/uptime", "/proc/loadavg", "/proc/meminfo", "/proc/mounts"} {
		c.addFile(&s, p)
	}

	var b strings.Builder
	for _, mount := range []string{"/", "/data", "/tmp", "/boot"} {
		var st syscall.Statfs_t
		if err := syscall.Statfs(c.path(mount), &st); err != nil {
			fmt.Fprintf(&b, "%s: %s\n", mount, err)
			continue
		}
		bs := uint64(st.Bsize)
		fmt.Fprintf(&b, "%s: %d KiB free of %d KiB\n", mount, st.Bavail*bs>>10, st.Blocks*bs>>10)
	}
	s.add("space", b.String())

	s.add("processes", c.processes())
	return s.b.String(), nil
}

// processes lists each process with its name, state and resident memory,
// from /proc/<pid>/stat and /proc/<pid>/status. It never reads cmdline: a
// command line can carry a key.
func (c *Collector) processes() string {
	entries, err := os.ReadDir(c.path("/proc"))
	if err != nil {
		return "(error: " + err.Error() + ")"
	}
	var b strings.Builder
	b.WriteString("PID STATE RSS_KB NAME\n")
	count := 0
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		if count++; count > maxProcesses {
			b.WriteString("[more processes left out]\n")
			break
		}
		stat, err := c.readFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err != nil {
			continue
		}
		name, state := parseStat(string(stat))
		rss := ""
		if status, err := c.readFile(fmt.Sprintf("/proc/%d/status", pid)); err == nil {
			for _, line := range strings.Split(string(status), "\n") {
				if strings.HasPrefix(line, "VmRSS:") {
					rss = strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, "VmRSS:")), " kB")
				}
			}
		}
		if rss == "" {
			rss = "-"
		}
		fmt.Fprintf(&b, "%d %s %s %s\n", pid, state, rss, name)
	}
	return b.String()
}

// parseStat returns the name and state from /proc/<pid>/stat. The name is in
// parentheses and may hold spaces and parentheses itself.
func parseStat(stat string) (name, state string) {
	open := strings.IndexByte(stat, '(')
	end := strings.LastIndexByte(stat, ')')
	if open < 0 || end < open {
		return "?", "?"
	}
	name = stat[open+1 : end]
	rest := strings.Fields(stat[end+1:])
	state = "?"
	if len(rest) > 0 {
		state = rest[0]
	}
	return name, state
}
