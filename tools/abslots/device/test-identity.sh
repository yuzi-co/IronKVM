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
# A directory can be bound, a single file cannot: passwd and every editor write
# by rename, and a rename replaces a single-file bind. So /etc/kvm and
# /root/.ssh are bound and change themselves, while /etc/shadow is copied and
# the server writes it back the moment a password change succeeds.
#
#   /data/identity                    bound over /etc/kvm
#   /data/identity-system/root-ssh    bound over /root/.ssh
#   /data/identity-system             shadow, passwd and the host keys, copied
S02=${1:-$(dirname "$0")/S02identity}
[ -f "$S02" ] || { echo "usage: test-identity.sh <S02identity>"; exit 1; }

WORK=$(mktemp -d)

# On Linux the bind mount in S02identity actually succeeds, and every case
# binds again over the same path. Without this the stack of binds outlives the
# test, its own `rm -rf` fails with EBUSY, and the build host collects a mount
# per run. On Windows the bind always fails, so this was invisible there and
# every local run looked clean.
unbind_one() {
    i=0
    while grep -q " $1 " /proc/mounts 2>/dev/null && [ "$i" -lt 32 ]; do
        umount "$1" 2>/dev/null || break
        i=$((i + 1))
    done
}

# Both bind targets. S02identity binds /etc/kvm and /root/.ssh, and a helper
# that only knew about the first would leak the second exactly the way this
# whole helper exists to prevent.
unbind() {
    unbind_one "$WORK/root/root/.ssh"
    unbind_one "$WORK/root/etc/kvm"
}
trap 'unbind; rm -rf "$WORK"' EXIT

# S02identity makes the host keys itself when the board has none, and the real
# ssh-keygen would put them in the host's own /etc/ssh: the script's default
# KEYGEN is an absolute path and -A ignores every variable this suite sets.
# This stands in for it. It takes the same `-A -f PREFIX` the script passes, it
# makes only the keys that are absent the way -A does, and it records each run
# so a case can assert that it did not happen.
#
# `ssh-keygen -A -f PREFIX` was confirmed against the board's own OpenSSH on
# 2026-08-22. RSA costs 20 seconds there, which is why this suite does not call
# the real one.
KEYGEN_STUB="$WORK/ssh-keygen-stub"
cat > "$KEYGEN_STUB" <<'STUB'
#!/bin/sh
prefix=
while [ $# -gt 0 ]; do
    case $1 in
        -f) prefix=$2; shift 2 ;;
        *)  shift ;;
    esac
done
echo "ran" >> "$KEYGEN_LOG"
# Deliberately no mkdir. The real ssh-keygen -A does not make the directory it
# writes into: with it absent it prints "Could not save your private key" for
# each type and still exits 0. Measured on the board on 2026-08-22. Creating it
# here would hide the mkdir the script has to do for itself.
[ -d "$prefix/etc/ssh" ] || exit 0
for t in rsa ecdsa ed25519; do
    [ -s "$prefix/etc/ssh/ssh_host_${t}_key" ] && continue
    printf 'MADE-HERE-%s
' "$t"        > "$prefix/etc/ssh/ssh_host_${t}_key"
    printf 'ssh-%s MADEPUB
' "$t"      > "$prefix/etc/ssh/ssh_host_${t}_key.pub"
done
STUB
chmod +x "$KEYGEN_STUB"
KEYGEN_LOG="$WORK/keygen.log"
export KEYGEN_LOG

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

