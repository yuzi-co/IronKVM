#!/bin/sh
# Check that a wedged VPSS channel is detected and the VI channel rebuilt.
#
#   test-vpss-wedge-rebuild.sh [kvm_vision.cpp]
#
# Not destructive: the source is only read, and the compiled part runs in a
# temporary directory.
#
# The fault this covers, measured on 2026-09-04: the VPSS channel stopped
# handing out frames while the VI device kept running. VIDevFPS held at about
# 50, the SendOK counter of the channel stayed frozen and its FrameRate read 0,
# and the kernel reported "CVI_VPSS_GetChnFrame fail" with "jobs wait(0)
# work(0) done(0)" once a second. It lasted 27 minutes and ended only because
# the server was restarted. An earlier episode the same day ran for 17 minutes.
#
# The board already knows how to recover. An HDMI resolution change sets
# reopen_cam_flag, check_kvmv calls cam->restart, and the channel comes back.
# That is what ended the 10:43 episode at 10:59. These cases check that the
# reader now asks for the same restart without waiting for the source to
# change, and that it asks rarely, because a reopen of the MMF channel is the
# operation that can exhaust the carveout heap.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
VIS=${1:-$ROOT/support/sg2002/additional/kvm/src/kvm_vision.cpp}

[ -f "$VIS" ] || { echo "missing: $VIS"; exit 2; }

CXX=${CXX:-g++}
command -v "$CXX" >/dev/null 2>&1 || {
    echo "test-vpss-wedge-rebuild.sh: needs $CXX, which is not on PATH." >&2
    exit 2
}

fails=0
note() {
    printf '  %-58s %s\n' "$1" "$2"
    case $2 in FAIL*) fails=$((fails + 1)) ;; esac
    return 0
}

# A function body with comment lines dropped, so prose about a call is not
# mistaken for the call. The return type may carry a star, because
# vi_subsystem_detection is declared "void* vi_subsystem_detection(...)".
body() { sed -n "/^[a-z0-9_ *]*[ *]$2(/,/^}/p" "$1" | grep -v '^[[:space:]]*//'; }

work=$(mktemp -d) || exit 2
trap 'rm -rf "$work"' EXIT INT TERM

echo "===== the reader reports what the channel did ====="

read_body=$(body "$VIS" kvmv_read_img)

# One site each. The frame site sits on the successful pop, so an encode that
# fails afterwards still counts as a live channel, and the empty-call site sits
# on the one return that means the whole call produced nothing.
frames=$(printf '%s\n' "$read_body" | grep -c 'wedge_note_frame(&vi_wedge)')
[ "$frames" = 1 ] \
    && note "the reader notes a frame in exactly one place" OK \
    || note "the reader notes a frame in $frames places, wanted 1" FAIL

noframes=$(printf '%s\n' "$read_body" | grep -c 'wedge_note_no_frame(&vi_wedge')
[ "$noframes" = 1 ] \
    && note "the reader notes an empty call in exactly one place" OK \
    || note "the reader notes an empty call in $noframes places, wanted 1" FAIL

# The frame site has to be the pop and not an encode. If it moved to a success
# return, then a channel that hands out frames the encoder rejects would look
# wedged, and the board would reopen the channel over an encoder fault.
printf '%s\n' "$read_body" | grep -A6 'mmf_vi_frame_pop_native' \
    | grep -q 'wedge_note_frame(&vi_wedge)' \
    && note "a frame is noted at the pop, not at an encode" OK \
    || note "a frame is noted at the pop, not at an encode" FAIL

# The empty-call site has to be the return that gives up, and not one of the
# early returns. Codes -4, -5, -6 and -7 all mean the reader never looked.
printf '%s\n' "$read_body" | grep -B12 'return IMG_NOT_EXIST;' \
    | grep -q 'wedge_note_no_frame(&vi_wedge' \
    && note "an empty call is noted only where the loop gives up" OK \
    || note "an empty call is noted only where the loop gives up" FAIL

