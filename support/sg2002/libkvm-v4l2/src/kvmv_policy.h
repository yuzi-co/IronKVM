/*
 * Decisions libkvm-v4l2 makes from what the drivers report. Kept free of
 * device access so the host unit test can drive every branch.
 */
#ifndef KVMV_POLICY_H
#define KVMV_POLICY_H

#include <stddef.h>
#include <stdint.h>
#include <linux/videodev2.h>

/* The same ranges the vendor library clamps to, but for the frame rate: the
 * capture runs at the source's rate, which can be above 60 Hz with an EDID
 * that offers it (ironkvm-dist trial 31), so 120 rather than 60. */
#define KVMV_BITRATE_MIN_KBPS 500
#define KVMV_BITRATE_MAX_KBPS 10000
#define KVMV_GOP_MIN 1
#define KVMV_GOP_MAX 100
#define KVMV_FPS_MIN 10
#define KVMV_FPS_MAX 120
#define KVMV_DEFAULT_GOP 30
#define KVMV_DEFAULT_FPS 60

/*
 * H.264 rate control on the Coda980 (issue #35, run sheet trial 10). The
 * encoder keeps the picture QP inside [min, max]. The initial delay is the
 * rate control's buffer: with none, the Coda980 holds every picture at QP 49
 * to 51.
 *
 * The maximum is the encoder's own, 51 (trial 51). Trial 10 held it to 42 so
 * text stayed legible on a busy screen, at the cost of the bitrate: scrolling
 * text then ran at two to five times the target and a screen changing
 * everywhere at fifteen (31 Mbit/s for 2), more than a WebRTC viewer takes, so
 * the viewer stalled. Quality gives way now, not the frame rate; the
 * output-rate guard below handles what even QP 51 cannot hold.
 */
#define KVMV_H264_MIN_QP 18
#define KVMV_H264_MAX_QP 51
#define KVMV_H264_VBV_DELAY_MS 1000

/*
 * The buffer for a longer GOP (ironkvm-dist#72, run sheet trial 64). The
 * Coda980's rate control lets the I picture fill the buffer in proportion to
 * the GOP: at 2000 kbit/s and 1000 ms, a 1080p keyframe on btop is about 100
 * kB at GOP 30 and 250 kB, the whole buffer, at GOP 99. The server's adaptive
 * keyframes ask for 2 s GOPs, and on slot B a keyframe costs about 1 ms of
 * the send path per 1.2 kB, so the buffer shrinks with the GOP to keep the I
 * picture near its GOP 30 size: 1000 ms * 30 / GOP, never under
 * KVMV_H264_VBV_DELAY_MIN_MS, above the cliff where the I picture falls apart
 * (150 to 200 ms, trial 61). Trial 64, 1080p, btop: GOP 99 at 300 ms gives
 * 110 kB keyframes, GOP 60 at 500 ms 173 kB (30 fps; 124 kB at GOP 30).
 */
#define KVMV_H264_VBV_DELAY_MIN_MS 300
unsigned int kvmv_h264_vbv_delay_ms(unsigned int gop);

/*
 * H.265 on the WAVE420L (ironkvm-dist#55, run sheet trials 13, 20 and 51):
 * the same maximum QP as H.264, a lower minimum and a longer initial delay.
 *
 * At QP 18 the WAVE420L stops at 2 to 3.4 Mbit/s on screen content whatever
 * is asked, up to 2 dB under the Coda980 at its QP 18; at 12 it spends more
 * of what is asked and passes the Coda980 (QP 8 gains nothing over 12). The
 * minimum only acts when the bitrate leaves room.
 *
 * With a 1 s GOP the IDR takes most of the bits on a screen, and a 1000 ms
 * buffer starves it: 2000 ms gives 1 to 6 dB more on still, changing and
 * scrolling text at 1 to 3 Mbit/s for 0 to 8% more bits. 3000 ms gains a
 * little more on scrolling text, but lets the stream run further from the
 * bitrate over a longer window, which a slow link sees as delay.
 */
