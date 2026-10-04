#include "video_watchdog.h"

#include <sys/stat.h>

video_wd_verdict_t video_wd_step(video_wd_t *wd, int enabled, int fed, uint8_t misses_max)
{
	if (!enabled) {
		wd->armed = 0;
		wd->misses = 0;
		return VIDEO_WD_IDLE;
	}
	if (fed == 1) {
		wd->armed = 1;
		wd->misses = 0;
		return VIDEO_WD_FED;
	}
	if (!wd->armed || fed != 0)
		return VIDEO_WD_IDLE;
	if (wd->misses < 255)
		wd->misses++;
	return wd->misses > misses_max ? VIDEO_WD_REBOOT : VIDEO_WD_MISSED;
}

int video_wd_is_regular_file(const char *path)
{
	struct stat st;

	return stat(path, &st) == 0 && S_ISREG(st.st_mode);
}
