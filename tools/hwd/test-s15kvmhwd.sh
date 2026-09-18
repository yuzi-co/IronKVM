#!/bin/sh
# What S15kvmhwd does, on a board that has the IronKVM hardware tools and on a
# board that does not.
#
#   test-s15kvmhwd.sh [path-to-S15kvmhwd]
#
# S15kvmhwd is a wrapper now. An image built by ironkvm-dist installs
# ironkvm-board and ironkvm-hwinit in /usr/bin, and the wrapper hands the work
# to them. A board on Sipeed's firmware has neither, and there the wrapper runs
# Sipeed's own code unchanged. Both paths are tested here, because the second
# one is what every deployed board runs today.
#
# == Why this suite rewrites the script ==
#
# S15kvmhwd writes /etc/kvm and /sys/class/gpio by their full names and no
# variable stands in for either. Two suites in ironkvm-dist parse the pin
# numbers, the register writes and the variant names straight out of this file,
# so the hooks that replace it cannot drift from it, and a variable would leave
# those suites nothing to read.
#
# So this suite copies the script and rewrites the two paths in the copy, then
# checks that the rewrite did what it says: the copy differs from the original,
# it still parses, and no occurrence of either path survived. The replacement
# directories are named so that neither of the two paths is a substring of what
# replaces it, or the last check could never fail. A rewrite that
# quietly missed one would point the test at the real /etc/kvm, and on the
# reference board that is a file the board boots from.
#
# i2cdetect, devmem, insmod, rmmod and reboot are stubs on PATH, so no case
# writes a register, touches a GPIO line or reboots anything. The suite refuses
# to start when PATH does not resolve reboot, i2cdetect and devmem to those
# stubs. It is meant to run on the reference board too, where a real devmem
# writes the SoC and a real reboot is a trip to the hardware: there is no
# remote power cycle here.
#
# ironkvm-board and ironkvm-hwinit are stubs as well, and they live in a second
# directory. A case adds that directory to PATH to be the board with the tools,
# and leaves it out to be the board without them.
S15=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S15kvmhwd}
[ -f "$S15" ] || { echo "no S15kvmhwd at $S15"; exit 2; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

ETC=$WORK/kvmconf
SYSFS=$WORK/gpio
BIN=$WORK/bin
TOOLS=$WORK/tools
SCRIPT=$WORK/S15kvmhwd
mkdir -p "$BIN" "$TOOLS" "$WORK/i2c"

# --- the copy under test ------------------------------------------------------

sed -e "s|/etc/kvm|$ETC|g" -e "s|/sys/class/gpio|$SYSFS|g" "$S15" > "$SCRIPT"

cmp -s "$S15" "$SCRIPT" && {
    echo "the path rewrite changed nothing, so this suite would drive the real /etc/kvm"; exit 2; }
sh -n "$SCRIPT" 2>/dev/null || {
    echo "the rewritten copy of S15kvmhwd does not parse, so no case can run"; exit 2; }
grep -q '/etc/kvm' "$SCRIPT" && {
    echo "/etc/kvm survived the rewrite, so this suite would drive the real one"; exit 2; }
grep -q '/sys/class/gpio' "$SCRIPT" && {
    echo "/sys/class/gpio survived the rewrite, so this suite would drive the real one"; exit 2; }

# --- the stubs ----------------------------------------------------------------

cat > "$BIN/i2cdetect" <<EOF
#!/bin/sh
# i2cdetect -ry <bus> <address> <address>
echo "i2cdetect \$*" >> "$WORK/i2c.log"
_s=\${3#0x}
if [ -f "$WORK/i2c/\$2.\$_s" ]; then
    echo "30: \$_s"
else
    echo "30: --"
fi
exit 0
EOF
cat > "$BIN/devmem" <<EOF
#!/bin/sh
echo "\$*" >> "$WORK/devmem.log"
exit 0
EOF
for n in insmod rmmod; do
    cat > "$BIN/$n" <<EOF
#!/bin/sh
echo "$n \$*" >> "$WORK/ko.log"
exit 0
EOF
done
cat > "$BIN/reboot" <<EOF
#!/bin/sh
echo "reboot \$*" >> "$WORK/reboot.log"
exit 0
EOF
chmod 755 "$BIN/i2cdetect" "$BIN/devmem" "$BIN/insmod" "$BIN/rmmod" "$BIN/reboot"

# ironkvm-board answers `get hdmi_bridge` from a file this suite writes. An
# empty file is the board that has no such key, which is every alpha board and
# every board that has not been detected yet, and the real tool exits 1 there.
cat > "$TOOLS/ironkvm-board" <<EOF
#!/bin/sh
echo "ironkvm-board \$*" >> "$WORK/tool.log"
if [ "\$1" = get ] && [ "\$2" = hdmi_bridge ]; then
    [ -s "$WORK/bridge" ] || exit 1
    cat "$WORK/bridge"
fi
exit 0
EOF
cat > "$TOOLS/ironkvm-hwinit" <<EOF
#!/bin/sh
echo "ironkvm-hwinit \$*" >> "$WORK/tool.log"
exit 0
EOF
chmod 755 "$TOOLS/ironkvm-board" "$TOOLS/ironkvm-hwinit"

BARE_PATH=$BIN:$PATH
TOOLED_PATH=$TOOLS:$BIN:$PATH

PATH=$BARE_PATH
export PATH
for n in reboot i2cdetect devmem; do
    [ "$(command -v "$n")" = "$BIN/$n" ] || {
        echo "PATH does not resolve $n to the stub, refusing to run S15kvmhwd"; exit 2; }
done

# --- driving the script -------------------------------------------------------

# present <bus>.<address>...: exactly these addresses answer, and no others
present() {
    rm -rf "$WORK/i2c"; mkdir -p "$WORK/i2c"
    for _a in "$@"; do : > "$WORK/i2c/$_a"; done
}

bridge() { printf '%s' "$1" > "$WORK/bridge"; }

reset() {
    rm -rf "$ETC" "$SYSFS"
    mkdir -p "$ETC" "$SYSFS"
    : > "$WORK/i2c.log"; : > "$WORK/devmem.log"; : > "$WORK/ko.log"
    : > "$WORK/tool.log"; : > "$WORK/reboot.log"; : > "$WORK/bridge"
}

# with <script> <subcommand>: the board that has the tools
with() { PATH=$TOOLED_PATH sh "$1" "$2" > "$WORK/out" 2>&1; echo $? > "$WORK/st"; }
# without <script> <subcommand>: the board on Sipeed's firmware
without() { PATH=$BARE_PATH sh "$1" "$2" > "$WORK/out" 2>&1; echo $? > "$WORK/st"; }

st() { cat "$WORK/st"; }
out() { cat "$WORK/out"; }
tools_called() { tr '\n' ';' < "$WORK/tool.log"; }
hw() { cat "$ETC/hw" 2>/dev/null; }
hdmi() { cat "$ETC/hdmi_version" 2>/dev/null; }
probes() { wc -l < "$WORK/i2c.log" 2>/dev/null | tr -d ' '; }

echo "===== the board that has ironkvm-board and ironkvm-hwinit ====="

reset
with "$SCRIPT" start
if [ "$(tools_called)" = "ironkvm-board detect;ironkvm-hwinit pinmux;" ]; then
    note "start detects the board and then muxes the pins" OK
else
    note "start detects the board and then muxes the pins: $(tools_called)" FAIL
fi

if [ "$(probes)" = 0 ] && [ ! -s "$WORK/devmem.log" ] && [ ! -s "$WORK/reboot.log" ]; then
    note "start runs no probe of its own and never reboots" OK
else
    note "start runs no probe of its own and never reboots ($(probes) probes)" FAIL
fi

reset
with "$SCRIPT" re-detect
if [ "$(tools_called)" = "ironkvm-board detect --again;ironkvm-hwinit pinmux;" ]; then
    note "re-detect asks for a fresh probe and then muxes the pins" OK
else
    note "re-detect asks for a fresh probe and then muxes the pins: $(tools_called)" FAIL
fi

reset
with "$SCRIPT" re-init
if [ "$(tools_called)" = "ironkvm-hwinit pinmux;" ]; then
    note "re-init muxes the pins and detects nothing" OK
else
    note "re-init muxes the pins and detects nothing: $(tools_called)" FAIL
fi

echo
echo "===== get_hdmi_version, which kvm_system reads as a file ====="

reset
bridge d
with "$SCRIPT" get_hdmi_version
if [ "$(hdmi)" = d ]; then
    note "the described bridge is written to the file kvm_system reads" OK
else
    note "the described bridge is written to the file kvm_system reads (got '$(hdmi)')" FAIL
fi

if [ "$(out)" = d ]; then
    note "get_hdmi_version prints the value as well" OK
else
    note "get_hdmi_version prints the value as well (got '$(out)')" FAIL
fi

if [ "$(probes)" = 0 ]; then
    note "a described bridge is not probed for a second time" OK
else
    note "a described bridge is not probed for a second time ($(probes) probes)" FAIL
fi

# /etc/kvm is on the boot card, so a value that has not changed is not written
# again. The file is given an old timestamp and the reference an older one, so
# the comparison cannot turn on the clock's resolution.
touch -t 200101010000 "$ETC/hdmi_version"
touch -t 200201010000 "$WORK/stamp"
with "$SCRIPT" get_hdmi_version
if [ ! "$ETC/hdmi_version" -nt "$WORK/stamp" ] && [ "$(hdmi)" = d ]; then
    note "a value that has not changed is not written to the card again" OK
else
    note "a value that has not changed is not written to the card again" FAIL
fi

bridge ux
with "$SCRIPT" get_hdmi_version
if [ "$(hdmi)" = ux ] && [ "$ETC/hdmi_version" -nt "$WORK/stamp" ]; then
    note "a value that has changed is written" OK
else
    note "a value that has changed is written (got '$(hdmi)')" FAIL
fi

# An alpha board has no hdmi_bridge key, because the detection probes the bridge
# on the beta and pcie path alone. kvm_system has always had a probed value
# there, so the probe is what keeps that board unchanged.
reset
present 4.2c
with "$SCRIPT" get_hdmi_version
if [ "$(hdmi)" = c ] && [ "$(probes)" -gt 0 ]; then
    note "a board whose description names no bridge is probed, as it is today" OK
else
    note "a board whose description names no bridge is probed, as it is today (got '$(hdmi)')" FAIL
fi

echo
echo "===== the surface a vendor binary sees ====="

reset
with "$SCRIPT" whatever
if [ "$(st)" = 0 ] && [ ! -s "$WORK/tool.log" ] && [ -z "$(hw)" ] && [ -z "$(hdmi)" ]; then
    note "an unknown argument exits 0 and does nothing" OK
else
    note "an unknown argument exits 0 and does nothing (got $(st))" FAIL
fi

reset
PATH=$TOOLED_PATH sh "$SCRIPT" > "$WORK/out" 2>&1; echo $? > "$WORK/st"
if [ "$(st)" = 0 ] && [ ! -s "$WORK/tool.log" ]; then
    note "no argument at all exits 0 and does nothing" OK
else
    note "no argument at all exits 0 and does nothing (got $(st))" FAIL
fi

# kvm_system copies this file from /kvmapp/system/init.d into /etc/init.d and
# calls it from both, so nothing in it may depend on where it sits.
mkdir -p "$WORK/etc.init.d" "$WORK/kvmapp.init.d"
cp "$SCRIPT" "$WORK/etc.init.d/S15kvmhwd"
cp "$SCRIPT" "$WORK/kvmapp.init.d/S15kvmhwd"
reset
with "$WORK/etc.init.d/S15kvmhwd" start
a=$(tools_called)
reset
with "$WORK/kvmapp.init.d/S15kvmhwd" start
b=$(tools_called)
if [ -n "$a" ] && [ "$a" = "$b" ]; then
    note "the copy in /etc/init.d and the copy in /kvmapp behave the same" OK
else
    note "the copy in /etc/init.d and the copy in /kvmapp behave the same ('$a' vs '$b')" FAIL
fi

echo
echo "===== the board on Sipeed's firmware, which has neither tool ====="

reset
present
without "$SCRIPT" start
if [ "$(hw)" = beta ] && [ "$(probes)" -gt 0 ]; then
    note "a board without ironkvm-board still detects" OK
else
    note "a board without ironkvm-board still detects (hw is '$(hw)', $(probes) probes)" FAIL
fi

reset
present 1.3d
without "$SCRIPT" start
if [ "$(hw)" = alpha ]; then
    note "an OLED on the I2C1 controller is still an alpha board" OK
else
    note "an OLED on the I2C1 controller is still an alpha board (hw is '$(hw)')" FAIL
fi

reset
present 5.3c
without "$SCRIPT" start
if [ "$(hw)" = pcie ] && [ -s "$WORK/reboot.log" ]; then
    note "a pcie board is still found and still reboots once" OK
else
    note "a pcie board is still found and still reboots once (hw is '$(hw)')" FAIL
fi

reset
present 4.2b
without "$SCRIPT" get_hdmi_version
if [ "$(hdmi)" = d ] && [ "$(out)" = d ]; then
    note "get_hdmi_version still probes the bus when there is no description" OK
else
    note "get_hdmi_version still probes the bus when there is no description (got '$(hdmi)')" FAIL
fi

reset
echo beta > "$ETC/hw"
without "$SCRIPT" re-init
if grep -q '0x030010E4' "$WORK/devmem.log" && [ "$(probes)" = 0 ]; then
    note "re-init still muxes the pins of the variant in /etc/kvm/hw" OK
else
    note "re-init still muxes the pins of the variant in /etc/kvm/hw" FAIL
fi

echo
echo "===== the script still parses ====="
sh -n "$S15" 2>/dev/null && note "sh -n accepts S15kvmhwd" OK || note "sh -n accepts S15kvmhwd" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