printf '%s\n' "$read_body" | grep -q 'kvmv_hdmi_signal_active()' \
    && note "the reader passes the VI device state, not a guess" OK \
    || note "the reader passes the VI device state, not a guess" FAIL

# Asking is the whole point, and the ask has to sit inside the branch the
# policy guards. Outside it the request would be raised on every empty call,
# which is the threshold and the cooldown both gone at once.
printf '%s\n' "$read_body" | grep -A4 'if (wedge_note_no_frame(&vi_wedge' \
    | grep -q '__atomic_store_n(&vi_wedge_rebuild_request, 1' \
    && note "the reader raises the request, and only when told to" OK \
    || note "the reader raises the request, and only when told to" FAIL

# debug() prints nothing unless the server passed a flag to kvmv_init that it
# does not pass, and it drops its arguments on the floor besides: it takes a
# variadic list and hands only the format to printf. Both of these events are
# rare, both are rate limited by the cooldown, and both describe a fault that
# otherwise leaves the board blank with nothing in the log to say why.
printf '%s\n' "$read_body" | grep -q 'printf("\[kvmv\]channel handed out nothing' \
    && note "the ask is logged whether or not debug is on" OK \
    || note "the ask is logged whether or not debug is on" FAIL

echo "===== the restart happens on the detection thread ====="

det_body=$(body "$VIS" vi_subsystem_detection)

printf '%s\n' "$det_body" | grep -q '__atomic_load_n(&vi_wedge_rebuild_request' \
    && note "the detection thread reads the request" OK \
    || note "the detection thread reads the request" FAIL

printf '%s\n' "$det_body" | grep -A2 '__atomic_load_n(&vi_wedge_rebuild_request' \
    | grep -q '__atomic_store_n(&vi_wedge_rebuild_request, 0' \
    && note "the request is cleared before the guards decide" OK \
    || note "the request is cleared before the guards decide" FAIL

printf '%s\n' "$det_body" | grep -q 'cam->restart' \
    && note "the detection thread is the one that restarts" OK \
    || note "the detection thread is the one that restarts" FAIL

# The reader must not restart the camera itself. vi_mutex is taken in exactly
# one place, so it does not order the reader against the detection thread, and
# two concurrent rebuilds are the case that exhausts the carveout heap.
printf '%s\n' "$read_body" | grep -q 'cam->restart' \
    && note "the reader does not restart the camera itself" FAIL \
    || note "the reader does not restart the camera itself" OK

for guard in 'vi_detect_state != 1' 'reopen_cam_flag == 0' \
             'hdmi_reading_flag == 0' 'hdmi_stop_flag != 1'; do
    printf '%s\n' "$det_body" | grep -q "$guard" \
        && note "the rebuild stands down on $guard" OK \
        || note "the rebuild stands down on $guard" FAIL
done

printf '%s\n' "$det_body" | grep -q 'printf("\[kvmv\]rebuilding the VI channel' \
    && note "the rebuild is logged whether or not debug is on" OK \
    || note "the rebuild is logged whether or not debug is on" FAIL

# A rebuild that never happened because something else was already
# transitioning looks exactly like a rebuild that happened and did not help,
# so the stood-down case needs a line of its own.
printf '%s\n' "$det_body" | grep -q 'printf("\[kvmv\]a rebuild was asked for while' \
    && note "a rebuild that stands down says so" OK \
    || note "a rebuild that stands down says so" FAIL

# vi_mutex really is taken in one place. The comment above the rebuild says so
# and the design depends on it, so a second lock site has to break this.
locks=$(grep -v '^[[:space:]]*//' "$VIS" \
    | grep -c 'pthread_mutex_timedlock(&vi_mutex\|pthread_mutex_lock(&vi_mutex)')
[ "$locks" = 1 ] \
    && note "vi_mutex still has exactly one lock site" OK \
    || note "vi_mutex has $locks lock sites, the design assumes 1" FAIL

echo "===== the policy, run for real ====="

