package hostlog

import (
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/events"
)

func TestSyslog3164(t *testing.T) {
	p := parseSyslog([]byte("<34>Oct 11 22:14:15 mymachine su[123]: 'su root' failed for lonvick\n"))
	if p.facility != 4 || p.severity != 2 || p.host != "mymachine" || p.tag != "su" || p.msg != "'su root' failed for lonvick" {
		t.Errorf("%+v", p)
	}

	// A day below ten is padded with a space.
	p = parseSyslog([]byte("<13>Oct  1 01:02:03 box kernel: hello"))
	if p.host != "box" || p.tag != "kernel" || p.msg != "hello" {
		t.Errorf("%+v", p)
	}
}

func TestSyslog5424(t *testing.T) {
	p := parseSyslog([]byte(`<165>1 2026-10-11T22:14:15.003Z mymachine.example.com evntslog 1234 ID47 [exampleSDID@32473 iut="3" eventSource="App \] x" eventID="1011"][b@1 k="v"] An application event`))
	if p.facility != 20 || p.severity != 5 || p.host != "mymachine.example.com" || p.tag != "evntslog" || p.msg != "An application event" {
		t.Errorf("%+v", p)
	}

	p = parseSyslog([]byte("<34>1 2026-10-11T22:14:15Z host app - - - \xef\xbb\xbfBOM message"))
	if p.msg != "BOM message" || p.host != "host" {
		t.Errorf("%+v", p)
	}

	p = parseSyslog([]byte("<34>1 2026-10-11T22:14:15Z - - - - - nilvalues"))
	if p.host != "" || p.tag != "" || p.msg != "nilvalues" {
		t.Errorf("%+v", p)
	}
}

func TestSyslogOfNeitherFormIsKeptWhole(t *testing.T) {
	p := parseSyslog([]byte("just some text: with a colon later"))
	if p.severity != -1 || p.facility != -1 || p.msg != "just some text: with a colon later" || p.tag != "" {
		t.Errorf("%+v", p)
	}

	// A broken PRI is not one.
	p = parseSyslog([]byte("<999>text"))
	if p.severity != -1 || p.msg != "<999>text" {
		t.Errorf("%+v", p)
	}
}

func TestNetconsolePlainSplitsLines(t *testing.T) {
	out := parseNetconsole([]byte("<6>[   12.3] first\n[   12.4] second\r\n\n"))
	if len(out) != 2 || out[0].severity != 6 || out[0].msg != "[   12.3] first" || out[1].severity != -1 || out[1].msg != "[   12.4] second" {
		t.Errorf("%+v", out)
	}
}

func TestNetconsoleExtendedCarriesTheLevel(t *testing.T) {
	out := parseNetconsole([]byte("6,339,5140900,-;NET: Registered protocol family 10\n SUBSYSTEM=net\n DEVICE=+x"))
	if len(out) != 1 || out[0].facility != 0 || out[0].severity != 6 || out[0].msg != "NET: Registered protocol family 10" {
		t.Errorf("%+v", out)
	}
	// Level 3 of facility 0 is an error; level 4 is facility 0, warning.
	out = parseNetconsole([]byte("3,1,2,c,more=x;oops"))
	if len(out) != 1 || out[0].severity != 3 || out[0].msg != "oops" {
		t.Errorf("%+v", out)
	}
}

func TestControlCharactersAndLongLinesAreCleaned(t *testing.T) {
	p := parseSyslog([]byte("<13>x: a\x1b[31mred\x07\tb\xff"))
	if strings.ContainsAny(p.msg, "\x1b\x07\t") || !strings.Contains(p.msg, "red") {
		t.Errorf("%q", p.msg)
	}
	long := parseSyslog([]byte("<13>" + strings.Repeat("é", 5000)))
	if len(long.msg) > maxMessage+3 {
		t.Errorf("%d bytes", len(long.msg))
	}
}

func e(msg string) Entry {
	return Entry{Time: time.Unix(0, 0), Source: "10.0.0.2", Proto: ProtoSyslog, Severity: 6, Msg: msg}
}

