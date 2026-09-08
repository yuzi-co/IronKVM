#!/bin/sh
# Check that kvm_vision.cpp writes its small files without a shell.
#
#   test-vision-no-shell.sh [path/to/kvm_vision.cpp]
#
# Not destructive: the source is only read.
#
# system(3) is not a way to write a file. It sets SIGINT and SIGQUIT to SIG_IGN
# for the whole process, runs the command, and puts the old handlers back. Two
# threads that call it at once race on that restore: the second one saves
# SIG_IGN as the handler to restore, the first one restores the real handler,
# and the second one then puts SIG_IGN back for good. The process keeps running
# and can no longer be interrupted.
#
# The server on this board does ignore SIGINT, measured from /proc/<pid>/status
# on 2026-08-19 and recorded in S95nanokvm. That observation is not evidence for
# the race above, because it already has a sufficient cause of its own: both
# services start as background jobs of a shell without job control, and POSIX
# makes such a shell set SIGINT to SIG_IGN in the child. The rule below removes
# the second mechanism, the one that takes SIGINT away from a process that had
# it, and the init script still has to send SIGTERM.
#
# The calls were all of the form `echo <value> > <path>`, and the HDMI detection
# thread and the watchdog thread run them while the main thread is still in
# kvmv_init. An open, a write and a close do the same work, reach no shell, cost
# no fork on a core a capture session already saturates, and touch no signal
# disposition.
#
# Two calls in this file run a program rather than write a file. They are named
# in the allowlist below, they stay, and they are checked for presence so this
# suite fails if somebody removes one and leaves the rule looking satisfied.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
SRC=${1:-$ROOT/support/sg2002/additional/kvm/src/kvm_vision.cpp}

[ -f "$SRC" ] || { echo "missing: $SRC"; exit 2; }

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

# The two shells this file is allowed to run. Each is a program invocation, not
# a file write, and neither runs on a thread's repeating path: the first ends
# the board, and the second runs once while kvmv_init reads the HDMI version.
ALLOWED='system("reboot");
system("/kvmapp/system/init.d/S15kvmhwd get_hdmi_version");'

echo "===== no shell writes a file ====="

# Every system() call that is not commented out, with its line number.
#
# Comments are dropped by their first non-space characters: // for a retired
# call left as a note, and / or * for the block comment on the helper below,
# which has to be able to name system(3) without failing this check. What is
# left has to look like a statement, so the line must also end in `);`.
calls=$(grep -n 'system[ ]*(' "$SRC" \
    | grep -v ':[[:space:]]*\(//\|\*\|/\*\)' \
    | grep ');[[:space:]]*$' || true)

if [ -z "$calls" ]; then
    note "kvm_vision.cpp calls system() at all" FAIL
else
    unexpected=0
    for n in $(printf '%s\n' "$calls" | cut -d: -f1); do
        text=$(sed -n "${n}p" "$SRC" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')
        case "
$ALLOWED
" in
            *"
$text
"*) continue ;;
        esac
        unexpected=$((unexpected + 1))
        note "line $n runs a shell: $text" FAIL
    done

    [ "$unexpected" = 0 ] && note "no system() call outside the allowlist" OK
fi

echo "===== the two allowed shells are still here ====="

# Without this the rule above passes trivially on a file that stopped calling
# system() because somebody deleted the reboot or the version probe.
printf '%s\n' "$ALLOWED" | while IFS= read -r want; do
    [ -n "$want" ] || continue
    if grep -qF "$want" "$SRC"; then
        printf '  %-64s %s\n' "still calls: $want" OK
    else
        printf '  %-64s %s\n' "still calls: $want" FAIL
    fi
done

missing=$(printf '%s\n' "$ALLOWED" | while IFS= read -r want; do
    [ -n "$want" ] || continue
    grep -qF "$want" "$SRC" || echo x
done | grep -c . )
fails=$((fails + missing))

echo "===== the file writes go through one helper ====="

# The replacement is a single helper rather than an open/write/close at each
# site, so a later change cannot reintroduce the shell at one of seventeen
# places without the reviewer seeing a new pattern.
#
# The name is matched with its opening parenthesis, so a renamed helper does not
# satisfy this by leaving the old name as a prefix of the new one. Sixteen call
# sites were replaced; the floor below is well under that and exists to reject a
# file that kept the definition and lost the callers.
used=$(grep -c 'write_small_file[ ]*(' "$SRC")
if grep -q 'static bool write_small_file[ ]*(' "$SRC" && [ "$used" -ge 10 ]; then
    note "write_small_file is defined and used ($used references)" OK
else
    note "write_small_file is defined and used ($used references)" FAIL
fi

# sync(2) is the same argument in one line: the shell was only ever there to
# reach a system call that this process can make directly.
if grep -q 'system[ ]*([ ]*"sync"' "$SRC"; then
    note "sync is a call, not a shell" FAIL
else
    note "sync is a call, not a shell" OK
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
    exit 0
fi
echo "$fails case(s) FAILED"
exit 1
