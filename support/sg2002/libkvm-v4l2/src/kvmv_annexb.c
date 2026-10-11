#include "kvmv_annexb.h"

#include <string.h>

#include "kvm_vision.h"

/*
 * Find the next start code at or after from. A 00 00 01 preceded by a zero is
 * reported as the 4-byte form, so the zero belongs to the start code and not to
 * the end of the previous NAL. Returns len when there is none.
 */
static size_t find_start_code(const uint8_t *buf, size_t len, size_t from,
			      size_t *code_len)
{
	size_t i;

	for (i = from; i + 3 <= len; i++) {
		if (buf[i] != 0 || buf[i + 1] != 0 || buf[i + 2] != 1)
			continue;
		if (i > from && buf[i - 1] == 0) {
			*code_len = 4;
			return i - 1;
		}
		*code_len = 3;
		return i;
	}
	*code_len = 0;
	return len;
}

static uint8_t nal_type(enum kvmv_codec codec, uint8_t header)
{
	return codec == KVMV_CODEC_KIND_HEVC ? (header >> 1) & 0x3f : header & 0x1f;
}

size_t kvmv_annexb_split_codec(enum kvmv_codec codec, const uint8_t *buf,
			       size_t len, struct kvmv_nal *nals, size_t max_nals)
{
	size_t count = 0;
	size_t code_len;
	size_t pos;

	if (buf == NULL || len == 0)
		return 0;

	pos = find_start_code(buf, len, 0, &code_len);
	while (pos < len && count < max_nals) {
		size_t start = pos + code_len;
		size_t next_len;
		size_t next = find_start_code(buf, len, start, &next_len);

		if (start < next) {
			nals[count].offset = pos;
			nals[count].start = start;
			nals[count].size = next - start;
			nals[count].type = nal_type(codec, buf[start]);
			count++;
		}
		pos = next;
		code_len = next_len;
	}
	return count;
}

size_t kvmv_annexb_split(const uint8_t *buf, size_t len, struct kvmv_nal *nals,
			 size_t max_nals)
{
	return kvmv_annexb_split_codec(KVMV_CODEC_KIND_H264, buf, len, nals, max_nals);
}

static void count_hevc(struct kvmv_au_info *info, uint8_t type)
{
	if (type <= KVMV_HEVC_NAL_RASL_R)
		info->slice++;
	else if (type >= KVMV_HEVC_NAL_BLA_W_LP && type <= KVMV_HEVC_NAL_CRA)
		info->idr++;
	else if (type == KVMV_HEVC_NAL_VPS)
		info->vps++;
	else if (type == KVMV_HEVC_NAL_SPS)
		info->sps++;
	else if (type == KVMV_HEVC_NAL_PPS)
		info->pps++;
	else if (type == KVMV_HEVC_NAL_SEI_PREFIX || type == KVMV_HEVC_NAL_SEI_SUFFIX)
		info->sei++;
	else if (type == KVMV_HEVC_NAL_AUD)
		info->aud++;
	else
		info->other++;
}

void kvmv_au_inspect_codec(enum kvmv_codec codec, const uint8_t *buf, size_t len,
			   struct kvmv_au_info *info)
{
	struct kvmv_nal nals[64];
	size_t count;
	size_t i;

	memset(info, 0, sizeof(*info));
	count = kvmv_annexb_split_codec(codec, buf, len, nals,
					sizeof(nals) / sizeof(nals[0]));
	info->nal_count = (unsigned int)count;
	for (i = 0; i < count; i++) {
		if (codec == KVMV_CODEC_KIND_HEVC) {
			count_hevc(info, nals[i].type);
			continue;
		}
		switch (nals[i].type) {
		case KVMV_NAL_IDR:
			info->idr++;
			break;
		case 1:
		case 2:
		case 3:
		case 4:
			info->slice++;
			break;
		case KVMV_NAL_SPS:
			info->sps++;
			break;
		case KVMV_NAL_PPS:
			info->pps++;
			break;
		case KVMV_NAL_SEI:
			info->sei++;
			break;
		case KVMV_NAL_AUD:
			info->aud++;
			break;
		default:
			info->other++;
			break;
		}
	}
}

void kvmv_au_inspect(const uint8_t *buf, size_t len, struct kvmv_au_info *info)
{
	kvmv_au_inspect_codec(KVMV_CODEC_KIND_H264, buf, len, info);
}

int kvmv_au_type(const struct kvmv_au_info *info)
{
	if (info->idr)
		return IMG_H264_TYPE_IF;
	if (info->slice)
		return IMG_H264_TYPE_PF;
	if (info->vps || info->sps || info->pps)
		return KVMV_AU_HEADERS_ONLY;
	return KVMV_AU_NO_PICTURE;
}

