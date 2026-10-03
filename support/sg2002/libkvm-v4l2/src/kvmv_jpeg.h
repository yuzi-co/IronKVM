/*
 * Software JPEG for kvmv_read_img (MJPEG) on the mainline kernel.
 *
 * There is no mainline driver for the SG2002's JPEG unit (ironkvm-dist#36), so
 * pictures are encoded on the CPU with libjpeg-turbo's TurboJPEG API, linked
 * statically: 4:2:0, fast integer DCT, from planar YUV, no RGB anywhere. The
 * board measurements (ironkvm-dist socs/sophgo-sg2002/mainline/jpeg-bench) put
 * one encode at about 210 ms at 1080p, 94 ms at 720p and 53 ms at 540p, all
 * of it on the only core, so encoding happens per call and never at the
 * stream rate.
 *
 * The work is split so the caller can hold the pipeline only while the
 * picture is read out: kvmv_jpeg_load copies an NV12 picture into the
 * encoder's own planes (the luma rows as they are, the chroma deinterleaved),
 * and kvmv_jpeg_compress encodes those planes with no device involved.
 */
#ifndef KVMV_JPEG_H
#define KVMV_JPEG_H

#include <stddef.h>
#include <stdint.h>

#include "kvmv_pipeline.h"

#define KVMV_JPEG_QUALITY_DEFAULT 80

struct kvmv_jpeg {
	void *tj; /* tjhandle, created on first use */
	uint8_t *plane[3]; /* Y, Cb, Cr, tightly packed */
	size_t plane_size[3];
	unsigned int width, height; /* of the loaded picture; 0 when none */
	uint8_t *out; /* TurboJPEG's output buffer, kept between calls */
	size_t out_size;
};

void kvmv_jpeg_init(struct kvmv_jpeg *jpeg);
void kvmv_jpeg_free(struct kvmv_jpeg *jpeg);

/*
 * Copy an NV12 picture into the encoder's planes. width and height must be
 * even. Returns 0, or -1 when out of memory or the picture is unusable.
 */
int kvmv_jpeg_load(struct kvmv_jpeg *jpeg, const struct kvmv_nv12 *image);

/*
 * Encode the loaded planes at quality (clamped to 1..100; 0 means the
 * default). On 0, *data and *size describe the JPEG, valid until the next
 * call. Returns -1 on failure, with the reason in error when it is not NULL.
 */
int kvmv_jpeg_compress(struct kvmv_jpeg *jpeg, int quality,
		       const uint8_t **data, size_t *size, char *error,
		       size_t error_size);

/* Split interleaved CbCr rows into planar Cb and Cr. Exposed for the test. */
void kvmv_nv12_split_chroma(const uint8_t *uv, unsigned int stride,
			    unsigned int width, unsigned int height,
			    uint8_t *cb, uint8_t *cr);

#endif
