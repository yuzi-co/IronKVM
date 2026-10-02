/*
 * libkvm.so for the SG2002 on the mainline kernel.
 *
 * Implements the kvm_vision.h ABI that NanoKVM-Server links against, over
 * V4L2 instead of Sophgo's MPI, so the server runs unchanged on either
 * kernel. See README.md beside this file for what is implemented and what is
 * not, and kvmv_pipeline.h for the pipeline itself.
 */
#define _GNU_SOURCE

#include <errno.h>
#include <fcntl.h>
#include <pthread.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/ioctl.h>
#include <sys/stat.h>
#include <time.h>
#include <unistd.h>
#include <linux/v4l2-subdev.h>

/*
 * The server's header declares set_h264_fps and set_capture_fps weak, so a
 * server binary starts against a library that predates them. Including it
 * as it is would make the definitions below weak as well. Rename the two
 * declarations out of the way; the definitions further down are ordinary
 * strong symbols, as in the vendor library.
 */
#define set_h264_fps kvmv_header_set_h264_fps
#define set_capture_fps kvmv_header_set_capture_fps
/* Everything the server calls is exported; the rest of the library is
 * hidden by -fvisibility=hidden. */
#pragma GCC visibility push(default)
#include "kvm_vision.h"
#pragma GCC visibility pop
#undef set_h264_fps
#undef set_capture_fps
#pragma GCC visibility push(default)
void set_h264_fps(uint8_t _fps);
void set_capture_fps(uint8_t _fps);
#pragma GCC visibility pop

#include "kvmv_annexb.h"
#include "kvmv_pipeline.h"
#include "kvmv_policy.h"
#include "kvmv_slots.h"

#define KVMV_CODEC_MJPEG 0
#define KVMV_CODEC_H264 1
#define KVMV_CODEC_H265 2

/* Return codes the vendor library documents and the server reports. */
#define KVMV_RET_RETRIEVING (-5) /* "Retrieving image, please wait" */
#define KVMV_RET_CHANGING (-4) /* "Modifying image resolution, please wait" */

#define KVMV_LOCK_TIMEOUT_S 1
#define KVMV_FRAME_TIMEOUT_MS 1000U
#define KVMV_NO_FRAME_LIMIT 3U
#define KVMV_KEY_ATTEMPTS 4
#define KVMV_KEY_DROP_LIMIT 60 /* delta frames dropped waiting for an IDR */
#define KVMV_DEFAULT_IDLE_MS 10000U
#define KVMV_RATE_WINDOW_MS 10000U

#define KVMV_STATE_DIR "/tmp/kvm"
#define KVMV_STATE_FILE "/tmp/kvm/state"
#define KVMV_WIDTH_FILE "/kvmapp/kvm/width"
#define KVMV_HEIGHT_FILE "/kvmapp/kvm/height"
#define KVMV_WATCHDOG_MODE "/etc/kvm/watchdog"
#define KVMV_WATCHDOG_TEMP "/tmp/watchdog"
#define KVMV_WATCHDOG_FILE "/tmp/nanokvm_wd"

/* ---- state --------------------------------------------------------- */

/* Held across every pipeline operation. kvmv_read_* takes it with a timeout,
 * as the vendor library takes vi_mutex. */
static pthread_mutex_t pipe_lock = PTHREAD_MUTEX_INITIALIZER;
/* Orders updates of the published HDMI state. */
static pthread_mutex_t signal_lock = PTHREAD_MUTEX_INITIALIZER;

/* Everything below up to the atomics is guarded by pipe_lock. */
static struct kvmv_pipe pipe_state = {
	.cap_fd = -1, .vpss_fd = -1, .enc_fd = -1, .heap_fd = -1,
};
static struct kvmv_devices devices;
static int devices_found;
static struct kvmv_ps_cache ps_cache;
static unsigned int pipe_width, pipe_height; /* the size the pipe was built for */
static uint32_t pipe_bitrate;
static int pipe_gop, pipe_fps;
static int need_key; /* 0, or 1 plus the delta frames dropped waiting for an IDR */
static unsigned int no_frame_count;
static uint64_t next_start_ms;
static int last_start_result;
static struct kvmv_rate rate;
static int overshoot_warned;
static char last_error[256];

static struct kvmv_slots slots;

