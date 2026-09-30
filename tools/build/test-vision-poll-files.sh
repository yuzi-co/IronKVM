#!/bin/sh
# Check the two files libkvm's threads poll: /etc/kvm/hdmi_mode and the ION
# carveout summary.
#
#   test-vision-poll-files.sh [kvm_vision.cpp]
#
# Not destructive: the source is only read, and the compiled part runs in a
# temporary directory with a temporary hdmi_mode file.
#
# The HDMI detection thread asked for the mode on every pass, and a pass is
# 100ms, or 10ms while a resolution is being chosen. Each ask was an access, an
# fopen, an fread and an fclose of a file that only this library writes. It now
# stats the file and reads it when the stat moved, or when its last read is a
# second old. The cases below hold the three things that must not change: a
# write by set_hdmi_mode is seen on the very next pass, an edit from outside is
# seen once the stat shows it, and a pass that finds nothing new returns 0 the
# way a read of an unchanged file did.
#
# The watchdog thread ran "cat | grep | awk" through popen every 500ms while
# /etc/kvm/watchdog or /tmp/watchdog exists: a shell and three programs forked
# on the one core. It reads the file itself now, and the parse is checked
# against the text the board prints, including the 100% that the old
# two-character slice read as 10.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
VIS=${1:-$ROOT/support/sg2002/additional/kvm/src/kvm_vision.cpp}

[ -f "$VIS" ] || { echo "missing: $VIS"; exit 2; }

CXX=${CXX:-g++}
command -v "$CXX" >/dev/null 2>&1 || {
    echo "test-vision-poll-files.sh: needs $CXX, which is not on PATH." >&2
    exit 2
}

fails=0
note() {
    printf '  %-58s %s\n' "$1" "$2"
    case $2 in FAIL*) fails=$((fails + 1)) ;; esac
    return 0
}

# A function body with comment lines dropped, as in the other suites here.
body() { sed -n "/^[a-z0-9_ *]*[ *]$2(/,/^}/p" "$1" | grep -v '^[[:space:]]*//'; }
# The definition alone, comments kept, for compiling: the first line that
# starts in column 0 and names the function, through its closing brace. The
# pattern in body() also matches an indented call, which is harmless for a
# grep and fatal for a compile.
code() {
    awk -v name="$2" '
        !on && $0 ~ "^[a-z][a-z0-9_ *]*[ *]" name "\\(" { on = 1 }
        on { print }
        on && /^}/ { exit }' "$1"
}

echo "===== the shape of the two pollers ====="

mode_body=$(body "$VIS" get_hdmi_mode)
set_body=$(body "$VIS" set_hdmi_mode)
ion_body=$(body "$VIS" chack_ion)

printf '%s\n' "$mode_body" | grep -q 'access(' \
    && note "get_hdmi_mode does not call access" FAIL \
    || note "get_hdmi_mode does not call access" OK

