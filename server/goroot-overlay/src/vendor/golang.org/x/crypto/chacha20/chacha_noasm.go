// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// NanoKVM: the toolchain's file with riscv64 added to the architectures that
// have their own xorKeyStreamBlocks. See server/goroot-overlay/overlay.sh.

//go:build (!arm64 && !s390x && !ppc64 && !ppc64le && !riscv64) || !gc || purego

package chacha20

const bufSize = blockSize

func (s *Cipher) xorKeyStreamBlocks(dst, src []byte) {
	s.xorKeyStreamBlocksGeneric(dst, src)
}
