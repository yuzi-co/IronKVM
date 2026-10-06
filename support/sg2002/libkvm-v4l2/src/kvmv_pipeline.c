#define _GNU_SOURCE
#include "kvmv_pipeline.h"
#include "kvmv_hwjpeg.h"

#include <dirent.h>
#include <errno.h>
#include <fcntl.h>
#include <linux/dma-buf.h>
#include <linux/dma-heap.h>
#include <linux/v4l2-subdev.h>
#include <poll.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/ioctl.h>
#include <sys/mman.h>
#include <time.h>
#include <unistd.h>

/* The bitstream buffer asked for. The bridge used the same 1 MiB. */
#define KVMV_BITSTREAM_SIZE (1024U * 1024U)
#define KVMV_PRIME_TIMEOUT_MS 2000U
#define KVMV_M2M_TIMEOUT_MS 1000U

static const char *const heap_paths[] = {
	"/dev/dma_heap/default_cma_region",
	"/dev/dma_heap/linux,cma",
	"/dev/dma_heap/reserved",
	"/dev/dma_heap/system",
};

static int default_ioctl(int fd, unsigned long request, void *arg)
{
	int ret;

	do {
		ret = ioctl(fd, request, arg);
	} while (ret == -1 && errno == EINTR);
	return ret;
}

int (*kvmv_ioctl_hook)(int fd, unsigned long request, void *arg) = default_ioctl;

static int xioctl(int fd, unsigned long request, void *arg)
{
	return kvmv_ioctl_hook(fd, request, arg);
}

static uint64_t now_ms(void)
{
	struct timespec ts;

	clock_gettime(CLOCK_MONOTONIC, &ts);
	return (uint64_t)ts.tv_sec * 1000U + (uint64_t)ts.tv_nsec / 1000000U;
}

static int fail(struct kvmv_pipe *p, const char *step)
{
	int saved = errno;

	snprintf(p->error, sizeof(p->error), "%s: %s", step,
		 saved ? strerror(saved) : "rejected");
	errno = saved;
	return -1;
}

static int fail_msg(struct kvmv_pipe *p, const char *fmt, ...)
	__attribute__((format(printf, 2, 3)));

static int fail_msg(struct kvmv_pipe *p, const char *fmt, ...)
{
	va_list args;

	va_start(args, fmt);
	vsnprintf(p->error, sizeof(p->error), fmt, args);
	va_end(args);
	errno = EINVAL;
	return -1;
}

/* ---- discovery ----------------------------------------------------- */

static int queue_offers(int fd, enum v4l2_buf_type type, uint32_t fourcc)
{
	struct v4l2_fmtdesc desc;
	unsigned int i;

	for (i = 0; i < 64; i++) {
		memset(&desc, 0, sizeof(desc));
		desc.index = i;
		desc.type = type;
		if (xioctl(fd, VIDIOC_ENUM_FMT, &desc))
			return 0;
		if (desc.pixelformat == fourcc)
			return 1;
	}
	return 0;
}

static void copy_path(char *dst, const char *src)
{
	snprintf(dst, KVMV_PATH_MAX, "%s", src);
}

int kvmv_find_subdev(char *path, size_t size)
{
	const char *env = getenv("KVMV_SUBDEV");
	DIR *dir;
	struct dirent *entry;
	int found = -1;

	if (env != NULL && *env) {
		snprintf(path, size, "%s", env);
		return 0;
	}
	path[0] = 0;
	dir = opendir("/sys/class/video4linux");
	if (dir == NULL)
		return -1;
	while ((entry = readdir(dir)) != NULL) {
		char name_path[300];
		char name[64] = { 0 };
		FILE *fp;

		if (strncmp(entry->d_name, "v4l-subdev", 10) != 0)
			continue;
		snprintf(name_path, sizeof(name_path),
			 "/sys/class/video4linux/%s/name", entry->d_name);
		fp = fopen(name_path, "re");
		if (fp == NULL)
			continue;
		if (fgets(name, sizeof(name), fp) == NULL)
			name[0] = 0;
		fclose(fp);
		if (kvmv_subdev_name_matches(name)) {
			snprintf(path, size, "/dev/%s", entry->d_name);
			found = 0;
			break;
		}
	}
	closedir(dir);
	return found;
}

int kvmv_find_devices(struct kvmv_devices *devices, char *missing,
		      size_t missing_size)
{
	const char *env;
	unsigned int n;

	memset(devices, 0, sizeof(*devices));

	for (n = 0; n < 64; n++) {
		char path[KVMV_PATH_MAX];
		struct v4l2_capability cap;
		uint32_t caps;
		enum kvmv_role role;
		int fd;

		snprintf(path, sizeof(path), "/dev/video%u", n);
		fd = open(path, O_RDWR | O_NONBLOCK | O_CLOEXEC);
		if (fd < 0)
			continue;
		memset(&cap, 0, sizeof(cap));
		if (xioctl(fd, VIDIOC_QUERYCAP, &cap)) {
			close(fd);
			continue;
		}
		caps = (cap.capabilities & V4L2_CAP_DEVICE_CAPS) ?
			cap.device_caps : cap.capabilities;
		role = kvmv_match_role((const char *)cap.driver, caps,
				       (caps & V4L2_CAP_VIDEO_M2M) &&
				       queue_offers(fd, V4L2_BUF_TYPE_VIDEO_CAPTURE,
						    V4L2_PIX_FMT_H264),
				       (caps & V4L2_CAP_VIDEO_M2M) &&
				       queue_offers(fd, V4L2_BUF_TYPE_VIDEO_CAPTURE,
						    V4L2_PIX_FMT_HEVC),
				       (caps & V4L2_CAP_VIDEO_M2M) &&
				       queue_offers(fd, V4L2_BUF_TYPE_VIDEO_OUTPUT,
						    V4L2_PIX_FMT_NV12));
		close(fd);

		if (role == KVMV_ROLE_CAPTURE && !devices->capture[0])
			copy_path(devices->capture, path);
		else if (role == KVMV_ROLE_SCALER && !devices->scaler[0])
			copy_path(devices->scaler, path);
		else if (role == KVMV_ROLE_ENCODER && !devices->encoder[0])
			copy_path(devices->encoder, path);
		else if (role == KVMV_ROLE_ENCODER_HEVC && !devices->encoder_hevc[0])
			copy_path(devices->encoder_hevc, path);
		else if (role == KVMV_ROLE_JPEG && !devices->jpeg[0])
			copy_path(devices->jpeg, path);
	}

	if ((env = getenv("KVMV_CAPTURE_DEV")) != NULL && *env)
		copy_path(devices->capture, env);
	if ((env = getenv("KVMV_SCALER_DEV")) != NULL && *env)
		copy_path(devices->scaler, env);
	if ((env = getenv("KVMV_ENCODER_DEV")) != NULL && *env)
		copy_path(devices->encoder, env);
	if ((env = getenv("KVMV_HEVC_ENCODER_DEV")) != NULL && *env)
		copy_path(devices->encoder_hevc, env);
	if ((env = getenv("KVMV_JPEG_DEV")) != NULL && *env)
		copy_path(devices->jpeg, env);
	kvmv_find_subdev(devices->subdev, sizeof(devices->subdev));

	if (devices->capture[0] && devices->scaler[0] &&
	    (devices->encoder[0] || devices->encoder_hevc[0]))
		return 0;
	snprintf(missing, missing_size, "%s%s%s",
		 devices->capture[0] ? "" : "capture (sg2002-capture) ",
		 devices->scaler[0] ? "" : "scaler (sg2002-vpss) ",
		 devices->encoder[0] || devices->encoder_hevc[0] ? "" :
		 "encoder (coda, NV12 to H.264, or wave420l, NV12 to H.265)");
	return -1;
}

enum kvmv_signal kvmv_query_signal(int subdev_fd, unsigned int need_width,
				   unsigned int need_height,
				   struct v4l2_dv_timings *out)
{
	struct v4l2_dv_timings timings;
	int ret;

	if (subdev_fd < 0)
		return KVMV_SIGNAL_UNKNOWN;
	memset(&timings, 0, sizeof(timings));
	/* VIDIOC_SUBDEV_QUERY_DV_TIMINGS is the same request number. */
	ret = xioctl(subdev_fd, VIDIOC_SUBDEV_QUERY_DV_TIMINGS, &timings);
	if (out != NULL)
		*out = timings;
	return kvmv_classify_timings(ret, ret ? errno : 0, &timings,
				     need_width, need_height);
}

/* ---- small V4L2 helpers -------------------------------------------- */

static int get_fmt(int fd, enum v4l2_buf_type type, struct v4l2_pix_format *pix)
{
	struct v4l2_format format;

	memset(&format, 0, sizeof(format));
	format.type = type;
	if (xioctl(fd, VIDIOC_G_FMT, &format))
		return -1;
	*pix = format.fmt.pix;
	return 0;
}

static int set_fmt(int fd, enum v4l2_buf_type type, uint32_t fourcc,
		   unsigned int width, unsigned int height,
		   const struct v4l2_pix_format *colour, uint32_t sizeimage,
		   struct v4l2_pix_format *actual)
{
	struct v4l2_format format;

	memset(&format, 0, sizeof(format));
	format.type = type;
	format.fmt.pix.width = width;
	format.fmt.pix.height = height;
	format.fmt.pix.pixelformat = fourcc;
	format.fmt.pix.field = V4L2_FIELD_NONE;
	format.fmt.pix.sizeimage = sizeimage;
	if (colour != NULL) {
		/* Without the magic the V4L2 core zeroes ycbcr_enc,
		 * quantization and xfer_func of a single-planar format before
		 * the driver sees them (v4l_sanitize_format). */
		format.fmt.pix.priv = V4L2_PIX_FMT_PRIV_MAGIC;
		format.fmt.pix.colorspace = colour->colorspace;
		format.fmt.pix.xfer_func = colour->xfer_func;
		format.fmt.pix.ycbcr_enc = colour->ycbcr_enc;
		format.fmt.pix.quantization = colour->quantization;
		/* A queue that reads the capture buffer as it is takes the
		 * capture's line too: the capture pads it to 16 bytes, which a
		 * 1366 pixel wide mode needs (ironkvm-dist patch 0937). */
		if (V4L2_TYPE_IS_OUTPUT(type) && colour->pixelformat == fourcc)
			format.fmt.pix.bytesperline = colour->bytesperline;
	}
	if (xioctl(fd, VIDIOC_S_FMT, &format))
		return -1;
	*actual = format.fmt.pix;
	return 0;
}

static int set_crop(int fd, enum v4l2_buf_type type, unsigned int width,
		    unsigned int height)
{
	struct v4l2_selection selection;

	memset(&selection, 0, sizeof(selection));
	selection.type = type;
	selection.target = V4L2_SEL_TGT_CROP;
	selection.r.width = width;
	selection.r.height = height;
	return xioctl(fd, VIDIOC_S_SELECTION, &selection);
}

