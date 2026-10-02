/*
 * H.264 Annex-B inspection for libkvm-v4l2.
 *
 * Nothing here touches a device, so all of it runs in the host unit test.
 */
#ifndef KVMV_ANNEXB_H
#define KVMV_ANNEXB_H

#include <stddef.h>
#include <stdint.h>

#define KVMV_NAL_SLICE 1
#define KVMV_NAL_IDR 5
#define KVMV_NAL_SEI 6
#define KVMV_NAL_SPS 7
#define KVMV_NAL_PPS 8
#define KVMV_NAL_AUD 9

/* What kvmv_au_type answers besides the IMG_ codes of kvm_vision.h. */
#define KVMV_AU_HEADERS_ONLY (-100) /* parameter sets and no picture */
#define KVMV_AU_NO_PICTURE (-101) /* nothing a decoder can show */

/* Largest SPS and PPS the cache keeps, start code included. */
#define KVMV_PS_MAX 256

struct kvmv_nal {
	size_t offset; /* first byte of the start code */
	size_t start; /* first byte after the start code (the NAL header) */
	size_t size; /* bytes from start to the next start code */
	uint8_t type;
};

struct kvmv_au_info {
	unsigned int nal_count;
	unsigned int sps;
	unsigned int pps;
	unsigned int idr;
	unsigned int slice; /* non-IDR coded slices, types 1 to 4 */
	unsigned int sei;
	unsigned int aud;
	unsigned int other;
};

/* Split an Annex-B buffer into NAL units. Returns how many were stored. */
size_t kvmv_annexb_split(const uint8_t *buf, size_t len, struct kvmv_nal *nals,
			 size_t max_nals);

/* Count the NAL types in one access unit. */
void kvmv_au_inspect(const uint8_t *buf, size_t len, struct kvmv_au_info *info);

/*
 * Classify one access unit the way the vendor library does: an IDR picture
 * anywhere in it makes it IMG_H264_TYPE_IF (3), any other coded slice makes it
 * IMG_H264_TYPE_PF (4). Parameter sets alone answer KVMV_AU_HEADERS_ONLY, and
 * a buffer with no picture at all answers KVMV_AU_NO_PICTURE.
 */
int kvmv_au_type(const struct kvmv_au_info *info);

/* The last SPS and PPS the encoder produced, each with a 4-byte start code. */
struct kvmv_ps_cache {
	uint8_t sps[KVMV_PS_MAX];
	size_t sps_len;
	uint8_t pps[KVMV_PS_MAX];
	size_t pps_len;
};

void kvmv_ps_cache_reset(struct kvmv_ps_cache *cache);
void kvmv_ps_cache_update(struct kvmv_ps_cache *cache, const uint8_t *buf,
			  size_t len);

/*
 * The bytes to put in front of an IDR access unit that does not carry both
 * parameter sets itself, so that every keyframe decodes from cold. Writes the
 * cached SPS and PPS into out and returns their length; returns 0 when the
 * unit needs nothing, when the cache lacks either set, or when out is too
 * small.
 */
size_t kvmv_ps_cache_prefix(const struct kvmv_ps_cache *cache,
			    const struct kvmv_au_info *info, uint8_t *out,
			    size_t out_size);

/*
 * Picture size from an SPS. nal points at the NAL header byte, without the
 * start code. Returns 0 on success.
 */
int kvmv_h264_sps_size(const uint8_t *nal, size_t len, unsigned int *width,
		       unsigned int *height);

#endif
