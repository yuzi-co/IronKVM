#!/bin/sh
# Check that kvm_system's update path seeds the screen settings only where none
# is kept.
#
#   test-screen-defaults.sh [path/to/kvm_system/main]
#
# Not destructive: everything happens in a temporary directory.
#
# new_app_init ran "echo 2000 > /kvmapp/kvm/qlty" and the same for fps, res and
# type on every tarball update. S95nanokvm links those paths to /etc/kvm/screen
# on /data before kvm_system starts, so each update reset the owner's saved
# settings (ironkvm-dist run sheet, trial 68). screen_defaults.cpp writes a
# default only into a file that is missing, empty or unusable. It has no
# MaixCDK dependency, so it builds here with the host compiler; the rest of
# kvm_system does not, and the last section only reads it.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
MAIN=${1:-$ROOT/support/sg2002/kvm_system/main}
SRC=$MAIN/lib/system_init/screen_defaults.cpp
INIT=$MAIN/lib/system_init/system_init.cpp

[ -f "$SRC" ] || { echo "missing: $SRC"; exit 2; }
[ -f "$INIT" ] || { echo "missing: $INIT"; exit 2; }
CXX=${CXX:-g++}
command -v "$CXX" > /dev/null 2>&1 || { echo "no C++ compiler ($CXX); cannot run here"; exit 2; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }
check() { if eval "$2"; then note "$1" OK; else note "$1" FAIL; fi; }

cat > "$work/t.cpp" <<'EOF'
#include "screen_defaults.h"
#include <stdio.h>

int main(int argc, char **argv)
{
	if (argc != 2) return 2;
	int n = seed_screen_defaults(argv[1]);
	fprintf(stderr, "%d\n", n);
	return n < 0 ? 1 : 0;
}
EOF

if ! "$CXX" -std=gnu++11 -Wall -Wextra -Werror -I"$MAIN/lib/system_init" \
        -o "$work/t" "$work/t.cpp" "$SRC" 2> "$work/cc.log"; then
    cat "$work/cc.log"
    echo "FAILED: screen_defaults.cpp does not build"
    exit 1
fi

# seed DIR: run it and keep the count it answers in $count.
seed() { count=$("$work/t" "$1" 2>&1 > /dev/null); }
# What a file holds, newline and all, as one line of hex.
bytes() { od -An -tx1 "$1" | tr -d ' \n'; }

echo "===== a fresh board gets the defaults ====="
d=$work/fresh; mkdir -p "$d"
seed "$d"
check "four files written" '[ "$count" = 4 ]'
check "qlty 3000" '[ "$(cat "$d/qlty")" = 3000 ]'
check "fps 30" '[ "$(cat "$d/fps")" = 30 ]'
check "res 720" '[ "$(cat "$d/res")" = 720 ]'
check "type h264" '[ "$(cat "$d/type")" = h264 ]'
check "written as the old redirect wrote them, with a newline" \
    '[ "$(bytes "$d/qlty")" = 333030300a ]'
seed "$d"
check "a second run writes nothing" '[ "$count" = 0 ]'

echo "===== an owner's settings are kept byte for byte ====="
d=$work/owner; mkdir -p "$d"
# As the server writes them: bare, no newline.
printf 'h264' > "$d/type"; printf '5000' > "$d/qlty"
printf '1080' > "$d/res";  printf '60' > "$d/fps"
for f in type qlty res fps; do bytes "$d/$f" > "$work/before.$f"; done
seed "$d"
check "nothing written" '[ "$count" = 0 ]'
for f in type qlty res fps; do
    check "$f unchanged" '[ "$(bytes "$d/$f")" = "$(cat "$work/before.$f")" ]'
done
d=$work/owner2; mkdir -p "$d"
printf 'mjpeg\n' > "$d/type"; printf '60\n' > "$d/qlty"
printf '0\n' > "$d/res"; printf '120\n' > "$d/fps"
seed "$d"
check "MJPEG quality 60, source resolution, 120 fps kept" '[ "$count" = 0 ]'
d=$work/owner3; mkdir -p "$d"
printf 'h265' > "$d/type"; printf '2000' > "$d/qlty"
printf '480' > "$d/res"; printf '10' > "$d/fps"
seed "$d"
check "h265, 2000 kbit/s, 480, 10 fps kept" '[ "$count" = 0 ]'

