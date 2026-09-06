#!/bin/sh
# Check the frame buffer pool in kvm_vision.cpp.
#
#   test-frame-buffer-pool.sh [path/to/kvm_vision.cpp]
#
# Not destructive: the source is only read, and the test binary is built and
# run inside mktemp. It needs a host C++ compiler and exits 2 without one.
#
# The pool is five short functions and a four element array, and it decides
# whether a frame allocation happens once or thirty times a second. It also
# decides whether a slot comes back after a failure. None of it can be reached
# from the device build: the library needs MMF, VENC and a camera, and there is
# no harness here that provides any of them.
#
# So lift the functions out of the file and compile them on their own. They
# depend on nothing but the struct, the array and to_roll, and a test that
# compiles the shipped text is a test of the shipped code rather than of a copy
# that drifts.
#
# What these cases are for, in order of how much they would cost to get wrong:
#
#   - A slot that is not handed back is gone for the life of the process. Four
#     of those and the stream stops with IMG_BUFFER_FULL for ever.
#   - A reserve that reallocates every time gives back the per frame mmap that
#     the whole change exists to remove.
#   - The ring used to look at one slot and give up if that slot was busy, with
#     three free ones beside it.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
SRC=${1:-$ROOT/support/sg2002/additional/kvm/src/kvm_vision.cpp}

[ -f "$SRC" ] || { echo "missing: $SRC"; exit 2; }
CXX=${CXX:-g++}
command -v "$CXX" > /dev/null 2>&1 || { echo "no C++ compiler ($CXX); cannot run here"; exit 2; }

. "$(dirname "$0")/asan-env.sh"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# Lift a function by name, from its signature line to the closing brace in
# column zero. Every function involved is written that way in this file.
lift() {
    sed -n "/^[A-Za-z_].*[ *]$1(/,/^}/p" "$SRC"
}

{
    echo '#include <cstdint>'
    echo '#include <cstdlib>'
    echo '#include <cstring>'
    echo '#include <cstdio>'
    echo '#define IMG_BUFFER_FULL -3'
    echo '#define IMG_VENC_ERROR  -2'
    echo '#define IMG_NOT_EXIST   -1'
    echo '#define IMG_MJPEG_TYPE   0'
    # The one line of context the lifted code needs, taken from the file so a
    # change to the ring size is picked up here too.
    grep '^#define kvmv_data_buffer_size' "$SRC"
    # Several structs in this file open with a bare "typedef struct {", so take
    # the one that ends with the name wanted: keep the lines since the last
    # opening and print them when that closing line arrives.
    awk '
        /^typedef struct \{$/ { n = 0; held[n++] = $0; taking = 1; next }
        taking { held[n++] = $0 }
        /^\} kvmv_data_t;$/ {
            if (taking) { for (i = 0; i < n; i++) print held[i] }
            taking = 0
        }
    ' "$SRC"
    echo 'kvmv_data_t kvmv_data_buffer[kvmv_data_buffer_size];'
    echo 'uint8_t kvmv_data_buffer_index = 0;'
    lift to_roll
    lift get_save_buffer
    lift release_save_buffer
    lift reserve_save_buffer
    lift free_kvmv_data
    lift free_all_kvmv_data
} > "$work/pool.cpp"

for fn in to_roll get_save_buffer release_save_buffer reserve_save_buffer free_kvmv_data free_all_kvmv_data; do
    grep -q "$fn" "$work/pool.cpp" || { echo "cannot lift $fn from $SRC"; exit 2; }
done

# The test body sits after the lifted code in one translation unit, so it can
# see the array directly and check the slot flags rather than inferring them.
cat > "$work/cases.cpp" <<'CASES'

int fails = 0;
static void note(const char *what, bool ok)
{
    printf("  %-58s %s\n", what, ok ? "OK" : "FAIL");
    if (!ok) fails++;
}

static void reset()
{
    free_all_kvmv_data();
    kvmv_data_buffer_index = 0;
}

