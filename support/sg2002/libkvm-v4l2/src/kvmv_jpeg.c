#include "kvmv_jpeg.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <turbojpeg.h>

void kvmv_jpeg_init(struct kvmv_jpeg *jpeg)
{
	memset(jpeg, 0, sizeof(*jpeg));
}

void kvmv_jpeg_free(struct kvmv_jpeg *jpeg)
{
	unsigned int i;

	if (jpeg->tj != NULL)
		tj3Destroy(jpeg->tj);
	tj3Free(jpeg->out);
	for (i = 0; i < 3; i++)
		free(jpeg->plane[i]);
	kvmv_jpeg_init(jpeg);
}

static int grow(uint8_t **buf, size_t *have, size_t need)
{
	uint8_t *grown;

	if (*have >= need)
		return 0;
	grown = realloc(*buf, need);
	if (grown == NULL)
		return -1;
	*buf = grown;
	*have = need;
	return 0;
}

void kvmv_nv12_split_chroma(const uint8_t *uv, unsigned int stride,
			    unsigned int width, unsigned int height,
			    uint8_t *cb, uint8_t *cr)
{
	unsigned int cw = width / 2, ch = height / 2;
	unsigned int x, y;

	for (y = 0; y < ch; y++) {
		const uint8_t *row = uv + (size_t)y * stride;

		for (x = 0; x < cw; x++) {
			cb[x] = row[2 * x];
			cr[x] = row[2 * x + 1];
		}
		cb += cw;
		cr += cw;
	}
}

int kvmv_jpeg_load(struct kvmv_jpeg *jpeg, const struct kvmv_nv12 *image)
{
	size_t luma, chroma;
	unsigned int y;

	jpeg->width = 0;
	jpeg->height = 0;
	if (image->width < 2 || image->height < 2 || (image->width & 1) ||
	    (image->height & 1) || image->stride < image->width)
		return -1;
	luma = (size_t)image->width * image->height;
	chroma = luma / 4;
	if (grow(&jpeg->plane[0], &jpeg->plane_size[0], luma) ||
	    grow(&jpeg->plane[1], &jpeg->plane_size[1], chroma) ||
	    grow(&jpeg->plane[2], &jpeg->plane_size[2], chroma))
		return -1;
	for (y = 0; y < image->height; y++)
		memcpy(jpeg->plane[0] + (size_t)y * image->width,
		       image->y + (size_t)y * image->stride, image->width);
	kvmv_nv12_split_chroma(image->uv, image->stride, image->width,
			       image->height, jpeg->plane[1], jpeg->plane[2]);
	jpeg->width = image->width;
	jpeg->height = image->height;
	return 0;
}

static void set_error(char *error, size_t error_size, const char *what,
		      const char *detail)
{
	if (error != NULL && error_size)
		snprintf(error, error_size, "%s: %s", what, detail);
}

int kvmv_jpeg_compress(struct kvmv_jpeg *jpeg, int quality,
		       const uint8_t **data, size_t *size, char *error,
		       size_t error_size)
{
	const unsigned char *planes[3];
	size_t out_size, need;

	*data = NULL;
	*size = 0;
	if (jpeg->width == 0) {
		set_error(error, error_size, "JPEG", "no picture loaded");
		return -1;
	}
	if (quality <= 0)
		quality = KVMV_JPEG_QUALITY_DEFAULT;
	if (quality > 100)
		quality = 100;
	if (jpeg->tj == NULL) {
		jpeg->tj = tj3Init(TJINIT_COMPRESS);
		if (jpeg->tj == NULL) {
			set_error(error, error_size, "tj3Init", tj3GetErrorStr(NULL));
			return -1;
		}
	}
	if (tj3Set(jpeg->tj, TJPARAM_SUBSAMP, TJSAMP_420) ||
	    tj3Set(jpeg->tj, TJPARAM_QUALITY, quality) ||
	    tj3Set(jpeg->tj, TJPARAM_FASTDCT, 1) ||
	    tj3Set(jpeg->tj, TJPARAM_NOREALLOC, 1)) {
		set_error(error, error_size, "tj3Set", tj3GetErrorStr(jpeg->tj));
		return -1;
	}

	/*
	 * One output buffer of the worst-case size, kept between calls, so a
	 * steady stream allocates nothing. It is several MB of address space
	 * at 1080p, but only the pages a picture writes are ever touched.
	 */
	need = tj3JPEGBufSize((int)jpeg->width, (int)jpeg->height, TJSAMP_420);
	if (need == 0) {
		set_error(error, error_size, "tj3JPEGBufSize", tj3GetErrorStr(NULL));
		return -1;
	}
	if (jpeg->out_size < need) {
		tj3Free(jpeg->out);
		jpeg->out_size = 0;
		jpeg->out = tj3Alloc(need);
		if (jpeg->out == NULL) {
			set_error(error, error_size, "tj3Alloc", "out of memory");
			return -1;
		}
		jpeg->out_size = need;
	}

	planes[0] = jpeg->plane[0];
	planes[1] = jpeg->plane[1];
	planes[2] = jpeg->plane[2];
	out_size = jpeg->out_size;
	if (tj3CompressFromYUVPlanes8(jpeg->tj, planes, (int)jpeg->width, NULL,
				      (int)jpeg->height, &jpeg->out, &out_size)) {
		set_error(error, error_size, "tj3CompressFromYUVPlanes8",
			  tj3GetErrorStr(jpeg->tj));
		return -1;
	}
	*data = jpeg->out;
	*size = out_size;
	return 0;
}
