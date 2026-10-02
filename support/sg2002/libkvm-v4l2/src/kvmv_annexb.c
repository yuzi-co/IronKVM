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

size_t kvmv_annexb_split(const uint8_t *buf, size_t len, struct kvmv_nal *nals,
			 size_t max_nals)
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
			nals[count].type = buf[start] & 0x1f;
			count++;
		}
		pos = next;
		code_len = next_len;
	}
	return count;
}

void kvmv_au_inspect(const uint8_t *buf, size_t len, struct kvmv_au_info *info)
{
	struct kvmv_nal nals[64];
	size_t count;
	size_t i;

	memset(info, 0, sizeof(*info));
	count = kvmv_annexb_split(buf, len, nals, sizeof(nals) / sizeof(nals[0]));
	info->nal_count = (unsigned int)count;
	for (i = 0; i < count; i++) {
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

int kvmv_au_type(const struct kvmv_au_info *info)
{
	if (info->idr)
		return IMG_H264_TYPE_IF;
	if (info->slice)
		return IMG_H264_TYPE_PF;
	if (info->sps || info->pps)
		return KVMV_AU_HEADERS_ONLY;
	return KVMV_AU_NO_PICTURE;
}

void kvmv_ps_cache_reset(struct kvmv_ps_cache *cache)
{
	cache->sps_len = 0;
	cache->pps_len = 0;
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
	size_t count = kvmv_annexb_split(buf, len, nals,
					 sizeof(nals) / sizeof(nals[0]));
	size_t i;

	for (i = 0; i < count; i++) {
		if (nals[i].type == KVMV_NAL_SPS)
			cache_store(cache->sps, &cache->sps_len,
				    buf + nals[i].start, nals[i].size);
		else if (nals[i].type == KVMV_NAL_PPS)
			cache_store(cache->pps, &cache->pps_len,
				    buf + nals[i].start, nals[i].size);
	}
}

size_t kvmv_ps_cache_prefix(const struct kvmv_ps_cache *cache,
			    const struct kvmv_au_info *info, uint8_t *out,
			    size_t out_size)
{
	size_t need;

	if (!info->idr || (info->sps && info->pps))
		return 0;
	if (cache->sps_len == 0 || cache->pps_len == 0)
		return 0;
	/*
	 * Both sets, even when the unit carries one of them: a PPS is parsed
	 * against the SPS it names, so the pair travels together.
	 */
	need = cache->sps_len + cache->pps_len;
	if (need > out_size)
		return 0;
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

int kvmv_h264_sps_size(const uint8_t *nal, size_t len, unsigned int *width,
		       unsigned int *height)
{
	struct bit_reader r;
	unsigned int profile;
	unsigned int chroma_format = 1;
	unsigned int separate_planes = 0;
	unsigned int width_mbs, height_units, frame_mbs_only;
	unsigned int crop_left = 0, crop_right = 0, crop_top = 0, crop_bottom = 0;
	unsigned int crop_x, crop_y;
	unsigned int w, h;

	if (nal == NULL || len < 4 || (nal[0] & 0x1f) != KVMV_NAL_SPS)
		return -1;

	reader_init(&r, nal + 1, len - 1);
	profile = read_bits(&r, 8);
	read_bits(&r, 8); /* constraint flags */
	read_bits(&r, 8); /* level */
	read_ue(&r); /* seq_parameter_set_id */

	if (profile == 100 || profile == 110 || profile == 122 || profile == 244 ||
	    profile == 44 || profile == 83 || profile == 86 || profile == 118 ||
	    profile == 128 || profile == 138 || profile == 139 || profile == 134 ||
	    profile == 135) {
		chroma_format = read_ue(&r);
		if (chroma_format == 3)
			separate_planes = read_bit(&r);
		read_ue(&r); /* bit_depth_luma_minus8 */
		read_ue(&r); /* bit_depth_chroma_minus8 */
		read_bit(&r); /* qpprime_y_zero_transform_bypass */
		if (read_bit(&r)) {
			unsigned int lists = chroma_format != 3 ? 8 : 12;
			unsigned int i;

			for (i = 0; i < lists; i++)
				if (read_bit(&r))
					skip_scaling_list(&r, i < 6 ? 16 : 64);
		}
	}

	read_ue(&r); /* log2_max_frame_num_minus4 */
	switch (read_ue(&r)) { /* pic_order_cnt_type */
	case 0:
		read_ue(&r);
		break;
	case 1: {
		unsigned int cycle, i;

		read_bit(&r);
		read_se(&r);
		read_se(&r);
		cycle = read_ue(&r);
		if (cycle > 255)
			return -1;
		for (i = 0; i < cycle; i++)
			read_se(&r);
		break;
	}
	default:
		break;
	}
	read_ue(&r); /* max_num_ref_frames */
	read_bit(&r); /* gaps_in_frame_num_value_allowed */
	width_mbs = read_ue(&r) + 1;
	height_units = read_ue(&r) + 1;
	frame_mbs_only = read_bit(&r);
	if (!frame_mbs_only)
		read_bit(&r); /* mb_adaptive_frame_field */
	read_bit(&r); /* direct_8x8_inference */
	if (read_bit(&r)) {
		crop_left = read_ue(&r);
		crop_right = read_ue(&r);
		crop_top = read_ue(&r);
		crop_bottom = read_ue(&r);
	}
	if (r.overrun)
		return -1;

	w = width_mbs * 16;
	h = (2 - frame_mbs_only) * height_units * 16;
	if (chroma_format == 0 || separate_planes) {
		crop_x = 1;
		crop_y = 2 - frame_mbs_only;
	} else {
		crop_x = chroma_format == 3 ? 1 : 2;
		crop_y = (chroma_format == 1 ? 2 : 1) * (2 - frame_mbs_only);
	}
	if (crop_x * (crop_left + crop_right) >= w ||
	    crop_y * (crop_top + crop_bottom) >= h)
		return -1;
	*width = w - crop_x * (crop_left + crop_right);
	*height = h - crop_y * (crop_top + crop_bottom);
	return 0;
}
