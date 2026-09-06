# Make AddressSanitizer usable, or say that it is not.
#
# Source this from a suite that compiles with -fsanitize=address, before it
# runs what it built:
#
#   . "$(dirname "$0")/asan-env.sh"
#   asan_run "$work/t"          # instead of "$work/t"
#
# ASan maps its shadow memory at a fixed address, and it needs the kernel to
# hand out no more entropy than that mapping allows: 28 bits on x86-64. The
# kernel behind Docker Desktop hands out 32. The mapping then fails for some
# address layouts and not others, so a suite does not fail, it fails about one
# run in four. Measured in the release host image on 2026-09-06:
# test-frame-buffer-pool.sh segfaulted 11 times in 40 runs, and passed every
# time when it was run alone often enough to look fine.
#
# Two things follow, and this file does both.
#
# ASAN_OPTIONS, so a failure is a failure. Without abort_on_error ASan takes a
# signal inside its own signal handler, prints AddressSanitizer:DEADLYSIGNAL,
# and takes the same signal again for ever: on 2026-09-06 that gave one suite
# its whole 1800 second budget under tools/run-tests.sh --container, and the
# sweep never reached the suites after it. handle_segv and handle_sigbus cover
# the faults ASan has nothing to report about yet, which the mutation suites
# found: a wild pointer raises the signal before abort_on_error is consulted.
# Nothing is lost by turning them off. An overrun near its allocation is caught
# by the redzone instrumentation and still names itself, and a wild one is
# still a fatal signal, which is what a suite acts on.
#
# asan_run, so the fault does not happen. Disabling ASLR for the child puts the
# mapping back within reach. setarch needs the personality syscall, which
# Docker's default seccomp profile blocks, so tools/run-tests.sh --container
# passes --security-opt seccomp=unconfined. With that, the same suite passed 30
# runs out of 30.
#
# If the entropy is too high and setarch cannot be used, ASAN_USABLE is 0 and
# the suite says so rather than reporting a fault it cannot stand behind.

ASAN_OPTIONS=${ASAN_OPTIONS:-abort_on_error=1:handle_segv=0:handle_sigbus=0}
export ASAN_OPTIONS

ASAN_USABLE=1
ASAN_NOTE=""
asan_no_aslr=0

asan_entropy=$(cat /proc/sys/vm/mmap_rnd_bits 2>/dev/null)
case $asan_entropy in
    ''|*[!0-9]*)
        # Not Linux, or the file is not readable. Nothing to work around.
        ;;
    *)
        if [ "$asan_entropy" -gt 28 ]; then
            if command -v setarch > /dev/null 2>&1 &&
               setarch "$(uname -m)" -R true > /dev/null 2>&1; then
                asan_no_aslr=1
            else
                ASAN_USABLE=0
                ASAN_NOTE="vm.mmap_rnd_bits is $asan_entropy and AddressSanitizer needs 28 or less; setarch -R is blocked here, so the sanitizer would fail at random"
            fi
        fi
        ;;
esac

# asan_run runs a sanitized binary the way it has to be run here.
asan_run() {
    if [ "$asan_no_aslr" = 1 ]; then
        setarch "$(uname -m)" -R "$@"
    else
        "$@"
    fi
}
