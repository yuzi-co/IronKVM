#define _GNU_SOURCE
#include "kvmv_pipeline.h"

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
				       queue_offers(fd, V4L2_BUF_TYPE_VIDEO_OUTPUT,
						    V4L2_PIX_FMT_NV12));
		close(fd);

		if (role == KVMV_ROLE_CAPTURE && !devices->capture[0])
			copy_path(devices->capture, path);
		else if (role == KVMV_ROLE_SCALER && !devices->scaler[0])
			copy_path(devices->scaler, path);
		else if (role == KVMV_ROLE_ENCODER && !devices->encoder[0])
			copy_path(devices->encoder, path);
	}

	if ((env = getenv("KVMV_CAPTURE_DEV")) != NULL && *env)
		copy_path(devices->capture, env);
	if ((env = getenv("KVMV_SCALER_DEV")) != NULL && *env)
		copy_path(devices->scaler, env);
	if ((env = getenv("KVMV_ENCODER_DEV")) != NULL && *env)
		copy_path(devices->encoder, env);
	kvmv_find_subdev(devices->subdev, sizeof(devices->subdev));

	if (devices->capture[0] && devices->scaler[0] && devices->encoder[0])
		return 0;
	snprintf(missing, missing_size, "%s%s%s",
		 devices->capture[0] ? "" : "capture (sg2002-capture) ",
		 devices->scaler[0] ? "" : "scaler (sg2002-vpss) ",
		 devices->encoder[0] ? "" : "encoder (coda, NV12 to H.264)");
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
		format.fmt.pix.colorspace = colour->colorspace;
		format.fmt.pix.xfer_func = colour->xfer_func;
		format.fmt.pix.ycbcr_enc = colour->ycbcr_enc;
		format.fmt.pix.quantization = colour->quantization;
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
	p->applied.bitrate_bps = -1;
	p->applied.gop = -1;
	p->applied.fps_numerator = -1;
	p->applied.fps_denominator = -1;
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
			     cfg->req_height, &p->plan))
		return fail_msg(p, "capture size %ux%u is unusable", cap->width,
				cap->height);

	/* The encoder first: its padded input surface dictates the layout the
	 * scaler has to write. */
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
	if (set_fmt(p->enc_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_PIX_FMT_H264,
		    p->plan.out_width, p->plan.out_height, cap,
		    KVMV_BITSTREAM_SIZE, &p->enc_cap_fmt))
		return fail(p, "encoder CAPTURE S_FMT");
	if (p->enc_cap_fmt.pixelformat != V4L2_PIX_FMT_H264)
		return fail_msg(p, "encoder does not produce H.264");

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
		set_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_GOP_SIZE, (int32_t)cfg->gop);
	read_back(p);

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

