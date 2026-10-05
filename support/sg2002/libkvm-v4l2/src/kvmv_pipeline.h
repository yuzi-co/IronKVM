/*
 * The V4L2 pipeline behind libkvm-v4l2:
 *
 *   capture node (UYVY, LT6911 over CSI-2)
 *     -> VPSS mem2mem scaler (UYVY to NV12, optional downscale)
 *     -> Coda980 mem2mem encoder (NV12 to H.264), or the WAVE420L (NV12 to
 *        H.265)
 *
 * H.265 at the capture's own size can skip the scaler: the WAVE420L reads the
 * UYVY capture buffer itself (kvmv_pipe_cfg.direct).
 *
 * Capture buffers are exported as dma-bufs and imported by the scaler. The
 * buffers between the scaler and the encoder come from a dma-heap and are
 * imported by both, so no frame is copied by the CPU. Only the bitstream is
 * copied out, into the buffer handed to the server.
 *
 * This is the pipeline nixos-nanokvm's sg2002-h264-bridge proved on the board
 * (about 51 fps at 1080p), run one frame per call instead of free running:
 * kvmv_read_* takes the newest captured frame, scales it, encodes it and
 * returns its access unit, which is what the vendor library does with its VI
 * channel.
 */
#ifndef KVMV_PIPELINE_H
#define KVMV_PIPELINE_H

#include <stddef.h>
#include <stdint.h>
#include <linux/videodev2.h>

#include "kvmv_annexb.h"
#include "kvmv_policy.h"

#define KVMV_PATH_MAX 64
#define KVMV_START_STEPS 10

struct kvmv_devices {
	char capture[KVMV_PATH_MAX];
	char scaler[KVMV_PATH_MAX];
	char encoder[KVMV_PATH_MAX]; /* H.264; empty when the kernel has none */
	char encoder_hevc[KVMV_PATH_MAX]; /* H.265 (wave420l); empty when none */
	char subdev[KVMV_PATH_MAX]; /* empty when the receiver has no node */
	char jpeg[KVMV_PATH_MAX]; /* sg2002-jpeg; empty when the kernel has none */
};

/*
 * Find the video nodes by driver and capability. Environment variables
 * KVMV_CAPTURE_DEV, KVMV_SCALER_DEV, KVMV_ENCODER_DEV, KVMV_HEVC_ENCODER_DEV,
 * KVMV_JPEG_DEV and KVMV_SUBDEV override discovery. The JPEG unit is optional,
 * and one encoder of the two is enough. Returns 0 when capture, scaler and an
 * encoder were found; otherwise -1, with the missing roles named in missing.
 */
int kvmv_find_devices(struct kvmv_devices *devices, char *missing,
		      size_t missing_size);

/* Find the HDMI receiver's subdevice node. Returns 0 when found. */
int kvmv_find_subdev(char *path, size_t size);

/*
 * Ask the receiver what it sees. need_width and need_height are the capture
 * node's frame size, or 0 to accept any. The timings are copied to out when
 * it is not NULL.
 */
enum kvmv_signal kvmv_query_signal(int subdev_fd, unsigned int need_width,
				   unsigned int need_height,
				   struct v4l2_dv_timings *out);

struct kvmv_buf {
	void *addr;
	size_t length;
	int fd;
};

struct kvmv_pipe_cfg {
	enum kvmv_codec codec; /* H.264 (the Coda980) or H.265 (the WAVE420L) */
	unsigned int req_width, req_height;
	uint32_t bitrate_bps;
	unsigned int gop;
	unsigned int fps;
	/* QP range, both 0 to leave the encoder's own; 0 ms keeps the
	 * encoder's initial rate-control delay. */
	int min_qp, max_qp;
	unsigned int vbv_delay_ms;
	unsigned int capture_buffers;
	unsigned int mid_buffers;
	unsigned int bitstream_buffers;
	/* Longest wait at a read for a capture frame about to complete,
	 * instead of taking one more than half a frame time old; 0 never. */
	unsigned int pickup_wait_ms;
	/* Start the capture first when kvmv_pipe_park left buffers that fit,
	 * so the receiver and the CSI link come up while the encoder is set
	 * up, instead of after it. */
	int early_capture;
	/* Keep the encoder node, its bitstream buffers and the middle buffers
	 * in kvmv_pipe_park, for the next start of the same codec and size. */
	int park_encoder;
	/* How the encoder is started: see enum kvmv_prime. */
	enum kvmv_prime prime;
	/*
	 * H.265 at the capture's own size: give the encoder the capture buffer
	 * itself, without the scaler, when it takes the capture format (the
	 * WAVE420L with ironkvm-dist patch 0933 reads packed UYVY). The
	 * pipeline falls back to the scaler when the encoder refuses it;
	 * struct kvmv_pipe's direct says which it got.
	 */
	int direct;
};

