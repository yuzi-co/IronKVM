/*
 * The native capture worker of libkvm-v4l2 (ironkvm-dist#91).
 *
 * kvmv_read_video blocks for the whole of a picture's scale and encode, 13 to
 * 26 ms at 1080p on the mainline kernel. Made from Go, that is a blocking cgo
 * call per frame, and the runtime hands its P to another thread on every one
 * of them. The worker makes the same call on a thread of the library's own:
 * the server fills in a request here, writes to request_fd, and waits for
 * done_fd in its netpoller; the worker reads, fills in the answer and writes
 * to done_fd. No cgo call is made per frame.
 *
 * The worker only reads when asked, so when to read stays with the server
 * (its read pacer and keyframe policy), exactly as with kvmv_read_video.
 *
 * Not part of kvm_vision.h, and not in Sipeed's library. The server looks the
 * two functions up with dlsym and reads frames with kvmv_read_video without
 * them. This header is shared by the server and libkvm-v4l2.
 */
#ifndef KVMV_NATIVE_CAPTURE_H_
#define KVMV_NATIVE_CAPTURE_H_

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Raised on any change to struct kvmv_native_capture. */
#define KVMV_NATIVE_CAPTURE_VERSION 1U

#define KVMV_NATIVE_CAPTURE_TIMES 5

struct kvmv_native_capture {
	uint32_t version; /* KVMV_NATIVE_CAPTURE_VERSION */
	uint32_t size; /* sizeof(struct kvmv_native_capture) */
	/* An eventfd the server writes 1 to once the request below is filled
	 * in. The worker blocks reading it. */
	int32_t request_fd;
	/* A nonblocking eventfd the worker writes 1 to once the answer below
	 * is filled in. The server dups it and waits for it in its poller. */
	int32_t done_fd;

	/* The request, written by the server before request_fd. req_seq last,
	 * with release order; the worker ignores a wakeup that brings no new
	 * one. The arguments are kvmv_read_video's. */
	uint32_t req_seq;
	uint16_t width;
	uint16_t height;
	uint16_t bitrate;
	uint8_t codec;
	uint8_t gop;
	uint8_t fps;
	uint8_t reserved[3];

	/* The answer, written by the worker before done_fd; res_seq last, with
	 * release order, set to the req_seq it answers. result is
	 * kvmv_read_video's. data points into the library's frame slots and
	 * stays valid until the server's next request: the worker gives the
	 * slot back when it takes that request. */
	uint32_t res_seq;
	int32_t result;
	uint32_t data_size;
	/* How many of times[] kvmv_video_times filled for this answer: 0
	 * unless result is a picture. */
	uint32_t ntimes;
	uint8_t *data;
	/* kvmv_video_times for the picture: CLOCK_MONOTONIC microseconds, [1] a
	 * duration. See abi-extensions.txt. */
	uint64_t times[KVMV_NATIVE_CAPTURE_TIMES];
	/* CLOCK_MONOTONIC and CLOCK_REALTIME in nanoseconds, read back to back
	 * after times[], so the server can put them on the wall clock without
	 * a call of its own. */
	int64_t mono_ns;
	int64_t real_ns;

	/* Counted by the worker, read with relaxed atomics. */
	uint64_t reads;
	uint64_t wakeups; /* request_fd reads, including ones with no request */
};

/*
 * Starts the worker if it is not running and answers its shared block, which
 * lives as long as the library. NULL when version is not the library's or
 * the worker cannot start. Calling it again while the worker runs answers the
 * same block. kvmv_deinit stops the worker, after the read it is in.
 */
struct kvmv_native_capture *kvmv_native_capture_open(uint32_t version);

/* Stops the worker, after the read it is in, and gives back the slot of its
 * last answer. */
void kvmv_native_capture_close(void);

#ifdef __cplusplus
}
#endif

#endif /* KVMV_NATIVE_CAPTURE_H_ */