#define KVMV_H265_MIN_QP 12
#define KVMV_H265_MAX_QP 51
#define KVMV_H265_VBV_DELAY_MS 2000

int kvmv_clamp(int value, int min, int max);
uint32_t kvmv_kbps_to_bps(int kbps);

struct kvmv_qp_range {
	int min_qp, max_qp;
};

/*
 * The QP range to ask for: the defaults above, or spec when it reads
 * "min:max" with 0 <= min <= max <= 51 (KVMV_H264_QP in the environment).
 * "0:51" leaves the encoder's own range alone.
 */
void kvmv_h264_qp_range(const char *spec, struct kvmv_qp_range *range);

/* The same for H.265: KVMV_H265_MIN_QP to KVMV_H265_MAX_QP, or spec. */
void kvmv_h265_qp_range(const char *spec, struct kvmv_qp_range *range);

/*
 * The H.264 bitrate in kbit/s for a stream of width x height when the caller
 * gives none: 3000 kbit/s at 1080p, the server's default. Half of it follows
 * the pixel count, because desktop text costs more per pixel the smaller it
 * is drawn: about 2200 kbit/s at 720p. Rounded to 100, within the range the
 * library accepts. 0 x 0 is the source size, taken as 1080p.
 */
int kvmv_default_kbps(unsigned int width, unsigned int height);

/* Encoder geometry for one pipeline build. */
struct kvmv_plan {
	unsigned int src_width, src_height; /* what the capture node delivers */
	unsigned int out_width, out_height; /* visible size of the stream */
	unsigned int coded_height; /* out_height rounded up to a macroblock */
};

/*
 * The largest picture either encoder takes: the Coda980 stops at 1920x1088,
 * which is 1080 visible lines.
 */
#define KVMV_ENCODE_MAX_WIDTH 1920
#define KVMV_ENCODE_MAX_HEIGHT 1080

/*
 * Work out the stream size. A requested size of 0 in either dimension follows
 * the source, as the vendor library does. With keep_shape the stream has the
 * source's shape at the requested height, or at the source's height when that
 * is smaller, and the requested width is not used: a 1920x1200 source asked
 * for 1920x1080 gives 1728x1080, 1024x768 asked for 1280x720 gives 960x720
 * (ironkvm-dist#37). Without it each side is held to the request on its own,
 * so a source of another shape is stretched to the request's (1920x1200 gives
 * 1920x1080). The VPSS only scales down here, so a request larger than the
 * source is held to the source. A size beyond the encoder's box is scaled
 * into it with its shape kept (1600x1200 gives 1440x1080). The width is
 * rounded down to a multiple of 16 so that the encoder's line stride and the
 * scaler's agree, and the height to an even number for 4:2:0. Returns 0, or
 * -1 when the source itself is unusable.
 */
int kvmv_plan_output(unsigned int src_width, unsigned int src_height,
		     unsigned int req_width, unsigned int req_height,
		     int keep_shape, struct kvmv_plan *plan);

/*
 * HDMI modes the capture can take (ironkvm-dist#37, patches 0936 to 0938).
 * A UYVY frame of the mode, its line rounded to 16 bytes and its height to
 * 16 lines, must be no larger than a 1920x1200 frame (4.4 MB). The capture
 * buffers share the 32 MiB media pool with the encoders' frames. From
 * ironkvm-dist patch 0940 the pool hands out the pages a buffer needs, so
 * three 1920x1200 buffers take 13.2 MiB, 1.2 more than three at 1080p;
 * before it each took an 8 MiB block and the encoders no longer fitted
 * (trial 44). The scaler takes 64 to 2880 pixels a side.
 */
#define KVMV_SOURCE_MIN_SIZE 64
#define KVMV_SOURCE_MAX_WIDTH 2880
#define KVMV_CAPTURE_FRAME_MAX (3840U * 1200U)

/* 1 when a source of width x height fits the capture, else 0. */
int kvmv_source_fits(unsigned int width, unsigned int height);