func TestTheRingNeverHoldsMoreThanItsCap(t *testing.T) {
	r := NewRing(4000)
	for i := 0; i < 1000; i++ {
		r.Add(e(strings.Repeat("x", 50)))
		if _, b := r.Stats(); b > 4000 {
			t.Fatalf("%d bytes after %d lines", b, i+1)
		}
	}
	n, b := r.Stats()
	if n == 0 || n > 40 || b > 4000 {
		t.Fatalf("%d entries, %d bytes", n, b)
	}
	got := r.Query(Filter{MaxSeverity: -1}, 0)
	if len(got) != n || got[len(got)-1].Seq != 1000 {
		t.Fatalf("newest is %d, %d entries", got[len(got)-1].Seq, len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i].Seq != got[i-1].Seq+1 {
			t.Fatal("entries are not consecutive")
		}
	}

	r.SetCap(500)
	if _, b := r.Stats(); b > 500 {
		t.Errorf("%d bytes after the cap shrank", b)
	}
}

func TestQueryFilters(t *testing.T) {
	r := NewRing(1 << 20)
	r.Add(Entry{Source: "a", Severity: 6, Msg: "Kernel panic - not syncing"})
	r.Add(Entry{Source: "b", Severity: 2, Tag: "sshd", Msg: "Failed password for root"})
	r.Add(Entry{Source: "a", Severity: -1, Msg: "no severity here"})
	r.Add(Entry{Source: "a", Severity: 7, Host: "Box", Msg: "debug noise"})

	count := func(f Filter) int { return len(r.Query(f, 0)) }
	if n := count(Filter{MaxSeverity: -1, Words: ParseWords("PANIC kernel")}); n != 1 {
		t.Errorf("two words in any case: %d", n)
	}
	if n := count(Filter{MaxSeverity: -1, Words: ParseWords("sshd failed")}); n != 1 {
		t.Errorf("the tag is searched: %d", n)
	}
	if n := count(Filter{MaxSeverity: -1, Words: ParseWords("box")}); n != 1 {
		t.Errorf("the host is searched: %d", n)
	}
	if n := count(Filter{MaxSeverity: 3}); n != 2 {
		t.Errorf("severity 3 or worse keeps the error and the unknown: %d", n)
	}
	if n := count(Filter{MaxSeverity: -1, Source: "b"}); n != 1 {
		t.Errorf("source: %d", n)
	}
	if n := count(Filter{MaxSeverity: -1, After: 2}); n != 2 {
		t.Errorf("after: %d", n)
	}
	if got := r.Query(Filter{MaxSeverity: -1}, 2); len(got) != 2 || got[1].Seq != 4 {
		t.Errorf("limit keeps the newest: %+v", got)
	}
}

func TestSubscribersGetNewLinesAndSlowOnesLoseThemQuietly(t *testing.T) {
	r := NewRing(1 << 20)
	ch, cancel := r.Subscribe()
	defer cancel()

	r.Add(e("one"))
	if got := <-ch; got.Msg != "one" {
		t.Errorf("%+v", got)
	}
	for i := 0; i < subscriberBuffer*2; i++ {
		r.Add(e("flood")) // must not block
	}
	cancel()
	r.Add(e("after"))
}

func TestTheDiskLogRotatesAndStaysUnderItsCap(t *testing.T) {
	dir := t.TempDir()
	d := NewDisk(dir, 4000)
	for i := 0; i < 300; i++ {
		en := e(strings.Repeat("y", 60))
		en.Seq = uint64(i + 1)
		if err := d.Append(en); err != nil {
			t.Fatal(err)
		}
	}
	if size := d.Size(); size > 4000+300 {
		t.Errorf("the log is %d bytes, cap 4000", size)
	}
	all := d.ReadAll()
	if len(all) == 0 || all[len(all)-1].Seq != 300 {
		t.Fatalf("%d entries read back", len(all))
	}
	for i := 1; i < len(all); i++ {
		if all[i].Seq != all[i-1].Seq+1 {
			t.Fatal("entries out of order across the two files")
		}
	}
	if st, err := os.Stat(filepath.Join(dir, diskCurrent)); err != nil || st.Mode().Perm()&0o077 != 0 {
		t.Errorf("the log file is not private: %v %v", st, err)
	}

	if err := d.Clear(); err != nil || d.Size() != 0 || len(d.ReadAll()) != 0 {
		t.Errorf("clear: %v", err)
	}
	if err := d.Append(e("again")); err != nil || len(d.ReadAll()) != 1 {
		t.Errorf("append after clear: %v", err)
	}
}

