#!/bin/sh
# Check that check-deployed reports what a device is running.
#
#   test-check-deployed.sh [path-to-check-deployed]
#
# The script under test needs a device, and this suite has none. It supplies
# one instead: a fake ssh earlier on PATH than the real one, which ignores its
# arguments and prints a report the case has written. Everything after the ssh
# call is then the real code, so the parsing, the comparison and the exit status
# are all exercised.
#
# The report is built from this checkout, so the matching case matches by
# construction rather than by a hash written down here that would rot on the
# next rebuild of a library.
#
# The probe that runs on the device is lifted out of the script and run against
# a directory this suite makes, which is the only way to reach it from here.
set -u

CHECK=${1:-$(dirname "$0")/check-deployed}
[ -f "$CHECK" ] || { echo "usage: test-check-deployed.sh <check-deployed>"; exit 1; }
CHECK=$(cd "$(dirname "$CHECK")" && pwd)/$(basename "$CHECK")

for tool in git gzip md5sum; do
    command -v "$tool" >/dev/null 2>&1 || {
        echo "SKIP: $tool is not on PATH"
        exit 2
    }
done

repo=$(git rev-parse --show-toplevel 2>/dev/null) || {
    echo "SKIP: not a git repository"
    exit 2
}

fails=0
note() { printf '  %-64s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# --- the fake device ---
#
# $REPORT names the file to print. The script under test pipes the probe into
# ssh on stdin, so that is read and discarded: leaving it unread would give the
# writing side a broken pipe.
mkdir -p "$work/bin"
cat > "$work/bin/ssh" <<'FAKE'
#!/bin/sh
cat > /dev/null
if [ "${FAKE_SSH_FAIL:-0}" = 1 ]; then
    echo "ssh: connect to host port 22: Connection refused" >&2
    exit 255
fi
[ -n "${REPORT:-}" ] && [ -f "$REPORT" ] && cat "$REPORT"
exit 0
FAKE
chmod +x "$work/bin/ssh"
PATH="$work/bin:$PATH"
export PATH

run() {
    # Prints the report and leaves the exit status in $status.
    ( cd "$repo" && sh "$CHECK" "$@" 2>&1 )
    return $?
}

# --- build a report that matches this checkout ---
head_md5() { git -C "$repo" cat-file -p "HEAD:$1" 2>/dev/null | md5sum | cut -d' ' -f1; }

content_md5() {
    case "$1" in
        *.gz) gzip -dc "$1" 2>/dev/null | md5sum | cut -d' ' -f1 ;;
        *)    md5sum "$1" 2>/dev/null | cut -d' ' -f1 ;;
    esac
}

sha=$(git -C "$repo" rev-parse --short HEAD)

build_report() {
    out=$1
    : > "$out"
    printf 'PID=442\n' >> "$out"
    printf 'EXE=/tmp/server/NanoKVM-Server\n' >> "$out"
    printf 'STAMP=dev.20260904.1302.%s\n' "${2:-$sha}" >> "$out"

    git -C "$repo" ls-tree --name-only HEAD server/dl_lib/ | while read -r path; do
        name=${path##*/}
        case "$name" in *.a) continue ;; esac
        sum=$(head_md5 "$path")
        printf 'FILE lib-disk %s %s\nFILE lib-run %s %s\n' "$name" "$sum" "$name" "$sum"
    done >> "$out"

    git -C "$repo" ls-tree --name-only HEAD kvmapp/system/init.d/ | while read -r path; do
        name=${path##*/}
        sum=$(head_md5 "$path")
        printf 'FILE initd-boot %s %s\nFILE initd-pkg %s %s\n' "$name" "$sum" "$name" "$sum"
    done >> "$out"

    if [ -d "$repo/web/dist" ]; then
        ( cd "$repo/web/dist" && find . -type f | sed 's#^\./##' | sort ) | while read -r name; do
            sum=$(content_md5 "$repo/web/dist/$name")
            printf 'FILE web-disk %s %s\nFILE web-run %s %s\n' "$name" "$sum" "$name" "$sum"
        done >> "$out"
    else
        printf 'ABSENT web-disk /kvmapp/server/web\nABSENT web-run /tmp/server/web\n' >> "$out"
    fi
}

echo "===== it refuses to guess ====="

out=$(cd "$repo" && sh "$CHECK" 2>&1); status=$?
[ "$status" -eq 2 ] && note "no device named exits 2" OK \
                    || note "no device named exits 2 (got $status)" FAIL
case "$out" in
    *usage*) note "no device named prints a usage line" OK ;;
    *)       note "no device named prints a usage line" FAIL ;;
