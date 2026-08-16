# The host release.sh expects, as an image.
#
# release.sh needs Linux with mtools, dosfstools, e2fsprogs, sfdisk, zstd,
# u-boot-tools, cpio, node and a Docker client. A Windows workstation has none
# of them, and a WSL shell needs a root password to install them. This image is
# that host. It drives the workstation's own Docker daemon through the mounted
# socket, which is how the MaixCDK cross-compile still runs.
#
# It must run as root. repack-boot.sh unpacks an initramfs that contains
# dev/console, and mknod is refused to an unprivileged user. The script's own
# check catches that and stops, so the failure is loud, but it stops the release.
#
# See README.md for the command that runs it.
FROM node:22-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
        fdisk util-linux dosfstools mtools e2fsprogs zstd xz-utils \
        u-boot-tools cpio device-tree-compiler openssl ca-certificates \
        coreutils tar gzip file patchelf \
    && rm -rf /var/lib/apt/lists/*

# The client only. The daemon is the workstation's, reached over the socket.
ARG DOCKER_VERSION=27.5.1
RUN curl -fsSL "https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz" \
        -o /tmp/docker.tgz \
    && tar -xzf /tmp/docker.tgz -C /tmp docker/docker \
    && install -m 0755 /tmp/docker/docker /usr/local/bin/docker \
    && rm -rf /tmp/docker.tgz /tmp/docker

# The same version .github/workflows/package.yml installs.
RUN npm install --global pnpm@11

# The checkout is bind-mounted from the workstation and its owner is not root
# inside the container. Without this git refuses it as dubious, and release.sh
# stops at its first check.
RUN git config --system --add safe.directory '*'
