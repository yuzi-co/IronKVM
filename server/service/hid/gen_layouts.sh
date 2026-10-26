#!/bin/sh
# Regenerate layout_gen.go and layout_xkb_test.go from xkeyboard-config.
#
#   sh server/service/hid/gen_layouts.sh
#   DOCKER=wslc sh server/service/hid/gen_layouts.sh
#
# The generator needs libxkbcommon and the xkb data, so it runs in a container.
# The image and the three Debian packages are pinned: a different
# xkeyboard-config can move a character to another key, and the tables must
# change only when this file does. If Debian drops a pinned version, the
# install fails rather than reading another one. Raise the pins together and
# read the diff of the two generated files before you commit it.
set -eu

IMAGE=golang:1.25.14@sha256:699337d620559a59b4a2bb298ad59611e535d2ee755a34cf2d2a98f37578dc80
XKB_DATA_VERSION=2.42-1        # xkb-data, which is xkeyboard-config 2.42
XKBCOMMON_VERSION=1.7.0-2      # libxkbcommon-dev
X11_DATA_VERSION=2:1.8.12-1    # libx11-data, the compose tables

DOCKER=${DOCKER:-docker}
HERE=$(cd "$(dirname "$0")" && pwd)
ROOT=$(cd "$HERE/../../.." && { pwd -W 2>/dev/null || pwd; })

MSYS_NO_PATHCONV=1 "$DOCKER" run --rm \
    -v "$ROOT:/repo" -w /repo/server/service/hid \
    -e CGO_ENABLED=1 \
    -e XKB_DATA_VERSION="${XKB_DATA_VERSION%-*}" \
    -e XKBCOMMON_VERSION="${XKBCOMMON_VERSION%-*}" \
    "$IMAGE" sh -euc "
        export DEBIAN_FRONTEND=noninteractive
        apt-get update -qq
        apt-get install -y -qq --no-install-recommends \
            xkb-data=$XKB_DATA_VERSION \
            libxkbcommon-dev=$XKBCOMMON_VERSION \
            libx11-data=$X11_DATA_VERSION >/dev/null
        go run gen_layouts.go
    "