# Lift the tunables, the state and both functions out of the shipped file, so
# what runs below is what ships rather than a copy that can drift away from it.
grep '^#define wedge_' "$VIS" > "$work/lifted.h"
awk '/^typedef struct \{/{buf=""} {buf = buf $0 "\n"} /^\} wedge_state_t;/{printf "%s", buf; exit}' \
    "$VIS" >> "$work/lifted.h"
sed -n '/^void wedge_note_frame(/,/^}/p' "$VIS" >> "$work/lifted.h"
sed -n '/^uint8_t wedge_note_no_frame(/,/^}/p' "$VIS" >> "$work/lifted.h"

grep -q 'wedge_note_no_frame(wedge_state_t \*w, uint32_t now_ms, uint8_t vi_live)' \
    "$work/lifted.h" \
    && note "the policy takes a clock and a VI state, not globals" OK \
    || note "the policy takes a clock and a VI state, not globals" FAIL

cat > "$work/main.cpp" <<'EOF'
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "lifted.h"

static int fails = 0;
static void check(const char *what, unsigned long got, unsigned long want)
{
    if (got != want) {
        printf("  %-58s FAIL (got %lu, wanted %lu)\n", what, got, want);
        fails++;
    } else {
        printf("  %-58s OK\n", what);
    }
}

// Feed n empty calls, one every step_ms, and count the rebuilds asked for.
static unsigned feed(wedge_state_t *w, unsigned n, uint32_t start_ms,
    uint32_t step_ms, uint8_t vi_live, uint32_t *end_ms)
{
    unsigned rebuilds = 0;
    uint32_t now = start_ms;
    for (unsigned i = 0; i < n; i++) {
        if (wedge_note_no_frame(w, now, vi_live)) {
            rebuilds++;
        }
        now += step_ms;
    }
    if (end_ms != NULL) {
        *end_ms = now;
    }
    return rebuilds;
}