/* Shared with the setters and the monitor thread: __atomic only. */
static int debug_enabled;
static int gop_setting = KVMV_DEFAULT_GOP;
static int fps_setting = KVMV_DEFAULT_FPS;
static int capture_fps_setting = KVMV_DEFAULT_FPS;
static int frame_detect_setting;
static int venc_auto_recyc_setting;
static int force_key;
static int pipe_running;
static uint64_t last_read_ms;
static int capture_enabled; /* kvmv_hdmi_control; 0 until the server enables it, as in the vendor library */
static int capture_stopped; /* kvmv_hdmi_control(0) until kvmv_hdmi_control(1) */
static int source_locked;
static int signal_active;
static unsigned int published_width, published_height;

static pthread_t monitor_thread;
static int monitor_valid;
static int monitor_exit;

/* ---- helpers ------------------------------------------------------- */

static uint64_t now_ms(void)
{
	struct timespec ts;

	clock_gettime(CLOCK_MONOTONIC, &ts);
	return (uint64_t)ts.tv_sec * 1000U + (uint64_t)ts.tv_nsec / 1000000U;
}

static void log_msg(const char *fmt, ...) __attribute__((format(printf, 1, 2)));
static void log_msg(const char *fmt, ...)
{
	va_list args;

	fputs("[kvmv-v4l2] ", stderr);
	va_start(args, fmt);
	vfprintf(stderr, fmt, args);
	va_end(args);
	fputc('\n', stderr);
}

#define DBG(...)                                                          \
	do {                                                              \
		if (__atomic_load_n(&debug_enabled, __ATOMIC_RELAXED))   \
			log_msg(__VA_ARGS__);                             \
	} while (0)

/* Log an error once until a different one replaces it, so a source that
 * stays away does not fill the log at the stream's frame rate. Caller holds
 * pipe_lock. */
static void log_error_once(const char *what, const char *detail)
{
	char message[sizeof(last_error)];

	snprintf(message, sizeof(message), "%s: %s", what, detail);
	if (strcmp(message, last_error) != 0) {
		memcpy(last_error, message, sizeof(last_error));
		log_msg("%s", message);
	}
}

static unsigned int env_uint(const char *name, unsigned int fallback)
{
	const char *value = getenv(name);
	char *end;
	unsigned long parsed;

	if (value == NULL || *value == 0)
		return fallback;
	parsed = strtoul(value, &end, 10);
	if (*end != 0 || parsed > 1000000UL)
		return fallback;
	return (unsigned int)parsed;
}

static void write_small_file(const char *path, unsigned int value)
{
	char text[16];
	int length = snprintf(text, sizeof(text), "%u\n", value);
	int fd = open(path, O_WRONLY | O_CREAT | O_TRUNC | O_CLOEXEC, 0644);

	if (fd < 0)
		return;
	if (write(fd, text, (size_t)length) != length)
		DBG("short write to %s", path);
	close(fd);
}

/* /tmp/kvm/state, written the way the vendor library writes it: a temporary
 * file renamed over the old one, "1\n" or "0\n". kvm_system reads it. */
static void publish_state_file(int active)
{
	char temp_path[] = KVMV_STATE_DIR "/.state.tmp.XXXXXX";
	const char *text = active ? "1\n" : "0\n";
	int fd;

	if (mkdir(KVMV_STATE_DIR, 0755) != 0 && errno != EEXIST)
		return;
	fd = mkstemp(temp_path);
	if (fd < 0)
		return;
	if (fchmod(fd, 0644) != 0 || write(fd, text, 2) != 2 || fsync(fd) != 0) {
		close(fd);
		unlink(temp_path);
		return;
	}
	close(fd);
	if (rename(temp_path, KVMV_STATE_FILE) != 0)
		unlink(temp_path);
}

/* The input resolution the server reports through GetInputResolution. */
static void publish_resolution(unsigned int width, unsigned int height)
{
	struct stat st;

	if (width == 0 || height == 0)
		return;
	if (__atomic_load_n(&published_width, __ATOMIC_RELAXED) == width &&
	    __atomic_load_n(&published_height, __ATOMIC_RELAXED) == height)
		return;
	__atomic_store_n(&published_width, width, __ATOMIC_RELAXED);
	__atomic_store_n(&published_height, height, __ATOMIC_RELAXED);
	log_msg("HDMI input %ux%u", width, height);
	if (stat("/kvmapp/kvm", &st) != 0 || !S_ISDIR(st.st_mode))
		return;
	write_small_file(KVMV_WIDTH_FILE, width);
	write_small_file(KVMV_HEIGHT_FILE, height);
}