/* What the encoder reports back. -1 where it would not say. */
struct kvmv_applied {
	int64_t bitrate_bps;
	int gop;
	int fps_numerator; /* frames */
	int fps_denominator; /* per this many seconds */
	int min_qp, max_qp; /* -1 on a kernel without the controls */
};

/*
 * Where the time of a read goes, summed over the frames since the last reset.
 * Microseconds. The library logs the averages with KVMV_DEBUG.
 */
struct kvmv_stage_times {
	uint64_t capture_us; /* waiting for a captured frame */
	uint64_t scale_us; /* VPSS, queue to dequeue */
	uint64_t encode_us; /* Coda, queue to dequeue */
	uint64_t copy_us; /* bitstream out of the encoder's buffer (the caller adds it) */
	uint64_t age_us; /* capture timestamp to the encoded frame, summed */
	uint64_t age_max_us; /* the largest of those */
	uint64_t pick_age_us; /* capture timestamp to the frame being taken, summed */
	uint64_t seq_gap; /* capture sequence numbers skipped between taken frames, summed */
	unsigned int pickup_waits; /* reads that waited for the next frame */
	unsigned int ahead; /* frames the scaler had started on during the encode before */
	uint64_t overlap_us; /* scaler started to encoder done, on reads that started it */
	unsigned int frames;
};

/* An NV12 picture the CPU can read. */
struct kvmv_nv12 {
	const uint8_t *y; /* luma plane, height rows of stride bytes */
	const uint8_t *uv; /* interleaved chroma, height / 2 rows of stride bytes */
	unsigned int width, height; /* the picture, without any padding */
	unsigned int stride;
	int fd; /* the dma-buf holding it, or -1 */
	size_t fd_size; /* that buffer's size */
	int full_range; /* 1 when the samples are full range (0..255) */
};

/* The scaler context behind kvmv_pipe_snapshot. */
struct kvmv_snap {
	int fd; /* -1 until the first snapshot */
	struct kvmv_plan plan;
	struct v4l2_pix_format fmt;
	struct kvmv_buf buf; /* from the dma-heap, mapped for reading */
	int out_on, cap_on;
	int synced; /* the CPU holds the buffer (DMA_BUF_SYNC_START done) */
};

