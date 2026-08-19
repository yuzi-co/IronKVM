#!/bin/sh
# Check that a server build carries no debug metadata.
#
#   test-strip-flags.sh              source checks, then a real cross-compile
#   SKIP_BUILD=1 test-strip-flags.sh source checks only
#
# Run it from the repository root. The build case needs Docker and the image
# from the Dockerfile beside this file.
#
# The server binary runs from /tmp, which is tmpfs, so every byte of it is
# pinned RAM on a board that has 166MB. A stock build carries 6.7MB of DWARF
# and symbol tables that nothing on the device reads: there is no debugger
# there, and Go builds its own stack traces from .gopclntab, which -s and -w
# leave alone. That is the whole reason for the flags, and the reason the last
# two cases below exist - a build that dropped .gopclntab would pass "no debug
# sections" while making every panic unreadable.
#
# Three entry points build this binary and all three must agree, because a
# binary is deployed from whichever one the operator reached for. Each has two
# branches, stamped and unstamped, and the flags have to be in both: the
# Makefile's unstamped branch used to pass no -ldflags at all.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
WORK=${WORK:-$(mktemp -d)}

fails=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

echo "===== every build entry point strips ====="

# Count the lines that reach the linker, not every line that mentions one.
# "go build" also appears in comments and in the echo that announces the step,
# and the Makefile names GO_LDFLAGS twice: once to define it and once to use
# it. Only the definition can carry the flags.
linker_lines() {
    case $1 in
        Makefile) grep -n '^GO_LDFLAGS[[:space:]]*:=' "$ROOT/$1" ;;
        *)        grep -n '[^a-z]go build ' "$ROOT/$1" | grep -v '^[0-9]*:[[:space:]]*#' | grep -v 'echo' ;;
    esac
}

for file in Makefile server/build.sh tools/build/build-app.sh; do
    lines=$(linker_lines "$file" | grep -c . || true)
    stripped=$(linker_lines "$file" | grep -c -- '-s -w' || true)

    [ "$lines" -gt 0 ] && [ "$stripped" = "$lines" ] \
        && note "$file passes -s -w on every build line" OK \
        || note "$file passes -s -w on every build line ($stripped/$lines)" FAIL
done

# A build with no stamp must still strip. The Makefile expresses the stamp as
# an optional fragment, so -s -w has to sit outside that fragment.
grep -q 'ldflags "-s -w' "$ROOT/Makefile" \
    && note "the Makefile strips even without a build stamp" OK \
    || note "the Makefile strips even without a build stamp" FAIL

if [ -n "$SKIP_BUILD" ]; then
    echo
    echo "(build case skipped)"
    echo
    [ "$fails" -eq 0 ] && echo "all cases passed" || echo "$fails case(s) FAILED"
    exit "$fails"
fi

echo
echo "===== the built binary carries no debug metadata ====="

IMAGE=nanokvm-app-builder
if ! command -v docker >/dev/null 2>&1 || ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
    # An absent builder is not a defect in the three build entry points, and
    # every case above did run. Keep their verdict, name what could not run,
    # and exit 2 so tools/run-tests.sh counts this as skipped. Reporting it as
    # "builder image is present FAIL" said the build was broken when nothing
    # about the build had been looked at.
    echo
    echo "(build case skipped: no $IMAGE image here)"
    echo "    build it with: docker build -t $IMAGE tools/build"
    echo
    if [ "$fails" -ne 0 ]
    then
        echo "$fails source case(s) FAILED"
        exit 1
    fi
    echo "all source cases passed"
    exit 2
fi
note "builder image $IMAGE is present" OK

# Build into a scratch name so a failed run cannot leave a half-linked binary
# where a deploy would find it.
mkdir -p "$WORK"
# MSYS_NO_PATHCONV keeps Git Bash from rewriting the container-side paths. It
# turns "-w /src" into "C:/Program Files/Git/src" and the run dies on an
# invalid working directory, which this suite then reported as a failed
# cross-compile. The variable means nothing outside Git Bash, so it costs a
# Linux run nothing.
if MSYS_NO_PATHCONV=1 docker run --rm \
    -v "$ROOT/server:/src" -v "$ROOT/tools/build:/build" \
    -v nanokvm-gopath:/gopath -v nanokvm-gocache:/gocache \
    -w /src -e GOPATH=/gopath -e GOCACHE=/gocache \
    -e BUILD_STAMP="test.stripcheck" \
    "$IMAGE" sh -c 'sh /build/build-app.sh >/dev/null 2>&1 && readelf -S -W NanoKVM-Server && echo "--SIZE--" && wc -c < NanoKVM-Server && echo "--STAMP--" && strings -a NanoKVM-Server | grep -c "test.stripcheck"' \
    > "$WORK/probe.txt" 2>"$WORK/probe.err"
then
    note "the cross-compile succeeds" OK
else
    note "the cross-compile succeeds" FAIL
    sed 's/^/    /' "$WORK/probe.err" | tail -20
    echo
    echo "$fails case(s) FAILED"
    exit "$fails"
fi

sections=$(awk '{ if ($2 ~ /^\./) print $2 }' "$WORK/probe.txt")

debug=$(echo "$sections" | grep -c '^\.debug' || true)
[ "$debug" = 0 ] \
    && note "no .debug_* section survives" OK \
    || note "no .debug_* section survives ($debug found)" FAIL

for s in .symtab .strtab; do
    if echo "$sections" | grep -qx -- "$s"; then
        note "$s is gone" FAIL
    else
        note "$s is gone" OK
    fi
done

# -s -w must not reach the table Go reads to build a panic trace.
echo "$sections" | grep -qx '\.gopclntab' \
    && note ".gopclntab survives, so panics stay readable" OK \
    || note ".gopclntab survives, so panics stay readable" FAIL

# The stamp is the only way to tell which build is on a device, and it is
# passed through the same -ldflags string as -s -w.
stamp=$(sed -n '/--STAMP--/,$p' "$WORK/probe.txt" | tail -1)
[ -n "$stamp" ] && [ "$stamp" != 0 ] \
    && note "the build stamp still reaches the binary" OK \
    || note "the build stamp still reaches the binary" FAIL

size=$(sed -n '/--SIZE--/,/--STAMP--/p' "$WORK/probe.txt" | sed -n 2p)
printf '  %-58s %s bytes\n' "built size" "${size:-unknown}"

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
fi
exit "$fails"