static void update_signal(void)
{
	int next, previous;

	pthread_mutex_lock(&signal_lock);
	next = __atomic_load_n(&capture_enabled, __ATOMIC_ACQUIRE) &&
	       __atomic_load_n(&source_locked, __ATOMIC_ACQUIRE);
	previous = __atomic_exchange_n(&signal_active, next, __ATOMIC_ACQ_REL);
	if (previous != next) {
		publish_state_file(next);
		DBG("HDMI signal %s", next ? "active" : "inactive");
	}
	pthread_mutex_unlock(&signal_lock);
}

static void set_source_locked(int locked)
{
	__atomic_store_n(&source_locked, locked ? 1 : 0, __ATOMIC_RELEASE);
	update_signal();
}

static void set_capture_enabled(int enabled)
{
	pthread_mutex_lock(&signal_lock);
	__atomic_store_n(&capture_enabled, enabled ? 1 : 0, __ATOMIC_RELEASE);
	__atomic_store_n(&signal_active, 0, __ATOMIC_RELEASE);
	publish_state_file(0);
	pthread_mutex_unlock(&signal_lock);
}

/* ---- pipeline lifecycle (pipe_lock held) --------------------------- */

static void pipe_down(const char *why)
{
	if (!pipe_state.running && pipe_state.cap_fd < 0 && pipe_state.enc_fd < 0)
		return;
	DBG("pipeline down: %s", why);
	kvmv_pipe_stop(&pipe_state);
	__atomic_store_n(&pipe_running, 0, __ATOMIC_RELEASE);
}

static int start_failed(int result, unsigned int retry_ms)
{
	last_start_result = result;
	next_start_ms = now_ms() + retry_ms;
	return result;
}

static int query_receiver(unsigned int need_width, unsigned int need_height)
{
	struct v4l2_dv_timings timings;
	enum kvmv_signal signal;
	int fd;

	if (!devices.subdev[0])
		return 0;
	fd = open(devices.subdev, O_RDWR | O_NONBLOCK | O_CLOEXEC);
	if (fd < 0)
		fd = open(devices.subdev, O_RDONLY | O_NONBLOCK | O_CLOEXEC);
	if (fd < 0)
		return 0;
	signal = kvmv_query_signal(fd, need_width, need_height, &timings);
	if (signal == KVMV_SIGNAL_OK) {
		/* Make the queried mode the active one. The capture driver's
		 * STREAMON refreshes it too; a read-only node refuses this, and
		 * that is fine. */
		ioctl(fd, VIDIOC_SUBDEV_S_DV_TIMINGS, &timings);
	}
	close(fd);

	if (signal != KVMV_SIGNAL_UNKNOWN)
		set_source_locked(kvmv_signal_present(signal));
	if (kvmv_signal_present(signal))
		publish_resolution(timings.bt.width, timings.bt.height);
	if (signal == KVMV_SIGNAL_UNSUPPORTED)
		log_error_once("HDMI source mode not supported by the capture node",
			       "set the host to 1920x1080 progressive");
	return kvmv_signal_result(signal);
}

/* The capture node's fixed frame size, to check the source against. */
static void capture_frame_size(unsigned int *width, unsigned int *height)
{
	struct v4l2_format format;
	int fd = open(devices.capture, O_RDWR | O_NONBLOCK | O_CLOEXEC);

	*width = 0;
	*height = 0;
	if (fd < 0)
		return;
	memset(&format, 0, sizeof(format));
	format.type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
	if (ioctl(fd, VIDIOC_G_FMT, &format) == 0) {
		*width = format.fmt.pix.width;
		*height = format.fmt.pix.height;
	}
	close(fd);
}

static void report_applied(uint32_t bitrate_bps, int gop, int fps)
{
	const struct kvmv_applied *a = &pipe_state.applied;

	log_msg("encoder holds bitrate %lld bit/s (asked %u), gop %d (asked %d), %d/%d fps (asked %d)",
		(long long)a->bitrate_bps, bitrate_bps, a->gop, gop,
		a->fps_numerator, a->fps_denominator, fps);
	if (a->bitrate_bps >= 0 && a->bitrate_bps != (int64_t)bitrate_bps)
		log_msg("encoder changed the bitrate it was given");
}