void kvmv_ps_cache_reset_codec(struct kvmv_ps_cache *cache, enum kvmv_codec codec)
{
	cache->codec = codec;
	cache->vps_len = 0;
	cache->sps_len = 0;
	cache->pps_len = 0;
}

void kvmv_ps_cache_reset(struct kvmv_ps_cache *cache)
{
	kvmv_ps_cache_reset_codec(cache, KVMV_CODEC_KIND_H264);
}

static void cache_store(uint8_t *dst, size_t *dst_len, const uint8_t *nal,
			size_t size)
{
	static const uint8_t start_code[4] = { 0, 0, 0, 1 };

	/* Drop trailing zero bytes; they belong to no parameter set. */
	while (size > 1 && nal[size - 1] == 0)
		size--;
	if (size + sizeof(start_code) > KVMV_PS_MAX)
		return;
	memcpy(dst, start_code, sizeof(start_code));
	memcpy(dst + sizeof(start_code), nal, size);
	*dst_len = size + sizeof(start_code);
}

void kvmv_ps_cache_update(struct kvmv_ps_cache *cache, const uint8_t *buf,
			  size_t len)
{
	struct kvmv_nal nals[64];
	int hevc = cache->codec == KVMV_CODEC_KIND_HEVC;
	size_t count = kvmv_annexb_split_codec(cache->codec, buf, len, nals,
					       sizeof(nals) / sizeof(nals[0]));
	size_t i;

	for (i = 0; i < count; i++) {
		const uint8_t *nal = buf + nals[i].start;

		if (hevc && nals[i].type == KVMV_HEVC_NAL_VPS)
			cache_store(cache->vps, &cache->vps_len, nal, nals[i].size);
		else if (nals[i].type == (hevc ? KVMV_HEVC_NAL_SPS : KVMV_NAL_SPS))
			cache_store(cache->sps, &cache->sps_len, nal, nals[i].size);
		else if (nals[i].type == (hevc ? KVMV_HEVC_NAL_PPS : KVMV_NAL_PPS))
			cache_store(cache->pps, &cache->pps_len, nal, nals[i].size);
	}
}

size_t kvmv_ps_cache_prefix(const struct kvmv_ps_cache *cache,
			    const struct kvmv_au_info *info, uint8_t *out,
			    size_t out_size)
{
	int hevc = cache->codec == KVMV_CODEC_KIND_HEVC;
	size_t need;

	if (!info->idr || ((info->vps || !hevc) && info->sps && info->pps))
		return 0;
	if ((hevc && cache->vps_len == 0) || cache->sps_len == 0 ||
	    cache->pps_len == 0)
		return 0;
	/*
	 * All of them, even when the unit carries some: a PPS is parsed against
	 * the SPS it names, and an H.265 SPS against its VPS, so they travel
	 * together.
	 */
	need = (hevc ? cache->vps_len : 0) + cache->sps_len + cache->pps_len;
	if (need > out_size)
		return 0;
	if (hevc) {
		memcpy(out, cache->vps, cache->vps_len);
		out += cache->vps_len;
	}
	memcpy(out, cache->sps, cache->sps_len);
	memcpy(out + cache->sps_len, cache->pps, cache->pps_len);
	return need;
}

/* ---- SPS parsing --------------------------------------------------- */

struct bit_reader {
	uint8_t rbsp[KVMV_PS_MAX];
	size_t len;
	size_t bit;
	int overrun;
};

static void reader_init(struct bit_reader *r, const uint8_t *nal, size_t len)
{
	size_t zeros = 0;
	size_t i;

	r->len = 0;
	r->bit = 0;
	r->overrun = 0;
	/* Remove emulation prevention bytes: 00 00 03 becomes 00 00. */
	for (i = 0; i < len && r->len < sizeof(r->rbsp); i++) {
		if (zeros >= 2 && nal[i] == 3) {
			zeros = 0;
			continue;
		}
		r->rbsp[r->len++] = nal[i];
		zeros = nal[i] == 0 ? zeros + 1 : 0;
	}
}

static unsigned int read_bit(struct bit_reader *r)
{
	unsigned int value;

	if (r->bit >= r->len * 8) {
		r->overrun = 1;
		return 0;
	}
	value = (r->rbsp[r->bit / 8] >> (7 - r->bit % 8)) & 1;
	r->bit++;
	return value;
}

static unsigned int read_bits(struct bit_reader *r, unsigned int count)
{
	unsigned int value = 0;

	while (count--)
		value = (value << 1) | read_bit(r);
	return value;
}

