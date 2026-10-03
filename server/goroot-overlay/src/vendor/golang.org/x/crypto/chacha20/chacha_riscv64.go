// NanoKVM: ChaCha20 in riscv64 assembly, for the TLS record layer. See
// server/goroot-overlay/overlay.sh for why this lives in the toolchain's
// vendor tree.

//go:build gc && !purego

package chacha20

import (
	"bytes"
	"os"
)

// bufSize is four blocks, as on arm64: a 32-byte Poly1305 key costs four
// blocks rather than one, which is noise beside a 16 kB TLS record.
const bufSize = 256

// useXTheadBb selects the assembly, which needs the T-Head rotate instruction
// (th.srriw, XTheadBb). It takes ChaCha20 from 20 to 28 MB/s (the generic Go
// code) to 43 to 54 MB/s on the SG2002's C906, measured on the 5.10 and 7.2
// kernels. RV64GC has no rotate; the same assembly with a shift-shift-or in
// its place ran no faster than the generic code, so any other core keeps the
// generic code.
var useXTheadBb = isTHeadC9xx()

// isTHeadC9xx reports a T-Head C906 or C910, the cores that implement
// XTheadBb. Executing th.srriw on any other core raises SIGILL, which the Go
// runtime turns into a crash rather than a panic, so the check has to be
// certain without trying the instruction.
//
// Linux has no flag for XTheadBb in /proc/cpuinfo or hwprobe. The device
// tree says it two ways:
//
//   - mainline: /cpus/cpu@0 is compatible with "thead,c906" (or c910)
//   - the Sophgo 5.10 kernel: the CPU node says only "riscv", the board is
//     compatible with "cvitek,cv18..." (CV180x, CV181x, SG200x), all of
//     which have a C906 as their RISC-V core
//
// Both kernels run on firmware that sets MXSTATUS.THEADISAEE, which these
// instructions need: Linux itself uses T-Head cache instructions on the C906
// and would not boot without it.
func isTHeadC9xx() bool {
	if cpu, err := os.ReadFile("/proc/device-tree/cpus/cpu@0/compatible"); err == nil {
		for _, c := range bytes.Split(cpu, []byte{0}) {
			if bytes.Equal(c, []byte("thead,c906")) || bytes.Equal(c, []byte("thead,c910")) {
				return true
			}
		}
	}
	if board, err := os.ReadFile("/proc/device-tree/compatible"); err == nil {
		for _, c := range bytes.Split(board, []byte{0}) {
			if bytes.HasPrefix(c, []byte("cvitek,cv18")) {
				return true
			}
		}
	}
	return false
}

//go:noescape
func xorKeyStreamTH(dst, src []byte, key *[8]uint32, nonce *[3]uint32, counter *uint32)

func (c *Cipher) xorKeyStreamBlocks(dst, src []byte) {
	if useXTheadBb {
		xorKeyStreamTH(dst, src, &c.key, &c.nonce, &c.counter)
	} else {
		c.xorKeyStreamBlocksGeneric(dst, src)
	}
}
