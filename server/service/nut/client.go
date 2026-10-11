package nut

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// This file is a client for the NUT network protocol (TCP 3493), the part of
// it a monitor needs: USERNAME, PASSWORD, LOGIN, LOGOUT and LIST VAR. The
// protocol is one line per command and per reply, so it is written here
// rather than pulled in as a library or shipped as upsmon: a Go NUT library
// would be a dependency of a few hundred lines, most of it commands this does
// not use, with no timeouts of its own, and upsmon would be a package, a
// config file and a process on a board with 158 MB of RAM. The protocol is
// described in docs/net-protocol.txt of the NUT sources.
//
// The connection is plain text. STARTTLS is not used, so on an untrusted
// network the password crosses it in the clear, as it does for upsmon without
// a certificate.

// maxLine bounds one reply line, and maxVars the lines of one list, so a
// server that talks nonsense cannot grow the board's memory.
const (
	maxLine = 8192
	maxVars = 4096
)

// ServerError is an ERR reply, such as ACCESS-DENIED or UNKNOWN-UPS.
type ServerError struct {
	Code string
}

func (e *ServerError) Error() string { return "the NUT server answered ERR " + e.Code }

// Client is one connection to a NUT server. It is not safe for concurrent use.
type Client struct {
	conn    net.Conn
	r       *bufio.Reader
	timeout time.Duration
}

// Dial connects to addr, a host:port. timeout bounds the connect and every
// exchange after it.
func Dial(ctx context.Context, addr string, timeout time.Duration) (*Client, error) {
	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return newClient(conn, timeout), nil
}

func newClient(conn net.Conn, timeout time.Duration) *Client {
	return &Client{conn: conn, r: bufio.NewReaderSize(conn, maxLine), timeout: timeout}
}

// Close drops the connection without a LOGOUT.
func (c *Client) Close() error { return c.conn.Close() }

// Authenticate sends the user name and the password. upsd keeps them for the
// session; it checks them when a command needs them, and LOGIN does.
func (c *Client) Authenticate(user, password string) error {
	if err := c.expectOK("USERNAME", user); err != nil {
		return err
	}
	return c.expectOK("PASSWORD", password)
}

// Login registers this connection as a client of ups. A primary's upsmon
// counts these before it cuts the UPS output, which is the point: while the
// KVM is logged in, the primary waits for the host to shut down.
func (c *Client) Login(ups string) error {
	return c.expectOK("LOGIN", ups)
}

// Logout says goodbye and closes the connection. A server that already hung up
// is not an error.
func (c *Client) Logout() error {
	err := c.send("LOGOUT")
	if err == nil {
		_, _ = c.readLine()
	}
	_ = c.Close()
	return err
}

// Vars returns every variable of ups.
func (c *Client) Vars(ups string) (map[string]string, error) {
	if err := c.send("LIST", "VAR", ups); err != nil {
		return nil, err
	}

	line, err := c.readLine()
	if err != nil {
		return nil, err
	}
	fields, err := split(line)
	if err != nil {
		return nil, err
	}
	if len(fields) < 2 || fields[0] != "BEGIN" || fields[1] != "LIST" {
		return nil, replyError(fields)
	}

	vars := map[string]string{}
	for {
		line, err := c.readLine()
		if err != nil {
			return nil, err
		}
		fields, err := split(line)
		if err != nil {
			return nil, err
		}
		switch {
		case len(fields) >= 1 && fields[0] == "END":
			return vars, nil
		case len(fields) == 4 && fields[0] == "VAR":
			if len(vars) >= maxVars {
				return nil, errors.New("the NUT server sent too many variables")
			}
			vars[fields[2]] = fields[3]
		case len(fields) >= 1 && fields[0] == "ERR":
			return nil, replyError(fields)
		default:
			return nil, fmt.Errorf("unexpected NUT reply %q", truncate(line))
		}
	}
}

func (c *Client) expectOK(cmd string, args ...string) error {
	if err := c.send(cmd, args...); err != nil {
		return err
	}
	line, err := c.readLine()
	if err != nil {
		return err
	}
	fields, err := split(line)
	if err != nil {
		return err
	}
	if len(fields) >= 1 && fields[0] == "OK" {
		return nil
	}
	return replyError(fields)
}

func replyError(fields []string) error {
	if len(fields) >= 2 && fields[0] == "ERR" {
		return &ServerError{Code: fields[1]}
	}
	return fmt.Errorf("unexpected NUT reply %q", truncate(strings.Join(fields, " ")))
}

func truncate(s string) string {
	if len(s) > 80 {
		return s[:80] + "..."
	}
	return s
}

// send writes one command. An argument with a space, a quote or a backslash is
// quoted; one with a control character is refused, because a newline in it
// would be a second command.
func (c *Client) send(cmd string, args ...string) error {
	var b strings.Builder
	b.WriteString(cmd)
	for _, a := range args {
		q, err := quote(a)
		if err != nil {
			return err
		}
		b.WriteByte(' ')
		b.WriteString(q)
	}
	b.WriteByte('\n')

	_ = c.conn.SetDeadline(time.Now().Add(c.timeout))
	_, err := c.conn.Write([]byte(b.String()))
	return err
}

func (c *Client) readLine() (string, error) {
	_ = c.conn.SetDeadline(time.Now().Add(c.timeout))
	line, err := c.r.ReadSlice('\n')
	if errors.Is(err, bufio.ErrBufferFull) {
		return "", errors.New("the NUT server sent a line that is too long")
	}
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(line), "\r\n"), nil
}

var errControl = errors.New("a NUT argument has a control character")

func quote(s string) (string, error) {
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			return "", errControl
		}
	}
	if s != "" && !strings.ContainsAny(s, " \t\"\\") {
		return s, nil
	}
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`, nil
}

// split breaks a protocol line into words. A word may be quoted, with a
// backslash escaping a quote or a backslash inside it.
func split(line string) ([]string, error) {
	var out []string
	var cur strings.Builder
	inWord, inQuote := false, false

	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case inQuote && ch == '\\' && i+1 < len(line):
			i++
			cur.WriteByte(line[i])
		case inQuote && ch == '"':
			inQuote = false
		case inQuote:
			cur.WriteByte(ch)
		case ch == '"':
			inQuote, inWord = true, true
		case ch == ' ' || ch == '\t':
			if inWord {
				out = append(out, cur.String())
				cur.Reset()
				inWord = false
			}
		default:
			cur.WriteByte(ch)
			inWord = true
		}
	}
	if inQuote {
		return nil, errors.New("the NUT server sent an unterminated quote")
	}
	if inWord {
		out = append(out, cur.String())
	}
	return out, nil
}
