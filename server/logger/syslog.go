package logger

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// The board's busybox syslogd listens on a datagram socket at /dev/log and,
// when the owner named a collector, forwards everything it receives. Sending
// the server's entries there as well as to its own file puts them beside the
// kernel's and the init scripts' on that collector.
const (
	syslogSocket = "/dev/log"
	syslogTag    = "NanoKVM-Server"

	// syslogQueue is how many entries may wait for the sender. A burst
	// larger than this loses the excess rather than holding up the caller.
	syslogQueue = 256

	// syslogWriteTimeout bounds one datagram. A syslogd that stopped reading
	// fills the socket's buffer, and a write then waits for room.
	syslogWriteTimeout = 100 * time.Millisecond

	// syslogRedial is how long the sender waits after a failed dial before
	// it tries again, dropping what arrives meanwhile.
	syslogRedial = time.Second

	// syslogFlushWait is how long a fatal or panic entry waits for the sender
	// to pass it on, since the process ends straight after.
	syslogFlushWait = 200 * time.Millisecond

	// facilityDaemon is LOG_DAEMON, already shifted.
	facilityDaemon = 3 << 3
)

// syslogEntry is what Fire hands the sender. It holds the message string
// rather than a formatted copy, so Fire itself allocates nothing.
type syslogEntry struct {
	level   logrus.Level
	time    time.Time
	message string
	// done, when set, is closed once the entry was written or dropped.
	done chan struct{}
}

// syslogHook sends each entry to the local syslog socket from its own
// goroutine. Nothing it does can block or fail a caller: a full queue, a
// missing socket and a syslogd that restarted all drop the entry quietly.
type syslogHook struct {
	queue  chan syslogEntry
	socket string
	tag    string
	pid    int

	conn     *net.UnixConn
	nextDial time.Time
	buf      []byte
}

func newSyslogHook(socket, tag string) *syslogHook {
	h := &syslogHook{
		queue:  make(chan syslogEntry, syslogQueue),
		socket: socket,
		tag:    tag,
		pid:    os.Getpid(),
		buf:    make([]byte, 0, 512),
	}
	go h.run()
	return h
}

func (h *syslogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *syslogHook) Fire(entry *logrus.Entry) error {
	e := syslogEntry{level: entry.Level, time: entry.Time, message: entry.Message}

	// A fatal entry exits the process and a panic may, so wait a moment for
	// either to leave. Every other entry is handed over and forgotten.
	if entry.Level <= logrus.FatalLevel {
		e.done = make(chan struct{})
	}

	select {
	case h.queue <- e:
	default:
		return nil
	}

	if e.done != nil {
		select {
		case <-e.done:
		case <-time.After(syslogFlushWait):
		}
	}
	return nil
}

func (h *syslogHook) run() {
	for e := range h.queue {
		h.send(e)
		if e.done != nil {
			close(e.done)
		}
	}
}

// send writes one datagram, dialling first if there is no connection. A write
// that fails closes the connection, so the next entry dials afresh: that is
// how a restarted syslogd is picked up.
func (h *syslogHook) send(e syslogEntry) {
	if h.conn == nil {
		now := time.Now()
		if now.Before(h.nextDial) {
			return
		}
		conn, err := net.DialUnix("unixgram", nil, &net.UnixAddr{Name: h.socket, Net: "unixgram"})
		if err != nil {
			h.nextDial = now.Add(syslogRedial)
			return
		}
		h.conn = conn
	}

	h.buf = appendSyslog(h.buf[:0], e, h.tag, h.pid)
	_ = h.conn.SetWriteDeadline(time.Now().Add(syslogWriteTimeout))
	if _, err := h.conn.Write(h.buf); err != nil {
		_ = h.conn.Close()
		h.conn = nil
	}
}

// appendSyslog formats one entry the way a local syslog client does:
// "<PRI>Mmm dd hh:mm:ss TAG[PID]: MESSAGE". syslogd adds the host name when it
// forwards. A message of several lines stays one message, its line breaks
// turned into spaces, so a collector keys every line on the tag.
func appendSyslog(dst []byte, e syslogEntry, tag string, pid int) []byte {
	dst = append(dst, '<')
	dst = strconv.AppendInt(dst, int64(facilityDaemon|severity(e.level)), 10)
	dst = append(dst, '>')
	dst = e.time.AppendFormat(dst, time.Stamp)
	dst = append(dst, ' ')
	dst = append(dst, tag...)
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, int64(pid), 10)
	dst = append(dst, "]: "...)

	msg := strings.TrimRight(e.message, "\r\n")
	for i := 0; i < len(msg); i++ {
		c := msg[i]
		if c == '\n' || c == '\r' {
			c = ' '
		}
		dst = append(dst, c)
	}
	return dst
}

// severity maps a logrus level to a syslog one, as logrus's own syslog hook
// does.
func severity(level logrus.Level) int {
	switch level {
	case logrus.PanicLevel, logrus.FatalLevel:
		return 2 // crit
	case logrus.ErrorLevel:
		return 3 // err
	case logrus.WarnLevel:
		return 4 // warning
	case logrus.InfoLevel:
		return 6 // info
	default:
		return 7 // debug
	}
}
