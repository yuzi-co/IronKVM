#!/bin/sh
# Run every test suite under tools/ and report one line for each.
#
#   run-tests.sh [substring]      run only the suites whose path matches
#   run-tests.sh --container      run the whole sweep in the release host image
#   TEST_TIMEOUT=600 run-tests.sh bound each suite, in seconds
#
# Prefer --container on a Windows workstation. The cost of a sweep there is
# process creation, not the tests: one suite that takes 12s in a Linux
# container takes 70s in Git Bash. The image also carries the tools that seven
# suites need, so they run instead of reporting that they cannot.
#
# The exit status of a suite is the whole contract:
#
#   0   every case passed
#   2   the suite cannot run here, and it says why on the last line: a tool it
#       needs is absent, an artefact no checkout carries was not supplied, or
#       it is destructive and refuses to run outside a throwaway container
#   1   a case failed, and that is a defect in what the suite tests
#
# That distinction is why this file exists. Running every suite bare used to
# report a dozen failures of which none was a defect: five wanted arguments,
# three wanted a tool, and two had harnesses that behaved one way on Windows
# and another on Linux. A dozen red lines that mean nothing are worse than no
# sweep at all, because they teach the next reader to stop running the sweep.
#
# A skipped suite is not a passing suite. Its count is printed on a line of its
# own so that nobody reads past it.
ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT" || exit 1


# --container re-runs this script inside the release host image. That is worth
# a flag rather than a line in a README for two reasons, both measured on
# 2026-08-19. It is about six times faster: tools/service/test-supervise.sh
# takes 70s in Git Bash and 12s in a Linux container, because the cost is
# process creation and not the tests. And the image carries sfdisk, zstd,
# e2fsprogs, mtools, parted and device-tree-compiler, so seven suites that skip
# on a Windows workstation run there instead.
if [ "$1" = --container ]
then
    shift
    IMAGE=ironkvm-release-host
    if ! command -v docker >/dev/null 2>&1 || ! docker image inspect "$IMAGE" >/dev/null 2>&1
    then
        echo "no $IMAGE image here. build it with:" >&2
        echo "  docker build -t $IMAGE tools/release" >&2
        exit 2
    fi
    # Mount the checkout at the same path inside as outside. tools/release
    # does the same, for the same reason: one suite cross-compiles by calling
    # the workstation's own daemon through that socket, and it bind-mounts a
    # directory of this checkout. The daemon resolves that path on the host, so
    # a checkout that sits at /repo in here and somewhere else out there hands
    # it a path that does not exist, and the cross-compile fails for a reason
    # that has nothing to do with the build.
    MSYS_NO_PATHCONV=1 exec docker run --rm \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v "$ROOT:$ROOT" -w "$ROOT" "$IMAGE" sh tools/run-tests.sh "$@"
fi
PATTERN=$1
TIMEOUT=${TEST_TIMEOUT:-1800}

OUT=$(mktemp -d)
trap 'rm -rf "$OUT"' EXIT

pass=0
skip=0
fail=0
slow=0
failed=

for t in $(find tools -name 'test-*.sh' | sort)
do
    if [ -n "$PATTERN" ]
    then
        case "$t" in
            *"$PATTERN"*) ;;
            *) continue ;;
        esac
    fi

    log="$OUT/$(echo "$t" | tr / _).log"
    start=$(date +%s)
    if command -v timeout >/dev/null 2>&1
    then
        timeout "$TIMEOUT" sh "$t" > "$log" 2>&1
    else
        sh "$t" > "$log" 2>&1
    fi
    rc=$?
    took=$(($(date +%s) - start))

    case $rc in
        0)
            pass=$((pass + 1))
            printf 'PASS  %4ss  %s\n' "$took" "$t"
            ;;
        2)
            skip=$((skip + 1))
            printf 'SKIP  %4ss  %s\n' "$took" "$t"
            tail -2 "$log" | sed 's/^/            /'
            ;;
        124)
            slow=$((slow + 1))
            failed="$failed $t"
            printf 'TIME  %4ss  %s (bound was TEST_TIMEOUT=%s)\n' "$took" "$t" "$TIMEOUT"
            echo "            it did not finish, which is not the same as a case failing."
            echo "            the mutation suites are the slow ones. Try --container."
            ;;
        *)
            fail=$((fail + 1))
            failed="$failed $t"
            printf 'FAIL  %4ss  %s (exit %s)\n' "$took" "$t" "$rc"
            tail -8 "$log" | sed 's/^/            /'
            ;;
    esac
done

echo
echo "$pass passed, $fail failed"
if [ "$slow" -gt 0 ]
then
    echo "$slow did not finish inside TEST_TIMEOUT=$TIMEOUT"
fi
if [ "$skip" -gt 0 ]
then
    echo "$skip skipped: those suites did not run, and a skip proves nothing"
fi
if [ $((fail + slow)) -gt 0 ]
then
    echo "failed:$failed"
    exit 1
fi
exit 0