esac

REPORT=; export REPORT
FAKE_SSH_FAIL=1; export FAKE_SSH_FAIL
out=$(run root@nowhere); status=$?
[ "$status" -eq 2 ] && note "an unreachable device exits 2, not 1" OK \
                    || note "an unreachable device exits 2, not 1 (got $status)" FAIL
FAKE_SSH_FAIL=0; export FAKE_SSH_FAIL

: > "$work/empty.report"
REPORT="$work/empty.report"; export REPORT
out=$(run root@nowhere); status=$?
[ "$status" -eq 2 ] && note "a device that answers nothing exits 2" OK \
                    || note "a device that answers nothing exits 2 (got $status)" FAIL

echo "===== a device that matches ====="

build_report "$work/match.report"
REPORT="$work/match.report"; export REPORT
out=$(run root@device); status=$?

[ "$status" -eq 0 ] && note "a matching device exits 0" OK \
                    || note "a matching device exits 0 (got $status)" FAIL

case "$out" in
    *CURRENT*) note "the server binary reads as CURRENT" OK ;;
    *)         note "the server binary reads as CURRENT" FAIL ;;
esac

case "$out" in
    *"do not match"*) note "a matching device says nothing about differences" FAIL ;;
    *)                note "a matching device says nothing about differences" OK ;;
esac

echo "===== a stale binary ====="

# HEAD~1 is a commit this checkout has, so the script can name it and count the
# distance. That is the case the whole script exists for.
prev=$(git -C "$repo" rev-parse --short HEAD~1 2>/dev/null)
if [ -z "$prev" ]; then
    note "a stale binary is named (needs two commits)" SKIP
else
    build_report "$work/stale.report" "$prev"
    REPORT="$work/stale.report"; export REPORT
    out=$(run root@device); status=$?

    [ "$status" -eq 1 ] && note "a stale binary exits 1" OK \
                        || note "a stale binary exits 1 (got $status)" FAIL
    case "$out" in
        *STALE*) note "a stale binary reads as STALE" OK ;;
        *)       note "a stale binary reads as STALE" FAIL ;;
    esac
    case "$out" in
        *"commits behind HEAD"*) note "it says how far behind" OK ;;
        *)                       note "it says how far behind" FAIL ;;
    esac
fi

# A stamp naming a commit this clone does not have must not be reported as
# current. It used to be possible to read "built from" as a match on the strength
# of the line existing at all.
build_report "$work/unknown.report" "0000000"
REPORT="$work/unknown.report"; export REPORT
out=$(run root@device); status=$?
[ "$status" -eq 1 ] && note "an unknown commit exits 1" OK \
                    || note "an unknown commit exits 1 (got $status)" FAIL
case "$out" in
    *"not a commit in this clone"*) note "an unknown commit is named as such" OK ;;
    *)                              note "an unknown commit is named as such" FAIL ;;
esac

# An unstamped binary is a release build. It is not a fault by itself, and it
# does mean the commit cannot be recovered, so it must not read as current.
build_report "$work/nostamp.report"
grep -v '^STAMP=' "$work/nostamp.report" > "$work/nostamp2.report"
printf 'STAMP=\n' >> "$work/nostamp2.report"
REPORT="$work/nostamp2.report"; export REPORT
out=$(run root@device); status=$?
case "$out" in
    *CURRENT*) note "an unstamped binary does not read as CURRENT" FAIL ;;
    *)         note "an unstamped binary does not read as CURRENT" OK ;;
esac

# No server at all is the loudest thing this can find.
build_report "$work/dead.report"
grep -v '^PID=' "$work/dead.report" > "$work/dead2.report"
printf 'PID=\n' >> "$work/dead2.report"
REPORT="$work/dead2.report"; export REPORT
out=$(run root@device); status=$?
case "$out" in
    *"no NanoKVM-Server process"*) note "a device with no server says so" OK ;;
    *)                             note "a device with no server says so" FAIL ;;
esac

echo "===== a library that does not match ====="

build_report "$work/lib.report"
# Corrupt libkvm.so in both copies, which is the state of a device that never
# received a rebuild.
sed 's/^FILE lib-disk libkvm.so .*/FILE lib-disk libkvm.so deadbeefdeadbeefdeadbeefdeadbeef/;
     s/^FILE lib-run libkvm.so .*/FILE lib-run libkvm.so deadbeefdeadbeefdeadbeefdeadbeef/' \
    "$work/lib.report" > "$work/lib2.report"
REPORT="$work/lib2.report"; export REPORT
out=$(run root@device); status=$?

