#ifndef VIDEO_WATCHDOG_H_
#define VIDEO_WATCHDOG_H_

// The video watchdog: kvm_system reboots the board when the server's video
// library stops touching /tmp/nanokvm_wd.
//
// It is on while /etc/kvm/watchdog exists (the owner's setting) or while
// /tmp/watchdog is a regular file (the vendor's temporary switch, which
// kvm_system makes during an application update and removes a minute later).
//
// Two things changed against the vendor loop (#59):
//
// - A directory at /tmp/watchdog does not count. The watchdog daemon that
//   S01hwdt starts made one at every boot (its default log directory is
//   /var/log/watchdog, and /var/log is a link to /tmp), which switched this on
//   on every board.
//
// - The watchdog arms on the first heartbeat after it is switched on, and
//   counts missed heartbeats only from then. kvm_system without a running
//   server (a board brought up by hand, a server stopped for maintenance before
//   it ever fed) no longer reboots ten seconds later. A server that fed and
//   then stopped, hung or crashed is still answered with a reboot after
//   VIDEO_WD_MISSES_MAX missed seconds, which is what the watchdog is for.
//
// No MaixCDK dependency, so a host test can build it alone
// (tools/kvm_system/test-video-watchdog.sh).

#include <stdint.h>

typedef struct {
	uint8_t armed;		// a heartbeat was seen since the watchdog was switched on
	uint8_t misses;		// seconds without a heartbeat since it armed
} video_wd_t;

typedef enum {
	VIDEO_WD_IDLE = 0,	// off, or on and waiting for the first heartbeat
	VIDEO_WD_FED,		// a heartbeat this second
	VIDEO_WD_MISSED,	// armed, no heartbeat this second, not yet too many
	VIDEO_WD_REBOOT,	// armed and too many seconds missed
} video_wd_verdict_t;

// One step, once a second. enabled: the watchdog is switched on. fed: a
// heartbeat arrived since the last step (1), none did (0), or it could not be
// told (-1, counted as neither).
video_wd_verdict_t video_wd_step(video_wd_t *wd, int enabled, int fed, uint8_t misses_max);

// 1 when the path exists and is a regular file.
int video_wd_is_regular_file(const char *path);

#endif
