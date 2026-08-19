# Build environment for the out-of-tree zram modules.
#
# build-modules.sh needs three things: the Xuantie cross-compiler that built the
# device kernel, the host packages a 5.10 kernel tree needs to configure and
# link a module, and modinfo to read the vermagic back out.
#
# The cross-compiler comes from nanokvm-app-builder, which already unpacks
# Sophgo's host-tools. Build that image first:
#
#   docker build -t nanokvm-app-builder tools/build
#   docker build -t nanokvm-zram-builder tools/zram
#
# The kernel tree does not belong on a Windows filesystem. Checking out
# linux_5.10 there fails outright, because the tree carries paths Windows
# refuses. Keep the source in a named volume and this image never touches the
# host filesystem for anything but the modules it emits.
FROM nanokvm-app-builder

# gcc-12 is here because the host half of a 5.10 build does not compile clean
# under this image's default gcc-13. Only scripts/ and the module linker run on
# the host; the modules themselves are built by the Xuantie cross-compiler, so
# the host compiler version cannot reach the artefact.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bc bison flex libssl-dev libelf-dev kmod cpio rsync python3 \
        gcc-12 g++-12 \
    && rm -rf /var/lib/apt/lists/* \
    && update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-12 100 \
    && update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-12 100

RUN riscv64-unknown-linux-musl-gcc --version | head -1 \
    && gcc --version | head -1 \
    && modinfo --version
