#!/bin/sh
# Check kvm_system's video watchdog decisions (#59).
#
#   test-video-watchdog.sh [path/to/kvm_system/main]
#
# Not destructive. video_watchdog.cpp has no MaixCDK dependency, so it builds
# here with the host compiler; the rest of kvm_system does not, and the last
# section only reads it.
#
# The watchdog reboots the board when the server's video library stops touching
# /tmp/nanokvm_wd. It used to count missed seconds from the moment it was
# switched on, and a directory at /tmp/watchdog switched it on: the watchdog
# daemon makes one at every boot. So kvm_system without a server rebooted the
# board ten seconds after it started.
ROOT=$(cd "$(dirname "$0")/../.." && pwd)
MAIN=${1:-$ROOT/support/sg2002/kvm_system/main}
SRC=$MAIN/lib/system_state/video_watchdog.cpp
STATE=$MAIN/lib/system_state/system_state.cpp
LOOP=$MAIN/src/main.cpp

for f in "$SRC" "$STATE" "$LOOP"; do
    [ -f "$f" ] || { echo "missing: $f"; exit 2; }
done
CXX=${CXX:-g++}
command -v "$CXX" > /dev/null 2>&1 || { echo "no C++ compiler ($CXX); cannot run here"; exit 2; }

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

fails=0
note() { printf '  %-58s %s\n' "$1" "$2"; [ "$2" = FAIL ] && fails=$((fails + 1)); return 0; }

cat > "$work/t.cpp" <<'EOF'
#include "video_watchdog.h"
#include <stdio.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>

static int fails = 0;
static void check(int ok, const char *what)
{
	printf("  %-58s %s\n", what, ok ? "OK" : "FAIL");
	if (!ok) fails++;
}

#define MAX 10

int main(int argc, char **argv)
{
	video_wd_t wd = {0, 0};
	video_wd_verdict_t v = VIDEO_WD_IDLE;
	int i, rebooted;

	// Switched on, no server: no heartbeat ever comes. The vendor loop
	// rebooted on the eleventh second.
	rebooted = 0;
	for (i = 0; i < 600; i++)
		if (video_wd_step(&wd, 1, 0, MAX) == VIDEO_WD_REBOOT) rebooted = 1;
	check(!rebooted, "on, no heartbeat for ten minutes: no reboot");
	check(!wd.armed, "on, no heartbeat: not armed");

	// The server starts and feeds.
	v = video_wd_step(&wd, 1, 1, MAX);
	check(v == VIDEO_WD_FED && wd.armed, "first heartbeat arms it");

	// Then it stops feeding: hung, crashed or stopped.
	for (i = 0; i < MAX; i++)
		v = video_wd_step(&wd, 1, 0, MAX);
	check(v == VIDEO_WD_MISSED, "armed, ten seconds missed: not yet");
	v = video_wd_step(&wd, 1, 0, MAX);
	check(v == VIDEO_WD_REBOOT, "armed, eleven seconds missed: reboot");

	// A heartbeat in time resets the count.
	wd.armed = 0; wd.misses = 0;
	video_wd_step(&wd, 1, 1, MAX);
	for (i = 0; i < 1000; i++) {
		v = video_wd_step(&wd, 1, (i % 5) == 4, MAX);
		if (v == VIDEO_WD_REBOOT) break;
	}
	check(v != VIDEO_WD_REBOOT, "a heartbeat every five seconds: no reboot");

	// Switching it off disarms; switching it on again waits for a heartbeat.
	video_wd_step(&wd, 1, 0, MAX);
	v = video_wd_step(&wd, 0, 0, MAX);
	check(v == VIDEO_WD_IDLE && !wd.armed && !wd.misses, "off: disarmed and the count cleared");
	rebooted = 0;
	for (i = 0; i < 60; i++)
		if (video_wd_step(&wd, 1, 0, MAX) == VIDEO_WD_REBOOT) rebooted = 1;
	check(!rebooted, "on again, no heartbeat: no reboot");

	// A heartbeat that could not be read (-1) counts as neither.
	wd.armed = 1; wd.misses = 3;
	v = video_wd_step(&wd, 1, -1, MAX);
	check(wd.misses == 3 && wd.armed, "unreadable heartbeat: count unchanged");

	// The count does not wrap back under the limit.
	wd.armed = 1; wd.misses = 0;
	for (i = 0; i < 400; i++)
		v = video_wd_step(&wd, 1, 0, MAX);
	check(v == VIDEO_WD_REBOOT, "the count saturates rather than wrapping");

	// Only a regular file at /tmp/watchdog is the vendor's switch.
	char file[512], dir[512];
	snprintf(file, sizeof(file), "%s/file", argv[1]);
	snprintf(dir, sizeof(dir), "%s/dir", argv[1]);
	FILE *f = fopen(file, "w"); if (f) fclose(f);
	mkdir(dir, 0755);
	check(video_wd_is_regular_file(file), "a regular file is a switch");
	check(!video_wd_is_regular_file(dir), "a directory is not (the daemon's log dir)");
	snprintf(file, sizeof(file), "%s/none", argv[1]);
	check(!video_wd_is_regular_file(file), "a missing path is not");

	(void)argc;
	return fails ? 1 : 0;
}
EOF

echo "===== the decisions ====="
if "$CXX" -std=c++11 -Wall -Werror -I"$MAIN/lib/system_state" "$work/t.cpp" "$SRC" -o "$work/t" 2> "$work/cc.log"; then
    "$work/t" "$work" || fails=$((fails + 1))
else
    cat "$work/cc.log"
    note "video_watchdog.cpp builds on the host" FAIL
fi

echo
echo "===== kvm_system uses them ====="
grep -q 'video_wd_step(' "$LOOP" \
    && note "the state loop decides through video_wd_step" OK \
    || note "the state loop does not use video_wd_step" FAIL
grep -q 'video_wd_is_regular_file(watchdog_temp_path)' "$STATE" \
    && note "/tmp/watchdog must be a regular file" OK \
    || note "/tmp/watchdog is checked some other way" FAIL
if grep -q 'access(watchdog_temp_path' "$STATE" && \
   sed -n '/watchdog_sf_is_open/,/^}/p' "$STATE" | grep -q 'access(watchdog_temp_path'; then
    note "watchdog_sf_is_open still takes any /tmp/watchdog" FAIL
else
    note "watchdog_sf_is_open ignores a directory" OK
fi

echo
if [ "$fails" -eq 0 ]; then
    echo "all cases passed"
else
    echo "$fails case(s) FAILED"
    exit 1
fi
