/*
 * Hardware JPEG for kvmv_read_img: the SG2002's JPEG unit through its V4L2
 * mem2mem driver, sg2002-jpeg (ironkvm-dist patch 0907, #36).
 *
 * The driver takes NV12 on its OUTPUT queue and gives a baseline 4:2:0 JPEG,
 * header included, on its CAPTURE queue. Quality is
 * V4L2_CID_JPEG_COMPRESSION_QUALITY. A 1080p picture takes about 5 ms on the
 * unit against about 210 ms in libjpeg-turbo on the CPU.
 *
 * The source is the snapshot scaler context's NV12 buffer. Two ways in:
 *
 *   import  the snapshot buffer's dma-buf is queued on OUTPUT as it is
 *           (V4L2_MEMORY_DMABUF), so no pixel passes through the CPU;
 *   copy    the picture is copied by the CPU into an OUTPUT buffer of the
 *           driver's own (V4L2_MEMORY_MMAP).
 *
 * Import is the default; KVMV_JPEG_IMPORT=0 selects the copy, for
 * comparison. Both answer the JPEG in the driver's CAPTURE buffer, which the
 * caller copies out with kvmv_copy_from_device before the next encode.
 *
 * The node is found by driver name. Without it (the vendor kernel, or a
 * mainline kernel without 0907) the library keeps the software encoder.
 */
#ifndef KVMV_HWJPEG_H
#define KVMV_HWJPEG_H

#include <stddef.h>
#include <stdint.h>

#include "kvmv_pipeline.h"

#define KVMV_HWJPEG_DRIVER "sg2002-jpeg"

struct kvmv_hwjpeg {
	int fd; /* -1 when closed */
	int import; /* the configured mode: 1 dma-buf import, 0 copy */
	unsigned int width, height, stride; /* the configured picture; 0 when none */
	uint32_t out_size; /* the OUTPUT sizeimage the driver asked for */
	int quality; /* the control's value, 0 when not set yet */
	int out_on, cap_on;
	struct kvmv_buf out; /* copy mode: the mapped OUTPUT buffer */
	struct kvmv_buf cap; /* the mapped CAPTURE buffer */
	char error[160];
};

void kvmv_hwjpeg_init(struct kvmv_hwjpeg *hw);

/* Open the node at path. Returns 0, or -1 with the reason in hw->error. */
int kvmv_hwjpeg_open(struct kvmv_hwjpeg *hw, const char *path);

/* Stop streaming, free the buffers and close the node. Safe when closed. */
void kvmv_hwjpeg_close(struct kvmv_hwjpeg *hw);

/*
 * Whether the unit can encode a picture of this size and stride as it is:
 * the driver rounds the width up to 16 and writes the rounded width into
 * the JPEG header, so any other width would show a stripe of padding.
 */
int kvmv_hwjpeg_fits(unsigned int width, unsigned int height, unsigned int stride);

/*
 * The bytes an NV12 source buffer needs for the unit: it reads chroma rows
 * up to the next multiple of 16 picture lines, past a tight 4:2:0 buffer
 * when the height is not a multiple of 16 (1080 is not).
 */
size_t kvmv_hwjpeg_src_size(unsigned int stride, unsigned int height);

/*
 * Encode image. With src_fd >= 0 its dma-buf (src_size bytes, holding the
 * picture at offset 0 with the chroma right after height rows of stride) is
 * imported; with src_fd < 0 the CPU copies image->y and image->uv. quality
 * is clamped to 1..100, 0 meaning def_quality.
 *
 * On 0, *data and *size describe the JPEG inside the CAPTURE buffer, which
 * is mapped uncached: copy it with kvmv_copy_from_device. It stays valid
 * until the next call. Returns -1 with the reason in hw->error.
 */
int kvmv_hwjpeg_encode(struct kvmv_hwjpeg *hw, const struct kvmv_nv12 *image,
		       int src_fd, size_t src_size, int quality, int def_quality,
		       const uint8_t **data, size_t *size);

#endif