func TestATornLastLineIsSkipped(t *testing.T) {
	dir := t.TempDir()
	d := NewDisk(dir, 1<<20)
	_ = d.Append(e("good"))
	_ = d.Close()
	f, _ := os.OpenFile(filepath.Join(dir, diskCurrent), os.O_APPEND|os.O_WRONLY, 0o600)
	_, _ = f.WriteString(`{"seq":2,"msg":"torn`)
	_ = f.Close()

	if got := d.ReadAll(); len(got) != 1 || got[0].Msg != "good" {
		t.Errorf("%+v", got)
	}
}

type fixture struct {
	svc      *Service
	set      config.HostLog
	setMu    sync.Mutex
	defaults []netip.Prefix
	board    netip.Addr
	boardOK  bool
	pubMu    sync.Mutex
	pub      []events.Event
	now      time.Time
}

func newFixture(t *testing.T, set config.HostLog) *fixture {
	t.Helper()
	f := &fixture{set: set.WithDefaults(), now: time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)}
	f.svc = New(Deps{
		Settings: func() config.HostLog {
			f.setMu.Lock()
			defer f.setMu.Unlock()
			return f.set
		},
		SetSettings: func(h config.HostLog) error {
			f.setMu.Lock()
			f.set = h.WithDefaults()
			f.setMu.Unlock()
			return nil
		},
		DefaultSources: func() []netip.Prefix { return f.defaults },
		USBBoard:       func() (netip.Addr, bool) { return f.board, f.boardOK },
		Publish: func(ev events.Event) {
			f.pubMu.Lock()
			f.pub = append(f.pub, ev)
			f.pubMu.Unlock()
		},
		DiskDir: filepath.Join(t.TempDir(), "hostlog"),
		Now:     func() time.Time { return f.now },
		Listen: func(network, address string) (net.PacketConn, error) {
			return net.ListenPacket(network, "127.0.0.1:0")
		},
	})
	return f
}

func (f *fixture) published() []events.Event {
	f.pubMu.Lock()
	defer f.pubMu.Unlock()
	return append([]events.Event(nil), f.pub...)
}

func addr(s string) netip.Addr { return netip.MustParseAddr(s) }

func TestOnlyTheHostIsHeardWhenItIsKnown(t *testing.T) {
	f := newFixture(t, config.HostLog{})
	f.defaults = []netip.Prefix{netip.MustParsePrefix("172.31.255.2/32")}
	f.svc.Reconcile()

	f.svc.Ingest(ProtoSyslog, addr("172.31.255.2"), []byte("<13>from the host"))
	f.svc.Ingest(ProtoSyslog, addr("192.168.1.9"), []byte("<13>from a stranger"))
	if lines := f.svc.Ring().Query(Filter{MaxSeverity: -1}, 0); len(lines) != 1 || lines[0].Msg != "from the host" {
		t.Errorf("%+v", lines)
	}
	if st := f.svc.State(); st.Filtered != 1 || len(st.Sources) != 1 || st.Sources[0] != "172.31.255.2/32" {
		t.Errorf("%+v", st)
	}
}

func TestAnyoneIsHeardWhenNoHostIsKnown(t *testing.T) {
	f := newFixture(t, config.HostLog{})
	f.svc.Reconcile()
	f.svc.Ingest(ProtoSyslog, addr("203.0.113.5"), []byte("<13>hello"))
	if n, _ := f.svc.Ring().Stats(); n != 1 {
		t.Errorf("%d lines", n)
	}
}

func TestExplicitSourcesReplaceTheDefault(t *testing.T) {
	f := newFixture(t, config.HostLog{Sources: []string{"10.1.0.0/16", "2001:db8::1"}})
	f.defaults = []netip.Prefix{netip.MustParsePrefix("172.31.255.2/32")}
	f.svc.Reconcile()

	f.svc.Ingest(ProtoSyslog, addr("10.1.2.3"), []byte("<13>in the range"))
	f.svc.Ingest(ProtoSyslog, addr("2001:db8::1"), []byte("<13>v6"))
	f.svc.Ingest(ProtoSyslog, addr("172.31.255.2"), []byte("<13>the default host is not named"))
	if n, _ := f.svc.Ring().Stats(); n != 2 {
		t.Errorf("%d lines, want 2", n)
	}
}