# The stat has to decide before the open, or the open happens every pass.
printf '%s\n' "$mode_body" | awk '
    /file_stamp_due\(/ && !seen_open { due = 1 }
    /fopen\(/ { seen_open = 1 }
    END { exit !(due && seen_open) }' \
    && note "get_hdmi_mode asks the stamp before it opens" OK \
    || note "get_hdmi_mode asks the stamp before it opens" FAIL

printf '%s\n' "$set_body" | grep -q 'hdmi_mode_stamp.valid = 0' \
    && note "set_hdmi_mode forces the next read" OK \
    || note "set_hdmi_mode forces the next read" FAIL

printf '%s\n' "$ion_body" | grep -q 'popen\|system(' \
    && note "chack_ion runs no shell" FAIL \
    || note "chack_ion runs no shell" OK

printf '%s\n' "$ion_body" | grep -q 'ion_usage_rate_from_summary(' \
    && note "chack_ion parses with the tested helper" OK \
    || note "chack_ion parses with the tested helper" FAIL

grep -v '^[[:space:]]*//' "$VIS" | grep -q 'popen(' \
    && note "kvm_vision.cpp calls popen nowhere" FAIL \
    || note "kvm_vision.cpp calls popen nowhere" OK

echo "===== the stamp, the mode file and the summary, run ====="

work=$(mktemp -d) || exit 2
trap 'rm -rf "$work"' EXIT INT TERM

stamp_type=$(awk '
    /^typedef struct \{/ { buf = ""; on = 1 }
    on { buf = buf $0 "\n" }
    on && /^\} *file_stamp_t;/ { printf "%s", buf; exit }
    on && /^\}/ { on = 0 }' "$VIS")
[ -n "$stamp_type" ] || note "file_stamp_t found in the source" FAIL

max_age=$(sed -n 's/^#define hdmi_mode_max_age_ms[[:space:]]*\([0-9]*\)U*.*/\1/p' "$VIS")
[ -n "$max_age" ] || { note "hdmi_mode_max_age_ms found" FAIL; max_age=1000; }

cat > "$work/t.cpp" <<EOF
#include <fcntl.h>
#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>

#define hdmi_mode_path "$work/hdmi_mode"
#define hdmi_mode_max_age_ms ${max_age}U

static uint32_t fake_ms = 5000;
namespace vi_state_shared { uint32_t monotonic_ms() { return fake_ms; } }

static int opens = 0;
static FILE *counted_fopen(const char *p, const char *m) { opens++; return fopen(p, m); }

struct { uint8_t hdmi_mode; } kvmv_cfg = { 0 };
static void debug(const char *, ...) {}

$(code "$VIS" write_small_file)
$stamp_type
$(code "$VIS" file_stamp_due)
$(code "$VIS" file_stamp_record)
$(sed -n '/^file_stamp_t hdmi_mode_stamp/p' "$VIS")
$(code "$VIS" set_hdmi_mode)
#define fopen counted_fopen
$(code "$VIS" get_hdmi_mode)
#undef fopen
$(code "$VIS" ion_usage_rate_from_summary)

static int fails = 0;
static void check(const char *what, int ok)
{
    printf("  %-58s %s\n", what, ok ? "OK" : "FAIL");
    if (!ok) fails++;
}

// An edit from outside, stamped with a chosen mtime so the test does not
// depend on the clock of the file system under it.
static void edit(const char *text, long sec)
{
    FILE *f = fopen(hdmi_mode_path, "w");
    fputs(text, f);
    fclose(f);
    struct timespec t[2] = { { sec, 0 }, { sec, 0 } };
    utimensat(AT_FDCWD, hdmi_mode_path, t, 0);
}

int main()
{
    // No file: mode 0, nothing opened, no change reported.
    unlink(hdmi_mode_path);
    kvmv_cfg.hdmi_mode = 2;
    int r = get_hdmi_mode();
    check("a missing file reads as mode 0 without an open", r == 0 && kvmv_cfg.hdmi_mode == 0 && opens == 0);

    edit("1\n", 1000);
    r = get_hdmi_mode();
    check("a new file is read and its mode reported", r == 1 && kvmv_cfg.hdmi_mode == 1 && opens == 1);

    fake_ms += 10;
    r = get_hdmi_mode();
    check("an unchanged file is not opened again", r == 0 && opens == 1);
    fake_ms += 100;
    for (int i = 0; i < 8; i++) { r |= get_hdmi_mode(); fake_ms += 100; }
    check("nine passes inside the age limit open nothing", r == 0 && opens == 1);

    fake_ms += 200;
    r = get_hdmi_mode();
    check("the age limit forces one read, same mode, no change", r == 0 && opens == 2);

    // Same length, new mtime: seen on the next pass.
    fake_ms += 10;
    edit("2\n", 1001);
    r = get_hdmi_mode();
    check("an outside edit with a new mtime is seen at once", r == 1 && kvmv_cfg.hdmi_mode == 2 && opens == 3);

    // Same length and same mtime: the stat cannot see it, the age limit does.
    fake_ms += 10;
    edit("0\n", 1001);
    r = get_hdmi_mode();
    check("an edit the stat cannot see waits for the age limit", r == 0 && kvmv_cfg.hdmi_mode == 2 && opens == 3);
    fake_ms += hdmi_mode_max_age_ms;
    r = get_hdmi_mode();
    check("and is read once the limit passes", r == 1 && kvmv_cfg.hdmi_mode == 0 && opens == 4);

    // The library's own write is seen on the very next pass, whatever the
    // stat says. That is what the state machine relies on after it sets 1.
    fake_ms += 10;
    set_hdmi_mode(1);
    struct timespec t[2] = { { 1001, 0 }, { 1001, 0 } };
    utimensat(AT_FDCWD, hdmi_mode_path, t, 0);
    r = get_hdmi_mode();
    check("set_hdmi_mode is seen on the next pass", r == 1 && kvmv_cfg.hdmi_mode == 1 && opens == 5);

    // Out of range is rewritten as 0, as before.
    fake_ms += 10;
    edit("7\n", 1002);
    r = get_hdmi_mode();
    check("an out of range mode reads as 0", r == 1 && kvmv_cfg.hdmi_mode == 0);

    unlink(hdmi_mode_path);
    r = get_hdmi_mode();
    fake_ms += 10;
    edit("0\n", 1002);
    int before = opens;
    r = get_hdmi_mode();
    check("a file removed and put back is read again", opens == before + 1);

    // The summary as the board prints it, 2026-09-30.
    const char *board =
        "Summary:\n"
        "[0] carveout heap size:78643200 bytes, used:31010816 bytes\n"
        "usage rate:40%, memory usage peak 31010816 bytes\n"
        "\n"
        "Details:\n";
    check("the board's summary reads 40", ion_usage_rate_from_summary(board) == 40);
    check("a single digit reads whole", ion_usage_rate_from_summary("usage rate:5%, x") == 5);
    check("100 reads as 100, not 10", ion_usage_rate_from_summary("usage rate:100%, x") == 100);
    check("95 reads as 95", ion_usage_rate_from_summary("x\nusage rate:95%,") == 95);
    check("no rate line is an error", ion_usage_rate_from_summary("Summary:\n") == -1);
    check("a rate line with no number is an error", ion_usage_rate_from_summary("usage rate:%") == -1);
    check("no text is an error", ion_usage_rate_from_summary(NULL) == -1);

    return fails != 0;
}
EOF

if "$CXX" -std=gnu++17 -Wall -Wno-unused-function -o "$work/t" "$work/t.cpp" 2>"$work/cc.log"; then
    "$work/t" >"$work/run.log" 2>&1 || fails=$((fails + 1))
    cat "$work/run.log"
else
    note "the extracted code compiles" FAIL
    sed 's/^/    /' "$work/cc.log" | head -30
fi

echo
if [ "$fails" = 0 ]; then
    echo "test-vision-poll-files.sh: all passed"
    exit 0
fi
echo "test-vision-poll-files.sh: $fails failed"
exit 1