static int alloc_mid(struct kvmv_pipe *p, unsigned int wanted)
{
	unsigned int i;
	int got;

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
 * Encode one black picture before any live frame, as the bridge does. It
 * claims the encoder's working memory before the capture buffers take theirs,
 * and gives the first live picture a reference. Its output is thrown away;
 * the caller asks for a keyframe for the first live frame.
 */
static int prime_encoder(struct kvmv_pipe *p)
{
	size_t luma = (size_t)p->enc_out_fmt.bytesperline * p->enc_out_fmt.height;
	size_t frame = luma * 3 / 2;
	struct v4l2_buffer buf;
	uint8_t *map;

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

int kvmv_pipe_start(struct kvmv_pipe *p, const struct kvmv_devices *d,
		    const struct kvmv_pipe_cfg *cfg)
{
	kvmv_pipe_init(p);

	p->cap_fd = open(d->capture, O_RDWR | O_NONBLOCK | O_CLOEXEC);
	if (p->cap_fd < 0) {
		fail(p, d->capture);
		goto out;
	}
	p->vpss_fd = open(d->scaler, O_RDWR | O_NONBLOCK | O_CLOEXEC);
	if (p->vpss_fd < 0) {
		fail(p, d->scaler);
		goto out;
	}
	p->enc_fd = open(d->encoder, O_RDWR | O_NONBLOCK | O_CLOEXEC);
	if (p->enc_fd < 0) {
		fail(p, d->encoder);
		goto out;
	}
	subscribe_source_change(p->cap_fd);

	if (kvmv_pipe_negotiate(p, cfg))
		goto out;
	if (alloc_mid(p, cfg->mid_buffers ? cfg->mid_buffers : 2))
		goto out;
	if (alloc_bitstream(p, cfg->bitstream_buffers ? cfg->bitstream_buffers : 3))
		goto out;
	if (prime_encoder(p))
		goto out;
	if (alloc_capture(p, cfg->capture_buffers ? cfg->capture_buffers : 2))
		goto out;
	/* Capture first: while the CSI driver streams, the VIP fabric clocks
	 * are on and VPSS register access is safe. */
	if (stream(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, 1)) {
		/* EPIPE here is the capture link refusing the source format. */
		fail(p, "capture STREAMON");
		goto out;
	}
	p->cap_on = 1;
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
	p->running = 1;
	return 0;

out: {
	char error[sizeof(p->error)];
	int saved = errno;

	memcpy(error, p->error, sizeof(error));
	kvmv_pipe_stop(p);
	memcpy(p->error, error, sizeof(error));
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

void kvmv_pipe_stop(struct kvmv_pipe *p)
{
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

/* Take the newest finished capture buffer and give the older ones back. */
static enum kvmv_pipe_status newest_capture(struct kvmv_pipe *p,
					    unsigned int timeout_ms,
					    unsigned int *index)
{
	uint64_t deadline = now_ms() + timeout_ms;
	int held = -1;

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
			if (held >= 0 &&
			    qbuf(p->cap_fd, V4L2_BUF_TYPE_VIDEO_CAPTURE,
				 V4L2_MEMORY_MMAP, (unsigned int)held, -1, 0))
				return fail(p, "capture requeue"), KVMV_PIPE_ERROR;
			held = (int)buf.index;
			continue;
		}
		if (errno != EAGAIN) {
			fail(p, "capture DQBUF");
			return stream_broken(errno) ? KVMV_PIPE_SOURCE_CHANGED :
						      KVMV_PIPE_ERROR;
		}
		if (held >= 0) {
			*index = (unsigned int)held;
			return KVMV_PIPE_OK;
		}
		now = now_ms();
		if (now >= deadline) {
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

enum kvmv_pipe_status kvmv_pipe_encode(struct kvmv_pipe *p,
				       unsigned int timeout_ms,
				       struct kvmv_encoded *out)
{
	enum kvmv_pipe_status status;
	struct v4l2_buffer buf;
	unsigned int ci, mi;

	memset(out, 0, sizeof(*out));
	if (!p->running)
		return fail_msg(p, "pipeline not running"), KVMV_PIPE_ERROR;
	if (source_changed(p->cap_fd)) {
		snprintf(p->error, sizeof(p->error), "source change event");
		return KVMV_PIPE_SOURCE_CHANGED;
	}

	status = newest_capture(p, timeout_ms, &ci);
	if (status != KVMV_PIPE_OK)
		return status;

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

	/* Encode the middle buffer. */
	if (qbuf(p->enc_fd, V4L2_BUF_TYPE_VIDEO_OUTPUT, V4L2_MEMORY_DMABUF, mi,
		 p->mid[mi].fd, p->enc_out_fmt.sizeimage))
		return fail(p, "encoder OUTPUT QBUF"), KVMV_PIPE_ERROR;
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
	if (out->index >= p->bs_count || (out->flags & V4L2_BUF_FLAG_ERROR) ||
	    out->size == 0 || out->size > p->bs[out->index].length) {
		fail_msg(p, "encoder returned buffer %u flags %#x size %zu",
			 out->index, out->flags, out->size);
		if (out->index < p->bs_count)
			kvmv_pipe_release(p, out);
		return KVMV_PIPE_ERROR;
	}
	out->data = p->bs[out->index].addr;
	return KVMV_PIPE_OK;
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
	if (set_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_GOP_SIZE, (int32_t)gop))
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
	if (set_ctrl(p->enc_fd, V4L2_CID_MPEG_VIDEO_FORCE_KEY_FRAME, 1))
		return fail(p, "force keyframe");
	return 0;
}
