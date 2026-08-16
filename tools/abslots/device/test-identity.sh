#!/bin/sh
# Check that S02identity carries the whole of a board's identity across a slot
# switch, not just /etc/kvm.
#
#   test-identity.sh [path-to-S02identity]
#
# The first A/B flash lost three things, and all three were identity that does
# not live in /etc/kvm:
#
#   /etc/shadow                  the root password reverted to the factory one,
#                                which with /boot/usb.acm enabled hands a root
#                                prompt to the managed host
#   /root/.ssh/authorized_keys   removed with /root as "operator cruft", so key
#                                login was impossible
#   /data itself                 unmounted, because S01fs hardcoded p3
#
# Two directories on /data, deliberately siblings rather than nested. The bind
# source has to contain exactly what /etc/kvm should contain, so anything else
# would show up inside /etc/kvm.
#
#   /data/identity          bound over /etc/kvm
#   /data/identity-system   shadow, passwd, authorized_keys, copied out
#
# shadow and authorized_keys are copied rather than bound. passwd rewrites its
# file by rename, which breaks a single-file bind mount, and a broken bind on
# /etc/shadow is a board nobody can log into.
S02=${1:-$(dirname "$0")/S02identity}
[ -f "$S02" ] || { echo "usage: test-identity.sh <S02identity>"; exit 1; }

WORK=$(mktemp -d)

# On Linux the bind mount in S02identity actually succeeds, and every case
# binds again over the same path. Without this the stack of binds outlives the
# test, its own `rm -rf` fails with EBUSY, and the build host collects a mount
# per run. On Windows the bind always fails, so this was invisible there and
# every local run looked clean.
unbind() {
    i=0
    while grep -q " $WORK/root/etc/kvm " /proc/mounts 2>/dev/null && [ "$i" -lt 32 ]; do
        umount "$WORK/root/etc/kvm" 2>/dev/null || break
        i=$((i + 1))
    done
}
trap 'unbind; rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

setup() {
    unbind
    rm -rf "$WORK/root" "$WORK/data"
    mkdir -p "$WORK/root/etc/kvm" "$WORK/root/root/.ssh" "$WORK/root/etc/ssh" "$WORK/data"
    printf 'root:$FACTORY:1::::::\n'  > "$WORK/root/etc/shadow"
    printf 'root:x:0:0:root:/root:/bin/sh\n' > "$WORK/root/etc/passwd"
    printf 'image-config\n'           > "$WORK/root/etc/kvm/server.yaml"
    printf 'IMAGE-HOST-KEY\n'         > "$WORK/root/etc/ssh/ssh_host_ed25519_key"
    printf 'ssh-ed25519 IMAGEPUB\n'   > "$WORK/root/etc/ssh/ssh_host_ed25519_key.pub"
}

# run <mounted?>  -- MOUNTS is the stub /proc/mounts content
run() {
    (
        IDENTITY="$WORK/data/identity"
        SYSDIR="$WORK/data/identity-system"
        TARGET="$WORK/root/etc/kvm"
        ETCDIR="$WORK/root/etc"
        ROOTHOME="$WORK/root/root"
        SSHDIR="$WORK/root/etc/ssh"
        DATA_MOUNTED="$1"
        export IDENTITY SYSDIR TARGET ETCDIR ROOTHOME SSHDIR DATA_MOUNTED
        sh "$S02" "${2:-start}"
    ) 2>&1
}

echo "===== a first boot seeds identity from the image ====="
setup
run yes > "$WORK/log1"

[ -f "$WORK/data/identity/server.yaml" ] \
    && note "the image's /etc/kvm seeds /data/identity" OK \
    || note "the image's /etc/kvm seeds /data/identity" FAIL

[ -f "$WORK/data/identity-system/shadow" ] \
    && note "the image's shadow seeds identity-system" OK \
    || note "the image's shadow seeds identity-system" FAIL

echo
echo "===== a later boot restores the device's identity, not the image's ====="
setup
mkdir -p "$WORK/data/identity" "$WORK/data/identity-system"
printf 'device-config\n'            > "$WORK/data/identity/server.yaml"
printf 'root:$REALHASH:1::::::\n'   > "$WORK/data/identity-system/shadow"
printf 'ssh-ed25519 AAAA vadim\n'   > "$WORK/data/identity-system/authorized_keys"
run yes > "$WORK/log2"

grep -q REALHASH "$WORK/root/etc/shadow" \
    && note "the stored root password is restored" OK \
    || note "the stored root password is restored, got: $(cat "$WORK/root/etc/shadow")" FAIL

grep -q 'vadim' "$WORK/root/root/.ssh/authorized_keys" 2>/dev/null \
    && note "authorized_keys is restored" OK \
    || note "authorized_keys is restored" FAIL

# Only meaningful where the filesystem honours chmod. Under Git Bash on Windows
# chmod 600 is a no-op and this reported a failure that did not exist, which is
# the same trap that made an earlier dispatch case test nothing at all. Say
# SKIP rather than pass quietly or fail wrongly.
: > "$WORK/permprobe"
chmod 600 "$WORK/permprobe"
if [ "$(ls -l "$WORK/permprobe" | cut -c1-10)" = "-rw-------" ]; then
    perm=$(ls -l "$WORK/root/root/.ssh/authorized_keys" 2>/dev/null | cut -c1-10)
    [ "$perm" = "-rw-------" ] \
        && note "authorized_keys is 0600" OK \
        || note "authorized_keys is $perm, want -rw-------" FAIL
