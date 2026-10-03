# libkvm.so for the SG2002 on the mainline kernel (V4L2). See README.md.
#
# In the app builder image (tools/build/Dockerfile), from the repository root:
#
#   docker run --rm -v "$PWD":/w -w /w/support/sg2002/libkvm-v4l2 \
#       nanokvm-app-builder make all test check-symbols
#
# all            libkvm.so and the on-board probe, for riscv64 musl
# test           host unit tests (NAL classification, negotiation against
#                mock drivers, policy, buffer slots, JPEG), with ASan and UBSan
# check-symbols  the exports against abi-symbols.txt, abi-extensions.txt and
#                the vendor library, and the NEEDED entries
# distclean      also remove deps/ (the libjpeg-turbo builds)
#
# The first build fetches and builds libjpeg-turbo, which needs cmake.

CROSS ?= riscv64-unknown-linux-musl-
CC := $(CROSS)gcc
NM := $(CROSS)nm
READELF := $(CROSS)readelf
HOSTCC ?= gcc

# The flags the server's cgo build uses for this core.
ARCH_FLAGS ?= -mcpu=c906fdv -march=rv64imafdcv0p7xthead -mcmodel=medany -mabi=lp64d

ROOT := ../../..
SERVER_INCLUDE := $(ROOT)/server/include
VENDOR_LIB := $(ROOT)/server/dl_lib/libkvm.so
BUILD := build

WARN := -Wall -Wextra -Wformat=2 -Wshadow -Wno-unused-parameter
CFLAGS_COMMON := -std=gnu11 -O2 $(WARN) -I$(SERVER_INCLUDE) -Isrc

LIB_SRCS := src/kvm_vision_v4l2.c src/kvmv_pipeline.c src/kvmv_policy.c \
	src/kvmv_annexb.c src/kvmv_slots.c src/kvmv_jpeg.c src/kvmv_hwjpeg.c
HEADERS := $(wildcard src/*.h) $(SERVER_INCLUDE)/kvm_vision.h

# libjpeg-turbo's TurboJPEG, built from a pinned release by
# tools/build-libjpeg-turbo.sh (needs cmake): once for the board, linked
# statically into libkvm.so with its symbols hidden, and once for the host
# test. Kept in deps/, which `make clean` leaves alone.
DEPS := deps
TURBO := $(DEPS)/turbo/prefix
TURBO_HOST := $(DEPS)/turbo-host/prefix
TURBO_TARBALL := $(CURDIR)/$(DEPS)/libjpeg-turbo-3.1.2.tar.gz

.PHONY: all test check-symbols clean distclean

all: $(BUILD)/libkvm.so $(BUILD)/kvmv-probe

$(BUILD):
	mkdir -p $@

$(TURBO)/lib/libturbojpeg.a: tools/build-libjpeg-turbo.sh
	CC=$(CC) CFLAGS_ARCH="$(ARCH_FLAGS)" SYSTEM=riscv64 \
		TURBO_TARBALL=$(TURBO_TARBALL) sh tools/build-libjpeg-turbo.sh $(DEPS)/turbo

$(TURBO_HOST)/lib/libturbojpeg.a: tools/build-libjpeg-turbo.sh
	CC=$(HOSTCC) TURBO_TARBALL=$(TURBO_TARBALL) \
		sh tools/build-libjpeg-turbo.sh $(DEPS)/turbo-host

# -z defs: anything not in libc is a link error, not a load-time surprise.
# The soname is what the server records in NEEDED.
#
# libgcc_s is linked with --no-as-needed although nothing in the library calls
# it. NanoKVM-Server's crtbegin calls __register_frame_info and
# __deregister_frame_info through the PLT. They are weak there, but the server
# is linked against Sipeed's library, which NEEDs libgcc_s, so the linker found
# them and gave them PLT slots; the server itself NEEDs only libkvm.so and
# libc.so. Loaded with a library that does not bring libgcc_s in, musl resolves
# the slots to 0 and the server jumps to address 0 at start.
# check-symbols.sh fails a library without this entry.
#
# --exclude-libs keeps TurboJPEG's API out of the exports even if a symbol
# escaped -fvisibility=hidden; check-symbols.sh fails any unlisted export.
$(BUILD)/libkvm.so: $(LIB_SRCS) $(HEADERS) $(TURBO)/lib/libturbojpeg.a | $(BUILD)
	$(CC) $(ARCH_FLAGS) $(CFLAGS_COMMON) -I$(TURBO)/include -fPIC -fvisibility=hidden -shared \
		-Wl,-soname,libkvm.so -Wl,--as-needed -Wl,-z,defs -Wl,--exclude-libs,ALL \
		-o $@ $(LIB_SRCS) $(TURBO)/lib/libturbojpeg.a -lpthread \
		-Wl,--no-as-needed -lgcc_s -Wl,--as-needed

$(BUILD)/kvmv-probe: tools/kvmv_probe.c src/kvmv_annexb.c $(HEADERS) | $(BUILD)
	$(CC) $(ARCH_FLAGS) $(CFLAGS_COMMON) -o $@ tools/kvmv_probe.c \
		src/kvmv_annexb.c -ldl -lpthread

$(BUILD)/test_host: test/test_host.c src/kvmv_annexb.c src/kvmv_policy.c \
		src/kvmv_slots.c src/kvmv_pipeline.c src/kvmv_jpeg.c src/kvmv_hwjpeg.c $(HEADERS) \
		$(TURBO_HOST)/lib/libturbojpeg.a | $(BUILD)
	$(HOSTCC) $(CFLAGS_COMMON) -I$(TURBO_HOST)/include -O1 -g \
		-fsanitize=address,undefined -fno-sanitize-recover=all -o $@ \
		test/test_host.c src/kvmv_annexb.c src/kvmv_policy.c src/kvmv_slots.c \
		src/kvmv_pipeline.c src/kvmv_jpeg.c src/kvmv_hwjpeg.c $(TURBO_HOST)/lib/libturbojpeg.a

test: $(BUILD)/test_host
	./$(BUILD)/test_host

check-symbols: $(BUILD)/libkvm.so
	READELF=$(READELF) EXTENSIONS=abi-extensions.txt \
		sh ./check-symbols.sh $(NM) $(BUILD)/libkvm.so abi-symbols.txt $(VENDOR_LIB)

clean:
	rm -rf $(BUILD)

distclean: clean
	rm -rf $(DEPS)
