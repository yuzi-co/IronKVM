FROM ubuntu:24.04 AS base_apt

# RUN echo 'Acquire::http::Proxy "http://cache.lan:3142";' > /etc/apt/apt.conf.d/00cacher
RUN userdel -r ubuntu
RUN apt update \
    && apt install wget git cmake build-essential python3 python3-venv python3-pip autoconf automake libtool gosu patchelf -y

ARG DOCKER_USER=build
ARG DOCKER_UID=1000
ARG DOCKER_GID=1000
# Reuse existing accounts for numeric IDs that are already present in the
# base image. All later ownership and user switches use numeric IDs, so a
# dedicated build user/group is only needed when the ID does not exist.
RUN set -eux; \
    if ! getent group "$DOCKER_GID" >/dev/null; then \
        groupadd --gid "$DOCKER_GID" "$DOCKER_USER"; \
    fi; \
    if ! getent passwd "$DOCKER_UID" >/dev/null; then \
        useradd --uid "$DOCKER_UID" --gid "$DOCKER_GID" --no-create-home \
            --home-dir "/home/$DOCKER_USER" --shell /bin/bash "$DOCKER_USER"; \
    fi; \
    install -d -o "$DOCKER_UID" -g "$DOCKER_GID" "/home/$DOCKER_USER"
USER $DOCKER_UID:$DOCKER_GID

FROM base_apt AS host_tools

# The URL carries no version and the server publishes no checksum, so this
# sha256 is the only thing that ties a build to a known toolchain. It was taken
# on 2026-08-29 by streaming the 840MB download through sha256sum. If it stops
# matching, do not raise it to whatever arrived: find out what changed at that
# URL first.
ARG HOST_TOOLS_URL=https://sophon-file.sophon.cn/sophon-prod-s3/drive/23/03/07/16/host-tools.tar.gz
ARG HOST_TOOLS_SHA256=ff9a58e8e192b20ea42e1d729c42d2219523209706cb3f0cf134582f6c70f805

USER root
RUN wget -O /tmp/host-tools.tar.gz "$HOST_TOOLS_URL" \
    && echo "$HOST_TOOLS_SHA256  /tmp/host-tools.tar.gz" | sha256sum -c - \
    && tar -C /usr/local -zxvf /tmp/host-tools.tar.gz  \
    && rm -f /tmp/host-tools.tar.gz \
    && export PATH=$PATH:/usr/local/host-tools/gcc/riscv64-linux-musl-x86_64/bin \
    && echo "Installed host-tools gcc" && riscv64-unknown-linux-musl-gcc -v

FROM base_apt AS golang

ARG GO_VERSION=1.25.0
# One sum for each architecture this image is built on. An architecture that is
# not listed stops the build. A silent skip is the failure this whole change
# exists to remove, so there is no fallback branch here.
ARG GO_SHA256_AMD64=2852af0cb20a13139b3448992e69b868e50ed0f8a1e5940ee1de9e19a123b613
ARG GO_SHA256_ARM64=05de75d6994a2783699815ee553bd5a9327d8b79991de36e38b66862782f54ae

USER root
# The mapping is explicit because `uname -m` does not name what go.dev
# publishes. The previous expression passed "aarch64" through unchanged and
# asked for a tarball that has never existed.
RUN set -eu; \
    case "$(uname -m)" in \
        x86_64)  goarch=amd64; gosum="$GO_SHA256_AMD64" ;; \
        aarch64) goarch=arm64; gosum="$GO_SHA256_ARM64" ;; \
        *) echo "no pinned go checksum for $(uname -m)" >&2; exit 1 ;; \
    esac; \
    wget -O /tmp/go.tar.gz "https://go.dev/dl/go${GO_VERSION}.linux-${goarch}.tar.gz"; \
    echo "$gosum  /tmp/go.tar.gz" | sha256sum -c -; \
    tar -C /usr/local -zxvf /tmp/go.tar.gz; \
    rm -f /tmp/go.tar.gz; \
    export PATH=$PATH:/usr/local/go/bin; \
    echo "Installed golang"; go version

COPY server /tmp/go_cache

ENV PATH="$PATH:/usr/local/go/bin"

RUN cd /tmp/go_cache \
    && echo "Caching go dep" \
    && go mod download \
    && go mod tidy \
    && cd / \
    && rm -rf /tmp/go_cache

FROM base_apt AS sdk

ENV HOME="/home/$DOCKER_USER"

# Pinned. A bare clone of the default branch made this image unreproducible:
# two builds of the same commit of this repository could link libkvm.so against
# different SDK sources, and nothing would report it.
#
# This commit is the one that built the libkvm.so committed under
# server/dl_lib. It is MaixCDK's head from 2026-08-20, the commit before it is
# from 2026-07-17, and the library was rebuilt on 2026-08-28, so no other
# commit can have produced it. The nearest release tag is v4.11.3.
ARG MAIXCDK_COMMIT=30f4b8b7e3f66ded9cfa64fb081b46e30bc27248

RUN cd ~ \
    && git clone https://github.com/Sipeed/MaixCDK \
    && cd MaixCDK \
    && git checkout --detach "$MAIXCDK_COMMIT" \
    && python3 -m venv . \
    && . ./bin/activate \
    && pip install -U -r requirements.txt

COPY --chown=$DOCKER_UID:$DOCKER_GID . /home/$DOCKER_USER/NanoKVM

RUN cd ~/MaixCDK \
    && . ./bin/activate \
    && cd ~/NanoKVM/support/sg2002 \
    && ./build kvm_system

FROM base_apt

# Install golang
COPY --from=golang /usr/local/go /usr/local/go
COPY --chown=$DOCKER_UID:$DOCKER_GID --from=golang /root/go /home/$DOCKER_USER/go
ENV PATH="$PATH:/usr/local/go/bin"

# Install host tools
COPY --from=host_tools /usr/local/host-tools /usr/local/host-tools
ENV PATH="$PATH:/usr/local/host-tools/gcc/riscv64-linux-musl-x86_64/bin"

# Install SDK
COPY --chown=$DOCKER_UID:$DOCKER_GID --from=sdk /home/$DOCKER_USER/MaixCDK /home/$DOCKER_USER/MaixCDK
COPY --chmod=+x docker/entrypoint /entrypoint

RUN echo "Verify build tools" \
    && echo "Installed host-tools gcc" && riscv64-unknown-linux-musl-gcc -v \
    && echo "Installed golang" && go version

USER root
RUN chmod 0755 /entrypoint \
 && sed "s/\$DOCKER_USER/$DOCKER_USER/g" -i /entrypoint

ENTRYPOINT ["/entrypoint"]
