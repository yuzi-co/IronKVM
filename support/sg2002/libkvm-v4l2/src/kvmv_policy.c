#include "kvmv_policy.h"

#include <errno.h>
#include <stdio.h>
#include <string.h>

#include "kvm_vision.h"

int kvmv_clamp(int value, int min, int max)
{
	if (value < min)
		return min;
	if (value > max)
		return max;
	return value;
}

uint32_t kvmv_kbps_to_bps(int kbps)
{
	return (uint32_t)kvmv_clamp(kbps, KVMV_BITRATE_MIN_KBPS,
				    KVMV_BITRATE_MAX_KBPS) * 1000U;
}

static void qp_range(const char *spec, int def_min, int def_max,
		     struct kvmv_qp_range *range)
{
	int min_qp, max_qp;
	char extra;

	range->min_qp = def_min;
	range->max_qp = def_max;
	if (spec == NULL ||
	    sscanf(spec, "%d:%d%c", &min_qp, &max_qp, &extra) != 2)
		return;
	if (min_qp < 0 || max_qp > 51 || min_qp > max_qp)
		return;
	range->min_qp = min_qp;
	range->max_qp = max_qp;
}

void kvmv_h264_qp_range(const char *spec, struct kvmv_qp_range *range)
{
	qp_range(spec, KVMV_H264_MIN_QP, KVMV_H264_MAX_QP, range);
}

void kvmv_h265_qp_range(const char *spec, struct kvmv_qp_range *range)
{
	qp_range(spec, KVMV_H265_MIN_QP, KVMV_H265_MAX_QP, range);
}

int kvmv_default_kbps(unsigned int width, unsigned int height)
{
	const uint64_t full = 1920 * 1080;
	uint64_t pixels = (uint64_t)width * height;
	uint64_t kbps;

	if (pixels == 0)
		pixels = full;
	kbps = 1500 + 1500 * pixels / full;
	kbps = (kbps + 50) / 100 * 100;
	return kvmv_clamp((int)(kbps > 100000 ? 100000 : kbps),
			  KVMV_BITRATE_MIN_KBPS, KVMV_BITRATE_MAX_KBPS);
}

/*
 * The source's shape at the requested height, or at the source's own height
 * when that is smaller. The width goes to the nearest multiple of 16 the
 * source has; when that moved it, the height is worked out again from the
 * width, so the picture keeps the source's ratio to within a line. A source
 * kept at its own height whose width is not a multiple of 16 (1366x768) only
 * loses the last few columns' worth of width, as without keep_shape.
 */
static void keep_shape_size(unsigned int src_width, unsigned int src_height,
			    unsigned int req_height, unsigned int *width,
			    unsigned int *height)
{
	uint64_t h = req_height > src_height ? src_height : req_height;
	uint64_t w = (h * src_width * 2 + src_height) / ((uint64_t)src_height * 2);
	uint64_t w16 = (w + 8) & ~(uint64_t)15;
	uint64_t max16 = src_width & ~15U;

	if (w16 > max16)
		w16 = max16;
	if (w16 != w && !(h == src_height && w16 == max16)) {
		uint64_t h2 = (w16 * src_height * 2 + src_width) / ((uint64_t)src_width * 2);

		if (h2 < h)
			h = h2;
	}
	*width = (unsigned int)w16;
	*height = (unsigned int)h;
}

int kvmv_plan_output(unsigned int src_width, unsigned int src_height,
		     unsigned int req_width, unsigned int req_height,
		     int keep_shape, struct kvmv_plan *plan)
{
	unsigned int width, height;

	memset(plan, 0, sizeof(*plan));
	if (src_width < 16 || src_height < 16)
		return -1;

	if (req_width == 0 || req_height == 0) {
		width = src_width;
		height = src_height;
	} else if (keep_shape) {
		keep_shape_size(src_width, src_height, req_height, &width, &height);
	} else {
		width = req_width > src_width ? src_width : req_width;
		height = req_height > src_height ? src_height : req_height;
	}
	if (width > KVMV_ENCODE_MAX_WIDTH || height > KVMV_ENCODE_MAX_HEIGHT) {
		/* Into the encoder's box, keeping the shape. */
		uint64_t w = width, h = height;

		if (w * KVMV_ENCODE_MAX_HEIGHT > h * KVMV_ENCODE_MAX_WIDTH) {
			height = (unsigned int)(h * KVMV_ENCODE_MAX_WIDTH / w);
			width = KVMV_ENCODE_MAX_WIDTH;
		} else {
			width = (unsigned int)(w * KVMV_ENCODE_MAX_HEIGHT / h);
			height = KVMV_ENCODE_MAX_HEIGHT;
		}
	}
	width &= ~15U;
	height &= ~1U;
	if (width < 16 || height < 16)
		return -1;

	plan->src_width = src_width;
	plan->src_height = src_height;
	plan->out_width = width;
	plan->out_height = height;
	plan->coded_height = (height + 15U) & ~15U;
	return 0;
}

int kvmv_source_fits(unsigned int width, unsigned int height)
{
	uint64_t line = ((uint64_t)width * 2 + 15) & ~(uint64_t)15;
	uint64_t lines = ((uint64_t)height + 15) & ~(uint64_t)15;

	if (width < KVMV_SOURCE_MIN_SIZE || height < KVMV_SOURCE_MIN_SIZE ||
	    width > KVMV_SOURCE_MAX_WIDTH || height > KVMV_SOURCE_MAX_WIDTH)
		return 0;
	return line * lines <= KVMV_CAPTURE_FRAME_MAX;
}

