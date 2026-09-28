package vm

import (
	"os"
	"os/exec"
	"regexp"
	"strings"

	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	BootHostnameFile = "/boot/hostname"
	EtcHostname      = "/etc/hostname"
	EtcHosts         = "/etc/hosts"
)

func (s *Service) SetHostname(c *gin.Context) {
	var req proto.SetHostnameReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if !validHostname(req.Hostname) {
		rsp.ErrRsp(c, -1, "invalid hostname")
		return
	}

	dataRead, err := os.ReadFile(EtcHostname)
	if err != nil {
		rsp.ErrRsp(c, -1, "read Hostname failed")
		return
	}

	oldHostname := strings.Replace(string(dataRead), "\n", "", -1)

	if oldHostname != req.Hostname {
		dataRead, err = os.ReadFile(EtcHosts)
		if err != nil {
			rsp.ErrRsp(c, -1, "read Hosts failed")
			return
		}

		data := []byte(renameHost(string(dataRead), oldHostname, req.Hostname))

		if err := os.WriteFile(EtcHosts, data, 0o644); err != nil {
			rsp.ErrRsp(c, -2, "failed to write data")
			return
		}
	}

	data := []byte(req.Hostname)

	if err := os.WriteFile(BootHostnameFile, data, 0o644); err != nil {
		rsp.ErrRsp(c, -2, "failed to write data")
		return
	}

	if err := os.WriteFile(EtcHostname, data, 0o644); err != nil {
		rsp.ErrRsp(c, -3, "failed to write data")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("set Hostname: %s", req.Hostname)

	_ = exec.Command("hostname", "-F", EtcHostname).Run()
}

func (s *Service) GetHostname(c *gin.Context) {
	var rsp proto.Response

	data, err := os.ReadFile(EtcHostname)
	if err != nil {
		rsp.ErrRsp(c, -1, "read Hostname failed")
		return
	}

	rsp.OkRspWithData(c, &proto.GetHostnameRsp{
		Hostname: strings.Replace(string(data), "\n", "", -1),
	})
	log.Debugf("get Hostname successful")
}

// hostnameLabel is one label of an RFC 1123 host name: letters, digits and
// hyphens, not starting or ending with a hyphen, at most 63 characters.
var hostnameLabel = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)

// validHostname reports whether name is an RFC 1123 host name the kernel will
// take: dot-separated labels, 64 characters at most, which is the kernel's
// limit and shorter than the 253 the RFC allows.
//
// The name is written to /etc/hosts and /etc/hostname and announced over mDNS,
// so a space, a newline or a slash in it breaks name resolution on the board.
func validHostname(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for _, label := range strings.Split(name, ".") {
		if !hostnameLabel.MatchString(label) {
			return false
		}
	}
	return true
}

// renameHost replaces the host name oldName with newName in the text of
// /etc/hosts. Only whole names change: a name is a whitespace-separated field
// after the address, and a comment is left alone. Replacing substrings turned
// "localhost" into "newnamehost" whenever the old name was "local".
func renameHost(hosts, oldName, newName string) string {
	if oldName == "" {
		return hosts
	}

	lines := strings.Split(hosts, "\n")
	for i, line := range lines {
		body, comment := line, ""
		if idx := strings.IndexByte(line, '#'); idx >= 0 {
			body, comment = line[:idx], line[idx:]
		}

		// Walk the fields keeping the whitespace between them, so a line that
		// does not change stays byte for byte the same.
		var out strings.Builder
		field := 0
		for j := 0; j < len(body); {
			if body[j] == ' ' || body[j] == '	' {
				out.WriteByte(body[j])
				j++
				continue
			}
			k := j
			for k < len(body) && body[k] != ' ' && body[k] != '	' {
				k++
			}
			token := body[j:k]
			// The first field is the address.
			if field > 0 && token == oldName {
				token = newName
			}
			out.WriteString(token)
			field++
			j = k
		}
		lines[i] = out.String() + comment
	}
	return strings.Join(lines, "\n")
}
