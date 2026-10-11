package nut

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

// fakeServer speaks enough of upsd for the client: USERNAME, PASSWORD, LOGIN,
// LOGOUT and LIST VAR. It is a real TCP server, so the tests cover the wire.
type fakeServer struct {
	t  *testing.T
	ln net.Listener

	mu       sync.Mutex
	ups      string
	vars     map[string]string
	user     string
	password string
	logged   int
	commands []string
	accepts  int
	conns    []net.Conn
}

func newFakeServer(t *testing.T) *fakeServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &fakeServer{t: t, ln: ln, ups: "ups", vars: map[string]string{
		"ups.status":      "OL",
		"battery.charge":  "100",
		"battery.runtime": "1800",
		"ups.mfr":         `Back "Ups" \ Co`,
	}}
	go s.serve()
	t.Cleanup(s.close)
	return s
}

func (s *fakeServer) addr() string { return s.ln.Addr().String() }

func (s *fakeServer) hostPort() (string, int) {
	h, p, _ := net.SplitHostPort(s.addr())
	var port int
	fmt.Sscanf(p, "%d", &port)
	return h, port
}

func (s *fakeServer) close() {
	_ = s.ln.Close()
	s.mu.Lock()
	for _, c := range s.conns {
		_ = c.Close()
	}
	s.conns = nil
	s.mu.Unlock()
}

func (s *fakeServer) set(status string, charge, runtime int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vars["ups.status"] = status
	s.vars["battery.charge"] = fmt.Sprint(charge)
	s.vars["battery.runtime"] = fmt.Sprint(runtime)
}

func (s *fakeServer) loggedIn() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.logged
}

func (s *fakeServer) acceptCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.accepts
}

func (s *fakeServer) saw(cmd string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.commands {
		if strings.HasPrefix(c, cmd) {
			return true
		}
	}
	return false
}

// waitFor polls until cond holds, for the server side of an exchange that the
// client does not wait for.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(2 * time.Millisecond)
	}
}

func (s *fakeServer) serve() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.mu.Lock()
		s.accepts++
		s.conns = append(s.conns, conn)
		s.mu.Unlock()
		go s.handle(conn)
	}
}

func (s *fakeServer) handle(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)

	user, pass := "", ""
	loggedIn := false
	defer func() {
		if loggedIn {
			s.mu.Lock()
			s.logged--
			s.mu.Unlock()
		}
	}()

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		s.mu.Lock()
		s.commands = append(s.commands, line)
		s.mu.Unlock()

		fields, err := split(line)
		if err != nil || len(fields) == 0 {
			fmt.Fprint(conn, "ERR UNKNOWN-COMMAND\n")
			continue
		}
		switch strings.ToUpper(fields[0]) {
		case "USERNAME":
			user = fields[1]
			fmt.Fprint(conn, "OK\n")
		case "PASSWORD":
			pass = fields[1]
			fmt.Fprint(conn, "OK\n")
		case "LOGIN":
			s.mu.Lock()
			ok := fields[1] == s.ups
			authed := s.user != "" && user == s.user && pass == s.password
			if ok && authed && !loggedIn {
				s.logged++
				loggedIn = true
			}
			s.mu.Unlock()
			switch {
			case !ok:
				fmt.Fprint(conn, "ERR UNKNOWN-UPS\n")
			case !authed:
				fmt.Fprint(conn, "ERR ACCESS-DENIED\n")
			default:
				fmt.Fprint(conn, "OK\n")
			}
		case "LOGOUT":
			fmt.Fprint(conn, "OK Goodbye\n")
			return
		case "LIST":
			s.mu.Lock()
			if len(fields) < 3 || fields[1] != "VAR" || fields[2] != s.ups {
				s.mu.Unlock()
				fmt.Fprint(conn, "ERR UNKNOWN-UPS\n")
				continue
			}
			var b strings.Builder
			fmt.Fprintf(&b, "BEGIN LIST VAR %s\n", s.ups)
			for k, v := range s.vars {
				esc := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(v)
				fmt.Fprintf(&b, "VAR %s %s \"%s\"\n", s.ups, k, esc)
			}
			fmt.Fprintf(&b, "END LIST VAR %s\n", s.ups)
			s.mu.Unlock()
			fmt.Fprint(conn, b.String())
		default:
			fmt.Fprint(conn, "ERR UNKNOWN-COMMAND\n")
		}
	}
}

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}
