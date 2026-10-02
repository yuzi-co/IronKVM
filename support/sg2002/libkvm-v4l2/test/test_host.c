/*
 * Host unit tests for libkvm-v4l2: Annex-B classification, parameter-set
 * handling, SPS parsing, the policy decisions, the frame slots, and format
 * negotiation against mock capture, VPSS and Coda drivers.
 *
 * Built and run by `make test`.
 */
#include <errno.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <linux/videodev2.h>

#include "kvm_vision.h"
#include "kvmv_annexb.h"
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

	CHECK_EQ(kvmv_match_role("sg2002-capture", cap, 0, 0), KVMV_ROLE_CAPTURE);
	CHECK_EQ(kvmv_match_role("sg2002-vpss", m2m, 0, 1), KVMV_ROLE_SCALER);
	CHECK_EQ(kvmv_match_role("coda", m2m, 1, 1), KVMV_ROLE_ENCODER);
	/* The Coda's decoder and JPEG nodes are not the encoder. */
	CHECK_EQ(kvmv_match_role("coda", m2m, 0, 1), KVMV_ROLE_NONE);
	CHECK_EQ(kvmv_match_role("coda", m2m, 1, 0), KVMV_ROLE_NONE);
	CHECK_EQ(kvmv_match_role("sg2002-capture", V4L2_CAP_VIDEO_CAPTURE, 0, 0),
		 KVMV_ROLE_NONE);
	CHECK_EQ(kvmv_match_role("uvcvideo", cap, 0, 0), KVMV_ROLE_NONE);
	CHECK_EQ(kvmv_match_role(NULL, cap, 0, 0), KVMV_ROLE_NONE);

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

int main(void)
{
	test_split();
	test_classify();
	test_ps_cache();
	test_sps();
	test_plan();
	test_timings();
	test_roles();
	test_clamps_and_rate();
	test_encoder_fps();
	test_copy_from_device();
	test_slots();
	test_negotiate();

	if (failures) {
		fprintf(stderr, "%d of %d checks failed\n", failures, checks);
		return 1;
	}
	printf("all %d checks passed\n", checks);
	return 0;
}
