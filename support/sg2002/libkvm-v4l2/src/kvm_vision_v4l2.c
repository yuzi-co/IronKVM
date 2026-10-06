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
#include "kvmv_jpeg.h"
#include "kvmv_hwjpeg.h"

#define KVMV_CODEC_MJPEG 0
#define KVMV_CODEC_H264 1
#define KVMV_CODEC_H265 2

/* Return codes the vendor library documents and the server reports. */
#define KVMV_RET_RETRIEVING (-5) /* "Retrieving image, please wait" */
#define KVMV_RET_CHANGING (-4) /* "Modifying image resolution, please wait" */

#define KVMV_LOCK_TIMEOUT_S 1
/* An MJPEG read may wait behind another one's encode, up to about 0.3 s at 1080p. */
#define KVMV_JPEG_LOCK_TIMEOUT_S 2
#define KVMV_FRAME_TIMEOUT_MS 1000U
/* The longest a read waits for a capture frame about to complete (see
 * pickup_wait_ms in kvmv_pipeline.c). 6 ms keeps an H.264 read, about 26 ms
 * at 1080p, inside a 30 fps frame time. */
#define KVMV_PICKUP_WAIT_MS 6U
#define KVMV_NO_FRAME_LIMIT 3U
#define KVMV_KEY_ATTEMPTS 4
#define KVMV_KEY_DROP_LIMIT 60 /* delta frames dropped waiting for an IDR */
#define KVMV_DEFAULT_IDLE_MS 10000U
/* How old a receiver answer from the monitor thread a build may use. It asks
 * every second while no pipeline runs. */
#define KVMV_RECEIVER_CACHE_MS 1500U
#define KVMV_RATE_WINDOW_MS 10000U
#define KVMV_FPS_WINDOW_MS 3000U
/* Room in front of a copied access unit for the parameter sets a keyframe
 * may need (H.265: VPS, SPS and PPS), so the unit is copied out of the
 * encoder once. */
#define KVMV_PREFIX_ROOM (3 * KVMV_PS_MAX)

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
/* One MJPEG picture at a time; taken before pipe_lock, never after. */
static pthread_mutex_t jpeg_lock = PTHREAD_MUTEX_INITIALIZER;
/* Guards claiming a slot: see claim_slot. */
static pthread_mutex_t slot_lock = PTHREAD_MUTEX_INITIALIZER;

/* Everything below up to the atomics is guarded by pipe_lock. */
static struct kvmv_pipe pipe_state = {
	.cap_fd = -1, .vpss_fd = -1, .enc_fd = -1, .heap_fd = -1,
	.ahead_mid = -1,
};
static struct kvmv_devices devices;
static int devices_found;
static struct kvmv_ps_cache ps_cache;
static unsigned int pipe_width, pipe_height; /* the size the pipe was built for */
/* The codec the pipe was built for, or is built for next by an MJPEG read. */
static enum kvmv_codec pipe_codec = KVMV_CODEC_KIND_H264;
static uint32_t pipe_bitrate;
/* Smoothed scale and encode times of this pipeline, for want_scale_ahead. */
static uint32_t scale_ewma_us, encode_ewma_us;
/* The last MJPEG or snapshot read, for want_scale_ahead. */
static uint64_t last_image_ms;
static int pipe_gop, pipe_fps;
static int need_key; /* 0, or 1 plus the delta frames dropped waiting for an IDR */
static unsigned int no_frame_count;
static uint64_t next_start_ms;
static int last_start_result;
static struct kvmv_rate rate;
static struct kvmv_rate fps_rate; /* a shorter window, for delivered_fps */
static int overshoot_warned;
static char last_error[256];
/*
 * H.265 at the capture's size straight from the capture buffer, without the
 * scaler (kvmv_pipe_cfg.direct, ironkvm-dist patch 0933): KVMV_HEVC_DIRECT=0
 * keeps the scaler. Off for good after a build that fails on it or a few
 * failed pictures in a row from it. Set once, under pipe_lock.
 */
#define KVMV_HEVC_DIRECT_DEFAULT 1U
#define KVMV_HEVC_DIRECT_FAILURE_LIMIT 3
static int hevc_direct_off = -1;
static int hevc_direct_failures;
static int hevc_direct_refused; /* logged that the encoder refused the format */

static struct kvmv_slots slots;
static struct kvmv_jpeg jpeg_state; /* guarded by jpeg_lock */
/* The JPEG unit, when the kernel has it. Guarded by pipe_lock: it encodes
 * from the snapshot buffer, which the pipeline owns. */
static struct kvmv_hwjpeg hw_jpeg = { .fd = -1, .out = { .fd = -1 }, .cap = { .fd = -1 } };
static int hw_jpeg_failures; /* consecutive; the unit is left alone at the limit */
static int hw_jpeg_announced;
static int hw_jpeg_cache_announced;
static int range_announced; /* 1 full, 2 limited */

/* Shared with the setters and the monitor thread: __atomic only. */
static int debug_enabled;
static int gop_setting = KVMV_DEFAULT_GOP;
static int fps_setting = KVMV_DEFAULT_FPS;
static int capture_fps_setting = KVMV_DEFAULT_FPS;
static int frame_detect_setting;
static int venc_auto_recyc_setting;
static int force_key;
/* Frames per second actually handed out, rounded; 0 until measured. Reset when
 * the asked rate changes. Written under pipe_lock, cleared by the setters. */
static int delivered_fps;
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