enum kvmv_signal kvmv_classify_timings(int ret, int err,
				       const struct v4l2_dv_timings *timings,
				       unsigned int need_width,
				       unsigned int need_height)
{
	const struct v4l2_bt_timings *bt;

	if (ret != 0) {
		switch (err) {
		case ENOLINK:
		case ENOLCK:
			return KVMV_SIGNAL_NONE;
		case ERANGE:
			return KVMV_SIGNAL_OUT_OF_RANGE;
		default:
			/* ENOTTY, ENOIOCTLCMD and the like: nothing to learn. */
			return KVMV_SIGNAL_UNKNOWN;
		}
	}
	if (timings == NULL || timings->type != V4L2_DV_BT_656_1120)
		return KVMV_SIGNAL_UNKNOWN;

	bt = &timings->bt;
	if (bt->width == 0 || bt->height == 0)
		return KVMV_SIGNAL_NONE;
	if (bt->interlaced)
		return KVMV_SIGNAL_UNSUPPORTED;
	if ((need_width && bt->width != need_width) ||
	    (need_height && bt->height != need_height))
		return KVMV_SIGNAL_UNSUPPORTED;
	return KVMV_SIGNAL_OK;
}

int kvmv_signal_present(enum kvmv_signal signal)
{
	return signal == KVMV_SIGNAL_OK || signal == KVMV_SIGNAL_UNSUPPORTED ||
	       signal == KVMV_SIGNAL_OUT_OF_RANGE;
}

int kvmv_signal_result(enum kvmv_signal signal)
{
	switch (signal) {
	case KVMV_SIGNAL_NONE:
		return IMG_NOT_EXIST; /* -1, what the vendor library says without HDMI */
	case KVMV_SIGNAL_UNSUPPORTED:
		return -6; /* "Unsupported resolution" */
	case KVMV_SIGNAL_OUT_OF_RANGE:
		return -7; /* "HDMI input resolution error" */
	case KVMV_SIGNAL_OK:
	case KVMV_SIGNAL_UNKNOWN:
	default:
		return 0;
	}
}

enum kvmv_role kvmv_match_role(const char *driver, uint32_t device_caps,
			       int encodes_h264, int encodes_hevc, int takes_nv12)
{
	if (driver == NULL)
		return KVMV_ROLE_NONE;
	if (!strcmp(driver, "sg2002-capture") &&
	    (device_caps & V4L2_CAP_VIDEO_CAPTURE) &&
	    (device_caps & V4L2_CAP_STREAMING))
		return KVMV_ROLE_CAPTURE;
	if (!(device_caps & V4L2_CAP_VIDEO_M2M) ||
	    !(device_caps & V4L2_CAP_STREAMING))
		return KVMV_ROLE_NONE;
	if (!strcmp(driver, "sg2002-vpss"))
		return KVMV_ROLE_SCALER;
	if (!strcmp(driver, "sg2002-jpeg"))
		return KVMV_ROLE_JPEG;
	/*
	 * The Coda registers several nodes (encoder, decoder, JPEG). The one
	 * wanted takes NV12 in and hands H.264 out.
	 */
	if (encodes_h264 && takes_nv12)
		return KVMV_ROLE_ENCODER;
	/* The WAVE420L's node: NV12 in, H.265 out. */
	if (encodes_hevc && takes_nv12)
		return KVMV_ROLE_ENCODER_HEVC;
	return KVMV_ROLE_NONE;
}

int kvmv_subdev_name_matches(const char *name)
{
	return name != NULL && (strstr(name, "lt6911") != NULL ||
				strstr(name, "LT6911") != NULL);
}

void kvmv_rate_reset(struct kvmv_rate *rate)
{
	memset(rate, 0, sizeof(*rate));
}

int kvmv_rate_add(struct kvmv_rate *rate, uint64_t now_ms, size_t bytes,
		  unsigned int window_ms, unsigned int *kbps, unsigned int *fps_x10)
{
	uint64_t elapsed;

	if (rate->frames == 0 && rate->bytes == 0)
		rate->window_start_ms = now_ms;
	rate->bytes += bytes;
	rate->frames++;
	elapsed = now_ms - rate->window_start_ms;
	if (elapsed < window_ms || elapsed == 0)
		return 0;
	*kbps = (unsigned int)(rate->bytes * 8U / elapsed);
	*fps_x10 = (unsigned int)((uint64_t)rate->frames * 10000U / elapsed);
	rate->bytes = 0;
	rate->frames = 0;
	return 1;
}

int kvmv_encoder_fps(int asked, int capture, int delivered, int current)
{
	int target = asked;
	int want;
	int diff;

	if (capture > 0 && capture < target)
		target = capture;
	if (target < 1)
		target = 1;
	want = target;
	if (delivered > 0 && delivered * 10 < target * 9)
		want = delivered;
	if (want == target || current <= 0)
		return want;
	diff = want > current ? want - current : current - want;
	if (diff * 10 <= current)
		return current;
	return want;
}

enum kvmv_prime kvmv_prime_mode(const char *env, int hevc)
{
	if (env != NULL && env[0] >= '0' && env[0] <= '2' && env[1] == 0)
		return (enum kvmv_prime)(env[0] - '0');
	(void)hevc;
	return KVMV_PRIME_NONE;
}

int kvmv_receiver_fresh(uint64_t seen_ms, uint64_t now_ms, unsigned int max_age_ms)
{
	if (max_age_ms == 0 || seen_ms == 0 || seen_ms > now_ms)
		return 0;
	return now_ms - seen_ms <= max_age_ms;
}
