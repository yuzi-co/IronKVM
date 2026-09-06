#!/bin/sh
# Prove that test-frame-buffer-pool.sh fails when the pool is wrong.
#
#   test-frame-buffer-pool-mutation.sh
#
# Not destructive: every mutation is made on a copy under mktemp, and the
# shipped source is only read.
#
# The pool test compiles code it lifts out of kvm_vision.cpp, so it is easy for
# it to look thorough and check nothing: a case that reads a field the mutation
# does not touch passes on a broken pool exactly as it does on a good one.
# Running it against deliberately broken copies is what separates the two.
#
# Each mutation here is a defect one edit away from the shipped text, and the
# first is the shipped text of a fortnight ago.
#
# NOT-APPLIED is a failure, not a skip. A sed that matches nothing leaves the
# copy identical, the test passes because nothing is wrong with it, and the run
# reads as proof when it proved nothing.
DIR=$(dirname "$0")
SRC=${1:-$DIR/../../support/sg2002/additional/kvm/src/kvm_vision.cpp}
TEST=$DIR/test-frame-buffer-pool.sh

[ -f "$SRC" ] || { echo "missing: $SRC"; exit 2; }
[ -f "$TEST" ] || { echo "missing: $TEST"; exit 2; }

# A mutant that never finishes is not a mutant the suite caught, and a suite
# that runs one unbounded takes its parent down with it: on 2026-09-06 a mutated
# source sent AddressSanitizer into a signal-handler loop, and this suite sat on
# it until tools/run-tests.sh killed the pair at 1800 seconds. A bound turns that
# into a line of output.
#
# The result is deliberately not counted as "the suite caught it". The suite
# reported nothing, so nothing is known, and pretending otherwise is how a
# mutation suite starts passing for the wrong reason.
MUTANT_TIMEOUT=${MUTANT_TIMEOUT:-120}
run_bounded() {
    if command -v timeout > /dev/null 2>&1; then
        timeout "$MUTANT_TIMEOUT" "$@"
        status=$?
        [ "$status" = 124 ] && return 124
        return "$status"
    fi
    "$@"
}

CXX=${CXX:-g++}
command -v "$CXX" > /dev/null 2>&1 || { echo "no C++ compiler ($CXX); cannot run here"; exit 2; }

pass=0
fail=0

try() {
    desc=$1
    shift

    d=$(mktemp -d)
    cp "$SRC" "$d/kvm_vision.cpp"
    "$@" "$d/kvm_vision.cpp"

    if cmp -s "$d/kvm_vision.cpp" "$SRC"; then
        printf '  %-14s %s\n' NOT-APPLIED "$desc"
        fail=$((fail + 1))
        rm -rf "$d"
        return
    fi

    run_bounded sh "$TEST" "$d/kvm_vision.cpp" > "$d/out" 2>&1
    status=$?
    if [ "$status" = 0 ]; then
        printf '  %-14s %s\n' SURVIVED "$desc"
        fail=$((fail + 1))
    elif [ "$status" = 2 ]; then
        # A mutation that stops the lift is not a mutation this suite tested.
        printf '  %-14s %s\n' UNTESTED "$desc"
        fail=$((fail + 1))
    elif [ "$status" = 124 ]; then
        # The suite never answered, so this mutation was not caught by it. A
        # non-zero exit from a timeout would otherwise read as "caught".
        printf '  %-14s %s (no answer in %ss)\n' HUNG "$desc" "$MUTANT_TIMEOUT"
        fail=$((fail + 1))
    else
        printf '  %-14s %s\n' caught "$desc"
        pass=$((pass + 1))
    fi
    rm -rf "$d"
}

# The ring as it was: one slot looked at, and three free buffers beside it
# reported as a full ring.
m_one_slot() {
    awk '
        /^kvmv_data_t\* get_save_buffer\(\)$/ { infn = 1 }
        infn && /^    for\(int i = 0; i < kvmv_data_buffer_size; i\+\+\)\{$/ {
            print "    for(int i = 0; i < 1; i++){"
            infn = 0
            next
        }
        { print }
    ' "$1" > "$1.new" && mv "$1.new" "$1"
}

# A slot handed out and never marked taken. Every request answers slot 1.
m_no_claim() { sed -i 's|^            buffer->in_use = 1;$|            buffer->in_use = 0;|' "$1"; }

# The release that does nothing. Four frames and the stream is over.
m_no_release() { sed -i 's|^        buffer->in_use = 0;$|        buffer->in_use = buffer->in_use;|' "$1"; }

# The reserve that forgets what it allocated, so every frame reallocates.
m_no_capacity() { sed -i 's|^    buffer->img_data_capacity = size;$|    buffer->img_data_capacity = 0;|' "$1"; }

# The reserve that hands back the old, too small allocation.
m_reserve_always_ok() {
    sed -i 's|^    if(buffer->p_img_data != NULL \&\& buffer->img_data_capacity >= size){$|    if(buffer->p_img_data != NULL){|' "$1"
}

# A zero length reserve accepted. realloc(p, 0) is allowed to free and answer
# NULL, which puts a NULL into a slot the caller is about to write.
m_reserve_zero() { sed -i 's|^    if(buffer == NULL \|\| size == 0){$|    if(buffer == NULL){|' "$1"; }

# free_kvmv_data without its NULL guard. Every unused slot holds a NULL
# pointer, so a NULL from the caller matches slot 0 and releases it.
m_free_no_null_guard() {
    sed -i 's|^    if(_pp_kvm_data == NULL \|\| \*_pp_kvm_data == NULL){$|    if(_pp_kvm_data == NULL){|' "$1"
}

# free_kvmv_data reading the frame type after clearing the flag, which is the
# ordering the comment above it argues against.
m_free_wrong_order() {
    awk '
        /^            uint8_t _type = kvmv_data_buffer\[i\]\.img_data_type;$/ { held = $0; next }
        /^            kvmv_data_buffer\[i\]\.in_use = 0;$/ && held != "" {
            print
            print "            kvmv_data_buffer[i].img_data_type = 0;"
            print held
            held = ""
            next
        }
        { print }
    ' "$1" > "$1.new" && mv "$1.new" "$1"
}

# Teardown that frees the memory and leaves the bookkeeping behind, so the pool
# reports capacity it no longer has.
m_free_all_keeps_capacity() {
    sed -i 's|^        kvmv_data_buffer\[i\].img_data_capacity = 0;$|        kvmv_data_buffer[i].img_data_capacity = 1;|' "$1"
}

# Teardown that leaves every slot marked taken.
m_free_all_keeps_flag() {
    sed -i 's|^        kvmv_data_buffer\[i\].in_use = 0;$|        kvmv_data_buffer[i].in_use = 1;|' "$1"
}

echo "===== every mutation must be caught ====="
try "the ring looks at one slot again"        m_one_slot
try "a handed out slot is not marked taken"   m_no_claim
try "release does not free the slot"          m_no_release
try "reserve forgets the capacity"            m_no_capacity
try "reserve accepts a too small allocation"  m_reserve_always_ok
try "reserve accepts a zero length"           m_reserve_zero
try "free drops its NULL guard"               m_free_no_null_guard
try "free reads the type after clearing"      m_free_wrong_order
try "teardown keeps the capacity"             m_free_all_keeps_capacity
try "teardown keeps the taken flag"           m_free_all_keeps_flag

echo
if [ "$fail" -eq 0 ]; then
    echo "===== all $pass mutations caught ====="
    exit 0
fi
echo "===== $fail of $((pass + fail)) mutations were not caught ====="
exit 1