static unsigned int read_ue(struct bit_reader *r)
{
	unsigned int zeros = 0;

	while (read_bit(r) == 0) {
		if (r->overrun || ++zeros > 31) {
			r->overrun = 1;
			return 0;
		}
	}
	if (zeros == 0)
		return 0;
	return ((1U << zeros) - 1) + read_bits(r, zeros);
}

static int read_se(struct bit_reader *r)
{
	unsigned int value = read_ue(r);

	if (value & 1)
		return (int)((value + 1) / 2);
	return -(int)(value / 2);
}

static void skip_scaling_list(struct bit_reader *r, unsigned int size)
{
	int last = 8;
	int next = 8;
	unsigned int i;

	for (i = 0; i < size && !r->overrun; i++) {
		if (next != 0)
			next = (last + read_se(r) + 256) % 256;
		last = next == 0 ? last : next;
	}
}

/* What an H.264 SPS says up to vui_parameters_present_flag. */
struct h264_sps_head {
	unsigned int chroma_format;
	unsigned int separate_planes;
	unsigned int max_num_ref_frames;
	unsigned int width_mbs, height_units, frame_mbs_only;
	unsigned int crop_left, crop_right, crop_top, crop_bottom;
};

/* Reads an SPS from after the NAL header up to, not including,
 * vui_parameters_present_flag. Returns 0, or -1 for one it cannot read. */
static int h264_sps_head(struct bit_reader *r, struct h264_sps_head *h)
{
	unsigned int profile;

	memset(h, 0, sizeof(*h));
	h->chroma_format = 1;
	profile = read_bits(r, 8);
	read_bits(r, 8); /* constraint flags */
	read_bits(r, 8); /* level */
	read_ue(r); /* seq_parameter_set_id */

	if (profile == 100 || profile == 110 || profile == 122 || profile == 244 ||
	    profile == 44 || profile == 83 || profile == 86 || profile == 118 ||
	    profile == 128 || profile == 138 || profile == 139 || profile == 134 ||
	    profile == 135) {
		h->chroma_format = read_ue(r);
		if (h->chroma_format == 3)
			h->separate_planes = read_bit(r);
		read_ue(r); /* bit_depth_luma_minus8 */
		read_ue(r); /* bit_depth_chroma_minus8 */
		read_bit(r); /* qpprime_y_zero_transform_bypass */
		if (read_bit(r)) {
			unsigned int lists = h->chroma_format != 3 ? 8 : 12;
			unsigned int i;

			for (i = 0; i < lists; i++)
				if (read_bit(r))
					skip_scaling_list(r, i < 6 ? 16 : 64);
		}
	}

	read_ue(r); /* log2_max_frame_num_minus4 */
	switch (read_ue(r)) { /* pic_order_cnt_type */
	case 0:
		read_ue(r);
		break;
	case 1: {
		unsigned int cycle, i;

		read_bit(r);
		read_se(r);
		read_se(r);
		cycle = read_ue(r);
		if (cycle > 255)
			return -1;
		for (i = 0; i < cycle; i++)
			read_se(r);
		break;
	}
	default:
		break;
	}
	h->max_num_ref_frames = read_ue(r);
	read_bit(r); /* gaps_in_frame_num_value_allowed */
	h->width_mbs = read_ue(r) + 1;
	h->height_units = read_ue(r) + 1;
	h->frame_mbs_only = read_bit(r);
	if (!h->frame_mbs_only)
		read_bit(r); /* mb_adaptive_frame_field */
	read_bit(r); /* direct_8x8_inference */
	if (read_bit(r)) {
		h->crop_left = read_ue(r);
		h->crop_right = read_ue(r);
		h->crop_top = read_ue(r);
		h->crop_bottom = read_ue(r);
	}
	return r->overrun ? -1 : 0;
}

int kvmv_h264_sps_size(const uint8_t *nal, size_t len, unsigned int *width,
		       unsigned int *height)
{
	struct bit_reader r;
	struct h264_sps_head h;
	unsigned int crop_x, crop_y;
	unsigned int w, h_px;

	if (nal == NULL || len < 4 || (nal[0] & 0x1f) != KVMV_NAL_SPS)
		return -1;

	reader_init(&r, nal + 1, len - 1);
	if (h264_sps_head(&r, &h))
		return -1;

	w = h.width_mbs * 16;
	h_px = (2 - h.frame_mbs_only) * h.height_units * 16;
	if (h.chroma_format == 0 || h.separate_planes) {
		crop_x = 1;
		crop_y = 2 - h.frame_mbs_only;
	} else {
		crop_x = h.chroma_format == 3 ? 1 : 2;
		crop_y = (h.chroma_format == 1 ? 2 : 1) * (2 - h.frame_mbs_only);
	}
	if (crop_x * (h.crop_left + h.crop_right) >= w ||
	    crop_y * (h.crop_top + h.crop_bottom) >= h_px)
		return -1;
	*width = w - crop_x * (h.crop_left + h.crop_right);
	*height = h_px - crop_y * (h.crop_top + h.crop_bottom);
	return 0;
}

