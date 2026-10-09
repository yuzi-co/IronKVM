#!/bin/sh
# Check that kvm_system's update path leaves an IronKVM slot image's boot
# scripts alone, and a stock board's update as it was.
#
#   test-slot-scripts.sh [path/to/kvm_system/main]
#
# Not destructive: everything happens in a temporary directory.
#
# After every tarball update new_app_init copies boot scripts from
# /kvmapp/system/init.d into /etc/init.d and deletes S30wifi on a board without
# wlan0. On IronKVM's vendor-kernel image that put back S00kmod, which the image
# removes because its hook runner loads the same modules, and deleted S30wifi,
# which the image keeps (ironkvm-dist run sheet, trial 70). slot_scripts.cpp
# reads the image's /usr/share/ironkvm/init.d.refuse, the list install.sh
# already honours. It has no MaixCDK dependency, so it builds here with the host
# compiler; the rest of kvm_system does not, and the last section only reads it.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
MAIN=${1:-$ROOT/support/sg2002/kvm_system/main}
SRC=$MAIN/lib/system_init/slot_scripts.cpp
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
#include "slot_scripts.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(int argc, char **argv)
{
	if (argc == 4 && strcmp(argv[1], "refuses") == 0) {
		printf("%d\n", slot_refuses(argv[2], argv[3]));
		return 0;
	}
	if (argc == 6 && strcmp(argv[1], "install") == 0) {
		int r = install_boot_script(argv[2], argv[3], argv[4], argv[5]);
		fprintf(stderr, "%d\n", r);
		return 0;
	}
	if (argc == 5 && strcmp(argv[1], "wifi") == 0) {
		static const char *names[] = { "install", "keep", "remove" };
		printf("%s\n", names[s30wifi_action(argv[2], atoi(argv[3]), atoi(argv[4]))]);
		return 0;
	}
	return 2;
}
EOF

if ! "$CXX" -std=gnu++11 -Wall -Wextra -Werror -I"$MAIN/lib/system_init" \
        -o "$work/t" "$work/t.cpp" "$SRC" 2> "$work/cc.log"; then
    cat "$work/cc.log"
    echo "FAILED: slot_scripts.cpp does not build"
    exit 1
fi
T=$work/t

# The three slots. No list is a stock Sipeed board. The vendor image refuses
# S00kmod only; the mainline image refuses more, S30wifi among them. Both lists
# are shaped as ironkvm-dist writes them: a comment header, one name a line.
NONE=$work/none.refuse
VENDOR=$work/vendor.refuse
cat > "$VENDOR" <<'EOF'
# Boot scripts an application update must not install on this slot.
#
# S00kmod: the hook runner loads the modules.
S00kmod
EOF
MAINLINE=$work/mainline.refuse
cat > "$MAINLINE" <<'EOF'
# Boot scripts an application update must not install on a mainline slot.
S00hwinit
S00kmod
S04media
S15kvmhwd
S25wifimod
S30wifi
S97oled-nudge
EOF

echo "===== reading the list ====="
check "no list refuses nothing" '[ "$($T refuses "$NONE" S00kmod)" = 0 ]'
check "the vendor list refuses S00kmod" '[ "$($T refuses "$VENDOR" S00kmod)" = 1 ]'
check "the vendor list does not refuse S30wifi" '[ "$($T refuses "$VENDOR" S30wifi)" = 0 ]'
check "the mainline list refuses S30wifi" '[ "$($T refuses "$MAINLINE" S30wifi)" = 1 ]'
check "a name inside a comment is not refused" '[ "$($T refuses "$VENDOR" hook)" = 0 ]'
check "a prefix of a listed name is not refused" '[ "$($T refuses "$VENDOR" S00km)" = 0 ]'
check "a longer name is not refused" '[ "$($T refuses "$VENDOR" S00kmodx)" = 0 ]'
printf '  S00kmod \r\n#S30wifi\n' > "$work/odd.refuse"
check "blanks and CRLF around a name are ignored" '[ "$($T refuses "$work/odd.refuse" S00kmod)" = 1 ]'
check "a commented-out name is not refused" '[ "$($T refuses "$work/odd.refuse" S30wifi)" = 0 ]'

echo "===== S30wifi ====="
# wifi LIST VENDOR WLAN0
for v in 0 1; do for w in 0 1; do
    want=remove; [ "$v$w" = 11 ] && want=install
    check "stock, vendor=$v wlan0=$w: $want, as before" '[ "$($T wifi "$NONE" $v $w)" = $want ]'
