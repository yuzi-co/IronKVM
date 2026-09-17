#!/bin/sh
# Exercise the deploy guard's decisions, taken straight out of the script that
# ships so the test cannot drift from it.
#
#   test-deploy-guard.sh [path-to-deploy-server]
#
# Not destructive: every case runs against a temporary directory tree and the
# probe is stubbed. No device is touched.
DG=${1:-$(dirname "$0")/deploy-server}
[ -f "$DG" ] || { echo "usage: test-deploy-guard.sh <deploy-server>"; exit 1; }

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

sed -n '/^# --- verdict ---$/,/^# --- end verdict ---$/p'   "$DG" > "$WORK/verdict.sh"
sed -n '/^# --- snapshot ---$/,/^# --- end snapshot ---$/p' "$DG" > "$WORK/snapshot.sh"
sed -n '/^# --- preflight ---$/,/^# --- end preflight ---$/p' "$DG" > "$WORK/preflight.sh"
sed -n '/^# --- settle ---$/,/^# --- end settle ---$/p' "$DG" > "$WORK/settle.sh"
sed -n '/^# --- running ---$/,/^# --- end running ---$/p' "$DG" > "$WORK/running.sh"
sed -n '/^# --- install ---$/,/^# --- end install ---$/p' "$DG" > "$WORK/install.sh"
[ -s "$WORK/running.sh" ] || { echo "could not extract the running block"; exit 1; }
[ -s "$WORK/install.sh" ] || { echo "could not extract the install block"; exit 1; }
[ -s "$WORK/verdict.sh" ]  || { echo "could not extract the verdict block"; exit 1; }
[ -s "$WORK/snapshot.sh" ] || { echo "could not extract the snapshot block"; exit 1; }
[ -s "$WORK/preflight.sh" ] || { echo "could not extract the preflight block"; exit 1; }
sed -n '/^# --- standoff ---$/,/^# --- end standoff ---$/p' "$DG" > "$WORK/standoff.sh"
sed -n '/^# --- probe url ---$/,/^# --- end probe url ---$/p' "$DG" > "$WORK/probeurl.sh"
[ -s "$WORK/probeurl.sh" ] || { echo "could not extract the probe url block"; exit 1; }
[ -s "$WORK/settle.sh" ] || { echo "could not extract the settle block"; exit 1; }
[ -s "$WORK/standoff.sh" ] || { echo "could not extract the standoff block"; exit 1; }

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