else
    note "authorized_keys is 0600 (this filesystem ignores chmod)" SKIP
fi

echo
echo "===== the ssh host key is identity too ====="
#
# A slot switch swaps the whole root filesystem, so it swaps /etc/ssh with it.
# Booting root A for the first time on 2026-08-16 changed the board's host key,
# and every client that had ever connected refused the new one. With
# BatchMode=yes ssh does not even print why: it just fails.
#
# The host key IS the board's identity. It belongs with the password and the
# authorized_keys, or every slot switch looks like an attack to every client.

setup
run yes > "$WORK/logk1"
[ -f "$WORK/data/identity-system/ssh/ssh_host_ed25519_key" ] \
    && note "a first boot seeds the host key from the image" OK \
    || note "a first boot seeds the host key from the image" FAIL

setup
mkdir -p "$WORK/data/identity" "$WORK/data/identity-system/ssh"
printf 'DEVICE-HOST-KEY\n'       > "$WORK/data/identity-system/ssh/ssh_host_ed25519_key"
printf 'ssh-ed25519 DEVICEPUB\n' > "$WORK/data/identity-system/ssh/ssh_host_ed25519_key.pub"
run yes > "$WORK/logk2"

grep -q DEVICE-HOST-KEY "$WORK/root/etc/ssh/ssh_host_ed25519_key" \
    && note "a later boot restores the board's own host key" OK \
    || note "a later boot kept the image's host key, so clients see a new board" FAIL

grep -q DEVICEPUB "$WORK/root/etc/ssh/ssh_host_ed25519_key.pub" \
    && note "the public half is restored with it" OK \
    || note "the public half is restored with it" FAIL

grep -q "host keys restored" "$WORK/logk2" \
    && note "and the restore is announced on the console" OK \
    || note "the restore happened but said nothing, so a boot log cannot show it" FAIL

if [ "$(ls -l "$WORK/permprobe" 2>/dev/null | cut -c1-10)" = "-rw-------" ]; then
    perm=$(ls -l "$WORK/root/etc/ssh/ssh_host_ed25519_key" 2>/dev/null | cut -c1-10)
    [ "$perm" = "-rw-------" ] \
        && note "a restored private host key is 0600" OK \
        || note "a restored private host key is $perm, want -rw-------" FAIL
else
    note "a restored private host key is 0600 (this filesystem ignores chmod)" SKIP
fi

# sshd refuses to start on a zero-length host key, and a board with no sshd is
# the situation this whole design exists to avoid.
setup
mkdir -p "$WORK/data/identity" "$WORK/data/identity-system/ssh"
: > "$WORK/data/identity-system/ssh/ssh_host_ed25519_key"
run yes > "$WORK/logk3"
grep -q IMAGE-HOST-KEY "$WORK/root/etc/ssh/ssh_host_ed25519_key" \
    && note "an empty stored host key is ignored" OK \
    || note "an empty stored host key replaced a working one" FAIL

# And save must capture them, or the first `slot identity save` after a
# regenerate silently stores everything except the key that matters.
setup
mkdir -p "$WORK/data/identity" "$WORK/data/identity-system"
printf 'REGENERATED\n' > "$WORK/root/etc/ssh/ssh_host_ed25519_key"
run yes save > "$WORK/logk4"
grep -q REGENERATED "$WORK/data/identity-system/ssh/ssh_host_ed25519_key" 2>/dev/null \
    && note "save captures the host key" OK \
    || note "save does not capture the host key" FAIL

# A restore that copied nothing must not say it restored something. Every
# fault in this project so far has been a step that reported success while
# doing nothing, so the log line has to be earned.
setup
mkdir -p "$WORK/data/identity" "$WORK/data/identity-system"
rm -rf "$WORK/root/etc/ssh"
run yes > "$WORK/logk5"
grep -q "host keys restored" "$WORK/logk5" \
    && note "it claims a host key restore that did not happen" FAIL \
    || note "no stored host key means no restore is announced" OK

echo
echo "===== a corrupt or empty stored shadow must not lock the board out ====="
setup
mkdir -p "$WORK/data/identity" "$WORK/data/identity-system"
: > "$WORK/data/identity-system/shadow"
run yes > "$WORK/log3"
grep -q FACTORY "$WORK/root/etc/shadow" \
    && note "an empty stored shadow is ignored" OK \
    || note "an empty stored shadow replaced the working one" FAIL

setup
mkdir -p "$WORK/data/identity" "$WORK/data/identity-system"
printf 'this is not a shadow file\n' > "$WORK/data/identity-system/shadow"
run yes > "$WORK/log4"
grep -q FACTORY "$WORK/root/etc/shadow" \
    && note "a shadow with no root entry is ignored" OK \
    || note "a shadow with no root entry replaced the working one" FAIL

echo
echo "===== no /data means stand aside, not refuse to boot ====="
setup
run no > "$WORK/log5"
grep -q FACTORY "$WORK/root/etc/shadow" \
    && note "without /data the image's own identity is kept" OK \
    || note "without /data the shadow was damaged" FAIL
grep -qi "not mounted" "$WORK/log5" \
    && note "and it says so on the console" OK \
    || note "and it says so on the console" FAIL

echo
echo "===== the script parses ====="
sh -n "$S02" 2>/dev/null && note "sh -n accepts S02identity" OK || note "sh -n accepts S02identity" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