static int set_ctrl(int fd, uint32_t id, int32_t value)
{
	struct v4l2_ext_control control;
	struct v4l2_ext_controls list;

	memset(&control, 0, sizeof(control));
	memset(&list, 0, sizeof(list));
	control.id = id;
	control.value = value;
	list.which = V4L2_CTRL_WHICH_CUR_VAL;
	list.count = 1;
	list.controls = &control;
	return xioctl(fd, VIDIOC_S_EXT_CTRLS, &list);
}

static int get_ctrl(int fd, uint32_t id, int32_t *value)
{
	struct v4l2_ext_control control;
	struct v4l2_ext_controls list;

	memset(&control, 0, sizeof(control));
	memset(&list, 0, sizeof(list));
	control.id = id;
	list.which = V4L2_CTRL_WHICH_CUR_VAL;
	list.count = 1;
	list.controls = &control;
	if (xioctl(fd, VIDIOC_G_EXT_CTRLS, &list))
		return -1;
	*value = control.value;
	return 0;
}

/*
 * value held to the control's range where the driver reports one: the Coda
 * takes a GOP of at most 99, the ABI up to 100.
 */
static int32_t ctrl_fit(int fd, uint32_t id, int32_t value)
{
	struct v4l2_queryctrl query;

	memset(&query, 0, sizeof(query));
	query.id = id;
	if (xioctl(fd, VIDIOC_QUERYCTRL, &query))
		return value;
	if (value > query.maximum)
		return query.maximum;
	if (value < query.minimum)
		return query.minimum;
	return value;
}

static int is_hevc(const struct kvmv_pipe *p)
{
	return p->codec == KVMV_CODEC_KIND_HEVC;
}

static uint32_t coded_fourcc(const struct kvmv_pipe *p)
{
	return is_hevc(p) ? V4L2_PIX_FMT_HEVC : V4L2_PIX_FMT_H264;
}

static uint32_t min_qp_ctrl(const struct kvmv_pipe *p)
{
	return is_hevc(p) ? V4L2_CID_MPEG_VIDEO_HEVC_MIN_QP : V4L2_CID_MPEG_VIDEO_H264_MIN_QP;
}

static uint32_t max_qp_ctrl(const struct kvmv_pipe *p)
{
	return is_hevc(p) ? V4L2_CID_MPEG_VIDEO_HEVC_MAX_QP : V4L2_CID_MPEG_VIDEO_H264_MAX_QP;
}

static void read_back(struct kvmv_pipe *p)
{
	struct v4l2_streamparm parm;
	int32_t value;

	p->applied.bitrate_bps = -1;
	p->applied.gop = -1;
	p->applied.fps_numerator = -1;
	p->applied.fps_denominator = -1;
	if (!get_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_BITRATE, &value))
		p->applied.bitrate_bps = value;
	if (!get_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_GOP_SIZE, &value))
		p->applied.gop = value;
	p->applied.min_qp = -1;
	p->applied.max_qp = -1;
	if (!get_ctrl(p->enc_fd, min_qp_ctrl(p), &value))
		p->applied.min_qp = value;
	if (!get_ctrl(p->enc_fd, max_qp_ctrl(p), &value))
		p->applied.max_qp = value;
	memset(&parm, 0, sizeof(parm));
	parm.type = V4L2_BUF_TYPE_VIDEO_OUTPUT;
	if (!xioctl(p->enc_fd, VIDIOC_G_PARM, &parm) &&
	    parm.parm.output.timeperframe.numerator) {
		/* timeperframe is seconds per frame, so the rate is its inverse. */
		p->applied.fps_numerator =
			(int)parm.parm.output.timeperframe.denominator;
		p->applied.fps_denominator =
			(int)parm.parm.output.timeperframe.numerator;
	}
}

static int set_frame_rate(struct kvmv_pipe *p, unsigned int fps)
{
	struct v4l2_streamparm parm;

	memset(&parm, 0, sizeof(parm));
	parm.type = V4L2_BUF_TYPE_VIDEO_OUTPUT;
	parm.parm.output.timeperframe.numerator = 1;
	parm.parm.output.timeperframe.denominator = fps;
	return xioctl(p->enc_fd, VIDIOC_S_PARM, &parm);
}

/* ---- negotiation --------------------------------------------------- */

void kvmv_pipe_init(struct kvmv_pipe *p)
{
	memset(p, 0, sizeof(*p));
	p->cap_fd = -1;
	p->vpss_fd = -1;
	p->enc_fd = -1;
	p->heap_fd = -1;
	p->snap.fd = -1;
	p->snap.buf.fd = -1;
	p->applied.bitrate_bps = -1;
	p->applied.gop = -1;
	p->applied.fps_numerator = -1;
	p->applied.fps_denominator = -1;
	p->applied.min_qp = -1;
	p->applied.max_qp = -1;
	p->ahead_mid = -1;
}

int kvmv_pipe_negotiate(struct kvmv_pipe *p, const struct kvmv_pipe_cfg *cfg)
{
	const struct v4l2_pix_format *cap = &p->cap_fmt;
	unsigned int bytes_per_pixel;

	if (get_fmt(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, &p->cap_fmt))
		return fail(p, "capture G_FMT");
	switch (cap->pixelformat) {
	case V4L2_PIX_FMT_UYVY:
	case V4L2_PIX_FMT_YUYV:
		bytes_per_pixel = 2;
		break;
	case V4L2_PIX_FMT_NV12:
	case V4L2_PIX_FMT_NV21:
		bytes_per_pixel = 1;
		break;
	default:
		return fail_msg(p, "capture delivers %.4s, not UYVY, YUYV, NV12 or NV21",
				(const char *)&cap->pixelformat);
	}
	if (cap->bytesperline < cap->width * bytes_per_pixel || !cap->sizeimage)
		return fail_msg(p, "capture stride %u is short for %ux%u",
				cap->bytesperline, cap->width, cap->height);
	if (kvmv_plan_output(cap->width, cap->height, cfg->req_width,
			     cfg->req_height, cfg->keep_shape, &p->plan))
		return fail_msg(p, "capture size %ux%u is unusable", cap->width,
				cap->height);

	/* The encoder first: its padded input surface dictates the layout the
	 * scaler has to write. A parked encoder holds its buffers, so its
	 * queues refuse S_FMT; it was parked with these formats, for this
	 * plan (take_parked_encoder). */
	if (p->enc_reused)
		goto rate_control;
	/*
	 * H.265 at the capture's size: the encoder may read the capture buffer
	 * as it is. It must keep the capture's format, size and line, or the
	 * scaler converts as before. A WAVE420L without ironkvm-dist patch
	 * 0933 answers NV12.
	 */
	p->direct = 0;
	if (cfg->direct && is_hevc(p) && bytes_per_pixel == 2 &&
	    p->plan.out_width == cap->width && p->plan.out_height == cap->height) {
		if (set_fmt(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, cap->pixelformat,
			    cap->width, cap->height, cap, 0, &p->enc_out_fmt))
			return fail(p, "encoder OUTPUT S_FMT");
		p->direct = p->enc_out_fmt.pixelformat == cap->pixelformat &&
			    p->enc_out_fmt.width == cap->width &&
			    p->enc_out_fmt.height == cap->height &&
			    p->enc_out_fmt.bytesperline == cap->bytesperline;
	}
	if (!p->direct) {
		if (set_fmt(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_PIX_FMT_NV12,
			    p->plan.out_width, p->plan.coded_height, cap, 0,
			    &p->enc_out_fmt))
			return fail(p, "encoder OUTPUT S_FMT");
		if (set_crop(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, p->plan.out_width,
			     p->plan.out_height)) {
			if (errno != EINVAL || p->plan.coded_height != p->plan.out_height)
				return fail(p, "encoder OUTPUT crop");
		}
		if (get_fmt(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, &p->enc_out_fmt))
			return fail(p, "encoder OUTPUT G_FMT");
		if (p->enc_out_fmt.pixelformat != V4L2_PIX_FMT_NV12 ||
		    p->enc_out_fmt.width < p->plan.out_width ||
		    p->enc_out_fmt.height < p->plan.out_height)
			return fail_msg(p, "encoder took %.4s %ux%u for NV12 %ux%u",
					(const char *)&p->enc_out_fmt.pixelformat,
					p->enc_out_fmt.width, p->enc_out_fmt.height,
					p->plan.out_width, p->plan.coded_height);
	}
	if (set_fmt(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, coded_fourcc(p),
		    p->plan.out_width, p->plan.out_height, cap,
		    KVMV_BITSTREAM_SIZE, &p->enc_cap_fmt))
		return fail(p, "encoder CAPTURE S_FMT");
	if (p->enc_cap_fmt.pixelformat != coded_fourcc(p))
		return fail_msg(p, "encoder does not produce %s",
				is_hevc(p) ? "H.265" : "H.264");

rate_control:
	/*
	 * Rate control. The frame rate matters to it: the encoder sizes a frame
	 * as bitrate divided by frame rate, and the bridge never set the rate,
	 * which is one candidate for the overshoot in issue #35. Each control is
	 * set on its own so one the driver lacks does not cost the others; what
	 * the encoder actually holds is read back below.
	 */
	if (cfg->fps)
		set_frame_rate(p, cfg->fps);
	set_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_FRAME_RC_ENABLE, 1);
	set_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_BITRATE_MODE,
		 V4L2_MPEG_VIDEO_BITRATE_MODE_CBR);
	set_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_HEADER_MODE,
		 V4L2_MPEG_VIDEO_HEADER_MODE_JOINED_WITH_1ST_FRAME);
	if (cfg->bitrate_bps)
		set_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_BITRATE,
			 (int32_t)cfg->bitrate_bps);
	if (cfg->gop)
		set_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_GOP_SIZE,
			 ctrl_fit(p->enc_fd, V4L2_CID_MPEG_VIDEO_GOP_SIZE,
				  (int32_t)cfg->gop));
	/*
	 * Picture quality (#35). Kernels before patch 0910 lack the minimum
	 * QP control on the Coda980 and ignore the maximum; each is set on
	 * its own, and what sticks is read back. MAX first, so a minimum
	 * above the encoder's current maximum is not refused. The WAVE420L
	 * takes the HEVC controls.
	 */
	if (cfg->max_qp)
		set_ctrl(p->enc_fd, max_qp_ctrl(p), cfg->max_qp);
	if (cfg->min_qp || cfg->max_qp)
		set_ctrl(p->enc_fd, min_qp_ctrl(p), cfg->min_qp);
	if (cfg->vbv_delay_ms)
		set_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_VBV_DELAY,
			 (int32_t)cfg->vbv_delay_ms);
	read_back(p);
	if (p->direct)
		return 0;

	/* The scaler: OUTPUT is the captured frame, CAPTURE is the encoder's
	 * input surface. */
	if (set_fmt(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, cap->pixelformat,
		    cap->width, cap->height, cap, 0, &p->vpss_in_fmt))
		return fail(p, "scaler OUTPUT S_FMT");
	if (p->vpss_in_fmt.bytesperline != cap->bytesperline ||
	    p->vpss_in_fmt.width != cap->width ||
	    p->vpss_in_fmt.height != cap->height)
		return fail_msg(p, "scaler wants %ux%u stride %u, capture gives %ux%u stride %u",
				p->vpss_in_fmt.width, p->vpss_in_fmt.height,
				p->vpss_in_fmt.bytesperline, cap->width,
				cap->height, cap->bytesperline);
	if (set_fmt(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_PIX_FMT_NV12,
		    p->enc_out_fmt.width, p->enc_out_fmt.height, cap, 0,
		    &p->vpss_out_fmt))
		return fail(p, "scaler CAPTURE S_FMT");
	if (p->vpss_out_fmt.pixelformat != V4L2_PIX_FMT_NV12 ||
	    p->vpss_out_fmt.bytesperline != p->enc_out_fmt.bytesperline ||
	    p->vpss_out_fmt.height != p->enc_out_fmt.height ||
	    p->vpss_out_fmt.sizeimage > p->enc_out_fmt.sizeimage)
		return fail_msg(p, "scaler surface %ux%u stride %u size %u does not match encoder %ux%u stride %u size %u",
				p->vpss_out_fmt.width, p->vpss_out_fmt.height,
				p->vpss_out_fmt.bytesperline,
				p->vpss_out_fmt.sizeimage,
				p->enc_out_fmt.width, p->enc_out_fmt.height,
				p->enc_out_fmt.bytesperline,
				p->enc_out_fmt.sizeimage);
	if (set_crop(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, p->plan.out_width,
		     p->plan.out_height)) {
		/* Without a crop the scaler fills the whole surface, which is
		 * only right when the surface has no padding. */
		if (errno != EINVAL || p->enc_out_fmt.height != p->plan.out_height ||
		    p->enc_out_fmt.width != p->plan.out_width)
			return fail(p, "scaler CAPTURE crop");
	}
	return 0;
}