/* ---- the low-delay VUI ----------------------------------------------- */

struct bit_writer {
	uint8_t buf[KVMV_PS_MAX];
	size_t bit;
	int overflow;
};

static void put_bit(struct bit_writer *w, unsigned int bit)
{
	if (w->bit >= sizeof(w->buf) * 8) {
		w->overflow = 1;
		return;
	}
	if (w->bit % 8 == 0)
		w->buf[w->bit / 8] = 0;
	if (bit)
		w->buf[w->bit / 8] |= (uint8_t)(0x80 >> (w->bit % 8));
	w->bit++;
}

static void put_ue(struct bit_writer *w, unsigned int value)
{
	unsigned int v = value + 1;
	int bits = 0;
	int i;

	while ((v >> bits) > 1)
		bits++;
	for (i = 0; i < bits; i++)
		put_bit(w, 0);
	for (i = bits; i >= 0; i--)
		put_bit(w, (v >> i) & 1);
}

/* video_signal_type_present_flag and what follows it, H.264 E.1.1 and H.265
 * E.2.1 alike: video_format 5 (unspecified), the range, and the colour
 * description. Without a colour, the flag alone, 0. */
static void put_video_signal(struct bit_writer *w, const struct kvmv_vui_colour *colour)
{
	int i;

	put_bit(w, colour != NULL); /* video_signal_type_present_flag */
	if (colour == NULL)
		return;
	for (i = 2; i >= 0; i--)
		put_bit(w, (5 >> i) & 1); /* video_format */
	put_bit(w, colour->full_range != 0); /* video_full_range_flag */
	put_bit(w, 1); /* colour_description_present_flag */
	for (i = 7; i >= 0; i--)
		put_bit(w, (colour->primaries >> i) & 1);
	for (i = 7; i >= 0; i--)
		put_bit(w, (colour->transfer >> i) & 1);
	for (i = 7; i >= 0; i--)
		put_bit(w, (colour->matrix >> i) & 1);
}

/* RBSP bits [from, to) of the reader, as they were. */
static void copy_bits(struct bit_writer *w, const struct bit_reader *r, size_t from,
		      size_t to)
{
	size_t bit;

	for (bit = from; bit < to; bit++)
		put_bit(w, (r->rbsp[bit / 8] >> (7 - bit % 8)) & 1);
}

/*
 * The NAL header, then the writer's RBSP (stop bit and alignment included)
 * with emulation prevention put back. Returns the length, or 0 when out is too
 * small or the writer overflowed.
 */
static size_t put_nal_out(const uint8_t *header, size_t header_len,
			  const struct bit_writer *w, uint8_t *out, size_t out_size)
{
	size_t bytes = w->bit / 8, n, i, zeros;

	if (w->overflow || w->bit % 8 || out_size < header_len)
		return 0;
	memcpy(out, header, header_len);
	n = header_len;
	zeros = 0;
	for (i = 0; i < bytes; i++) {
		if (zeros >= 2 && w->buf[i] <= 3) {
			if (n >= out_size)
				return 0;
			out[n++] = 3;
			zeros = 0;
		}
		if (n >= out_size)
			return 0;
		out[n++] = w->buf[i];
		zeros = w->buf[i] == 0 ? zeros + 1 : 0;
	}
	return n;
}

