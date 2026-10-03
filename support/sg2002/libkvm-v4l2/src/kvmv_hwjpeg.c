#define _GNU_SOURCE
#include "kvmv_hwjpeg.h"

#include <errno.h>
#include <fcntl.h>
#include <poll.h>
#include <stdio.h>
#include <string.h>
#include <sys/ioctl.h>
#include <sys/mman.h>
#include <time.h>
#include <unistd.h>

#define KVMV_HWJPEG_TIMEOUT_MS 1000U

static int xioctl(int fd, unsigned long request, void *arg)
{
	return kvmv_ioctl_hook(fd, request, arg);
}

static int fail(struct kvmv_hwjpeg *hw, const char *step)
{
	int saved = errno;

	snprintf(hw->error, sizeof(hw->error), "JPEG unit: %s: %s", step,
		 saved ? strerror(saved) : "rejected");
	errno = saved;
	return -1;
}

void kvmv_hwjpeg_init(struct kvmv_hwjpeg *hw)
{
	memset(hw, 0, sizeof(*hw));
	hw->fd = -1;
	hw->out.fd = -1;
	hw->cap.fd = -1;
}

int kvmv_hwjpeg_fits(unsigned int width, unsigned int height, unsigned int stride)
{
	return width >= 16 && height >= 16 && width <= 8192 && height <= 8192 &&
	       (width % 16) == 0 && (height % 2) == 0 && (stride % 16) == 0 &&
	       stride >= width;
}

size_t kvmv_hwjpeg_src_size(unsigned int stride, unsigned int height)
{
	return (size_t)stride * height + (size_t)stride * ((height + 15U) & ~15U) / 2;
}

int kvmv_hwjpeg_open(struct kvmv_hwjpeg *hw, const char *path)
{
	kvmv_hwjpeg_init(hw);
	hw->fd = open(path, O_RDWR | O_NONBLOCK | O_CLOEXEC);
	if (hw->fd < 0)
		return fail(hw, path);
	return 0;
}

static void unmap(struct kvmv_buf *buf)
{
	if (buf->addr != NULL)
		munmap(buf->addr, buf->length);
	buf->addr = NULL;
	buf->length = 0;
}

static void request(int fd, enum v4l2_buf_type type, enum v4l2_memory memory,
		    unsigned int count, int *got)
{
	struct v4l2_requestbuffers req;

	memset(&req, 0, sizeof(req));
	req.count = count;
	req.type = type;
	req.memory = memory;
	if (xioctl(fd, VIDIOC_REQBUFS, &req))
		*got = -1;
	else
		*got = (int)req.count;
}

/* Back to an open node with no buffers. */
static void release(struct kvmv_hwjpeg *hw)
{
	enum v4l2_buf_type type;
	int got;

	if (hw->fd < 0)
		return;
	if (hw->out_on) {
		type = V4L2_BUF_TYPE_VIDEO_OUTPUT;
		xioctl(hw->fd, VIDIOC_STREAMOFF, &type);
	}
	if (hw->cap_on) {
		type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
		xioctl(hw->fd, VIDIOC_STREAMOFF, &type);
	}
	hw->out_on = hw->cap_on = 0;
	unmap(&hw->out);
	unmap(&hw->cap);
	if (hw->width) {
		request(hw->fd, V4L2_BUF_TYPE_VIDEO_OUTPUT,
			hw->import ? V4L2_MEMORY_DMABUF : V4L2_MEMORY_MMAP, 0, &got);
		request(hw->fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP, 0, &got);
	}
	hw->width = hw->height = hw->stride = 0;
	hw->out_size = 0;
}

void kvmv_hwjpeg_close(struct kvmv_hwjpeg *hw)
{
	release(hw);
	if (hw->fd >= 0)
		close(hw->fd);
	kvmv_hwjpeg_init(hw);
}

static int map_one(struct kvmv_hwjpeg *hw, enum v4l2_buf_type type, int prot,
		   struct kvmv_buf *buf)
{
	struct v4l2_buffer vbuf;

	memset(&vbuf, 0, sizeof(vbuf));
	vbuf.type = type;
	vbuf.memory = V4L2_MEMORY_MMAP;
	vbuf.index = 0;
	if (xioctl(hw->fd, VIDIOC_QUERYBUF, &vbuf))
		return fail(hw, "QUERYBUF");
	buf->addr = mmap(NULL, vbuf.length, prot, MAP_SHARED, hw->fd, vbuf.m.offset);
	if (buf->addr == MAP_FAILED) {
		buf->addr = NULL;
		return fail(hw, "mmap");
	}
	buf->length = vbuf.length;
	return 0;
}

