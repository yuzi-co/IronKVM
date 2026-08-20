#!/bin/sh
# Check that libkvm's threads can still be told to stop.
#
#   test-libkvm-thread-exit.sh [path-to-libkvm.so]
#
# kvmv_deinit sets kvmv_cfg.try_exit_thread and then joins the two threads that
# watch it. Neither the flag nor the join is visible in the shipped library, so
# this reads the machine code instead of the source: it is the compiled artefact
# that gets deployed, and the fault this exists to catch was invisible in the
# source.
#
# The fault, measured on the device 2026-08-20. watchdog_sf_feed is declared
# void* and returned nothing, so control fell off the end of a function that
# owes a value. That is undefined behaviour, and GCC is entitled to assume it
# never happens: it concluded the break out of the loop was unreachable and
# deleted the test that reaches it. The loop then held no reference to the exit
# flag at all. kvmv_deinit's join waited for a thread with no way out, the
# teardown never reached mmf_deinit, and every stop of the server left a VI
# channel pool and an ISP shared buffer in the carveout - 6,516,736 bytes a
# time, which only a reboot returns.
#
# Making the flag atomic did not fix it on its own. The load appeared and its
# result was still never examined, because the branch that consumed it was
# still unreachable. So this suite checks for the branch, not for the load.
#
# The offset of the flag inside kvmv_cfg is read out of kvmv_deinit rather than
# written down here: whichever byte kvmv_deinit stores to is the one the thread
# loops have to test, and the two cannot drift apart if the test derives one
# from the other.
set -u

DIR=$(dirname "$0")
LIB=${1:-$DIR/../../server/dl_lib/libkvm.so}

[ -f "$LIB" ] || {
    echo "test-libkvm-thread-exit.sh: no library at $LIB."
    exit 2
}

# The toolchain lives in the MaixCDK builder image, not on a workstation.
OBJDUMP=
NM=
for prefix in riscv64-unknown-linux-musl- riscv64-linux-gnu- riscv64-linux-musl- ''; do
    if command -v "${prefix}objdump" >/dev/null 2>&1 && command -v "${prefix}nm" >/dev/null 2>&1; then
        if "${prefix}objdump" -f "$LIB" 2>/dev/null | grep -qi riscv; then
            OBJDUMP=${prefix}objdump
            NM=${prefix}nm
            break
        fi
    fi
done

[ -n "$OBJDUMP" ] || {
    echo "test-libkvm-thread-exit.sh: no objdump that reads riscv64 on PATH."
    echo "run it in the MaixCDK builder image, where the cross binutils live."
    exit 2
}

fails=0
note() { printf '  %-62s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

sym_addr() {
    "$NM" -D "$LIB" 2>/dev/null | awk -v want="$1" '$3 ~ want { print $1; exit }'
}

# Disassemble one function by name, from its symbol to the next one.
disasm() {
    start=$(sym_addr "$1")
    [ -n "$start" ] || return 1
    end=$("$NM" -D "$LIB" 2>/dev/null | awk '{ print $1 }' | sort | awk -v s="$start" '
        $1 > s { print $1; exit }')
    [ -n "$end" ] || return 1
    "$OBJDUMP" -d --start-address="0x$start" --stop-address="0x$end" "$LIB" 2>/dev/null
}

DEINIT=$(sym_addr '^kvmv_deinit$')
if [ -z "$DEINIT" ]; then
    echo "test-libkvm-thread-exit.sh: no kvmv_deinit in $LIB."
    echo "the library changed shape, so nothing below is measured against it."
    exit 2
fi

echo "===== kvmv_deinit still asks the threads to stop ====="

disasm '^kvmv_deinit$' > "$work/deinit.txt" || {
    echo "test-libkvm-thread-exit.sh: cannot disassemble kvmv_deinit."
    exit 2
}

# The store of 1 into a byte of kvmv_cfg. It is the only sb in the function that
# writes a non-zero constant, and the joins follow it.
FLAG_OFF=$(awk '
    /\tli\t/ && /,1$/  { reg = $0; sub(/.*\tli\t/, "", reg); sub(/,1$/, "", reg); ones[reg] = 1 }
    /\tsb\t/ {
        f = $0; sub(/.*\tsb\t/, "", f)
        split(f, p, ",")
        if (p[1] in ones) {
            off = p[2]; sub(/\(.*/, "", off)
            print off
            exit
        }
    }' "$work/deinit.txt")

if [ -z "$FLAG_OFF" ]; then
    note "kvmv_deinit stores the exit flag" FAIL
    echo "  no store of a set byte in kvmv_deinit: the request is never made"
else
    note "kvmv_deinit stores the exit flag at offset $FLAG_OFF" OK
fi

grep -q 'pthread_join' "$work/deinit.txt" \
    && note "kvmv_deinit joins the threads it asked to stop" OK \
    || note "kvmv_deinit joins the threads it asked to stop" FAIL

echo
echo "===== each thread tests the flag, and can leave on it ====="

for fn in '_Z16watchdog_sf_feedPv:watchdog_sf_feed' '_Z22vi_subsystem_detectionPv:vi_subsystem_detection'; do
    sym=${fn%%:*}
    name=${fn##*:}

    if ! disasm "^$sym\$" > "$work/$name.txt"; then
        note "$name is in the library" FAIL
        continue
    fi

    if [ -z "$FLAG_OFF" ]; then
        note "$name tests the flag" FAIL
        continue
    fi

    # A load of that byte, and a conditional branch that consumes it close
    # behind. The register the load lands in has to be the one the branch reads:
    # a load whose result nothing examines is exactly the state the fixed
    # library came out of the compiler in on the first attempt.
    if awk -v off="$FLAG_OFF" '
        BEGIN { pending = 0 }
        {
            line = $0
            if (line ~ /\tlbu\t/) {
                f = line; sub(/.*\tlbu\t/, "", f)
                split(f, p, ",")
                o = p[2]; sub(/\(.*/, "", o)
                if (o == off) { reg = p[1]; pending = 8; next }
            }
            if (pending > 0) {
                # andi/mv may sit between the load and the branch.
                if (line ~ /\tandi\t/ || line ~ /\tmv\t/) {
                    f = line; sub(/.*\t(andi|mv)\t/, "", f)
                    split(f, p, ",")
                    if (p[2] == reg) reg = p[1]
                }
                if (line ~ /\tb(eq|ne|eqz|nez|ltu|geu)/) {
                    f = line; sub(/.*\tb[a-z]*\t/, "", f)
                    split(f, p, ",")
                    if (p[1] == reg || p[2] == reg) { print "yes"; exit }
                }
                pending = pending - 1
            }
        }' "$work/$name.txt" | grep -q yes
    then
        note "$name tests the flag and branches on it" OK
    else
        note "$name tests the flag and branches on it" FAIL
        echo "  the loop cannot leave: the compiler dropped the test"
    fi

    # The return that keeps the break reachable. Without it the break falls off
    # the end of a void* function, and the branch above disappears again.
    if awk -F'	' '$3 ~ /^(ret|jr)$/ { found = 1 } END { exit !found }' "$work/$name.txt"; then
        note "$name has a reachable return" OK
    else
        note "$name has a reachable return" FAIL
    fi
done

echo
if [ "$fails" -eq 0 ]; then
    echo "===== libkvm's threads can be told to stop ====="
    exit 0
fi

echo "===== $fails check(s) failed: a stop will leak the carveout ====="
exit 1