static int pipe_up(unsigned int width, unsigned int height, uint32_t bitrate_bps)
{
	struct kvmv_pipe_cfg cfg;
	unsigned int need_width, need_height;
	char missing[128];
	int result;

	if (next_start_ms && now_ms() < next_start_ms)
		return last_start_result;

	if (!devices_found) {
		if (kvmv_find_devices(&devices, missing, sizeof(missing))) {
			log_error_once("V4L2 nodes not found", missing);
			return start_failed(IMG_VENC_ERROR, 2000);
		}
		devices_found = 1;
		log_msg("capture %s, scaler %s, encoder %s, receiver %s",
			devices.capture, devices.scaler, devices.encoder,
			devices.subdev[0] ? devices.subdev : "(no subdevice node)");
	}

	capture_frame_size(&need_width, &need_height);
	result = query_receiver(need_width, need_height);
	if (result != 0)
		return start_failed(result, 500);

	memset(&cfg, 0, sizeof(cfg));
	cfg.req_width = width;
	cfg.req_height = height;
	cfg.bitrate_bps = bitrate_bps;
	cfg.gop = (unsigned int)__atomic_load_n(&gop_setting, __ATOMIC_ACQUIRE);
	cfg.fps = (unsigned int)__atomic_load_n(&fps_setting, __ATOMIC_ACQUIRE);
	cfg.capture_buffers = env_uint("KVMV_CAPTURE_BUFFERS", 2);
	cfg.mid_buffers = env_uint("KVMV_MID_BUFFERS", 2);
	cfg.bitstream_buffers = env_uint("KVMV_BITSTREAM_BUFFERS", 3);

	if (kvmv_pipe_start(&pipe_state, &devices, &cfg)) {
		int err = errno;

		log_error_once("pipeline start failed", pipe_state.error);
		if (err == ENOENT || err == ENODEV || err == ENXIO)
			devices_found = 0;
		/* The capture link refuses a source format it cannot take with
		 * EPIPE at STREAMON. */
		if (err == EPIPE)
			return start_failed(-6, 1000);
		return start_failed(IMG_VENC_ERROR, 1000);
	}

	pipe_width = width;
	pipe_height = height;
	pipe_bitrate = bitrate_bps;
	pipe_gop = (int)cfg.gop;
	pipe_fps = (int)cfg.fps;
	next_start_ms = 0;
	no_frame_count = 0;
	overshoot_warned = 0;
	last_error[0] = 0;
	kvmv_ps_cache_reset(&ps_cache);
	kvmv_rate_reset(&rate);
	/* The first picture after a build must be decodable from cold. */
	need_key = 1;
	kvmv_pipe_force_key(&pipe_state);
	__atomic_store_n(&force_key, 0, __ATOMIC_RELEASE);
	__atomic_store_n(&pipe_running, 1, __ATOMIC_RELEASE);
	set_source_locked(1);
	publish_resolution(pipe_state.plan.src_width, pipe_state.plan.src_height);

	log_msg("pipeline %ux%u -> %ux%u H.264",
		pipe_state.plan.src_width, pipe_state.plan.src_height,
		pipe_state.plan.out_width, pipe_state.plan.out_height);
	report_applied(bitrate_bps, pipe_gop, pipe_fps);
	return 0;
}

/* Apply settings that changed since the pipe was built. A change the encoder
 * refuses at runtime rebuilds the pipeline, which is what the vendor library
 * does for every one of them. */