echo "===== through the S95nanokvm links ====="
shared=$work/data/screen; app=$work/kvmapp/kvm
mkdir -p "$shared" "$app"
for f in type qlty res fps; do ln -s "$shared/$f" "$app/$f"; done
printf 'h264' > "$shared/type"; printf '5000' > "$shared/qlty"
printf '1080' > "$shared/res"
seed "$app"
check "only the missing fps written" '[ "$count" = 1 ]'
check "the shared qlty is still 5000" '[ "$(cat "$shared/qlty")" = 5000 ]'
check "the shared res is still 1080" '[ "$(cat "$shared/res")" = 1080 ]'
check "the missing shared fps was made through its link" '[ "$(cat "$shared/fps")" = 30 ]'
for f in type qlty res fps; do
    check "$f is still a link" '[ -L "$app/$f" ]'
done

echo "===== a missing, empty or unusable file is replaced ====="
d=$work/bad; mkdir -p "$d"
: > "$d/qlty"; printf 'abc' > "$d/fps"; printf '123' > "$d/res"; printf 'vp9' > "$d/type"
seed "$d"
check "four files written" '[ "$count" = 4 ]'
check "empty qlty becomes 3000" '[ "$(cat "$d/qlty")" = 3000 ]'
check "fps abc becomes 30" '[ "$(cat "$d/fps")" = 30 ]'
check "res 123 becomes 720" '[ "$(cat "$d/res")" = 720 ]'
check "type vp9 becomes h264" '[ "$(cat "$d/type")" = h264 ]'
d=$work/bad2; mkdir -p "$d"
printf '0' > "$d/qlty"; printf '0' > "$d/fps"; printf '99999999' > "$d/res"; printf '' > "$d/type"
seed "$d"
check "qlty 0, fps 0, res 99999999 and an empty type replaced" '[ "$count" = 4 ]'
d=$work/bad3; mkdir -p "$d"
head -c 100 /dev/zero | tr '\0' '7' > "$d/qlty"
seed "$d"
check "a file too long for a setting is replaced" '[ "$(cat "$d/qlty")" = 3000 ]'

echo "===== the update path uses it ====="
code=$(grep -v '^[[:space:]]*//' "$INIT")
if printf '%s\n' "$code" | grep -Eq 'echo [^>]*> */kvmapp/kvm/(fps|qlty|res|type)([^a-z_]|$)'; then
    note "new_app_init writes no screen setting with a redirect" FAIL
else
    note "new_app_init writes no screen setting with a redirect" OK
fi
if printf '%s\n' "$code" | grep -q 'seed_screen_defaults("/kvmapp/kvm")'; then
    note "new_app_init seeds /kvmapp/kvm" OK
else
    note "new_app_init seeds /kvmapp/kvm" FAIL
fi

echo "===== one default bitrate everywhere ====="
SCREEN_GO=$ROOT/server/common/screen.go
WEB=$ROOT/web/src/pages/desktop/menu/screen/constants.ts
PACKAGE=$ROOT/scripts/package.sh
check "kvm_system seeds 3000" \
    'grep -q "^#define SCREEN_DEFAULT_QLTY[[:space:]]*\"3000\"" "$MAIN/lib/system_init/screen_defaults.h"'
if [ -f "$SCREEN_GO" ]; then
    check "the server's DefaultBitRate is 3000" 'grep -q "^const DefaultBitRate = 3000$" "$SCREEN_GO"'
fi
if [ -f "$WEB" ]; then
    check "the web's DEFAULT_BIT_RATE is 3000" 'grep -q "^export const DEFAULT_BIT_RATE = 3000;$" "$WEB"'
fi
if [ -f "$PACKAGE" ]; then
    line=$(grep 'STAGE/kvm/qlty' "$PACKAGE")
    case $line in
        "printf '3000\\n'"*) note "the tarball seeds qlty 3000" OK ;;
        *) note "the tarball seeds qlty 3000" FAIL ;;
    esac
fi

echo
if [ "$fails" -gt 0 ]; then
    echo "FAILED: $fails"
    exit 1
fi
echo "all passed"
