#!/bin/sh
# Check the exit record that service/failsafe decides from.
#
#   test-exit-record.sh [path-to-S95nanokvm] [path-to-S98supervise]
#
# S95nanokvm and S98supervise both start the server through run_recorded,
# which waits for it and writes one file per exit: the status, the Go crash
# report and the log's last lines. A clean exit removes the records instead.
# The block is taken out of the scripts that ship and run against stand-in
# servers in a scratch directory. Nothing on the board is touched.
#
# Exits 0 when every case passes, 1 when one fails, 2 when it cannot run here
# (it needs /proc/uptime).
S95=${1:-$(dirname "$0")/../../kvmapp/system/init.d/S95nanokvm}
SV=${2:-$(dirname "$0")/S98supervise}
[ -f "$S95" ] && [ -f "$SV" ] || { echo "usage: test-exit-record.sh <S95nanokvm> <S98supervise>"; exit 1; }
[ -r /proc/uptime ] || { echo "cannot run here: no /proc/uptime"; exit 2; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

extract() { sed -n '/^# --- exit record ---$/,/^# --- end exit record ---$/p' "$1"; }

echo "===== one block, in both scripts ====="
extract "$S95" > "$WORK/block.sh"
extract "$SV" > "$WORK/block.sv"
[ -s "$WORK/block.sh" ] || { echo "could not extract the exit record block"; exit 1; }
if cmp -s "$WORK/block.sh" "$WORK/block.sv"; then
    note "S98supervise carries the same exit record" OK
else
    note "S98supervise carries a different exit record" FAIL
fi

grep -q '^        run_recorded "\$SERVER_DST/NanoKVM-Server" >> "\$SERVER_LOG" 2>&1 &$' "$S95" \
    && note "S95nanokvm starts the server through run_recorded" OK \
    || note "S95nanokvm starts the server without a record" FAIL
grep -q 'run_recorded "\$SERVER_BIN" < /dev/null >> "\$SERVER_LOG" 2>&1 &$' "$SV" \
    && note "S98supervise restarts the server through run_recorded" OK \
    || note "S98supervise restarts the server without a record" FAIL

# No other way to start the server remains in either script.
n=$(grep -c 'NanoKVM-Server" >>\|SERVER_BIN" <' "$S95" "$SV" | awk -F: '{ s += $2 } END { print s }')
[ "$n" -eq 2 ] && note "both start lines are the recorded ones" OK \
               || note "$n start lines, want 2" FAIL

echo
echo "===== the records ====="
DIR=$WORK/failsafe
LOGF=$WORK/server.log
# The server makes the directory at its start; the stand-in does not.
mkdir -p "$DIR"

# A stand-in server. It prints to the log as the real one does, may leave a
# crash report under its own pid, and ends as it is told.
cat > "$WORK/server" <<'EOF'
#!/bin/sh
echo "--- NanoKVM-Server started, pid $$"
echo "line from run $RUN"
if [ -n "$REPORT" ]; then
    printf '%s\n' "$REPORT" > "$FAILSAFE_DIR/crash.$$"
fi
case "$END" in
    segv) kill -SEGV $$ ;;
    kill) kill -KILL $$ ;;
    *) exit "$END" ;;
esac
EOF
chmod +x "$WORK/server"

run() {   # run in the background, as the scripts do, and wait for the recorder
    (
        FAILSAFE_DIR=$DIR
        SERVER_LOG=$LOGF
        export FAILSAFE_DIR RUN END REPORT
        . "$WORK/block.sh"
        run_recorded "$WORK/server" >> "$LOGF" 2>&1 &
        wait $!
    )
}

records() { ls "$DIR/exits" 2>/dev/null | grep -cv '\.new$'; }
of_run() { grep -l "line from run $1$" "$DIR"/exits/* | head -1; }

RUN=1; END=2; REPORT='fatal error: unexpected signal during runtime execution
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x1]
NanoKVM-Server/common._Cfunc_kvmv_init(0x0)'; run
[ "$(records)" -eq 1 ] && note "a crash leaves one record" OK || note "a crash left $(records) records" FAIL
rec=$(of_run 1)
grep -qx 'status 2' "$rec" && note "the record holds the status" OK || note "no status 2 in the record" FAIL
grep -q '^uptime [0-9][0-9]*$' "$rec" && note "the record holds the uptime" OK || note "no uptime in the record" FAIL
grep -q '_Cfunc_kvmv_init' "$rec" && note "the record holds the crash report" OK || note "no crash report in the record" FAIL
grep -q 'line from run 1' "$rec" && note "the record holds the log's last lines" OK || note "no log in the record" FAIL
ls "$DIR"/crash.* > /dev/null 2>&1 && note "the crash report was left behind" FAIL || note "the crash report was taken" OK

RUN=2; END='segv'; REPORT=; run
rec=$(of_run 2)
[ "$(records)" -eq 2 ] && note "a second crash adds a second record" OK || note "$(records) records after two crashes" FAIL
grep -qx 'status 139' "$rec" && note "a SIGSEGV is status 139" OK || note "SIGSEGV recorded as: $(grep '^status' "$rec")" FAIL
sed -n '/^--- crash$/,/^--- log$/p' "$rec" | grep -qv '^---' \
    && note "a crash with no report records none" FAIL \
    || note "a crash with no report records none" OK

RUN=3; END='kill'; REPORT=; run
grep -qx "status 137" "$(of_run 3)" && note "a SIGKILL is status 137" OK || note "SIGKILL not recorded as 137" FAIL

RUN=4; END=0; REPORT=; run
[ "$(records)" -eq 0 ] && note "a clean exit removes every record" OK || note "a clean exit left $(records) records" FAIL

# The directory is made when it is missing, and a server that never runs
# leaves nothing to record.
rm -rf "$DIR"
RUN=5; END=1; REPORT=; run
[ "$(records)" -eq 1 ] && note "a missing directory is made" OK || note "a missing directory stayed missing" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases pass"
    exit 0
fi
echo "$fails case(s) FAIL"
exit 1