[ "$status" -eq 1 ] && note "a stale library exits 1" OK \
                    || note "a stale library exits 1 (got $status)" FAIL
case "$out" in
    *libkvm.so*) note "the stale library is named" OK ;;
    *)           note "the stale library is named" FAIL ;;
esac
# Both copies are reported, because which one is wrong decides the repair.
case "$out" in
    *"differs, serving"*|*"differs"*"differs"*) note "both copies are reported" OK ;;
    *)                                          note "both copies are reported" FAIL ;;
esac

echo "===== one copy right and the other not ====="

build_report "$work/lag.report"
# Installed to /kvmapp and not picked up by the running copy in /tmp. This is a
# restart pending, not a failed deploy, and the two must not read the same.
sed 's/^FILE lib-run libkvm.so .*/FILE lib-run libkvm.so deadbeefdeadbeefdeadbeefdeadbeef/' \
    "$work/lag.report" > "$work/lag2.report"
REPORT="$work/lag2.report"; export REPORT
out=$(run root@device); status=$?

[ "$status" -eq 1 ] && note "a half-applied deploy exits 1" OK \
                    || note "a half-applied deploy exits 1 (got $status)" FAIL
case "$out" in
    *"right in one copy and not the other"*)
        note "a half-applied deploy is not called a mismatch" OK ;;
    *)  note "a half-applied deploy is not called a mismatch" FAIL ;;
esac

echo "===== the probe that runs on the device ====="

# Lifted out of the script rather than copied, so a change to one is a change
# to both.
sed -n '/^emit() {/,/^}/p' "$CHECK" > "$work/emit.sh"
if [ ! -s "$work/emit.sh" ]; then
    note "check-deployed defines emit" FAIL
else
    note "check-deployed defines emit" OK

    . "$work/emit.sh"

    mkdir -p "$work/tree"

    # Big enough and repetitive enough that two compression levels produce two
    # different deflate streams. A short payload compresses to the same bytes
    # at every level, which would make the pair below prove nothing: the first
    # version of this suite used "hello" and the case passed with the rule it
    # was written to test deleted.
    i=0
    while [ "$i" -lt 200 ]; do
        printf 'the same line over and over %s\n' "$i"
        i=$((i + 1))
    done > "$work/tree/plain.txt"

    # The same content compressed two ways, which is the shape of the real
    # difference: the device's assets and a local build carry different gzip
    # bytes for identical content, so comparing the compressed bytes reported
    # all 19 of them as changed.
    gzip -1 -c "$work/tree/plain.txt" > "$work/tree/a.gz"
    gzip -9 -c "$work/tree/plain.txt" > "$work/tree/b.gz"

    if cmp -s "$work/tree/a.gz" "$work/tree/b.gz"; then
        note "the two .gz fixtures differ in their bytes" FAIL
    else
        note "the two .gz fixtures differ in their bytes" OK
    fi

    got=$(emit slot "$work/tree")

    a=$(echo "$got" | awk '$3 == "a.gz" { print $4 }')
    b=$(echo "$got" | awk '$3 == "b.gz" { print $4 }')
    want=$(md5sum "$work/tree/plain.txt" | cut -d' ' -f1)

    [ -n "$a" ] && [ "$a" = "$b" ] \
        && note "two .gz holding the same bytes hash the same" OK \
        || note "two .gz holding the same bytes hash the same (got '$a' and '$b')" FAIL

    [ "$a" = "$want" ] \
        && note "a .gz is hashed by what it holds" OK \
        || note "a .gz is hashed by what it holds (got '$a')" FAIL

    plain=$(echo "$got" | awk '$3 == "plain.txt" { print $4 }')
    [ "$plain" = "$want" ] \
        && note "a plain file is hashed by its own bytes" OK \
        || note "a plain file is hashed by its own bytes (got '$plain')" FAIL

    absent=$(emit slot "$work/tree/nothing-here")
    case "$absent" in
        ABSENT*) note "a directory that does not exist reports ABSENT" OK ;;
        *)       note "a directory that does not exist reports ABSENT (got '$absent')" FAIL ;;
    esac
fi

echo "===== it reads nothing through the CRLF filter ====="

# git cat-file is the only reader that applies no conversion. A working tree on
# a Windows host holds CRLF for blobs that have none, so comparing those bytes
# against the device would report every shell script as different.
grep -q 'git cat-file -p "HEAD:' "$CHECK" \
    && note "repository hashes come from git cat-file" OK \
    || note "repository hashes come from git cat-file" FAIL

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