echo "===== keep or restore ====="
# The guard exists because a manual backup only helps if someone is watching.
# The verdict must depend on whether the service answers, nothing else.
verdict_case() {
    desc="$1"; healthy="$2"; matches="$3"; want="$4"
    got=$(HEALTHY="$healthy" MATCHES="$matches" WORK="$WORK" sh -c '
        serving() { [ "$HEALTHY" = yes ]; }
        running_matches_installed() { [ "$MATCHES" = yes ]; }
        . "$WORK/verdict.sh"
        verdict
    ')
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

verdict_case "answers, and the running copy is the one installed" yes yes keep
verdict_case "does not answer"                                  no  yes restore

# The failure this was rewritten after. tmpfs filled up, the candidate copied in
# truncated, and the guard said OK - because S95nanokvm then ran a copy of the
# binary from /tmp and that copy was still the old good one. Serving proves a
# server is up. It does not prove it is the server just installed.
verdict_case "answers, but an old server is running"            yes no  restore

echo
echo "===== the known-good copy is only replaced by a proven one ====="
# The trap this closes: deploy twice in a row and a naive backup would snapshot
# the first broken binary as the thing to fall back to. A snapshot is only taken
# when the server that is running right now is answering.
snap_case() {
    desc="$1"; healthy="$2"; existing="$3"; want="$4"
    D="$WORK/d"; rm -rf "$D"; mkdir -p "$D/known-good"
    printf 'running\n' > "$D/current"
    [ "$existing" = none ] || printf '%s\n' "$existing" > "$D/known-good/binary"

    HEALTHY="$healthy" D="$D" WORK="$WORK" sh -c '
        serving() { [ "$HEALTHY" = yes ]; }
        CURRENT="$D/current"; GOOD="$D/known-good/binary"
        . "$WORK/snapshot.sh"
        snapshot_if_proven
    ' > /dev/null 2>&1

    if [ -f "$D/known-good/binary" ]; then got=$(cat "$D/known-good/binary"); else got=none; fi
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

snap_case "running server is healthy, so snapshot it"        yes "old"  "running"
snap_case "running server is sick, keep the old known-good"  no  "old"  "old"
snap_case "running server is sick and nothing saved yet"     no  none   none
snap_case "healthy with nothing saved yet"                   yes none   "running"

echo
echo "===== a failed restore must be loud ====="
# If the restore also fails to come up, the box is in the state this script
# exists to prevent. It must report that rather than exit 0 and look successful.
got=$(WORK="$WORK" sh -c '
    serving() { return 1; }
    . "$WORK/verdict.sh"
    if restore_worked; then echo quiet; else echo loud; fi
')
[ "$got" = "loud" ] && note "restore that does not come up reports failure" OK \
                    || note "restore failure reported as $got" FAIL

got=$(WORK="$WORK" sh -c '
    serving() { return 0; }
    . "$WORK/verdict.sh"
    if restore_worked; then echo quiet; else echo loud; fi
')
[ "$got" = "quiet" ] && note "restore that comes up reports success" OK \
                     || note "successful restore reported as $got" FAIL

echo
echo "===== refuse before touching anything ====="
# A short copy is silent, and it is what this script was written after. The
# new file is held beside the old one until the rename, so the filesystem that
# holds CURRENT needs room for it first.
pre_case() {
    desc="$1"; elf="$2"; free="$3"; want="$4"
    got=$(ELF="$elf" FREE="$free" NEEDED=24000 WORK="$WORK" sh -c '
        candidate_is_elf()  { [ "$ELF" = yes ]; }
        install_free_kb()   { echo "$FREE"; }
        candidate_size_kb() { echo "$NEEDED"; }
        . "$WORK/preflight.sh"
        preflight && echo proceed || echo refuse
    ')
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

pre_case "intact candidate and room beside CURRENT"   yes 48000 proceed
pre_case "candidate is not an ELF"                    no  48000 refuse
pre_case "no room to hold the new file beside the old" yes 20000 refuse

# The boundary is 1.5x the candidate: one candidate for the file held beside the
# old one, and half again as margin. NEEDED is 24000, so the line is 36000.
pre_case "exactly at the 1.5x boundary"             yes 36000 proceed
pre_case "one kB under the boundary"                yes 35999 refuse

echo
echo
echo "===== waiting for the restart, not for any answer ====="
# The bug this replaced: the guard called wait_for_service, which returned true
# the moment ANY server answered - and right after `restart` the one answering is
# still the old process, because the restart is detached and has not killed it
# yet. The verdict was then taken against a /tmp copy that had not been replaced,
# so it read "serves, but not the binary just installed" and rolled back a good
# build. Measured on a device: installed and FAILED were stamped the same second.
#
# So the wait has to be for the state the guard actually wants, and it has to
# keep asking until the deadline instead of judging once.
settle_case() {
    desc="$1"; script="$2"; want="$3"
    got=$(WORK="$WORK" SCRIPT="$script" sh -c '
        eval "$SCRIPT"
        . "$WORK/settle.sh"
        settled_within 5 && echo keep || echo restore
    ')
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

# The old process answers immediately and the new binary is not running yet.
# Giving up here is exactly the bug.
settle_case "old server still answering, restart not landed yet"     'n=0; serving() { return 0; }; running_matches_installed() { return 1; }' restore

# The normal case: the restart lands a second or two in. The guard must wait for
# it rather than judge on the first look.
settle_case "restart lands after a moment"     'n=0; serving() { return 0; }; running_matches_installed() { n=$((n+1)); [ "$n" -ge 3 ]; }' keep

settle_case "healthy straight away"     'serving() { return 0; }; running_matches_installed() { return 0; }' keep

# A candidate that cannot start never serves, and the guard must still give up
# rather than hang.
settle_case "candidate never serves"     'serving() { return 1; }; running_matches_installed() { return 1; }' restore

# A server that comes up slowly under SD contention must not be failed early.
settle_case "slow start, answers on the third look"     'n=0; serving() { n=$((n+1)); [ "$n" -ge 3 ]; }; running_matches_installed() { return 0; }' keep

echo
echo "===== the rollback is judged the same way ====="
# The same race, in reverse: right after restoring, the process still running is
# the bad one, and `serving` alone would report "rolled back, serving again"
# while the binary the guard just removed was still the one answering.
# Both paths, counted rather than located: an assertion about how many lines
# apart two statements sit breaks on the next edit and proves nothing anyway.
calls=$(grep -c 'settled_within "\$DEPLOY_TIMEOUT"' "$DG")
[ "$calls" = 2 ]     && note "both the deploy and the rollback wait for the restart to land" OK     || note "settled_within is called $calls time(s), want 2 (deploy and rollback)" FAIL

# The racy helper must be gone, not merely bypassed on one path.
grep -q 'wait_for_service' "$DG"     && note "wait_for_service is still present and can be reached again" FAIL     || note "the helper that judged on the first answer is gone" OK

echo
echo "===== the running server is judged by what it maps ====="
# /tmp/server is a link to /kvmapp/server, so a file at that path compared with
# CURRENT is a file compared with itself. What tells an old process from a new
# one is the kernel's own record: a mapping of a file that was renamed over is
# printed with " (deleted)" after its path.
#
# PROC stands in for /proc, and pidof is stubbed to name the pids under it. Each
# line of the second argument is "pid|maps line".
running_case() {
    desc="$1"; maps="$2"; current="$3"; want="$4"
    P="$WORK/proc"; rm -rf "$P"; mkdir -p "$P"
    printf '%s\n' "$maps" | while IFS='|' read -r pid line; do
        [ -n "$pid" ] || continue
        mkdir -p "$P/$pid"
        printf '%s\n' "$line" >> "$P/$pid/maps"
    done
    got=$(PROC="$P" CURRENT="$current" WORK="$WORK" sh -c '
        pidof() { ls "$PROC"; }
        . "$WORK/running.sh"
        running_matches_installed && echo new || echo old
    ' 2>/dev/null)
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

mkdir -p "$WORK/app/dl_lib"
: > "$WORK/app/NanoKVM-Server"
: > "$WORK/app/dl_lib/libkvm.so"

running_case "a process started from the installed binary" \
"100|3fb0000000-3fb1000000 r-xp 00000000 b3:03 1234   $WORK/app/NanoKVM-Server" \
    "$WORK/app/NanoKVM-Server" new

running_case "the old process, whose binary was renamed over" \
"100|3fb0000000-3fb1000000 r-xp 00000000 b3:03 1234   $WORK/app/NanoKVM-Server (deleted)" \
    "$WORK/app/NanoKVM-Server" old

running_case "no server process at all" "" "$WORK/app/NanoKVM-Server" old

# The deploy of a library passes CURRENT=.../libkvm.so. The binary is not what
# changed, so only the library's mapping can say which process is new.
running_case "a library deploy, the library freshly mapped" \
"100|3fb2000000-3fb2100000 r-xp 00000000 b3:03 99     $WORK/app/dl_lib/libkvm.so" \
    "$WORK/app/dl_lib/libkvm.so" new

running_case "a library deploy, the old library still mapped" \
"100|3fb0000000-3fb1000000 r-xp 00000000 b3:03 1234   $WORK/app/NanoKVM-Server
100|3fb2000000-3fb2100000 r-xp 00000000 b3:03 98     $WORK/app/dl_lib/libkvm.so (deleted)" \
    "$WORK/app/dl_lib/libkvm.so" old

# A restart that has started the new server while the old one is still dying.
running_case "old and new process side by side" \
"100|3fb0000000-3fb1000000 r-xp 00000000 b3:03 1234   $WORK/app/NanoKVM-Server (deleted)
200|3fb0000000-3fb1000000 r-xp 00000000 b3:03 1235   $WORK/app/NanoKVM-Server" \
    "$WORK/app/NanoKVM-Server" new

# A path that only starts with CURRENT is a different file.
running_case "a mapping of NanoKVM-Server.deploy-new is not the binary" \
"100|3fb0000000-3fb1000000 r-xp 00000000 b3:03 1234   $WORK/app/NanoKVM-Server.deploy-new" \
    "$WORK/app/NanoKVM-Server" old

# The kernel prints the resolved path, so a CURRENT given through the link must
# be resolved before it is compared. Git Bash on Windows copies instead of
# linking, so this case runs only where ln -s makes a link.
if ln -s "$WORK/app" "$WORK/link" 2>/dev/null && [ -L "$WORK/link" ]; then
    running_case "CURRENT named through the /tmp/server link" \
"100|3fb0000000-3fb1000000 r-xp 00000000 b3:03 1234   $WORK/app/NanoKVM-Server" \
        "$WORK/link/NanoKVM-Server" new
fi

echo
echo "===== the install renames, and never writes over the running file ====="
# The server runs from /kvmapp/server. cp over its executable fails with
# ETXTBSY, and cp over a library it has loaded truncates pages it is executing.
# A rename gives the file a new inode and leaves the old one to the process.
I="$WORK/inst"
rm -rf "$I"; mkdir -p "$I"
printf 'old\n' > "$I/current"
printf 'new\n' > "$I/candidate"
exec 3< "$I/current"
before=$(ls -i "$I/current" | awk '{print $1}')

if WORK="$WORK" I="$I" sh -c '. "$WORK/install.sh"; install_file "$I/candidate" "$I/current"'; then
    note "install_file succeeds" OK
else
    note "install_file succeeds" FAIL
fi
after=$(ls -i "$I/current" | awk '{print $1}')
[ "$(cat "$I/current")" = new ] \
    && note "the destination holds the new content" OK \
    || note "the destination holds the new content" FAIL
[ "$before" != "$after" ] \
    && note "the destination is a new inode, renamed into place" OK \
    || note "the destination is a new inode, renamed into place" FAIL
[ "$(cat <&3)" = old ] \
    && note "a process holding the old file still reads the old content" OK \
    || note "a process holding the old file still reads the old content" FAIL
exec 3<&-
[ -e "$I/current.deploy-new" ] \
    && note "no temporary file is left behind" FAIL \
    || note "no temporary file is left behind" OK

# A copy that does not match its source must not be renamed into place: a short
# copy is how a deploy once installed a truncated binary.
printf 'old\n' > "$I/current"
got=$(WORK="$WORK" I="$I" sh -c '
    cmp() { return 1; }
    . "$WORK/install.sh"
    install_file "$I/candidate" "$I/current" && echo installed || echo refused
' 2>/dev/null)
[ "$got" = refused ] && [ "$(cat "$I/current")" = old ] \
    && note "a copy that does not match the source is not installed" OK \
    || note "a copy that does not match the source is not installed" FAIL
[ -e "$I/current.deploy-new" ] \
    && note "a refused copy is cleaned up" FAIL \
    || note "a refused copy is cleaned up" OK

# Both installs go through it: the candidate and the rollback.
calls=$(grep -c '^    if ! install_file ' "$DG")
[ "$calls" = 2 ] \
    && note "the deploy and the rollback both install by rename" OK \
    || note "install_file is used $calls time(s), want 2 (deploy and rollback)" FAIL
grep -qE '^[[:space:]]*cp "\$(NEW|GOOD)" "\$CURRENT"' "$DG" \
    && note "nothing copies straight onto CURRENT" FAIL \
    || note "nothing copies straight onto CURRENT" OK

echo
echo "===== the script still parses ====="
sh -n "$DG" && note "deploy-server is valid shell" OK || note "deploy-server does not parse" FAIL
grep -q 'setsid' "$DG" \
    && note "detaches, so a dropped ssh cannot abandon it mid-deploy" OK \
    || note "does not detach - a dropped ssh would leave it half done" FAIL

echo
echo "===== standing off the supervisor ====="
# S98supervise polls every five seconds and acts on a server it finds gone or
# not answering. A deploy stops the server and restarts it, which looks like
# exactly that. On 2026-08-17 the supervisor started the server in the middle of
# a deploy and then killed it 63 seconds later as hung, while the deploy was
# still deciding whether its own candidate worked.
#
# Unlike the updater, this script is not the thing being replaced, so it can and
# must clean up after itself.
standoff_case() {
    desc="$1"; body="$2"; want="$3"
    got=$(WORK="$WORK" BODY="$body" sh -c '
        UPDATE_MARKER=$WORK/marker
        export UPDATE_MARKER
        rm -f "$UPDATE_MARKER"
        . "$WORK/standoff.sh"
        eval "$BODY"
        [ -f "$UPDATE_MARKER" ] && echo present || echo absent
    ' 2>/dev/null)
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

standoff_case "hold_supervisor writes the marker"      "hold_supervisor"                  present
standoff_case "the marker carries a timestamp"         "hold_supervisor; grep -q \"^[0-9][0-9]*\$\" \"\$UPDATE_MARKER\" || rm -f \"\$UPDATE_MARKER\"" present
standoff_case "release_supervisor removes it"          "hold_supervisor; release_supervisor" absent
standoff_case "releasing without holding is harmless"  "release_supervisor"               absent
standoff_case "holding twice is harmless"              "hold_supervisor; hold_supervisor" present

# The marker must not survive the script. A deploy that exits leaving it behind
# suspends the supervisor for the whole stand-off window, and the one thing worse
# than a supervisor that acts too early is one that never acts at all.
grep -q 'trap .*release_supervisor' "$DG" \
    && note "the marker is cleared on every exit path, including a failure" OK \
    || note "the marker can outlive the deploy" FAIL

# The two scripts have to name the same file or the stand-off protects nothing.
DEPLOY_MARKER=$(sed -n 's|^UPDATE_MARKER=${UPDATE_MARKER:-\(.*\)}$|\1|p' "$DG" | head -1)
SUPERVISE_MARKER=$(sed -n 's|^UPDATE_MARKER=${SUPERVISE_UPDATE_MARKER:-\(.*\)}$|\1|p' \
    "$(dirname "$DG")/../service/S98supervise" | head -1)
[ -n "$DEPLOY_MARKER" ] && [ "$DEPLOY_MARKER" = "$SUPERVISE_MARKER" ] \
    && note "deploy and supervisor agree on the marker path ($DEPLOY_MARKER)" OK \
    || note "marker paths disagree: deploy '$DEPLOY_MARKER' vs supervisor '$SUPERVISE_MARKER'" FAIL

echo
echo "===== the probe follows the protocol the server is configured for ====="

# The guard waits for 200. With `proto: https` the server answers 307 on plain
# HTTP, so a probe fixed at http:// would fail on a board that is serving
# perfectly and the guard would restore the previous binary over a good one.
# That happened on 2026-08-19, the first deploy after the protocol was switched.
url_case() {
    desc=$1; want=$2
    cfg=$WORK/server.yaml
    cat > "$cfg"
    got=$(SERVER_CONFIG="$cfg" sh -c '. "$1"; default_probe_url' _ "$WORK/probeurl.sh")
    [ "$got" = "$want" ] && note "$desc -> $got" OK || note "$desc -> $got, want $want" FAIL
}

url_case "https on the standard port" "https://127.0.0.1/" <<'YAML'
proto: https
port:
    http: 80
    https: 443
YAML

url_case "http on the standard port" "http://127.0.0.1/" <<'YAML'
proto: http
port:
    http: 80
    https: 443
YAML

# A port that is not the scheme's default has to appear, or the probe asks the
# wrong socket and the guard reads a healthy server as a dead one.
url_case "https on a port of its own" "https://127.0.0.1:8443/" <<'YAML'
proto: https
port:
    http: 80
    https: 8443
YAML

url_case "http on a port of its own" "http://127.0.0.1:8080/" <<'YAML'
proto: http
port:
    http: 8080
    https: 443
YAML

# The scheme decides which port is read. Reading the wrong one is how a probe
# ends up on a socket nothing is listening to.
url_case "https ignores the http port" "https://127.0.0.1:8443/" <<'YAML'
proto: https
port:
    http: 8080
    https: 8443
YAML

# The lookup is scoped to the port block. Any other block is free to carry a key
# of the same name, and taking the first match in the file would aim the probe
# at whatever that block meant.
url_case "the port comes from the port block, not another one" "https://127.0.0.1/" <<'YAML'
proto: https
mirror:
    https: 8443
port:
    http: 80
    https: 443
YAML

url_case "a config with no proto falls back to http" "http://127.0.0.1/" <<'YAML'
port:
    http: 80
YAML

got=$(SERVER_CONFIG="$WORK/there-is-no-such-file.yaml"       sh -c '. "$1"; default_probe_url' _ "$WORK/probeurl.sh")
[ "$got" = "http://127.0.0.1/" ]     && note "an unreadable config falls back to http -> $got" OK     || note "an unreadable config gave $got, want http://127.0.0.1/" FAIL

# It has to be the default rather than the value, or an operator cannot point
# the probe somewhere else.
grep -q 'DEPLOY_URL=${DEPLOY_URL:-$(default_probe_url)}' "$DG"     && note "DEPLOY_URL still overrides what the config says" OK     || note "DEPLOY_URL is not an overridable default" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "===== all deploy guard cases pass ====="
else
    echo "===== $fails case(s) failed ====="
    exit 1
fi
