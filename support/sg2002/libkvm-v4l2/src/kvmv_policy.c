#include "kvmv_policy.h"

#include <errno.h>
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

int kvmv_plan_output(unsigned int src_width, unsigned int src_height,
		     unsigned int req_width, unsigned int req_height,
		     struct kvmv_plan *plan)
{
	unsigned int width, height;

	memset(plan, 0, sizeof(*plan));
	if (src_width < 16 || src_height < 16)
		return -1;

	if (req_width == 0 || req_height == 0) {
		width = src_width;
		height = src_height;
	} else {
		width = req_width > src_width ? src_width : req_width;
		height = req_height > src_height ? src_height : req_height;
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
			       int encodes_h264, int takes_nv12)
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
	/*
	 * The Coda registers several nodes (encoder, decoder, JPEG). The one
	 * wanted takes NV12 in and hands H.264 out.
	 */
	if (encodes_h264 && takes_nv12)
		return KVMV_ROLE_ENCODER;
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
