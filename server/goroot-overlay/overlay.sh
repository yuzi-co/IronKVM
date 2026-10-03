#!/bin/sh
# Print the path of a `go build -overlay` file that puts the files under
# goroot-overlay/src over the Go toolchain's own source tree:
#
#   go build -overlay "$(sh goroot-overlay/overlay.sh)" ...
#
# Run from server/, inside the builder image or anywhere with the same Go.
#
# Why this exists (ironkvm-dist#60). Over HTTPS the board spends most of the
# core sealing TLS records. The server negotiates TLS 1.3 with
# ChaCha20-Poly1305, the fastest suite on a core without AES instructions,
# but Go has no riscv64 assembly for either half of it. On the C906 the
# generic Go code seals about 11 MB/s (ChaCha20 20 MB/s, Poly1305 70 MB/s),
# and a 1080p MJPEG picture is about 300 kB.
#
# crypto/tls takes both from the toolchain's vendored copy of
# golang.org/x/crypto (GOROOT/src/vendor), not from the module graph, so a
# go.mod replace cannot reach them. An overlay can: it swaps files in any
# package, the standard library's included, for this build only, without
# touching the toolchain. It adds chacha_riscv64.{go,s} and sum_riscv64.s, and
# replaces chacha_noasm.go, sum_asm.go and mac_noasm.go with the same files
# with riscv64 moved into their build constraints.
#
# The replaced file has to match the toolchain's, or the build would silently
# mix two versions of the package. The check below compares checksums of the
# files the assembly depends on and stops the build when the toolchain has
# moved. When that happens: diff the new chacha_noasm.go and chacha_generic.go
# against the ones these hashes describe, carry any change into the overlay,
# run the test on a device (below), and update the hashes.
#
# The test, on a device:
#
#   o=$(sh goroot-overlay/overlay.sh)
#   for p in chacha20 internal/poly1305; do
#       GOARCH=riscv64 CGO_ENABLED=0 go test -c -overlay "$o" \
#           -o "$(basename $p).test" "vendor/golang.org/x/crypto/$p"
#   done
#   scp chacha20.test poly1305.test root@<device>:/tmp/
#   ssh root@<device> 'for t in /tmp/chacha20.test /tmp/poly1305.test; do $t -test.v -test.bench .; rm $t; done'
#
# Turning it off: NANOKVM_GOROOT_OVERLAY=off makes this print an empty
# overlay, and the build uses the toolchain's own generic Go code. The server
# is then correct and slower over HTTPS; nothing else changes.
#
# Where the code comes from:
#   chacha_riscv64.s   written for NanoKVM, checked against the generic code
#   sum_riscv64.s      Go's own sum_loong64.s (BSD), translated instruction
#                      for instruction
#   the .go files      the toolchain's, with riscv64 added to the constraints
#
# What it was checked with, on the device (C906, both kernels): the tests
# beside the files, x/crypto v0.39.0's own chacha20, chacha20poly1305 and
# poly1305 suites with these files dropped in, crypto/tls -test.short
# built with the overlay, and Wycheproof's ChaCha20-Poly1305 (325 tests) and
# XChaCha20-Poly1305 (315 tests) vectors on the C906 (vendor 5.10 kernel), all
# passing, with chacha_riscv64_active_test.go confirming that the assembly
# was the path in use. The Wycheproof test reads the vector files named by
# WP_CHACHA and WP_XCHACHA (testvectors_v1/ in github.com/C2SP/wycheproof) and
# skips when they are unset:
#
#   GOARCH=riscv64 CGO_ENABLED=0 go test -c -overlay "$o" #       -o aead.test vendor/golang.org/x/crypto/chacha20poly1305
#   WP_CHACHA=chacha20_poly1305_test.json #       WP_XCHACHA=xchacha20_poly1305_test.json ./aead.test -test.v
set -e

if [ "${NANOKVM_GOROOT_OVERLAY:-on}" = off ]; then
    out=$(mktemp "${TMPDIR:-/tmp}/goroot-overlay.XXXXXX")
    echo '{"Replace":{}}' > "$out"
    echo "$out"
    exit 0
fi

here=$(cd "$(dirname "$0")" && pwd)
goroot=$(go env GOROOT)
xc=src/vendor/golang.org/x/crypto

# sha256 of the toolchain files this overlay was written against (Go 1.25.0,
# golang.org/x/crypto v0.39.0): the ones it replaces, and the ones whose
# types and calling conventions the assembly relies on.
check() {
    sum=$(sha256sum "$goroot/$xc/$1" | cut -d' ' -f1)
    if [ "$sum" != "$2" ]; then
        echo "goroot-overlay: $goroot/$xc/$1 is not the file the overlay was written for" >&2
        echo "  have $sum" >&2
        echo "  want $2" >&2
        echo "  see the notes in $0" >&2
        exit 1
    fi
}
check chacha20/chacha_noasm.go bbdb67ceb30ef13efc54e17935f47e00744b96754e03364e606d4f1f8f9085c1
check chacha20/chacha_generic.go 34403e82b1387b4402b00ce30c1364508c333f4bdbe671321690c6ebaa8d3180
check internal/poly1305/sum_asm.go ce5f94aedd0ce3349a8a8655607bd0159d354902e3298169cf9350fe235382b7
check internal/poly1305/mac_noasm.go f5308cd6f14bab1b00963eeae8137d63fc5452db7f7d49276a56f457bfd89d2d
check internal/poly1305/sum_generic.go b0094a2895d5bda42dcaaf57c0b31fc914c3b6c6aa6237aab0100a6d78346933

out=$(mktemp "${TMPDIR:-/tmp}/goroot-overlay.XXXXXX")
{
    printf '{"Replace":{'
    sep=
    for f in $(cd "$here" && find src -type f | sort); do
        printf '%s"%s":"%s"' "$sep" "$goroot/$f" "$here/$f"
        sep=,
    done
    printf '}}\n'
} > "$out"
echo "$out"
