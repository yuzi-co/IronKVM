package hostlog

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Protocols, as an entry names the listener that received it.
const (
	ProtoSyslog     = "syslog"
	ProtoNetconsole = "netconsole"
)

// maxMessage bounds one message. A host cannot make the ring hold more than
// this per line.
const maxMessage = 2048

// parsed is one message before it is stamped and stored.
type parsed struct {
	facility int // -1 when the sender gave none
	severity int // -1 when the sender gave none
	host     string
	tag      string
	msg      string
}

// clean makes text safe to store and to show: valid UTF-8, no control
// characters (an escape sequence in a log line must not reach a terminal or
// hide text), and no longer than maxMessage.
func clean(s string) string {
	s = strings.ToValidUTF8(s, "\uFFFD")
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\t':
			b.WriteRune(' ')
		case r < 0x20 || r == 0x7f:
			// dropped
		default:
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) > maxMessage {
		cut := maxMessage
		for cut > 0 && !utf8.RuneStart(out[cut]) {
			cut--
		}
		out = out[:cut] + "..."
	}
	return strings.TrimSpace(out)
}

// splitPRI reads a leading <PRI>. ok is false when there is none.
func splitPRI(s string) (facility, severity int, rest string, ok bool) {
	if len(s) < 3 || s[0] != '<' {
		return -1, -1, s, false
	}
	end := strings.IndexByte(s, '>')
	if end < 2 || end > 4 {
		return -1, -1, s, false
	}
	n, err := strconv.Atoi(s[1:end])
	if err != nil || n < 0 || n > 191 {
		return -1, -1, s, false
	}
	return n >> 3, n & 7, s[end+1:], true
}

var (
	// RFC 3164: "Mmm dd hh:mm:ss HOST TAG[pid]: msg". The day may be padded
	// with a space.
	rfc3164Time = regexp.MustCompile(`^[A-Z][a-z]{2} [ 0-9][0-9] [0-9]{2}:[0-9]{2}:[0-9]{2} `)
	tagRe       = regexp.MustCompile(`^([^\s:\[\]]{1,48})(?:\[[0-9]+\])?: ?`)
)

// parseSyslog reads one syslog datagram, RFC 5424 or RFC 3164. A datagram of
// neither form is kept whole as the message.
func parseSyslog(data []byte) parsed {
	text := strings.TrimRight(string(data), "\x00\r\n ")
	p := parsed{facility: -1, severity: -1}

	fac, sev, rest, ok := splitPRI(text)
	if ok {
		p.facility, p.severity = fac, sev
	}
	text = rest

	if ok && parse5424(text, &p) {
		return p
	}

	if rfc3164Time.MatchString(text) {
		text = text[16:]
		if i := strings.IndexByte(text, ' '); i > 0 {
			p.host = clean(text[:i])
			text = text[i+1:]
		}
	}
	if m := tagRe.FindStringSubmatch(text); m != nil {
		p.tag = clean(m[1])
		text = text[len(m[0]):]
	}
	p.msg = clean(text)
	return p
}

// parse5424 reads "VERSION TIMESTAMP HOST APP PROCID MSGID SD MSG" after the
// PRI. It reports false when the text is not that form.
func parse5424(text string, p *parsed) bool {
	i := 0
	for i < len(text) && text[i] >= '0' && text[i] <= '9' {
		i++
	}
	if i == 0 || i > 2 || i >= len(text) || text[i] != ' ' {
		return false
	}
	fields := strings.SplitN(text[i+1:], " ", 6)
	if len(fields) < 6 {
		return false
	}
	// fields: timestamp, host, app, procid, msgid, then SD and message.
	rest := fields[5]
	switch {
	case strings.HasPrefix(rest, "- "), rest == "-":
		rest = strings.TrimPrefix(strings.TrimPrefix(rest, "-"), " ")
	case strings.HasPrefix(rest, "["):
		rest = skipStructuredData(rest)
	default:
		return false
	}

	if fields[1] != "-" {
		p.host = clean(fields[1])
	}
	if fields[2] != "-" {
		p.tag = clean(fields[2])
	}
	rest = strings.TrimPrefix(rest, "\xef\xbb\xbf")
	p.msg = clean(rest)
	return true
}

// skipStructuredData drops the [id k="v"]... elements at the start of s and
// returns what follows, without the space after them. A "]" inside a quoted
// value is escaped as \].
func skipStructuredData(s string) string {
	i := 0
	for i < len(s) && s[i] == '[' {
		inQuote := false
		j := i + 1
		for ; j < len(s); j++ {
			c := s[j]
			if c == '\\' {
				j++
				continue
			}
			if c == '"' {
				inQuote = !inQuote
				continue
			}
			if c == ']' && !inQuote {
				break
			}
		}
		if j >= len(s) {
			return ""
		}
		i = j + 1
	}
	return strings.TrimPrefix(s[i:], " ")
}

// extNetconsole is the extended format the kernel sends when netconsole is
// loaded with a leading "+": "LEVEL,SEQ,TIMESTAMP,FLAG[,KEY=VALUE];MESSAGE",
// then optional continuation lines that begin with a space.
var extNetconsole = regexp.MustCompile(`^([0-9]{1,3}),[0-9]+,[0-9]+,[-cC](?:,[^;]*)?;`)

// parseNetconsole reads one netconsole datagram into one entry per line. The
// plain format is the kernel's text, a line or several; the extended one
// carries the level.
func parseNetconsole(data []byte) []parsed {
	text := strings.TrimRight(string(data), "\x00\r\n ")

	if m := extNetconsole.FindStringSubmatch(text); m != nil {
		level, _ := strconv.Atoi(m[1])
		first := text[len(m[0]):]
		if nl := strings.IndexByte(first, '\n'); nl >= 0 {
			first = first[:nl]
		}
		if level > 191 {
			level = 191
		}
		return []parsed{{facility: level >> 3, severity: level & 7, msg: clean(first)}}
	}

	var out []parsed
	for _, line := range strings.Split(text, "\n") {
		p := parsed{facility: -1, severity: -1}
		if fac, sev, rest, ok := splitPRI(line); ok {
			p.facility, p.severity, line = fac, sev, rest
		}
		if p.msg = clean(line); p.msg != "" {
			out = append(out, p)
		}
	}
	return out
}
