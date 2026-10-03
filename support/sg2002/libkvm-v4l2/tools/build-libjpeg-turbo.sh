#!/bin/sh
# Build libjpeg-turbo's static TurboJPEG library for linking into libkvm.so.
#
#   build-libjpeg-turbo.sh OUT_DIR
#
# Environment:
#   CC          compiler (default cc); the riscv64 musl cross compiler for the
#               board, the host compiler for the unit test
#   CFLAGS_ARCH extra flags, e.g. the -mcpu/-march flags the library uses
#   SYSTEM      cmake CMAKE_SYSTEM_PROCESSOR for a cross build (e.g. riscv64);
#               empty for a host build
#   TURBO_TARBALL  where the source tarball is kept (default: beside OUT_DIR);
#               downloaded when missing, checked against SHA256 either way
#
# Produces OUT_DIR/prefix/{include,lib/libturbojpeg.a}. The objects are
# position independent, since they end up in a shared library, and built with
# hidden visibility, so libkvm.so does not export TurboJPEG's API.
#
# Needs cmake (not in the app builder image before cmake was added to
# tools/build/Dockerfile; `apt-get install -y cmake` in an older one).
#
# libjpeg-turbo has no SIMD for RISC-V, so WITH_SIMD is off and it runs its C
# paths. See ironkvm-dist socs/sophgo-sg2002/mainline/jpeg-bench for the
# measurements behind choosing it (TurboJPEG API, fast DCT).
set -eu

VERSION=3.1.2
SHA256=8f0012234b464ce50890c490f18194f913a7b1f4e6a03d6644179fa0f867d0cf
URL=https://github.com/libjpeg-turbo/libjpeg-turbo/releases/download/$VERSION/libjpeg-turbo-$VERSION.tar.gz

out=$1
cc=${CC:-cc}
flags=${CFLAGS_ARCH:-}
system=${SYSTEM:-}

if ! command -v cmake >/dev/null 2>&1; then
	echo "build-libjpeg-turbo.sh: cmake not found (apt-get install -y cmake)" >&2
	exit 1
fi

mkdir -p "$out"
out=$(cd "$out" && pwd)
tarball=${TURBO_TARBALL:-$(dirname "$out")/libjpeg-turbo-$VERSION.tar.gz}
if [ ! -f "$tarball" ]; then
	wget -q -O "$tarball.part" "$URL"
	mv "$tarball.part" "$tarball"
fi
echo "$SHA256  $tarball" | sha256sum -c - >/dev/null

rm -rf "$out/src" "$out/cmake" "$out/prefix"
mkdir -p "$out/src"
tar -C "$out/src" --strip-components=1 -xzf "$tarball"

set -- -S "$out/src" -B "$out/cmake" \
	-DCMAKE_C_COMPILER="$cc" \
	-DCMAKE_BUILD_TYPE=Release \
	-DCMAKE_C_FLAGS_RELEASE="-O2 -DNDEBUG $flags -fvisibility=hidden" \
	-DCMAKE_POSITION_INDEPENDENT_CODE=ON \
	-DENABLE_SHARED=0 -DENABLE_STATIC=1 -DWITH_TURBOJPEG=1 \
	-DWITH_SIMD=0 -DWITH_TOOLS=0 -DWITH_TESTS=0 \
	-DCMAKE_INSTALL_PREFIX="$out/prefix" -DCMAKE_INSTALL_LIBDIR=lib
if [ -n "$system" ]; then
	set -- "$@" -DCMAKE_SYSTEM_NAME=Linux -DCMAKE_SYSTEM_PROCESSOR="$system"
fi
cmake "$@" >"$out/cmake.log" 2>&1 || { tail -20 "$out/cmake.log" >&2; exit 1; }
cmake --build "$out/cmake" --target turbojpeg-static -j"$(nproc)" >"$out/build.log" 2>&1 ||
	{ tail -20 "$out/build.log" >&2; exit 1; }
mkdir -p "$out/prefix/lib" "$out/prefix/include"
cp "$out/cmake/libturbojpeg.a" "$out/prefix/lib/"
cp "$out/src/src/turbojpeg.h" "$out/prefix/include/" 2>/dev/null ||
	cp "$out/src/turbojpeg.h" "$out/prefix/include/"
echo "libjpeg-turbo $VERSION: $out/prefix/lib/libturbojpeg.a"