static int apply_settings(unsigned int width, unsigned int height,
			  uint32_t bitrate_bps)
{
	int gop = __atomic_load_n(&gop_setting, __ATOMIC_ACQUIRE);
	int fps = __atomic_load_n(&fps_setting, __ATOMIC_ACQUIRE);
	int rebuild = 0;
	int changed = 0;

	if (gop != pipe_gop) {
		if (kvmv_pipe_set_gop(&pipe_state, (unsigned int)gop))
			rebuild = 1;
		pipe_gop = gop;
		changed = 1;
	}
	if (!rebuild && fps != pipe_fps) {
		if (kvmv_pipe_set_fps(&pipe_state, (unsigned int)fps))
			rebuild = 1;
		pipe_fps = fps;
		changed = 1;
	}
	if (!rebuild && bitrate_bps != pipe_bitrate) {
		if (kvmv_pipe_set_bitrate(&pipe_state, bitrate_bps))
			rebuild = 1;
		pipe_bitrate = bitrate_bps;
		kvmv_rate_reset(&rate);
		overshoot_warned = 0;
		changed = 1;
	}
	if (!rebuild && __atomic_exchange_n(&force_key, 0, __ATOMIC_ACQ_REL)) {
		if (kvmv_pipe_force_key(&pipe_state))
			rebuild = 1;
	}
	if (rebuild) {
		DBG("rebuilding: %s", pipe_state.error);
		pipe_down("setting refused at runtime");
		return pipe_up(width, height, bitrate_bps);
	}
	if (changed)
		report_applied(bitrate_bps, gop, fps);
	return 0;
}

static void watch_rate(size_t bytes)
{
	unsigned int kbps, fps_x10;
	unsigned int target = pipe_bitrate / 1000U;

	if (!kvmv_rate_add(&rate, now_ms(), bytes, KVMV_RATE_WINDOW_MS, &kbps,
			   &fps_x10))
		return;
	DBG("output %u kbit/s at %u.%u fps, target %u kbit/s", kbps,
	    fps_x10 / 10, fps_x10 % 10, target);
	if (!overshoot_warned && target && kbps > target * 2) {
		overshoot_warned = 1;
		log_msg("output runs at %u kbit/s against a %u kbit/s target (%u.%u fps); see ironkvm-dist#35",
			kbps, target, fps_x10 / 10, fps_x10 % 10);
	}
}

static int lock_pipe(void)
{
	struct timespec ts;

	clock_gettime(CLOCK_REALTIME, &ts);
	ts.tv_sec += KVMV_LOCK_TIMEOUT_S;
	return pthread_mutex_timedlock(&pipe_lock, &ts);
}

static int read_video_locked(unsigned int width, unsigned int height,
			     uint32_t bitrate_bps, uint8_t **data,
			     uint32_t *size)
{
	int attempt;
	int result;

	if (pipe_state.running && (width != pipe_width || height != pipe_height))
		pipe_down("output size changed");
	if (!pipe_state.running) {
		result = pipe_up(width, height, bitrate_bps);
		if (result != 0)
			return result;
	} else {
		result = apply_settings(width, height, bitrate_bps);
		if (result != 0)
			return result;
	}

	for (attempt = 0; attempt < KVMV_KEY_ATTEMPTS; attempt++) {
		struct kvmv_encoded encoded;
		struct kvmv_au_info info;
		struct kvmv_slot *slot;
		uint8_t prefix[2 * KVMV_PS_MAX];
		size_t prefix_len;
		int type;

		switch (kvmv_pipe_encode(&pipe_state, KVMV_FRAME_TIMEOUT_MS, &encoded)) {
		case KVMV_PIPE_OK:
			break;
		case KVMV_PIPE_NO_FRAME:
			DBG("%s", pipe_state.error);
			if (++no_frame_count >= KVMV_NO_FRAME_LIMIT) {
				/* Rebuild from the receiver's answer next time. */
				set_source_locked(0);
				pipe_down("capture delivers nothing");
			}
			return IMG_NOT_EXIST;
		case KVMV_PIPE_SOURCE_CHANGED:
			log_msg("HDMI source changed (%s), rebuilding", pipe_state.error);
			pipe_down("source changed");
			return KVMV_RET_CHANGING;
		case KVMV_PIPE_ERROR:
		default:
			log_error_once("frame failed", pipe_state.error);
			pipe_down("frame failed");
			/* Rebuild after a pause rather than on every read: a
			 * build costs a priming encode and a burst of ioctls. */
			return start_failed(IMG_VENC_ERROR, 500);
		}
		no_frame_count = 0;

		kvmv_au_inspect(encoded.data, encoded.size, &info);
		kvmv_ps_cache_update(&ps_cache, encoded.data, encoded.size);
		type = kvmv_au_type(&info);
		if (type != IMG_H264_TYPE_IF && type != IMG_H264_TYPE_PF) {
			DBG("access unit with no picture (%u NALs), skipped", info.nal_count);
			kvmv_pipe_release(&pipe_state, &encoded);
			continue;
		}
		if (need_key && type != IMG_H264_TYPE_IF) {
			/* A stream that starts on a delta frame shows garbage
			 * until the next keyframe. Ask again and drop this one,
			 * but not for ever: an encoder that never produces an
			 * IDR on request would otherwise stop the stream. */
			if (++need_key > KVMV_KEY_DROP_LIMIT) {
				log_msg("no IDR after %d forced keyframe requests; streaming without one",
					KVMV_KEY_DROP_LIMIT);
				need_key = 0;
			} else {
				DBG("delta frame before the first keyframe, dropped");
				kvmv_pipe_release(&pipe_state, &encoded);
				kvmv_pipe_force_key(&pipe_state);
				continue;
			}
		}

		prefix_len = kvmv_ps_cache_prefix(&ps_cache, &info, prefix,
						  sizeof(prefix));
		slot = kvmv_slot_claim(&slots, (uint32_t)(prefix_len + encoded.size));
		if (slot == NULL) {
			kvmv_pipe_release(&pipe_state, &encoded);
			return IMG_BUFFER_FULL;
		}
		memcpy(slot->data, prefix, prefix_len);
		memcpy(slot->data + prefix_len, encoded.data, encoded.size);
		kvmv_pipe_release(&pipe_state, &encoded);
		slot->size = (uint32_t)(prefix_len + encoded.size);
		slot->type = (uint8_t)type;
		need_key = 0;
		set_source_locked(1);
		watch_rate(slot->size);

		*data = slot->data;
		*size = slot->size;
		return type;
	}
	return IMG_NOT_EXIST;
}