/* ---- buffers ------------------------------------------------------- */

static int request(int fd, enum v4l2_buf_type type, enum v4l2_memory memory,
		   unsigned int count)
{
	struct v4l2_requestbuffers req;

	memset(&req, 0, sizeof(req));
	req.count = count;
	req.type = type;
	req.memory = memory;
	if (xioctl(fd, VIDIOC_REQBUFS, &req))
		return -1;
	return (int)req.count;
}

static int qbuf(int fd, enum v4l2_buf_type type, enum v4l2_memory memory,
		unsigned int index, int dmabuf_fd, unsigned int bytesused)
{
	struct v4l2_buffer buf;

	memset(&buf, 0, sizeof(buf));
	buf.type = type;
	buf.memory = memory;
	buf.index = index;
	buf.bytesused = bytesused;
	if (memory == V4L2_MEMORY_DMABUF)
		buf.m.fd = dmabuf_fd;
	return xioctl(fd, VIDIOC_QBUF, &buf);
}

static int dqbuf(int fd, enum v4l2_buf_type type, enum v4l2_memory memory,
		 struct v4l2_buffer *buf)
{
	memset(buf, 0, sizeof(*buf));
	buf->type = type;
	buf->memory = memory;
	return xioctl(fd, VIDIOC_DQBUF, buf);
}

static int stream(int fd, enum v4l2_buf_type type, int on)
{
	return xioctl(fd, on ? VIDIOC_STREAMON : VIDIOC_STREAMOFF, &type);
}

/* Dequeue one buffer, waiting up to timeout_ms. ETIMEDOUT on timeout. */
static int wait_dqbuf(int fd, enum v4l2_buf_type type, enum v4l2_memory memory,
		      short events, unsigned int timeout_ms,
		      struct v4l2_buffer *buf)
{
	uint64_t deadline = now_ms() + timeout_ms;

	for (;;) {
		struct pollfd pfd = { .fd = fd, .events = events };
		uint64_t now;
		int ret;

		if (dqbuf(fd, type, memory, buf) == 0)
			return 0;
		if (errno != EAGAIN)
			return -1;
		now = now_ms();
		if (now >= deadline) {
			errno = ETIMEDOUT;
			return -1;
		}
		ret = poll(&pfd, 1, (int)(deadline - now));
		if (ret < 0 && errno != EINTR)
			return -1;
		if (ret > 0 && (pfd.revents & POLLERR) && !(pfd.revents & events)) {
			/* One more try: the error may race a completion. */
			if (dqbuf(fd, type, memory, buf) == 0)
				return 0;
			if (errno == EAGAIN)
				errno = EIO;
			return -1;
		}
	}
}

static int open_heap(void)
{
	unsigned int i;

	for (i = 0; i < sizeof(heap_paths) / sizeof(heap_paths[0]); i++) {
		int fd = open(heap_paths[i], O_RDONLY | O_CLOEXEC);

		if (fd >= 0)
			return fd;
	}
	return -1;
}

static int sync_dmabuf(int fd, uint64_t flags)
{
	struct dma_buf_sync sync = { .flags = flags };

	return xioctl(fd, DMA_BUF_IOCTL_SYNC, &sync);
}

static int alloc_mid(struct kvmv_pipe *p, unsigned int wanted,
		     unsigned int capture_wanted)
{
	unsigned int i;
	int got;

	if (p->direct) {
		/*
		 * No middle buffers: the encoder's OUTPUT queue takes the
		 * capture buffers' dma-bufs, slot i for capture buffer i. The
		 * heap stays open for the snapshot context.
		 */
		if (!p->enc_reused) {
			got = request(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT,
				      V4L2_MEMORY_DMABUF, capture_wanted);
			if (got < 1)
				return fail(p, "encoder OUTPUT REQBUFS");
			p->enc_out_count = (unsigned int)got;
		}
		if (p->heap_fd < 0)
			p->heap_fd = open_heap();
		if (p->heap_fd < 0)
			return fail(p, "open /dev/dma_heap");
		return 0;
	}
	if (p->enc_reused) {
		/* The buffers and the encoder's OUTPUT queue came parked; only
		 * the new scaler context needs its queue. */
		got = request(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE,
			      V4L2_MEMORY_DMABUF, p->mid_count);
		if (got != (int)p->mid_count)
			return fail(p, "scaler CAPTURE REQBUFS");
		return 0;
	}
	got = request(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE,
		      V4L2_MEMORY_DMABUF, wanted);
	if (got < 1)
		return fail(p, "scaler CAPTURE REQBUFS");
	p->mid_count = (unsigned int)got;
	got = request(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF,
		      p->mid_count);
	if (got < (int)p->mid_count)
		return fail(p, "encoder OUTPUT REQBUFS");

	p->heap_fd = open_heap();
	if (p->heap_fd < 0)
		return fail(p, "open /dev/dma_heap");
	p->mid = calloc(p->mid_count, sizeof(*p->mid));
	if (p->mid == NULL)
		return fail(p, "calloc");
	for (i = 0; i < p->mid_count; i++)
		p->mid[i].fd = -1;
	for (i = 0; i < p->mid_count; i++) {
		struct dma_heap_allocation_data alloc;

		memset(&alloc, 0, sizeof(alloc));
		alloc.len = p->enc_out_fmt.sizeimage;
		alloc.fd_flags = O_CLOEXEC | O_RDWR;
		if (xioctl(p->heap_fd, DMA_HEAP_IOCTL_ALLOC, &alloc))
			return fail(p, "DMA_HEAP_IOCTL_ALLOC");
		p->mid[i].fd = (int)alloc.fd;
		p->mid[i].length = p->enc_out_fmt.sizeimage;
	}
	return 0;
}

static int alloc_bitstream(struct kvmv_pipe *p, unsigned int wanted)
{
	unsigned int i;
	int got;

	if (p->enc_reused) {
		/* Mapped already; STREAMOFF gave them back to user space. */
		for (i = 0; i < p->bs_count; i++)
			if (qbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE,
				 V4L2_MEMORY_MMAP, i, -1, 0))
				return fail(p, "encoder CAPTURE QBUF");
		return 0;
	}
	got = request(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
		      wanted);
	if (got < 1)
		return fail(p, "encoder CAPTURE REQBUFS");
	p->bs = calloc((size_t)got, sizeof(*p->bs));
	if (p->bs == NULL)
		return fail(p, "calloc");
	p->bs_count = (unsigned int)got;
	for (i = 0; i < p->bs_count; i++) {
		struct v4l2_buffer buf;

		p->bs[i].fd = -1;
		memset(&buf, 0, sizeof(buf));
		buf.type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
		buf.memory = V4L2_MEMORY_MMAP;
		buf.index = i;
		if (xioctl(p->enc_fd, VIDIOC_QUERYBUF, &buf))
			return fail(p, "encoder CAPTURE QUERYBUF");
		p->bs[i].addr = mmap(NULL, buf.length, PROT_READ, MAP_SHARED,
				     p->enc_fd, buf.m.offset);
		if (p->bs[i].addr == MAP_FAILED) {
			p->bs[i].addr = NULL;
			return fail(p, "encoder CAPTURE mmap");
		}
		p->bs[i].length = buf.length;
		if (qbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
			 i, -1, 0))
			return fail(p, "encoder CAPTURE QBUF");
	}
	return 0;
}

static int alloc_capture(struct kvmv_pipe *p, unsigned int wanted)
{
	unsigned int i;
	int got;

	got = request(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
		      wanted);
	if (got < 1)
		return fail(p, "capture REQBUFS");
	p->cap = calloc((size_t)got, sizeof(*p->cap));
	if (p->cap == NULL)
		return fail(p, "calloc");
	p->cap_count = (unsigned int)got;
	for (i = 0; i < p->cap_count; i++)
		p->cap[i].fd = -1;
	for (i = 0; i < p->cap_count; i++) {
		struct v4l2_exportbuffer exp;

		memset(&exp, 0, sizeof(exp));
		exp.type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
		exp.index = i;
		exp.flags = O_CLOEXEC | O_RDWR;
		if (xioctl(p->cap_fd, VIDIOC_EXPBUF, &exp))
			return fail(p, "capture EXPBUF");
		p->cap[i].fd = exp.fd;
		p->cap[i].length = p->cap_fmt.sizeimage;
		if (qbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
			 i, -1, 0))
			return fail(p, "capture QBUF");
	}
	return 0;
}

