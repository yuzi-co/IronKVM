/*
 * The native capture worker. See kvmv_worker.h and kvmv_native_capture.h.
 */
#define _GNU_SOURCE

#include <errno.h>
#include <string.h>
#include <sys/eventfd.h>
#include <time.h>
#include <unistd.h>

#include "kvmv_worker.h"

static void signal_fd(int fd)
{
	uint64_t one = 1;

	while (write(fd, &one, sizeof(one)) < 0 && errno == EINTR)
		;
}

static int64_t clock_ns(clockid_t clock)
{
	struct timespec ts;

	clock_gettime(clock, &ts);
	return (int64_t)ts.tv_sec * 1000000000LL + ts.tv_nsec;
}

static void give_back(struct kvmv_worker *w)
{
	if (w->held != NULL) {
		w->ops->release(&w->held);
		w->held = NULL;
	}
}

/* One request: the server's previous answer is done with, so its slot goes
 * back before the read, which may need it. */
static void serve(struct kvmv_worker *w, uint32_t seq)
{
	struct kvmv_native_capture *s = &w->shared;
	uint8_t *data = NULL;
	uint32_t size = 0;
	uint64_t times[KVMV_NATIVE_CAPTURE_TIMES];
	int ntimes = 0;
	int result;
	int i;

	give_back(w);
	result = w->ops->read(s->width, s->height, s->codec, s->bitrate, s->gop,
			      s->fps, &data, &size);
	if (result >= 0 && data != NULL) {
		w->held = data;
		if (size > 0 && w->ops->times != NULL)
			ntimes = w->ops->times(times, KVMV_NATIVE_CAPTURE_TIMES);
	} else {
		data = NULL;
		size = 0;
	}
	if (ntimes < 0)
		ntimes = 0;
	for (i = 0; i < ntimes; i++)
		s->times[i] = times[i];
	s->ntimes = (uint32_t)ntimes;
	s->mono_ns = clock_ns(CLOCK_MONOTONIC);
	s->real_ns = clock_ns(CLOCK_REALTIME);
	s->result = result;
	s->data = data;
	s->data_size = size;
	__atomic_add_fetch(&s->reads, 1, __ATOMIC_RELAXED);
	__atomic_store_n(&s->res_seq, seq, __ATOMIC_RELEASE);
	signal_fd(s->done_fd);
}

static void *worker_main(void *arg)
{
	struct kvmv_worker *w = arg;
	struct kvmv_native_capture *s = &w->shared;

	for (;;) {
		uint64_t count;
		uint32_t seq;
		ssize_t n = read(s->request_fd, &count, sizeof(count));

		if (n < 0 && errno == EINTR)
			continue;
		if (__atomic_load_n(&w->stop, __ATOMIC_ACQUIRE))
			break;
		__atomic_add_fetch(&s->wakeups, 1, __ATOMIC_RELAXED);
		if (n != (ssize_t)sizeof(count))
			continue;
		seq = __atomic_load_n(&s->req_seq, __ATOMIC_ACQUIRE);
		if (seq == w->last_seq)
			continue;
		w->last_seq = seq;
		serve(w, seq);
	}
	give_back(w);
	return NULL;
}

int kvmv_worker_init(struct kvmv_worker *w, const struct kvmv_worker_ops *ops)
{
	struct kvmv_native_capture *s = &w->shared;
	int request_fd, done_fd;

	if (s->version == KVMV_NATIVE_CAPTURE_VERSION)
		return 0;
	request_fd = eventfd(0, EFD_CLOEXEC);
	if (request_fd < 0)
		return -1;
	done_fd = eventfd(0, EFD_CLOEXEC | EFD_NONBLOCK);
	if (done_fd < 0) {
		int saved = errno;

		close(request_fd);
		errno = saved;
		return -1;
	}
	memset(w, 0, sizeof(*w));
	w->ops = ops;
	s->size = sizeof(*s);
	s->request_fd = request_fd;
	s->done_fd = done_fd;
	s->version = KVMV_NATIVE_CAPTURE_VERSION;
	return 0;
}

int kvmv_worker_start(struct kvmv_worker *w)
{
	int err;

	if (w->running)
		return 0;
	__atomic_store_n(&w->stop, 0, __ATOMIC_RELEASE);
	err = pthread_create(&w->thread, NULL, worker_main, w);
	if (err != 0) {
		errno = err;
		return -1;
	}
	pthread_setname_np(w->thread, "kvmv-capture");
	w->running = 1;
	return 0;
}

void kvmv_worker_stop(struct kvmv_worker *w)
{
	if (!w->running)
		return;
	__atomic_store_n(&w->stop, 1, __ATOMIC_RELEASE);
	signal_fd(w->shared.request_fd);
	pthread_join(w->thread, NULL);
	w->running = 0;
}
