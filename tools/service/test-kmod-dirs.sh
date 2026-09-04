#!/bin/sh
# Check that S00kmod prefers the modules the install package ships.
#
#   test-kmod-dirs.sh [S00kmod]
#
# Not destructive: the script under test is run with a stub insmod on PATH and
# module directories in a temporary tree. No module is ever inserted.
#
# The fault this covers: S00kmod loaded every module from /mnt/system/ko, so a
# rebuilt module in /kvmapp/system/ko was shipped, installed, and never run.
# soph_vi.ko was patched on 2026-09-03 to keep four idle ISP threads out of the
# load average, and the board loaded the stock module at every boot for the
# five days afterwards. The load average stayed at about 4 and looked like the
# patch had not worked.
#
# This is the first init script the board runs and it loads 22 modules, so the
# order matters as much as the directory. A board that loads soph_vi before
# soph_base has no video and no way in.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
SCRIPT=${1:-$ROOT/kvmapp/system/init.d/S00kmod}

[ -f "$SCRIPT" ] || { echo "missing: $SCRIPT"; exit 2; }

fails=0
note() {
    printf '  %-58s %s\n' "$1" "$2"
    case $2 in FAIL*) fails=$((fails + 1)) ;; esac
    return 0
}

work=$(mktemp -d) || exit 2
trap 'rm -rf "$work"' EXIT INT TERM

# The order the stock script loaded them in. Written out rather than derived
# from the script under test, so that a reordering is a failure here and not a
# silently updated expectation.
ORDER='soph_sys.ko soph_base.ko soph_rtos_cmdqu.ko soph_fast_image.ko
soph_mipi_rx.ko soph_snsr_i2c.ko soph_vi.ko soph_vpss.ko soph_dwa.ko
soph_vo.ko soph_rgn.ko soph_wdt.ko soph_clock_cooling.ko soph_tpu.ko
soph_vcodec.ko soph_jpeg.ko soph_vc_driver.ko soph_rtc.ko soph_ive.ko
soph_mon.ko soph_pwm.ko soph_wiegand.ko'

# A stub insmod that records the path it was given and whatever else, and
# refuses any module named in $work/refuse.
mkdir -p "$work/bin"
cat > "$work/bin/insmod" <<'EOF'
#!/bin/sh
if [ -f "$INSMOD_REFUSE" ] && grep -qxF "$1" "$INSMOD_REFUSE"; then
    echo "refused $1" >> "$INSMOD_LOG.refused"
    exit 1
fi
echo "$*" >> "$INSMOD_LOG"
exit 0
EOF
chmod +x "$work/bin/insmod"

# /etc/profile is sourced by the script and does not exist off the device.
mkdir -p "$work/etc"
: > "$work/etc/profile"

# Two module directories, standing in for the ones on the board.
PKG="$work/kvmapp/system/ko"
STOCK="$work/mnt/system/ko"
mkdir -p "$PKG" "$STOCK"
for m in $ORDER; do
    echo "stock $m" > "$STOCK/$m"
done

run() {
    # run [refuse-list-file] -- runs the script with the stubs in place
    rm -f "$work/log" "$work/log.refused"
    INSMOD_LOG="$work/log" INSMOD_REFUSE="$1" \
    KMOD_DIRS="$PKG $STOCK" \
    PATH="$work/bin:$PATH" \
        sh "$work/under-test" start > "$work/out" 2>&1
    echo $? > "$work/rc"
}

# The script sources /etc/profile by absolute path, which cannot be redirected
# from here, so the copy under test sources the temporary one instead. That is
# the only edit, and the case below checks the original still has the line.
sed "s#^\t\. /etc/profile#\t. $work/etc/profile#" "$SCRIPT" > "$work/under-test"
grep -q '^	\. /etc/profile' "$SCRIPT" \
    && note "the shipped script still sources /etc/profile" OK \
    || note "the shipped script still sources /etc/profile" FAIL

echo "===== the package copy wins ====="

# Only soph_vi differs on the real board, so that is the one shipped here.
echo "package soph_vi.ko" > "$PKG/soph_vi.ko"
run /dev/null

grep -q "^$PKG/soph_vi.ko" "$work/log" \
    && note "soph_vi comes from the install package" OK \
    || note "soph_vi comes from the install package" FAIL

grep -q "^$STOCK/soph_vi.ko" "$work/log" \
    && note "the stock soph_vi is not loaded as well" FAIL \
    || note "the stock soph_vi is not loaded as well" OK

