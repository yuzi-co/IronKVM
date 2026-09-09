#ifndef __KVM_MMF_HPP__
#define __KVM_MMF_HPP__

#include "stdint.h"

typedef struct {
    uint8_t *data[8];
    int data_size[8];
    int count;
} mmf_stream_t;

typedef struct {
    uint8_t type;           // 0, jpg; 1, h265; 2, h264
    int w;
	int h;
	int fmt;
	uint8_t jpg_quality;	// jpeg
	int gop;				// h264/h265
	int intput_fps;			// h264/h265
	int output_fps;			// h264/h265
	int bitrate;			// h264/h265
} mmf_venc_cfg_t;

// init sys
int mmf_init(void);
int mmf_deinit(void);
int mmf_try_deinit(bool force);
bool mmf_is_init(void);

// manage vi channels(vi->vpssgroup->vpss->frame)
int mmf_get_vi_unused_channel(void);
int mmf_vi_init(void);
int mmf_vi_deinit(void);
int mmf_add_vi_channel_with_enc(int ch, int width, int height, int format);
int mmf_add_vi_channel(int ch, int width, int height, int format);
int mmf_del_vi_channel(int ch);
int mmf_del_vi_channel_all(void);
int mmf_reset_vi_channel(int ch, int width, int height, int format);
bool mmf_vi_chn_is_open(int ch);
int mmf_vi_aligned_width(int ch);
void mmf_set_vi_hmirror(int ch, bool en);
void mmf_set_vi_vflip(int ch, bool en);
void mmf_get_vi_hmirror(int ch, bool *en);
void mmf_get_vi_vflip(int ch, bool *en);

// get vi frame
int mmf_vi_frame_pop(int ch, void **data, int *len, int *width, int *height, int *format);
// Take a VI frame without mapping it into this address space. The frame can
// go to an H.264 encoder with mmf_venc_push_vi(), and it must reach either
// that or mmf_vi_frame_release().
int mmf_vi_frame_pop_native(int ch, int *len, int *width, int *height, int *format);
// Set how many frames a second VI channel ch hands out. 0, or anything at
// or above the sensor rate, means every frame. Remembered across a channel
// rebuild. Returns 0 on success.
int mmf_vi_set_chn_fps(int ch, int dst_fps);
// Map a frame taken with mmf_vi_frame_pop_native so the CPU may read it.
// Returns NULL if there is no such frame. The frame still has to reach an
// encoder push or mmf_vi_frame_release afterwards.
void *mmf_vi_frame_map(int ch);
void mmf_vi_frame_free(int ch);
// Release the current VI frame immediately. mmf_vi_frame_free() defers
// release so the frame can be sent directly to VENC without a second copy.
void mmf_vi_frame_release(int ch);

// invert format
int mmf_invert_format_to_maix(int mmf_format);
int mmf_invert_format_to_mmf(int maix_format);

// venc
int mmf_enc_jpg_init(int ch, int w, int h, int format, int quality);
int mmf_enc_jpg_deinit(int ch);
int mmf_enc_jpg_push(int ch, uint8_t *data, int w, int h, int format);
int mmf_enc_jpg_push_with_quality(int ch, uint8_t *data, int w, int h, int format, int quality);
// Encode a frame taken with mmf_vi_frame_pop_native, without mapping it.
int mmf_enc_jpg_push_vi_with_quality(int ch, int vi_ch, int quality);
int mmf_enc_jpg_pop(int ch, uint8_t **data, int *size);
int mmf_enc_jpg_free(int ch);
int mmf_add_venc_channel(int ch, mmf_venc_cfg_t *cfg);
int mmf_del_venc_channel(int ch);
int mmf_del_venc_channel_all();
int mmf_venc_push(int ch, uint8_t *data, int w, int h, int format);
int mmf_venc_push_vi(int ch, int vi_ch);
int mmf_venc_pop(int ch, mmf_stream_t *stream);
int mmf_venc_free(int ch);

#endif // __KVM_MMF_HPP__
