/*
 * H.264 and H.265 Annex-B inspection for libkvm-v4l2.
 *
 * Nothing here touches a device, so all of it runs in the host unit test.
 */
#ifndef KVMV_ANNEXB_H
#define KVMV_ANNEXB_H

#include <stddef.h>
#include <stdint.h>

/* Which bitstream a buffer holds: the NAL header differs. */
enum kvmv_codec {
	KVMV_CODEC_KIND_H264 = 0,
	KVMV_CODEC_KIND_HEVC = 1,
};

/* H.264 NAL unit types */
#define KVMV_NAL_SLICE 1
#define KVMV_NAL_IDR 5
#define KVMV_NAL_SEI 6
#define KVMV_NAL_SPS 7
#define KVMV_NAL_PPS 8
#define KVMV_NAL_AUD 9

/*
 * H.265 NAL unit types. 0 to 9 are the coded slices of ordinary pictures; 16
 * to 21 the random access points (IRAP: BLA, IDR, CRA), which a decoder can
 * start at; 22 to 31 are reserved.
 */
#define KVMV_HEVC_NAL_TRAIL_N 0
#define KVMV_HEVC_NAL_RASL_R 9
#define KVMV_HEVC_NAL_BLA_W_LP 16
#define KVMV_HEVC_NAL_IDR_W_RADL 19
#define KVMV_HEVC_NAL_IDR_N_LP 20
#define KVMV_HEVC_NAL_CRA 21
#define KVMV_HEVC_NAL_VPS 32
#define KVMV_HEVC_NAL_SPS 33
#define KVMV_HEVC_NAL_PPS 34
#define KVMV_HEVC_NAL_AUD 35
#define KVMV_HEVC_NAL_SEI_PREFIX 39
#define KVMV_HEVC_NAL_SEI_SUFFIX 40

/* What kvmv_au_type answers besides the IMG_ codes of kvm_vision.h. */
#define KVMV_AU_HEADERS_ONLY (-100) /* parameter sets and no picture */
#define KVMV_AU_NO_PICTURE (-101) /* nothing a decoder can show */

/* Largest VPS, SPS and PPS the cache keeps, start code included. */
#define KVMV_PS_MAX 256

struct kvmv_nal {
	size_t offset; /* first byte of the start code */
	size_t start; /* first byte after the start code (the NAL header) */
	size_t size; /* bytes from start to the next start code */
	uint8_t type; /* nal_unit_type, read as the codec defines it */
};

struct kvmv_au_info {
	unsigned int nal_count;
	unsigned int vps; /* H.265 only */
	unsigned int sps;
	unsigned int pps;
	unsigned int idr; /* H.264 IDR slices; H.265 IRAP slices (16 to 21) */
	unsigned int slice; /* other coded slices: H.264 1 to 4, H.265 0 to 9 */
	unsigned int sei;
	unsigned int aud;
	unsigned int other;
};

/*
 * Split an Annex-B buffer into NAL units. Returns how many were stored.
 * kvmv_annexb_split reads H.264 NAL headers.
 */
size_t kvmv_annexb_split_codec(enum kvmv_codec codec, const uint8_t *buf,
			       size_t len, struct kvmv_nal *nals, size_t max_nals);
size_t kvmv_annexb_split(const uint8_t *buf, size_t len, struct kvmv_nal *nals,
			 size_t max_nals);

/* Count the NAL types in one access unit. kvmv_au_inspect is H.264. */
void kvmv_au_inspect_codec(enum kvmv_codec codec, const uint8_t *buf, size_t len,
			   struct kvmv_au_info *info);
void kvmv_au_inspect(const uint8_t *buf, size_t len, struct kvmv_au_info *info);

/*
 * Classify one access unit the way the vendor library does: a random access
 * picture anywhere in it (an H.264 IDR, an H.265 IRAP) makes it
 * IMG_H264_TYPE_IF (3), any other coded slice makes it IMG_H264_TYPE_PF (4).
 * Parameter sets alone answer KVMV_AU_HEADERS_ONLY, and a buffer with no
 * picture at all answers KVMV_AU_NO_PICTURE.
 */
int kvmv_au_type(const struct kvmv_au_info *info);

/*
 * The last parameter sets the encoder produced, each with a 4-byte start
 * code: SPS and PPS for H.264, VPS, SPS and PPS for H.265.
 */
struct kvmv_ps_cache {
	enum kvmv_codec codec;
	uint8_t vps[KVMV_PS_MAX];
	size_t vps_len;
	uint8_t sps[KVMV_PS_MAX];
	size_t sps_len;
	uint8_t pps[KVMV_PS_MAX];
	size_t pps_len;
};

/* Empty the cache for a stream of codec; kvmv_ps_cache_reset is H.264. */
void kvmv_ps_cache_reset_codec(struct kvmv_ps_cache *cache, enum kvmv_codec codec);
void kvmv_ps_cache_reset(struct kvmv_ps_cache *cache);
void kvmv_ps_cache_update(struct kvmv_ps_cache *cache, const uint8_t *buf,
			  size_t len);

/*
 * The bytes to put in front of a keyframe access unit that does not carry all
 * of its parameter sets itself, so that every keyframe decodes from cold.
 * Writes the cached sets (VPS for H.265, then SPS, then PPS) into out and
 * returns their length; returns 0 when the unit needs nothing, when the cache
 * lacks a set, or when out is too small.
 */
size_t kvmv_ps_cache_prefix(const struct kvmv_ps_cache *cache,
			    const struct kvmv_au_info *info, uint8_t *out,
			    size_t out_size);

/*
 * Clear vps_extension_flag in every H.265 VPS of buf, in place. The WAVE420L
 * sets it with no extension behind it, and Chrome's hardware HEVC decoder
 * refuses a VPS with the flag set (every picture is a decoding error); ffmpeg
 * ignores it. The flag becomes 0 and the stop bit follows it; bytes the
 * extension took become trailing zero bytes, so the length does not change.
 * Returns how many VPS were changed.
 */
unsigned int kvmv_hevc_clear_vps_extension(uint8_t *buf, size_t len);

/*
 * Picture size from an SPS, the conformance window or frame cropping applied.
 * nal points at the NAL header, without the start code. Returns 0 on success.
 */
int kvmv_h264_sps_size(const uint8_t *nal, size_t len, unsigned int *width,
		       unsigned int *height);
int kvmv_hevc_sps_size(const uint8_t *nal, size_t len, unsigned int *width,
		       unsigned int *height);

#endif