size_t kvmv_h264_sps_low_delay(const uint8_t *nal, size_t len, uint8_t *out,
			       size_t out_size, const struct kvmv_vui_colour *colour)
{
	struct bit_reader r;
	struct h264_sps_head h;
	struct bit_writer w;

	if (nal == NULL || out == NULL || len < 4 || len - 1 > KVMV_PS_MAX ||
	    (nal[0] & 0x1f) != KVMV_NAL_SPS)
		return 0;
	reader_init(&r, nal + 1, len - 1);
	if (h264_sps_head(&r, &h) || read_bit(&r) != 0 || r.overrun)
		return 0; /* unreadable, or it has a VUI already */

	/* The RBSP up to the VUI flag as it was, then the flag set. */
	w.bit = 0;
	w.overflow = 0;
	copy_bits(&w, &r, 0, r.bit - 1);
	put_bit(&w, 1); /* vui_parameters_present_flag */
	put_bit(&w, 0); /* aspect_ratio_info_present_flag */
	put_bit(&w, 0); /* overscan_info_present_flag */
	put_video_signal(&w, colour);
	put_bit(&w, 0); /* chroma_loc_info_present_flag */
	put_bit(&w, 0); /* timing_info_present_flag */
	put_bit(&w, 0); /* nal_hrd_parameters_present_flag */
	put_bit(&w, 0); /* vcl_hrd_parameters_present_flag */
	put_bit(&w, 0); /* pic_struct_present_flag */
	put_bit(&w, 1); /* bitstream_restriction_flag */
	put_bit(&w, 1); /* motion_vectors_over_pic_boundaries_flag */
	put_ue(&w, 2); /* max_bytes_per_pic_denom (the default) */
	put_ue(&w, 1); /* max_bits_per_mb_denom (the default) */
	put_ue(&w, 16); /* log2_max_mv_length_horizontal */
	put_ue(&w, 16); /* log2_max_mv_length_vertical */
	put_ue(&w, 0); /* max_num_reorder_frames */
	put_ue(&w, h.max_num_ref_frames ? h.max_num_ref_frames : 1); /* max_dec_frame_buffering */
	put_bit(&w, 1); /* rbsp_stop_one_bit */
	while (w.bit % 8)
		put_bit(&w, 0);
	return put_nal_out(nal, 1, &w, out, out_size);
}

typedef size_t (*sps_rewrite_fn)(const uint8_t *nal, size_t len, uint8_t *out,
				 size_t out_size, const struct kvmv_vui_colour *colour);

/* Rewrites every SPS of an access unit in place with fn; see kvmv_annexb.h. */
static size_t au_rewrite_sps(enum kvmv_codec codec, uint8_t *buf, size_t len,
			     size_t capacity, sps_rewrite_fn fn,
			     const struct kvmv_vui_colour *colour)
{
	struct kvmv_nal nals[16];
	unsigned int sps_type = codec == KVMV_CODEC_KIND_HEVC ? KVMV_HEVC_NAL_SPS : KVMV_NAL_SPS;
	size_t count, i;

	if (buf == NULL)
		return len;
	count = kvmv_annexb_split_codec(codec, buf, len, nals, 16);
	/* From the last, so the offsets of the earlier ones stay right. */
	for (i = count; i-- > 0;) {
		uint8_t sps[KVMV_PS_MAX + 16];
		size_t size, old;

		if (nals[i].type != sps_type)
			continue;
		old = nals[i].size;
		/* Trailing zero bytes belong to no NAL; leave them out. */
		while (old > 1 && buf[nals[i].start + old - 1] == 0)
			old--;
		size = fn(buf + nals[i].start, old, sps, sizeof(sps), colour);
		if (size == 0 || len - old + size > capacity)
			continue;
		memmove(buf + nals[i].start + size, buf + nals[i].start + old,
			len - nals[i].start - old);
		memcpy(buf + nals[i].start, sps, size);
		len = len - old + size;
	}
	return len;
}

size_t kvmv_h264_au_low_delay(uint8_t *buf, size_t len, size_t capacity,
			      const struct kvmv_vui_colour *colour)
{
	return au_rewrite_sps(KVMV_CODEC_KIND_H264, buf, len, capacity,
			      kvmv_h264_sps_low_delay, colour);
}

/*
 * profile_tier_level() of an H.265 VPS or SPS with profile_present_flag 1:
 * 88 bits of general profile, 8 of level, then the sub-layers.
 */
static void skip_profile_tier_level(struct bit_reader *r, unsigned int max_sub_layers_minus1)
{
	unsigned int profile_present[8] = { 0 }, level_present[8] = { 0 };
	unsigned int i;

	read_bits(r, 8); /* profile space, tier, profile idc */
	read_bits(r, 32); /* profile compatibility flags */
	read_bits(r, 4); /* progressive, interlaced, non-packed, frame-only */
	read_bits(r, 32); /* the 43 reserved or constraint bits ... */
	read_bits(r, 11);
	read_bits(r, 1); /* ... and the inbld or reserved bit */
	read_bits(r, 8); /* general_level_idc */
	for (i = 0; i < max_sub_layers_minus1; i++) {
		profile_present[i] = read_bit(r);
		level_present[i] = read_bit(r);
	}
	if (max_sub_layers_minus1 > 0)
		for (i = max_sub_layers_minus1; i < 8; i++)
			read_bits(r, 2);
	for (i = 0; i < max_sub_layers_minus1; i++) {
		if (profile_present[i]) {
			read_bits(r, 32);
			read_bits(r, 32);
			read_bits(r, 24);
		}
		if (level_present[i])
			read_bits(r, 8);
	}
}