setup() {
    unbind
    rm -rf "$WORK/root" "$WORK/data"
    mkdir -p "$WORK/root/etc/kvm" "$WORK/root/root/.ssh" "$WORK/root/etc/ssh" "$WORK/data"
    : > "$KEYGEN_LOG"
    printf 'root:$FACTORY:1::::::\n'  > "$WORK/root/etc/shadow"
    printf 'root:x:0:0:root:/root:/bin/sh\n' > "$WORK/root/etc/passwd"
    printf 'image-config\n'           > "$WORK/root/etc/kvm/server.yaml"
    printf 'IMAGE-HOST-KEY\n'         > "$WORK/root/etc/ssh/ssh_host_ed25519_key"
    printf 'ssh-ed25519 IMAGEPUB\n'   > "$WORK/root/etc/ssh/ssh_host_ed25519_key.pub"
    # All three types, because /etc/ssh on a board that has keys carries all
    # three and S50sshd's own guard reads the RSA one. A fixture with only an
    # ed25519 key describes a board sshd would top up behind this script's back.
    printf 'IMAGE-RSA-KEY\n'          > "$WORK/root/etc/ssh/ssh_host_rsa_key"
    printf 'ssh-rsa IMAGEPUB\n'       > "$WORK/root/etc/ssh/ssh_host_rsa_key.pub"
    printf 'IMAGE-ECDSA-KEY\n'        > "$WORK/root/etc/ssh/ssh_host_ecdsa_key"
    printf 'ecdsa-sha2 IMAGEPUB\n'    > "$WORK/root/etc/ssh/ssh_host_ecdsa_key.pub"
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
        ROOTSSH="$WORK/data/identity-system/root-ssh"
        DATA_MOUNTED="$1"
        KEYGEN="$KEYGEN_STUB"
        KEYGEN_PREFIX="$WORK/root"
        export IDENTITY SYSDIR TARGET ETCDIR ROOTHOME SSHDIR ROOTSSH DATA_MOUNTED
        export KEYGEN KEYGEN_PREFIX
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
echo "===== /root/.ssh is a directory bind, so a key edit needs no command ====="
#
# authorized_keys used to be copied out of /data at boot and copied back only by
# `slot identity save`. A command an operator has to remember after editing a
# file is not a mechanism: they do not, nothing tells them it was needed, and
# the next slot switch silently restores the old file.
#
# A single-file bind is not the answer, because an editor writes by rename and
# the rename replaces the bind. A DIRECTORY bind survives that: the rename
# happens inside the bound directory, so it lands on /data either way.

setup
mkdir -p "$WORK/data/identity" "$WORK/data/identity-system/root-ssh"
printf 'ssh-ed25519 AAAA bound\n' > "$WORK/data/identity-system/root-ssh/authorized_keys"
run yes > "$WORK/logb1"

grep -q bound "$WORK/root/root/.ssh/authorized_keys" 2>/dev/null \
    && note "the stored key is visible under /root/.ssh" OK \
    || note "the stored key is visible under /root/.ssh" FAIL

# The real proof, and it only runs where a bind mount works. Write through
# /root/.ssh the way ssh-copy-id would, and it has to appear on /data with no
# further step.
if grep -q " $WORK/root/root/.ssh " /proc/mounts 2>/dev/null; then
    printf 'ssh-ed25519 AAAA added-later\n' >> "$WORK/root/root/.ssh/authorized_keys"
    grep -q added-later "$WORK/data/identity-system/root-ssh/authorized_keys" \
        && note "a key added afterwards lands on /data by itself" OK \
        || note "a key added afterwards did not reach /data" FAIL

    # An editor writes by rename. That is what breaks a single-file bind.
    printf 'ssh-ed25519 AAAA renamed-in\n' > "$WORK/root/root/.ssh/.tmpkey"
    mv "$WORK/root/root/.ssh/.tmpkey" "$WORK/root/root/.ssh/authorized_keys"
    grep -q renamed-in "$WORK/data/identity-system/root-ssh/authorized_keys" \
        && note "and so does a file replaced by rename" OK \
        || note "a rename broke the link to /data" FAIL
else
    note "a key added afterwards lands on /data (no bind mount here)" SKIP
    note "and so does a file replaced by rename (no bind mount here)" SKIP
fi

# Boards built before this change keep their key beside the shadow. It has to
# come across, or the first boot after an upgrade loses key login.
setup
mkdir -p "$WORK/data/identity" "$WORK/data/identity-system"
printf 'ssh-ed25519 AAAA legacy\n' > "$WORK/data/identity-system/authorized_keys"
run yes > "$WORK/logb2"
grep -q legacy "$WORK/data/identity-system/root-ssh/authorized_keys" 2>/dev/null \
    && note "a legacy authorized_keys is migrated into root-ssh" OK \
    || note "a legacy authorized_keys is not migrated, so key login is lost" FAIL
grep -q legacy "$WORK/root/root/.ssh/authorized_keys" 2>/dev/null \
    && note "and it is still readable at /root/.ssh" OK \
    || note "and it is still readable at /root/.ssh" FAIL

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
echo "===== a card whose image ships no host key still gets a stored one ====="
#
# This is the case a real flash produces, and the one the suite used to miss.
# Its fixture wrote a host key into the image, so the seed always had something
# to copy. A real card has none: sshd makes them from S50sshd, 48 scripts after
# S02identity, and the seed is guarded on a directory it has already made, so it
# never runs again. Measured on a card flashed with 1.0.1 on 2026-08-22:
# /data/identity-system/ssh did not exist, and the board would have shown a
# different fingerprint on its other slot with no message to say why.

setup
rm -rf "$WORK/root/etc/ssh"
run yes > "$WORK/logg1"

[ -s "$WORK/root/etc/ssh/ssh_host_ed25519_key" ] \
    && note "a board with no host key gets one before sshd starts" OK \
    || note "a board with no host key gets one before sshd starts" FAIL

grep -q MADE-HERE "$WORK/data/identity-system/ssh/ssh_host_ed25519_key" 2>/dev/null \
    && note "and the key it made is stored on /data" OK \
    || note "the key was made and never stored, so a slot switch loses it" FAIL

grep -q "generating them" "$WORK/logg1" \
    && note "and the console says it generated them" OK \
    || note "and the console says it generated them" FAIL

# The point of storing it is the next boot of the other slot, which has its own
# empty /etc/ssh. It has to come back byte for byte or the fingerprint changed.
kept=$(cat "$WORK/data/identity-system/ssh/ssh_host_ed25519_key" 2>/dev/null)
rm -rf "$WORK/root/etc/ssh"
run yes > "$WORK/logg2"
[ "$(cat "$WORK/root/etc/ssh/ssh_host_ed25519_key" 2>/dev/null)" = "$kept" ] \
    && note "the other slot restores that same key" OK \
    || note "the other slot restores that same key" FAIL

echo
echo "===== an image that does ship a host key has it stored, not replaced ====="

setup
run yes > "$WORK/logg3"
grep -q IMAGE-HOST-KEY "$WORK/data/identity-system/ssh/ssh_host_ed25519_key" 2>/dev/null \
    && note "the image's own host key is what gets stored" OK \
    || note "the image's own host key is what gets stored" FAIL
[ ! -s "$KEYGEN_LOG" ] \
    && note "and no key is generated over it" OK \
    || note "a key was generated although the image carried one" FAIL

echo
echo "===== a board that already stores host keys does not make new ones ====="

setup
mkdir -p "$WORK/data/identity" "$WORK/data/identity-system/ssh"
printf 'DEVICE-HOST-KEY\n'       > "$WORK/data/identity-system/ssh/ssh_host_ed25519_key"
printf 'ssh-ed25519 DEVICEPUB\n' > "$WORK/data/identity-system/ssh/ssh_host_ed25519_key.pub"
run yes > "$WORK/logg4"
[ ! -s "$KEYGEN_LOG" ] \
    && note "a restore runs no generator" OK \
    || note "a restore ran the generator, which would cost 20s on every boot" FAIL

echo
echo "===== a board seeded before this change is repaired on its next boot ====="
#
# The card flashed on 2026-08-22 is in exactly this state: identity-system
# exists, so the seed will not run again, and it holds no ssh directory. Without
# a repair path such a board keeps its unsaved key for ever.

setup
mkdir -p "$WORK/data/identity" "$WORK/data/identity-system"
printf 'ALREADY-RUNNING\n' > "$WORK/root/etc/ssh/ssh_host_ed25519_key"
run yes > "$WORK/logg5"
grep -q ALREADY-RUNNING "$WORK/data/identity-system/ssh/ssh_host_ed25519_key" 2>/dev/null \
    && note "the running key is stored although the seed is long past" OK \
    || note "the running key is stored although the seed is long past" FAIL
grep -q "host keys stored" "$WORK/logg5" \
    && note "and the console says so" OK \
    || note "and the console says so" FAIL

echo
echo "===== no ssh-keygen is not a boot failure ====="
#
# A board without ssh-keygen has no sshd either, so a host key would be of no
# use to it. Standing aside beats stopping the boot of every other service.

setup
rm -rf "$WORK/root/etc/ssh"
(
    IDENTITY="$WORK/data/identity"
    SYSDIR="$WORK/data/identity-system"
    TARGET="$WORK/root/etc/kvm"
    ETCDIR="$WORK/root/etc"
    ROOTHOME="$WORK/root/root"
    SSHDIR="$WORK/root/etc/ssh"
    ROOTSSH="$WORK/data/identity-system/root-ssh"
    DATA_MOUNTED=yes
    KEYGEN="$WORK/no-such-keygen"
    KEYGEN_PREFIX="$WORK/root"
    export IDENTITY SYSDIR TARGET ETCDIR ROOTHOME SSHDIR ROOTSSH DATA_MOUNTED
    export KEYGEN KEYGEN_PREFIX
    sh "$S02" start
) > "$WORK/logg6" 2>&1
rc=$?
[ "$rc" -eq 0 ] \
    && note "the boot continues without ssh-keygen" OK \
    || note "no ssh-keygen stopped the boot with rc=$rc" FAIL
grep -q "no ssh-keygen to make them" "$WORK/logg6" \
    && note "and it says which tool is missing" OK \
    || note "and it says which tool is missing" FAIL
grep -q "identity-system" "$WORK/logg6" \
    && note "the rest of the identity is still restored" OK \
    || note "the rest of the identity is still restored" FAIL

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
    exit 1
fi
