package vpn

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
)

// CgroupDir is the addons memory group that ironkvm-dist's S01cgroups makes.
// A variable for the tests.
var CgroupDir = "/sys/fs/cgroup/addons"

// clockTicks is USER_HZ, the unit of the start time in /proc/<pid>/stat. It is
// 100 on every Linux this server runs on.
const clockTicks = 100

// UptimeSec is how long the process has run, in whole seconds, or 0 when it
// cannot be read.
func UptimeSec(pid int) int64 {
	stat, err := os.ReadFile(filepath.Join(addon.ProcDir, strconv.Itoa(pid), "stat"))
	if err != nil {
		return 0
	}
	// The command name is in parentheses and may hold spaces, so the fields
	// are counted from the last ')'. After it, field 3 (state) is index 0, and
	// field 22 (starttime) is index 19.
	s := string(stat)
	i := strings.LastIndexByte(s, ')')
	if i < 0 {
		return 0
	}
	fields := strings.Fields(s[i+1:])
	if len(fields) < 20 {
		return 0
	}
	start, err := strconv.ParseInt(fields[19], 10, 64)
	if err != nil {
		return 0
	}
	up, err := os.ReadFile(filepath.Join(addon.ProcDir, "uptime"))
	if err != nil {
		return 0
	}
	first, _, _ := strings.Cut(strings.TrimSpace(string(up)), " ")
	system, err := strconv.ParseFloat(first, 64)
	if err != nil {
		return 0
	}
	if sec := int64(system) - start/clockTicks; sec > 0 {
		return sec
	}
	return 0
}

// RSS is the process's resident memory in bytes, from VmRSS, or 0.
func RSS(pid int) int64 {
	b, err := os.ReadFile(filepath.Join(addon.ProcDir, strconv.Itoa(pid), "status"))
	if err != nil {
		return 0
	}
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		rest, ok := strings.CutPrefix(sc.Text(), "VmRSS:")
		if !ok {
			continue
		}
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			return 0
		}
		kb, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return 0
		}
		return kb * 1024
	}
	return 0
}

// GroupMemory reads the addons group's use and its two limits, in bytes. A
// limit of "max", or a group that is not there, reads as 0.
func GroupMemory() (current, high, max int64) {
	return readBytes("memory.current"), readBytes("memory.high"), readBytes("memory.max")
}

func readBytes(name string) int64 {
	b, err := os.ReadFile(filepath.Join(CgroupDir, name))
	if err != nil {
		return 0
	}
	v, err := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// Fill adds what both add-ons report the same way: start at boot, the other
// VPN when it blocks this one, and the daemon's uptime and memory beside its
// group's.
func Fill(st *proto.VpnStatus, d addon.Daemon) {
	st.BootEnabled = addon.BootEnabled(d)
	st.BlockedBy = addon.BlockedBy(d.Name)
	if pid, ok := addon.PID(d); ok {
		st.UptimeSec = UptimeSec(pid)
		st.Memory.DaemonRSS = RSS(pid)
	}
	st.Memory.GroupCurrent, st.Memory.GroupHigh, st.Memory.GroupMax = GroupMemory()
	if st.Peers == nil {
		st.Peers = []proto.VpnPeer{}
	}
}
