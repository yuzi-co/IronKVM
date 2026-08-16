#!/bin/sh
# Hardware acceptance test for the boot-script rollback.
#
#   acceptance-initd-rollback.sh root@<device>
#
# DO NOT RUN THIS UNATTENDED. It deliberately makes the board unreachable and
# relies on the mechanism under test to bring it back.
#
# Before running it, confirm both backstops:
#   - the recovery slot is populated  (dumpe2fs -h /dev/mmcblk0p5)
#   - the ACM console answers on the managed host (/dev/ttyACM0)
# There is no remote power cycle. If both backstops are gone the only recovery
# is pulling the card.
#
# WHAT IT BREAKS, AND WHY THAT ONE
#
# It breaks S30eth, not S01fs. The watchdog acts on REACHABILITY, and a board
# with a broken S01fs usually still answers ssh, so the watchdog would call it
# healthy and never roll anything back. Losing the network costs both doors,
# which is the state the rollback exists for.
#
# The script that replaces S30eth is syntactically valid on purpose. install.sh
# refuses one that fails `sh -n`, so a syntax error would test that guard
# instead of this one.
#
# WHAT SHOULD HAPPEN
#
#   boot 1   S00awatchdog arms, S30eth exits before configuring the network
#            the board is unreachable, the watcher waits out its deadline
#            escalate finds an unspent manifest, restores /etc/init.d, reboots
#   boot 2   S30eth is the original again, the board comes back
#
# If the restore does nothing, escalate falls through to the recovery marker and
# the next boot lands in recovery, which answers ssh. That is the second
# backstop, and it is why this is safe to run at all.

set -u

DEV=${1:?usage: acceptance-initd-rollback.sh root@<device>}
DEADLINE_WAIT=${DEADLINE_WAIT:-900}

say()  { echo "==> $*"; }
fail() { echo "FAIL: $*" >&2; exit 1; }

ssh_dev() { timeout 30 ssh -o ConnectTimeout=8 -o BatchMode=yes "$DEV" "$@" 2>/dev/null; }

say "preflight"
ssh_dev true || fail "the board does not answer ssh; fix that before breaking it"

RECOVERY=$(ssh_dev 'dumpe2fs -h /dev/mmcblk0p5 2>/dev/null | grep -c "Filesystem volume"')
[ "$RECOVERY" = "1" ] || fail "the recovery slot is empty; there would be no backstop"

ssh_dev 'grep -q "initd rollback" /etc/init.d/S00awatchdog' \
    || fail "/etc/init.d/S00awatchdog has no rollback block; deploy it first"
ssh_dev '[ -x /kvmapp/system/install.sh ]' \
    || fail "/kvmapp/system/install.sh is missing; deploy it first"

ORIGINAL=$(ssh_dev 'sha256sum /etc/init.d/S30eth | cut -d" " -f1')
[ -n "$ORIGINAL" ] || fail "cannot read /etc/init.d/S30eth"
say "S30eth is $ORIGINAL"
say "slot: $(ssh_dev 'slot status | tr "\n" " "')"

say "staging a valid script that brings up no network"
ssh_dev 'rm -rf /tmp/acceptance-src && mkdir -p /tmp/acceptance-src && cat > /tmp/acceptance-src/S30eth <<"EOS"
#!/bin/sh
# Deliberately inert, for the boot-script rollback acceptance test.
# Valid shell, so install.sh accepts it; configures nothing, so the board loses
# its only way in and the watchdog has to undo this.
echo "acceptance: S30eth doing nothing" > /dev/kmsg 2>/dev/null
exit 0
EOS
chmod 755 /tmp/acceptance-src/S30eth' || fail "could not stage the broken script"

say "running the install hook"
ssh_dev 'INSTALL_SRC=/tmp/acceptance-src /kvmapp/system/install.sh' || fail "install.sh failed"

ssh_dev '[ -f /root/.ironkvm/initd-backup/manifest ]' \
    || fail "install.sh left no manifest, so there is nothing to roll back"
MANIFEST=$(ssh_dev 'cat /root/.ironkvm/initd-backup/manifest')
say "manifest: $MANIFEST"
echo "$MANIFEST" | grep -q '^S30eth yes$' || fail "the manifest does not record S30eth as replaced"

say "rebooting; the board should go away and come back on its own"
ssh_dev 'sync; reboot' > /dev/null 2>&1

start=$(date +%s)
gone=0
back=0
while [ $(( $(date +%s) - start )) -lt "$DEADLINE_WAIT" ]; do
    if ssh_dev true; then
        [ "$gone" = 1 ] && { back=1; break; }
    else
        gone=1
    fi
    sleep 10
done

elapsed=$(( $(date +%s) - start ))
[ "$gone" = 1 ] || fail "the board never went away; the break did not take effect"
[ "$back" = 1 ] || fail "the board did not come back within ${DEADLINE_WAIT}s. Reach it over the ACM console on the managed host"

say "the board returned after ${elapsed}s"

say "verifying the rollback rather than assuming it"
NOW=$(ssh_dev 'sha256sum /etc/init.d/S30eth | cut -d" " -f1')
[ "$NOW" = "$ORIGINAL" ] || fail "S30eth is $NOW, expected the original $ORIGINAL"

ssh_dev '[ -f /root/.ironkvm/initd-backup/manifest.done ]' \
    || fail "the manifest was not spent, so it would undo the next update too"
ssh_dev '[ ! -f /root/.ironkvm/initd-backup/manifest ]' \
    || fail "an unspent manifest is still present"

ssh_dev 'grep -q "restored the previous" /watchdog.log' \
    || fail "the watchdog log does not record a restore; the board may have come back another way"

say "watchdog log:"
ssh_dev 'tail -6 /watchdog.log' | sed 's/^/    /'

ssh_dev '[ ! -f /boot/recovery ]' \
    || fail "the recovery marker is set, so the board fell back rather than repairing itself"

say "slot: $(ssh_dev 'slot status | tr "\n" " "')"
say "PASS: the board undid the update by itself in ${elapsed}s and did not need recovery"