struct kvmv_pipe {
	enum kvmv_codec codec;
	int cap_fd, vpss_fd, enc_fd, heap_fd;
	char scaler_path[KVMV_PATH_MAX];
	char enc_path[KVMV_PATH_MAX];
	struct kvmv_snap snap;
	struct v4l2_pix_format cap_fmt;
	struct v4l2_pix_format vpss_in_fmt, vpss_out_fmt;
	struct v4l2_pix_format enc_out_fmt, enc_cap_fmt;
	struct kvmv_plan plan;
	struct kvmv_buf *cap;
	unsigned int cap_count;
	struct kvmv_buf *mid;
	unsigned int mid_count;
	unsigned int next_mid;
	struct kvmv_buf *bs;
	unsigned int bs_count;
	/*
	 * The encoder reads the capture buffers (kvmv_pipe_cfg.direct): its
	 * OUTPUT queue imports them by dma-buf, index for index, enc_out_count
	 * slots, and the scaler context of this pipe is not used. No middle
	 * buffers then.
	 */
	int direct;
	int want_direct; /* kvmv_pipe_cfg.direct, for kvmv_pipe_park */
	unsigned int enc_out_count;
	int cap_on, vpss_out_on, vpss_cap_on, enc_out_on, enc_cap_on;
	int running;
	struct kvmv_applied applied;
	struct kvmv_stage_times times;
	uint64_t cap_ts_us; /* CLOCK_MONOTONIC capture time of the frame being scaled */
	uint32_t cap_seq; /* its sequence number */
	uint32_t cap_interval_us; /* the source's frame time, from the timestamps */
	uint32_t pickup_wait_max_us; /* kvmv_pipe_cfg.pickup_wait_ms */
	/* With an early capture start: frames that completed before this
	 * CLOCK_MONOTONIC time are given back, not taken. 0 once a frame was. */
	uint64_t fresh_after_us;
	int early; /* this start began with the capture (cfg.early_capture) */
	int cap_allocated; /* this start allocated the capture buffers */
	int park_encoder; /* cfg.park_encoder */
	int enc_reused; /* the encoder and its buffers came from kvmv_pipe_park */
	/* KVMV_PRIME_ASYNC: the priming picture's output is still to be
	 * collected, and a keyframe asked for meanwhile is asked for after it. */
	int prime_pending;
	int key_after_prime;
	/*
	 * Scale-ahead. With scale_ahead set (by the caller, before each
	 * kvmv_pipe_encode), a capture frame that completes while the encoder
	 * works on the picture before it goes to the scaler at once, into the
	 * other middle buffer, and the next kvmv_pipe_encode takes that picture
	 * instead of waiting for and scaling a frame of its own. The scaler and
	 * the encoder then work side by side instead of one after the other.
	 * ahead_mid is the middle buffer being filled, -1 for none; ahead_ci
	 * the capture buffer the scaler reads; ahead_ts_us that frame's capture
	 * time.
	 */
	int scale_ahead;
	int ahead_mid;
	unsigned int ahead_ci;
	uint64_t ahead_ts_us;
	/* Scale-ahead starts no earlier than this after the encoder was given
	 * its picture (KVMV_AHEAD_DELAY_US), 0 at once; ahead_start_us is when
	 * it last started, for times.overlap_us. */
	uint32_t ahead_delay_us;
	uint64_t ahead_start_us;
	/* kvmv_pipe_start's steps, microseconds from its start: open, early
	 * capture STREAMON, negotiate, middle buffers, bitstream buffers,
	 * encoder STREAMON, priming encode, capture buffers, capture STREAMON,
	 * scaler STREAMON. For KVMV_DEBUG. */
	uint32_t start_us[KVMV_START_STEPS];
	char error[192];
};

enum kvmv_pipe_status {
	KVMV_PIPE_OK = 0,
	KVMV_PIPE_NO_FRAME, /* the capture node delivered nothing in time */
	KVMV_PIPE_SOURCE_CHANGED, /* the source changed mode under the pipeline */
	KVMV_PIPE_ERROR,
};

struct kvmv_encoded {
	unsigned int index;
	const uint8_t *data;
	size_t size;
	uint32_t flags;
};

/* Every ioctl the pipeline makes goes through this, so a test can stand in
 * for the drivers. */
extern int (*kvmv_ioctl_hook)(int fd, unsigned long request, void *arg);

void kvmv_pipe_init(struct kvmv_pipe *pipe);

/*
 * Set formats, crops, frame rate and rate control on already open nodes, and
 * check that the scaler's output surface is laid out exactly as the encoder
 * expects its input. No buffers are touched, so this is what the host test
 * drives through kvmv_ioctl_hook.
 */
int kvmv_pipe_negotiate(struct kvmv_pipe *pipe, const struct kvmv_pipe_cfg *cfg);

/* Open the nodes, negotiate, allocate, prime the encoder and start streaming. */
int kvmv_pipe_start(struct kvmv_pipe *pipe, const struct kvmv_devices *devices,
		    const struct kvmv_pipe_cfg *cfg);

/* Stop streaming and release everything. Safe on a stopped pipe. */
void kvmv_pipe_stop(struct kvmv_pipe *pipe);

/*
 * As kvmv_pipe_stop, but keep the capture node open with its buffers, so the
 * next kvmv_pipe_start takes them instead of allocating and clearing new ones
 * (about 50 ms for three 1080p frames). The capture is streamed off, so its
 * clocks and DMA stop. With cfg.park_encoder the encoder node is kept too,
 * streamed off (no sequence, no codec SRAM, no reference frames), with its
 * bitstream buffers and the middle buffers; the next start of the same codec
 * at the same size takes them instead of allocating and clearing new ones.
 * kvmv_pipe_unpark releases all of it.
 */
void kvmv_pipe_park(struct kvmv_pipe *pipe);
void kvmv_pipe_unpark(void);

/*
 * Take the newest captured frame and scale it into a middle buffer. On
 * KVMV_PIPE_OK *mid names the buffer, which holds the NV12 picture for
 * kvmv_pipe_encode_mid.
 */
enum kvmv_pipe_status kvmv_pipe_scale(struct kvmv_pipe *pipe,
				      unsigned int timeout_ms, unsigned int *mid);