done; done
check "vendor image, 5.10 with wlan0: install" '[ "$($T wifi "$VENDOR" 1 1)" = install ]'
check "vendor image, 5.10 without wlan0: keep (trial 70)" '[ "$($T wifi "$VENDOR" 1 0)" = keep ]'
for v in 0 1; do for w in 0 1; do
    check "mainline image, vendor=$v wlan0=$w: remove" '[ "$($T wifi "$MAINLINE" $v $w)" = remove ]'
done; done

echo "===== an update on each slot ====="
# The copies new_app_init makes on a 5.10 kernel, through install_boot_script,
# into an /etc/init.d shaped like the slot's.
pkg=$work/pkg; mkdir -p "$pkg"
for s in S00kmod S01fs S03usbdev S15kvmhwd S30eth S30wifi S50sshd S95nanokvm; do
    echo "# package $s" > "$pkg/$s"
done
update() { # LIST ETC
    for s in S00kmod S01fs S03usbdev S15kvmhwd S30eth S50sshd S95nanokvm; do
        "$T" install "$1" "$pkg" "$2" "$s" > "$2.log" 2>/dev/null || return 1
        cat "$2.log" >> "$2.out"
    done
}
etc=$work/stock; mkdir -p "$etc"; : > "$etc.out"
update "$NONE" "$etc"
check "stock: S00kmod installed" '[ -f "$etc/S00kmod" ]'
check "stock: every script installed" '[ "$(ls "$etc" | wc -l)" = 7 ]'
check "stock: nothing reported" '[ ! -s "$etc.out" ]'
etc=$work/vendor; mkdir -p "$etc"; : > "$etc.out"
echo "# image S30wifi" > "$etc/S30wifi"
update "$VENDOR" "$etc"
check "vendor image: S00kmod not installed" '[ ! -e "$etc/S00kmod" ]'
check "vendor image: the other six installed" \
    '[ -f "$etc/S01fs" ] && [ -f "$etc/S03usbdev" ] && [ -f "$etc/S15kvmhwd" ] && [ -f "$etc/S30eth" ] && [ -f "$etc/S50sshd" ] && [ -f "$etc/S95nanokvm" ]'
check "vendor image: the refusal is reported" \
    'grep -qx "new_app_init: S00kmod is refused by this slot ($VENDOR), not installed" "$etc.out"'
check "vendor image: S30wifi untouched by the copies" '[ "$(cat "$etc/S30wifi")" = "# image S30wifi" ]'
etc=$work/mainline; mkdir -p "$etc"; : > "$etc.out"
update "$MAINLINE" "$etc"
check "mainline image: S00kmod and S15kvmhwd not installed" '[ ! -e "$etc/S00kmod" ] && [ ! -e "$etc/S15kvmhwd" ]'
check "mainline image: S01fs, S03usbdev, S30eth, S50sshd, S95nanokvm installed" \
    '[ "$(ls "$etc" | wc -l)" = 5 ]'

echo "===== the update path uses it ====="
code=$(grep -v '^[[:space:]]*//' "$INIT")
body=$(printf '%s\n' "$code" | sed -n '/^void new_app_init(void)/,/^}/p')
check "new_app_init found" '[ -n "$body" ]'
# The one direct copy left is S30wifi's, which s30wifi_action decides.
n=$(printf '%s\n' "$body" | grep -c 'cp -f /kvmapp/system/init.d/')
check "every other boot script goes through install_boot_script" '[ "$n" = 1 ]'
check "that copy is S30wifi" \
    'printf "%s\n" "$body" | grep "cp -f /kvmapp/system/init.d/" | grep -q "S30wifi /etc/init.d/"'
for s in S00kmod S01fs S03usbdev S15kvmhwd S30eth S50sshd S95nanokvm S98tailscaled; do
    check "$s through install_boot_script" \
        'printf "%s\n" "$body" | grep -q "install_boot_script(refuse, pkg, etc, \"$s\")"'
done
check "S30wifi decided by s30wifi_action" \
    'printf "%s\n" "$body" | grep -q "s30wifi_action(refuse, vendor, kvm_wifi_exist())"'
check "the list is install.sh's" \
    'grep -q "^#define SLOT_REFUSE_LIST[[:space:]]*\"/usr/share/ironkvm/init.d.refuse\"" "$MAIN/lib/system_init/slot_scripts.h"'
INSTALL_SH=$ROOT/kvmapp/system/install.sh
if [ -f "$INSTALL_SH" ]; then
    check "install.sh reads the same list" \
        'grep -q "^REFUSE=\${INSTALL_REFUSE:-/usr/share/ironkvm/init.d.refuse}$" "$INSTALL_SH"'
fi

echo
if [ "$fails" -gt 0 ]; then
    echo "FAILED: $fails"
    exit 1
fi
echo "all passed"
