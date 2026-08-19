#!/bin/sh
# Prove that test-server-stop.sh fails when stop_services is wrong.
#
#   test-server-stop-mutation.sh
#
# Not destructive: every mutation is made on a copy under mktemp, and the
# shipped script is only ever read.
#
# The guard this covers is two lines, and two lines are easy to test in a way
# that passes whatever they say. A case that only ever asserts "the tree is
# still there" passes on a script that never cleans up at all, which on a tmpfs
# board is its own defect. Running the cases against deliberately broken copies
# is what separates a guard from a comment.
#
# NOT-APPLIED is a failure, not a skip. A sed that matches nothing leaves the
# copy identical, the cases pass because there is nothing wrong with it, and the
# run reads as proof when it proved nothing.
DIR=$(dirname "$0")
S95=${1:-$DIR/../../kvmapp/system/init.d/S95nanokvm}
TEST=$DIR/test-server-stop.sh
for f in "$S95" "$TEST"; do
    [ -f "$f" ] || { echo "missing: $f"; exit 1; }
done

pass=0
fail=0

try() {
    desc=$1
    shift

    d=$(mktemp -d)
    cp "$S95" "$d/S95nanokvm"
    "$@" "$d/S95nanokvm"

    if cmp -s "$d/S95nanokvm" "$S95"; then
        printf '  %-14s %s\n' NOT-APPLIED "$desc"
        fail=$((fail + 1))
        rm -rf "$d"
        return
    fi

    if sh "$TEST" "$d/S95nanokvm" > /dev/null 2>&1; then
        printf '  %-14s %s\n' SURVIVED "$desc"
        fail=$((fail + 1))
    else
        printf '  %-14s %s\n' caught "$desc"
        pass=$((pass + 1))
    fi
    rm -rf "$d"
}

# The defect as it shipped: remove both copies whatever is still running. This
# is the one that cost a working board its whole web UI.
m_unconditional() {
    sed -i 's@^    pidof [A-Za-z_]* >/dev/null 2>&1 || rm -rf@    rm -rf@' "$1"
}

# The lazy fix: stop removing anything. tmpfs is RAM here, so a copy left behind
# is pinned memory until the board reboots.
m_noremove() {
    sed -i '\@rm -rf "\$SERVER_DST"@d; \@rm -rf "\$SYSTEM_DST"@d' "$1"
}

# One process decides both copies. They start and stop separately, so a server
# that survived would keep kvm_system's copy in RAM for nothing.
m_shared() {
    sed -i 's@^    pidof kvm_system >/dev/null@    pidof NanoKVM-Server >/dev/null@' "$1"
}

# The guard inverted: keep the copy when the process is gone, and remove it from
# underneath the process that is still using it.
m_inverted() {
    sed -i 's@^    pidof NanoKVM-Server >/dev/null 2>&1 ||@    pidof NanoKVM-Server >/dev/null 2>\&1 \&\&@' "$1"
}

echo "===== every mutation must be caught ====="
try "the removal goes back to unconditional" m_unconditional
try "nothing is ever removed"                m_noremove
try "one process decides both copies"        m_shared
try "the guard is inverted"                  m_inverted

echo
if [ "$fail" -eq 0 ]; then
    echo "===== all $pass mutations caught ====="
else
    echo "===== $fail of $((pass + fail)) mutations were not caught ====="
    exit 1
fi