static int read_frame(uint16_t width, uint16_t height, uint8_t codec,
		      int bitrate_kbps, uint8_t **data, uint32_t *size)
{
	static int mjpeg_warned, hevc_warned;
	int result;

	if (data == NULL || size == NULL)
		return IMG_NOT_EXIST;
	*data = NULL;
	*size = 0;

	if (codec == KVMV_CODEC_MJPEG) {
		if (!__atomic_exchange_n(&mjpeg_warned, 1, __ATOMIC_ACQ_REL))
			log_msg("MJPEG is not available on the mainline kernel: there is no driver for the SG2002 JPEG encoder (ironkvm-dist#36); use H.264");
		return IMG_VENC_ERROR;
	}
	if (codec != KVMV_CODEC_H264) {
		if (!__atomic_exchange_n(&hevc_warned, 1, __ATOMIC_ACQ_REL))
			log_msg("codec %u is not available: the Coda980 encodes H.264 only", codec);
		return IMG_VENC_ERROR;
	}

	if (lock_pipe() != 0)
		return KVMV_RET_RETRIEVING;
	__atomic_store_n(&last_read_ms, now_ms(), __ATOMIC_RELEASE);
	if (__atomic_load_n(&capture_stopped, __ATOMIC_ACQUIRE)) {
		pthread_mutex_unlock(&pipe_lock);
		return IMG_NOT_EXIST;
	}
	result = read_video_locked(width, height, kvmv_kbps_to_bps(bitrate_kbps),
				   data, size);
	pthread_mutex_unlock(&pipe_lock);
	return result;
}

/* ---- monitor thread ------------------------------------------------ */

static int watchdog_enabled(void)
{
	return access(KVMV_WATCHDOG_MODE, F_OK) == 0 ||
	       access(KVMV_WATCHDOG_TEMP, F_OK) == 0;
}

static void feed_watchdog(void)
{
	int fd = open(KVMV_WATCHDOG_FILE, O_WRONLY | O_CREAT | O_TRUNC | O_CLOEXEC, 0644);

	if (fd >= 0)
		close(fd);
}

/*
 * Every half second: feed the watchdog file, as the vendor library's thread
 * does. Every second: while no pipeline runs, ask the receiver whether a
 * source is there, so HasHDMISignal and /tmp/kvm/state stay true without a
 * viewer; and take an unread pipeline down, so an idle board is not running
 * capture DMA for nobody.
 */