func TestAFloodIsRateLimited(t *testing.T) {
	f := newFixture(t, config.HostLog{})
	f.svc.Reconcile()
	for i := 0; i < maxBurst*3; i++ {
		f.svc.Ingest(ProtoNetconsole, addr("10.0.0.2"), []byte("spam"))
	}
	n, _ := f.svc.Ring().Stats()
	if n > maxBurst+10 || n < maxBurst-10 {
		t.Errorf("%d lines kept of %d", n, maxBurst*3)
	}
	if f.svc.State().RateLimited == 0 {
		t.Error("the drops were not counted")
	}

	f.now = f.now.Add(time.Second)
	before, _ := f.svc.Ring().Stats()
	f.svc.Ingest(ProtoNetconsole, addr("10.0.0.2"), []byte("later"))
	after, _ := f.svc.Ring().Stats()
	if after != before+1 {
		t.Error("the bucket did not refill")
	}
}

func TestAPhraseRaisesAnEventAtMostOncePerSecond(t *testing.T) {
	f := newFixture(t, config.HostLog{})
	f.svc.Reconcile()

	f.svc.Ingest(ProtoNetconsole, addr("10.0.0.2"), []byte("all fine here"))
	f.svc.Ingest(ProtoNetconsole, addr("10.0.0.2"), []byte("Kernel panic - not syncing: Fatal exception"))
	f.svc.Ingest(ProtoNetconsole, addr("10.0.0.2"), []byte("kernel PANIC again"))
	got := f.published()
	if len(got) != 1 || got[0].Type != events.HostLogMatch || !strings.Contains(got[0].Message, "Kernel panic") || got[0].Fields["source"] != "10.0.0.2" {
		t.Fatalf("%+v", got)
	}

	f.now = f.now.Add(2 * time.Second)
	f.svc.Ingest(ProtoNetconsole, addr("10.0.0.2"), []byte("BUG: unable to handle kernel NULL pointer"))
	if n := len(f.published()); n != 2 {
		t.Errorf("%d events after a second", n)
	}
}

func TestAPatternOfTheOwnersIsUsed(t *testing.T) {
	f := newFixture(t, config.HostLog{Patterns: []string{`disk (\w+) failed`}})
	f.svc.Reconcile()

	f.svc.Ingest(ProtoSyslog, addr("10.0.0.2"), []byte("<3>Disk sda FAILED"))
	f.svc.Ingest(ProtoSyslog, addr("10.0.0.2"), []byte("<3>Kernel panic is not in the owner's list now"))
	if got := f.published(); len(got) != 1 {
		t.Errorf("%+v", got)
	}
}

