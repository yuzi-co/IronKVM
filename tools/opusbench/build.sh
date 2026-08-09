#!/bin/bash
# Rebuild libopus for the NanoKVM's C906 core, install it into the server tree,
# and build the measurement harness.
#
# Run this from the repository root, inside the builder image:
#
#   MSYS_NO_PATHCONV=1 docker run --rm -v "$PWD:/home/build/NanoKVM" \
#     nanokvm-builder-local-$(id -u)-$(id -g) \
#     bash /home/build/NanoKVM/tools/opusbench/build.sh
#
# The archive and the headers it installs are committed. The benchmark binary
# is not: it is built when somebody wants to measure.
set -e

VERSION=1.5.2
SHA256=65c1d2f78b9f2fb20082c38cbe47c951ad5839345876e46941612ee87f9a7ce1

HOST=riscv64-unknown-linux-musl
CC=$HOST-gcc

# The same flags the Go cgo build uses in the repository Makefile. The board
# has a hardware FPU, so libopus is built in floating point. A fixed-point
# build was measured at 9.78% of the core against 7.88% for float, at
# complexity 5, stereo, 96 kbit/s.
ARCH_FLAGS="-mcpu=c906fdv -march=rv64imafdcv0p7xthead -mcmodel=medany -mabi=lp64d"

root=$(cd "$(dirname "$0")/../.." && pwd)
work=$root/tools/opusbench/.build

mkdir -p "$work"
cd "$work"

if [ ! -f "opus-$VERSION.tar.gz" ]; then
    curl -sSL -o "opus-$VERSION.tar.gz" \
        "https://downloads.xiph.org/releases/opus/opus-$VERSION.tar.gz"
fi

echo "$SHA256  opus-$VERSION.tar.gz" | sha256sum -c -

rm -rf "opus-$VERSION" build inst
tar xzf "opus-$VERSION.tar.gz"

mkdir build
cd build

../opus-$VERSION/configure \
    --host="$HOST" \
    --prefix="$work/inst" \
    --disable-shared --enable-static \
    --disable-doc --disable-extra-programs \
    CFLAGS="-O3 $ARCH_FLAGS" >configure.log 2>&1

make -j"$(nproc)" >make.log 2>&1
make install >>make.log 2>&1

cd "$work"

install -D -m 644 inst/lib/libopus.a "$root/server/dl_lib/libopus.a"
for header in opus.h opus_defines.h opus_types.h; do
    install -D -m 644 "inst/include/opus/$header" "$root/server/include/opus/$header"
done

$CC -O3 $ARCH_FLAGS -static \
    -I "$root/server/include/opus" \
    "$root/tools/opusbench/opusbench.c" \
    "$root/server/dl_lib/libopus.a" \
    -lm -o "$root/tools/opusbench/opusbench"

echo "installed server/dl_lib/libopus.a ($(stat -c %s "$root/server/dl_lib/libopus.a") bytes)"
echo "built tools/opusbench/opusbench ($(stat -c %s "$root/tools/opusbench/opusbench") bytes)"
