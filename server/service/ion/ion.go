package ion

import (
	"os"
	"path/filepath"
	"sync"
)

// Root is the carveout debugfs directory. It is a variable so that tests can
// point it at a fixture directory.
//
// Do not add /proc/cvitek/vb to anything in this package. Reading that file
// blocks the reader forever in uninterruptible sleep, and the reader cannot be
// killed.
var Root = "/sys/kernel/debug/ion/cvi_carveout_heap_dump"

// orphanCost is what one generation leaves behind when it does not release what
// it held. No userspace action returns it. It is the floor under Reserve,
// because a board with less than this free cannot absorb one more.
//
// Only a crash strands memory now. An ordinary stop releases everything it
// took, on every delivery path, since kvmv_deinit stopped decrementing a
// reference count that a served JPEG frame had already pushed out of reach.
// tools/vidiag/venc-leak.cpp measures both libraries through that sequence.
//
// There are two figures, and this is the larger one on purpose.
//
// An idle pipeline strands one VI channel pool and one ISP_SHARED_BUFFER_0 when
// its process dies without a teardown: 6,516,736 bytes, measured four times.
// The dead process leaves 19,050,496, and the next start takes the VI pool back
// - _free_leak_memory_of_vb reclaims a leaked VbPool, and nothing reclaims the
// rest.
//
// A pipeline that has encoded strands 11,636,736: two VENC_1_ReconFrameBuf, one
// VCODEC_H264_FW_Buffer, one VENC_1_BitStreamBuffer, one VENC_1_H264_WorkBuffer,
// the encoder's VbPool and one ISP_SHARED_BUFFER_0. Nothing takes any of that
// back. The vcodec driver builds with CVI_H26X_USE_ION_FW_BUFFER, which drops
// its own release path, and soph_sys never walks its buffer list on close.
//
// Understating this is what lets a grade read "ok" up to the allocation that
// segfaults the server, so the constant carries the cost of the worse case.
// See tools/ionmem/README.md.
const orphanCost = 11636736

// writeFile is a variable so a test can make the peak reset fail while the
// counter stays readable. The container that runs the tests is root, and root
// ignores the write bit, so the failure cannot be staged on the filesystem.
var writeFile = os.WriteFile

// Status is one reading of the carveout.
type Status struct {
	Total       uint64
	Used        uint64
	Free        uint64
	UsageRate   int
	Generations int
	Reserve     uint64
	// Measured is false when Reserve is the configured floor rather than a
	// value this process observed.
	Measured bool
	Verdict  string
}

var (
	mu           sync.Mutex
	allocAtStart uint64
	baselineOK   bool
	peakReset    bool
	reserveFloor uint64
)

// Init records the allocation level at startup and resets the peak watermark,
// so that a later reading measures what this process needed rather than what
// the board happened to hold.
//
// The working set is cumulative over delivery paths: a board that only serves
// screenshots needs less than one that also streams H264. A fixed constant is
// therefore wrong in both directions, and the floor exists only to cover the
// window before this process has captured anything.
//
// Every failure here is survivable. A board whose peak cannot be reset falls
// back to the floor and still reports a graded verdict.
func Init(floor uint64) {
	mu.Lock()
	defer mu.Unlock()

	reserveFloor = floor
	allocAtStart = 0
	baselineOK = false
	peakReset = false

	alloc, err := readCounter("alloc_mem")
	if err != nil {
		return
	}
	allocAtStart = alloc
	baselineOK = true

	if err := writeFile(filepath.Join(Root, "peak"), []byte("0"), 0o644); err == nil {
		peakReset = true
	}
}

// Read takes one reading. It never returns an error: a carveout it cannot read
// is reported as unavailable, and the UI shows nothing.
func Read() Status {
	mu.Lock()
	base, haveBase, reset, floor := allocAtStart, baselineOK, peakReset, reserveFloor
	mu.Unlock()

	total, errTotal := readCounter("total_mem")
	used, errUsed := readCounter("alloc_mem")
	if errTotal != nil || errUsed != nil || total == 0 {
		return Status{Verdict: VerdictUnavailable}
	}

	// total_mem and alloc_mem are two separate, non-atomic reads of files the
	// kernel can update between them, so a torn read can briefly report used >
	// total. Guard and clamp the same way in both derived fields rather than
	// letting one of them show a value a torn read cannot actually produce.
	status := Status{Total: total, Used: used}
	if used <= total {
		status.Free = total - used
		status.UsageRate = int(used * 100 / total)
	} else {
		status.UsageRate = 100
	}

	if body, err := os.ReadFile(filepath.Join(Root, "summary")); err == nil {
		if buffers, err := ParseSummary(string(body)); err == nil {
			status.Generations = CountGenerations(buffers)
		}
	}

	// What has to stay free is what this generation has not allocated yet, and
	// never less than what one crash costs.
	//
	// peak - base is what this generation has already taken since it started.
	// It is spent, not pending: those buffers are held right now and the free
	// space is what is left beside them. Requiring the carveout to hold that
	// much a second time was right only while a dying process kept its buffers,
	// which it did until the teardown was fixed on 2026-08-20. A clean stop
	// releases them before the next generation asks for them, so counting the
	// working set as a future cost double-counts it.
	//
	// Measured on this board after the fix: a fully exercised generation holds
	// 42,942,464 bytes and the stop gives all of it back. Grading it against
	// its own high-water mark called a healthy board critical, and on the 75MB
	// carveout the same arithmetic had already been calling it amber.
	//
	// So the reserve is what is still ahead of this generation: the floor, less
	// whatever of it the generation has already spent. A board that has not
	// opened a stream still needs room to open one. A board that has opened
	// every stream needs room for none of it again.
	//
	// The orphan cost is the part that never goes away. A process that dies
	// without running its teardown - a SIGKILL, a segfault, the dispose
	// timeout expiring - still leaves one VI channel pool and one ISP shared
	// buffer behind, and nothing but a reboot returns those. A board with less
	// than that free is one crash from a carveout it cannot recover, whatever
	// it has or has not allocated, so the reserve never falls below it.
	status.Reserve = orphanCost
	if floor > orphanCost {
		status.Reserve = floor
	}

	if haveBase && reset {
		if peak, err := readCounter("peak"); err == nil && peak >= base {
			spent := peak - base
			pending := uint64(0)
			if floor > spent {
				pending = floor - spent
			}
			status.Reserve = pending
			if orphanCost > pending {
				status.Reserve = orphanCost
			}
			status.Measured = true
		}
	}

	status.Verdict = Verdict(status.Free, status.Reserve)
	return status
}

func readCounter(name string) (uint64, error) {
	body, err := os.ReadFile(filepath.Join(Root, name))
	if err != nil {
		return 0, err
	}
	return ParseCounter(string(body))
}
