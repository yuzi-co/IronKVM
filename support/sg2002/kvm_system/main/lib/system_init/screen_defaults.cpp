#include "screen_defaults.h"

#include <ctype.h>
#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

// Larger than anything a setting holds. A file that does not fit is not one
// this code or the server wrote, and is treated as unusable.
#define SETTING_MAX	32

static const struct {
	const char *name;
	const char *value;
} screen_defaults[] = {
	{ "fps",  SCREEN_DEFAULT_FPS  },
	{ "qlty", SCREEN_DEFAULT_QLTY },
	{ "res",  SCREEN_DEFAULT_RES  },
	{ "type", SCREEN_DEFAULT_TYPE },
};

// The text without leading and trailing whitespace, into out. The server
// writes values bare and the old shell redirect wrote them with a newline;
// both are the same value.
static void trim(const char *text, char *out, size_t len)
{
	while(*text != '\0' && isspace((unsigned char)*text)) text++;
	size_t n = strlen(text);
	while(n > 0 && isspace((unsigned char)text[n - 1])) n--;
	if(n >= len) n = len - 1;
	memcpy(out, text, n);
	out[n] = '\0';
}

// A decimal number with nothing else in it, or -1.
static long number(const char *text)
{
	if(*text == '\0') return -1;
	for(const char *p = text; *p != '\0'; p++){
		if(!isdigit((unsigned char)*p)) return -1;
	}
	if(strlen(text) > 6) return -1;
	return strtol(text, NULL, 10);
}

int screen_setting_valid(const char *name, const char *text)
{
	char value[SETTING_MAX];
	trim(text, value, sizeof(value));

	if(strcmp(name, "type") == 0){
		// What the server writes, and the H.265 spelling a hand edit might use.
		// kvm_system's reader only looks at the first letter.
		return strcmp(value, "h264") == 0 || strcmp(value, "h265") == 0 ||
			strcmp(value, "mjpeg") == 0;
	}

	long n = number(value);
	if(strcmp(name, "res") == 0){
		// The heights the server's ResolutionMap knows; 0 is the source's own.
		return n == 0 || n == 480 || n == 600 || n == 720 || n == 1080;
	}
	if(strcmp(name, "fps") == 0){
		// The server clamps to 10..120 itself; any frame rate is a choice.
		return n >= 1 && n <= 1000;
	}
	if(strcmp(name, "qlty") == 0){
		// A JPEG quality up to 100 or a bitrate in kbit/s above it, in the
		// server's 16 bits.
		return n >= 1 && n <= 65535;
	}

	return 0;
}

// Reads the file at path into buf. Answers 1 when it held something that fits,
// 0 when it is missing, empty, unreadable or too large.
static int read_setting(const char *path, char *buf, size_t len)
{
	int fd = open(path, O_RDONLY);
	if(fd < 0) return 0;

	ssize_t n = read(fd, buf, len);
	close(fd);
	if(n <= 0 || (size_t)n >= len) return 0;

	buf[n] = '\0';
	return 1;
}

static int write_setting(const char *path, const char *value)
{
	// O_CREAT without O_NOFOLLOW: a link is followed, so a link to a shared
	// copy that does not exist yet makes the shared copy.
	int fd = open(path, O_WRONLY | O_CREAT | O_TRUNC, 0644);
	if(fd < 0) return -1;

	char line[SETTING_MAX];
	int len = snprintf(line, sizeof(line), "%s\n", value);
	int ok = write(fd, line, len) == len;
	if(close(fd) != 0) ok = 0;

	return ok ? 0 : -1;
}

int seed_screen_defaults(const char *dir)
{
	int written = 0;
	int failed = 0;

	for(size_t i = 0; i < sizeof(screen_defaults) / sizeof(screen_defaults[0]); i++){
		char path[256];
		char held[SETTING_MAX + 1];
		snprintf(path, sizeof(path), "%s/%s", dir, screen_defaults[i].name);

		if(read_setting(path, held, sizeof(held)) &&
		   screen_setting_valid(screen_defaults[i].name, held)){
			continue;
		}

		if(write_setting(path, screen_defaults[i].value) != 0){
			printf("screen defaults: could not write %s: %s\n", path, strerror(errno));
			failed = 1;
			continue;
		}
		printf("screen defaults: %s had no usable value, wrote %s\n",
		       path, screen_defaults[i].value);
		written++;
	}

	return failed ? -1 : written;
}