static void *monitor_main(void *arg)
{
	char subdev[KVMV_PATH_MAX] = { 0 };
	unsigned int idle_ms = env_uint("KVMV_IDLE_MS", KVMV_DEFAULT_IDLE_MS);
	unsigned int tick = 0;
	int fd = -1;

	(void)arg;
	while (!__atomic_load_n(&monitor_exit, __ATOMIC_ACQUIRE)) {
		struct timespec step = { 0, 100 * 1000 * 1000 };
		int i;

		for (i = 0; i < 5 && !__atomic_load_n(&monitor_exit, __ATOMIC_ACQUIRE); i++)
			nanosleep(&step, NULL);
		if (__atomic_load_n(&monitor_exit, __ATOMIC_ACQUIRE))
			break;
		tick++;

		if (watchdog_enabled())
			feed_watchdog();
		if (tick % 2)
			continue;

		if (!__atomic_load_n(&pipe_running, __ATOMIC_ACQUIRE) &&
		    __atomic_load_n(&capture_enabled, __ATOMIC_ACQUIRE) &&
		    !__atomic_load_n(&capture_stopped, __ATOMIC_ACQUIRE)) {
			struct v4l2_dv_timings timings;
			enum kvmv_signal signal;

			if (fd < 0) {
				if (!subdev[0])
					kvmv_find_subdev(subdev, sizeof(subdev));
				if (subdev[0])
					fd = open(subdev, O_RDONLY | O_NONBLOCK | O_CLOEXEC);
			}
			signal = kvmv_query_signal(fd, 0, 0, &timings);
			if (signal != KVMV_SIGNAL_UNKNOWN) {
				set_source_locked(kvmv_signal_present(signal));
				if (kvmv_signal_present(signal))
					publish_resolution(timings.bt.width,
							   timings.bt.height);
			} else if (fd >= 0 && (errno == ENODEV || errno == EBADF)) {
				close(fd);
				fd = -1;
			}
		}

		if (idle_ms && __atomic_load_n(&pipe_running, __ATOMIC_ACQUIRE) &&
		    now_ms() - __atomic_load_n(&last_read_ms, __ATOMIC_ACQUIRE) > idle_ms &&
		    pthread_mutex_trylock(&pipe_lock) == 0) {
			if (now_ms() - __atomic_load_n(&last_read_ms, __ATOMIC_ACQUIRE) > idle_ms)
				pipe_down("no reads");
			pthread_mutex_unlock(&pipe_lock);
		}
	}
	if (fd >= 0)
		close(fd);
	return NULL;
}

/* ---- the ABI ------------------------------------------------------- */

void kvmv_init(uint8_t _debug_info_en)
{
	const char *env = getenv("KVMV_DEBUG");

	__atomic_store_n(&debug_enabled,
			 _debug_info_en != 0 || (env != NULL && *env && *env != '0'),
			 __ATOMIC_RELEASE);
	__atomic_store_n(&monitor_exit, 0, __ATOMIC_RELEASE);
	if (!monitor_valid) {
		if (pthread_create(&monitor_thread, NULL, monitor_main, NULL) == 0)
			monitor_valid = 1;
		else
			log_msg("could not start the monitor thread");
	}
	DBG("initialised");
}

void kvmv_deinit(void)
{
	__atomic_store_n(&monitor_exit, 1, __ATOMIC_RELEASE);
	if (monitor_valid) {
		pthread_join(monitor_thread, NULL);
		monitor_valid = 0;
	}
	pthread_mutex_lock(&pipe_lock);
	pipe_down("deinit");
	devices_found = 0;
	next_start_ms = 0;
	pthread_mutex_unlock(&pipe_lock);
	set_capture_enabled(0);
	__atomic_store_n(&source_locked, 0, __ATOMIC_RELEASE);
	kvmv_slots_free_all(&slots);
}

void set_venc_auto_recyc(uint8_t _enable)
{
	/* The vendor library used this to free the other encoder on a codec
	 * switch. There is one encoder here; kept for the ABI. */
	__atomic_store_n(&venc_auto_recyc_setting, _enable ? 1 : 0, __ATOMIC_RELAXED);
}

int kvmv_read_img(uint16_t _width, uint16_t _height, uint8_t _type,
		  uint16_t _qlty, uint8_t **_pp_kvm_data,
		  uint32_t *_p_kvmv_data_size)
{
	return read_frame(_width, _height, _type, _qlty, _pp_kvm_data,
			  _p_kvmv_data_size);
}