int kvmv_hevc_sps_size(const uint8_t *nal, size_t len, unsigned int *width,
		       unsigned int *height)
{
	struct bit_reader r;
	unsigned int max_sub_layers_minus1, chroma_format, separate_planes = 0;
	unsigned int w, h, crop_x, crop_y;
	unsigned int left = 0, right = 0, top = 0, bottom = 0;

	if (nal == NULL || len < 4 ||
	    ((nal[0] >> 1) & 0x3f) != KVMV_HEVC_NAL_SPS)
		return -1;

	reader_init(&r, nal + 2, len - 2);
	read_bits(&r, 4); /* sps_video_parameter_set_id */
	max_sub_layers_minus1 = read_bits(&r, 3);
	read_bit(&r); /* sps_temporal_id_nesting_flag */
	skip_profile_tier_level(&r, max_sub_layers_minus1);
	read_ue(&r); /* sps_seq_parameter_set_id */
	chroma_format = read_ue(&r);
	if (chroma_format == 3)
		separate_planes = read_bit(&r);
	w = read_ue(&r);
	h = read_ue(&r);
	if (read_bit(&r)) { /* conformance_window_flag */
		left = read_ue(&r);
		right = read_ue(&r);
		top = read_ue(&r);
		bottom = read_ue(&r);
	}
	if (r.overrun || chroma_format > 3 || w == 0 || h == 0)
		return -1;

	if (chroma_format == 0 || separate_planes) {
		crop_x = 1;
		crop_y = 1;
	} else {
		crop_x = chroma_format == 3 ? 1 : 2;
		crop_y = chroma_format == 1 ? 2 : 1;
	}
	if (crop_x * (left + right) >= w || crop_y * (top + bottom) >= h)
		return -1;
	*width = w - crop_x * (left + right);
	*height = h - crop_y * (top + bottom);
	return 0;
}

/*
 * The NAL byte that holds RBSP bit 'bit' of a NAL starting at nal (header
 * included), skipping emulation prevention bytes. Returns the offset, or -1.
 */
static long rbsp_bit_to_nal_byte(const uint8_t *nal, size_t len, size_t bit)
{
	size_t want = bit / 8, rbsp = 0, zeros = 0, i;

	for (i = 0; i < len; i++) {
		if (zeros >= 2 && nal[i] == 3) {
			zeros = 0;
			continue;
		}
		if (rbsp == want)
			return (long)i;
		rbsp++;
		zeros = nal[i] == 0 ? zeros + 1 : 0;
	}
	return -1;
}

static int clear_one_vps_extension(uint8_t *nal, size_t len)
{
	struct bit_reader r;
	unsigned int max_sub_layers_minus1, i, n, layer_sets, max_layer_id;
	size_t bit;
	long at;

	/* The reader starts after the 2-byte NAL header; bits count from there. */
	reader_init(&r, nal + 2, len - 2);
	read_bits(&r, 4); /* vps_video_parameter_set_id */
	read_bits(&r, 2); /* base layer internal, available */
	read_bits(&r, 6); /* vps_max_layers_minus1 */
	max_sub_layers_minus1 = read_bits(&r, 3);
	read_bit(&r); /* vps_temporal_id_nesting_flag */
	read_bits(&r, 16); /* vps_reserved_0xffff_16bits */
	skip_profile_tier_level(&r, max_sub_layers_minus1);
	n = read_bit(&r) ? max_sub_layers_minus1 + 1 : 1; /* ordering info present */
	for (i = 0; i < n; i++) {
		read_ue(&r);
		read_ue(&r);
		read_ue(&r);
	}
	max_layer_id = read_bits(&r, 6);
	layer_sets = read_ue(&r);
	if (layer_sets > 1023)
		return 0;
	for (i = 1; i <= layer_sets; i++)
		read_bits(&r, max_layer_id + 1); /* layer_id_included_flag */
	if (read_bit(&r)) { /* vps_timing_info_present_flag */
		read_bits(&r, 32);
		read_bits(&r, 32);
		if (read_bit(&r)) /* vps_poc_proportional_to_timing_flag */
			read_ue(&r);
		if (read_ue(&r) != 0) /* vps_num_hrd_parameters: not handled */
			return 0;
	}
	if (r.overrun || !read_bit(&r)) /* vps_extension_flag */
		return 0;

	/* The flag's bit, counted in the RBSP after the NAL header. */
	bit = r.bit - 1 + 16;
	at = rbsp_bit_to_nal_byte(nal, len, bit);
	if (at < 0)
		return 0;
	/* Flag 0, the stop bit after it, zeros to the end of the byte. */
	{
		unsigned int pos = bit % 8; /* from the most significant bit */
		uint8_t keep = (uint8_t)(0xff00 >> pos);
		uint8_t stop = (uint8_t)(0x80 >> (pos + 1));

		if (pos == 7) {
			nal[at] &= keep;
			if ((size_t)at + 1 >= len)
				return 0;
			nal[at + 1] = 0x80;
			at++;
		} else {
			nal[at] = (uint8_t)((nal[at] & keep) | stop);
		}
	}
	/* What the extension took is now trailing zero bytes. */
	memset(nal + at + 1, 0, len - (size_t)at - 1);
	return 1;
}

