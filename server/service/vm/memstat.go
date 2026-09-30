package vm

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
)

// memZramDir is zram's sysfs directory. A variable so the tests can point it
// at a fixture; /proc is addon.ProcDir and the add-ons' group vpn.CgroupDir,
// which the tests move the same way.
var memZramDir = zramSysfsDir

// memSelfPID is this server's pid. A variable for the tests.
var memSelfPID = os.Getpid

// kvmSystemComm is the name kvm_system runs under, as /proc/<pid>/comm has it.
const kvmSystemComm = "kvm_system"

// kvmSystemPID remembers where kvm_system was last found, so a poll reads one
// comm file rather than every process's while it keeps running.
var kvmSystemPID struct {
	sync.Mutex
	pid int
}

// GetMemory reports RAM, swap and the main consumers of memory. Every reading
// is a file read, with no command run, because the page asks every few
// seconds while it is open. A source that cannot be read is left out, so a
// board without zram or the add-ons' group still answers.
func (s *Service) GetMemory(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, readMemory())
}

func readMemory() *proto.GetMemoryRsp {
	proc := addon.ProcDir
	data := &proto.GetMemoryRsp{
		Swaps:     parseSwaps(readFileString(filepath.Join(proc, "swaps"))),
		Processes: []proto.MemoryProcess{},
	}

	info := parseMeminfo(readFileString(filepath.Join(proc, "meminfo")))
	data.Total, data.Available, data.Free = info["MemTotal"], info["MemAvailable"], info["MemFree"]

	for _, swap := range data.Swaps {
		if swap.Kind == "zram" {
			data.ZramMemUsed = parseMmStat(readFileString(filepath.Join(memZramDir, "mm_stat"))).MemUsed
			break
		}
	}

	if rss := vpn.RSS(memSelfPID()); rss > 0 {
		data.Processes = append(data.Processes, proto.MemoryProcess{Name: "NanoKVM-Server", RSS: rss})
	}
	if pid, ok := findKvmSystem(proc); ok {
		if rss := vpn.RSS(pid); rss > 0 {
			data.Processes = append(data.Processes, proto.MemoryProcess{Name: kvmSystemComm, RSS: rss})
		}
	}
	for _, d := range healthVpns() {
		pid, ok := addon.PID(d)
		if !ok {
			continue
		}
		if rss := vpn.RSS(pid); rss > 0 {
			data.Processes = append(data.Processes, proto.MemoryProcess{Name: d.Process, RSS: rss})
		}
	}

	if current, high, max := vpn.GroupMemory(); current > 0 {
		data.Addons = &proto.MemoryGroup{Current: current, High: high, Max: max}
	}

	return data
}

// findKvmSystem returns kvm_system's pid: the one it had last time while it
// still names kvm_system, or else the first process that does.
func findKvmSystem(proc string) (int, bool) {
	kvmSystemPID.Lock()
	defer kvmSystemPID.Unlock()

	if pid := kvmSystemPID.pid; pid > 0 && readComm(proc, pid) == kvmSystemComm {
		return pid, true
	}
	kvmSystemPID.pid = 0

	entries, err := os.ReadDir(proc)
	if err != nil {
		return 0, false
	}
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil || pid <= 0 {
			continue
		}
		if readComm(proc, pid) == kvmSystemComm {
			kvmSystemPID.pid = pid
			return pid, true
		}
	}
	return 0, false
}

func readComm(proc string, pid int) string {
	return strings.TrimSpace(readFileString(filepath.Join(proc, strconv.Itoa(pid), "comm")))
}

// parseMeminfo reads the "Name: value kB" lines of /proc/meminfo, in bytes.
func parseMeminfo(content string) map[string]int64 {
	values := make(map[string]int64)
	for _, line := range strings.Split(content, "\n") {
		name, rest, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		value, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			continue
		}
		if len(fields) > 1 && fields[1] == "kB" {
			value *= 1024
		}
		values[strings.TrimSpace(name)] = value
	}
	return values
}

// parseSwaps reads every device of /proc/swaps, whose sizes are in KiB. A zram
// device is told apart from other partitions, since it lives in RAM.
func parseSwaps(content string) []proto.MemorySwap {
	swaps := []proto.MemorySwap{}
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 || fields[0] == "Filename" {
			continue
		}
		size, err1 := strconv.ParseInt(fields[2], 10, 64)
		used, err2 := strconv.ParseInt(fields[3], 10, 64)
		if err1 != nil || err2 != nil {
			continue
		}
		kind := fields[1]
		if strings.HasPrefix(fields[0], "/dev/zram") {
			kind = "zram"
		}
		swaps = append(swaps, proto.MemorySwap{Name: fields[0], Kind: kind, Size: size * 1024, Used: used * 1024})
	}
	return swaps
}