int main(void)
{
    // The board fails about once a second in this fault, so a step of 1000ms
    // is the shape of the real thing.
    {
        wedge_state_t w = { 0, 0, 0, wedge_first_cooldown_ms, 0 };
        check("one failure short of the threshold asks for nothing",
            feed(&w, wedge_fail_threshold - 1, 1000, 1000, 1, NULL), 0);
    }
    {
        wedge_state_t w = { 0, 0, 0, wedge_first_cooldown_ms, 0 };
        check("the threshold with a live VI device asks once",
            feed(&w, wedge_fail_threshold, 1000, 1000, 1, NULL), 1);
    }
    {
        // A source that stopped is the common case, and it must never reach a
        // rebuild however long it lasts.
        wedge_state_t w = { 0, 0, 0, wedge_first_cooldown_ms, 0 };
        check("a dead VI device never asks, however long",
            feed(&w, 5000, 1000, 1000, 0, NULL), 0);
    }
    {
        // A reader that spins far faster than the board does must not reach a
        // rebuild on count alone.
        wedge_state_t w = { 0, 0, 0, wedge_first_cooldown_ms, 0 };
        check("a fast spin inside the minimum run asks for nothing",
            feed(&w, wedge_fail_threshold * 4, 1000, 1, 1, NULL), 0);
    }
    {
        // The same spin does ask once it has also lasted long enough.
        wedge_state_t w = { 0, 0, 0, wedge_first_cooldown_ms, 0 };
        unsigned first = feed(&w, wedge_fail_threshold * 4, 1000, 1, 1, NULL);
        unsigned later = wedge_note_no_frame(&w, 1000 + wedge_min_run_ms, 1) ? 1 : 0;
        check("the same spin asks once the minimum run has passed",
            first + later, 1);
    }
    {
        // A frame is the whole recovery signal.
        wedge_state_t w = { 0, 0, 0, wedge_first_cooldown_ms, 0 };
        feed(&w, wedge_fail_threshold - 1, 1000, 1000, 1, NULL);
        wedge_note_frame(&w);
        check("a frame clears the run, so a short one asks for nothing",
            feed(&w, wedge_fail_threshold - 1, 100000, 1000, 1, NULL), 0);
    }
    {
        wedge_state_t w = { 0, 0, 0, wedge_first_cooldown_ms, 0 };
        uint32_t t = 1000;
        feed(&w, wedge_fail_threshold, t, 1000, 1, &t);
        // Still failing, but inside the first cooldown.
        check("a second ask inside the cooldown is refused",
            wedge_note_no_frame(&w, w.last_rebuild_ms + wedge_first_cooldown_ms - 1, 1) ? 1 : 0, 0);
        check("a second ask after the cooldown is allowed",
            wedge_note_no_frame(&w, w.last_rebuild_ms + wedge_first_cooldown_ms, 1) ? 1 : 0, 1);
    }
    {
        // The backoff, checked as a sequence rather than as a single step.
        // Each ask is made at the earliest moment the previous cooldown allows.
        wedge_state_t w = { 0, 0, 0, wedge_first_cooldown_ms, 0 };
        uint32_t t = 1000;
        feed(&w, wedge_fail_threshold, t, 1000, 1, &t);
        uint32_t want[5] = { 10000, 20000, 40000, 60000, 60000 };
        unsigned ok = 0;
        for (int i = 0; i < 5; i++) {
            uint32_t at = w.last_rebuild_ms + want[i];
            if (wedge_note_no_frame(&w, at - 1, 1) == 0 &&
                    wedge_note_no_frame(&w, at, 1) != 0) {
                ok++;
            }
        }
        check("the cooldown backs off 10s 20s 40s 60s and caps", ok, 5);
    }
    {
        wedge_state_t w = { 0, 0, 0, wedge_first_cooldown_ms, 0 };
        uint32_t t = 1000;
        feed(&w, wedge_fail_threshold, t, 1000, 1, &t);
        wedge_note_no_frame(&w, w.last_rebuild_ms + wedge_first_cooldown_ms, 1);
        wedge_note_frame(&w);
        check("a frame puts the cooldown back to the first one",
            w.cooldown_ms, wedge_first_cooldown_ms);
    }
    {
        // The monotonic clock wraps every 49 days and the board is expected to
        // outlive that. Every comparison is unsigned, so a run that straddles
        // the wrap has to behave like any other.
        wedge_state_t w = { 0, 0, 0, wedge_first_cooldown_ms, 0 };
        uint32_t start = 0xFFFFFFFFU - (wedge_fail_threshold / 2) * 1000U;
        check("a run that straddles the clock wrap still asks",
            feed(&w, wedge_fail_threshold, start, 1000, 1, NULL), 1);
    }
    {
        // If the counter wrapped to 0 it would take first_fail_ms with it and
        // the run would restart, so a channel wedged for long enough would
        // stop being recovered.
        wedge_state_t w = { 0, 0, 0, wedge_first_cooldown_ms, 0 };
        w.fail_count = 0xFFFFFFFFU;
        w.first_fail_ms = 1000;
        wedge_note_no_frame(&w, 100000, 1);
        check("the failure counter saturates rather than wrapping",
            w.fail_count, 0xFFFFFFFFU);
        check("a saturated counter keeps the start of its run",
            w.first_fail_ms, 1000);
    }
    {
        // Neither is called with a null pointer today, and neither may crash
        // if that changes. ASAN would catch the dereference.
        wedge_note_frame(NULL);
        check("a null state asks for nothing and does not crash",
            wedge_note_no_frame(NULL, 1000, 1) ? 1 : 0, 0);
    }

    return fails == 0 ? 0 : 1;
}
EOF

if "$CXX" -std=c++11 -O1 -g -fsanitize=address,undefined -I"$work" \
        -o "$work/t" "$work/main.cpp" 2>"$work/cc.log"; then
    "$work/t" || fails=$((fails + 1))
else
    note "the lifted wedge policy compiles" FAIL
    sed -n '1,15p' "$work/cc.log"
fi

echo
if [ "$fails" = 0 ]; then
    echo "test-vpss-wedge-rebuild.sh: all cases passed"
    exit 0
fi
echo "test-vpss-wedge-rebuild.sh: $fails case(s) failed"
exit 1