enum kvmv_signal {
	KVMV_SIGNAL_UNKNOWN = 0, /* no subdevice to ask, or it cannot answer */
	KVMV_SIGNAL_NONE, /* no cable, no lock */
	KVMV_SIGNAL_OK, /* locked, and the capture node can take it */
	KVMV_SIGNAL_UNSUPPORTED, /* locked, but not a mode the capture takes */
	KVMV_SIGNAL_OUT_OF_RANGE, /* the receiver reports timings it refuses */
};

/*
 * Classify a VIDIOC_QUERY_DV_TIMINGS answer. ret and err are the ioctl return
 * and errno. need_width and need_height are a frame size the mode must have,
 * or 0 to accept any progressive mode.
 */
enum kvmv_signal kvmv_classify_timings(int ret, int err,
				       const struct v4l2_dv_timings *timings,
				       unsigned int need_width,
				       unsigned int need_height);

/* Whether the receiver sees a source at all, whatever its mode. */
int kvmv_signal_present(enum kvmv_signal signal);

/*
 * The kvmv_read_* return for a signal that stops a pipeline build, or 0 when
 * the build may go ahead.
 */
int kvmv_signal_result(enum kvmv_signal signal);

enum kvmv_role {
	KVMV_ROLE_NONE = 0,
	KVMV_ROLE_CAPTURE,
	KVMV_ROLE_SCALER,
	KVMV_ROLE_ENCODER,
	KVMV_ROLE_JPEG, /* the hardware JPEG encoder, optional */
	KVMV_ROLE_ENCODER_HEVC, /* NV12 to H.265 (the WAVE420L), optional */
};

/*
 * Which part of the pipeline a video node plays, from VIDIOC_QUERYCAP and
 * whether its capture queue offers H.264. Node numbers are not stable, so
 * discovery goes by driver name and capability, never by /dev/videoN.
 */
enum kvmv_role kvmv_match_role(const char *driver, uint32_t device_caps,
			       int encodes_h264, int encodes_hevc, int takes_nv12);

int kvmv_subdev_name_matches(const char *name);

/*
 * Measured output rate over a window, for the overshoot diagnostic in
 * issue #35. kvmv_rate_add answers 1 and sets *kbps when a window closes.
 */
struct kvmv_rate {
	uint64_t window_start_ms;
	uint64_t bytes;
	unsigned int frames;
};

void kvmv_rate_reset(struct kvmv_rate *rate);
int kvmv_rate_add(struct kvmv_rate *rate, uint64_t now_ms, size_t bytes,
		  unsigned int window_ms, unsigned int *kbps, unsigned int *fps_x10);

/*
 * Output-rate guard (ironkvm-dist#35, run sheet trial 51). On a screen that
 * changes everywhere at once the encoders cannot reach a low bitrate even at
 * their highest QP, and a stream far over its bitrate stalls a WebRTC viewer:
 * its sender falls behind and drops frames up to the next keyframe. The guard
 * keeps the bytes of the pictures handed out over the last window_ms and
 * tells the caller to leave the next picture out while they reach the limit,
 * so any window of that length carries at most the limit and one picture.
 * A picture left out is never encoded, so no reference goes missing. The
 * library asks only while delta pictures cost their share of the bitrate
 * (bitrate / frame rate) or more: after a calm screen's keyframe the window
 * can be full with nothing left to save.
 */
#define KVMV_GUARD_SLOTS 256

struct kvmv_guard {
	uint64_t at_ms[KVMV_GUARD_SLOTS];
	uint32_t bytes[KVMV_GUARD_SLOTS];
	unsigned int head, count; /* oldest entry, entries held */
	uint64_t sum; /* bytes of the entries held */
};

/*
 * The limit in percent of the bitrate, over one second: 140 by default
 * (KVMV_RATE_GUARD in the environment, 0 turns the guard off). With one
 * picture on top it holds a second to about 1.5 times the bitrate on the
 * busiest screen, and stays out of the way of a rate control that holds its
 * target: a calm screen never reaches it.
 */
#define KVMV_GUARD_PERCENT 140
#define KVMV_GUARD_WINDOW_MS 1000

