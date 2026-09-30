package logger

import (
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestSyslogFormatCarriesPriorityTagAndPid(t *testing.T) {
	at := time.Date(2026, time.September, 3, 7, 4, 5, 0, time.UTC)
	cases := []struct {
		level logrus.Level
		pri   string
	}{
		{logrus.FatalLevel, "<26>"},
		{logrus.ErrorLevel, "<27>"},
		{logrus.WarnLevel, "<28>"},
		{logrus.InfoLevel, "<30>"},
		{logrus.DebugLevel, "<31>"},
		{logrus.TraceLevel, "<31>"},
	}
	for _, c := range cases {
		got := string(appendSyslog(nil, syslogEntry{level: c.level, time: at, message: "hello"}, "NanoKVM-Server", 42))
		want := c.pri + "Sep  3 07:04:05 NanoKVM-Server[42]: hello"
		if got != want {
			t.Errorf("%s: got %q, want %q", c.level, got, want)
		}
	}
}

func TestSyslogKeepsAMultiLineEntryAsOneMessage(t *testing.T) {
	e := syslogEntry{level: logrus.ErrorLevel, time: time.Unix(0, 0).UTC(), message: "panic: boom\ngoroutine 1:\r\n\tmain.go:1\n"}
	got := string(appendSyslog(nil, e, "t", 1))
	if !strings.HasSuffix(got, "]: panic: boom goroutine 1:  \tmain.go:1") {
		t.Fatalf("got %q", got)
	}
}

func TestSyslogNeverBlocksWithoutASocket(t *testing.T) {
	h := newSyslogHook(filepath.Join(t.TempDir(), "absent"), "t")

	entry := logrus.NewEntry(logrus.New())
	entry.Level = logrus.InfoLevel
	entry.Message = "nobody listens"

	start := time.Now()
	for i := 0; i < 10*syslogQueue; i++ {
		if err := h.Fire(entry); err != nil {
			t.Fatalf("Fire returned %s", err)
		}
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("firing into a missing socket took %s", elapsed)
	}
}

func TestSyslogDeliversAndRedialsAfterSyslogdRestarts(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no unixgram sockets")
	}

	path := filepath.Join(t.TempDir(), "log")
	listen := func() *net.UnixConn {
		conn, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: path, Net: "unixgram"})
		if err != nil {
			t.Fatalf("listen: %s", err)
		}
		return conn
	}

	h := newSyslogHook(path, "NanoKVM-Server")
	fire := func(msg string) {
		entry := logrus.NewEntry(logrus.New())
		entry.Level = logrus.WarnLevel
		entry.Time = time.Now()
		entry.Message = msg
		_ = h.Fire(entry)
	}
	receive := func(conn *net.UnixConn, msg string) {
		t.Helper()
		buf := make([]byte, 1024)
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			fire(msg)
			_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			n, err := conn.Read(buf)
			if err != nil {
				continue
			}
			got := string(buf[:n])
			if !strings.HasPrefix(got, "<28>") || !strings.Contains(got, " NanoKVM-Server[") || !strings.HasSuffix(got, "]: "+msg) {
				t.Fatalf("unexpected datagram %q", got)
			}
			return
		}
		t.Fatalf("%q never arrived", msg)
	}

	first := listen()
	receive(first, "before")

	// syslogd restarts: the socket goes and a new one takes its path.
	_ = first.Close()
	_ = os.Remove(path)
	second := listen()
	defer second.Close()
	receive(second, "after")
}