static uint64_t now_us(void)
{
	struct timespec ts;

	clock_gettime(CLOCK_MONOTONIC, &ts);
	return (uint64_t)ts.tv_sec * 1000000U + (uint64_t)ts.tv_nsec / 1000U;
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

static void request_key(void);
static unsigned int capture_width_known, capture_height_known;
static void receiver_remember(enum kvmv_signal signal,
			      const struct v4l2_dv_timings *timings);

/*
 * Tear the pipeline down. park keeps the capture node and its buffers for
 * the next build (kvmv_pipe_park): for a pipeline nobody reads, a codec
 * switch or a new output size, where the source is unchanged. Anything else
 * releases them too.
 */
static void pipe_down_keep(const char *why, int park)
{
	if (!park) {
		kvmv_pipe_unpark();
		/* The source may have changed: ask the nodes again. */
		capture_width_known = 0;
		capture_height_known = 0;
		receiver_remember(KVMV_SIGNAL_NONE, NULL);
	}
	if (!pipe_state.running && pipe_state.cap_fd < 0 && pipe_state.enc_fd < 0)
		return;
	DBG("pipeline down: %s%s", why, park ? " (capture buffers kept)" : "");
	if (park)
		kvmv_pipe_park(&pipe_state);
	else
		kvmv_pipe_stop(&pipe_state);
	/* The unit imports the snapshot buffer; let both go together. Closing
	 * the node also turns the unit's clocks off. */
	kvmv_hwjpeg_close(&hw_jpeg);
	__atomic_store_n(&pipe_running, 0, __ATOMIC_RELEASE);
}

static void pipe_down(const char *why)
{
	pipe_down_keep(why, 0);
}

static int start_failed(int result, unsigned int retry_ms)
{
	last_start_result = result;
	next_start_ms = now_ms() + retry_ms;
	return result;
}

/*
 * Ask the receiver for the source's mode. On a source the capture can take,
 * *width and *height are its size; they stay 0 when there is no receiver to
 * ask. The query also makes the mode the receiver's active format, which the
 * capture node follows (ironkvm-dist patch 0936).
 */
static int query_receiver(unsigned int *width, unsigned int *height)
{
	struct v4l2_dv_timings timings;
	enum kvmv_signal signal;
	int fd;

	*width = 0;
	*height = 0;
	if (!devices.subdev[0])
		return 0;
	fd = open(devices.subdev, O_RDWR | O_NONBLOCK | O_CLOEXEC);
	if (fd < 0)
		fd = open(devices.subdev, O_RDONLY | O_NONBLOCK | O_CLOEXEC);
	if (fd < 0)
		return 0;
	signal = kvmv_query_signal(fd, 0, 0, &timings);
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
		log_error_once("HDMI source mode not supported", "interlaced");
	if (signal == KVMV_SIGNAL_OK) {
		*width = timings.bt.width;
		*height = timings.bt.height;
	}
	return kvmv_signal_result(signal);
}

/* The capture node's frame size. Asked once (capture_width_known,
 * capture_height_known), and again after a pipe_down that does not park
 * (source changes) or when the source's size differs. */
static void capture_frame_size(unsigned int *width, unsigned int *height)
{
	struct v4l2_format format;
	int fd;

	*width = capture_width_known;
	*height = capture_height_known;
	if (*width && *height)
		return;
	fd = open(devices.capture, O_RDWR | O_NONBLOCK | O_CLOEXEC);
	if (fd < 0)
		return;
	memset(&format, 0, sizeof(format));
	format.type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
	if (ioctl(fd, VIDIOC_G_FMT, &format) == 0) {
		*width = format.fmt.pix.width;
		*height = format.fmt.pix.height;
		capture_width_known = *width;
		capture_height_known = *height;
	}
	close(fd);
}

/*
 * The monitor thread's last answer from the receiver, taken while no pipeline
 * ran: a source in a mode with a frame size, or receiver_seen_ms 0. A build
 * soon after uses it instead of asking again (KVMV_RECEIVER_CACHE_MS).
 */
static pthread_mutex_t receiver_lock = PTHREAD_MUTEX_INITIALIZER;
static struct v4l2_dv_timings receiver_timings;
static uint64_t receiver_seen_ms;

static void receiver_remember(enum kvmv_signal signal,
			      const struct v4l2_dv_timings *timings)
{
	pthread_mutex_lock(&receiver_lock);
	if (signal == KVMV_SIGNAL_OK) {
		receiver_timings = *timings;
		receiver_seen_ms = now_ms();
	} else {
		receiver_seen_ms = 0;
	}
	pthread_mutex_unlock(&receiver_lock);
}

/* 1 when the monitor saw, recently enough, a source in a progressive mode;
 * *width and *height are its size. */
static int receiver_recent(unsigned int *width, unsigned int *height)
{
	unsigned int max_age = env_uint("KVMV_RECEIVER_CACHE_MS", KVMV_RECEIVER_CACHE_MS);
	struct v4l2_dv_timings timings;
	int fresh;

	pthread_mutex_lock(&receiver_lock);
	fresh = kvmv_receiver_fresh(receiver_seen_ms, now_ms(), max_age);
	timings = receiver_timings;
	pthread_mutex_unlock(&receiver_lock);
	if (!fresh || kvmv_classify_timings(0, 0, &timings, 0, 0) != KVMV_SIGNAL_OK)
		return 0;
	set_source_locked(1);
	publish_resolution(timings.bt.width, timings.bt.height);
	*width = timings.bt.width;
	*height = timings.bt.height;
	return 1;
}

/*
 * Check the source's mode against what the capture can take and what the
 * capture node reports, and have them agree. A parked capture node keeps the
 * size it had when it was parked, so a different source size unparks it and
 * asks again. Returns 0, or the read result for a mode that cannot be taken.
 */
static int match_capture(unsigned int src_width, unsigned int src_height)
{
	unsigned int width, height;
	char detail[96];

	if (!src_width || !src_height)
		return 0; /* no receiver to ask: the capture node's size it is */
	if (!kvmv_source_fits(src_width, src_height)) {
		snprintf(detail, sizeof(detail), "%ux%u is larger than a capture buffer holds",
			 src_width, src_height);
		log_error_once("HDMI source mode not supported", detail);
		return -6;
	}
	capture_frame_size(&width, &height);
	if (width == src_width && height == src_height)
		return 0;
	/* Parked buffers pin the old size, and the remembered size is old. */
	kvmv_pipe_unpark();
	capture_width_known = 0;
	capture_height_known = 0;
	capture_frame_size(&width, &height);
	if (width == src_width && height == src_height) {
		log_msg("capture follows the source to %ux%u", width, height);
		return 0;
	}
	snprintf(detail, sizeof(detail),
		 "source %ux%u, capture node %ux%u (kernel without ironkvm-dist patch 0936)",
		 src_width, src_height, width, height);
	log_error_once("HDMI source mode not supported by the capture node", detail);
	return -6;
}

static void report_applied(uint32_t bitrate_bps, int gop, int fps)
{
	const struct kvmv_applied *a = &pipe_state.applied;

	log_msg("encoder holds bitrate %lld bit/s (asked %u), gop %d (asked %d), %d/%d fps (asked %d)",
		(long long)a->bitrate_bps, bitrate_bps, a->gop, gop,
		a->fps_numerator, a->fps_denominator, fps);
	if (a->bitrate_bps >= 0 && a->bitrate_bps != (int64_t)bitrate_bps)
		log_msg("encoder changed the bitrate it was given");
	log_msg("encoder QP range %d..%d (-1: no control)", a->min_qp, a->max_qp);
}

/* The rate to tell the encoder: see kvmv_encoder_fps. current is what it
 * holds now, 0 for a new pipeline. */
static int encoder_fps(int current)
{
	return kvmv_encoder_fps(__atomic_load_n(&fps_setting, __ATOMIC_ACQUIRE),
				__atomic_load_n(&capture_fps_setting, __ATOMIC_ACQUIRE),
				__atomic_load_n(&delivered_fps, __ATOMIC_ACQUIRE),
				current);
}

static const char *codec_name(enum kvmv_codec codec)
{
	return codec == KVMV_CODEC_KIND_HEVC ? "H.265" : "H.264";
}

static int find_devices_locked(void)
{
	char missing[128];

	if (devices_found)
		return 0;
	if (kvmv_find_devices(&devices, missing, sizeof(missing))) {
		log_error_once("V4L2 nodes not found", missing);
		return -1;
	}
	devices_found = 1;
	log_msg("capture %s, scaler %s, H.264 encoder %s, H.265 encoder %s, receiver %s",
		devices.capture, devices.scaler,
		devices.encoder[0] ? devices.encoder : "(none)",
		devices.encoder_hevc[0] ? devices.encoder_hevc : "(none)",
		devices.subdev[0] ? devices.subdev : "(no subdevice node)");
	return 0;
}

static int have_encoder(enum kvmv_codec codec)
{
	return codec == KVMV_CODEC_KIND_HEVC ? devices.encoder_hevc[0] != 0 :
					       devices.encoder[0] != 0;
}

static int pipe_up(unsigned int width, unsigned int height, uint32_t bitrate_bps)
{
	struct kvmv_pipe_cfg cfg;
	struct kvmv_qp_range qp;
	unsigned int src_width = 0, src_height = 0;
	enum kvmv_codec codec = pipe_codec;
	int hevc = codec == KVMV_CODEC_KIND_HEVC;
	uint64_t query_us;
	int receiver_cached;
	int result;

	if (next_start_ms && now_ms() < next_start_ms)
		return last_start_result;

	if (find_devices_locked())
		return start_failed(IMG_VENC_ERROR, 2000);
	/* An MJPEG read builds the pipe too; it takes whichever encoder exists. */
	if (!have_encoder(codec)) {
		codec = hevc ? KVMV_CODEC_KIND_H264 : KVMV_CODEC_KIND_HEVC;
		hevc = !hevc;
		pipe_codec = codec;
	}

	query_us = now_us();
	receiver_cached = receiver_recent(&src_width, &src_height);
	if (!receiver_cached) {
		result = query_receiver(&src_width, &src_height);
		if (result != 0)
			return start_failed(result, 500);
	}
	result = match_capture(src_width, src_height);
	if (result != 0)
		return start_failed(result, 1000);
	query_us = now_us() - query_us;

	memset(&cfg, 0, sizeof(cfg));
	cfg.codec = codec;
	cfg.req_width = width;
	cfg.req_height = height;
	cfg.bitrate_bps = bitrate_bps;
	cfg.gop = (unsigned int)__atomic_load_n(&gop_setting, __ATOMIC_ACQUIRE);
	cfg.fps = (unsigned int)encoder_fps(0);
	if (hevc) {
		kvmv_h265_qp_range(getenv("KVMV_H265_QP"), &qp);
		cfg.vbv_delay_ms = env_uint("KVMV_H265_VBV_DELAY_MS", KVMV_H265_VBV_DELAY_MS);
	} else {
		kvmv_h264_qp_range(getenv("KVMV_H264_QP"), &qp);
		cfg.vbv_delay_ms = env_uint("KVMV_H264_VBV_DELAY_MS", KVMV_H264_VBV_DELAY_MS);
	}
	cfg.min_qp = qp.min_qp;
	cfg.max_qp = qp.max_qp;
	cfg.capture_buffers = env_uint("KVMV_CAPTURE_BUFFERS", 3);
	cfg.pickup_wait_ms = env_uint("KVMV_PICKUP_WAIT_MS", KVMV_PICKUP_WAIT_MS);
	cfg.mid_buffers = env_uint("KVMV_MID_BUFFERS", 2);
	cfg.bitstream_buffers = env_uint("KVMV_BITSTREAM_BUFFERS", 3);
	cfg.early_capture = (int)env_uint("KVMV_EARLY_CAPTURE", 1);
	cfg.park_encoder = (int)env_uint("KVMV_PARK_ENCODER", 1);
	cfg.prime = kvmv_prime_mode(getenv("KVMV_PRIME"), hevc);
	if (hevc_direct_off < 0)
		hevc_direct_off = env_uint("KVMV_HEVC_DIRECT", KVMV_HEVC_DIRECT_DEFAULT) == 0;
	cfg.direct = hevc && !hevc_direct_off;

	if (kvmv_pipe_start(&pipe_state, &devices, &cfg)) {
		int err = errno;

		log_error_once("pipeline start failed", pipe_state.error);
		/* EINVAL: the capture buffers do not fit the encoder (no
		 * ironkvm-dist patch 0921), or a format was refused. */
		if (cfg.direct && err == EINVAL) {
			/* The next build goes through the scaler. */
			log_msg("H.265 from the capture buffer failed (%s); the scaler converts from now on",
				pipe_state.error);
			hevc_direct_off = 1;
		}
		if (err == ENOENT || err == ENODEV || err == ENXIO)
			devices_found = 0;
		/* The capture link refuses a source format that differs from
		 * the capture node's with EPIPE at STREAMON: the mode changed
		 * after the receiver was asked. Ask again on the next read. */
		if (err == EPIPE) {
			capture_width_known = 0;
			capture_height_known = 0;
			receiver_remember(KVMV_SIGNAL_NONE, NULL);
			kvmv_pipe_unpark();
			return start_failed(KVMV_RET_CHANGING, 500);
		}
		/* The other codec's encoder holds the codec SRAM. */
		if (err == EBUSY)
			log_msg("the %s encoder is busy: the other codec is streaming",
				codec_name(codec));
		return start_failed(IMG_VENC_ERROR, 1000);
	}

	DBG("pipeline start, ms: receiver %.1f%s, then open %.1f, early capture on %.1f, formats %.1f, middle buffers %.1f, bitstream buffers %.1f, encoder on %.1f, priming %.1f (mode %d), capture buffers %.1f, capture on %.1f, scaler on %.1f%s%s%s",
	    query_us / 1000.0, receiver_cached ? " (cached)" : "",
	    pipe_state.start_us[0] / 1000.0, pipe_state.start_us[1] / 1000.0,
	    pipe_state.start_us[2] / 1000.0, pipe_state.start_us[3] / 1000.0,
	    pipe_state.start_us[4] / 1000.0, pipe_state.start_us[5] / 1000.0,
	    pipe_state.start_us[6] / 1000.0, (int)cfg.prime,
	    pipe_state.start_us[7] / 1000.0,
	    pipe_state.start_us[8] / 1000.0, pipe_state.start_us[9] / 1000.0,
	    pipe_state.early ? " (capture started first)" : "",
	    pipe_state.enc_reused ? " (encoder kept)" : "",
	    pipe_state.cap_allocated ? " (capture buffers allocated before the encoder on)" : "");
	if (pipe_state.cap_allocated &&
	    pipe_state.cap_count < (cfg.capture_buffers ? cfg.capture_buffers : 2))
		log_msg("capture got %u of %u buffers: video_pool is short",
			pipe_state.cap_count, cfg.capture_buffers ? cfg.capture_buffers : 2);
	pipe_width = width;
	pipe_height = height;
	pipe_bitrate = bitrate_bps;
	pipe_gop = (int)cfg.gop;
	pipe_fps = (int)cfg.fps;
	next_start_ms = 0;
	no_frame_count = 0;
	overshoot_warned = 0;
	last_error[0] = 0;
	kvmv_ps_cache_reset_codec(&ps_cache, codec);
	kvmv_rate_reset(&rate);
	kvmv_rate_reset(&fps_rate);
	/* The first picture after a build must be decodable from cold. */
	need_key = 1;
	request_key();
	__atomic_store_n(&force_key, 0, __ATOMIC_RELEASE);
	__atomic_store_n(&pipe_running, 1, __ATOMIC_RELEASE);
	set_source_locked(1);
	publish_resolution(pipe_state.plan.src_width, pipe_state.plan.src_height);

	log_msg("pipeline %ux%u -> %ux%u %s%s",
		pipe_state.plan.src_width, pipe_state.plan.src_height,
		pipe_state.plan.out_width, pipe_state.plan.out_height,
		codec_name(codec),
		pipe_state.direct ? ", the encoder reads the capture buffer (no scaler)" : "");
	if (cfg.direct && !pipe_state.direct &&
	    pipe_state.plan.out_width == pipe_state.plan.src_width &&
	    pipe_state.plan.out_height == pipe_state.plan.src_height &&
	    !__atomic_exchange_n(&hevc_direct_refused, 1, __ATOMIC_RELAXED))
		log_msg("H.265: the encoder does not take the capture's %.4s (ironkvm-dist patch 0933); the scaler converts",
			(const char *)&pipe_state.cap_fmt.pixelformat);
	report_applied(bitrate_bps, pipe_gop, pipe_fps);
	return 0;
}

/*
 * Ask the encoder for a keyframe next. An encoder without the control (the
 * coda driver before ironkvm-dist patch 0912 creates none) produces IDRs only
 * at its GOP boundaries: drop delta frames until the next one instead. That
 * waits at most one GOP; a rebuild, which this used to do, primes the encoder
 * with a picture of its own and so always waits for a whole GOP.
 */
static int key_control_missing;

static void request_key(void)
{
	if (kvmv_pipe_force_key(&pipe_state) == 0)
		return;
	if (!key_control_missing) {
		key_control_missing = 1;
		log_msg("%s; keyframes wait for the GOP (ironkvm-dist#37, patch 0912)",
			pipe_state.error);
	}
	if (!need_key)
		need_key = 1;
}

/* Apply settings that changed since the pipe was built. A change the encoder
 * refuses at runtime rebuilds the pipeline, which is what the vendor library
 * does for every one of them. */
static int apply_settings(unsigned int width, unsigned int height,
			  uint32_t bitrate_bps)
{
	int gop = __atomic_load_n(&gop_setting, __ATOMIC_ACQUIRE);
	int fps = encoder_fps(pipe_fps);
	int rebuild = 0;
	int changed = 0;

	if (gop != pipe_gop) {
		/* A kernel without the keyframe control (no 0912) also
		 * applies a GOP only at stream start: rebuild there. */
		if (kvmv_pipe_set_gop(&pipe_state, (unsigned int)gop) ||
		    key_control_missing)
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
	if (!rebuild && __atomic_exchange_n(&force_key, 0, __ATOMIC_ACQ_REL))
		request_key();
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
	uint64_t now = now_ms();

	if (kvmv_rate_add(&fps_rate, now, 0, KVMV_FPS_WINDOW_MS, &kbps, &fps_x10))
		__atomic_store_n(&delivered_fps, (int)((fps_x10 + 5) / 10),
				 __ATOMIC_RELEASE);
	if (!kvmv_rate_add(&rate, now, bytes, KVMV_RATE_WINDOW_MS, &kbps,
			   &fps_x10))
		return;
	DBG("output %u kbit/s at %u.%u fps, target %u kbit/s, encoder told %d fps",
	    kbps, fps_x10 / 10, fps_x10 % 10, target, pipe_fps);
	if (pipe_state.times.frames) {
		const struct kvmv_stage_times *t = &pipe_state.times;
		unsigned int n = t->frames;

		DBG("per frame: capture wait %llu us, scale %llu us, encode %llu us, copy %llu us, capture age at encoded %llu us (max %llu), at pickup %llu us, capture frames per read x100 %llu, waited for the next frame %u, scaled ahead %u of %u (smoothed scale %u us, encode %u us), scaler beside the encoder %llu us",
		    (unsigned long long)(t->capture_us / n),
		    (unsigned long long)(t->scale_us / n),
		    (unsigned long long)(t->encode_us / n),
		    (unsigned long long)(t->copy_us / n),
		    (unsigned long long)(t->age_us / n),
		    (unsigned long long)t->age_max_us,
		    (unsigned long long)(t->pick_age_us / n),
		    (unsigned long long)(t->seq_gap * 100 / n),
		    t->pickup_waits, t->ahead, n, scale_ewma_us, encode_ewma_us,
		    (unsigned long long)(t->overlap_us / n));
		memset(&pipe_state.times, 0, sizeof(pipe_state.times));
	}
	if (!overshoot_warned && target && kbps > target * 2) {
		overshoot_warned = 1;
		log_msg("output runs at %u kbit/s against a %u kbit/s target (%u.%u fps); see ironkvm-dist#35",
			kbps, target, fps_x10 / 10, fps_x10 % 10);
	}
}

static int timed_lock(pthread_mutex_t *lock, unsigned int seconds)
{
	struct timespec ts;

	clock_gettime(CLOCK_REALTIME, &ts);
	ts.tv_sec += seconds;
	return pthread_mutex_timedlock(lock, &ts);
}

/* The slots are claimed from the H.264 path under pipe_lock and from the
 * MJPEG path outside it, so claiming takes a lock of its own. */
static struct kvmv_slot *claim_slot(uint32_t size)
{
	struct kvmv_slot *slot;

	pthread_mutex_lock(&slot_lock);
	slot = kvmv_slot_claim(&slots, size);
	pthread_mutex_unlock(&slot_lock);
	return slot;
}

/* What a read answers for a frame the pipeline could not produce. */
static int pipe_failure(enum kvmv_pipe_status status)
{
	switch (status) {
	case KVMV_PIPE_OK:
		return 0;
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
		if (pipe_state.direct && ++hevc_direct_failures >= KVMV_HEVC_DIRECT_FAILURE_LIMIT) {
			log_msg("H.265: %d failed pictures in a row from the capture buffer; the scaler converts from now on",
				hevc_direct_failures);
			hevc_direct_off = 1;
		}
		pipe_down("frame failed");
		/* Rebuild after a pause rather than on every read: a build
		 * costs a priming encode and a burst of ioctls. */
		return start_failed(IMG_VENC_ERROR, 500);
	}
}

/* When the pipeline build for a video read began, until its first picture
 * (KVMV_DEBUG). */
static uint64_t cold_start_us;

/*
 * Scale-ahead (struct kvmv_pipe): the scaler takes the next frame while the
 * encoder works on this one. KVMV_SCALE_AHEAD=0 never, 1 always, 2 (the
 * default) when the frame time the server asks for is shorter than a read
 * that scales and then encodes: 1080p at 60 fps (about 8 + 13 ms against
 * 16.7), not 30 fps, where a frame scaled ahead would only be older than the
 * one the next read could take. The two stage times are smoothed per
 * pipeline: the scale from reads that scaled their own frame, the encode from
 * every read but the first, which starts the encoder's sequence
 * (scale_ewma_us and encode_ewma_us, above).
 *
 * Not while MJPEG or snapshot reads come too (in the last second): they take
 * capture frames and the scaler between the video reads, a frame scaled ahead
 * then holds one of the three capture buffers for longer, and kvmv-probe's
 * H.264 at 60 fps beside a 30 fps MJPEG reader fell from 28 to 16 fps.
 */

static int want_scale_ahead(void)
{
	static int mode = -1;
	int fps;

	if (mode < 0)
		mode = (int)env_uint("KVMV_SCALE_AHEAD", 2);
	if (mode != 2)
		return mode == 1;
	if (last_image_ms && now_ms() - last_image_ms < 1000U)
		return 0;
	fps = __atomic_load_n(&capture_fps_setting, __ATOMIC_ACQUIRE);
	if (fps <= 0 || !scale_ewma_us || !encode_ewma_us)
		return 0;
	return 1000000U / (unsigned int)fps < scale_ewma_us + encode_ewma_us;
}

/*
 * KVMV_AHEAD_DELAY_US: scale-ahead gives the scaler a frame no earlier than
 * this long after the encoder took its picture, so the two overlap less in
 * the DRAM (trial 33). 0, the default, starts it at once.
 */
static uint32_t ahead_delay_us(void)
{
	static long delay = -1;

	if (delay < 0)
		delay = (long)env_uint("KVMV_AHEAD_DELAY_US", 0);
	return (uint32_t)delay;
}

static uint32_t ewma(uint32_t avg, uint64_t sample)
{
	if (sample > 1000000U)
		sample = 1000000U;
	return avg ? (uint32_t)((avg * 7ULL + sample) / 8U) : (uint32_t)sample;
}

static int read_video_locked(unsigned int width, unsigned int height,
			     uint32_t bitrate_bps, enum kvmv_codec codec,
			     uint8_t **data, uint32_t *size)
{
	int attempt;
	int result;

	if (find_devices_locked())
		return start_failed(IMG_VENC_ERROR, 2000);
	if (!have_encoder(codec)) {
		log_error_once(codec_name(codec), "no encoder for this codec");
		return IMG_VENC_ERROR;
	}
	/*
	 * A codec switch is a new pipe. The old encoder stops streaming first,
	 * which releases the codec SRAM the new one needs.
	 */
	if (codec != pipe_codec) {
		pipe_down_keep("codec changed", 1);
		log_msg("codec %s -> %s", codec_name(pipe_codec), codec_name(codec));
		pipe_codec = codec;
	}
	if (pipe_state.running && (width != pipe_width || height != pipe_height))
		pipe_down_keep("output size changed", 1);
	if (!pipe_state.running) {
		cold_start_us = now_us();
		scale_ewma_us = 0;
		encode_ewma_us = 0;
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
		uint8_t *unit;
		size_t unit_size, prefix_len;
		uint64_t copy_start;
		int type;
		uint64_t scale0 = pipe_state.times.scale_us;
		uint64_t encode0 = pipe_state.times.encode_us;
		unsigned int ahead0 = pipe_state.times.ahead;

		pipe_state.scale_ahead = want_scale_ahead();
		pipe_state.ahead_delay_us = ahead_delay_us();
		result = pipe_failure(kvmv_pipe_encode(&pipe_state, KVMV_FRAME_TIMEOUT_MS,
						       &encoded));
		if (result != 0)
			return result;
		no_frame_count = 0;
		hevc_direct_failures = 0;
		if (!cold_start_us) {
			if (pipe_state.times.ahead == ahead0)
				scale_ewma_us = ewma(scale_ewma_us,
						     pipe_state.times.scale_us - scale0);
			encode_ewma_us = ewma(encode_ewma_us,
					      pipe_state.times.encode_us - encode0);
		}

		/*
		 * Copy the unit out of the encoder's buffer before looking at
		 * it. That buffer is mapped uncached, and the NAL scans below
		 * read it byte by byte, twice: done on the mapping, that cost
		 * one DRAM round trip per byte per scan, which at 1080p is a
		 * large part of a frame time. The copy leaves room in front
		 * for the parameter sets a keyframe may need.
		 */
		slot = claim_slot((uint32_t)(KVMV_PREFIX_ROOM + encoded.size));
		if (slot == NULL) {
			kvmv_pipe_release(&pipe_state, &encoded);
			return IMG_BUFFER_FULL;
		}
		unit = slot->data + KVMV_PREFIX_ROOM;
		unit_size = encoded.size;
		copy_start = now_us();
		kvmv_copy_from_device(unit, encoded.data, unit_size);
		pipe_state.times.copy_us += now_us() - copy_start;
		kvmv_pipe_release(&pipe_state, &encoded);

		kvmv_au_inspect_codec(pipe_state.codec, unit, unit_size, &info);
		/* Chrome refuses the WAVE420L's VPS with its empty extension. */
		if (pipe_state.codec == KVMV_CODEC_KIND_HEVC && info.vps)
			kvmv_hevc_clear_vps_extension(unit, unit_size);
		kvmv_ps_cache_update(&ps_cache, unit, unit_size);
		type = kvmv_au_type(&info);
		if (type != IMG_H264_TYPE_IF && type != IMG_H264_TYPE_PF) {
			DBG("access unit with no picture (%u NALs), skipped", info.nal_count);
			kvmv_slot_abandon(slot);
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
				kvmv_slot_abandon(slot);
				request_key();
				continue;
			}
		}

		/* The prefix goes straight into the room in front of the unit,
		 * then the unit moves down to meet it (cached memory). */
		prefix_len = kvmv_ps_cache_prefix(&ps_cache, &info, slot->data,
						  KVMV_PREFIX_ROOM);
		if (prefix_len != KVMV_PREFIX_ROOM)
			memmove(slot->data + prefix_len, unit, unit_size);
		slot->size = (uint32_t)(prefix_len + unit_size);
		slot->type = (uint8_t)type;
		need_key = 0;
		set_source_locked(1);
		watch_rate(slot->size);
		if (cold_start_us) {
			const struct kvmv_stage_times *t = &pipe_state.times;

			DBG("first picture %llu us after the pipeline build began; %u frames encoded for it: capture wait %llu us, scale %llu us, encode %llu us, copy %llu us",
			    (unsigned long long)(now_us() - cold_start_us), t->frames,
			    (unsigned long long)t->capture_us,
			    (unsigned long long)t->scale_us,
			    (unsigned long long)t->encode_us,
			    (unsigned long long)t->copy_us);
			cold_start_us = 0;
		}

		*data = slot->data;
		*size = slot->size;
		return type;
	}
	return IMG_NOT_EXIST;
}

/* Capture runs inside the H.264 pipeline. Build it at the size and bitrate
 * the stream last used, so a stream that comes back does not rebuild it
 * straight away. Caller holds pipe_lock. */
static int image_pipe_up(void)
{
	if (pipe_state.running)
		return 0;
	return pipe_up(pipe_width, pipe_height,
		       pipe_bitrate ? pipe_bitrate :
		       kvmv_kbps_to_bps(kvmv_default_kbps(pipe_width, pipe_height)));
}

static void announce_range(int full_range)
{
	int state = full_range ? 1 : 2;

	if (__atomic_exchange_n(&range_announced, state, __ATOMIC_RELAXED) == state)
		return;
	if (full_range)
		log_msg("MJPEG: the scaler expands the capture to full range");
	else
		log_msg("MJPEG: the scaler keeps limited range (no ironkvm-dist patch 0908); pictures look flat");
}

/* A return of read_hw_locked: use the software encoder for this picture. */
#define KVMV_USE_SOFTWARE 1000
/* Consecutive hardware failures after which only software is used. */
#define KVMV_HWJPEG_FAILURE_LIMIT 3

/* Whether this read goes to the JPEG unit; opens it on first use. */
static int hw_jpeg_ready(void)
{
	const char *mode = getenv("KVMV_JPEG");

	if (mode != NULL && !strcmp(mode, "sw"))
		return 0;
	if (!devices.jpeg[0] || hw_jpeg_failures >= KVMV_HWJPEG_FAILURE_LIMIT)
		return 0;
	if (hw_jpeg.fd >= 0)
		return 1;
	if (kvmv_hwjpeg_open(&hw_jpeg, devices.jpeg)) {
		log_msg("%s; software JPEG", hw_jpeg.error);
		hw_jpeg_failures = KVMV_HWJPEG_FAILURE_LIMIT;
		return 0;
	}
	if (!__atomic_exchange_n(&hw_jpeg_announced, 1, __ATOMIC_RELAXED))
		log_msg("MJPEG: hardware JPEG encoder %s, %s", devices.jpeg,
			env_uint("KVMV_JPEG_IMPORT", 1) ? "dma-buf import" : "CPU copy");
	return 1;
}

/*
 * Direct MJPEG off: KVMV_JPEG_DIRECT=0, or a JPEG unit without packed input
 * (ironkvm-dist patch 0920). Set once, under pipe_lock.
 */
static int hw_jpeg_direct_off = -1;
static int hw_jpeg_direct_announced;

/* A return of read_direct_locked: take the picture through the scaler. */
#define KVMV_USE_SCALER 1001

/*
 * An MJPEG read at the capture's own size, without the scaler: the JPEG unit
 * reads the UYVY capture buffer itself and expands its limited range to full
 * range (ironkvm-dist patch 0920). At 1080p the scaler only converted the
 * format and the range, at the cost of reading 4 MB and writing 3 MB that
 * the unit then read back, for every picture. Answers KVMV_USE_SCALER when
 * this picture needs the scaler (another size, another capture format) or
 * the unit cannot do it.
 */
static int read_direct_locked(unsigned int width, unsigned int height, int quality,
			      uint8_t **data, uint32_t *size)
{
	const struct v4l2_pix_format *cap = &pipe_state.cap_fmt;
	struct kvmv_frame frame;
	struct kvmv_plan plan;
	struct kvmv_slot *slot;
	const uint8_t *jpeg;
	size_t jpeg_size;
	uint64_t t0, t1, t2;
	int result, rc, saved;

	if (hw_jpeg_direct_off < 0)
		hw_jpeg_direct_off = env_uint("KVMV_JPEG_DIRECT", 1) == 0;
	if (hw_jpeg_direct_off || cap->pixelformat != V4L2_PIX_FMT_UYVY ||
	    kvmv_plan_output(cap->width, cap->height, width, height, &plan) ||
	    plan.out_width != cap->width || plan.out_height != cap->height)
		return KVMV_USE_SCALER;

	t0 = now_us();
	result = pipe_failure(kvmv_pipe_lend(&pipe_state, KVMV_FRAME_TIMEOUT_MS, &frame));
	if (result != 0)
		return result;
	no_frame_count = 0;
	set_source_locked(1);

	t1 = now_us();
	rc = kvmv_hwjpeg_encode_frame(&hw_jpeg, frame.fd, frame.size, frame.index,
				      pipe_state.cap_count, cap->width, cap->height,
				      cap->bytesperline, cap->pixelformat,
				      cap->quantization != V4L2_QUANTIZATION_FULL_RANGE,
				      quality, KVMV_JPEG_QUALITY_DEFAULT, &jpeg, &jpeg_size);
	saved = errno;
	if (kvmv_pipe_give_back(&pipe_state, &frame))
		return pipe_failure(KVMV_PIPE_ERROR);
	if (rc) {
		if (saved == ENOTSUP) {
			log_msg("MJPEG: %s; the scaler converts every picture", hw_jpeg.error);
			hw_jpeg_direct_off = 1;
			return KVMV_USE_SCALER;
		}
		log_msg("%s", hw_jpeg.error);
		if (++hw_jpeg_failures >= KVMV_HWJPEG_FAILURE_LIMIT) {
			log_msg("MJPEG: %d hardware failures in a row; software JPEG from now on",
				hw_jpeg_failures);
			kvmv_hwjpeg_close(&hw_jpeg);
		}
		return KVMV_USE_SOFTWARE;
	}
	hw_jpeg_failures = 0;
	if (!__atomic_exchange_n(&hw_jpeg_direct_announced, 1, __ATOMIC_RELAXED))
		log_msg("MJPEG: the JPEG unit reads the capture buffer at %ux%u, %s range expanded, no scaler",
			cap->width, cap->height,
			cap->quantization != V4L2_QUANTIZATION_FULL_RANGE ? "limited" : "full (not)");

	t2 = now_us();
	slot = claim_slot((uint32_t)jpeg_size);
	if (slot == NULL)
		return IMG_BUFFER_FULL;
	kvmv_copy_from_device(slot->data, jpeg, jpeg_size);
	slot->size = (uint32_t)jpeg_size;
	slot->type = IMG_MJPEG_TYPE;
	DBG("MJPEG %ux%u q%d (unit, direct, %s): picture %llu us, encode %llu us, copy %llu us, %zu bytes",
	    cap->width, cap->height, quality,
	    hw_jpeg.cap_cached ? "cached" : "uncached",
	    (unsigned long long)(t1 - t0), (unsigned long long)(t2 - t1),
	    (unsigned long long)(now_us() - t2), jpeg_size);
	*data = slot->data;
	*size = slot->size;
	return IMG_MJPEG_TYPE;
}

/*
 * An MJPEG read through the JPEG unit: the snapshot scaler context's buffer
 * goes straight to the unit (or is copied into its own buffer with
 * KVMV_JPEG_IMPORT=0), and the JPEG is copied once, into the slot. About
 * 5 ms at 1080p, so it runs under pipe_lock. Answers KVMV_USE_SOFTWARE when
 * the unit cannot take this picture.
 */
static int read_hw_locked(unsigned int width, unsigned int height, int quality,
			  uint8_t **data, uint32_t *size)
{
	int import = env_uint("KVMV_JPEG_IMPORT", 1) != 0;
	struct kvmv_nv12 image;
	struct kvmv_slot *slot;
	const uint8_t *jpeg;
	size_t jpeg_size;
	uint64_t t0, t1, t2;
	int result;

	result = read_direct_locked(width, height, quality, data, size);
	if (result != KVMV_USE_SCALER)
		return result;

	t0 = now_us();
	result = pipe_failure(kvmv_pipe_snapshot(&pipe_state, width, height,
						 KVMV_FRAME_TIMEOUT_MS, !import, &image));
	if (result != 0)
		return result;
	no_frame_count = 0;
	set_source_locked(1);
	announce_range(image.full_range);
	if (!kvmv_hwjpeg_fits(image.width, image.height, image.stride)) {
		kvmv_pipe_snapshot_done(&pipe_state);
		DBG("MJPEG %ux%u: the JPEG unit needs a width in steps of 16; software",
		    image.width, image.height);
		return KVMV_USE_SOFTWARE;
	}

	t1 = now_us();
	if (kvmv_hwjpeg_encode(&hw_jpeg, &image, import ? image.fd : -1,
			       image.fd_size, quality, KVMV_JPEG_QUALITY_DEFAULT,
			       &jpeg, &jpeg_size)) {
		kvmv_pipe_snapshot_done(&pipe_state);
		log_msg("%s", hw_jpeg.error);
		if (++hw_jpeg_failures >= KVMV_HWJPEG_FAILURE_LIMIT) {
			log_msg("MJPEG: %d hardware failures in a row; software JPEG from now on",
				hw_jpeg_failures);
			kvmv_hwjpeg_close(&hw_jpeg);
		}
		return KVMV_USE_SOFTWARE;
	}
	hw_jpeg_failures = 0;
	kvmv_pipe_snapshot_done(&pipe_state);
	if (!__atomic_exchange_n(&hw_jpeg_cache_announced, 1, __ATOMIC_RELAXED))
		log_msg("MJPEG: the JPEG comes back in a %s buffer",
			hw_jpeg.cap_cached ? "cached" : "coherent (uncached; no ironkvm-dist patch 0909)");

	t2 = now_us();
	slot = claim_slot((uint32_t)jpeg_size);
	if (slot == NULL)
		return IMG_BUFFER_FULL;
	kvmv_copy_from_device(slot->data, jpeg, jpeg_size);
	slot->size = (uint32_t)jpeg_size;
	slot->type = IMG_MJPEG_TYPE;
	DBG("MJPEG %ux%u q%d (unit, %s, %s): picture %llu us, encode %llu us, copy %llu us, %zu bytes",
	    image.width, image.height, quality, import ? "import" : "copy",
	    hw_jpeg.cap_cached ? "cached" : "uncached",
	    (unsigned long long)(t1 - t0), (unsigned long long)(t2 - t1),
	    (unsigned long long)(now_us() - t2), jpeg_size);
	*data = slot->data;
	*size = slot->size;
	return IMG_MJPEG_TYPE;
}

/*
 * Take a picture for a software MJPEG read: the newest captured frame, scaled
 * by the snapshot scaler context and copied into the JPEG encoder's planes.
 * Caller holds pipe_lock; the encode itself happens after it is released.
 */
static int snapshot_locked(unsigned int width, unsigned int height)
{
	struct kvmv_nv12 image;
	int result;

	result = pipe_failure(kvmv_pipe_snapshot(&pipe_state, width, height,
						 KVMV_FRAME_TIMEOUT_MS, 1, &image));
	if (result != 0)
		return result;
	no_frame_count = 0;
	announce_range(image.full_range);
	result = kvmv_jpeg_load(&jpeg_state, &image);
	kvmv_pipe_snapshot_done(&pipe_state);
	if (result != 0) {
		log_error_once("JPEG", "cannot hold the picture (out of memory?)");
		return IMG_VENC_ERROR;
	}
	set_source_locked(1);
	return 0;
}

static int read_image(uint16_t width, uint16_t height, int quality,
		      uint8_t **data, uint32_t *size)
{
	const uint8_t *jpeg;
	struct kvmv_slot *slot;
	size_t jpeg_size;
	uint64_t t0, t1;
	char error[160];
	int result;

	/*
	 * jpeg_lock keeps one picture in the encoder at a time and is always
	 * taken before pipe_lock. With the JPEG unit the whole read, about
	 * 10 ms at 1080p, runs under pipe_lock. In software pipe_lock is held
	 * only to take the picture (a capture wait, one scale, a copy); the
	 * encode, a fifth of a second at 1080p, runs without it, so the H.264
	 * stream keeps going.
	 */
	if (timed_lock(&jpeg_lock, KVMV_JPEG_LOCK_TIMEOUT_S) != 0)
		return KVMV_RET_RETRIEVING;
	if (timed_lock(&pipe_lock, KVMV_LOCK_TIMEOUT_S) != 0) {
		pthread_mutex_unlock(&jpeg_lock);
		return KVMV_RET_RETRIEVING;
	}
	__atomic_store_n(&last_read_ms, now_ms(), __ATOMIC_RELEASE);
	last_image_ms = now_ms();
	if (__atomic_load_n(&capture_stopped, __ATOMIC_ACQUIRE)) {
		pthread_mutex_unlock(&pipe_lock);
		pthread_mutex_unlock(&jpeg_lock);
		return IMG_NOT_EXIST;
	}
	result = image_pipe_up();
	if (result == 0 && hw_jpeg_ready()) {
		result = read_hw_locked(width, height, quality, data, size);
		if (result != KVMV_USE_SOFTWARE) {
			pthread_mutex_unlock(&pipe_lock);
			pthread_mutex_unlock(&jpeg_lock);
			return result;
		}
		result = 0;
	}
	t0 = now_us();
	if (result == 0)
		result = snapshot_locked(width, height);
	pthread_mutex_unlock(&pipe_lock);
	if (result != 0) {
		pthread_mutex_unlock(&jpeg_lock);
		return result;
	}

	t1 = now_us();
	if (kvmv_jpeg_compress(&jpeg_state, quality, &jpeg, &jpeg_size, error,
			       sizeof(error))) {
		pthread_mutex_unlock(&jpeg_lock);
		log_msg("%s", error);
		return IMG_VENC_ERROR;
	}
	slot = claim_slot((uint32_t)jpeg_size);
	if (slot == NULL) {
		pthread_mutex_unlock(&jpeg_lock);
		return IMG_BUFFER_FULL;
	}
	memcpy(slot->data, jpeg, jpeg_size);
	slot->size = (uint32_t)jpeg_size;
	slot->type = IMG_MJPEG_TYPE;
	DBG("MJPEG %ux%u q%d: picture %llu us, encode %llu us, %zu bytes",
	    jpeg_state.width, jpeg_state.height, quality,
	    (unsigned long long)(t1 - t0), (unsigned long long)(now_us() - t1),
	    jpeg_size);
	pthread_mutex_unlock(&jpeg_lock);

	*data = slot->data;
	*size = slot->size;
	return IMG_MJPEG_TYPE;
}

static int read_frame(uint16_t width, uint16_t height, uint8_t codec,
		      int bitrate_kbps, uint8_t **data, uint32_t *size)
{
	int result;

	if (data == NULL || size == NULL)
		return IMG_NOT_EXIST;
	*data = NULL;
	*size = 0;

	if (codec == KVMV_CODEC_MJPEG)
		return read_image(width, height, bitrate_kbps, data, size);
	if (codec != KVMV_CODEC_H264 && codec != KVMV_CODEC_H265)
		return IMG_VENC_ERROR;

	if (timed_lock(&pipe_lock, KVMV_LOCK_TIMEOUT_S) != 0)
		return KVMV_RET_RETRIEVING;
	__atomic_store_n(&last_read_ms, now_ms(), __ATOMIC_RELEASE);
	if (__atomic_load_n(&capture_stopped, __ATOMIC_ACQUIRE)) {
		pthread_mutex_unlock(&pipe_lock);
		return IMG_NOT_EXIST;
	}
	/* 0 is outside the ABI's 500..10000: give the stream its size's default. */
	if (bitrate_kbps == 0)
		bitrate_kbps = kvmv_default_kbps(width, height);
	result = read_video_locked(width, height, kvmv_kbps_to_bps(bitrate_kbps),
				   codec == KVMV_CODEC_H265 ? KVMV_CODEC_KIND_HEVC :
							      KVMV_CODEC_KIND_H264,
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
			receiver_remember(signal, &timings);
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
				pipe_down_keep("no reads", 1);
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
	key_control_missing = 0;
	hw_jpeg_failures = 0;
	next_start_ms = 0;
	pthread_mutex_unlock(&pipe_lock);
	pthread_mutex_lock(&jpeg_lock);
	kvmv_jpeg_free(&jpeg_state);
	pthread_mutex_unlock(&jpeg_lock);
	set_capture_enabled(0);
	__atomic_store_n(&source_locked, 0, __ATOMIC_RELEASE);
	pthread_mutex_lock(&slot_lock);
	kvmv_slots_free_all(&slots);
	pthread_mutex_unlock(&slot_lock);
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
		set_h264_fps(_fps);
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
	pthread_mutex_lock(&slot_lock);
	kvmv_slots_free_all(&slots);
	pthread_mutex_unlock(&slot_lock);
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

	if (__atomic_exchange_n(&fps_setting, fps, __ATOMIC_ACQ_REL) != fps) {
		/* A new rate starts from what was asked, not from what the
		 * old one delivered. */
		__atomic_store_n(&delivered_fps, 0, __ATOMIC_RELEASE);
		DBG("set_h264_fps %d", fps);
	}
}

void set_capture_fps(uint8_t _fps)
{
	int fps = kvmv_clamp(_fps, KVMV_FPS_MIN, KVMV_FPS_MAX);

	/*
	 * The capture node has no frame interval control and runs at the
	 * source's rate, but nothing beyond its DMA happens to a frame nobody
	 * reads: each read takes the newest frame and the rest go straight back
	 * to the driver, unscaled and unencoded. What this does change is the
	 * rate the encoder is told: frames cannot reach it faster than this, so
	 * the lower of this and set_h264_fps is what its rate control budgets
	 * for (applied on the next read).
	 */
	if (__atomic_exchange_n(&capture_fps_setting, fps, __ATOMIC_ACQ_REL) != fps) {
		__atomic_store_n(&delivered_fps, 0, __ATOMIC_RELEASE);
		DBG("set_capture_fps %d", fps);
	}
}

void set_frame_detact(uint8_t _frame_detact)
{
	/* Frame detection skips unchanged MJPEG frames. Recorded only: every
	 * MJPEG read here encodes a picture (answer 5, "not changed", is never
	 * given). */
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

/*
 * Not part of kvm_vision.h, and not in Sipeed's library: which codecs this
 * library can deliver, by kvmv_read_img/kvmv_read_video codec number (0 MJPEG,
 * 1 H.264, 2 H.265). NanoKVM-Server looks it up with dlsym, so it runs with
 * either library, and treats its absence as "all three" (the vendor library).
 * Without it a server restored to H.265 asks for codec 2 forever and every
 * read answers -2, which is what left the web UI without video.
 */
__attribute__((visibility("default"))) uint8_t kvmv_codec_supported(uint8_t _codec);
uint8_t kvmv_codec_supported(uint8_t _codec)
{
	uint8_t supported = 0;

	if (_codec == KVMV_CODEC_MJPEG)
		return 1;
	if (_codec != KVMV_CODEC_H264 && _codec != KVMV_CODEC_H265)
		return 0;
	/*
	 * By the encoder nodes the kernel has: the Coda980 for H.264, the
	 * WAVE420L (ironkvm-dist#55) for H.265. Discovery opens the nodes
	 * once; a pipeline busy for longer than the lock timeout answers from
	 * what is known, or the vendor library's "yes" for H.264.
	 */
	if (timed_lock(&pipe_lock, KVMV_LOCK_TIMEOUT_S) != 0)
		return devices_found ? (uint8_t)have_encoder(_codec == KVMV_CODEC_H265 ?
							   KVMV_CODEC_KIND_HEVC :
							   KVMV_CODEC_KIND_H264) :
				       _codec == KVMV_CODEC_H264;
	if (find_devices_locked() == 0)
		supported = (uint8_t)have_encoder(_codec == KVMV_CODEC_H265 ?
						  KVMV_CODEC_KIND_HEVC :
						  KVMV_CODEC_KIND_H264);
	pthread_mutex_unlock(&pipe_lock);
	return supported;
}