static int configure(struct kvmv_hwjpeg *hw, unsigned int width,
		     unsigned int height, unsigned int stride, int import)
{
	struct v4l2_format fmt;
	enum v4l2_buf_type type;
	int got;

	release(hw);

	memset(&fmt, 0, sizeof(fmt));
	fmt.type = V4L2_BUF_TYPE_VIDEO_OUTPUT;
	fmt.fmt.pix.width = width;
	fmt.fmt.pix.height = height;
	fmt.fmt.pix.pixelformat = V4L2_PIX_FMT_NV12;
	fmt.fmt.pix.field = V4L2_FIELD_NONE;
	fmt.fmt.pix.bytesperline = stride;
	if (xioctl(hw->fd, VIDIOC_S_FMT, &fmt))
		return fail(hw, "OUTPUT S_FMT");
	if (fmt.fmt.pix.pixelformat != V4L2_PIX_FMT_NV12 ||
	    fmt.fmt.pix.width != width || fmt.fmt.pix.height != height ||
	    fmt.fmt.pix.bytesperline != stride || fmt.fmt.pix.sizeimage == 0) {
		snprintf(hw->error, sizeof(hw->error),
			 "JPEG unit took %.4s %ux%u stride %u for NV12 %ux%u stride %u",
			 (const char *)&fmt.fmt.pix.pixelformat, fmt.fmt.pix.width,
			 fmt.fmt.pix.height, fmt.fmt.pix.bytesperline, width, height,
			 stride);
		errno = EINVAL;
		return -1;
	}
	hw->out_size = fmt.fmt.pix.sizeimage;

	memset(&fmt, 0, sizeof(fmt));
	fmt.type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
	fmt.fmt.pix.width = width;
	fmt.fmt.pix.height = height;
	fmt.fmt.pix.pixelformat = V4L2_PIX_FMT_JPEG;
	if (xioctl(hw->fd, VIDIOC_S_FMT, &fmt))
		return fail(hw, "CAPTURE S_FMT");
	if (fmt.fmt.pix.pixelformat != V4L2_PIX_FMT_JPEG) {
		snprintf(hw->error, sizeof(hw->error), "JPEG unit does not produce JPEG");
		errno = EINVAL;
		return -1;
	}

	/* From here on release() has buffers to free. */
	hw->width = width;
	hw->height = height;
	hw->stride = stride;
	hw->import = import;

	request(hw->fd, V4L2_BUF_TYPE_VIDEO_OUTPUT,
		import ? V4L2_MEMORY_DMABUF : V4L2_MEMORY_MMAP, 1, &got);
	if (got < 1)
		return fail(hw, "OUTPUT REQBUFS");
	if (!import && map_one(hw, V4L2_BUF_TYPE_VIDEO_OUTPUT,
			       PROT_READ | PROT_WRITE, &hw->out))
		return -1;
	request(hw->fd, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP, 1, &got);
	if (got < 1)
		return fail(hw, "CAPTURE REQBUFS");
	if (map_one(hw, V4L2_BUF_TYPE_VIDEO_CAPTURE, PROT_READ, &hw->cap))
		return -1;

	type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
	if (xioctl(hw->fd, VIDIOC_STREAMON, &type))
		return fail(hw, "CAPTURE STREAMON");
	hw->cap_on = 1;
	type = V4L2_BUF_TYPE_VIDEO_OUTPUT;
	if (xioctl(hw->fd, VIDIOC_STREAMON, &type))
		return fail(hw, "OUTPUT STREAMON");
	hw->out_on = 1;
	return 0;
}

static int set_quality(struct kvmv_hwjpeg *hw, int quality)
{
	struct v4l2_ext_control control;
	struct v4l2_ext_controls list;

	if (quality == hw->quality)
		return 0;
	memset(&control, 0, sizeof(control));
	memset(&list, 0, sizeof(list));
	control.id = V4L2_CID_JPEG_COMPRESSION_QUALITY;
	control.value = quality;
	list.which = V4L2_CTRL_WHICH_CUR_VAL;
	list.count = 1;
	list.controls = &control;
	if (xioctl(hw->fd, VIDIOC_S_EXT_CTRLS, &list))
		return fail(hw, "set V4L2_CID_JPEG_COMPRESSION_QUALITY");
	hw->quality = quality;
	return 0;
}

static uint64_t now_ms(void)
{
	struct timespec ts;

	clock_gettime(CLOCK_MONOTONIC, &ts);
	return (uint64_t)ts.tv_sec * 1000U + (uint64_t)ts.tv_nsec / 1000000U;
}

static int wait_dqbuf(struct kvmv_hwjpeg *hw, enum v4l2_buf_type type,
		      enum v4l2_memory memory, short events, struct v4l2_buffer *buf)
{
	uint64_t deadline = now_ms() + KVMV_HWJPEG_TIMEOUT_MS;

