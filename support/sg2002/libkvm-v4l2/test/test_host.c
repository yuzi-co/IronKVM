/*
 * Host unit tests for libkvm-v4l2: Annex-B classification, parameter-set
 * handling, SPS parsing, the policy decisions, the frame slots, format
 * negotiation against mock capture, VPSS and Coda drivers, and the JPEG unit
 * against a mock sg2002-jpeg.
 *
 * Built and run by `make test`.
 */
#define _GNU_SOURCE
#include <errno.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <unistd.h>
#include <linux/videodev2.h>
#include <turbojpeg.h>

#include "kvm_vision.h"
#include "kvmv_annexb.h"
#include "kvmv_jpeg.h"
#include "kvmv_hwjpeg.h"
#include "kvmv_pipeline.h"
#include "kvmv_policy.h"
#include "kvmv_slots.h"

static int failures;
static int checks;

#define CHECK(cond)                                                          \
	do {                                                                 \
		checks++;                                                    \
		if (!(cond)) {                                               \
			failures++;                                          \
			fprintf(stderr, "%s:%d: CHECK failed: %s\n", __FILE__, \
				__LINE__, #cond);                            \
		}                                                            \
	} while (0)

#define CHECK_EQ(a, b)                                                       \
	do {                                                                 \
		long long _a = (long long)(a), _b = (long long)(b);          \
		checks++;                                                    \
		if (_a != _b) {                                              \
			failures++;                                          \
			fprintf(stderr, "%s:%d: %s == %lld, expected %lld\n", \
				__FILE__, __LINE__, #a, _a, _b);             \
		}                                                            \
	} while (0)

/* ---- Annex-B builders ---------------------------------------------- */

struct stream {
	uint8_t data[4096];
	size_t len;
};

static void put_nal(struct stream *s, int four_byte, const uint8_t *nal, size_t n)
{
	if (four_byte)
		s->data[s->len++] = 0;
	s->data[s->len++] = 0;
	s->data[s->len++] = 0;
	s->data[s->len++] = 1;
	memcpy(s->data + s->len, nal, n);
	s->len += n;
}

static const uint8_t sps_nal[] = { 0x67, 0x42, 0xc0, 0x28, 0xda, 0x01, 0xe0, 0x08, 0x9f, 0x96 };
static const uint8_t pps_nal[] = { 0x68, 0xce, 0x3c, 0x80 };
static const uint8_t idr_nal[] = { 0x65, 0x88, 0x84, 0x00, 0x33, 0xff };
static const uint8_t p_nal[] = { 0x41, 0x9a, 0x24, 0x6c, 0x00 , 0x10 };
static const uint8_t sei_nal[] = { 0x06, 0x05, 0x01, 0x80 };
static const uint8_t aud_nal[] = { 0x09, 0xf0 };

static void test_split(void)
{
	struct stream s = { .len = 0 };
	struct kvmv_nal nals[8];
	size_t n;

	put_nal(&s, 1, sps_nal, sizeof(sps_nal));
	put_nal(&s, 0, pps_nal, sizeof(pps_nal));
	put_nal(&s, 1, idr_nal, sizeof(idr_nal));
	n = kvmv_annexb_split(s.data, s.len, nals, 8);
	CHECK_EQ(n, 3);
	CHECK_EQ(nals[0].type, 7);
	CHECK_EQ(nals[0].offset, 0);
	CHECK_EQ(nals[0].start, 4);
	CHECK_EQ(nals[0].size, sizeof(sps_nal));
	CHECK_EQ(nals[1].type, 8);
	CHECK_EQ(nals[1].size, sizeof(pps_nal));
	CHECK_EQ(nals[2].type, 5);
	/* The 4-byte start code's leading zero is not part of the PPS. */
	CHECK_EQ(nals[2].offset, nals[1].start + sizeof(pps_nal));
	CHECK_EQ(nals[2].size, sizeof(idr_nal));

	CHECK_EQ(kvmv_annexb_split(NULL, 0, nals, 8), 0);
	CHECK_EQ(kvmv_annexb_split((const uint8_t *)"\x12\x34\x56", 3, nals, 8), 0);
	/* A start code with nothing after it is no NAL. */
	CHECK_EQ(kvmv_annexb_split((const uint8_t *)"\0\0\0\1", 4, nals, 8), 0);
	/* max_nals is honoured. */
	CHECK_EQ(kvmv_annexb_split(s.data, s.len, nals, 2), 2);
}

static int classify(const struct stream *s, struct kvmv_au_info *info)
{
	kvmv_au_inspect(s->data, s->len, info);
	return kvmv_au_type(info);
}

static void test_classify(void)
{
	struct kvmv_au_info info;
	struct stream key = { .len = 0 }, delta = { .len = 0 }, headers = { .len = 0 },
		      bare_idr = { .len = 0 }, noise = { .len = 0 };

	put_nal(&key, 1, aud_nal, sizeof(aud_nal));
	put_nal(&key, 1, sps_nal, sizeof(sps_nal));
	put_nal(&key, 1, pps_nal, sizeof(pps_nal));
	put_nal(&key, 0, sei_nal, sizeof(sei_nal));
	put_nal(&key, 1, idr_nal, sizeof(idr_nal));
	CHECK_EQ(classify(&key, &info), IMG_H264_TYPE_IF);
	CHECK_EQ(info.nal_count, 5);
	CHECK_EQ(info.sps, 1);
	CHECK_EQ(info.pps, 1);
	CHECK_EQ(info.idr, 1);
	CHECK_EQ(info.aud, 1);
	CHECK_EQ(info.sei, 1);

	put_nal(&delta, 1, p_nal, sizeof(p_nal));
	CHECK_EQ(classify(&delta, &info), IMG_H264_TYPE_PF);
	CHECK_EQ(info.slice, 1);

	put_nal(&headers, 1, sps_nal, sizeof(sps_nal));
	put_nal(&headers, 1, pps_nal, sizeof(pps_nal));
	CHECK_EQ(classify(&headers, &info), KVMV_AU_HEADERS_ONLY);

	put_nal(&bare_idr, 0, idr_nal, sizeof(idr_nal));
	CHECK_EQ(classify(&bare_idr, &info), IMG_H264_TYPE_IF);
	CHECK_EQ(info.sps, 0);

	memset(noise.data, 0x5a, 32);
	noise.len = 32;
	CHECK_EQ(classify(&noise, &info), KVMV_AU_NO_PICTURE);
}

static void test_ps_cache(void)
{
	struct kvmv_ps_cache cache;
	struct kvmv_au_info info;
	struct stream key = { .len = 0 }, bare_idr = { .len = 0 }, delta = { .len = 0 };
	uint8_t out[2 * KVMV_PS_MAX];
	size_t n;

	kvmv_ps_cache_reset(&cache);
	put_nal(&bare_idr, 0, idr_nal, sizeof(idr_nal));
	kvmv_au_inspect(bare_idr.data, bare_idr.len, &info);
	/* Nothing cached yet: nothing to prefix. */
	CHECK_EQ(kvmv_ps_cache_prefix(&cache, &info, out, sizeof(out)), 0);

	put_nal(&key, 1, sps_nal, sizeof(sps_nal));
	put_nal(&key, 1, pps_nal, sizeof(pps_nal));
	put_nal(&key, 1, idr_nal, sizeof(idr_nal));
	kvmv_ps_cache_update(&cache, key.data, key.len);
	CHECK_EQ(cache.sps_len, 4 + sizeof(sps_nal));
	CHECK_EQ(cache.pps_len, 4 + sizeof(pps_nal));

	/* A complete keyframe needs nothing. */
	kvmv_au_inspect(key.data, key.len, &info);
	CHECK_EQ(kvmv_ps_cache_prefix(&cache, &info, out, sizeof(out)), 0);

	/* A bare IDR gets SPS then PPS, each with a 4-byte start code. */
	kvmv_au_inspect(bare_idr.data, bare_idr.len, &info);
	n = kvmv_ps_cache_prefix(&cache, &info, out, sizeof(out));
	CHECK_EQ(n, 8 + sizeof(sps_nal) + sizeof(pps_nal));
	CHECK(memcmp(out, "\0\0\0\1", 4) == 0);
	CHECK(memcmp(out + 4, sps_nal, sizeof(sps_nal)) == 0);
	CHECK(memcmp(out + 8 + sizeof(sps_nal), pps_nal, sizeof(pps_nal)) == 0);

	/* The prefixed unit classifies as a complete keyframe. */
	{
		struct stream joined = { .len = 0 };

		memcpy(joined.data, out, n);
		memcpy(joined.data + n, bare_idr.data, bare_idr.len);
		joined.len = n + bare_idr.len;
		kvmv_au_inspect(joined.data, joined.len, &info);
		CHECK_EQ(info.sps, 1);
		CHECK_EQ(info.pps, 1);
		CHECK_EQ(info.idr, 1);
	}

	/* Delta frames are never prefixed. */
	put_nal(&delta, 1, p_nal, sizeof(p_nal));
	kvmv_au_inspect(delta.data, delta.len, &info);
	CHECK_EQ(kvmv_ps_cache_prefix(&cache, &info, out, sizeof(out)), 0);

	/* Too small an output: nothing rather than half. */
	kvmv_au_inspect(bare_idr.data, bare_idr.len, &info);
	CHECK_EQ(kvmv_ps_cache_prefix(&cache, &info, out, 8), 0);

	/* Trailing zero bytes do not end up in the cache. */
	{
		struct stream padded = { .len = 0 };
		uint8_t sps_padded[sizeof(sps_nal) + 2];

		memcpy(sps_padded, sps_nal, sizeof(sps_nal));
		sps_padded[sizeof(sps_nal)] = 0;
		sps_padded[sizeof(sps_nal) + 1] = 0;
		put_nal(&padded, 1, sps_padded, sizeof(sps_padded));
		kvmv_ps_cache_update(&cache, padded.data, padded.len);
		CHECK_EQ(cache.sps_len, 4 + sizeof(sps_nal));
	}
}

/* ---- SPS ----------------------------------------------------------- */

struct bit_writer {
	uint8_t rbsp[128];
	size_t bits;
};

static void put_bit(struct bit_writer *w, unsigned int bit)
{
	if (bit)
		w->rbsp[w->bits / 8] |= (uint8_t)(0x80 >> (w->bits % 8));
	w->bits++;
}

static void put_bits(struct bit_writer *w, unsigned int value, unsigned int n)
{
	while (n--)
		put_bit(w, (value >> n) & 1);
}

static void put_ue(struct bit_writer *w, unsigned int value)
{
	unsigned int v = value + 1;
	unsigned int len = 0;

	while ((v >> len) > 1)
		len++;
	put_bits(w, 0, len);
	put_bits(w, v, len + 1);
}

/* Finish the RBSP and escape it into a NAL with emulation prevention. */
static size_t finish_nal(struct bit_writer *w, uint8_t *nal, size_t *escapes)
{
	size_t bytes, i, n = 0;
	unsigned int zeros = 0;

	put_bit(w, 1); /* rbsp_stop_one_bit */
	bytes = (w->bits + 7) / 8;
	*escapes = 0;
	nal[n++] = 0x67;
	for (i = 0; i < bytes; i++) {
		if (zeros >= 2 && w->rbsp[i] <= 3) {
			nal[n++] = 3;
			(*escapes)++;
			zeros = 0;
		}
		nal[n++] = w->rbsp[i];
		zeros = w->rbsp[i] == 0 ? zeros + 1 : 0;
	}
	return n;
}

static size_t make_sps(uint8_t *nal, unsigned int profile, unsigned int sps_id,
		       unsigned int width_mbs, unsigned int height_mbs,
		       unsigned int crop_bottom, size_t *escapes)
{
	struct bit_writer w;

	memset(&w, 0, sizeof(w));
	put_bits(&w, profile, 8);
	put_bits(&w, 0, 8); /* constraint flags */
	put_bits(&w, 0, 8); /* level, zero to force an escape below */
	put_ue(&w, sps_id);
	if (profile == 100) {
		put_ue(&w, 1); /* chroma_format_idc 4:2:0 */
		put_ue(&w, 0);
		put_ue(&w, 0);
		put_bit(&w, 0);
		put_bit(&w, 1); /* seq_scaling_matrix_present */
		{
			unsigned int i;

			for (i = 0; i < 8; i++) {
				put_bit(&w, i == 0); /* list 0 present */
				if (i == 0) {
					unsigned int j;

					/* delta_scale 0 for every entry */
					for (j = 0; j < 16; j++)
						put_ue(&w, 0);
				}
			}
		}
	}
	put_ue(&w, 0); /* log2_max_frame_num_minus4 */
	put_ue(&w, 0); /* pic_order_cnt_type */
	put_ue(&w, 0); /* log2_max_pic_order_cnt_lsb_minus4 */
	put_ue(&w, 1); /* max_num_ref_frames */
	put_bit(&w, 0);
	put_ue(&w, width_mbs - 1);
	put_ue(&w, height_mbs - 1);
	put_bit(&w, 1); /* frame_mbs_only */
	put_bit(&w, 1); /* direct_8x8_inference */
	put_bit(&w, crop_bottom != 0);
	if (crop_bottom) {
		put_ue(&w, 0);
		put_ue(&w, 0);
		put_ue(&w, 0);
		put_ue(&w, crop_bottom);
	}
	put_bit(&w, 0); /* vui_parameters_present */
	return finish_nal(&w, nal, escapes);
}

static void test_sps(void)
{
	uint8_t nal[160];
	unsigned int w = 0, h = 0;
	size_t escapes;
	size_t n;

	/* 1920x1088 coded, cropped by 8 lines to 1080. sps_id 63 makes the
	 * fourth byte 0x02 after two zero bytes, so an escape is needed. */
	n = make_sps(nal, 66, 63, 120, 68, 4, &escapes);
	CHECK(escapes >= 1);
	CHECK_EQ(kvmv_h264_sps_size(nal, n, &w, &h), 0);
	CHECK_EQ(w, 1920);
	CHECK_EQ(h, 1080);

	n = make_sps(nal, 100, 0, 80, 45, 0, &escapes);
	CHECK_EQ(kvmv_h264_sps_size(nal, n, &w, &h), 0);
	CHECK_EQ(w, 1280);
	CHECK_EQ(h, 720);

	/* Not an SPS, or cut short. */
	CHECK(kvmv_h264_sps_size(pps_nal, sizeof(pps_nal), &w, &h) != 0);
	n = make_sps(nal, 66, 0, 120, 68, 4, &escapes);
	CHECK(kvmv_h264_sps_size(nal, 5, &w, &h) != 0);
}

/* ---- H.265 --------------------------------------------------------- */

/* The parameter sets the WAVE420L wrote for 1920x1080 Main, level 4.1 (trial 13). */
static const uint8_t hevc_vps[] = {
	0x40, 0x01, 0x0c, 0x01, 0xff, 0xff, 0x01, 0x60, 0x00, 0x00, 0x03, 0x00,
	0x00, 0x03, 0x00, 0x00, 0x03, 0x00, 0x00, 0x03, 0x00, 0x7b, 0xac, 0x0c,
	0x00, 0x00, 0x03, 0x00, 0x04, 0x00, 0x00, 0x03, 0x00, 0x79, 0xa0,
};
static const uint8_t hevc_sps[] = {
	0x42, 0x01, 0x01, 0x01, 0x60, 0x00, 0x00, 0x03, 0x00, 0x00, 0x03, 0x00,
	0x00, 0x03, 0x00, 0x00, 0x03, 0x00, 0x7b, 0xa0, 0x03, 0xc0, 0x80, 0x10,
	0xe5, 0x96, 0xb9, 0x24, 0xc1, 0xae, 0x59, 0x90,
};
static const uint8_t hevc_pps[] = { 0x44, 0x01, 0xe0, 0x72, 0xb0, 0x26, 0x40 };
static const uint8_t hevc_idr[] = { 0x26, 0x01, 0xac, 0x19, 0x60, 0xe0 }; /* type 19 */
static const uint8_t hevc_cra[] = { 0x2a, 0x01, 0xac, 0x19, 0x60, 0xe0 }; /* type 21 */
static const uint8_t hevc_trail[] = { 0x02, 0x01, 0xd0, 0x08, 0x31, 0x22 }; /* type 1 */

static void test_hevc(void)
{
	const enum kvmv_codec hevc = KVMV_CODEC_KIND_HEVC;
	struct stream key = { .len = 0 }, delta = { .len = 0 }, bare = { .len = 0 },
		      cra = { .len = 0 }, headers = { .len = 0 };
	struct kvmv_au_info info;
	struct kvmv_ps_cache cache;
	struct kvmv_nal nals[8];
	uint8_t out[3 * KVMV_PS_MAX];
	unsigned int w = 0, h = 0;
	size_t n;

	put_nal(&key, 1, hevc_vps, sizeof(hevc_vps));
	put_nal(&key, 1, hevc_sps, sizeof(hevc_sps));
	put_nal(&key, 1, hevc_pps, sizeof(hevc_pps));
	put_nal(&key, 1, hevc_idr, sizeof(hevc_idr));
	CHECK_EQ(kvmv_annexb_split_codec(hevc, key.data, key.len, nals, 8), 4);
	CHECK_EQ(nals[0].type, KVMV_HEVC_NAL_VPS);
	CHECK_EQ(nals[1].type, KVMV_HEVC_NAL_SPS);
	CHECK_EQ(nals[2].type, KVMV_HEVC_NAL_PPS);
	CHECK_EQ(nals[3].type, KVMV_HEVC_NAL_IDR_W_RADL);
	kvmv_au_inspect_codec(hevc, key.data, key.len, &info);
	CHECK_EQ(kvmv_au_type(&info), IMG_H264_TYPE_IF);
	CHECK_EQ(info.vps, 1);
	CHECK_EQ(info.sps, 1);
	CHECK_EQ(info.pps, 1);
	CHECK_EQ(info.idr, 1);

	/* Read as H.264, the same bytes are no keyframe. */
	kvmv_au_inspect(key.data, key.len, &info);
	CHECK(info.idr == 0);

	put_nal(&delta, 1, hevc_trail, sizeof(hevc_trail));
	kvmv_au_inspect_codec(hevc, delta.data, delta.len, &info);
	CHECK_EQ(kvmv_au_type(&info), IMG_H264_TYPE_PF);

	/* A CRA is a random access point too. */
	put_nal(&cra, 1, hevc_cra, sizeof(hevc_cra));
	kvmv_au_inspect_codec(hevc, cra.data, cra.len, &info);
	CHECK_EQ(kvmv_au_type(&info), IMG_H264_TYPE_IF);

	put_nal(&headers, 1, hevc_vps, sizeof(hevc_vps));
	kvmv_au_inspect_codec(hevc, headers.data, headers.len, &info);
	CHECK_EQ(kvmv_au_type(&info), KVMV_AU_HEADERS_ONLY);

	/* The cache gives a bare IDR the VPS, SPS and PPS, in that order. */
	kvmv_ps_cache_reset_codec(&cache, hevc);
	put_nal(&bare, 0, hevc_idr, sizeof(hevc_idr));
	kvmv_au_inspect_codec(hevc, bare.data, bare.len, &info);
	CHECK_EQ(kvmv_ps_cache_prefix(&cache, &info, out, sizeof(out)), 0);
	kvmv_ps_cache_update(&cache, key.data, key.len);
	CHECK_EQ(cache.vps_len, 4 + sizeof(hevc_vps));
	CHECK_EQ(cache.sps_len, 4 + sizeof(hevc_sps));
	CHECK_EQ(cache.pps_len, 4 + sizeof(hevc_pps));
	n = kvmv_ps_cache_prefix(&cache, &info, out, sizeof(out));
	CHECK_EQ(n, 12 + sizeof(hevc_vps) + sizeof(hevc_sps) + sizeof(hevc_pps));
	CHECK(memcmp(out + 4, hevc_vps, sizeof(hevc_vps)) == 0);
	CHECK(memcmp(out + 8 + sizeof(hevc_vps), hevc_sps, sizeof(hevc_sps)) == 0);
	/* A complete keyframe and a delta frame need nothing. */
	kvmv_au_inspect_codec(hevc, key.data, key.len, &info);
	CHECK_EQ(kvmv_ps_cache_prefix(&cache, &info, out, sizeof(out)), 0);
	kvmv_au_inspect_codec(hevc, delta.data, delta.len, &info);
	CHECK_EQ(kvmv_ps_cache_prefix(&cache, &info, out, sizeof(out)), 0);
	/* SPS and PPS without the VPS are not complete. */
	{
		struct stream partial = { .len = 0 };

		put_nal(&partial, 1, hevc_sps, sizeof(hevc_sps));
		put_nal(&partial, 1, hevc_pps, sizeof(hevc_pps));
		put_nal(&partial, 1, hevc_idr, sizeof(hevc_idr));
		kvmv_au_inspect_codec(hevc, partial.data, partial.len, &info);
		CHECK(kvmv_ps_cache_prefix(&cache, &info, out, sizeof(out)) > 0);
	}

	/* The VPS loses its empty extension: the flag bit and nothing else. */
	{
		struct stream fix = { .len = 0 };
		uint8_t want[sizeof(hevc_vps)];
		size_t i, vps_at;

		put_nal(&fix, 1, hevc_vps, sizeof(hevc_vps));
		put_nal(&fix, 1, hevc_sps, sizeof(hevc_sps));
		put_nal(&fix, 1, hevc_idr, sizeof(hevc_idr));
		memcpy(want, hevc_vps, sizeof(want));
		want[sizeof(want) - 1] = 0x40; /* 0xa0: flag 1, data 0, stop; now flag 0, stop */
		CHECK_EQ(kvmv_hevc_clear_vps_extension(fix.data, fix.len), 1);
		vps_at = 4;
		CHECK(memcmp(fix.data + vps_at, want, sizeof(want)) == 0);
		/* The rest of the unit is untouched, and a second pass finds nothing. */
		i = 4 + sizeof(hevc_vps) + 4;
		CHECK(memcmp(fix.data + i, hevc_sps, sizeof(hevc_sps)) == 0);
		CHECK_EQ(kvmv_hevc_clear_vps_extension(fix.data, fix.len), 0);
		/* An H.264 buffer is left alone. */
		CHECK_EQ(kvmv_hevc_clear_vps_extension(key.data, 0), 0);
	}

	CHECK_EQ(kvmv_hevc_sps_size(hevc_sps, sizeof(hevc_sps), &w, &h), 0);
	CHECK_EQ(w, 1920);
	CHECK_EQ(h, 1080);
	CHECK(kvmv_hevc_sps_size(hevc_vps, sizeof(hevc_vps), &w, &h) != 0);
	CHECK(kvmv_hevc_sps_size(hevc_sps, 10, &w, &h) != 0);
}

/* ---- policy -------------------------------------------------------- */

static void test_plan(void)
{
	struct kvmv_plan plan;

	CHECK_EQ(kvmv_plan_output(1920, 1080, 0, 0, &plan), 0);
	CHECK_EQ(plan.out_width, 1920);
	CHECK_EQ(plan.out_height, 1080);
	CHECK_EQ(plan.coded_height, 1088);

	CHECK_EQ(kvmv_plan_output(1920, 1080, 1280, 720, &plan), 0);
	CHECK_EQ(plan.out_width, 1280);
	CHECK_EQ(plan.out_height, 720);
	CHECK_EQ(plan.coded_height, 720);

	/* 1366 is not a multiple of 16; the stride has to be. */
	CHECK_EQ(kvmv_plan_output(1920, 1080, 1366, 768, &plan), 0);
	CHECK_EQ(plan.out_width, 1360);
	CHECK_EQ(plan.out_height, 768);

	/* Odd heights are made even for 4:2:0. */
	CHECK_EQ(kvmv_plan_output(1920, 1080, 800, 601, &plan), 0);
	CHECK_EQ(plan.out_height, 600);
	CHECK_EQ(plan.coded_height, 608);

	/* No upscaling. */
	CHECK_EQ(kvmv_plan_output(1920, 1080, 2560, 1440, &plan), 0);
	CHECK_EQ(plan.out_width, 1920);
	CHECK_EQ(plan.out_height, 1080);

	/* One dimension zero follows the source, as the vendor library does. */
	CHECK_EQ(kvmv_plan_output(1920, 1080, 1280, 0, &plan), 0);
	CHECK_EQ(plan.out_width, 1920);

	CHECK(kvmv_plan_output(0, 0, 0, 0, &plan) != 0);
	CHECK(kvmv_plan_output(1920, 1080, 8, 8, &plan) != 0);
}

static void test_timings(void)
{
	struct v4l2_dv_timings t;

	memset(&t, 0, sizeof(t));
	t.type = V4L2_DV_BT_656_1120;
	t.bt.width = 1920;
	t.bt.height = 1080;
	CHECK_EQ(kvmv_classify_timings(0, 0, &t, 1920, 1080), KVMV_SIGNAL_OK);
	CHECK_EQ(kvmv_classify_timings(0, 0, &t, 0, 0), KVMV_SIGNAL_OK);
	t.bt.width = 1280;
	t.bt.height = 720;
	CHECK_EQ(kvmv_classify_timings(0, 0, &t, 1920, 1080), KVMV_SIGNAL_UNSUPPORTED);
	CHECK_EQ(kvmv_classify_timings(0, 0, &t, 0, 0), KVMV_SIGNAL_OK);
	t.bt.width = 1920;
	t.bt.height = 1080;
	t.bt.interlaced = 1;
	CHECK_EQ(kvmv_classify_timings(0, 0, &t, 1920, 1080), KVMV_SIGNAL_UNSUPPORTED);
	t.bt.interlaced = 0;
	t.bt.width = 0;
	CHECK_EQ(kvmv_classify_timings(0, 0, &t, 1920, 1080), KVMV_SIGNAL_NONE);

	CHECK_EQ(kvmv_classify_timings(-1, ENOLINK, &t, 0, 0), KVMV_SIGNAL_NONE);
	CHECK_EQ(kvmv_classify_timings(-1, ENOLCK, &t, 0, 0), KVMV_SIGNAL_NONE);
	CHECK_EQ(kvmv_classify_timings(-1, ERANGE, &t, 0, 0), KVMV_SIGNAL_OUT_OF_RANGE);
	CHECK_EQ(kvmv_classify_timings(-1, ENOTTY, &t, 0, 0), KVMV_SIGNAL_UNKNOWN);

	CHECK_EQ(kvmv_signal_result(KVMV_SIGNAL_OK), 0);
	CHECK_EQ(kvmv_signal_result(KVMV_SIGNAL_UNKNOWN), 0);
	CHECK_EQ(kvmv_signal_result(KVMV_SIGNAL_NONE), IMG_NOT_EXIST);
	CHECK_EQ(kvmv_signal_result(KVMV_SIGNAL_UNSUPPORTED), -6);
	CHECK_EQ(kvmv_signal_result(KVMV_SIGNAL_OUT_OF_RANGE), -7);

	CHECK(kvmv_signal_present(KVMV_SIGNAL_UNSUPPORTED));
	CHECK(!kvmv_signal_present(KVMV_SIGNAL_NONE));
	CHECK(!kvmv_signal_present(KVMV_SIGNAL_UNKNOWN));
}

static void test_roles(void)
{
	const uint32_t cap = V4L2_CAP_VIDEO_CAPTURE | V4L2_CAP_STREAMING;
	const uint32_t m2m = V4L2_CAP_VIDEO_M2M | V4L2_CAP_STREAMING;

	CHECK_EQ(kvmv_match_role("sg2002-capture", cap, 0, 0, 0), KVMV_ROLE_CAPTURE);
	CHECK_EQ(kvmv_match_role("sg2002-vpss", m2m, 0, 0, 1), KVMV_ROLE_SCALER);
	CHECK_EQ(kvmv_match_role("coda", m2m, 1, 0, 1), KVMV_ROLE_ENCODER);
	CHECK_EQ(kvmv_match_role("sg2002-jpeg", m2m, 0, 0, 1), KVMV_ROLE_JPEG);
	CHECK_EQ(kvmv_match_role("sg2002-jpeg", cap, 0, 0, 0), KVMV_ROLE_NONE);
	/* The Coda's decoder and JPEG nodes are not the encoder. */
	CHECK_EQ(kvmv_match_role("coda", m2m, 0, 0, 1), KVMV_ROLE_NONE);
	CHECK_EQ(kvmv_match_role("coda", m2m, 1, 0, 0), KVMV_ROLE_NONE);
	CHECK_EQ(kvmv_match_role("sg2002-capture", V4L2_CAP_VIDEO_CAPTURE, 0, 0, 0),
		 KVMV_ROLE_NONE);
	CHECK_EQ(kvmv_match_role("uvcvideo", cap, 0, 0, 0), KVMV_ROLE_NONE);
	/* The WAVE420L: NV12 in, H.265 out. A node offering both goes to H.264. */
	CHECK_EQ(kvmv_match_role("wave420l", m2m, 0, 1, 1), KVMV_ROLE_ENCODER_HEVC);
	CHECK_EQ(kvmv_match_role("wave420l", m2m, 0, 1, 0), KVMV_ROLE_NONE);
	CHECK_EQ(kvmv_match_role("both", m2m, 1, 1, 1), KVMV_ROLE_ENCODER);
	CHECK_EQ(kvmv_match_role(NULL, cap, 0, 0, 0), KVMV_ROLE_NONE);

	CHECK(kvmv_subdev_name_matches("lt6911uxe 4-002b\n"));
	CHECK(!kvmv_subdev_name_matches("ov5647 2-0036"));
	CHECK(!kvmv_subdev_name_matches(NULL));
}

static void test_clamps_and_rate(void)
{
	struct kvmv_rate rate;
	unsigned int kbps = 0, fps_x10 = 0;
	int i, closed = 0;

	CHECK_EQ(kvmv_kbps_to_bps(0), 500000);
	CHECK_EQ(kvmv_kbps_to_bps(4000), 4000000);
	CHECK_EQ(kvmv_kbps_to_bps(65535), 10000000);
	CHECK_EQ(kvmv_clamp(0, KVMV_GOP_MIN, KVMV_GOP_MAX), 1);
	CHECK_EQ(kvmv_clamp(255, KVMV_FPS_MIN, KVMV_FPS_MAX), 60);

	/* Default bitrates by size. */
	CHECK_EQ(kvmv_default_kbps(1920, 1080), 3000);
	CHECK_EQ(kvmv_default_kbps(0, 0), 3000);
	CHECK_EQ(kvmv_default_kbps(1280, 720), 2200);
	CHECK_EQ(kvmv_default_kbps(640, 480), 1700);
	CHECK_EQ(kvmv_default_kbps(3840, 2160), 7500);
	CHECK_EQ(kvmv_default_kbps(65535, 65535), 10000);

	/* The QP range: defaults, an override, and overrides that do not parse. */
	{
		struct kvmv_qp_range r;

		kvmv_h264_qp_range(NULL, &r);
		CHECK_EQ(r.min_qp, KVMV_H264_MIN_QP);
		CHECK_EQ(r.max_qp, KVMV_H264_MAX_QP);
		kvmv_h264_qp_range("20:36", &r);
		CHECK_EQ(r.min_qp, 20);
		CHECK_EQ(r.max_qp, 36);
		kvmv_h264_qp_range("0:51", &r);
		CHECK_EQ(r.min_qp, 0);
		CHECK_EQ(r.max_qp, 51);
		kvmv_h264_qp_range("40:30", &r);
		CHECK_EQ(r.min_qp, KVMV_H264_MIN_QP);
		kvmv_h264_qp_range("10:52", &r);
		CHECK_EQ(r.max_qp, KVMV_H264_MAX_QP);
		kvmv_h264_qp_range("10:40x", &r);
		CHECK_EQ(r.min_qp, KVMV_H264_MIN_QP);
		kvmv_h264_qp_range("", &r);
		CHECK_EQ(r.max_qp, KVMV_H264_MAX_QP);
	}

	/* 30 frames a second of 12500 bytes for a second is 3000 kbit/s. */
	kvmv_rate_reset(&rate);
	for (i = 0; i <= 30; i++)
		closed += kvmv_rate_add(&rate, 1000 + (uint64_t)i * 1000 / 30, 12500,
					1000, &kbps, &fps_x10);
	CHECK_EQ(closed, 1);
	CHECK(kbps >= 3000 && kbps <= 3200);
	CHECK(fps_x10 >= 300 && fps_x10 <= 320);
}

static void test_encoder_fps(void)
{
	/* Nothing measured: the asked rate, capped by the capture rate. */
	CHECK_EQ(kvmv_encoder_fps(30, 30, 0, 0), 30);
	CHECK_EQ(kvmv_encoder_fps(30, 20, 0, 0), 20);
	CHECK_EQ(kvmv_encoder_fps(30, 60, 0, 0), 30);
	CHECK_EQ(kvmv_encoder_fps(30, 0, 0, 0), 30);
	/* Delivering half: the encoder is told the delivered rate. */
	CHECK_EQ(kvmv_encoder_fps(30, 30, 15, 30), 15);
	CHECK_EQ(kvmv_encoder_fps(30, 30, 15, 0), 15);
	/* Within 10% of the target counts as the target. */
	CHECK_EQ(kvmv_encoder_fps(30, 30, 28, 15), 30);
	CHECK_EQ(kvmv_encoder_fps(30, 30, 29, 30), 30);
	/* Small wobble while tracking costs no ioctl; a real move does. */
	CHECK_EQ(kvmv_encoder_fps(30, 30, 16, 15), 15);
	CHECK_EQ(kvmv_encoder_fps(30, 30, 20, 15), 20);
	/* Never faster than asked, never below 1. */
	CHECK_EQ(kvmv_encoder_fps(30, 30, 45, 30), 30);
	CHECK_EQ(kvmv_encoder_fps(0, 0, 0, 0), 1);
}

static void test_jpeg(void)
{
	enum { W = 96, H = 64, STRIDE = 128 };
	static uint8_t nv12[STRIDE * H * 3 / 2];
	uint8_t cb[(W / 2) * (H / 2)], cr[(W / 2) * (H / 2)];
	struct kvmv_nv12 image;
	struct kvmv_jpeg jpeg;
	const uint8_t *data;
	size_t size;
	char error[128];
	tjhandle tj;
	int x, y, ok;

	/* Flat colour with garbage in the stride padding, which must not
	 * reach the picture. */
	memset(nv12, 0xff, sizeof(nv12));
	for (y = 0; y < H; y++)
		memset(nv12 + y * STRIDE, 100, W);
	for (y = 0; y < H / 2; y++)
		for (x = 0; x < W / 2; x++) {
			nv12[STRIDE * H + y * STRIDE + 2 * x] = 80;
			nv12[STRIDE * H + y * STRIDE + 2 * x + 1] = 200;
		}

	kvmv_nv12_split_chroma(nv12 + STRIDE * H, STRIDE, W, H, cb, cr);
	ok = 1;
	for (x = 0; x < (W / 2) * (H / 2); x++)
		if (cb[x] != 80 || cr[x] != 200)
			ok = 0;
	CHECK(ok);

	memset(&image, 0, sizeof(image));
	image.y = nv12;
	image.uv = nv12 + STRIDE * H;
	image.width = W;
	image.height = H;
	image.stride = STRIDE;
	kvmv_jpeg_init(&jpeg);
	CHECK_EQ(kvmv_jpeg_compress(&jpeg, 80, &data, &size, error, sizeof(error)), -1);
	CHECK_EQ(kvmv_jpeg_load(&jpeg, &image), 0);
	CHECK_EQ(kvmv_jpeg_compress(&jpeg, 80, &data, &size, error, sizeof(error)), 0);
	CHECK(size > 100);
	CHECK(data[0] == 0xff && data[1] == 0xd8);
	CHECK(data[size - 2] == 0xff && data[size - 1] == 0xd9);

	/* It decodes, at the size given, 4:2:0, to the colour given. */
	tj = tj3Init(TJINIT_DECOMPRESS);
	CHECK(tj != NULL);
	if (tj != NULL) {
		uint8_t out[W * H + 2 * (W / 2) * (H / 2)];

		CHECK_EQ(tj3DecompressHeader(tj, data, size), 0);
		CHECK_EQ(tj3Get(tj, TJPARAM_JPEGWIDTH), W);
		CHECK_EQ(tj3Get(tj, TJPARAM_JPEGHEIGHT), H);
		CHECK_EQ(tj3Get(tj, TJPARAM_SUBSAMP), TJSAMP_420);
		CHECK_EQ(tj3DecompressToYUV8(tj, data, size, out, 1), 0);
		CHECK(abs(out[W * H / 2 + W / 2] - 100) <= 2);
		CHECK(abs(out[W * H + 10] - 80) <= 2);
		CHECK(abs(out[W * H + (W / 2) * (H / 2) + 10] - 200) <= 2);
		tj3Destroy(tj);
	}

	/* A second picture reuses the buffers; an odd size is refused. */
	CHECK_EQ(kvmv_jpeg_compress(&jpeg, 0, &data, &size, error, sizeof(error)), 0);
	image.width = W - 1;
	CHECK_EQ(kvmv_jpeg_load(&jpeg, &image), -1);
	kvmv_jpeg_free(&jpeg);
}

static void test_copy_from_device(void)
{
	uint8_t src[300], dst[320];
	size_t off, len;
	int ok = 1;

	for (off = 0; off < sizeof(src); off++)
		src[off] = (uint8_t)(off * 7 + 3);
	/* Every alignment of both ends, every short length and a long one. */
	for (off = 0; off < 9; off++) {
		for (len = 0; len < 40; len++) {
			memset(dst, 0xee, sizeof(dst));
			kvmv_copy_from_device(dst + off, src + (off % 3), len);
			if (memcmp(dst + off, src + (off % 3), len) != 0 ||
			    (off && dst[off - 1] != 0xee) || dst[off + len] != 0xee)
				ok = 0;
		}
		memset(dst, 0xee, sizeof(dst));
		kvmv_copy_from_device(dst + off, src, 290);
		if (memcmp(dst + off, src, 290) != 0 || dst[off + 290] != 0xee)
			ok = 0;
	}
	CHECK(ok);
}

static void test_slots(void)
{
	struct kvmv_slots slots;
	struct kvmv_slot *taken[KVMV_SLOT_COUNT];
	struct kvmv_slot *slot;
	int i;

	memset(&slots, 0, sizeof(slots));
	for (i = 0; i < KVMV_SLOT_COUNT; i++) {
		taken[i] = kvmv_slot_claim(&slots, 1000);
		CHECK(taken[i] != NULL);
		taken[i]->type = (uint8_t)(i % 2 ? IMG_H264_TYPE_PF : IMG_H264_TYPE_IF);
	}
	/* All four out: the fifth claim is IMG_BUFFER_FULL's case. */
	CHECK(kvmv_slot_claim(&slots, 1000) == NULL);
	CHECK_EQ(kvmv_slot_release(&slots, taken[1]->data), IMG_H264_TYPE_PF);
	/* Releasing twice finds nothing the second time. */
	CHECK_EQ(kvmv_slot_release(&slots, taken[1]->data), -1);
	CHECK_EQ(kvmv_slot_release(&slots, NULL), -1);
	/* The freed slot grows to fit and is reused. */
	slot = kvmv_slot_claim(&slots, 5000);
	CHECK(slot == taken[1]);
	CHECK(slot->capacity >= 5000);
	memset(slot->data, 0xab, 5000);
	kvmv_slots_free_all(&slots);
	CHECK(slots.slot[0].data == NULL);
	CHECK(kvmv_slot_claim(&slots, 0) == NULL);
	kvmv_slots_free_all(&slots);
}

/* ---- negotiation against mock drivers ------------------------------ */

#define FD_CAPTURE 10
#define FD_VPSS 11
#define FD_CODA 12

struct mock {
	uint32_t capture_fourcc;
	unsigned int vpss_align; /* line stride alignment of the VPSS output */
	int vpss_has_crop;
	int coda_rejects_gop;
	int coda_bitrate_divisor; /* the encoder keeps bitrate / this */
	int coda_lacks_qp; /* a kernel without patch 0910 */
	int32_t min_qp, max_qp, vbv_delay;
	struct v4l2_pix_format vpss_out, vpss_cap, coda_out, coda_cap;
	struct v4l2_rect vpss_crop, coda_crop;
	int32_t bitrate, gop;
	unsigned int fps;
	unsigned int s_fmt_calls;
};

static struct mock mock;

static unsigned int align_up(unsigned int v, unsigned int a)
{
	return (v + a - 1) / a * a;
}

static int mock_fmt(int fd, struct v4l2_format *f, int set)
{
	struct v4l2_pix_format *pix = &f->fmt.pix;
	int output = f->type == V4L2_BUF_TYPE_VIDEO_OUTPUT;

	if (fd == FD_CAPTURE) {
		if (set)
			return errno = EINVAL, -1;
		memset(pix, 0, sizeof(*pix));
		pix->width = 1920;
		pix->height = 1080;
		pix->pixelformat = mock.capture_fourcc;
		pix->bytesperline = 3840;
		pix->sizeimage = 3840 * 1080;
		pix->colorspace = V4L2_COLORSPACE_REC709;
		return 0;
	}
	if (fd == FD_VPSS) {
		struct v4l2_pix_format *slot = output ? &mock.vpss_out : &mock.vpss_cap;

		if (set) {
			mock.s_fmt_calls++;
			if (output) {
				pix->bytesperline = pix->width * 2;
				pix->sizeimage = pix->bytesperline * pix->height;
			} else {
				if (pix->pixelformat != V4L2_PIX_FMT_NV12 &&
				    pix->pixelformat != V4L2_PIX_FMT_NV21)
					pix->pixelformat = V4L2_PIX_FMT_NV12;
				pix->bytesperline = align_up(pix->width, mock.vpss_align);
				pix->sizeimage = pix->bytesperline * pix->height * 3 / 2;
			}
			*slot = *pix;
			return 0;
		}
		*pix = *slot;
		return 0;
	}
	if (fd == FD_CODA) {
		struct v4l2_pix_format *slot = output ? &mock.coda_out : &mock.coda_cap;

		if (set) {
			mock.s_fmt_calls++;
			if (output) {
				pix->pixelformat = V4L2_PIX_FMT_NV12;
				pix->width = align_up(pix->width, 16);
				pix->height = align_up(pix->height, 16);
				pix->bytesperline = pix->width;
				pix->sizeimage = pix->width * pix->height * 3 / 2;
			} else {
				pix->pixelformat = V4L2_PIX_FMT_H264;
				pix->bytesperline = 0;
				if (pix->sizeimage < 256 * 1024)
					pix->sizeimage = 256 * 1024;
			}
			*slot = *pix;
			return 0;
		}
		*pix = *slot;
		return 0;
	}
	return errno = EBADF, -1;
}

static int mock_ctrls(int fd, struct v4l2_ext_controls *list, int set)
{
	unsigned int i;

	if (fd != FD_CODA)
		return errno = ENOTTY, -1;
	for (i = 0; i < list->count; i++) {
		struct v4l2_ext_control *c = &list->controls[i];

		switch (c->id) {
		case V4L2_CID_MPEG_VIDEO_BITRATE:
			if (set)
				mock.bitrate = c->value / mock.coda_bitrate_divisor;
			else
				c->value = mock.bitrate;
			break;
		case V4L2_CID_MPEG_VIDEO_GOP_SIZE:
			if (mock.coda_rejects_gop)
				return errno = EINVAL, -1;
			if (set)
				mock.gop = c->value;
			else
				c->value = mock.gop;
			break;
		case V4L2_CID_MPEG_VIDEO_H264_MIN_QP:
		case V4L2_CID_MPEG_VIDEO_H264_MAX_QP: {
			int32_t *qp = c->id == V4L2_CID_MPEG_VIDEO_H264_MIN_QP ?
				      &mock.min_qp : &mock.max_qp;

			if (mock.coda_lacks_qp)
				return errno = EINVAL, -1;
			if (set)
				*qp = c->value;
			else
				c->value = *qp;
			break;
		}
		case V4L2_CID_MPEG_VIDEO_VBV_DELAY:
			if (set)
				mock.vbv_delay = c->value;
			else
				c->value = mock.vbv_delay;
			break;
		case V4L2_CID_MPEG_VIDEO_FRAME_RC_ENABLE:
		case V4L2_CID_MPEG_VIDEO_FORCE_KEY_FRAME:
			break;
		default:
			/* The Coda has no bitrate mode or header mode control. */
			return errno = EINVAL, -1;
		}
	}
	return 0;
}

static int mock_ioctl(int fd, unsigned long request, void *arg)
{
	switch (request) {
	case VIDIOC_G_FMT:
		return mock_fmt(fd, arg, 0);
	case VIDIOC_S_FMT:
		return mock_fmt(fd, arg, 1);
	case VIDIOC_S_SELECTION: {
		struct v4l2_selection *sel = arg;

		if (fd == FD_CODA && sel->type == V4L2_BUF_TYPE_VIDEO_OUTPUT) {
			if (sel->r.width > mock.coda_out.width ||
			    sel->r.height > mock.coda_out.height)
				return errno = EINVAL, -1;
			mock.coda_crop = sel->r;
			return 0;
		}
		if (fd == FD_VPSS && sel->type == V4L2_BUF_TYPE_VIDEO_CAPTURE) {
			if (!mock.vpss_has_crop)
				return errno = EINVAL, -1;
			mock.vpss_crop = sel->r;
			return 0;
		}
		return errno = EINVAL, -1;
	}
	case VIDIOC_S_PARM: {
		struct v4l2_streamparm *parm = arg;

		if (fd != FD_CODA || parm->type != V4L2_BUF_TYPE_VIDEO_OUTPUT)
			return errno = EINVAL, -1;
		mock.fps = parm->parm.output.timeperframe.denominator /
			   parm->parm.output.timeperframe.numerator;
		return 0;
	}
	case VIDIOC_G_PARM: {
		struct v4l2_streamparm *parm = arg;

		if (fd != FD_CODA || parm->type != V4L2_BUF_TYPE_VIDEO_OUTPUT)
			return errno = EINVAL, -1;
		parm->parm.output.timeperframe.numerator = 1;
		parm->parm.output.timeperframe.denominator = mock.fps;
		return 0;
	}
	case VIDIOC_S_EXT_CTRLS:
		return mock_ctrls(fd, arg, 1);
	case VIDIOC_G_EXT_CTRLS:
		return mock_ctrls(fd, arg, 0);
	case VIDIOC_QUERYCTRL: {
		struct v4l2_queryctrl *query = arg;

		/* The Coda's GOP control stops at 99. */
		if (fd != FD_CODA || query->id != V4L2_CID_MPEG_VIDEO_GOP_SIZE)
			return errno = EINVAL, -1;
		query->minimum = 0;
		query->maximum = 99;
		return 0;
	}
	default:
		return errno = ENOTTY, -1;
	}
}

static void mock_reset(void)
{
	memset(&mock, 0, sizeof(mock));
	mock.capture_fourcc = V4L2_PIX_FMT_UYVY;
	mock.vpss_align = 16;
	mock.vpss_has_crop = 1;
	mock.coda_bitrate_divisor = 1;
	mock.gop = 16;
	mock.fps = 30;
}

static int negotiate(struct kvmv_pipe *p, unsigned int w, unsigned int h,
		     uint32_t bps, unsigned int gop, unsigned int fps)
{
	struct kvmv_pipe_cfg cfg;

	memset(&cfg, 0, sizeof(cfg));
	cfg.req_width = w;
	cfg.req_height = h;
	cfg.bitrate_bps = bps;
	cfg.gop = gop;
	cfg.fps = fps;
	kvmv_pipe_init(p);
	p->cap_fd = FD_CAPTURE;
	p->vpss_fd = FD_VPSS;
	p->enc_fd = FD_CODA;
	return kvmv_pipe_negotiate(p, &cfg);
}

static int negotiate_qp(struct kvmv_pipe *p, int min_qp, int max_qp,
			unsigned int vbv_delay_ms)
{
	struct kvmv_pipe_cfg cfg;

	memset(&cfg, 0, sizeof(cfg));
	cfg.bitrate_bps = 4000000;
	cfg.gop = 30;
	cfg.fps = 30;
	cfg.min_qp = min_qp;
	cfg.max_qp = max_qp;
	cfg.vbv_delay_ms = vbv_delay_ms;
	kvmv_pipe_init(p);
	p->cap_fd = FD_CAPTURE;
	p->vpss_fd = FD_VPSS;
	p->enc_fd = FD_CODA;
	return kvmv_pipe_negotiate(p, &cfg);
}

static void test_negotiate(void)
{
	struct kvmv_pipe p;

	kvmv_ioctl_hook = mock_ioctl;

	/* 1080p through: the encoder pads to 1088 and crops back. */
	mock_reset();
	CHECK_EQ(negotiate(&p, 0, 0, 4000000, 30, 30), 0);
	CHECK_EQ(p.plan.out_width, 1920);
	CHECK_EQ(p.plan.out_height, 1080);
	CHECK_EQ(p.enc_out_fmt.height, 1088);
	CHECK_EQ(mock.coda_crop.height, 1080);
	CHECK_EQ(p.vpss_out_fmt.bytesperline, p.enc_out_fmt.bytesperline);
	CHECK_EQ(p.vpss_out_fmt.height, 1088);
	CHECK_EQ(mock.vpss_crop.width, 1920);
	CHECK_EQ(mock.vpss_crop.height, 1080);
	CHECK_EQ(p.vpss_in_fmt.pixelformat, V4L2_PIX_FMT_UYVY);
	CHECK_EQ(p.enc_cap_fmt.pixelformat, V4L2_PIX_FMT_H264);
	CHECK(p.enc_cap_fmt.sizeimage >= 1024 * 1024);
	/* Rate control as set, and read back. */
	CHECK_EQ(mock.fps, 30);
	CHECK_EQ(p.applied.bitrate_bps, 4000000);
	CHECK_EQ(p.applied.gop, 30);
	CHECK_EQ(p.applied.fps_numerator, 30);
	CHECK_EQ(p.applied.fps_denominator, 1);

	/* Downscale to 720p. */
	mock_reset();
	CHECK_EQ(negotiate(&p, 1280, 720, 2000000, 60, 60), 0);
	CHECK_EQ(p.enc_out_fmt.width, 1280);
	CHECK_EQ(p.enc_out_fmt.height, 720);
	CHECK_EQ(p.vpss_out_fmt.width, 1280);
	CHECK_EQ(p.vpss_in_fmt.width, 1920);
	CHECK_EQ(mock.fps, 60);

	/* The QP range and initial delay go to the encoder, and are read back. */
	mock_reset();
	CHECK_EQ(negotiate_qp(&p, 18, 42, 1000), 0);
	CHECK_EQ(mock.min_qp, 18);
	CHECK_EQ(mock.max_qp, 42);
	CHECK_EQ(mock.vbv_delay, 1000);
	CHECK_EQ(p.applied.min_qp, 18);
	CHECK_EQ(p.applied.max_qp, 42);

	/* Without the QP controls the stream still comes up. */
	mock_reset();
	mock.coda_lacks_qp = 1;
	CHECK_EQ(negotiate_qp(&p, 18, 42, 1000), 0);
	CHECK_EQ(p.applied.min_qp, -1);
	CHECK_EQ(p.applied.max_qp, -1);
	CHECK_EQ(p.applied.bitrate_bps, 4000000);

	/* 0 leaves the encoder's own range and delay alone. */
	mock_reset();
	mock.min_qp = 12;
	mock.max_qp = 51;
	CHECK_EQ(negotiate_qp(&p, 0, 0, 0), 0);
	CHECK_EQ(mock.min_qp, 12);
	CHECK_EQ(mock.max_qp, 51);
	CHECK_EQ(mock.vbv_delay, 0);

	/* The ABI allows a GOP of 100; the Coda stops at 99. */
	mock_reset();
	CHECK_EQ(negotiate(&p, 0, 0, 4000000, 100, 30), 0);
	CHECK_EQ(mock.gop, 99);

	/* An encoder that keeps a different bitrate is reported as such. */
	mock_reset();
	mock.coda_bitrate_divisor = 2;
	CHECK_EQ(negotiate(&p, 0, 0, 4000000, 30, 30), 0);
	CHECK_EQ(p.applied.bitrate_bps, 2000000);

	/* A refused control costs only that control. */
	mock_reset();
	mock.coda_rejects_gop = 1;
	CHECK_EQ(negotiate(&p, 0, 0, 4000000, 30, 30), 0);
	CHECK_EQ(p.applied.gop, -1);
	CHECK_EQ(p.applied.bitrate_bps, 4000000);

	/* Strides that disagree would scramble every line: refuse. */
	mock_reset();
	mock.vpss_align = 64;
	CHECK(negotiate(&p, 1366, 768, 4000000, 30, 30) != 0);
	CHECK(strstr(p.error, "does not match") != NULL);

	/* No scaler crop: fine without padding, refused with it. */
	mock_reset();
	mock.vpss_has_crop = 0;
	CHECK_EQ(negotiate(&p, 1280, 720, 4000000, 30, 30), 0);
	CHECK(negotiate(&p, 0, 0, 4000000, 30, 30) != 0);
	CHECK(strstr(p.error, "scaler CAPTURE crop") != NULL);

	/* A capture format the scaler cannot read. */
	mock_reset();
	mock.capture_fourcc = V4L2_PIX_FMT_SRGGB12P;
	CHECK(negotiate(&p, 0, 0, 4000000, 30, 30) != 0);
	CHECK_EQ(mock.s_fmt_calls, 0);
}

/*
 * The JPEG unit (kvmv_hwjpeg) against a mock sg2002-jpeg. The node is a
 * memfd, so the library's mmap of the OUTPUT and CAPTURE buffers works: the
 * OUTPUT buffer sits at offset 0, the CAPTURE buffer at JM_CAP_OFFSET. The
 * mock "encodes" by writing SOI, the quality, the first source byte (from its
 * OUTPUT buffer in copy mode; 0xd1 for an imported dma-buf) and EOI.
 */
#define JM_CAP_OFFSET 65536
#define JM_SIZE (2 * JM_CAP_OFFSET)
#define JM_IMPORTED_BYTE 0xd1

static struct {
	int fd;
	unsigned int width, height, stride;
	uint32_t out_size;
	int quality, ctrl_calls, s_fmt_calls;
	enum v4l2_memory out_memory;
	int out_queued, cap_queued, done, out_done;
	int import_fd;
	uint32_t import_length, out_bytesused;
	int fail_next;
	int cache_hints;
	int streaming;
} jm;

static int jm_ioctl(int fd, unsigned long request, void *arg)
{
	if (fd != jm.fd)
		return errno = EBADF, -1;
	switch (request) {
	case VIDIOC_S_FMT: {
		struct v4l2_pix_format *pix = &((struct v4l2_format *)arg)->fmt.pix;

		jm.s_fmt_calls++;
		if (((struct v4l2_format *)arg)->type == V4L2_BUF_TYPE_VIDEO_CAPTURE) {
			pix->pixelformat = V4L2_PIX_FMT_JPEG;
			pix->sizeimage = 4096;
			return 0;
		}
		/* As sg2002-jpeg: width to 16, stride to 16 and at least the
		 * width, chroma read to the next 16 lines. */
		pix->pixelformat = V4L2_PIX_FMT_NV12;
		pix->width = align_up(pix->width, 16);
		if (pix->bytesperline < pix->width)
			pix->bytesperline = pix->width;
		pix->bytesperline = align_up(pix->bytesperline, 16);
		pix->sizeimage = pix->bytesperline * pix->height +
				 pix->bytesperline * align_up(pix->height, 16) / 2;
		jm.width = pix->width;
		jm.height = pix->height;
		jm.stride = pix->bytesperline;
		jm.out_size = pix->sizeimage;
		return 0;
	}
	case VIDIOC_REQBUFS: {
		struct v4l2_requestbuffers *req = arg;

		if (req->type == V4L2_BUF_TYPE_VIDEO_OUTPUT)
			jm.out_memory = req->memory;
		/* The flags byte after capabilities: kept only with cache hints. */
		if (req->type == V4L2_BUF_TYPE_VIDEO_OUTPUT || !jm.cache_hints)
			((uint8_t *)req)[16] = 0;
		if (req->count > 1)
			req->count = 1;
		return 0;
	}
	case VIDIOC_QUERYBUF: {
		struct v4l2_buffer *buf = arg;

		if (buf->type == V4L2_BUF_TYPE_VIDEO_OUTPUT) {
			buf->m.offset = 0;
			buf->length = jm.out_size;
		} else {
			buf->m.offset = JM_CAP_OFFSET;
			buf->length = 4096;
		}
		return 0;
	}
	case VIDIOC_STREAMON:
		jm.streaming++;
		return 0;
	case VIDIOC_STREAMOFF:
		jm.streaming--;
		return 0;
	case VIDIOC_S_EXT_CTRLS: {
		struct v4l2_ext_controls *list = arg;

		if (list->count != 1 ||
		    list->controls[0].id != V4L2_CID_JPEG_COMPRESSION_QUALITY)
			return errno = EINVAL, -1;
		jm.quality = list->controls[0].value;
		jm.ctrl_calls++;
		return 0;
	}
	case VIDIOC_QBUF: {
		struct v4l2_buffer *buf = arg;

		if (buf->type == V4L2_BUF_TYPE_VIDEO_CAPTURE) {
			jm.cap_queued = 1;
			return 0;
		}
		if (buf->memory != jm.out_memory)
			return errno = EINVAL, -1;
		if (buf->memory == V4L2_MEMORY_DMABUF) {
			if (buf->length && buf->length < jm.out_size)
				return errno = EINVAL, -1;
			jm.import_fd = buf->m.fd;
			jm.import_length = buf->length;
		}
		jm.out_bytesused = buf->bytesused;
		jm.out_queued = 1;
		return 0;
	}
	case VIDIOC_DQBUF: {
		struct v4l2_buffer *buf = arg;

		if (buf->type == V4L2_BUF_TYPE_VIDEO_CAPTURE) {
			uint8_t pic[5] = { 0xff, 0xd8, 0, 0, 0xff };
			uint8_t eoi = 0xd9;

			if (!jm.cap_queued || !jm.out_queued)
				return errno = EAGAIN, -1;
			pic[2] = (uint8_t)jm.quality;
			if (jm.out_memory == V4L2_MEMORY_MMAP) {
				if (pread(jm.fd, &pic[3], 1, 0) != 1)
					return errno = EIO, -1;
			} else {
				pic[3] = JM_IMPORTED_BYTE;
			}
			if (pwrite(jm.fd, pic, 5, JM_CAP_OFFSET) != 5 ||
			    pwrite(jm.fd, &eoi, 1, JM_CAP_OFFSET + 5) != 1)
				return errno = EIO, -1;
			buf->index = 0;
			buf->bytesused = jm.fail_next ? 0 : 6;
			buf->flags = jm.fail_next ? V4L2_BUF_FLAG_ERROR : 0;
			jm.fail_next = 0;
			jm.cap_queued = 0;
			jm.out_queued = 0;
			jm.out_done = 1;
			return 0;
		}
		if (!jm.out_done)
			return errno = EAGAIN, -1;
		jm.out_done = 0;
		buf->index = 0;
		return 0;
	}
	default:
		return errno = ENOTTY, -1;
	}
}

static void test_hwjpeg(void)
{
	enum { W = 32, H = 16, STRIDE = 32 };
	static uint8_t nv12[STRIDE * H * 3 / 2];
	int (*saved)(int, unsigned long, void *) = kvmv_ioctl_hook;
	struct kvmv_hwjpeg hw;
	struct kvmv_nv12 image;
	const uint8_t *data;
	size_t size, need;
	char path[64];

	CHECK(kvmv_hwjpeg_fits(1920, 1080, 1920));
	CHECK(kvmv_hwjpeg_fits(1280, 720, 1280));
	CHECK(kvmv_hwjpeg_fits(960, 540, 960));
	CHECK(!kvmv_hwjpeg_fits(1366, 768, 1376)); /* the header would say 1376 */
	CHECK(!kvmv_hwjpeg_fits(1920, 1081, 1920));
	CHECK(!kvmv_hwjpeg_fits(1920, 1080, 1928));
	/* 1080 lines: four chroma rows past a tight 4:2:0 buffer. */
	CHECK_EQ(kvmv_hwjpeg_src_size(1920, 1080), 1920 * 1080 + 1920 * 1088 / 2);
	CHECK_EQ(kvmv_hwjpeg_src_size(1280, 720), 1280 * 720 * 3 / 2);

	memset(&jm, 0, sizeof(jm));
	jm.fd = memfd_create("jpeg-mock", 0);
	CHECK(jm.fd >= 0);
	if (jm.fd < 0)
		return;
	CHECK_EQ(ftruncate(jm.fd, JM_SIZE), 0);
	kvmv_ioctl_hook = jm_ioctl;

	/* Open through /proc so the library gets a descriptor of its own on
	 * the same memfd; the mock knows it by number. */
	snprintf(path, sizeof(path), "/proc/self/fd/%d", jm.fd);
	CHECK_EQ(kvmv_hwjpeg_open(&hw, path), 0);
	jm.fd = hw.fd;

	memset(nv12, 0x42, STRIDE * H);
	memset(nv12 + STRIDE * H, 0x80, STRIDE * H / 2);
	memset(&image, 0, sizeof(image));
	image.y = nv12;
	image.uv = nv12 + STRIDE * H;
	image.width = W;
	image.height = H;
	image.stride = STRIDE;

	/* Copy mode: the picture reaches the driver's buffer, quality 0 is
	 * the default and is set once. */
	CHECK_EQ(kvmv_hwjpeg_encode(&hw, &image, -1, 0, 0, 80, &data, &size), 0);
	CHECK_EQ(size, 6);
	CHECK(data != NULL && data[0] == 0xff && data[1] == 0xd8 && data[5] == 0xd9);
	CHECK(data != NULL && data[2] == 80 && data[3] == 0x42);
	CHECK_EQ(jm.out_memory, V4L2_MEMORY_MMAP);
	CHECK_EQ(jm.out_bytesused, jm.out_size);
	CHECK_EQ(jm.ctrl_calls, 1);
	CHECK_EQ(hw.cap_cached, 0);
	CHECK_EQ(jm.streaming, 2);
	CHECK_EQ(kvmv_hwjpeg_encode(&hw, &image, -1, 0, 80, 80, &data, &size), 0);
	CHECK_EQ(jm.ctrl_calls, 1);
	CHECK_EQ(jm.s_fmt_calls, 2); /* configured once */

	/* A driver with cache hints keeps the flag: a cached buffer. */
	jm.cache_hints = 1;
	/* Import mode reconfigures, queues the dma-buf with its size, and
	 * clamps the quality. */
	need = kvmv_hwjpeg_src_size(STRIDE, H);
	CHECK_EQ(kvmv_hwjpeg_encode(&hw, &image, 77, need, 150, 80, &data, &size), 0);
	CHECK_EQ(jm.out_memory, V4L2_MEMORY_DMABUF);
	CHECK_EQ(jm.import_fd, 77);
	CHECK_EQ(jm.import_length, need);
	CHECK_EQ(hw.cap_cached, 1);
	CHECK(data != NULL && data[2] == 100 && data[3] == JM_IMPORTED_BYTE);
	CHECK_EQ(jm.s_fmt_calls, 4);
	CHECK_EQ(jm.streaming, 2);

	/* A source buffer too small for the unit's reads is refused. */
	CHECK_EQ(kvmv_hwjpeg_encode(&hw, &image, 77, need - 1, 80, 80, &data, &size), -1);
	CHECK(data == NULL);

	/* A picture the driver marks bad is an error, and the next one works. */
	jm.fail_next = 1;
	CHECK_EQ(kvmv_hwjpeg_encode(&hw, &image, 77, need, 80, 80, &data, &size), -1);
	CHECK(data == NULL && size == 0);
	CHECK_EQ(kvmv_hwjpeg_encode(&hw, &image, 77, need, 80, 80, &data, &size), 0);

	/* A width the unit would round is not taken. */
	image.width = W - 2;
	CHECK_EQ(kvmv_hwjpeg_encode(&hw, &image, 77, need, 80, 80, &data, &size), -1);

	kvmv_hwjpeg_close(&hw);
	CHECK_EQ(hw.fd, -1);
	CHECK_EQ(jm.streaming, 0);
	kvmv_ioctl_hook = saved;
}

int main(void)
{
	test_split();
	test_classify();
	test_ps_cache();
	test_sps();
	test_hevc();
	test_plan();
	test_timings();
	test_roles();
	test_clamps_and_rate();
	test_encoder_fps();
	test_copy_from_device();
	test_jpeg();
	test_hwjpeg();
	test_slots();
	test_negotiate();

	if (failures) {
		fprintf(stderr, "%d of %d checks failed\n", failures, checks);
		return 1;
	}
	printf("all %d checks passed\n", checks);
	return 0;
}