int kvmv_read_video(uint16_t _width, uint16_t _height, uint8_t _codec,
		    uint16_t _bitrate, uint8_t _gop, uint8_t _fps,
		    uint8_t **_pp_kvm_data, uint32_t *_p_kvmv_data_size)
{
	/* 0 keeps what set_h264_gop and set_h264_fps last set. */
	if (_gop != 0)
		__atomic_store_n(&gop_setting,
				 kvmv_clamp(_gop, KVMV_GOP_MIN, KVMV_GOP_MAX),
				 __ATOMIC_RELEASE);
	if (_fps != 0)
		__atomic_store_n(&fps_setting,
				 kvmv_clamp(_fps, KVMV_FPS_MIN, KVMV_FPS_MAX),
				 __ATOMIC_RELEASE);
	if (_codec != KVMV_CODEC_H264 && _codec != KVMV_CODEC_H265) {
		if (_pp_kvm_data)
			*_pp_kvm_data = NULL;
		if (_p_kvmv_data_size)
			*_p_kvmv_data_size = 0;
		return IMG_VENC_ERROR;
	}
	return read_frame(_width, _height, _codec, _bitrate, _pp_kvm_data,
			  _p_kvmv_data_size);
}

int free_kvmv_data(uint8_t **_pp_kvm_data)
{
	if (_pp_kvm_data == NULL || *_pp_kvm_data == NULL)
		return IMG_NOT_EXIST;
	return kvmv_slot_release(&slots, *_pp_kvm_data);
}

void free_all_kvmv_data(void)
{
	kvmv_slots_free_all(&slots);
}

void set_h264_gop(uint8_t _gop)
{
	__atomic_store_n(&gop_setting, kvmv_clamp(_gop, KVMV_GOP_MIN, KVMV_GOP_MAX),
			 __ATOMIC_RELEASE);
	/* The vendor library rebuilds its encoder here, so the next frame is a
	 * keyframe, and the server relies on that when a stream starts. */
	__atomic_store_n(&force_key, 1, __ATOMIC_RELEASE);
	DBG("set_h264_gop %d", __atomic_load_n(&gop_setting, __ATOMIC_RELAXED));
}

void set_h264_fps(uint8_t _fps)
{
	int fps = kvmv_clamp(_fps, KVMV_FPS_MIN, KVMV_FPS_MAX);

	if (__atomic_exchange_n(&fps_setting, fps, __ATOMIC_ACQ_REL) != fps)
		DBG("set_h264_fps %d", fps);
}

void set_capture_fps(uint8_t _fps)
{
	/*
	 * Recorded only. The capture node has no frame interval control and
	 * runs at the source's rate, but nothing beyond its DMA happens to a
	 * frame nobody reads: each read takes the newest frame and the rest go
	 * straight back to the driver, unscaled and unencoded.
	 */
	__atomic_store_n(&capture_fps_setting,
			 kvmv_clamp(_fps, KVMV_FPS_MIN, KVMV_FPS_MAX),
			 __ATOMIC_RELAXED);
}

void set_frame_detact(uint8_t _frame_detact)
{
	/* Frame detection skips unchanged MJPEG frames. There is no MJPEG here. */
	__atomic_store_n(&frame_detect_setting, kvmv_clamp(_frame_detact, 0, 100),
			 __ATOMIC_RELAXED);
}

uint8_t kvmv_hdmi_control(uint8_t _en)
{
	/*
	 * The vendor library powers the receiver through a GPIO on the PCIe
	 * board and answers -1 everywhere else. Here capture is stopped and
	 * started in software on every board, and the receiver stays powered:
	 * nothing reads frames while it is off, the pipeline is torn down, and
	 * reads answer IMG_NOT_EXIST until it is enabled again.
	 */
	if (_en == 0) {
		set_capture_enabled(0);
		__atomic_store_n(&capture_stopped, 1, __ATOMIC_RELEASE);
		pthread_mutex_lock(&pipe_lock);
		pipe_down("HDMI capture disabled");
		pthread_mutex_unlock(&pipe_lock);
		return 0;
	}
	__atomic_store_n(&capture_stopped, 0, __ATOMIC_RELEASE);
	set_capture_enabled(1);
	update_signal();
	return 0;
}

uint8_t kvmv_hdmi_signal_active(void)
{
	return __atomic_load_n(&signal_active, __ATOMIC_ACQUIRE) ? 1 : 0;
}
