/*
 * Decisions libkvm-v4l2 makes from what the drivers report. Kept free of
 * device access so the host unit test can drive every branch.
 */
#ifndef KVMV_POLICY_H
#define KVMV_POLICY_H

#include <stddef.h>
#include <stdint.h>
#include <linux/videodev2.h>

/* The same ranges the vendor library clamps to. */
#define KVMV_BITRATE_MIN_KBPS 500
#define KVMV_BITRATE_MAX_KBPS 10000
#define KVMV_GOP_MIN 1
#define KVMV_GOP_MAX 100
#define KVMV_FPS_MIN 10
#define KVMV_FPS_MAX 60
#define KVMV_DEFAULT_GOP 30
#define KVMV_DEFAULT_FPS 60
#define KVMV_DEFAULT_KBPS 1000

int kvmv_clamp(int value, int min, int max);
uint32_t kvmv_kbps_to_bps(int kbps);

/* Encoder geometry for one pipeline build. */
struct kvmv_plan {
	unsigned int src_width, src_height; /* what the capture node delivers */
	unsigned int out_width, out_height; /* visible size of the stream */
	unsigned int coded_height; /* out_height rounded up to a macroblock */
};

/*
 * Work out the stream size. A requested size of 0 in either dimension follows
 * the source, as the vendor library does. The VPSS only scales down here, so
 * a request larger than the source is held to the source. The width is
 * rounded down to a multiple of 16 so that the encoder's line stride and the
 * scaler's agree, and the height to an even number for 4:2:0. Returns 0, or
 * -1 when the source itself is unusable.
 */
int kvmv_plan_output(unsigned int src_width, unsigned int src_height,
		     unsigned int req_width, unsigned int req_height,
		     struct kvmv_plan *plan);

enum kvmv_signal {
	KVMV_SIGNAL_UNKNOWN = 0, /* no subdevice to ask, or it cannot answer */
	KVMV_SIGNAL_NONE, /* no cable, no lock */
	KVMV_SIGNAL_OK, /* locked, and the capture node can take it */
	KVMV_SIGNAL_UNSUPPORTED, /* locked, but not a mode the capture takes */
	KVMV_SIGNAL_OUT_OF_RANGE, /* the receiver reports timings it refuses */
};

/*
 * Classify a VIDIOC_QUERY_DV_TIMINGS answer. ret and err are the ioctl return
 * and errno. need_width and need_height are the capture node's fixed frame
 * size, or 0 to accept any.
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
};

/*
 * Which part of the pipeline a video node plays, from VIDIOC_QUERYCAP and
 * whether its capture queue offers H.264. Node numbers are not stable, so
 * discovery goes by driver name and capability, never by /dev/videoN.
 */
enum kvmv_role kvmv_match_role(const char *driver, uint32_t device_caps,
			       int encodes_h264, int takes_nv12);

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

#endif
