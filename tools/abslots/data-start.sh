#!/bin/sh
# Print the sector the data partition starts at, derived from the table.
#
#   data-start.sh [path-to-partition.sfdisk]
#
# The shipped table declares p1 to p5 and stops. The data partition is made on
# the first boot of a card, at whatever size that card allows, so its start is a
# derived number and not a table entry. Three places need it and none of them may
# keep its own copy: build-card.sh sizes the image with it, build-image.sh writes
# it into /etc/nanokvm-slots.conf, and S01fs reads it from there.
#
#   start = the end of the recovery slot + 8192
#
# 8192 sectors is 4 MiB. It is the gap the table already leaves ahead of a
# logical partition for its EBR, and 4 MiB is the erase block size that matters
# on this card. The rule gives 10543104, which is where the data partition sat in
# the fixed table this replaced. A card made either way has one layout.
set -eu

TABLE=${1:-$(dirname "$0")/partition.sfdisk}
[ -f "$TABLE" ] || { echo "no such table: $TABLE" >&2; exit 1; }

field() {
    sed -n "s/^$1 *: *.*$2=\([0-9]*\).*/\1/p" "$TABLE" | head -1
}

# A table that still declares a data partition is a table from before this
# change. Refuse it rather than print a number that contradicts it.
if [ -n "$(field 6 start)" ]; then
    echo "$TABLE declares partition 6, and the data partition is made on the device" >&2
    exit 1
fi

START=$(field 5 start)
SIZE=$(field 5 size)
if [ -z "$START" ] || [ -z "$SIZE" ]; then
    echo "$TABLE does not describe partition 5" >&2
    exit 1
fi

echo $((START + SIZE + 8192))