grep -q "^$STOCK/soph_base.ko" "$work/log" \
    && note "a module absent from the package comes from stock" OK \
    || note "a module absent from the package comes from stock" FAIL

echo "===== order and parameters are unchanged ====="

got=$(awk '{print $1}' "$work/log" | sed 's#.*/##' | tr '\n' ' ' | sed 's/ $//')
want=$(printf '%s' "$ORDER" | tr '\n' ' ' | tr -s ' ' | sed 's/ $//')
[ "$got" = "$want" ] \
    && note "all 22 modules load, in the original order" OK \
    || note "the load order changed" FAIL

count=$(wc -l < "$work/log" | tr -d ' ')
[ "$count" = 22 ] \
    && note "exactly 22 modules are loaded" OK \
    || note "$count modules loaded, wanted 22" FAIL

grep -q "soph_vc_driver.ko MaxVencChnNum=9 MaxVdecChnNum=9" "$work/log" \
    && note "soph_vc_driver keeps its channel counts" OK \
    || note "soph_vc_driver keeps its channel counts" FAIL

# A module with no parameters must not be handed an empty argument, which
# insmod would read as a second file name.
grep -q "^$STOCK/soph_sys.ko$" "$work/log" \
    && note "a module without parameters is given none" OK \
    || note "a module without parameters is given none" FAIL

echo "===== failure does not take the board down ====="

# A rebuilt module the kernel rejects must fall through to the stock one
# rather than leaving the pipeline without it.
echo "$PKG/soph_vi.ko" > "$work/refuse"
run "$work/refuse"
grep -q "^$STOCK/soph_vi.ko" "$work/log" \
    && note "a package module that is refused falls back to stock" OK \
    || note "a package module that is refused falls back to stock" FAIL
[ "$(wc -l < "$work/log" | tr -d ' ')" = 22 ] \
    && note "the fallback still leaves 22 modules loaded" OK \
    || note "the fallback still leaves 22 modules loaded" FAIL

# A module missing from both directories is reported and does not stop the run.
rm -f "$PKG/soph_ive.ko" "$STOCK/soph_ive.ko"
run /dev/null
grep -q "soph_ive.ko not loaded" "$work/out" \
    && note "a module missing everywhere is reported" OK \
    || note "a module missing everywhere is reported" FAIL
grep -q "^$STOCK/soph_mon.ko" "$work/log" \
    && note "the modules after a missing one still load" OK \
    || note "the modules after a missing one still load" FAIL
[ "$(cat "$work/rc")" = 0 ] \
    && note "a missing module still exits 0" OK \
    || note "a missing module exits $(cat "$work/rc"), wanted 0" FAIL
echo "stock soph_ive.ko" > "$STOCK/soph_ive.ko"

echo "===== the contract with the rest of the boot ====="

run /dev/null
# printf without a newline, then echo, so the whole thing is one line. Match
# it as one: an anchored "OK" never matched and the case passed for nothing
# until that was noticed.
grep -q "^load kernel module: OK$" "$work/out" \
    && note "the script still prints the boot line and OK" OK \
    || note "the script still prints the boot line and OK" FAIL
[ "$(cat "$work/rc")" = 0 ] \
    && note "the script still exits 0 on start" OK \
    || note "the script still exits 0 on start" FAIL

# Anything other than start does nothing at all, which is what rcS relies on
# when it runs the K-links at shutdown.
rm -f "$work/log"
INSMOD_LOG="$work/log" INSMOD_REFUSE=/dev/null KMOD_DIRS="$PKG $STOCK" \
    PATH="$work/bin:$PATH" sh "$work/under-test" stop > "$work/out2" 2>&1
[ ! -f "$work/log" ] \
    && note "stop loads nothing" OK \
    || note "stop loads nothing" FAIL

# The directory order is the whole fix, so it has to be readable as such.
grep -q 'KMOD_DIRS=.*kvmapp/system/ko.*mnt/system/ko' "$SCRIPT" \
    && note "the package directory is searched before stock" OK \
    || note "the package directory is searched before stock" FAIL

# The old form loaded everything relative to one hardcoded directory.
grep -q '^	cd /mnt/system/ko' "$SCRIPT" \
    && note "the hardcoded cd into the stock directory is gone" FAIL \
    || note "the hardcoded cd into the stock directory is gone" OK

echo
if [ "$fails" = 0 ]; then
    echo "test-kmod-dirs.sh: all cases passed"
    exit 0
fi
echo "test-kmod-dirs.sh: $fails case(s) failed"
exit 1