	for (;;) {
		struct pollfd pfd = { .fd = hw->fd, .events = events };
		uint64_t now;

		memset(buf, 0, sizeof(*buf));
		buf->type = type;
		buf->memory = memory;
		if (xioctl(hw->fd, VIDIOC_DQBUF, buf) == 0)
			return 0;
		if (errno != EAGAIN)
			return -1;
		now = now_ms();
		if (now >= deadline) {
			errno = ETIMEDOUT;
			return -1;
		}
		if (poll(&pfd, 1, (int)(deadline - now)) < 0 && errno != EINTR)
			return -1;
	}
}

static void copy_rows(uint8_t *dst, unsigned int dst_stride, const uint8_t *src,
		      unsigned int src_stride, unsigned int width, unsigned int rows)
{
	unsigned int y;

	if (dst_stride == src_stride) {
		memcpy(dst, src, (size_t)src_stride * rows);
		return;
	}
	for (y = 0; y < rows; y++)
		memcpy(dst + (size_t)y * dst_stride, src + (size_t)y * src_stride, width);
}

int kvmv_hwjpeg_encode(struct kvmv_hwjpeg *hw, const struct kvmv_nv12 *image,
		       int src_fd, size_t src_size, int quality, int def_quality,
		       const uint8_t **data, size_t *size)
{
	struct v4l2_buffer buf;
	int import = src_fd >= 0;

	*data = NULL;
	*size = 0;
	if (hw->fd < 0) {
		snprintf(hw->error, sizeof(hw->error), "JPEG unit not open");
		errno = EBADF;
		return -1;
	}
	if (!kvmv_hwjpeg_fits(image->width, image->height, image->stride)) {
		snprintf(hw->error, sizeof(hw->error),
			 "JPEG unit cannot take %ux%u stride %u", image->width,
			 image->height, image->stride);
		errno = EINVAL;
		return -1;
	}
	if (hw->width != image->width || hw->height != image->height ||
	    hw->stride != image->stride || hw->import != import || !hw->out_on) {
		if (configure(hw, image->width, image->height, image->stride, import)) {
			int saved = errno;

			release(hw);
			errno = saved;
			return -1;
		}
	}
	if (import && src_size < hw->out_size) {
		snprintf(hw->error, sizeof(hw->error),
			 "JPEG unit needs a %u byte source, the buffer has %zu",
			 hw->out_size, src_size);
		errno = EINVAL;
		return -1;
	}

	if (quality <= 0)
		quality = def_quality;
	if (quality > 100)
		quality = 100;
	if (set_quality(hw, quality))
		return -1;

	memset(&buf, 0, sizeof(buf));
	buf.type = V4L2_BUF_TYPE_VIDEO_CAPTURE;
	buf.memory = V4L2_MEMORY_MMAP;
	buf.index = 0;
	if (xioctl(hw->fd, VIDIOC_QBUF, &buf))
		return fail(hw, "CAPTURE QBUF");

	memset(&buf, 0, sizeof(buf));
	buf.type = V4L2_BUF_TYPE_VIDEO_OUTPUT;
	buf.index = 0;
	buf.bytesused = hw->out_size;
	if (import) {
		buf.memory = V4L2_MEMORY_DMABUF;
		buf.m.fd = src_fd;
		buf.length = (uint32_t)src_size;
	} else {
		buf.memory = V4L2_MEMORY_MMAP;
		copy_rows(hw->out.addr, hw->stride, image->y, image->stride,
			  image->width, image->height);
		copy_rows((uint8_t *)hw->out.addr + (size_t)hw->stride * hw->height,
			  hw->stride, image->uv, image->stride, image->width,
			  image->height / 2);
	}
	if (xioctl(hw->fd, VIDIOC_QBUF, &buf)) {
		fail(hw, "OUTPUT QBUF");
		/* The CAPTURE buffer is queued: start over on the next call. */
		release(hw);
		return -1;
	}

	if (wait_dqbuf(hw, V4L2_BUF_TYPE_VIDEO_CAPTURE, V4L2_MEMORY_MMAP, POLLIN, &buf)) {
		fail(hw, "CAPTURE DQBUF");
		release(hw);
		return -1;
	}
	*size = buf.bytesused;
	if (buf.flags & V4L2_BUF_FLAG_ERROR || buf.bytesused < 4 ||
	    buf.bytesused > hw->cap.length) {
		snprintf(hw->error, sizeof(hw->error),
			 "JPEG unit returned flags %#x, %u bytes", buf.flags,
			 buf.bytesused);
		*size = 0;
	}
	if (wait_dqbuf(hw, V4L2_BUF_TYPE_VIDEO_OUTPUT,
		       import ? V4L2_MEMORY_DMABUF : V4L2_MEMORY_MMAP, POLLOUT, &buf)) {
		fail(hw, "OUTPUT DQBUF");
		release(hw);
		*size = 0;
		return -1;
	}
	if (*size == 0) {
		errno = EIO;
		return -1;
	}
	*data = hw->cap.addr;
	return 0;
}