void kvmv_guard_reset(struct kvmv_guard *guard);
/* A picture of bytes handed out at now_ms. */
void kvmv_guard_add(struct kvmv_guard *guard, uint64_t now_ms, uint32_t bytes);
/*
 * 1 when the pictures handed out in the window_ms before now_ms already hold
 * limit_bytes or more, so the next one should be left out; 0 otherwise, and
 * always 0 for a limit of 0.
 */
int kvmv_guard_full(struct kvmv_guard *guard, uint64_t now_ms,
		    uint64_t limit_bytes, unsigned int window_ms);
/* The byte limit for bitrate_bps at percent over window_ms; 0 for percent 0. */
uint64_t kvmv_guard_limit(uint32_t bitrate_bps, unsigned int percent,
			  unsigned int window_ms);

/*
 * The frame rate to give the encoder's OUTPUT queue (VIDIOC_S_PARM). With rate
 * control on, the Coda budgets bitrate / rate per frame, so the rate has to be
 * the one frames actually arrive at, or every frame gets a share of a stream
 * that never comes and the output runs over its bitrate.
 *
 * asked is set_h264_fps (or the read's fps argument), capture is
 * set_capture_fps; frames cannot arrive faster than either. delivered is the
 * measured rate of frames handed out, rounded, or 0 before there is one; when
 * it falls more than 10% short of the target, it is the rate given. current
 * is what the encoder holds now (0 for none): a change smaller than 10% of it
 * while tracking the delivered rate is not worth an ioctl, so current is
 * answered. Never below 1.
 */
int kvmv_encoder_fps(int asked, int capture, int delivered, int current);

/*
 * How a pipeline start brings the encoder up.
 *
 * KVMV_PRIME_NONE: both queues streamed on, nothing encoded. The WAVE420L
 * starts its sequence at STREAMON and makes its first picture an IDR, so the
 * first live picture is the first output. The Coda980 waits for a queued
 * picture (min_queued_buffers 1), so its sequence setup moves into the first
 * read; in trial 24 that encode took no longer than one after a priming
 * picture (12 ms), and the build was 23 ms shorter. Both encoders' first output is an IDR with
 * its parameter sets.
 *
 * KVMV_PRIME_WAIT: one black picture encoded and its output thrown away
 * before the start returns, as before trial 24.
 *
 * KVMV_PRIME_ASYNC: the black picture is queued and the encoder streamed on,
 * which sets the sequence up at once, but the start does not wait for its
 * output; the first read collects it while the capture and the scaler run.
 */
enum kvmv_prime {
	KVMV_PRIME_NONE = 0,
	KVMV_PRIME_WAIT = 1,
	KVMV_PRIME_ASYNC = 2,
};

/*
 * The mode for an encoder, from KVMV_PRIME ("0", "1", "2", or NULL or empty
 * for the default): none for either encoder (trial 24: async came to 66 ms
 * from build to first H.264 picture, none to 59). hevc is kept for a
 * per-encoder default.
 */
enum kvmv_prime kvmv_prime_mode(const char *env, int hevc);

/*
 * Whether a receiver answer taken at seen_ms may stand in for a fresh query
 * at now_ms: max_age_ms 0 never, and an answer from the future (a clock that
 * went back) never.
 */
int kvmv_receiver_fresh(uint64_t seen_ms, uint64_t now_ms, unsigned int max_age_ms);

/*
 * How long a pipeline build that found no usable source waits before the
 * next try. After an HDMI source change the receiver reports no signal until
 * it has locked to the new mode, about 0.5 to 1 s after the change on the
 * LT6911UXC (trial 56). A build retried every 500 ms then started up to
 * 500 ms after the signal was back, on every mode change; within
 * KVMV_CHANGE_WINDOW_MS of a change it retries every KVMV_CHANGE_RETRY_MS
 * instead. Outside it (no cable, the host asleep) the slow interval stands,
 * slow_ms. change_ms is when the source last changed, 0 for never.
 */
#define KVMV_CHANGE_WINDOW_MS 5000U
#define KVMV_CHANGE_RETRY_MS 100U
unsigned int kvmv_start_retry_ms(uint64_t change_ms, uint64_t now_ms,
				 unsigned int slow_ms);

#endif
