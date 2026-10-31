/*
 * The native capture worker (ironkvm-dist#91): a thread that makes
 * kvmv_read_video calls on the server's behalf and hands the answers over
 * through the shared block of kvmv_native_capture.h and two eventfds. See that
 * header for the protocol.
 *
 * The reads go through ops, so the host test drives the worker with a fake
 * library.
 */
#ifndef KVMV_WORKER_H
#define KVMV_WORKER_H

#include <pthread.h>
#include <stdint.h>

#include "kvmv_native_capture.h"

struct kvmv_worker_ops {
	int (*read)(uint16_t width, uint16_t height, uint8_t codec,
		    uint16_t bitrate, uint8_t gop, uint8_t fps, uint8_t **data,
		    uint32_t *size);
	int (*release)(uint8_t **data);
	int (*times)(uint64_t *times, int n);
};

struct kvmv_worker {
	struct kvmv_native_capture shared;
	const struct kvmv_worker_ops *ops;
	pthread_t thread;
	int running; /* the thread exists; guarded by the caller */
	int stop; /* __atomic */
	uint32_t last_seq; /* the worker thread's own */
	uint8_t *held; /* the slot of the last answer; the worker thread's own */
};

/* Sets up the block and its eventfds, once; a second call does nothing.
 * Answers 0, or -1 with errno when an eventfd cannot be made. */
int kvmv_worker_init(struct kvmv_worker *w, const struct kvmv_worker_ops *ops);
/* Starts the thread unless it runs. Answers 0 or -1. */
int kvmv_worker_start(struct kvmv_worker *w);
/* Stops and joins the thread, after the read it is in, and releases the
 * slot it holds. Does nothing when it does not run. */
void kvmv_worker_stop(struct kvmv_worker *w);

#endif
