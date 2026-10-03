#!/bin/sh
# Check that every server build entry point builds with the goroot overlay,
# the riscv64 ChaCha20 and Poly1305 for crypto/tls (server/goroot-overlay,
# ironkvm-dist#60).
#
#   test-goroot-overlay.sh
#
# Run it from the repository root. The cases that run overlay.sh need Go on
# the PATH; without it they are skipped and the script exits 2.
#
# A binary is deployed from whichever entry point the operator reached for, so
# one that left the overlay out would ship a server that does MJPEG over HTTPS
# at about half the frame rate, and nothing would say so.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

echo "===== every build entry point passes the overlay ====="

build_lines() {
    grep -n '[^a-z]go build ' "$ROOT/$1" | grep -v '^[0-9]*:[[:space:]]*#' | grep -v 'echo'
}

for file in Makefile server/build.sh tools/build/build-app.sh tools/release/release.sh; do
    lines=$(build_lines "$file" | grep -c . || true)
    with=$(build_lines "$file" | grep -c -- '-overlay ' || true)
    [ "$lines" -gt 0 ] && [ "$with" = "$lines" ] \
        && note "$file passes -overlay on every build line" OK \
        || note "$file passes -overlay on every build line ($with/$lines)" FAIL
done

if ! command -v go >/dev/null 2>&1; then
    echo
    if [ "$fails" -ne 0 ]; then
        echo "$fails case(s) FAILED"
        exit 1
    fi
    echo "the source cases passed; go is not on the PATH, so overlay.sh did not run"
    exit 2
fi

echo
echo "===== overlay.sh ====="

cd "$ROOT/server"
if out=$(sh goroot-overlay/overlay.sh 2>"${TMPDIR:-/tmp}/overlay.err"); then
    n=$(grep -o '"[^"]*":"[^"]*"' "$out" | grep -c . || true)
    [ "$n" = "$(find goroot-overlay/src -type f | grep -c .)" ] \
        && note "it maps every file under goroot-overlay/src ($n)" OK \
        || note "it maps every file under goroot-overlay/src ($n)" FAIL
    rm -f "$out"
else
    # The toolchain moved under the overlay. overlay.sh says what to do.
    note "the toolchain is the one the overlay was written for" FAIL
    sed 's/^/    /' "${TMPDIR:-/tmp}/overlay.err"
fi

out=$(NANOKVM_GOROOT_OVERLAY=off sh goroot-overlay/overlay.sh)
grep -qx '{"Replace":{}}' "$out" \
    && note "NANOKVM_GOROOT_OVERLAY=off gives an empty overlay" OK \
    || note "NANOKVM_GOROOT_OVERLAY=off gives an empty overlay" FAIL
rm -f "$out"

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