/*
 * Stream the encoder on (see enum kvmv_prime). With a priming picture, one
 * black picture goes in before any live frame, as the bridge does: on the
 * Coda980 that is what starts the sequence (its OUTPUT queue waits for a
 * picture), which claims the encoder's working memory. The capture buffers
 * are allocated before this (kvmv_pipe_start). Its output is thrown away;
 * the caller asks for a keyframe for the first live frame.
 */
static uint64_t now_us(void);

static int prime_encoder(struct kvmv_pipe *p, enum kvmv_prime mode,
			 uint64_t start0)
{
	size_t luma = (size_t)p->enc_out_fmt.bytesperline * p->enc_out_fmt.height;
	size_t frame = luma * 3 / 2;
	struct v4l2_buffer buf;
	uint8_t *map;

	/* The first scale goes into mid[1] while mid[0] is being encoded. */
	if (mode == KVMV_PRIME_ASYNC && p->mid_count < 2)
		mode = KVMV_PRIME_WAIT;
	/* Reading the capture buffers, the encoder has no surface of its own
	 * to prime with (and no priming is the default). */
	if (p->direct)
		mode = KVMV_PRIME_NONE;
	if (mode == KVMV_PRIME_NONE) {
		if (stream(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, 1))
			return fail(p, "encoder CAPTURE STREAMON");
		p->enc_cap_on = 1;
		if (stream(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, 1))
			return fail(p, "encoder OUTPUT STREAMON");
		p->enc_out_on = 1;
		p->start_us[5] = (uint32_t)(now_us() - start0);
		return 0;
	}

	if (frame > p->mid[0].length)
		return fail_msg(p, "encoder surface %zu exceeds its buffer %zu",
				frame, p->mid[0].length);
	map = mmap(NULL, p->mid[0].length, PROT_READ | PROT_WRITE, MAP_SHARED,
		   p->mid[0].fd, 0);
	if (map == MAP_FAILED)
		return fail(p, "mmap middle buffer");
	sync_dmabuf(p->mid[0].fd, DMA_BUF_SYNC_START | DMA_BUF_SYNC_WRITE);
	memset(map, 16, luma);
	memset(map + luma, 128, frame - luma);
	sync_dmabuf(p->mid[0].fd, DMA_BUF_SYNC_END | DMA_BUF_SYNC_WRITE);
	munmap(map, p->mid[0].length);

	if (qbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF, 0,
		 p->mid[0].fd, p->enc_out_fmt.sizeimage))
		return fail(p, "encoder priming QBUF");
	if (stream(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, 1))
		return fail(p, "encoder CAPTURE STREAMON");
	p->enc_cap_on = 1;
	if (stream(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, 1))
		return fail(p, "encoder OUTPUT STREAMON");
	p->enc_out_on = 1;
	/* STREAMON has set the encoder up: its buffers and sequence. */
	p->start_us[5] = (uint32_t)(now_us() - start0);

	if (mode == KVMV_PRIME_ASYNC) {
		/* The encode runs while the capture and the scaler start; the
		 * first kvmv_pipe_encode_mid collects it. */
		p->prime_pending = 1;
		p->next_mid = 1;
		return 0;
	}
	if (wait_dqbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
		       POLLIN, KVMV_PRIME_TIMEOUT_MS, &buf))
		return fail(p, "encoder priming output");
	if (qbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
		 buf.index, -1, 0))
		return fail(p, "encoder CAPTURE requeue");
	if (wait_dqbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF,
		       POLLOUT, KVMV_M2M_TIMEOUT_MS, &buf))
		return fail(p, "encoder priming input");
	return 0;
}

static void subscribe_source_change(int fd)
{
	struct v4l2_event_subscription sub;

	memset(&sub, 0, sizeof(sub));
	sub.type = V4L2_EVENT_SOURCE_CHANGE;
	xioctl(fd, VIDIOC_SUBSCRIBE_EVENT, &sub);
}

/* The capture node and buffers kvmv_pipe_park kept, for the next start. */
static struct {
	int fd;
	struct kvmv_buf *bufs;
	unsigned int count;
	uint32_t sizeimage;
} parked = { .fd = -1 };

/*
 * The encoder node kvmv_pipe_park kept (cfg.park_encoder), streamed off, with
 * its bitstream buffers (mapped) and the middle buffers its OUTPUT queue has
 * imported. Its queues keep their formats, which hold for this capture format
 * and plan only.
 */
static struct {
	int fd;
	enum kvmv_codec codec;
	char path[KVMV_PATH_MAX];
	struct v4l2_pix_format cap_fmt;
	struct kvmv_plan plan;
	struct v4l2_pix_format enc_out_fmt, enc_cap_fmt;
	struct kvmv_buf *bs, *mid;
	unsigned int bs_count, mid_count;
	int heap_fd;
	/* kvmv_pipe_cfg.direct as asked, and what the pipe got. */
	int want_direct, direct;
	unsigned int enc_out_count;
} parked_enc = { .fd = -1, .heap_fd = -1 };

static void free_bufs(struct kvmv_buf **bufs, unsigned int *count);

static void unpark_encoder(void)
{
	/* Mappings first: the node's buffers live until both are gone. */
	free_bufs(&parked_enc.bs, &parked_enc.bs_count);
	free_bufs(&parked_enc.mid, &parked_enc.mid_count);
	if (parked_enc.fd >= 0)
		close(parked_enc.fd);
	if (parked_enc.heap_fd >= 0)
		close(parked_enc.heap_fd);
	parked_enc.fd = -1;
	parked_enc.heap_fd = -1;
}

void kvmv_pipe_unpark(void)
{
	free_bufs(&parked.bufs, &parked.count);
	if (parked.fd >= 0)
		close(parked.fd);
	parked.fd = -1;
	unpark_encoder();
}

static int same_pix(const struct v4l2_pix_format *a, const struct v4l2_pix_format *b)
{
	return a->width == b->width && a->height == b->height &&
	       a->pixelformat == b->pixelformat &&
	       a->bytesperline == b->bytesperline &&
	       a->sizeimage == b->sizeimage && a->colorspace == b->colorspace &&
	       a->ycbcr_enc == b->ycbcr_enc &&
	       a->quantization == b->quantization &&
	       a->xfer_func == b->xfer_func;
}

/*
 * Take the parked encoder when it was parked for this codec, node, capture
 * format, output plan and buffer counts; otherwise let it go, so its memory
 * is free before new buffers are allocated. Returns 1 when taken.
 */
static int take_parked_encoder(struct kvmv_pipe *p, const struct kvmv_pipe_cfg *cfg,
			       const char *encoder)
{
	struct v4l2_pix_format cap;
	struct kvmv_plan plan;

	if (parked_enc.fd < 0)
		return 0;
	if (!cfg->park_encoder || parked_enc.codec != cfg->codec ||
	    strcmp(parked_enc.path, encoder) != 0 ||
	    parked_enc.want_direct != cfg->direct ||
	    (!parked_enc.direct &&
	     parked_enc.mid_count != (cfg->mid_buffers ? cfg->mid_buffers : 2)) ||
	    parked_enc.bs_count != (cfg->bitstream_buffers ? cfg->bitstream_buffers : 3) ||
	    get_fmt(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, &cap) ||
	    !same_pix(&cap, &parked_enc.cap_fmt) ||
	    kvmv_plan_output(cap.width, cap.height, cfg->req_width,
			     cfg->req_height, cfg->keep_shape, &plan) ||
	    memcmp(&plan, &parked_enc.plan, sizeof(plan)) != 0) {
		unpark_encoder();
		return 0;
	}
	p->enc_fd = parked_enc.fd;
	p->heap_fd = parked_enc.heap_fd;
	p->bs = parked_enc.bs;
	p->bs_count = parked_enc.bs_count;
	p->mid = parked_enc.mid;
	p->mid_count = parked_enc.mid_count;
	p->enc_out_fmt = parked_enc.enc_out_fmt;
	p->enc_cap_fmt = parked_enc.enc_cap_fmt;
	p->direct = parked_enc.direct;
	p->enc_out_count = parked_enc.enc_out_count;
	parked_enc.fd = -1;
	parked_enc.heap_fd = -1;
	parked_enc.bs = NULL;
	parked_enc.mid = NULL;
	parked_enc.bs_count = 0;
	parked_enc.mid_count = 0;
	p->enc_reused = 1;
	return 1;
}

/* Queue the parked buffers on the capture node instead of allocating. */
static int adopt_capture(struct kvmv_pipe *p)
{
	unsigned int i;

	p->cap = parked.bufs;
	p->cap_count = parked.count;
	parked.bufs = NULL;
	parked.count = 0;
	for (i = 0; i < p->cap_count; i++)
		if (qbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
			 i, -1, 0))
			return fail(p, "capture QBUF");
	return 0;
}

static uint64_t now_us(void);
#define START_STAMP(i) (p->start_us[i] = (uint32_t)(now_us() - start0))

