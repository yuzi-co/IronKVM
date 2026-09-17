#!/bin/sh
# Check that S95nanokvm shares the screen settings between the two slots.
#
#   test-screen-settings.sh [path-to-S95nanokvm]
#
# /kvmapp is part of each slot's root filesystem, so every slot kept its own
# stream type, frame rate, quality, resolution, codec and GOP, and booting the
# other slot quietly changed them. share_screen_settings turns each file into a
# link to /etc/kvm/screen, which S02identity puts on /data.
#
# This runs the shipped function against a scratch tree, the way
# test-runtime-state.sh does, so it fails if the function stops doing what its
# comment claims.
S95=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S95nanokvm}
[ -f "$S95" ] || { echo "usage: test-screen-settings.sh <S95nanokvm>"; exit 1; }

fails=0
note() { printf '  %-66s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

NAMES="type fps qlty res codec gop"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# Git Bash copies instead of linking. Every case below is about links, so a
# run there proves nothing either way.
ln -s "$work" "$work/linkprobe" 2>/dev/null
if [ ! -L "$work/linkprobe" ]; then
    echo "this filesystem has no symlinks; run it under Linux or busybox"
    exit 2
fi
rm -f "$work/linkprobe"

sed -n '/^SETTINGS_DIR=/p;/^SCREEN_SETTINGS=/p;/^share_screen_settings() {/,/^}/p' "$S95" \
    | sed "s|/kvmapp/kvm|$work/kvmapp/kvm|g" > "$work/funcs.sh"

if ! grep -q '^share_screen_settings() {' "$work/funcs.sh"; then
    note "the script still defines share_screen_settings" FAIL
    echo
    echo "$fails case(s) FAILED"
    exit 1
fi

shared="$work/data/identity/screen"

run() {
    SETTINGS_DIR=$shared sh -c ". '$work/funcs.sh'; share_screen_settings" > "$work/err" 2>&1
}

# A slot's own tree, holding the values that slot last had.
slot_tree() {
    rm -rf "$work/kvmapp"
    mkdir -p "$work/kvmapp/kvm"
    for name in $NAMES; do
        echo "$1-$name" > "$work/kvmapp/kvm/$name"
    done
    echo 1080 > "$work/kvmapp/kvm/height"
    echo 1920 > "$work/kvmapp/kvm/width"
}

echo "===== the first slot seeds the shared copy ====="
rm -rf "$work/data"
slot_tree a
if run; then note "share_screen_settings runs without error" OK
else note "share_screen_settings runs without error" FAIL; sed 's/^/    /' "$work/err"; fi

for name in $NAMES; do
    [ -L "$work/kvmapp/kvm/$name" ] \
        && [ "$(readlink "$work/kvmapp/kvm/$name")" = "$shared/$name" ] \
        && note "$name is a link to the shared directory" OK \
        || note "$name is a link to the shared directory" FAIL
    [ "$(cat "$shared/$name" 2>/dev/null)" = "a-$name" ] \
        && note "$name keeps the value the slot had" OK \
        || note "$name keeps the value the slot had" FAIL
done

for name in width height; do
    [ -f "$work/kvmapp/kvm/$name" ] && [ ! -L "$work/kvmapp/kvm/$name" ] \
        && note "$name, which describes the source, is left alone" OK \
        || note "$name, which describes the source, is left alone" FAIL
done

echo
echo "===== the other slot reads the same settings ====="
# Its own files hold what it had before the change. The shared copy wins, or a
# slot switch goes on changing the settings.
slot_tree b
run
for name in $NAMES; do
    [ "$(cat "$work/kvmapp/kvm/$name" 2>/dev/null)" = "a-$name" ] \
        && note "$name reads the shared value, not this slot's old one" OK \
        || note "$name reads the shared value, not this slot's old one" FAIL
done

echo
echo "===== a write through the link lands in the shared copy ====="
# kvm_system writes with a shell redirect, and the server with os.WriteFile.
# Both follow the link. A writer that replaced the link would put the file
# back on the slot.
echo 60 > "$work/kvmapp/kvm/fps"
[ "$(cat "$shared/fps")" = 60 ] && [ -L "$work/kvmapp/kvm/fps" ] \
    && note "a redirect updates the shared file and keeps the link" OK \
    || note "a redirect updates the shared file and keeps the link" FAIL

echo
echo "===== a setting nobody has chosen yet ====="
# codec and gop are absent on a board that never changed them. The link is made
# anyway, so the first write goes to the shared directory.
rm -rf "$work/data"
slot_tree c
rm -f "$work/kvmapp/kvm/codec"
run
[ -L "$work/kvmapp/kvm/codec" ] && [ ! -e "$shared/codec" ] \
    && note "an absent setting gets a link and no invented value" OK \
    || note "an absent setting gets a link and no invented value" FAIL
echo 2 > "$work/kvmapp/kvm/codec"
[ "$(cat "$shared/codec" 2>/dev/null)" = 2 ] \
    && note "the first write through it creates the shared file" OK \
    || note "the first write through it creates the shared file" FAIL

echo
echo "===== a restart leaves the links alone ====="
# start_services runs on every restart. Recreating a link is a directory write
# on the boot medium for no reason.
before=$(ls -li "$work/kvmapp/kvm" | awk '{print $1, $NF}' | sort)
run
after=$(ls -li "$work/kvmapp/kvm" | awk '{print $1, $NF}' | sort)
[ "$before" = "$after" ] \
    && note "a second run does not recreate any link" OK \
    || note "a second run does not recreate any link" FAIL

echo
echo "===== no shared directory, nothing lost ====="
# A directory that cannot be made must not cost the slot its own settings.
rm -rf "$work/data"
slot_tree d
rm -f "$work/kvmapp/kvm/codec"
mkdir -p "$work/data/identity"
: > "$shared"
run
st=$?
[ "$st" = 0 ] \
    && note "exits 0 when the shared directory cannot be made" OK \
    || note "exits 0 when the shared directory cannot be made (got $st)" FAIL
[ -f "$work/kvmapp/kvm/fps" ] && [ ! -L "$work/kvmapp/kvm/fps" ] \
    && [ "$(cat "$work/kvmapp/kvm/fps")" = d-fps ] \
    && note "the slot keeps its own files" OK \
    || note "the slot keeps its own files" FAIL
# A link into a directory that does not exist is a setting no writer can save.
[ ! -L "$work/kvmapp/kvm/codec" ] \
    && note "an absent setting gets no link that leads nowhere" OK \
    || note "an absent setting gets no link that leads nowhere" FAIL

echo
echo "===== a copy that fails keeps the slot's file ====="
# The directory exists, but one name cannot be written. Deleting the slot's
# file before the value is safe elsewhere would lose it.
# A full card is the likely cause; a cp that refuses fps stands in for it.
rm -rf "$work/data"
slot_tree e
mkdir -p "$work/bin"
cat > "$work/bin/cp" <<'STUB'
#!/bin/sh
case "$2" in */fps) exit 1 ;; esac
exec /bin/cp "$@"
STUB
chmod 755 "$work/bin/cp"
PATH="$work/bin:$PATH" run
rm -rf "$work/bin"
[ -f "$work/kvmapp/kvm/fps" ] && [ ! -L "$work/kvmapp/kvm/fps" ] \
    && [ "$(cat "$work/kvmapp/kvm/fps")" = e-fps ] \
    && note "the file whose copy failed is still there" OK \
    || note "the file whose copy failed is still there" FAIL
[ "$(cat "$shared/qlty" 2>/dev/null)" = e-qlty ] \
    && note "the other settings are still shared" OK \
    || note "the other settings are still shared" FAIL

echo
echo "===== start runs it ====="
# The function is only useful if the start path calls it, before either process
# reads a setting.
sed -n '/^start_services() {/,/^}/p' "$S95" | grep -q '^    share_screen_settings$' \
    && note "start_services calls share_screen_settings" OK \
    || note "start_services calls share_screen_settings" FAIL

echo
echo "===== the script still parses ====="
sh -n "$S95" 2>/dev/null \
    && note "sh -n accepts the script" OK \
    || note "sh -n rejects the script" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