unsigned int kvmv_hevc_clear_vps_extension(uint8_t *buf, size_t len)
{
	struct kvmv_nal nals[64];
	size_t count = kvmv_annexb_split_codec(KVMV_CODEC_KIND_HEVC, buf, len, nals,
					       sizeof(nals) / sizeof(nals[0]));
	unsigned int changed = 0;
	size_t i;

	for (i = 0; i < count; i++)
		if (nals[i].type == KVMV_HEVC_NAL_VPS && nals[i].size > 4)
			changed += clear_one_vps_extension(buf + nals[i].start, nals[i].size);
	return changed;
}

/* scaling_list_data() of an H.265 SPS or PPS (7.3.4). */
static void skip_hevc_scaling_list_data(struct bit_reader *r)
{
	unsigned int size_id, matrix_id, i, coefs;

	for (size_id = 0; size_id < 4; size_id++) {
		for (matrix_id = 0; matrix_id < 6; matrix_id += size_id == 3 ? 3 : 1) {
			if (!read_bit(r)) { /* scaling_list_pred_mode_flag */
				read_ue(r); /* scaling_list_pred_matrix_id_delta */
				continue;
			}
			coefs = 1U << (4 + (size_id << 1));
			if (coefs > 64)
				coefs = 64;
			if (size_id > 1)
				read_se(r); /* scaling_list_dc_coef_minus8 */
			for (i = 0; i < coefs && !r->overrun; i++)
				read_se(r); /* scaling_list_delta_coef */
		}
	}
}

/* The SPS's short-term reference picture sets (7.3.7). Returns -1 for sets it
 * cannot read. */
static int skip_hevc_st_ref_pic_sets(struct bit_reader *r, unsigned int count)
{
	unsigned int deltas[64];
	unsigned int idx, j, n;

	if (count > 64)
		return -1;
	for (idx = 0; idx < count && !r->overrun; idx++) {
		if (idx != 0 && read_bit(r)) { /* inter_ref_pic_set_prediction_flag */
			read_bit(r); /* delta_rps_sign */
			read_ue(r); /* abs_delta_rps_minus1 */
			n = 0;
			/* In the SPS the reference is the set before this one. */
			for (j = 0; j <= deltas[idx - 1]; j++) {
				unsigned int used = read_bit(r);

				if (used || read_bit(r)) /* use_delta_flag */
					n++;
			}
		} else {
			unsigned int negative = read_ue(r), positive = read_ue(r);

			if (negative > 16 || positive > 16)
				return -1;
			n = negative + positive;
			for (j = 0; j < n; j++) {
				read_ue(r); /* delta_poc_minus1 */
				read_bit(r); /* used_by_curr_pic_flag */
			}
		}
		if (n > 32)
			return -1;
		deltas[idx] = n;
	}
	return r->overrun ? -1 : 0;
}