int kvmv_pipe_start(struct kvmv_pipe *p, const struct kvmv_devices *d,
		    const struct kvmv_pipe_cfg *cfg)
{
	const char *encoder;
	uint64_t start0 = now_us();

	kvmv_pipe_init(p);
	p->codec = cfg->codec;
	p->pickup_wait_max_us = cfg->pickup_wait_ms * 1000U;
	p->park_encoder = cfg->park_encoder;
	p->want_direct = cfg->direct;
	copy_path(p->scaler_path, d->scaler);
	encoder = is_hevc(p) ? d->encoder_hevc : d->encoder;
	if (!encoder[0]) {
		fail_msg(p, "no %s encoder", is_hevc(p) ? "H.265" : "H.264");
		errno = ENODEV;
		return -1;
	}
	copy_path(p->enc_path, encoder);

	if (parked.fd >= 0) {
		p->cap_fd = parked.fd;
		parked.fd = -1;
	} else {
		p->cap_fd = open(d->capture, O_RDWR | O_NONBLOCK | O_CLOEXEC);
	}
	if (p->cap_fd < 0) {
		fail(p, d->capture);
		goto out;
	}
	p->vpss_fd = open(d->scaler, O_RDWR | O_NONBLOCK | O_CLOEXEC);
	if (p->vpss_fd < 0) {
		fail(p, d->scaler);
		goto out;
	}
	/*
	 * The H.264 and H.265 encoders share the codec SRAM and take turns:
	 * the one that is not streaming refuses STREAMON with EBUSY.
	 */
	if (!take_parked_encoder(p, cfg, encoder))
		p->enc_fd = open(encoder, O_RDWR | O_NONBLOCK | O_CLOEXEC);
	if (p->enc_fd < 0) {
		fail(p, encoder);
		goto out;
	}
	subscribe_source_change(p->cap_fd);
	START_STAMP(0);

	/*
	 * Buffers kept by kvmv_pipe_park need no allocation, so the capture
	 * can start first. Its STREAMON brings the receiver and the CSI link
	 * up, and the first frames arrive while the encoder below is set up
	 * and primes, which takes several frame times. Started last, as on a
	 * first build, the pipeline waits for both one after the other.
	 * Capture before the scaler either way: while the CSI driver streams,
	 * the VIP fabric clocks are on and VPSS register access is safe.
	 */
	if (cfg->early_capture && parked.bufs &&
	    get_fmt(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, &p->cap_fmt) == 0 &&
	    parked.sizeimage == p->cap_fmt.sizeimage &&
	    parked.count == (cfg->capture_buffers ? cfg->capture_buffers : 2)) {
		if (adopt_capture(p))
			goto out;
		if (stream(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, 1)) {
			fail(p, "capture STREAMON");
			p->cap_on_failed = 1;
			goto out;
		}
		p->cap_on = 1;
		p->early = 1;
	}
	START_STAMP(1);

	if (kvmv_pipe_negotiate(p, cfg))
		goto out;
	START_STAMP(2);
	if (alloc_mid(p, cfg->mid_buffers ? cfg->mid_buffers : 2,
		      cfg->capture_buffers ? cfg->capture_buffers : 2))
		goto out;
	START_STAMP(3);
	if (alloc_bitstream(p, cfg->bitstream_buffers ? cfg->bitstream_buffers : 3))
		goto out;
	START_STAMP(4);
	/*
	 * Nothing parked that fits: the node was opened afresh, or holds
	 * buffers of another size, which REQBUFS replaces. Allocate before the
	 * encoder streams on. The capture buffers and the encoders' reference
	 * frames all come from video_pool, a 32 MiB device pool (a 1080p
	 * capture buffer is 4 MB; before ironkvm-dist patch 0940 the pool
	 * rounded each buffer up to a power of two). The WAVE420L takes its
	 * reference frames at STREAMON; allocated after them, on a first
	 * build with H.265, the capture got fewer buffers than it asked for
	 * (trial 25). Parked, that short set never matched the count again, so
	 * every later build freed it and came up short once more. This
	 * order is the one every later build has anyway: the parked capture
	 * buffers stay while the encoders come and go.
	 */
	if (!p->early &&
	    !(parked.bufs && parked.sizeimage == p->cap_fmt.sizeimage &&
	      parked.count == (cfg->capture_buffers ? cfg->capture_buffers : 2))) {
		free_bufs(&parked.bufs, &parked.count);
		if (alloc_capture(p, cfg->capture_buffers ? cfg->capture_buffers : 2))
			goto out;
		p->cap_allocated = 1;
	}
	if (prime_encoder(p, cfg->prime, start0))
		goto out;
	START_STAMP(6);
	if (p->early || p->cap != NULL) {
		/* Adopted or allocated above. */
	} else if (adopt_capture(p)) {
		goto out;
	}
	if (p->direct) {
		/*
		 * Each capture buffer goes to the encoder slot of its index, and
		 * must hold what the encoder reads: whole 16-line rows, which
		 * the capture allocates with ironkvm-dist patch 0921.
		 */
		off_t size = lseek(p->cap[0].fd, 0, SEEK_END);

		if (p->cap_count > p->enc_out_count) {
			fail_msg(p, "encoder took %u capture buffers of %u",
				 p->enc_out_count, p->cap_count);
			goto out;
		}
		if (size < (off_t)p->enc_out_fmt.sizeimage) {
			fail_msg(p, "capture buffer of %lld bytes is short of the encoder's %u",
				 (long long)size, p->enc_out_fmt.sizeimage);
			goto out;
		}
	}
	START_STAMP(7);
	if (!p->cap_on) {
		if (stream(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, 1)) {
			/* EPIPE here is the capture link refusing the source
			 * format; EINVAL, ENOLINK or ERANGE the receiver's
			 * answer while the source changes mode. */
			fail(p, "capture STREAMON");
			p->cap_on_failed = 1;
			goto out;
		}
		p->cap_on = 1;
	}
	START_STAMP(8);
	if (!p->direct) {
		if (request(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF,
			    p->cap_count) < (int)p->cap_count) {
			fail(p, "scaler OUTPUT REQBUFS");
			goto out;
		}
		if (stream(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, 1)) {
			fail(p, "scaler OUTPUT STREAMON");
			goto out;
		}
		p->vpss_out_on = 1;
		if (stream(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, 1)) {
			fail(p, "scaler CAPTURE STREAMON");
			goto out;
		}
		p->vpss_cap_on = 1;
	}
	START_STAMP(9);
	/* Frames that completed while the encoder was set up are as old as
	 * that took: the first read waits for the next one instead. */
	if (p->early)
		p->fresh_after_us = now_us();
	p->running = 1;
	return 0;

out: {
	char error[sizeof(p->error)];
	int saved = errno;
	int cap_on_failed = p->cap_on_failed;

	memcpy(error, p->error, sizeof(error));
	kvmv_pipe_stop(p);
	memcpy(p->error, error, sizeof(error));
	p->cap_on_failed = cap_on_failed;
	errno = saved;
	return -1;
}
}

static void free_bufs(struct kvmv_buf **bufs, unsigned int *count)
{
	unsigned int i;

	if (*bufs == NULL)
		return;
	for (i = 0; i < *count; i++) {
		if ((*bufs)[i].addr != NULL)
			munmap((*bufs)[i].addr, (*bufs)[i].length);
		if ((*bufs)[i].fd >= 0)
			close((*bufs)[i].fd);
	}
	free(*bufs);
	*bufs = NULL;
	*count = 0;
}

static void snap_stop(struct kvmv_pipe *p);

void kvmv_pipe_stop(struct kvmv_pipe *p)
{
	/* The snapshot context imports the capture buffers: it goes first. */
	snap_stop(p);
	if (p->vpss_out_on)
		stream(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, 0);
	if (p->vpss_cap_on)
		stream(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, 0);
	if (p->cap_on)
		stream(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, 0);
	if (p->enc_out_on)
		stream(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, 0);
	if (p->enc_cap_on)
		stream(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, 0);
	free_bufs(&p->bs, &p->bs_count);
	free_bufs(&p->mid, &p->mid_count);
	free_bufs(&p->cap, &p->cap_count);
	/* Closing the nodes frees every queue's buffers in the drivers. */
	if (p->heap_fd >= 0)
		close(p->heap_fd);
	if (p->vpss_fd >= 0)
		close(p->vpss_fd);
	if (p->enc_fd >= 0)
		close(p->enc_fd);
	if (p->cap_fd >= 0)
		close(p->cap_fd);
	kvmv_pipe_init(p);
}

void kvmv_pipe_park(struct kvmv_pipe *p)
{
	kvmv_pipe_unpark();
	if (p->cap_fd < 0 || p->cap == NULL) {
		kvmv_pipe_stop(p);
		return;
	}
	/* The snapshot context imports the capture buffers: it goes first,
	 * then the scaler that reads them, then the capture itself. */
	snap_stop(p);
	if (p->vpss_out_on)
		stream(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, 0);
	p->vpss_out_on = 0;
	if (p->cap_on)
		stream(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, 0);
	p->cap_on = 0;
	parked.fd = p->cap_fd;
	parked.bufs = p->cap;
	parked.count = p->cap_count;
	parked.sizeimage = p->cap_fmt.sizeimage;
	if (p->park_encoder && p->running && p->enc_fd >= 0 && p->bs != NULL &&
	    (p->mid != NULL || p->direct) && p->heap_fd >= 0) {
		/*
		 * Streamed off, the encoder ends its sequence: it frees its
		 * reference frames, gives the codec SRAM back and, for the
		 * WAVE420L, powers its core down. What stays is the node and
		 * the buffers user space holds.
		 */
		if (p->enc_out_on)
			stream(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, 0);
		if (p->enc_cap_on)
			stream(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, 0);
		p->enc_out_on = 0;
		p->enc_cap_on = 0;
		parked_enc.fd = p->enc_fd;
		parked_enc.codec = p->codec;
		snprintf(parked_enc.path, sizeof(parked_enc.path), "%s", p->enc_path);
		parked_enc.cap_fmt = p->cap_fmt;
		parked_enc.plan = p->plan;
		parked_enc.enc_out_fmt = p->enc_out_fmt;
		parked_enc.enc_cap_fmt = p->enc_cap_fmt;
		parked_enc.bs = p->bs;
		parked_enc.bs_count = p->bs_count;
		parked_enc.mid = p->mid;
		parked_enc.mid_count = p->mid_count;
		parked_enc.heap_fd = p->heap_fd;
		parked_enc.want_direct = p->want_direct;
		parked_enc.direct = p->direct;
		parked_enc.enc_out_count = p->enc_out_count;
		p->enc_fd = -1;
		p->heap_fd = -1;
		p->bs = NULL;
		p->bs_count = 0;
		p->mid = NULL;
		p->mid_count = 0;
	}
	p->cap_fd = -1;
	p->cap = NULL;
	p->cap_count = 0;
	kvmv_pipe_stop(p);
}

/* ---- per frame ----------------------------------------------------- */

static int source_changed(int fd)
{
	struct v4l2_event event;
	int changed = 0;

	while (xioctl(fd, VIDIOC_DQEVENT, &event) == 0)
		if (event.type == V4L2_EVENT_SOURCE_CHANGE)
			changed = 1;
	return changed;
}

static int stream_broken(int err)
{
	/* vb2 answers EIO once the driver marked the queue failed, which the
	 * capture driver does when the source changes while streaming. */
	return err == EIO || err == EPIPE || err == ENODEV || err == ENOLINK;
}

/*
 * How long to wait for the next capture frame instead of taking the one held,
 * in ms, or 0. The reader comes at its own rate, so the newest frame is
 * anything from 0 to a frame time old when it arrives, and what it takes
 * reaches the viewer that much later. When the next frame is due sooner than
 * the held one is old, and within pickup_wait_max_us, the newer one is the
 * fresher picture by the time it is delivered.
 */
static unsigned int pickup_wait_ms(const struct kvmv_pipe *p)
{
	uint64_t now = now_us();
	uint64_t age, due;

	if (!p->pickup_wait_max_us || !p->cap_interval_us || now < p->cap_ts_us)
		return 0;
	age = now - p->cap_ts_us;
	if (age >= p->cap_interval_us || age * 2 <= p->cap_interval_us)
		return 0;
	due = p->cap_interval_us - age;
	if (due > p->pickup_wait_max_us)
		return 0;
	/* A little over, for the interrupt and the wake-up. */
	return (unsigned int)((due + 2000U + 999U) / 1000U);
}

