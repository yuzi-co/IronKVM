# App-only cross-compile toolchain for NanoKVM-Server.
#
# This is the golang + host_tools half of docker/Dockerfile without the MaixCDK
# stage. That stage exists to build kvm_system (`make support`) and is a large
# clone and C++ build, neither of which is needed when only the Go server
# changed. patchelf is included because the RPATH patch is required and
# `make app` does not perform it.
FROM ubuntu:24.04

ARG GO_VERSION=1.25.0
ARG HOST_TOOLS_URL=https://sophon-file.sophon.cn/sophon-prod-s3/drive/23/03/07/16/host-tools.tar.gz
# Both sums must stay equal to the ones in docker/Dockerfile. The two images
# download the same two archives, and an image that verifies while the other
# does not is the same hole with one more place to look for it.
ARG HOST_TOOLS_SHA256=ff9a58e8e192b20ea42e1d729c42d2219523209706cb3f0cf134582f6c70f805
ARG GO_SHA256_AMD64=2852af0cb20a13139b3448992e69b868e50ed0f8a1e5940ee1de9e19a123b613

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        wget ca-certificates build-essential patchelf git \
    && rm -rf /var/lib/apt/lists/*

RUN wget -q -O /tmp/host-tools.tar.gz "$HOST_TOOLS_URL" \
    && echo "$HOST_TOOLS_SHA256  /tmp/host-tools.tar.gz" | sha256sum -c - \
    && tar -C /usr/local -zxf /tmp/host-tools.tar.gz \
    && rm -f /tmp/host-tools.tar.gz

# This image is amd64 only, which is why one sum is enough here and
# docker/Dockerfile needs a case statement.
RUN wget -q -O /tmp/go.tar.gz "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" \
    && echo "$GO_SHA256_AMD64  /tmp/go.tar.gz" | sha256sum -c - \
    && tar -C /usr/local -zxf /tmp/go.tar.gz \
    && rm -f /tmp/go.tar.gz

ENV PATH="/usr/local/go/bin:/usr/local/host-tools/gcc/riscv64-linux-musl-x86_64/bin:${PATH}"
ENV GOFLAGS="-buildvcs=false"

RUN riscv64-unknown-linux-musl-gcc -v && go version && patchelf --version
