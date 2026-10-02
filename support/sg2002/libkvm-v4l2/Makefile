# libkvm.so for the SG2002 on the mainline kernel (V4L2). See README.md.
#
# In the app builder image (tools/build/Dockerfile), from the repository root:
#
#   docker run --rm -v "$PWD":/w -w /w/support/sg2002/libkvm-v4l2 \
#       nanokvm-app-builder make all test check-symbols
#
# all            libkvm.so and the on-board probe, for riscv64 musl
# test           host unit tests (NAL classification, negotiation against
#                mock drivers, policy, buffer slots), with ASan and UBSan
# check-symbols  the exports against abi-symbols.txt and the vendor library

CROSS ?= riscv64-unknown-linux-musl-
CC := $(CROSS)gcc
NM := $(CROSS)nm
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
	src/kvmv_annexb.c src/kvmv_slots.c
HEADERS := $(wildcard src/*.h) $(SERVER_INCLUDE)/kvm_vision.h

.PHONY: all test check-symbols clean

all: $(BUILD)/libkvm.so $(BUILD)/kvmv-probe

$(BUILD):
	mkdir -p $@

# -z defs: anything not in libc is a link error, not a load-time surprise.
# The soname is what the server records in NEEDED.
$(BUILD)/libkvm.so: $(LIB_SRCS) $(HEADERS) | $(BUILD)
	$(CC) $(ARCH_FLAGS) $(CFLAGS_COMMON) -fPIC -fvisibility=hidden -shared \
		-Wl,-soname,libkvm.so -Wl,--as-needed -Wl,-z,defs \
		-o $@ $(LIB_SRCS) -lpthread

$(BUILD)/kvmv-probe: tools/kvmv_probe.c src/kvmv_annexb.c $(HEADERS) | $(BUILD)
	$(CC) $(ARCH_FLAGS) $(CFLAGS_COMMON) -o $@ tools/kvmv_probe.c \
		src/kvmv_annexb.c -ldl -lpthread

$(BUILD)/test_host: test/test_host.c src/kvmv_annexb.c src/kvmv_policy.c \
		src/kvmv_slots.c src/kvmv_pipeline.c $(HEADERS) | $(BUILD)
	$(HOSTCC) $(CFLAGS_COMMON) -O1 -g -fsanitize=address,undefined \
		-fno-sanitize-recover=all -o $@ test/test_host.c \
		src/kvmv_annexb.c src/kvmv_policy.c src/kvmv_slots.c \
		src/kvmv_pipeline.c

test: $(BUILD)/test_host
	./$(BUILD)/test_host

check-symbols: $(BUILD)/libkvm.so
	sh ./check-symbols.sh $(NM) $(BUILD)/libkvm.so abi-symbols.txt $(VENDOR_LIB)

clean:
	rm -rf $(BUILD)