/* Take the newest finished capture buffer and give the older ones back. */
static enum kvmv_pipe_status newest_capture(struct kvmv_pipe *p,
					    unsigned int timeout_ms,
					    unsigned int *index)
{
	uint64_t deadline = now_ms() + timeout_ms;
	int held = -1;
	int waited = 0;

	for (;;) {
		struct v4l2_buffer buf;
		struct pollfd pfd;
		uint64_t now;

		if (dqbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
			  &buf) == 0) {
			if (buf.flags & V4L2_BUF_FLAG_ERROR) {
				if (qbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE,
					 V4L2_MEMORY_MMAP, buf.index, -1, 0))
					return fail(p, "capture requeue"),
					       KVMV_PIPE_ERROR;
				continue;
			}
			if (p->fresh_after_us &&
			    (uint64_t)buf.timestamp.tv_sec * 1000000U +
			    (uint64_t)buf.timestamp.tv_usec < p->fresh_after_us) {
				/* Completed before the pipeline was ready. */
				if (qbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE,
					 V4L2_MEMORY_MMAP, buf.index, -1, 0))
					return fail(p, "capture requeue"),
					       KVMV_PIPE_ERROR;
				continue;
			}
			if (held >= 0 &&
			    qbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE,
				 V4L2_MEMORY_MMAP, (unsigned int)held, -1, 0))
				return fail(p, "capture requeue"), KVMV_PIPE_ERROR;
			held = (int)buf.index;
			{
				uint64_t ts = (uint64_t)buf.timestamp.tv_sec * 1000000U +
					      (uint64_t)buf.timestamp.tv_usec;
				uint32_t frames = buf.sequence - p->cap_seq;

				/* The frame time, from frames that both reached a
				 * buffer: 5 to 50 ms, or it is not a frame time. */
				if (p->cap_ts_us && frames && ts > p->cap_ts_us) {
					uint64_t iv = (ts - p->cap_ts_us) / frames;

					if (iv >= 5000 && iv <= 50000)
						p->cap_interval_us = (uint32_t)iv;
				}
				p->cap_ts_us = ts;
				p->times.seq_gap += frames;
				p->cap_seq = buf.sequence;
			}
			continue;
		}
		if (errno != EAGAIN) {
			fail(p, "capture DQBUF");
			return stream_broken(errno) ? KVMV_PIPE_SOURCE_CHANGED :
						      KVMV_PIPE_ERROR;
		}
		if (held >= 0) {
			unsigned int wait = waited ? 0 : pickup_wait_ms(p);

			if (wait == 0) {
				*index = (unsigned int)held;
				p->fresh_after_us = 0;
				return KVMV_PIPE_OK;
			}
			/* Once: poll for the next frame, then take the newest. */
			waited = 1;
			p->times.pickup_waits++;
			deadline = now_ms() + wait;
		}
		now = now_ms();
		if (now >= deadline) {
			if (held >= 0) {
				*index = (unsigned int)held;
				p->fresh_after_us = 0;
				return KVMV_PIPE_OK;
			}
			snprintf(p->error, sizeof(p->error),
				 "no frame from the capture node in %u ms",
				 timeout_ms);
			return KVMV_PIPE_NO_FRAME;
		}
		pfd.fd = p->cap_fd;
		pfd.events = POLLIN | POLLPRI;
		pfd.revents = 0;
		if (poll(&pfd, 1, (int)(deadline - now)) < 0 && errno != EINTR)
			return fail(p, "capture poll"), KVMV_PIPE_ERROR;
		if ((pfd.revents & POLLPRI) && source_changed(p->cap_fd)) {
			snprintf(p->error, sizeof(p->error), "source change event");
			return KVMV_PIPE_SOURCE_CHANGED;
		}
		if ((pfd.revents & POLLERR) && !(pfd.revents & POLLIN)) {
			snprintf(p->error, sizeof(p->error),
				 "capture queue reports an error");
			return KVMV_PIPE_SOURCE_CHANGED;
		}
	}
}

static uint64_t now_us(void)
{
	struct timespec ts;

	clock_gettime(CLOCK_MONOTONIC, &ts);
	return (uint64_t)ts.tv_sec * 1000000U + (uint64_t)ts.tv_nsec / 1000U;
}

enum kvmv_pipe_status kvmv_pipe_scale(struct kvmv_pipe *p,
				      unsigned int timeout_ms, unsigned int *mid)
{
	enum kvmv_pipe_status status;
	struct v4l2_buffer buf;
	unsigned int ci, mi;
	uint64_t t0, t1;

	if (!p->running)
		return fail_msg(p, "pipeline not running"), KVMV_PIPE_ERROR;
	if (source_changed(p->cap_fd)) {
		snprintf(p->error, sizeof(p->error), "source change event");
		return KVMV_PIPE_SOURCE_CHANGED;
	}

	t0 = now_us();
	status = newest_capture(p, timeout_ms, &ci);
	if (status != KVMV_PIPE_OK)
		return status;
	t1 = now_us();
	p->times.capture_us += t1 - t0;
	if (p->cap_ts_us && t1 > p->cap_ts_us)
		p->times.pick_age_us += t1 - p->cap_ts_us;

	/* Scale: the capture buffer in, a middle buffer out. */
	mi = p->next_mid;
	p->next_mid = (p->next_mid + 1) % p->mid_count;
	if (qbuf(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF, ci,
		 p->cap[ci].fd, p->cap_fmt.sizeimage))
		return fail(p, "scaler OUTPUT QBUF"), KVMV_PIPE_ERROR;
	if (qbuf(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_DMABUF, mi,
		 p->mid[mi].fd, 0))
		return fail(p, "scaler CAPTURE QBUF"), KVMV_PIPE_ERROR;
	if (wait_dqbuf(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_DMABUF,
		       POLLIN, KVMV_M2M_TIMEOUT_MS, &buf))
		return fail(p, "scaler CAPTURE DQBUF"), KVMV_PIPE_ERROR;
	if (buf.index != mi || (buf.flags & V4L2_BUF_FLAG_ERROR))
		return fail_msg(p, "scaler returned buffer %u flags %#x for %u",
				buf.index, buf.flags, mi), KVMV_PIPE_ERROR;
	if (wait_dqbuf(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF,
		       POLLOUT, KVMV_M2M_TIMEOUT_MS, &buf))
		return fail(p, "scaler OUTPUT DQBUF"), KVMV_PIPE_ERROR;
	if (qbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
		 buf.index, -1, 0))
		return fail(p, "capture requeue"), KVMV_PIPE_ERROR;
	p->times.scale_us += now_us() - t1;
	*mid = mi;
	return KVMV_PIPE_OK;
}

/*
 * KVMV_PRIME_ASYNC: take the priming picture's output and input back, before
 * the first live picture goes in, and ask for the keyframe that was held back
 * for it. Mostly done by then: the encode ran while the capture frame was
 * waited for and scaled.
 */
static int collect_prime(struct kvmv_pipe *p)
{
	struct v4l2_buffer buf;

	if (!p->prime_pending)
		return 0;
	if (wait_dqbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
		       POLLIN, KVMV_PRIME_TIMEOUT_MS, &buf))
		return fail(p, "encoder priming output");
	if (qbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
		 buf.index, -1, 0))
		return fail(p, "encoder CAPTURE requeue");
	if (wait_dqbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF,
		       POLLOUT, KVMV_M2M_TIMEOUT_MS, &buf))
		return fail(p, "encoder priming input");
	p->prime_pending = 0;
	if (p->key_after_prime) {
		p->key_after_prime = 0;
		if (set_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_FORCE_KEY_FRAME, 1))
			return fail(p, "force keyframe");
	}
	return 0;
}

/*
 * Scale-ahead: give the scaler the capture frame that just completed, into the
 * middle buffer the encoder is not reading. Called while the encoder works.
 * Nothing to take is not an error: the next read scales its own frame.
 */
static enum kvmv_pipe_status start_ahead(struct kvmv_pipe *p)
{
	enum kvmv_pipe_status status;
	uint32_t pickup = p->pickup_wait_max_us;
	unsigned int ci, mi;

	/* Take what has completed; no waiting for a frame about to. */
	p->pickup_wait_max_us = 0;
	status = newest_capture(p, 0, &ci);
	p->pickup_wait_max_us = pickup;
	if (status == KVMV_PIPE_NO_FRAME) {
		p->error[0] = '\0';
		return KVMV_PIPE_OK;
	}
	if (status != KVMV_PIPE_OK)
		return status;
	mi = p->next_mid;
	p->next_mid = (p->next_mid + 1) % p->mid_count;
	if (qbuf(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF, ci,
		 p->cap[ci].fd, p->cap_fmt.sizeimage))
		return fail(p, "scaler OUTPUT QBUF (ahead)"), KVMV_PIPE_ERROR;
	if (qbuf(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_DMABUF, mi,
		 p->mid[mi].fd, 0))
		return fail(p, "scaler CAPTURE QBUF (ahead)"), KVMV_PIPE_ERROR;
	p->ahead_mid = (int)mi;
	p->ahead_ci = ci;
	p->ahead_ts_us = p->cap_ts_us;
	p->ahead_start_us = now_us();
	return KVMV_PIPE_OK;
}

/*
 * Wait for the encoder's output as wait_dqbuf would, but watch the capture
 * node too, and start the scaler on a frame that completes meanwhile. Returns
 * once the encoder has output or the wait is over; wait_dqbuf then takes it
 * or reports the timeout.
 */
static enum kvmv_pipe_status encode_wait_ahead(struct kvmv_pipe *p,
						uint64_t enc_start_us)
{
	uint64_t deadline = now_ms() + KVMV_M2M_TIMEOUT_MS;
	uint64_t not_before = enc_start_us + p->ahead_delay_us;

	while (p->ahead_mid < 0) {
		struct pollfd pfd[2];
		uint64_t now = now_ms();
		uint64_t us = now_us();
		int early = us < not_before;
		int timeout = (int)(deadline - now);

		if (now >= deadline)
			return KVMV_PIPE_OK;
		/* Before the delay is over only the encoder is watched, then
		 * whatever completed meanwhile goes to the scaler. */
		if (early && (int)((not_before - us + 999U) / 1000U) < timeout)
			timeout = (int)((not_before - us + 999U) / 1000U);
		pfd[0].fd = p->enc_fd;
		pfd[0].events = POLLIN;
		pfd[0].revents = 0;
		pfd[1].fd = p->cap_fd;
		pfd[1].events = POLLIN;
		pfd[1].revents = 0;
		if (poll(pfd, early ? 1 : 2, timeout) < 0) {
			if (errno == EINTR)
				continue;
			return fail(p, "encoder poll"), KVMV_PIPE_ERROR;
		}
		if (pfd[0].revents)
			return early ? start_ahead(p) : KVMV_PIPE_OK;
		if (early)
			continue;
		/* A source change or a queue error is the next read's to see. */
		if (pfd[1].revents & ~POLLIN)
			return KVMV_PIPE_OK;
		if (pfd[1].revents & POLLIN)
			return start_ahead(p);
	}
	return KVMV_PIPE_OK;
}