int main()
{
    printf("===== the pool hands out every slot =====\n");
    reset();

    kvmv_data_t *taken[kvmv_data_buffer_size];
    bool distinct = true;
    for (int i = 0; i < kvmv_data_buffer_size; i++) {
        taken[i] = get_save_buffer();
        if (taken[i] == NULL) distinct = false;
        for (int j = 0; j < i; j++) {
            if (taken[i] == taken[j]) distinct = false;
        }
    }
    note("all four slots are handed out, each one once", distinct);
    note("a fifth request answers NULL", get_save_buffer() == NULL);

    // The version this replaces advanced the index by one and gave up unless
    // that single slot was free.
    release_save_buffer(taken[2]);
    kvmv_data_t *again = get_save_buffer();
    note("a released slot is found wherever it sits in the ring", again == taken[2]);

    printf("\n===== an allocation outlives the frame in it =====\n");
    reset();

    kvmv_data_t *b = get_save_buffer();
    note("reserve of 0 bytes is refused", !reserve_save_buffer(b, 0));
    note("reserve of a NULL slot is refused", !reserve_save_buffer(NULL, 16));

    note("the first reserve allocates", reserve_save_buffer(b, 4096) && b->p_img_data != NULL);
    uint8_t *first = b->p_img_data;
    memset(first, 0xA5, 4096);

    note("a smaller reserve keeps the same allocation",
         reserve_save_buffer(b, 1024) && b->p_img_data == first);
    note("an equal reserve keeps the same allocation",
         reserve_save_buffer(b, 4096) && b->p_img_data == first);
    note("capacity does not shrink", b->img_data_capacity == 4096);

    note("a larger reserve grows capacity",
         reserve_save_buffer(b, 8192) && b->img_data_capacity == 8192);

    printf("\n===== a slot comes back, its memory does not =====\n");
    reset();

    b = get_save_buffer();
    reserve_save_buffer(b, 2048);
    b->img_data_type = 4;
    uint8_t *held = b->p_img_data;

    int type = free_kvmv_data(&held);
    note("the frame type survives the release", type == 4);
    note("the slot is free again", b->in_use == 0);
    note("the allocation is kept", b->p_img_data != NULL);

    // The ring keeps rotating, so the slot comes back around rather than
    // straight back. What matters is that when it does, the memory is the
    // memory it had, and a reserve that fits does not call the allocator.
    kvmv_data_t *back = NULL;
    for (int i = 0; i < kvmv_data_buffer_size && back != b; i++) {
        kvmv_data_t *s = get_save_buffer();
        if (s == b) back = s;
        else release_save_buffer(s);
    }
    note("the slot comes back around", back == b);
    note("it reuses the allocation it had",
         reserve_save_buffer(b, 2048) && b->p_img_data == held);

    printf("\n===== a release the pool does not own changes nothing =====\n");
    reset();

    b = get_save_buffer();
    reserve_save_buffer(b, 512);
    uint8_t stranger[8];
    uint8_t *p = stranger;
    note("a foreign pointer answers IMG_NOT_EXIST", free_kvmv_data(&p) == IMG_NOT_EXIST);
    note("a foreign pointer does not free a slot", b->in_use == 1);

    p = NULL;
    note("a NULL pointer answers IMG_NOT_EXIST", free_kvmv_data(&p) == IMG_NOT_EXIST);
    note("a NULL argument answers IMG_NOT_EXIST", free_kvmv_data(NULL) == IMG_NOT_EXIST);

    // Every slot starts with p_img_data NULL, and a NULL from the caller must
    // not match those and release one of them.
    reset();
    p = NULL;
    free_kvmv_data(&p);
    bool none_taken = true;
    for (int i = 0; i < kvmv_data_buffer_size; i++) {
        if (kvmv_data_buffer[i].in_use != 0) none_taken = false;
    }
    note("a NULL pointer does not match an unused slot", none_taken);

    printf("\n===== teardown returns everything =====\n");
    reset();
    for (int i = 0; i < kvmv_data_buffer_size; i++) {
        kvmv_data_t *s = get_save_buffer();
        reserve_save_buffer(s, 1024 * (i + 1));
    }
    free_all_kvmv_data();
    bool cleared = true;
    for (int i = 0; i < kvmv_data_buffer_size; i++) {
        if (kvmv_data_buffer[i].p_img_data != NULL) cleared = false;
        if (kvmv_data_buffer[i].img_data_capacity != 0) cleared = false;
        if (kvmv_data_buffer[i].in_use != 0) cleared = false;
    }
    note("free_all clears the pointer, the capacity and the flag", cleared);
    note("the pool is usable again after teardown", get_save_buffer() != NULL);

    printf("\n");
    if (fails == 0) {
        printf("all cases passed\n");
        return 0;
    }
    printf("%d case(s) FAILED\n", fails);
    return 1;
}
CASES


cat "$work/pool.cpp" "$work/cases.cpp" > "$work/test.cpp"

# The sanitizer is only used when it can be trusted here. asan-env.sh explains
# what makes it untrustworthy and why the answer is not simply to drop it.
if [ "$ASAN_USABLE" = 1 ]; then
    sanitize="-fsanitize=address,undefined"
else
    sanitize=""
    printf '  %-58s %s\n' "the sanitizer runs" "SKIP ($ASAN_NOTE)"
fi

if ! "$CXX" -std=c++17 -O1 $sanitize -o "$work/test" "$work/test.cpp" 2> "$work/cc.log"; then
    # Without a sanitizer runtime the build still has to work.
    if ! "$CXX" -std=c++17 -O1 -o "$work/test" "$work/test.cpp" 2>> "$work/cc.log"; then
        echo "the lifted pool does not compile:"
        sed 's/^/    /' "$work/cc.log"
        exit 1
    fi
fi

asan_run "$work/test"
status=$?
[ -s "$work/cc.log" ] && grep -q warning "$work/cc.log" && {
    echo
    echo "compiler warnings:"
    sed 's/^/    /' "$work/cc.log"
}
exit $status