/* Encode a middle buffer filled by kvmv_pipe_scale. */
enum kvmv_pipe_status kvmv_pipe_encode_mid(struct kvmv_pipe *pipe,
					   unsigned int mid,
					   struct kvmv_encoded *out);

/*
 * Take the newest frame, scale it and encode it; with pipe->direct, encode the
 * capture buffer itself and give it back to the capture once the encoder has
 * read it. On KVMV_PIPE_OK the caller
 * copies out->data with kvmv_copy_from_device and then calls kvmv_pipe_release.
 * With pipe->scale_ahead the picture may be one the scaler started on during
 * the previous call (see struct kvmv_pipe), and this call may leave the next
 * one with the scaler.
 */
enum kvmv_pipe_status kvmv_pipe_encode(struct kvmv_pipe *pipe,
				       unsigned int timeout_ms,
				       struct kvmv_encoded *out);
void kvmv_pipe_release(struct kvmv_pipe *pipe, const struct kvmv_encoded *encoded);

/*
 * Snapshots for kvmv_read_img. A second context on the scaler node takes the
 * same captured frames the H.264 path takes (it imports the same capture
 * dma-bufs) and scales them to the picture size asked for, into one NV12
 * buffer of its own from the dma-heap. Its size is independent of the H.264
 * stream's, so a screenshot or VNC reader never rebuilds the encoder, and the
 * scaler's mem2mem queue serialises its jobs with the H.264 path's.
 *
 * kvmv_pipe_snapshot fills the buffer and, with cpu_read, makes it readable
 * by the CPU (a cached mapping, synced for reading); the caller reads the
 * picture and then calls kvmv_pipe_snapshot_done. Without cpu_read the
 * buffer is left to devices (the JPEG unit imports image->fd) and no cache
 * maintenance is done. The context is set up on the first snapshot, rebuilt
 * when the size changes, and torn down with the pipeline.
 *
 * The scaler is asked for full range, which JPEG wants; image->full_range
 * says whether it gave it (a kernel without ironkvm-dist patch 0908 keeps
 * the source's limited range). The buffer is sized for the JPEG unit too
 * (kvmv_hwjpeg_src_size), which reads chroma rows past a tight 4:2:0 frame.
 */
enum kvmv_pipe_status kvmv_pipe_snapshot(struct kvmv_pipe *pipe,
					 unsigned int width, unsigned int height,
					 unsigned int timeout_ms, int cpu_read,
					 struct kvmv_nv12 *image);
void kvmv_pipe_snapshot_done(struct kvmv_pipe *pipe);

/*
 * A captured frame lent as it is, to a device that reads the capture format
 * itself: the JPEG unit with ironkvm-dist patch 0920 takes the UYVY capture
 * buffer at the capture's own size, so an MJPEG picture needs no scaler
 * pass. kvmv_pipe_lend takes the newest frame, as a snapshot does, and
 * describes it in *frame; kvmv_pipe_give_back queues it to the capture
 * again. In between the frame is the caller's, and the pipeline must not be
 * stopped.
 */
struct kvmv_frame {
	unsigned int index; /* the capture buffer */
	int fd; /* its dma-buf */
	size_t size; /* that buffer's size */
	const struct v4l2_pix_format *fmt; /* the capture format */
};

enum kvmv_pipe_status kvmv_pipe_lend(struct kvmv_pipe *pipe,
				     unsigned int timeout_ms,
				     struct kvmv_frame *frame);
int kvmv_pipe_give_back(struct kvmv_pipe *pipe, const struct kvmv_frame *frame);

/*
 * Copy out of the encoder's bitstream buffer. That buffer comes from the
 * Coda's no-map shared-dma-pool, so user space sees it uncached and every load
 * goes to DRAM: copy it once, in 8-byte loads where alignment allows, and
 * parse the copy, never the mapping.
 */
void kvmv_copy_from_device(uint8_t *dst, const uint8_t *src, size_t len);

/* Runtime changes. Each returns 0 when the encoder took it; the caller
 * rebuilds the pipeline otherwise. */
int kvmv_pipe_set_bitrate(struct kvmv_pipe *pipe, uint32_t bitrate_bps);
int kvmv_pipe_set_gop(struct kvmv_pipe *pipe, unsigned int gop);
int kvmv_pipe_set_fps(struct kvmv_pipe *pipe, unsigned int fps);
int kvmv_pipe_force_key(struct kvmv_pipe *pipe);

#endif