/* Collect the picture start_ahead gave the scaler. */
static enum kvmv_pipe_status finish_ahead(struct kvmv_pipe *p, unsigned int *mid)
{
	struct v4l2_buffer buf;
	unsigned int mi = (unsigned int)p->ahead_mid;
	uint64_t t0 = now_us();

	p->ahead_mid = -1;
	if (wait_dqbuf(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_DMABUF,
		       POLLIN, KVMV_M2M_TIMEOUT_MS, &buf))
		return fail(p, "scaler CAPTURE DQBUF (ahead)"), KVMV_PIPE_ERROR;
	if (buf.index != mi || (buf.flags & V4L2_BUF_FLAG_ERROR))
		return fail_msg(p, "scaler returned buffer %u flags %#x for %u (ahead)",
				buf.index, buf.flags, mi), KVMV_PIPE_ERROR;
	if (wait_dqbuf(p->vpss_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF,
		       POLLOUT, KVMV_M2M_TIMEOUT_MS, &buf))
		return fail(p, "scaler OUTPUT DQBUF (ahead)"), KVMV_PIPE_ERROR;
	if (qbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
		 buf.index, -1, 0))
		return fail(p, "capture requeue"), KVMV_PIPE_ERROR;
	/* What is left of the scale once the read wants it. */
	p->times.scale_us += now_us() - t0;
	p->times.ahead++;
	p->cap_ts_us = p->ahead_ts_us;
	*mid = mi;
	return KVMV_PIPE_OK;
}

/*
 * The encoder has its picture (queued at t0, captured at frame_ts): take the
 * access unit and the picture back. *input_back says whether the encoder gave
 * the picture back, so its buffer may be reused.
 */
static enum kvmv_pipe_status finish_encode(struct kvmv_pipe *p, uint64_t t0,
					   uint64_t frame_ts,
					   struct kvmv_encoded *out,
					   int *input_back)
{
	struct v4l2_buffer buf;

	*input_back = 0;
	if (wait_dqbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
		       POLLIN, KVMV_M2M_TIMEOUT_MS, &buf))
		return fail(p, "encoder CAPTURE DQBUF"), KVMV_PIPE_ERROR;
	out->index = buf.index;
	out->flags = buf.flags;
	out->size = buf.bytesused;
	if (wait_dqbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF,
		       POLLOUT, KVMV_M2M_TIMEOUT_MS, &buf)) {
		fail(p, "encoder OUTPUT DQBUF");
		kvmv_pipe_release(p, out);
		return KVMV_PIPE_ERROR;
	}
	*input_back = 1;
	if (out->index >= p->bs_count || (out->flags & V4L2_BUF_FLAG_ERROR) ||
	    out->size == 0 || out->size > p->bs[out->index].length) {
		fail_msg(p, "encoder returned buffer %u flags %#x size %zu",
			 out->index, out->flags, out->size);
		if (out->index < p->bs_count)
			kvmv_pipe_release(p, out);
		return KVMV_PIPE_ERROR;
	}
	out->data = p->bs[out->index].addr;
	{
		uint64_t t1 = now_us();
		uint64_t age = frame_ts && t1 > frame_ts ? t1 - frame_ts : 0;

		p->times.encode_us += t1 - t0;
		if (p->ahead_mid >= 0 && p->ahead_start_us >= t0 &&
		    t1 > p->ahead_start_us)
			p->times.overlap_us += t1 - p->ahead_start_us;
		p->times.age_us += age;
		if (age > p->times.age_max_us)
			p->times.age_max_us = age;
	}
	p->times.frames++;
	return KVMV_PIPE_OK;
}

enum kvmv_pipe_status kvmv_pipe_encode_mid(struct kvmv_pipe *p,
					   unsigned int mi,
					   struct kvmv_encoded *out)
{
	uint64_t t0 = now_us();
	/* This picture's capture time: start_ahead moves cap_ts_us on. */
	uint64_t frame_ts = p->cap_ts_us;
	int input_back;

	memset(out, 0, sizeof(*out));
	if (!p->running || mi >= p->mid_count)
		return fail_msg(p, "no scaled picture to encode"), KVMV_PIPE_ERROR;
	if (collect_prime(p))
		return KVMV_PIPE_ERROR;
	if (qbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF, mi,
		 p->mid[mi].fd, p->enc_out_fmt.sizeimage))
		return fail(p, "encoder OUTPUT QBUF"), KVMV_PIPE_ERROR;
	if (p->scale_ahead && p->ahead_mid < 0 && p->mid_count >= 2) {
		enum kvmv_pipe_status status = encode_wait_ahead(p, t0);

		if (status != KVMV_PIPE_OK)
			return status;
	}
	return finish_encode(p, t0, frame_ts, out, &input_back);
}

/*
 * The direct path: the newest capture buffer goes to the encoder as it is
 * (its OUTPUT slot of the same index) and back to the capture once the
 * encoder has read it. No scaler, no middle buffer.
 */
static enum kvmv_pipe_status encode_direct(struct kvmv_pipe *p,
					   unsigned int timeout_ms,
					   struct kvmv_encoded *out)
{
	enum kvmv_pipe_status status;
	unsigned int ci;
	uint64_t t0, t1;
	int input_back;

	if (!p->running)
		return fail_msg(p, "pipeline not running"), KVMV_PIPE_ERROR;
	if (source_changed(p->cap_fd)) {
		snprintf(p->error, sizeof(p->error), "source change event");
		return KVMV_PIPE_SOURCE_CHANGED;
	}
	t0 = now_us();
	status = newest_capture(p, timeout_ms, &ci);
	if (status != KVMV_PIPE_OK)
		return status;
	t1 = now_us();
	p->times.capture_us += t1 - t0;
	if (p->cap_ts_us && t1 > p->cap_ts_us)
		p->times.pick_age_us += t1 - p->cap_ts_us;
	if (ci >= p->enc_out_count ||
	    qbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF, ci,
		 p->cap[ci].fd, p->enc_out_fmt.sizeimage)) {
		fail(p, "encoder OUTPUT QBUF (capture buffer)");
		qbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP, ci, -1, 0);
		return KVMV_PIPE_ERROR;
	}
	status = finish_encode(p, t1, p->cap_ts_us, out, &input_back);
	/* Still with the encoder after a failure: the pipe is torn down. */
	if (input_back &&
	    qbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP, ci, -1, 0) &&
	    status == KVMV_PIPE_OK) {
		fail(p, "capture requeue");
		kvmv_pipe_release(p, out);
		return KVMV_PIPE_ERROR;
	}
	return status;
}

enum kvmv_pipe_status kvmv_pipe_encode(struct kvmv_pipe *p,
				       unsigned int timeout_ms,
				       struct kvmv_encoded *out)
{
	enum kvmv_pipe_status status;
	unsigned int mi;

	memset(out, 0, sizeof(*out));
	if (p->direct)
		return encode_direct(p, timeout_ms, out);
	if (p->ahead_mid >= 0) {
		uint64_t now = now_us();
		/* More than two frame times old (the reads paused, or came
		 * slower than the frames): collect it and take a newer one.
		 * Never less than two 60 Hz frame times: from a faster source
		 * (trial 31, 100 to 110 Hz) a picture scaled ahead is younger
		 * than that whenever the reads keep up with the encoder, and
		 * two source frame times (18 ms at 110 Hz) is shorter than
		 * one encode plus the wait for the next read, which threw
		 * away half of them. */
		uint64_t iv = p->cap_interval_us ? p->cap_interval_us : 16667U;
		uint64_t stale = 2ULL * (iv < 16667U ? 16667U : iv);

		status = finish_ahead(p, &mi);
		if (status != KVMV_PIPE_OK)
			return status;
		if (now > p->ahead_ts_us && now - p->ahead_ts_us > stale) {
			p->times.ahead--;
			status = kvmv_pipe_scale(p, timeout_ms, &mi);
		}
	} else {
		status = kvmv_pipe_scale(p, timeout_ms, &mi);
	}
	if (status != KVMV_PIPE_OK)
		return status;
	return kvmv_pipe_encode_mid(p, mi, out);
}

/* ---- snapshots ----------------------------------------------------- */

static void snap_stop(struct kvmv_pipe *p)
{
	struct kvmv_snap *s = &p->snap;

	if (s->synced)
		sync_dmabuf(s->buf.fd, DMA_BUF_SYNC_END | DMA_BUF_SYNC_READ);
	if (s->out_on)
		stream(s->fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, 0);
	if (s->cap_on)
		stream(s->fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, 0);
	if (s->buf.addr != NULL)
		munmap(s->buf.addr, s->buf.length);
	if (s->buf.fd >= 0)
		close(s->buf.fd);
	if (s->fd >= 0)
		close(s->fd);
	memset(s, 0, sizeof(*s));
	s->fd = -1;
	s->buf.fd = -1;
}