func TestListenersFollowTheSettings(t *testing.T) {
	f := newFixture(t, config.HostLog{})
	f.svc.Reconcile()
	if st := f.svc.State(); st.Syslog.State != "off" || st.Netconsole.State != "off" {
		t.Fatalf("both off by default: %+v", st)
	}

	f.setMu.Lock()
	f.set.Syslog.Enabled = true
	f.setMu.Unlock()
	f.svc.Reconcile()
	st := f.svc.State()
	if st.Syslog.State != "listening" || st.Netconsole.State != "off" {
		t.Fatalf("%+v", st)
	}

	// Send a datagram through the real socket.
	l := f.svc.listeners[ProtoSyslog]
	conn, err := net.Dial("udp", l.conn.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, _ = conn.Write([]byte("<13>Oct 11 10:00:00 h app: over the wire"))
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if n, _ := f.svc.Ring().Stats(); n == 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	lines := f.svc.Ring().Query(Filter{MaxSeverity: -1}, 0)
	if len(lines) != 1 || lines[0].Msg != "over the wire" || lines[0].Source != "127.0.0.1" || lines[0].Tag != "app" {
		t.Fatalf("%+v", lines)
	}

	f.setMu.Lock()
	f.set.Syslog.Enabled = false
	f.setMu.Unlock()
	f.svc.Reconcile()
	if st := f.svc.State(); st.Syslog.State != "off" {
		t.Errorf("%+v", st)
	}
}

func TestTheUSBBindWaitsForTheLink(t *testing.T) {
	f := newFixture(t, config.HostLog{Syslog: config.HostLogListener{Enabled: true}, Bind: "usb"})
	f.svc.Reconcile()
	if st := f.svc.State(); st.Syslog.State != "waiting" {
		t.Fatalf("%+v", st)
	}

	f.board, f.boardOK = addr("172.31.255.1"), true
	f.svc.Reconcile()
	if st := f.svc.State(); st.Syslog.State != "listening" || !strings.HasPrefix(st.Syslog.Addr, "172.31.255.1:") {
		t.Fatalf("%+v", st)
	}
}

func TestABindFailureIsReportedNotFatal(t *testing.T) {
	f := newFixture(t, config.HostLog{Syslog: config.HostLogListener{Enabled: true}})
	f.svc.deps.Listen = func(string, string) (net.PacketConn, error) {
		return nil, &net.OpError{Op: "listen", Err: os.ErrPermission}
	}
	f.svc.Reconcile()
	if st := f.svc.State(); st.Syslog.State != "error" || st.Syslog.Error == "" {
		t.Errorf("%+v", st)
	}
}

func TestTheDiskLogIsOffUntilAskedForAndComesBackAfterARestart(t *testing.T) {
	f := newFixture(t, config.HostLog{})
	f.svc.Reconcile()
	f.svc.Ingest(ProtoSyslog, addr("10.0.0.2"), []byte("<13>memory only"))
	if _, err := os.Stat(f.svc.deps.DiskDir); err == nil {
		t.Fatal("something was written to disk with the disk log off")
	}

	f.setMu.Lock()
	f.set.Disk = true
	f.setMu.Unlock()
	f.svc.Reconcile()
	f.svc.Ingest(ProtoSyslog, addr("10.0.0.2"), []byte("<13>on disk"))
	if f.svc.State().DiskBytes == 0 {
		t.Fatal("nothing reached the disk log")
	}

	// A new run reads the disk log back.
	g := newFixture(t, config.HostLog{Disk: true})
	g.svc.deps.DiskDir = f.svc.deps.DiskDir
	g.svc.Reconcile()
	lines := g.svc.Ring().Query(Filter{MaxSeverity: -1}, 0)
	if len(lines) != 1 || lines[0].Msg != "on disk" {
		t.Errorf("%+v", lines)
	}

	if err := g.svc.Clear(); err != nil {
		t.Fatal(err)
	}
	if n, _ := g.svc.Ring().Stats(); n != 0 || g.svc.State().DiskBytes != 0 {
		t.Error("clear left lines behind")
	}
}

func TestSettingsAreChecked(t *testing.T) {
	good := settingsOf(config.HostLog{}.WithDefaults())
	if _, err := good.validate(); err != nil {
		t.Fatalf("the defaults do not validate: %s", err)
	}

	cases := map[string]func(*Settings){
		"port zero":    func(s *Settings) { s.Syslog.Port = 0 },
		"port too big": func(s *Settings) { s.Netconsole.Port = 70000 },
		"same port": func(s *Settings) {
			s.Syslog.Enabled, s.Netconsole.Enabled, s.Netconsole.Port = true, true, s.Syslog.Port
		},
		"bad bind":          func(s *Settings) { s.Bind = "eth0" },
		"bad source":        func(s *Settings) { s.Sources = []string{"not-an-address"} },
		"too many sources":  func(s *Settings) { s.Sources = make([]string, maxSources+1) },
		"memory too small":  func(s *Settings) { s.MemoryKB = 1 },
		"memory too big":    func(s *Settings) { s.MemoryKB = 1 << 20 },
		"disk too small":    func(s *Settings) { s.DiskKB = 1 },
		"bad pattern":       func(s *Settings) { s.Patterns = []string{"("} },
		"empty pattern":     func(s *Settings) { s.Patterns = []string{""} },
		"too many patterns": func(s *Settings) { s.Patterns = make([]string, maxPatterns+1) },
	}
	for name, mutate := range cases {
		s := settingsOf(config.HostLog{}.WithDefaults())
		mutate(&s)
		if _, err := s.validate(); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestDefaultPatternsCompileAndMatchAPanic(t *testing.T) {
	res, err := patternsOf(config.DefaultHostLogPatterns)
	if err != nil {
		t.Fatal(err)
	}
	hit := func(s string) bool {
		for _, re := range res {
			if re.MatchString(s) {
				return true
			}
		}
		return false
	}
	for _, s := range []string{"Kernel panic - not syncing", "Oops: 0002 [#1] SMP", "BUG: soft lockup", "Out of memory: Killed process 1"} {
		if !hit(s) {
			t.Errorf("%q does not match", s)
		}
	}
	if hit("everything is fine") || hit("Oopsie daisy") {
		t.Error("a harmless line matched")
	}
}
