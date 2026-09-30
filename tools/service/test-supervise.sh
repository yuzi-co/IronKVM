#!/bin/sh
# Exercise the supervisor's decisions, taken straight out of the script that
# ships so the test cannot drift from it.
#
#   test-supervise.sh [path-to-S98supervise] [path-to-S95nanokvm]
#
# Not destructive: nothing is started or killed. The probes are stubbed.
SV=${1:-$(dirname "$0")/S98supervise}
S95=${2:-$(dirname "$0")/../../kvmapp/system/init.d/S95nanokvm}
[ -f "$SV" ] || { echo "usage: test-supervise.sh <S98supervise>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

sed -n '/^# --- decide ---$/,/^# --- end decide ---$/p'   "$SV" > "$WORK/decide.sh"
sed -n '/^# --- backoff ---$/,/^# --- end backoff ---$/p' "$SV" > "$WORK/backoff.sh"
sed -n '/^# --- cure ---$/,/^# --- end cure ---$/p'       "$SV" > "$WORK/cure.sh"
sed -n '/^# --- escalate ---$/,/^# --- end escalate ---$/p' "$SV" > "$WORK/escalate.sh"
sed -n '/^# --- count ---$/,/^# --- end count ---$/p' "$SV" > "$WORK/count.sh"
sed -n '/^# --- act ---$/,/^# --- end act ---$/p' "$SV" > "$WORK/act.sh"
sed -n '/^# --- ion ---$/,/^# --- end ion ---$/p' "$SV" > "$WORK/ion.sh"
sed -n '/^# --- updating ---$/,/^# --- end updating ---$/p' "$SV" > "$WORK/updating.sh"
sed -n '/^# --- ssh door ---$/,/^# --- end ssh door ---$/p' "$SV" > "$WORK/sshdoor.sh"
sed -n '/^# --- procs ---$/,/^# --- end procs ---$/p' "$SV" > "$WORK/procs.sh"
[ -s "$WORK/procs.sh" ] || { echo "could not extract the procs block"; exit 1; }
[ -s "$WORK/sshdoor.sh" ] || { echo "could not extract the ssh door block"; exit 1; }
[ -s "$WORK/updating.sh" ] || { echo "could not extract the updating block"; exit 1; }
[ -s "$WORK/decide.sh" ]  || { echo "could not extract the decide block"; exit 1; }
[ -s "$WORK/backoff.sh" ] || { echo "could not extract the backoff block"; exit 1; }
[ -s "$WORK/cure.sh" ]    || { echo "could not extract the cure block"; exit 1; }
[ -s "$WORK/escalate.sh" ] || { echo "could not extract the escalate block"; exit 1; }
[ -s "$WORK/count.sh" ] || { echo "could not extract the count block"; exit 1; }
[ -s "$WORK/act.sh" ] || { echo "could not extract the act block"; exit 1; }
[ -s "$WORK/ion.sh" ] || { echo "could not extract the ion block"; exit 1; }

fails=0
note() { printf '  %-60s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

echo "===== crashed, or stopped on purpose? ====="
# The distinction costs nothing to make: `S95nanokvm stop` removes /tmp/server
# after killing the process, so the binary's presence is the operator's intent.
# Without this the supervisor would fight every deliberate stop.
# serving has three answers, so the probe column has three values. Anything
# that is not yes or no stands for "the probe could not run at all", which is
# what the shipped serving reports when curl is missing. The loop runs the probe
# and hands action its status.
decide_case() {
    desc="$1"; binary="$2"; running="$3"; serving="$4"; unhealthy="$5"; want="$6"
    got=$(BIN="$binary" RUN="$running" SRV="$serving" UNW="$unhealthy" WORK="$WORK" sh -c '
        binary_present()  { [ "$BIN" = yes ]; }
        process_running() { [ "$RUN" = yes ]; }
        unhealthy_for()   { echo "$UNW"; }
        . "$WORK/decide.sh"
        case "$SRV" in yes) ans=0 ;; no) ans=1 ;; *) ans=2 ;; esac
        action "$ans"
        echo "$ACTION"
    ')
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

#           binary running serving unhealthy-for  want
decide_case "running and answering"          yes yes yes 0   healthy
decide_case "binary staged but no process"   yes no  no  0   restart
decide_case "stopped on purpose"             no  no  no  0   stopped
# Odd but real during a restart: the old process is still dying while /tmp/server
# has already been removed. Interfering there would race S95nanokvm.
decide_case "no binary, process still dying" no  yes no  0   healthy

echo
echo "  --- alive but not answering: a hang"
# The failure the supervisor could not see. A process that is up and not serving
# looks healthiest of all from outside, and pidof cannot tell the difference.
#
# A board reboot is the wrong response: it costs 17s, risks the boot path, and can
# land on a slot without any of this. Restarting the server is cheaper and safer,
# so the hardware watchdog is reserved for a kernel lockup that userspace cannot
# act on at all.
decide_case "just stopped answering, still inside grace" yes yes no 10  healthy
decide_case "not answering for a minute: hung"           yes yes no 60  hung
decide_case "not answering for a long time"              yes yes no 600 hung

# Never on ambiguity. A curl that cannot run, a probe that errors, a slow start
# under heavy IO - none of those are worth killing a working KVM for, so the
# grace period has to be generous and the default has to be inaction.
decide_case "answering again before the threshold"       yes yes yes 59  healthy

echo
echo "  --- an answer makes pidof unnecessary"
# The healthy pass is the one that runs every INTERVAL for the life of the board,
# and pidof is a fork and a walk of /proc. When the probe answered, the process
# is up, so action must not ask. A pidof that would say otherwise cannot change
# the verdict: something answered on the port.
pidof_asked() {   # $1 = probe status handed to action; prints verdict and pidof count
    PS="$1" WORK="$WORK" sh -c '
        : > "$WORK/pidof.calls"
        binary_present()  { true; }
        process_running() { echo x >> "$WORK/pidof.calls"; true; }
        unhealthy_for()   { echo 0; }
        . "$WORK/decide.sh"
        action "$PS"
        echo "$ACTION $(wc -l < "$WORK/pidof.calls" | tr -d " ")"
    '
}
got=$(pidof_asked 0)
[ "$got" = "healthy 0" ] && note "answered: healthy, pidof not asked -> $got" OK \
                         || note "answered: [$got], want [healthy 0]" FAIL
got=$(pidof_asked 1)
[ "$got" = "healthy 1" ] && note "silent: pidof still decides -> $got" OK \
                         || note "silent: [$got], want [healthy 1]" FAIL
got=$(pidof_asked 2)
[ "$got" = "healthy 1" ] && note "probe could not run: pidof still decides -> $got" OK \
                         || note "probe could not run: [$got], want [healthy 1]" FAIL
# The shortcut is for an answer only. Nothing answered and nothing is running is
# a crash, and it must still read as one.
decide_case "silent and no process: pidof is what says so" yes no no 0 restart

echo
echo "===== standing off while an update is in progress ====="
# On 2026-08-17 an update stopped the server, this script noticed 5s later, and
# started it 10 seconds into the install. That server died 3 seconds after,
# because its own files were being moved out from under it, and the board
# rebooted. The supervisor has to leave a running update alone.
#
# Two things bound the stand-off, because a marker nothing clears is a
# supervisor that never supervises again. The marker lives in tmpfs, so a reboot
# clears it, and it is ignored once it is older than the window.
updating_case() {
    desc="$1"; age="$2"; want="$3"
    got=$(AGE="$age" WORK="$WORK" sh -c '
        M=$WORK/marker
        rm -f "$M"
        [ "$AGE" = absent ] || date +%s > "$M"
        UPDATE_MARKER=$M UPDATE_STANDOFF=300
        export UPDATE_MARKER UPDATE_STANDOFF
        . "$WORK/updating.sh"
        if updating; then echo yes; else echo no; fi
    ' 2>/dev/null)
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

# Written the same way the shipped script sees it, because touch -d with an
# arithmetic expression is not portable to busybox.
updating_age() {
    desc="$1"; seconds="$2"; want="$3"
    got=$(SEC="$seconds" WORK="$WORK" sh -c '
        M=$WORK/marker
        rm -f "$M"
        echo $(( $(date +%s) - SEC )) > "$M"
        UPDATE_MARKER=$M UPDATE_STANDOFF=300
        export UPDATE_MARKER UPDATE_STANDOFF
        . "$WORK/updating.sh"
        if updating; then echo yes; else echo no; fi
    ' 2>/dev/null)
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

updating_case "no marker at all"                    absent no
updating_case "a marker written just now"           fresh  yes
updating_age  "a marker 30s old, inside the window" 30     yes
updating_age  "a marker 299s old, at the edge"      299    yes
updating_age  "a marker 301s old, past the window"  301    no
updating_age  "a marker an hour old"                3600   no

# A marker whose age cannot be read must resume supervision, not suspend it.
# Standing off for ever means a dead server nothing restarts, and only a reboot
# clears that; acting early only risks the fault this block prevents, which is
# rare and recoverable. The dangerous direction is the one that fails silent.
got=$(WORK="$WORK" sh -c '
    M=$WORK/marker; rm -f "$M"; : > "$M"
    UPDATE_MARKER=$M UPDATE_STANDOFF=300
    export UPDATE_MARKER UPDATE_STANDOFF
    . "$WORK/updating.sh"
    if updating; then echo yes; else echo no; fi
' 2>/dev/null)
[ "$got" = no ] && note "an empty marker resumes supervision -> $got" OK \
                || note "an empty marker resumes supervision -> $got, want no" FAIL

got=$(WORK="$WORK" sh -c '
    M=$WORK/marker; rm -f "$M"; echo "not-a-number" > "$M"
    UPDATE_MARKER=$M UPDATE_STANDOFF=300
    export UPDATE_MARKER UPDATE_STANDOFF
    . "$WORK/updating.sh"
    if updating; then echo yes; else echo no; fi
' 2>/dev/null)
[ "$got" = no ] && note "a marker holding rubbish resumes supervision -> $got" OK \
                || note "a marker holding rubbish resumes supervision -> $got, want no" FAIL

# The age comes from the marker's CONTENTS, never from its mtime. busybox on the
# device is built without FEATURE_STAT_FORMAT: `stat -c %Y` reports
# "unrecognized option: c" there, so the age could never be read and the
# stand-off never engaged. The container these tests run in has a busybox with
# the feature, which is why this passed everywhere except on hardware.
# Comments are stripped first: the reason this rule exists is written in the
# script, and a guard that fires on its own explanation is a guard that has to
# be deleted the moment somebody documents anything.
grep -v '^[[:space:]]*#' "$SV" | grep -q 'stat -c' \
    && note "the age is not read with stat -c, which the device lacks" FAIL \
    || note "the age is not read with stat -c, which the device lacks" OK

echo
echo "  --- and the decision uses it"
# The stand-off is worth nothing unless action() consults it. The case that
# matters is exactly the one that caused the fault: binary staged, no process,
# which without this returns restart.
updating_decides() {
    desc="$1"; marker="$2"; want="$3"
    got=$(MK="$marker" WORK="$WORK" sh -c '
        M=$WORK/marker; rm -f "$M"
        [ "$MK" = yes ] && date +%s > "$M"
        UPDATE_MARKER=$M UPDATE_STANDOFF=300
        export UPDATE_MARKER UPDATE_STANDOFF
        binary_present()  { true; }
        process_running() { false; }
        unhealthy_for()   { echo 0; }
        . "$WORK/updating.sh"
        . "$WORK/decide.sh"
        action 1
        echo "$ACTION"
    ' 2>/dev/null)
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

updating_decides "binary staged, no process, update running" yes updating
updating_decides "binary staged, no process, no update"      no  restart

echo
echo "  --- a probe that cannot run is not evidence of anything"
# serving reports a third answer for "curl is missing, so nothing was measured".
# action must read that as serving, at any silence, forever. The alternative is
# that a board without curl reports hung on every poll, and the supervisor then
# kills and restarts a perfectly healthy KVM once a minute - which is worse than
# the reboot cycle this third answer exists to close.
decide_case "the probe cannot run, exactly at the threshold" yes yes unavailable 60   healthy
decide_case "the probe cannot run, hours of silence"         yes yes unavailable 9999 healthy

echo
echo "===== curing a hang means killing it first ====="
# S95nanokvm's restart uses killall, which is SIGTERM. A wedged process may never
# act on that, and while it lives it holds port 80 - so the replacement could not
# bind and the hang would survive its own cure.
got=$(WORK="$WORK" sh -c '
    force_kill()   { echo "kill"; }
    wait_gone()    { echo "waited"; }
    full_restart() { echo "restart"; }
    . "$WORK/cure.sh"
    cure_hung
' | tr '
' ' ')
[ "$got" = "kill waited restart " ]     && note "SIGKILL, wait for it to go, then the normal restart" OK     || note "order was [$got], want [kill waited restart ]" FAIL

# killall -9 cannot clear uninterruptible sleep. A reader of /proc/cvitek/vb
# blocks in D state on this board, so the one hang that most needs a reboot is
# exactly the hang cure_hung cannot fix - and the answer used to be discarded.
got=$(WORK="$WORK" sh -c '
    force_kill()   { echo "kill"; }
    wait_gone()    { echo "waited"; return 1; }
    full_restart() { echo "restart"; }
    . "$WORK/cure.sh"
    cure_hung; echo "rc=$?"
' | tr '\n' ' ')
[ "$got" = "kill waited restart rc=1 " ] \
    && note "a process that will not die still restarts, and says so" OK \
    || note "order was [$got], want [kill waited restart rc=1 ]" FAIL

echo
echo "===== the retry delay grows and is capped ====="
# A binary that can never start must not be retried in a tight loop: that burns
# the one core and writes a log line every pass. It must also never give up,
# because this is the device you reach for when nothing else answers.
got=$(WORK="$WORK" sh -c '. "$WORK/backoff.sh"; d=$(first_delay); i=0; while [ $i -lt 8 ]; do printf "%s " "$d"; d=$(next_delay "$d"); i=$((i+1)); done')
want="5 10 20 40 60 60 60 60 "
[ "$got" = "$want" ] && note "delay walks 5 10 20 40 then holds at 60" OK \
                     || note "got [$got] want [$want]" FAIL

got=$(WORK="$WORK" sh -c '. "$WORK/backoff.sh"; next_delay 0')
[ "$got" -gt 0 ] && note "a zero delay still advances (no tight loop)" OK \
                 || note "next_delay 0 = $got" FAIL

echo
echo "===== a run that lasted resets the delay ====="
# Otherwise a board that crashes once a day would creep to the cap and stay
# there, so the next real crash waits a minute for no reason.
reset_case() {
    desc="$1"; uptime="$2"; current="$3"; want="$4"
    got=$(WORK="$WORK" sh -c ". \"\$WORK/backoff.sh\"; delay_after_run $uptime $current")
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

reset_case "ran 5 minutes then died, so start over"   300 60 5
reset_case "died again after 3 seconds, so keep backing off" 3 20 40
reset_case "died right at the threshold"              60 40 5

echo
echo "===== a restarted server still reports where libkvm fails ====="
# The supervisor restarts a crashed server itself, rather than through
# S95nanokvm, because a full restart also stops and restarts kvm_system for
# nothing. It used to copy 36MB back into tmpfs as well.
# That shortcut has to carry the redirection with it.
#
# libkvm reports a capture pipeline that does not start with printf, and that
# output is the only record of the failure. A server started with its output on
# /dev/null is a server nobody can debug, and a crash is when the record matters
# most. S98vidiag would go on reading a file that nothing writes to, and the
# file would still be there, so nothing would look wrong.
if grep -q '"\$SERVER_BIN" < /dev/null >> "\$SERVER_LOG" 2>&1 &' "$SV"; then
    note "the crash restart sends the server's output to the log" OK
else
    note "the crash restart discards the server's output" FAIL
fi

# One path, spelled in two scripts, drifts. S98vidiag reads one file, so a
# second spelling here means the collector follows a file nobody writes.
sv_log=$(sed -n 's/^SERVER_LOG=\(.*\)$/\1/p' "$SV")
s95_log=$(sed -n 's/^SERVER_LOG=\(.*\)$/\1/p' "$S95")
if [ -z "$sv_log" ]; then
    note "the supervisor never names the log" FAIL
elif [ "$sv_log" = "$s95_log" ]; then
    note "both scripts name $sv_log" OK
else
    note "the supervisor says $sv_log, S95nanokvm says $s95_log" FAIL
fi

# kvm_system drives the OLED. Pointing it at the server's log would mix two
# programs into one record, so its output goes to syslog under its own tag,
# and to /dev/null on a board without logger. See "kvm_system's output goes to
# syslog" below for how it behaves.
if grep -q '"\$SYSTEM_BIN" < /dev/null > /dev/null 2>&1 &' "$SV" &&
   grep -q '"\$SYSTEM_BIN" < /dev/null >&8 2>&8 8>&- &' "$SV"; then
    note "kvm_system writes to syslog, or to /dev/null without it" OK
else
    note "kvm_system no longer starts the way it did" FAIL
fi

# The gap between S95nanokvm killing kvm_system and starting it again is five
# seconds wide at most, and this poll used to fill it. Two of them then drove
# the OLED at once and wrote each other's pixels to the wrong column, which is
# corruption no redraw clears because fields are drawn only when they change.
if grep -q 'if \[ "\$state" != updating \] && ! system_running && system_binary_present; then' "$SV"; then
    note "kvm_system is not started during a stand-off" OK
else
    note "kvm_system can still be started in the middle of a restart" FAIL
fi

echo
echo "===== when restarting cannot work, reboot ====="
# S98supervise restarts and never reboots, which is right for a hung server and
# wrong for an exhausted ION carveout: the allocation is leaked inside the kernel
# modules, no userspace action frees it, and the server dies again in under a
# second. On 2026-08-04 that produced 23 restarts over 22 minutes into a
# guaranteed failure, and it would have continued indefinitely.
escalate_case() {
    desc="$1"; verdict="$2"; short="$3"; cures="$4"; up="$5"; ever="$6"; want="$7"
    got=$(WORK="$WORK" sh -c ". \"\$WORK/escalate.sh\"; should_reboot $verdict $short $cures $up $ever")
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

#              desc                                            verdict short cures up   ever want
escalate_case "five short runs on a board that has been up"    restart 5  0  3600 yes  yes
escalate_case "four short runs is not a loop yet"              restart 4  0  3600 yes  no
escalate_case "no short runs at all"                           restart 0  0  3600 yes  no
escalate_case "two failed cures on a board that has been up"   hung    0  2  3600 yes  yes
escalate_case "one failed cure is not enough"                  hung    0  1  3600 yes  no

# A deliberate stop is the operator's intent, and a healthy server has nothing
# wrong with it. Neither is ever a reason to take the KVM away from someone.
escalate_case "a deliberate stop is never escalated"           stopped 99 99 3600 yes  no
escalate_case "a healthy server is never escalated"            healthy 99 99 3600 yes  no

echo
echo "  --- a board that has never served since boot is never rebooted"
# The floor sets the period of a reboot cycle. It does not prevent one: a fault
# present from boot keeps producing restart or hung verdicts, the counters do not
# survive a reboot, and the moment uptime crosses the floor the same sequence
# escalates again. Measured against the shipped loop, that is every 10.5 minutes
# forever - a board that answers nothing and cannot be repaired.
#
# A reboot cures a board that worked and then broke. It is never a cure for a
# server that has not answered once since this boot: an unreadable certificate,
# a binary that skipped patchelf, a truncated copy from a full SD card - all of
# them leave the binary staged and executable, so the verdict is restart, and
# all of them survive a reboot unchanged.
escalate_case "a crash loop that never answered since boot"    restart 9  0  3600 no   no
escalate_case "a hang that never answered since boot"          hung    0  9  3600 no   no

echo
echo "  --- the floor: the one check between this and a board that must be opened"
# A board that crash-loops out of boot reaches the escalation at roughly five
# minutes of uptime: 135s of backoff plus five runs of at most 30s, on top of a
# 20s boot. The floor sits at ten minutes, so that case is blocked with about
# twice the margin it needs - firmly, not narrowly.
#
# The consequence is deliberate. After one reboot the fault either goes away, or
# it returns at low uptime and no second reboot happens: the board stays up and
# reachable over ssh for a person to work on, which is what happens today anyway.
# A leak that refills faster than the floor is a leak a reboot cannot cure.
escalate_case "crash loop one second under the floor"          restart 5  0  599  yes  no
escalate_case "crash loop straight out of boot"                restart 9  0  0    yes  no
escalate_case "hang one second under the floor"                hung    0  2  599  yes  no
escalate_case "hang straight out of boot"                      hung    0  9  0    yes  no
escalate_case "crash loop exactly at the floor"                restart 5  0  600  yes  yes
escalate_case "hang exactly at the floor"                      hung    0  2  600  yes  yes

echo
echo "  --- a guard that inverts its own meaning on bad input is not a guard"
# `[ "" -lt 600 ]` is an error, not a false comparison. The `if` is false, so the
# floor is skipped and the board reboots at whatever uptime it has. `up` comes
# from `cut -d. -f1 /proc/uptime`, and any failure to fork that - memory
# pressure, PID exhaustion, a stalled filesystem - yields an empty string. A typo
# in SUPERVISE_REBOOT_FLOOR opens the same hole from the other side.
#
# These cases cannot use escalate_case: field splitting cannot produce an empty
# argument, and the threshold ones have to set a variable before sourcing.
got=$(WORK="$WORK" sh -c '. "$WORK/escalate.sh"; should_reboot hung 0 2 "" yes')
[ "$got" = no ] && note "an uptime that could not be read -> $got" OK \
                || note "an uptime that could not be read -> $got, want no" FAIL

got=$(WORK="$WORK" sh -c '. "$WORK/escalate.sh"; should_reboot restart 5 0 abc yes')
[ "$got" = no ] && note "an uptime that is not a number -> $got" OK \
                || note "an uptime that is not a number -> $got, want no" FAIL

got=$(WORK="$WORK" sh -c 'REBOOT_FLOOR=ten; . "$WORK/escalate.sh"; should_reboot restart 5 0 10 yes')
[ "$got" = no ] && note "a floor that is not a number -> $got" OK \
                || note "a floor that is not a number -> $got, want no" FAIL

# Digits are not a range. busybox `[` answers "out of range" on a value wider
# than the comparison can hold, and that is an error, so it skips the floor by
# exactly the route an empty string does. /proc/uptime cannot produce one. A
# typo in SUPERVISE_REBOOT_FLOOR can, and it would turn an operator's "never
# reboot this board" into "reboot this board at any uptime".
got=$(WORK="$WORK" sh -c '. "$WORK/escalate.sh"; should_reboot restart 5 0 99999999999999999999 yes')
[ "$got" = no ] && note "an uptime too wide for the comparison -> $got" OK \
                || note "an uptime too wide for the comparison -> $got, want no" FAIL

got=$(WORK="$WORK" sh -c 'REBOOT_FLOOR=99999999999999999999; . "$WORK/escalate.sh"; should_reboot restart 5 0 3600 yes')
[ "$got" = no ] && note "a floor too wide for the comparison -> $got" OK \
                || note "a floor too wide for the comparison -> $got, want no" FAIL

echo
echo "===== counting the runs that did not last ====="
# The threshold is 30s and not the one second the process actually survives.
# watch_loop sleeps INTERVAL between checks and resets `started` after each
# start, so `ran` is quantised to the poll and never reports below about five
# seconds. A five-second threshold would be an assertion that can never be true.
short_case() {
    desc="$1"; ran="$2"; current="$3"; want="$4"
    got=$(WORK="$WORK" sh -c ". \"\$WORK/count.sh\"; next_short_runs $ran $current")
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

short_case "died as soon as it started"        1   0 1
short_case "died at the poll interval"         5   0 1
short_case "just under the threshold"          29  2 3
short_case "exactly at the threshold, so reset" 30 4 0
short_case "a run that lasted, so reset"       300 4 0

echo
echo "===== counting the cures that did not work ====="
# A hung verdict that arrives after a cure proves that cure did not work. The
# first hung verdict of a fault follows no cure at all, so it counts nothing -
# otherwise the very first hang would be one step from a reboot.
cures_case() {
    desc="$1"; cures="$2"; current="$3"; want="$4"
    got=$(WORK="$WORK" sh -c ". \"\$WORK/count.sh\"; next_failed_cures $cures $current")
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

cures_case "the first hang, nothing tried yet"  0 0 0
cures_case "hung again after one cure"          1 0 1
cures_case "hung again after two cures"         2 1 2

echo
echo "===== clearing the counters needs an answer, not just a verdict ====="
# action() reports healthy for a process that is up and not answering yet,
# inside HANG_AFTER, and the hang branch resets LAST_OK after every cure - so
# the very next poll reports healthy too. Clearing on the verdict name alone
# would wipe the counters between every cure and the hung verdict that should
# follow it, and the counted hang escalation could never reach its threshold.
clear_case() {
    desc="$1"; verdict="$2"; answered="$3"; want="$4"
    got=$(WORK="$WORK" sh -c ". \"\$WORK/count.sh\"; should_clear $verdict $answered && echo yes || echo no")
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

#          desc                                                 verdict answered want
clear_case "answering: the fault is over"                       healthy yes      yes
clear_case "up but not answering yet, inside the grace"         healthy no       no
# The supervisor's own cure is S95nanokvm restart, which removes /tmp/server and
# stages it again. For that whole window there is no process and no binary, so
# the verdict is the one a deliberate stop gives - and clearing there wipes the
# cure counters. When the stage was a 36MB copy, that was on exactly the slow SD
# card the fault arrives with. Measured: a
# 20-second re-stage escalated on the third hung verdict, a 40-second re-stage
# never escalated at all. An operator who stops the server and brings it back
# produces an answering healthy poll, which clears the counters safely.
clear_case "mid-cure, while S95nanokvm re-stages /tmp"          stopped no       no
clear_case "stopped but something answered - not a cure signal" stopped yes      no
clear_case "hung and answered this pass - still mid-cure"       hung    yes      no
clear_case "hung and silent - stale counts must survive"        hung    no       no
clear_case "gone and answered this pass - impossible but safe"  restart yes      no
clear_case "gone and silent"                                    restart no       no

echo
echo "===== rebooting, and refusing to ====="
# SUPERVISE_NO_REBOOT exists for the hardware test and for an operator who wants
# to leave a board in its failed state to investigate it.
# Asserting silence would pass on any breakage that produces nothing at all, so
# this asserts the decision was written down. That log line is the whole point of
# the switch: it exists for the hardware test and for an operator who wants the
# board left in its failed state, and both need to read what it would have done.
got=$(WORK="$WORK" sh -c '
    SUPERVISE_NO_REBOOT=1
    NO_REBOOT=1
    . "$WORK/act.sh"
    log()             { echo "LOG: $*"; }
    warn()            { echo "WARN: $*"; }
    capture_bounded() { echo "captured"; }
    sync()            { echo "synced"; }
    reboot()          { echo "REBOOTED"; }
    sleep()           { :; }
    escalate "test"
' | tr '\n' ' ')
# A reboot decision is a warning, whether or not it is carried out.
want="WARN: would reboot (test), but SUPERVISE_NO_REBOOT is set "
[ "$got" = "$want" ] && note "SUPERVISE_NO_REBOOT=1 records the decision and does nothing else" OK \
                     || note "SUPERVISE_NO_REBOOT=1 did [$got], want [$want]" FAIL

got=$(WORK="$WORK" sh -c '
    NO_REBOOT=0
    . "$WORK/act.sh"
    log()             { echo "LOG: $*"; }
    warn()            { :; }
    capture_bounded() { echo "captured"; }
    sync()            { echo "synced"; }
    reboot()          { echo "REBOOTED"; }
    sleep()           { :; }
    escalate "test"
' | tr '\n' ' ')
[ "$got" = "captured synced REBOOTED " ] \
    && note "evidence is captured and synced before the reboot" OK \
    || note "order was [$got], want [captured synced REBOOTED ]" FAIL

echo
echo "  --- the capture must never be able to block the reboot"
# /tmp does not survive a reboot and dmesg rolls within ten minutes, so evidence
# not taken here is gone. But a capture that wedges means the board never
# reboots, and the guard becomes the fault. Uptime outranks evidence.
start=$(date +%s)
WORK="$WORK" sh -c '
    . "$WORK/act.sh"
    log()              { :; }
    warn()             { :; }
    capture_evidence() { sleep 60; }
    capture_bounded "test"
' > /dev/null 2>&1
elapsed=$(( $(date +%s) - start ))
[ "$elapsed" -lt 20 ] && note "a wedged capture is abandoned after ~10s (took ${elapsed}s)" OK \
                      || note "a wedged capture held the reboot for ${elapsed}s" FAIL

# A reader of /proc/cvitek/vb blocks forever in uninterruptible sleep and cannot
# be killed. Reading it here would mean the board never reboots at all.
if grep -v '^[[:space:]]*#' "$SV" | grep -q '/proc/cvitek/vb'; then
    note "the capture reads /proc/cvitek/vb and would wedge the board" FAIL
else
    note "nothing in the script reads /proc/cvitek/vb" OK
fi

echo
echo "  --- /data is on the SD card, so the evidence is capped"
got=$(WORK="$WORK" sh -c '
    d=$(mktemp -d)
    mkdir -p "$d/kvm-diag"
    for s in 01 02 03 04 05; do mkdir -p "$d/kvm-diag/reboot-2026080$s-000000"; done
    . "$WORK/act.sh"
    cd "$d/kvm-diag" || exit 1
    prune_reboot_dirs
    ls -d "$d"/kvm-diag/reboot-* 2>/dev/null | wc -l
    rm -rf "$d"
')
[ "$got" = "3" ] && note "five reboot directories are pruned to 3" OK \
                 || note "pruning left $got directories, want 3" FAIL

echo
echo "===== the carveout is recorded at each restart ====="
# The carveout erodes with restarts, not with uptime, so the only place this can
# be measured is here, at the moment a restart happens.
ion_case() {   # $1 = name, $2 = fixture dir, $3 = expected line, or "" for none
    : > "$WORK/ion.log"
    (
        # ION_DIR is read by the block through its ${ION_DIR:-...} default, so it
        # is set inside the subshell that sources the block and nowhere else.
        ION_DIR=$2
        log() { echo "$*" >> "$WORK/ion.log"; }
        . "$WORK/ion.sh"
        ion_line
    ) 2> "$WORK/ion.err"
    rc=$?
    got=$(cat "$WORK/ion.log")
    err=$(cat "$WORK/ion.err")
    # Every path through ion_line ends in an explicit `return 0`, and it never
    # writes to stderr. A guard that has been removed does not just skip a line
    # - on the zero-total fixture it lets the shell divide by zero, which exits
    # nonzero and prints to stderr rather than quietly producing empty output.
    # Checking only the log line would call that "caught" for free and prove
    # nothing about the guard.
    if [ "$got" = "$3" ] && [ "$rc" -eq 0 ] && [ -z "$err" ]; then
        note "$1 -> [$got]" OK
    else
        note "$1 -> [$got] rc=$rc err=[$err], want [$3] rc=0" FAIL
    fi
}

mkfixture() {   # $1 = dir, $2 = alloc, $3 = total, $4 = generations
    mkdir -p "$1"
    echo "$2" > "$1/alloc_mem"
    echo "$3" > "$1/total_mem"
    {
        echo "Details:"
        i=0
        while [ "$i" -lt "$4" ]; do
            echo "               0           294912         8bef4000                1 ISP_SHARED_BUFFER_0"
            i=$(( i + 1 ))
        done
        echo "minimum ion allocate unit = 4096"
    } > "$1/summary"
}

mkfixture "$WORK/ion-clean"  19050496 78643200 1
mkfixture "$WORK/ion-orphan" 49459200 78643200 2
mkfixture "$WORK/ion-zero"   19050496 0        1

ion_case "a healthy board reports one generation" \
    "$WORK/ion-clean"  "ion 19050496/78643200 24% gen=1"
ion_case "an orphaned generation is counted" \
    "$WORK/ion-orphan" "ion 49459200/78643200 62% gen=2"
# ion-absent is deliberately never created by mkfixture. A board without the
# debugfs entry must write no line at all.
ion_case "a missing carveout writes nothing" \
    "$WORK/ion-absent" ""
ion_case "a zero total writes nothing rather than dividing by it" \
    "$WORK/ion-zero"   ""

# A summary that cannot be read must not lose the counters.
rm -f "$WORK/ion-clean/summary"
ion_case "a missing summary still reports the counters" \
    "$WORK/ion-clean"  "ion 19050496/78643200 24% gen=0"

# Not in the brief's fixture set: every mkfixture value above is a plain digit
# string, so a case guard that stopped rejecting non-numeric input would leave
# every case above unchanged and pass anyway - the debugfs file is text ("carveout
# heap size:..." on a kernel where the split integer files do not exist), so a
# malformed total is a real state, not a hypothetical one.
mkdir -p "$WORK/ion-total-garbage"
echo 19050496 > "$WORK/ion-total-garbage/alloc_mem"
echo "carveout heap size:78643200 bytes" > "$WORK/ion-total-garbage/total_mem"
ion_case "a non-numeric total writes nothing rather than being accepted" \
    "$WORK/ion-total-garbage" ""

# Anchored to the call site inside full_restart, not to the function name: a
# grep for "ion_line" alone would also match its own definition, so a function
# that is defined and never called would pass every case above.
got=$(sed -n '/^full_restart()/,/^}/p' "$SV" | grep -c '^[[:space:]]*ion_line$')
[ "$got" = "1" ] && note "full_restart actually calls ion_line" OK \
                 || note "full_restart calls ion_line $got times, want 1" FAIL

# full_restart is the hang cure, not the common case. A server that simply
# died is relaunched inline in watch_loop, and that path erodes the carveout
# exactly the same way - a dead process keeps its whole ION working set
# whichever path restarts it. Hardware acceptance found this path recording
# nothing: killing the server and letting it come back through this branch
# logged no ion line at all. Anchored to the "if action = restart" guard
# around the direct launch, not to the function name, so this cannot be
# satisfied by full_restart's own call or by the definition.
got=$(sed -n '/^[[:space:]]*if \[ "\$ACTION" = restart \]; then$/,/^[[:space:]]*fi$/p' "$SV" | grep -c '^[[:space:]]*ion_line$')
[ "$got" = "1" ] && note "the inline restart branch actually calls ion_line" OK \
                 || note "the inline restart branch calls ion_line $got times, want 1" FAIL

echo
echo "===== the script still parses ====="
sh -n "$SV" && note "S98supervise is valid shell" OK || note "S98supervise does not parse" FAIL
# This check used to grep for the redirection string, and passed while the
# behaviour was wrong: measured over ssh, `start` printed its line and then held
# the session open until the client gave up after five minutes. Redirecting the
# loop's stdio to /dev/null is not enough on busybox - it needs its own session.
#
# Whether it really detaches can only be shown on a device, by timing how long
# `start` takes to return. This asserts the mechanism is present; the timing is
# the evidence, and it belongs in the deploy notes rather than here.
grep -q 'setsid "\$0" __watch' "$SV" \
    && note "detaches with setsid, not merely redirected fds" OK \
    || note "no setsid - start would hold the calling ssh session" FAIL
grep -q '__watch)' "$SV" \
    && note "the entry point setsid re-enters is handled" OK \
    || note "setsid re-enters a subcommand the script does not handle" FAIL

# cure_hung was once defined and never called while every case above stayed
# green: testable in isolation, unreachable from the loop. The same trap applies
# to everything this file extracts, so each new decision gets a wiring check too.
# Like the setsid check these assert a string, so the real evidence is the
# on-device test that crash-loops the server and watches the board come back.
grep -qE '^[[:space:]]+if cure_hung; then$' "$SV" \
    && note "the hang branch actually calls cure_hung" OK \
    || note "cure_hung is defined but never reached" FAIL
grep -qE 'should_reboot restart' "$SV" \
    && note "the restart branch actually asks should_reboot" OK \
    || note "should_reboot is defined but no crash loop reaches it" FAIL
grep -qE 'should_reboot hung' "$SV" \
    && note "the hang branch actually asks should_reboot" OK \
    || note "should_reboot is defined but no hang reaches it" FAIL
# One pattern matching both call sites meant deleting either alone left it
# green - the other call kept the string in the file. Anchored one per site,
# the way every other wiring check here is one-to-one with the call it proves.
grep -qE '^[[:space:]]+escalate "\$hang_reason"$' "$SV" \
    && note "the hang branch actually calls escalate" OK \
    || note "escalate is defined but the hang branch never reaches it" FAIL

# The reason string is what lands in the evidence directory, and it is the only
# thing that says which hang this was. A process that will not die after SIGKILL
# is not two cures that did not work: no cure was attempted, and none would help.
# Collapsing the two strings parses and changes no decision, so only a check that
# both spellings exist can see it.
counted=$(grep -c 'hang_reason="hung: \$failed_cures cures did not restore service"' "$SV")
unkillable=$(grep -c 'hang_reason="hung: the process did not leave after SIGKILL"' "$SV")
if [ "$counted" -eq 1 ] && [ "$unkillable" -eq 1 ]; then
    note "the two hang reasons say different things" OK
else
    note "hang reasons: $counted counted, $unkillable unkillable, want one of each" FAIL
fi
grep -qE '^[[:space:]]+escalate "crash loop: ' "$SV" \
    && note "the restart branch actually calls escalate" OK \
    || note "escalate is defined but the restart branch never reaches it" FAIL
grep -qE '^[[:space:]]+short_runs=\$\(next_short_runs ' "$SV" \
    && note "the restart branch actually updates short_runs" OK \
    || note "next_short_runs is defined and never called" FAIL
grep -qE '^[[:space:]]+failed_cures=\$\(next_failed_cures ' "$SV" \
    && note "the hang branch actually updates failed_cures" OK \
    || note "next_failed_cures is defined and never called" FAIL
grep -qE '^[[:space:]]+if should_clear ' "$SV" \
    && note "the loop actually asks should_clear before wiping the counters" OK \
    || note "should_clear is defined and never called" FAIL

# serving fails open on purpose: a probe that cannot run must never kill a
# working KVM. That answer cannot also be the answer the latch reads, or "the
# probe could not run" is recorded as "the server answered" and a board without
# curl gets the reboot cycle the latch exists to prevent. Three answers, and
# only 0 means the server answered.
grep -qE '^[[:space:]]+\[ -x "\$CURL_BIN" \] \|\| return 2$' "$SV" \
    && note "a probe that cannot run says so, rather than saying success" OK \
    || note "a missing curl is indistinguishable from an answering server" FAIL

# The latch has to reach both decisions, or half the reboot cycle comes back.
grep -qE '^[[:space:]]+served_ever=yes$' "$SV" \
    && note "the loop sets the latch when the probe answers" OK \
    || note "nothing ever sets served_ever, so no board could reboot" FAIL

# watch_loop is a while loop with side effects, so this suite cannot drive it
# and the latch's own assignment has no unit case. What can be asserted is its
# shape: the one assignment in the file sits directly inside the branch that
# tests for status 0, so no other status can reach it. Deleting the branch or
# widening it to -ne 1 makes this report FAIL.
latch_line=$(sed -n '/^[[:space:]]*if \[ "\$answered" -eq 0 \]; then$/{n;s/^[[:space:]]*//;p;}' "$SV")
latch_count=$(grep -c '^[[:space:]]*served_ever=yes$' "$SV")
if [ "$latch_line" = "served_ever=yes" ] && [ "$latch_count" -eq 1 ]; then
    note "the latch is set only where the probe answered" OK
else
    note "the latch is set outside the answered branch (guarded [$latch_line], $latch_count assignments)" FAIL
fi
grep -qE 'should_reboot restart .*"\$served_ever"' "$SV" \
    && note "the restart branch passes the latch" OK \
    || note "the restart branch judges without the latch" FAIL
grep -qE 'should_reboot hung .*"\$served_ever"' "$SV" \
    && note "the hang branch passes the latch" OK \
    || note "the hang branch judges without the latch" FAIL

# served_ever is a per-boot latch, not a counter. should_clear must never reach
# it: a fault present from boot would then escalate as soon as the floor let it,
# the counters would not survive the reboot, and the identical sequence would
# repeat every REBOOT_FLOOR seconds for as long as the board had power.
if [ "$(grep -c '^[[:space:]]*served_ever=no$' "$SV")" -eq 1 ]; then
    note "the latch is set once at boot and never cleared" OK
else
    note "served_ever is assigned no in more than one place, so it is a counter" FAIL
fi

echo
echo "===== the other door ====="
# sshd is how this board gets repaired when the server is the thing that broke,
# and nothing watched it before. It is judged on its own: the cases below are
# about the door alone, and the two shape assertions at the end are what keep it
# from ever reaching the reboot ladder.

# ssh_case <description> <probe answer> <seconds down> <seconds since cure> <want>
ssh_case() {
    desc="$1"; answered="$2"; down="$3"; since="$4"; want="$5"
    # The owner's off switch points at a path that does not exist, so a run
    # on a board whose owner turned SSH off still tests the table.
    got=$(WORK="$WORK" SSH_STOP_FLAG="$WORK/no-ssh-stop" sh -c '
        . "$WORK/sshdoor.sh"
        ssh_action "$1" "$2" "$3" && echo restart || echo none
    ' sh "$answered" "$down" "$since")
    if [ "$got" = "$want" ]; then
        note "$desc -> $got" OK
    else
        note "$desc -> $got, want $want" FAIL
    fi
}

#         description                              answer down  since  want
ssh_case "listening, so nothing to do"             0     0      999    none
ssh_case "listening after a long outage"           0     9999   999    none
ssh_case "the probe could not run"                 2     9999   999    none
ssh_case "down, but not for long enough"           1     30     999    none
ssh_case "down for exactly the threshold"          1     60     999    restart
ssh_case "down for longer"                         1     600    999    restart
ssh_case "down, but a cure is still backing off"   1     600    30     none
ssh_case "down, and the backoff has expired"       1     600    300    restart

# The same fail-closed rule should_reboot documents. An error inside `[` is not
# a false comparison: it skips the guard entirely.
ssh_case "an empty seconds-down"                   1     ""     999    none
ssh_case "a non-numeric seconds-down"              1     abc    999    none
ssh_case "a negative seconds-down"                 1     -5     999    none
ssh_case "an empty backoff clock"                  1     600    ""     none
ssh_case "a seconds-down too wide to compare"      1     99999999999 999 none

# SSH that the owner turned off is not a fault. The web UI's switch writes
# /etc/kvm/ssh_stop, and the supervisor used to "cure" the missing listener
# every five minutes for as long as the switch stayed off, writing a line to
# the SD card each time. S50sshd honoured the flag, so nothing started, but the
# door kept treating a setting as a failure.
touch "$WORK/ssh-stop"
got=$(WORK="$WORK" SSH_STOP_FLAG="$WORK/ssh-stop" sh -c '
    . "$WORK/sshdoor.sh"
    ssh_action 1 600 999 && echo restart || echo none
')
[ "$got" = none ] && note "down, and the owner turned SSH off -> $got" OK \
    || note "down, and the owner turned SSH off -> $got, want none" FAIL

echo
echo "===== the door is a listener, not a process ====="
# `pidof sshd` passes on a board nobody can reach: a wedged sshd keeps its pid
# and its port. This tests the shipped probe against socket tables laid out the
# way the kernel writes /proc/net/tcp and tcp6: local address, remote address,
# state, with the port in hex and 0A meaning LISTEN.
HDR='  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode'
TAIL='00000000:00000000 00:00000000 00000000     0        0 1234 1 0000000000000000 100 0 0 10 0'
V6ANY=00000000000000000000000000000000

# table <proc dir> <tcp|tcp6> <local port hex> <state hex> ...
table() {
    dir=$1; name=$2; shift 2
    mkdir -p "$dir/net"
    echo "$HDR" > "$dir/net/$name"
    n=0
    while [ "$#" -ge 2 ]; do
        if [ "$name" = tcp6 ]; then addr=$V6ANY; else addr=00000000; fi
        echo "   $n: $addr:$1 $addr:0000 $2 $TAIL" >> "$dir/net/$name"
        n=$(( n + 1 )); shift 2
    done
}

# 0x0016 is 22, 0x0050 is 80, 0x08AE is 2222. 01 is ESTABLISHED.
table "$WORK/p-open"      tcp  0050 0A 0016 0A
table "$WORK/p-open"      tcp6
table "$WORK/p-open6"     tcp  0050 0A
table "$WORK/p-open6"     tcp6 0016 0A
table "$WORK/p-shut"      tcp  0050 0A
table "$WORK/p-shut"      tcp6 0050 0A
table "$WORK/p-connected" tcp  0050 0A 0016 01
table "$WORK/p-2222"      tcp  0050 0A 08AE 0A
table "$WORK/p-only4"     tcp  0016 0A
mkdir -p "$WORK/p-none"

# port_files <ssh_port contents, or absent> <drop-in: yes or no>
# S50sshd writes the drop-in for any port it accepted, so it is there unless a
# case says otherwise.
port_files() {
    SSH_PORT_FILE="$WORK/ssh_port"
    SSH_PORT_DROPIN="$WORK/ironkvm-port.conf"
    rm -f "$SSH_PORT_FILE" "$SSH_PORT_DROPIN"
    [ "$1" = absent ] || printf '%s\n' "$1" > "$SSH_PORT_FILE"
    [ "${2:-yes}" = no ] || printf 'Port %s\n' "$1" > "$SSH_PORT_DROPIN"
}

# probe <proc dir> <ssh_port contents, or absent> [drop-in: yes or no]
probe() {
    (
        PROC="$WORK/$1"
        port_files "$2" "$3"
        . "$WORK/sshdoor.sh"
        ssh_listening
        echo $?
    )
}

probe_case() {   # description, proc dir, port file, want, [drop-in]
    got=$(probe "$2" "$3" "$5")
    [ "$got" = "$4" ] && note "$1 -> $got" OK || note "$1 -> $got, want $4" FAIL
}

#          description                                     proc        port    want
probe_case "a listener on 22 reads as open"                p-open      absent  0
probe_case "a listener on 22 in tcp6 alone reads as open"  p-open6     absent  0
probe_case "a listener in tcp alone, tcp6 missing, is open" p-only4    absent  0
probe_case "no listener on 22 reads as shut"               p-shut      absent  1
# A connection on 22 is not a listener: a board whose sshd died keeps the
# session that was open when it did until that session ends.
probe_case "an established connection on 22 is not the door" p-connected absent 1
# 2222 is a listener that contains the digits and is not the door.
probe_case "a listener on 2222 is not mistaken for 22"     p-2222      absent  1

echo
echo "  --- the owner's port"
# Settings > SSH can move sshd off 22. A probe fixed at 22 reads that board's
# working door as shut and restarts its sshd every SSH_CURE_BACKOFF, for ever.
probe_case "port 2222 chosen, sshd on 2222"                p-2222      2222    0
probe_case "port 2222 chosen, nothing on 2222 or 22"       p-shut      2222    1
# Only the chosen port is the door. Something else listening on 22 must not
# hide an sshd that died on 2222.
probe_case "port 2222 chosen, only 22 listens"             p-only4     2222    1
probe_case "port 2222 chosen, 22 listens and 2222 does not" p-open     2222    1
# With 22 chosen, a listener on 2222 is not the door either.
probe_case "port 22 by default, only 2222 listens"         p-2222      absent  1
# S50sshd falls back to 22 when sshd rejects the port, and removes the drop-in
# when it does. That door is open on 22, so it must read as open.
probe_case "port 2222 refused, sshd fell back to 22"       p-only4     2222    0  no
probe_case "port 2222 refused, nothing on 22"              p-2222      2222    1  no

# ssh_port applies S50sshd's rules: anything it would not use means 22.
port_case() {   # description, file contents or absent, want, [drop-in]
    got=$(
        port_files "$2" "$4"
        . "$WORK/sshdoor.sh"
        ssh_port
        echo "$SSH_PORT"
    )
    [ "$got" = "$3" ] && note "$1 -> $got" OK || note "$1 -> $got, want $3" FAIL
}

port_case "no port file"                        absent                22
port_case "an empty port file"                  ""                    22
port_case "a chosen port"                       2222                  2222
port_case "the highest port"                    65535                 65535
port_case "one past the highest port"           65536                 22
port_case "a number too wide to compare"        99999999999999999999  22
port_case "a leading zero, which sh reads as octal" 022               22
port_case "zero"                                0                     22
port_case "not a number"                        ssh                   22
port_case "a negative number"                   -22                   22
port_case "a chosen port sshd refused"          2222                  22    no

# A board whose socket tables cannot be read measured nothing, and must say so
# rather than report the door shut.
probe_case "no socket tables: the probe could not run"     p-none      absent  2

# The log line names the port that was probed, or an owner on 2222 reads that
# 22 was down and goes looking for the wrong fault.
grep -q 'warn "nothing has listened on port \$SSH_PORT for ' "$SV" \
    && note "the restart line names the probed port" OK \
    || note "the restart line does not name the probed port" FAIL

echo
echo "===== kvm_system is tracked by pid, not found by pidof every pass ====="
# pidof forks and walks all of /proc, every pass, for a process that is almost
# always there. The loop remembers the pid and reads /proc/<pid>/comm; the name
# check is what keeps a reused pid from passing for kvm_system.
mkdir -p "$WORK/pp/123" "$WORK/pp/124" "$WORK/pp/456"
echo kvm_system > "$WORK/pp/123/comm"
echo sh         > "$WORK/pp/124/comm"
echo kvm_system > "$WORK/pp/456/comm"

# sys_case <description> <SYS_PID before> <what pidof prints> <want: rc pid pidof-calls>
sys_case() {
    got=$(SP="$2" OUT="$3" WORK="$WORK" sh -c '
        PROC=$WORK/pp
        : > "$WORK/sys.calls"
        pidof() { echo x >> "$WORK/sys.calls"; [ -n "$OUT" ] && echo "$OUT"; }
        . "$WORK/procs.sh"
        SYS_PID=$SP
        system_running; rc=$?
        echo "$rc ${SYS_PID:--} $(wc -l < "$WORK/sys.calls" | tr -d " ")"
    ')
    [ "$got" = "$4" ] && note "$1 -> $got" OK || note "$1 -> [$got], want [$4]" FAIL
}

#        description                                 before  pidof      want
sys_case "first pass: pidof finds it and it is kept" ""      "123"      "0 123 1"
sys_case "tracked and alive: pidof is not asked"     123     ""         "0 123 0"
sys_case "the pid was reused by something else"      124     ""         "1 - 1"
sys_case "reused, and pidof finds the real one"      124     "456"      "0 456 1"
sys_case "gone, and pidof reports several"           999     "456 789"  "0 456 1"
sys_case "never seen and not running"                ""      ""         "1 - 1"

# The pid the loop starts is the pid it tracks, or the next pass asks pidof for
# a process it started itself. Both ways of starting it, with syslog and
# without, have to keep the pid.
tracked=0
for start in '"\$SYSTEM_BIN" < /dev/null > /dev/null 2>&1 &' '"\$SYSTEM_BIN" < /dev/null >&8 2>&8 8>&- &'
do
    grep -A1 "^[[:space:]]*$start\$" "$SV" | grep -q '^[[:space:]]*SYS_PID=\$!$' \
        && tracked=$(( tracked + 1 ))
done
[ "$tracked" -eq 2 ] \
    && note "a kvm_system this starts is tracked from the start" OK \
    || note "a kvm_system this starts is not tracked" FAIL
loop_start=$(sed -n '/^watch_loop() {$/,/^}$/p' "$SV" | grep -A1 '^[[:space:]]*start_system$')
case "$loop_start" in
    *'warn "kvm_system was gone, started it as pid $SYS_PID"'*)
        note "the loop starts kvm_system through start_system" OK ;;
    *)  note "the loop does not start kvm_system through start_system" FAIL ;;
esac

echo
echo "===== kvm_system's output goes to syslog ====="
# kvm_system's output went to /dev/null, and from S95nanokvm to wherever that
# script's output went. It now goes through one logger per kvm_system, fed by a
# fifo so that $! stays kvm_system: a pipeline would hand the loop the logger's
# pid, and the loop would track a process that is not kvm_system.
sed -n '/^# --- syslog pipe ---$/,/^# --- end syslog pipe ---$/p' "$SV" > "$WORK/syslog.sh"
[ -s "$WORK/syslog.sh" ] || { echo "could not extract the syslog pipe block"; exit 1; }

# Three scripts start a program this way. One block, spelled three times,
# drifts, so the copies must match.
for other in "$S95" "$(dirname "$S95")/S98tailscaled"
do
    sed -n '/^# --- syslog pipe ---$/,/^# --- end syslog pipe ---$/p' "$other" > "$WORK/syslog.other"
    if cmp -s "$WORK/syslog.sh" "$WORK/syslog.other"; then
        note "$(basename "$other") carries the same syslog pipe" OK
    else
        note "$(basename "$other") carries a different syslog pipe" FAIL
    fi
done

grep -q '^        /tmp/kvm_system/kvm_system >&8 2>&8 8>&- &$' "$S95" \
    && note "S95nanokvm sends kvm_system to syslog too" OK \
    || note "S95nanokvm does not send kvm_system to syslog" FAIL

# The rest runs the block for real: a fifo, a stub logger and a stub
# kvm_system. It needs /proc/<pid>/comm, which Linux has and Git Bash does not.
if [ -r /proc/self/comm ] && command -v mkfifo > /dev/null 2>&1; then
    SL="$WORK/sl"
    mkdir -p "$SL/bin"
    # Named kvm_system, so the kernel names the process that and comm_is can
    # find it. It says one line on each stream, then one line a tick while the
    # talk file exists.
    cat > "$SL/bin/kvm_system" <<'STUB'
#!/bin/sh
echo "to stdout"
echo "to stderr" >&2
trap 'exit 0' TERM
while :; do
    [ -e "$TALK" ] && echo tick
    sleep 0.1
done
STUB
    # Records how it was called, each line it read, and its end of file.
    cat > "$SL/bin/logger" <<'STUB'
#!/bin/sh
echo "args $*" >> "$LOGGER_OUT"
while IFS= read -r line; do echo "line $line" >> "$LOGGER_OUT"; done
echo eof >> "$LOGGER_OUT"
STUB
    chmod +x "$SL/bin/kvm_system" "$SL/bin/logger"

    # wait_for <what> <shell test>, up to three seconds.
    wait_for() {
        i=0
        while ! eval "$2"; do
            i=$(( i + 1 ))
            [ "$i" -ge 30 ] && return 1
            sleep 0.1
        done
        return 0
    }
    # The bracket keeps grep from finding its own command line.
    logger_pids() {
        grep -l "$SL/bin/logge[r]" /proc/[0-9]*/cmdline 2>/dev/null \
            | sed 's|^/proc/\([0-9]*\)/cmdline$|\1|'
    }

    # One run: start it, check what reached the logger, stop it.
    : > "$SL/out"
    rm -f "$SL/talk"
    SYSLOG_LOGGER="$SL/bin/logger" SYSLOG_FIFO_DIR="$SL" LOGGER_OUT="$SL/out" TALK="$SL/talk" \
        SYSTEM_BIN="$SL/bin/kvm_system" WORK="$WORK" SL="$SL" sh -c '
        . "$WORK/syslog.sh"
        . "$WORK/procs.sh"
        start_system
        echo "$SYS_PID" > "$SL/pid"
        { [ -e "/proc/$$/fd/8" ] || [ -e "/proc/$$/fd/9" ]; } && echo open > "$SL/fds"
        # The child is named sh until its exec lands.
        i=0
        until comm_is "$SYS_PID" kvm_system || [ "$i" -ge 30 ]; do
            i=$(( i + 1 )); sleep 0.1
        done
        comm_is "$SYS_PID" kvm_system && echo tracked > "$SL/tracked"
        exit 0
    '
    pid=$(cat "$SL/pid" 2>/dev/null)

    [ -e "$SL/tracked" ] && note "SYS_PID is kvm_system, not its logger" OK \
                         || note "SYS_PID $pid is not kvm_system" FAIL
    [ -e "$SL/fds" ] && note "the starting shell keeps the pipe open" FAIL \
                     || note "the starting shell closes its ends of the pipe" OK
    leftovers=$(ls -A "$SL" | grep -c '^\.syslog-')
    [ "$leftovers" -eq 0 ] && note "no fifo is left behind" OK \
                           || note "$leftovers fifo(s) left behind" FAIL

    if wait_for "both lines" 'grep -q "^line to stderr$" "$SL/out" && grep -q "^line to stdout$" "$SL/out"'; then
        note "stdout and stderr both reach the logger" OK
    else
        note "the logger got: $(tr '\n' '|' < "$SL/out")" FAIL
    fi
    grep -q '^args -t kvm_system -p daemon.info$' "$SL/out" \
        && note "the logger tags it kvm_system at daemon.info" OK \
        || note "the logger was called: $(grep '^args' "$SL/out")" FAIL

    # A stop or a killall sends TERM. The logger stays for as long as
    # kvm_system has something to say.
    lp=$(logger_pids)
    [ -n "$lp" ] && kill -TERM $lp 2>/dev/null
    sleep 0.3
    [ -n "$lp" ] && [ -n "$(logger_pids)" ] \
        && note "the logger ignores SIGTERM" OK \
        || note "the logger died of SIGTERM" FAIL

    # Stopping kvm_system ends its logger, so a restart leaves none behind.
    kill "$pid" 2>/dev/null
    if wait_for "the logger to go" '[ -z "$(logger_pids)" ]'; then
        note "stopping kvm_system ends its logger" OK
    else
        note "the logger outlived kvm_system: $(logger_pids)" FAIL
        kill -9 $(logger_pids) 2>/dev/null
    fi
    grep -q '^eof$' "$SL/out" && note "the logger read to the end" OK \
                              || note "the logger did not see end of file" FAIL

    # A logger that is gone must not wedge kvm_system on a full pipe. It
    # dies of SIGPIPE at its next write, and the loop starts the pair again.
    : > "$SL/out"
    SYSLOG_LOGGER="$SL/bin/logger" SYSLOG_FIFO_DIR="$SL" LOGGER_OUT="$SL/out" TALK="$SL/talk" \
        SYSTEM_BIN="$SL/bin/kvm_system" WORK="$WORK" SL="$SL" sh -c '
        . "$WORK/syslog.sh"
        . "$WORK/procs.sh"
        start_system
        echo "$SYS_PID" > "$SL/pid"
    '
    pid=$(cat "$SL/pid" 2>/dev/null)
    wait_for "the logger" '[ -n "$(logger_pids)" ]'
    kill -9 $(logger_pids) 2>/dev/null
    : > "$SL/talk"
    if wait_for "kvm_system to go" '! kill -0 "$pid" 2>/dev/null'; then
        note "a dead logger ends kvm_system instead of blocking it" OK
    else
        note "kvm_system $pid outlived its logger" FAIL
        kill -9 "$pid" 2>/dev/null
    fi
    rm -f "$SL/talk"

    # Without logger nothing changes: /dev/null, tracked, no logger started.
    SYSLOG_LOGGER= SYSLOG_FIFO_DIR="$SL" TALK="$SL/talk" \
        SYSTEM_BIN="$SL/bin/kvm_system" WORK="$WORK" SL="$SL" sh -c '
        . "$WORK/syslog.sh"
        . "$WORK/procs.sh"
        start_system
        echo "$SYS_PID" > "$SL/pid"
        i=0
        until comm_is "$SYS_PID" kvm_system || [ "$i" -ge 30 ]; do
            i=$(( i + 1 )); sleep 0.1
        done
        comm_is "$SYS_PID" kvm_system && echo tracked > "$SL/tracked2"
    '
    pid=$(cat "$SL/pid" 2>/dev/null)
    [ -e "$SL/tracked2" ] && [ -z "$(logger_pids)" ] \
        && note "without logger kvm_system starts as before" OK \
        || note "without logger: tracked=$([ -e "$SL/tracked2" ] && echo yes || echo no) loggers=$(logger_pids)" FAIL
    kill "$pid" 2>/dev/null
else
    note "the syslog pipe itself (needs Linux /proc and mkfifo)" SKIP
fi

echo
echo "===== a healthy pass costs a sleep and a curl ====="
# The loop runs every INTERVAL for the life of the board, on one core. Each of
# these used to fork on every pass: measured on the board at about 3.5% CPU and
# 3.6 forks a second at idle. What is asserted is that none has come back into
# watch_loop; comments are stripped so the reasons can still be written down.
loop_code=$(sed -n '/^watch_loop() {$/,/^}$/p' "$SV" | grep -v '^[[:space:]]*#')
[ -n "$loop_code" ] || note "could not find watch_loop" FAIL
for pat in '$(now)' '$(action)' '$(should_clear' '$(ssh_action' 'netstat' 'command -v' 'pidof' '/proc/uptime'
do
    case "$loop_code" in
        *"$pat"*) note "watch_loop no longer uses $pat" FAIL ;;
        *)        note "watch_loop does not use $pat" OK ;;
    esac
done

echo
echo "===== the ssh door never reaches the reboot ladder ====="
# This board serves video and HID through the web door. Restarting a working KVM
# because a maintenance daemon will not start is the worst trade in the file, so
# the ssh cure must stay out of should_reboot and out of escalate.
if grep -n 'should_reboot' "$SV" | grep -qi 'ssh'; then
    note "the ssh door feeds the reboot decision" FAIL
else
    note "the ssh door never feeds the reboot decision" OK
fi

if grep -n 'escalate ' "$SV" | grep -qi 'ssh'; then
    note "the ssh door can escalate to a reboot" FAIL
else
    note "the ssh door cannot escalate to a reboot" OK
fi

# An update replaces the boot scripts. Restarting one in the middle of that is
# the fault the stand-off exists to prevent, so the cure sits inside the guard.
if sed -n '/if \[ "\$state" != updating \]; then/,/^        fi$/p' "$SV" | grep -q 'cure_ssh'; then
    note "the ssh cure stands off during an update" OK
else
    note "the ssh cure runs during an update" FAIL
fi

echo
echo "===== each event line goes to the file and to syslog ====="
# The file on /data survives a reboot; syslog is what reaches the owner's
# collector. A line must land in both, and a board without logger must still
# get the file and a zero status, since callers chain on log.
sed -n '/^# --- log ---$/,/^# --- end log ---$/p' "$SV" > "$WORK/log.sh"
[ -s "$WORK/log.sh" ] || { echo "could not extract the log block"; exit 1; }

cat > "$WORK/logger" <<'STUB'
#!/bin/sh
printf '%s|' "$@" >> "$LOGGER_OUT"
echo >> "$LOGGER_OUT"
STUB
chmod +x "$WORK/logger"

rm -f "$WORK/events.log" "$WORK/logger.out"
LOGGER_OUT="$WORK/logger.out" SUPERVISE_LOGGER="$WORK/logger" LOG="$WORK/events.log" sh -c '
    . "$0"
    log "-n rebooting: crash loop"
' "$WORK/log.sh"
if grep -q ' -n rebooting: crash loop$' "$WORK/events.log" 2>/dev/null; then
    note "the line is in the file" OK
else
    note "the line is not in the file" FAIL
fi
if [ "$(cat "$WORK/logger.out" 2>/dev/null)" = "-t|supervise|-p|daemon.notice|--|-n rebooting: crash loop|" ]; then
    note "the line goes to syslog tagged supervise" OK
else
    note "syslog got: $(cat "$WORK/logger.out" 2>/dev/null)" FAIL
fi

# warn is the same line at daemon.warning, so a collector can alert on what
# this script does to the board without the routine lines.
rm -f "$WORK/events.log" "$WORK/logger.out"
LOGGER_OUT="$WORK/logger.out" SUPERVISE_LOGGER="$WORK/logger" LOG="$WORK/events.log" sh -c '
    . "$0"
    warn "rebooting: crash loop"
' "$WORK/log.sh"
if grep -q ' rebooting: crash loop$' "$WORK/events.log" 2>/dev/null; then
    note "a warning is in the file" OK
else
    note "a warning is not in the file" FAIL
fi
if [ "$(cat "$WORK/logger.out" 2>/dev/null)" = "-t|supervise|-p|daemon.warning|--|rebooting: crash loop|" ]; then
    note "a warning goes to syslog at daemon.warning" OK
else
    note "syslog got: $(cat "$WORK/logger.out" 2>/dev/null)" FAIL
fi

# Which lines are warnings. Each of these reboots the board, decides to, or
# kills or restarts a process; everything else stays at notice.
for line in \
    'warn "the evidence capture did not finish' \
    'warn "would reboot (' \
    'warn "rebooting: ' \
    'warn "NanoKVM-Server is up but has not answered' \
    'warn "the process did not leave after SIGKILL"' \
    'warn "NanoKVM-Server is gone after' \
    'warn "started $SERVER_BIN as pid' \
    'warn "kvm_system was gone' \
    'warn "nothing has listened on port' \
    'log "ion ' \
    'log "an update is in progress' \
    'log "no update in progress any more' \
    'log "supervisor started' \
    'log "supervisor stopped"'
do
    if grep -qF "$line" "$SV"; then
        note "${line%% *} ${line#* }" OK
    else
        note "missing: $line" FAIL
    fi
done
n_warn=$(grep -c '^[[:space:]]*warn "' "$SV")
[ "$n_warn" -eq 9 ] && note "nine lines are warnings" OK \
                    || note "$n_warn lines are warnings, want 9" FAIL

rm -f "$WORK/events.log"
SUPERVISE_LOGGER= LOG="$WORK/events.log" sh -c '
    . "$0"
    log "no logger here"
' "$WORK/log.sh"
status=$?
if [ "$status" -eq 0 ] && grep -q ' no logger here$' "$WORK/events.log" 2>/dev/null; then
    note "without logger the line still reaches the file" OK
else
    note "without logger: status $status, file $(cat "$WORK/events.log" 2>/dev/null)" FAIL
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "===== all supervisor cases pass ====="
else
    echo "===== $fails case(s) failed ====="
    exit 1
fi