size_t kvmv_hevc_sps_colour(const uint8_t *nal, size_t len, uint8_t *out,
			    size_t out_size, const struct kvmv_vui_colour *colour)
{
	struct bit_reader r;
	struct bit_writer w;
	unsigned int max_sub_layers_minus1, i, n, log2_max_poc_lsb;
	size_t vui_flag, end;

	if (nal == NULL || out == NULL || colour == NULL || len < 4 ||
	    len - 2 > KVMV_PS_MAX || ((nal[0] >> 1) & 0x3f) != KVMV_HEVC_NAL_SPS)
		return 0;

	/* 7.3.2.2, up to vui_parameters_present_flag. */
	reader_init(&r, nal + 2, len - 2);
	read_bits(&r, 4); /* sps_video_parameter_set_id */
	max_sub_layers_minus1 = read_bits(&r, 3);
	read_bit(&r); /* sps_temporal_id_nesting_flag */
	skip_profile_tier_level(&r, max_sub_layers_minus1);
	read_ue(&r); /* sps_seq_parameter_set_id */
	if (read_ue(&r) == 3) /* chroma_format_idc */
		read_bit(&r); /* separate_colour_plane_flag */
	read_ue(&r); /* pic_width_in_luma_samples */
	read_ue(&r); /* pic_height_in_luma_samples */
	if (read_bit(&r)) { /* conformance_window_flag */
		for (i = 0; i < 4; i++)
			read_ue(&r);
	}
	read_ue(&r); /* bit_depth_luma_minus8 */
	read_ue(&r); /* bit_depth_chroma_minus8 */
	log2_max_poc_lsb = read_ue(&r) + 4;
	if (log2_max_poc_lsb > 16)
		return 0;
	i = read_bit(&r) ? 0 : max_sub_layers_minus1; /* sub_layer_ordering_info_present */
	for (; i <= max_sub_layers_minus1; i++) {
		read_ue(&r);
		read_ue(&r);
		read_ue(&r);
	}
	for (i = 0; i < 6; i++)
		read_ue(&r); /* the coding and transform block sizes and depths */
	if (read_bit(&r) && read_bit(&r)) /* scaling_list_enabled, sps_..._data_present */
		skip_hevc_scaling_list_data(&r);
	read_bit(&r); /* amp_enabled_flag */
	read_bit(&r); /* sample_adaptive_offset_enabled_flag */
	if (read_bit(&r)) { /* pcm_enabled_flag */
		read_bits(&r, 8);
		read_ue(&r);
		read_ue(&r);
		read_bit(&r);
	}
	if (skip_hevc_st_ref_pic_sets(&r, read_ue(&r)))
		return 0;
	if (read_bit(&r)) { /* long_term_ref_pics_present_flag */
		n = read_ue(&r);
		if (n > 32)
			return 0;
		for (i = 0; i < n; i++) {
			read_bits(&r, log2_max_poc_lsb);
			read_bit(&r);
		}
	}
	read_bit(&r); /* sps_temporal_mvp_enabled_flag */
	read_bit(&r); /* strong_intra_smoothing_enabled_flag */
	vui_flag = r.bit;
	if (read_bit(&r)) { /* vui_parameters_present_flag: E.2.1 */
		if (read_bit(&r) && read_bits(&r, 8) == 255) /* aspect_ratio_idc */
			read_bits(&r, 32);
		if (read_bit(&r)) /* overscan_info_present_flag */
			read_bit(&r);
		if (r.overrun || read_bit(&r)) /* video_signal_type_present_flag */
			return 0; /* unreadable, or it says already */
	}
	if (r.overrun)
		return 0;

	/* The rbsp_stop_one_bit: the last bit set. */
	end = r.len * 8;
	while (end > r.bit && !((r.rbsp[(end - 1) / 8] >> (7 - (end - 1) % 8)) & 1))
		end--;
	if (end <= r.bit)
		return 0;
	end--;

	w.bit = 0;
	w.overflow = 0;
	if (r.bit == vui_flag + 1) {
		/* No VUI: one with the video signal type alone. */
		copy_bits(&w, &r, 0, vui_flag);
		put_bit(&w, 1); /* vui_parameters_present_flag */
		put_bit(&w, 0); /* aspect_ratio_info_present_flag */
		put_bit(&w, 0); /* overscan_info_present_flag */
		put_video_signal(&w, colour);
		put_bit(&w, 0); /* chroma_loc_info_present_flag */
		put_bit(&w, 0); /* neutral_chroma_indication_flag */
		put_bit(&w, 0); /* field_seq_flag */
		put_bit(&w, 0); /* frame_field_info_present_flag */
		put_bit(&w, 0); /* default_display_window_flag */
		put_bit(&w, 0); /* vui_timing_info_present_flag */
		put_bit(&w, 0); /* bitstream_restriction_flag */
	} else {
		/* A VUI without it: the flag set and the fields put in. */
		copy_bits(&w, &r, 0, r.bit - 1);
		put_video_signal(&w, colour);
	}
	copy_bits(&w, &r, r.bit, end);
	put_bit(&w, 1); /* rbsp_stop_one_bit */
	while (w.bit % 8)
		put_bit(&w, 0);
	return put_nal_out(nal, 2, &w, out, out_size);
}

size_t kvmv_hevc_au_colour(uint8_t *buf, size_t len, size_t capacity,
			   const struct kvmv_vui_colour *colour)
{
	if (colour == NULL)
		return len;
	return au_rewrite_sps(KVMV_CODEC_KIND_HEVC, buf, len, capacity,
			      kvmv_hevc_sps_colour, colour);
}