static int snap_start(struct kvmv_pipe *p, unsigned int width,
		      unsigned int height, int keep_shape)
{
	struct kvmv_snap *s = &p->snap;
	const struct v4l2_pix_format *cap = &p->cap_fmt;
	struct v4l2_pix_format in, full;
	struct dma_heap_allocation_data alloc;
	size_t need, jpeg_need;

	if (kvmv_plan_output(cap->width, cap->height, width, height, keep_shape,
			     &s->plan))
		return fail_msg(p, "snapshot size %ux%u is unusable", width, height);
	s->fd = open(p->scaler_path, O_RDWR | O_NONBLOCK | O_CLOEXEC);
	if (s->fd < 0)
		return fail(p, "open scaler for snapshots");

	if (set_fmt(s->fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, cap->pixelformat,
		    cap->width, cap->height, cap, 0, &in))
		return fail(p, "snapshot scaler OUTPUT S_FMT");
	if (in.bytesperline != cap->bytesperline || in.width != cap->width ||
	    in.height != cap->height)
		return fail_msg(p, "snapshot scaler wants %ux%u stride %u, capture gives %ux%u stride %u",
				in.width, in.height, in.bytesperline, cap->width,
				cap->height, cap->bytesperline);
	/*
	 * Snapshots become JPEGs, and JPEG (JFIF) is full range. The capture is
	 * limited range; a scaler with ironkvm-dist patch 0908 expands it when
	 * asked, one without answers with the source's range, and the picture
	 * is then encoded as it is. The H.264 context never asks: H.264 carries
	 * limited range as standard.
	 */
	full = in;
	full.quantization = V4L2_QUANTIZATION_FULL_RANGE;
	if (set_fmt(s->fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_PIX_FMT_NV12,
		    s->plan.out_width, s->plan.out_height, &full, 0, &s->fmt))
		return fail(p, "snapshot scaler CAPTURE S_FMT");
	need = (size_t)s->fmt.bytesperline * s->fmt.height * 3 / 2;
	if (s->fmt.pixelformat != V4L2_PIX_FMT_NV12 ||
	    s->fmt.width < s->plan.out_width || s->fmt.height < s->plan.out_height ||
	    s->fmt.bytesperline < s->plan.out_width || s->fmt.sizeimage < need)
		return fail_msg(p, "snapshot scaler took %.4s %ux%u stride %u size %u for NV12 %ux%u",
				(const char *)&s->fmt.pixelformat, s->fmt.width,
				s->fmt.height, s->fmt.bytesperline,
				s->fmt.sizeimage, s->plan.out_width,
				s->plan.out_height);
	if (set_crop(s->fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, s->plan.out_width,
		     s->plan.out_height)) {
		if (errno != EINVAL || s->fmt.width != s->plan.out_width ||
		    s->fmt.height != s->plan.out_height)
			return fail(p, "snapshot scaler CAPTURE crop");
	}

	if (request(s->fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_DMABUF, 1) < 1)
		return fail(p, "snapshot scaler CAPTURE REQBUFS");
	if (request(s->fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF,
		    p->cap_count) < (int)p->cap_count)
		return fail(p, "snapshot scaler OUTPUT REQBUFS");
	/* Room for the JPEG unit's reads past the last chroma row. */
	need = s->fmt.sizeimage;
	jpeg_need = kvmv_hwjpeg_src_size(s->fmt.bytesperline, s->fmt.height);
	if (jpeg_need > need)
		need = jpeg_need;
	need = (need + 4095U) & ~(size_t)4095U;
	memset(&alloc, 0, sizeof(alloc));
	alloc.len = need;
	alloc.fd_flags = O_CLOEXEC | O_RDWR;
	if (xioctl(p->heap_fd, DMA_HEAP_IOCTL_ALLOC, &alloc))
		return fail(p, "snapshot DMA_HEAP_IOCTL_ALLOC");
	s->buf.fd = (int)alloc.fd;
	s->buf.length = need;
	/* A dma-heap buffer maps cached; the sync calls keep it coherent. */
	s->buf.addr = mmap(NULL, s->buf.length, PROT_READ, MAP_SHARED,
			   s->buf.fd, 0);
	if (s->buf.addr == MAP_FAILED) {
		s->buf.addr = NULL;
		return fail(p, "snapshot mmap");
	}
	if (stream(s->fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, 1))
		return fail(p, "snapshot scaler OUTPUT STREAMON");
	s->out_on = 1;
	if (stream(s->fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, 1))
		return fail(p, "snapshot scaler CAPTURE STREAMON");
	s->cap_on = 1;
	return 0;
}

enum kvmv_pipe_status kvmv_pipe_snapshot(struct kvmv_pipe *p,
					 unsigned int width, unsigned int height,
					 int keep_shape, unsigned int timeout_ms,
					 int cpu_read, struct kvmv_nv12 *image)
{
	struct kvmv_snap *s = &p->snap;
	enum kvmv_pipe_status status;
	struct v4l2_buffer buf;
	struct kvmv_plan want;
	unsigned int ci;

	memset(image, 0, sizeof(*image));
	if (!p->running)
		return fail_msg(p, "pipeline not running"), KVMV_PIPE_ERROR;
	kvmv_pipe_snapshot_done(p);
	if (s->fd >= 0 &&
	    (kvmv_plan_output(p->cap_fmt.width, p->cap_fmt.height, width, height,
			      keep_shape, &want) ||
	     want.out_width != s->plan.out_width ||
	     want.out_height != s->plan.out_height))
		snap_stop(p);
	if (s->fd < 0 && snap_start(p, width, height, keep_shape)) {
		int saved = errno;
		char error[sizeof(p->error)];

		memcpy(error, p->error, sizeof(error));
		snap_stop(p);
		memcpy(p->error, error, sizeof(error));
		errno = saved;
		return KVMV_PIPE_ERROR;
	}

	if (source_changed(p->cap_fd)) {
		snprintf(p->error, sizeof(p->error), "source change event");
		return KVMV_PIPE_SOURCE_CHANGED;
	}
	status = newest_capture(p, timeout_ms, &ci);
	if (status != KVMV_PIPE_OK)
		return status;

	if (qbuf(s->fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF, ci,
		 p->cap[ci].fd, p->cap_fmt.sizeimage)) {
		fail(p, "snapshot scaler OUTPUT QBUF");
		qbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP, ci,
		     -1, 0);
		return KVMV_PIPE_ERROR;
	}
	if (qbuf(s->fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_DMABUF, 0,
		 s->buf.fd, 0))
		return fail(p, "snapshot scaler CAPTURE QBUF"), KVMV_PIPE_ERROR;
	if (wait_dqbuf(s->fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_DMABUF,
		       POLLIN, KVMV_M2M_TIMEOUT_MS, &buf))
		return fail(p, "snapshot scaler CAPTURE DQBUF"), KVMV_PIPE_ERROR;
	if (buf.flags & V4L2_BUF_FLAG_ERROR)
		return fail_msg(p, "snapshot scaler returned an error buffer"),
		       KVMV_PIPE_ERROR;
	if (wait_dqbuf(s->fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF,
		       POLLOUT, KVMV_M2M_TIMEOUT_MS, &buf))
		return fail(p, "snapshot scaler OUTPUT DQBUF"), KVMV_PIPE_ERROR;
	if (qbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
		 buf.index, -1, 0))
		return fail(p, "capture requeue"), KVMV_PIPE_ERROR;

	if (cpu_read) {
		sync_dmabuf(s->buf.fd, DMA_BUF_SYNC_START | DMA_BUF_SYNC_READ);
		s->synced = 1;
	}
	image->y = s->buf.addr;
	image->uv = (const uint8_t *)s->buf.addr +
		    (size_t)s->fmt.bytesperline * s->fmt.height;
	image->width = s->plan.out_width;
	image->height = s->plan.out_height;
	image->stride = s->fmt.bytesperline;
	image->fd = s->buf.fd;
	image->fd_size = s->buf.length;
	image->full_range = s->fmt.quantization == V4L2_QUANTIZATION_FULL_RANGE;
	return KVMV_PIPE_OK;
}

enum kvmv_pipe_status kvmv_pipe_lend(struct kvmv_pipe *p,
				     unsigned int timeout_ms,
				     struct kvmv_frame *frame)
{
	enum kvmv_pipe_status status;
	unsigned int ci;

	memset(frame, 0, sizeof(*frame));
	frame->fd = -1;
	if (!p->running)
		return fail_msg(p, "pipeline not running"), KVMV_PIPE_ERROR;
	if (source_changed(p->cap_fd)) {
		snprintf(p->error, sizeof(p->error), "source change event");
		return KVMV_PIPE_SOURCE_CHANGED;
	}
	status = newest_capture(p, timeout_ms, &ci);
	if (status != KVMV_PIPE_OK)
		return status;
	frame->index = ci;
	frame->fd = p->cap[ci].fd;
	frame->size = p->cap[ci].length;
	{
		/*
		 * The buffer can be larger than the frame: with ironkvm-dist
		 * patch 0921 it has room for the 16-line MCU rows the JPEG unit
		 * reads. A dma-buf tells its size through lseek.
		 */
		off_t end = lseek(frame->fd, 0, SEEK_END);

		if (end > 0 && (size_t)end > frame->size)
			frame->size = (size_t)end;
	}
	frame->fmt = &p->cap_fmt;
	return KVMV_PIPE_OK;
}

int kvmv_pipe_give_back(struct kvmv_pipe *p, const struct kvmv_frame *frame)
{
	if (qbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
		 frame->index, -1, 0))
		return fail(p, "capture requeue");
	return 0;
}

void kvmv_pipe_snapshot_done(struct kvmv_pipe *p)
{
	struct kvmv_snap *s = &p->snap;

	if (!s->synced)
		return;
	sync_dmabuf(s->buf.fd, DMA_BUF_SYNC_END | DMA_BUF_SYNC_READ);
	s->synced = 0;
}

void kvmv_copy_from_device(uint8_t *dst, const uint8_t *src, size_t len)
{
	/*
	 * Each load from an uncached mapping is a DRAM round trip, so the
	 * number of loads is the cost: 8 bytes each where both ends allow it.
	 * The mapping starts on a page, so src is aligned in practice; dst is
	 * the caller's.
	 */
	if ((((uintptr_t)dst | (uintptr_t)src) & 7U) == 0) {
		uint64_t *d = (uint64_t *)(void *)dst;
		const volatile uint64_t *s = (const volatile uint64_t *)(const void *)src;
		size_t words = len / 8;
		size_t i;

		for (i = 0; i + 4 <= words; i += 4) {
			uint64_t a = s[i], b = s[i + 1], c = s[i + 2], e = s[i + 3];

			d[i] = a;
			d[i + 1] = b;
			d[i + 2] = c;
			d[i + 3] = e;
		}
		for (; i < words; i++)
			d[i] = s[i];
		dst += words * 8;
		src += words * 8;
		len -= words * 8;
	}
	memcpy(dst, src, len);
}

void kvmv_pipe_release(struct kvmv_pipe *p, const struct kvmv_encoded *e)
{
	if (p->running && e->index < p->bs_count)
		qbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP,
		     e->index, -1, 0);
}

int kvmv_pipe_set_bitrate(struct kvmv_pipe *p, uint32_t bitrate_bps)
{
	if (set_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_BITRATE, (int32_t)bitrate_bps))
		return fail(p, "set bitrate");
	read_back(p);
	return 0;
}

int kvmv_pipe_set_gop(struct kvmv_pipe *p, unsigned int gop)
{
	if (set_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_GOP_SIZE,
		     ctrl_fit(p->enc_fd, V4L2_CID_MPEG_VIDEO_GOP_SIZE,
			      (int32_t)gop)))
		return fail(p, "set GOP");
	read_back(p);
	return 0;
}

int kvmv_pipe_set_fps(struct kvmv_pipe *p, unsigned int fps)
{
	if (set_frame_rate(p, fps))
		return fail(p, "set frame rate");
	read_back(p);
	return 0;
}

int kvmv_pipe_force_key(struct kvmv_pipe *p)
{
	if (p->prime_pending) {
		/*
		 * Asked now, the key could go to the priming picture if the
		 * encoder has not taken it yet: hold it for the first live one
		 * (collect_prime). Answer as the control would.
		 */
		struct v4l2_queryctrl query;

		memset(&query, 0, sizeof(query));
		query.id = V4L2_CID_MPEG_VIDEO_FORCE_KEY_FRAME;
		if (xioctl(p->enc_fd, VIDIOC_QUERYCTRL, &query))
			return fail(p, "force keyframe");
		p->key_after_prime = 1;
		return 0;
	}
	if (set_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_FORCE_KEY_FRAME, 1))
		return fail(p, "force keyframe");
	return 0;
}
