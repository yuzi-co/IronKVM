/*
 * kvmv-probe: exercise a libkvm.so on the board through the kvm_vision.h ABI,
 * the way NanoKVM-Server does, and say what came out.
 *
 *   kvmv-probe [-l ./libkvm.so] [-n frames] [-b kbit/s] [-g gop] [-f fps]
 *              [-q jpeg quality] [-o out.h264] [-j dir] [-s] [-v]
 *
 * Stop NanoKVM-Server first: it holds the same devices. The library is loaded
 * with dlopen, so the same binary can be pointed at the vendor library for a
 * comparison. Exit status is the number of failed checks.
 */
#define _GNU_SOURCE
#include <dlfcn.h>
#include <errno.h>
#include <pthread.h>
#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#include "kvmv_annexb.h"

#define IMG_VENC_ERROR (-2)
#define IMG_NOT_EXIST (-1)
#define TYPE_KEY 3
#define TYPE_DELTA 4

typedef void (*init_fn)(uint8_t);
typedef void (*void_fn)(void);
typedef int (*read_img_fn)(uint16_t, uint16_t, uint8_t, uint16_t, uint8_t **, uint32_t *);
typedef int (*read_video_fn)(uint16_t, uint16_t, uint8_t, uint16_t, uint8_t, uint8_t,
			     uint8_t **, uint32_t *);
typedef int (*free_fn)(uint8_t **);
typedef void (*u8_fn)(uint8_t);
typedef uint8_t (*u8_u8_fn)(uint8_t);
typedef uint8_t (*u8_void_fn)(void);

static struct {
	init_fn kvmv_init;
	void_fn kvmv_deinit;
	read_img_fn kvmv_read_img;
	read_video_fn kvmv_read_video;
	free_fn free_kvmv_data;
	void_fn free_all_kvmv_data;
	u8_fn set_h264_gop;
	u8_fn set_h264_fps;
	u8_fn set_capture_fps;
	u8_fn set_frame_detact;
	u8_fn set_venc_auto_recyc;
	u8_u8_fn kvmv_hdmi_control;
	u8_void_fn kvmv_hdmi_signal_active;
} api;

static int failures;
static int verbose;
static FILE *out_file;

static void result(int ok, const char *fmt, ...) __attribute__((format(printf, 2, 3)));
static void result(int ok, const char *fmt, ...)
{
	va_list args;

	printf("%s ", ok ? "PASS" : "FAIL");
	va_start(args, fmt);
	vprintf(fmt, args);
	va_end(args);
	putchar('\n');
	fflush(stdout);
	if (!ok)
		failures++;
}

static void note(const char *fmt, ...) __attribute__((format(printf, 1, 2)));
static void note(const char *fmt, ...)
{
	va_list args;

	printf("     ");
	va_start(args, fmt);
	vprintf(fmt, args);
	va_end(args);
	putchar('\n');
	fflush(stdout);
}

static uint64_t now_us(void)
{
	struct timespec ts;

	clock_gettime(CLOCK_MONOTONIC, &ts);
	return (uint64_t)ts.tv_sec * 1000000U + (uint64_t)ts.tv_nsec / 1000U;
}

static void sleep_until(uint64_t when_us)
{
	uint64_t now = now_us();

	if (when_us > now) {
		struct timespec ts = {
			.tv_sec = (time_t)((when_us - now) / 1000000U),
			.tv_nsec = (long)((when_us - now) % 1000000U) * 1000,
		};
		nanosleep(&ts, NULL);
	}
}

static int load(const char *path)
{
	void *lib = dlopen(path, RTLD_NOW | RTLD_GLOBAL);
	int missing = 0;

	if (lib == NULL) {
		result(0, "dlopen %s: %s", path, dlerror());
		return -1;
	}
#define SYM(name)                                                         \
	do {                                                              \
		*(void **)&api.name = dlsym(lib, #name);                  \
		if (api.name == NULL) {                                   \
			note("missing symbol %s", #name);                 \
			missing++;                                        \
		}                                                         \
	} while (0)
	SYM(kvmv_init);
	SYM(kvmv_deinit);
	SYM(kvmv_read_img);
	SYM(kvmv_read_video);
	SYM(free_kvmv_data);
	SYM(free_all_kvmv_data);
	SYM(set_h264_gop);
	SYM(set_h264_fps);
	SYM(set_capture_fps);
	SYM(set_frame_detact);
	SYM(set_venc_auto_recyc);
	SYM(kvmv_hdmi_control);
	SYM(kvmv_hdmi_signal_active);
#undef SYM
	result(missing == 0, "all 13 kvm_vision.h symbols resolve in %s", path);
	return missing ? -1 : 0;
}

/* The codec run() reads: 1 H.264, 2 H.265. */
static uint8_t run_codec = 1;

struct run_stats {
	unsigned int frames, keys, deltas, errors;
	unsigned int first_key_index; /* UINT32_MAX when none */
	unsigned int keys_complete; /* keyframes carrying (VPS,) SPS, PPS and IDR */
	unsigned int max_key_gap, min_key_gap;
	unsigned int sps_width, sps_height;
	int first_type;
	int error_codes[8]; /* counts of -1..-7 */
	uint64_t bytes;
	uint64_t elapsed_us;
};

static void inspect_frame(struct run_stats *st, const uint8_t *data, uint32_t size,
			  int type, unsigned int *since_key)
{
	enum kvmv_codec kind = run_codec == 2 ? KVMV_CODEC_KIND_HEVC : KVMV_CODEC_KIND_H264;
	struct kvmv_au_info info;

	kvmv_au_inspect_codec(kind, data, size, &info);
	if (type == TYPE_KEY) {
		struct kvmv_nal nals[32];
		size_t n = kvmv_annexb_split_codec(kind, data, size, nals, 32), i;

		st->keys++;
		if ((info.vps || kind != KVMV_CODEC_KIND_HEVC) && info.sps && info.pps && info.idr)
			st->keys_complete++;
		for (i = 0; i < n; i++) {
			if (kind == KVMV_CODEC_KIND_HEVC && nals[i].type == KVMV_HEVC_NAL_SPS)
				kvmv_hevc_sps_size(data + nals[i].start, nals[i].size,
						   &st->sps_width, &st->sps_height);
			else if (kind == KVMV_CODEC_KIND_H264 && nals[i].type == KVMV_NAL_SPS)
				kvmv_h264_sps_size(data + nals[i].start, nals[i].size,
						   &st->sps_width, &st->sps_height);
		}
		if (st->keys > 1) {
			if (*since_key > st->max_key_gap)
				st->max_key_gap = *since_key;
			if (*since_key < st->min_key_gap)
				st->min_key_gap = *since_key;
		}
		*since_key = 0;
	} else {
		st->deltas++;
	}
	if ((type == TYPE_KEY) != (info.idr > 0))
		note("frame %u: type %d but IDR count %u", st->frames, type, info.idr);
	(*since_key)++;
}

/* Read count frames paced at fps, like the server's ticker. */
static void run(struct run_stats *st, unsigned int count, uint16_t w, uint16_t h,
		uint16_t kbps, uint8_t gop, uint8_t fps, unsigned int max_wait_ms)
{
	uint64_t start = now_us(), next = start;
	uint64_t give_up = start + (uint64_t)max_wait_ms * 1000U;
	unsigned int since_key = 0;

	memset(st, 0, sizeof(*st));
	st->first_key_index = UINT32_MAX;
	st->min_key_gap = UINT32_MAX;
	st->first_type = 0;
	while (st->frames < count && now_us() < give_up) {
		uint8_t *data = NULL;
		uint32_t size = 0;
		int type;

		sleep_until(next);
		next += 1000000U / fps;
		type = api.kvmv_read_video(w, h, run_codec, kbps, gop, 0, &data, &size);
		if (type < 0) {
			st->errors++;
			if (type >= -7)
				st->error_codes[-type]++;
			if (verbose)
				note("read -> %d", type);
			continue;
		}
		if (st->frames == 0)
			st->first_type = type;
		if (type == TYPE_KEY && st->first_key_index == UINT32_MAX)
			st->first_key_index = st->frames;
		inspect_frame(st, data, size, type, &since_key);
		st->bytes += size;
		if (out_file)
			fwrite(data, 1, size, out_file);
		api.free_kvmv_data(&data);
		st->frames++;
		if (verbose)
			note("frame %u type %d size %u", st->frames, type, size);
	}
	st->elapsed_us = now_us() - start;
}

static void report(const struct run_stats *st, unsigned int fps)
{
	double secs = (double)st->elapsed_us / 1e6;
	double kbps = st->frames ? (double)st->bytes * 8.0 / 1000.0 /
				   ((double)st->frames / (double)fps) : 0;

	note("%u frames (%u key, %u delta) in %.2f s = %.1f fps; %u errors "
	     "(-1:%d -2:%d -3:%d -4:%d -5:%d -6:%d -7:%d)",
	     st->frames, st->keys, st->deltas, secs,
	     secs > 0 ? st->frames / secs : 0, st->errors, st->error_codes[1],
	     st->error_codes[2], st->error_codes[3], st->error_codes[4],
	     st->error_codes[5], st->error_codes[6], st->error_codes[7]);
	note("stream rate at the nominal %u fps: %.0f kbit/s; key gap %u..%u; SPS %ux%u",
	     fps, kbps, st->min_key_gap == UINT32_MAX ? 0 : st->min_key_gap,
	     st->max_key_gap, st->sps_width, st->sps_height);
}

/* The picture size in a JPEG's SOF segment. Returns 0 when found. */
static int jpeg_size(const uint8_t *d, uint32_t n, unsigned int *w, unsigned int *h)
{
	uint32_t i = 2;

	if (n < 4 || d[0] != 0xff || d[1] != 0xd8)
		return -1;
	while (i + 4 <= n) {
		uint8_t marker;
		uint32_t len;

		if (d[i] != 0xff)
			return -1;
		marker = d[i + 1];
		len = (uint32_t)d[i + 2] << 8 | d[i + 3];
		if (marker >= 0xc0 && marker <= 0xc3) {
			if (i + 9 > n)
				return -1;
			*h = (unsigned int)d[i + 5] << 8 | d[i + 6];
			*w = (unsigned int)d[i + 7] << 8 | d[i + 8];
			return 0;
		}
		if (marker == 0xda)
			return -1;
		i += 2 + len;
	}
	return -1;
}

static int cmp_u64(const void *a, const void *b)
{
	uint64_t x = *(const uint64_t *)a, y = *(const uint64_t *)b;

	return x < y ? -1 : x > y;
}

static const char *jpeg_dir;

/* kvmv_read_img MJPEG at one size: each read's time, and the picture checked. */
static void run_mjpeg(unsigned int w, unsigned int h, unsigned int quality,
		      unsigned int count)
{
	uint64_t times[64];
	unsigned int ok = 0, good = 0, i, sw = 0, sh = 0;
	uint64_t bytes = 0, total;
	int last = 0;

	if (count > 64)
		count = 64;
	for (i = 0; i < count; i++) {
		uint8_t *data = NULL;
		uint32_t size = 0;
		uint64_t t0 = now_us();
		int ret = api.kvmv_read_img((uint16_t)w, (uint16_t)h, 0,
					    (uint16_t)quality, &data, &size);

		times[i] = now_us() - t0;
		last = ret;
		if (ret != 0 || data == NULL)
			continue;
		ok++;
		bytes += size;
		if (size > 4 && data[size - 2] == 0xff && data[size - 1] == 0xd9 &&
		    jpeg_size(data, size, &sw, &sh) == 0 && sw == w && sh == h)
			good++;
		if (jpeg_dir && i == count - 1) {
			char path[256];
			FILE *fp;

			snprintf(path, sizeof(path), "%s/probe-%ux%u.jpg", jpeg_dir, w, h);
			if ((fp = fopen(path, "wb")) != NULL) {
				fwrite(data, 1, size, fp);
				fclose(fp);
				note("wrote %s", path);
			}
		}
		api.free_kvmv_data(&data);
	}
	/* The first read may build the pipeline and the scaler context. */
	for (i = 1, total = 0; i < count; i++)
		total += times[i];
	qsort(times + 1, count - 1, sizeof(times[0]), cmp_u64);
	note("MJPEG %ux%u q%u: first read %.1f ms, then median %.1f ms, min %.1f, max %.1f per frame (%.1f fps back to back); %u bytes average",
	     w, h, quality, times[0] / 1000.0, times[1 + (count - 1) / 2] / 1000.0,
	     times[1] / 1000.0, times[count - 1] / 1000.0,
	     total ? (count - 1) * 1e6 / (double)total : 0.0,
	     ok ? (unsigned int)(bytes / ok) : 0);
	result(ok == count, "MJPEG %ux%u: %u of %u reads answered 0 (last %d)", w, h, ok,
	       count, last);
	result(good == ok && ok > 0, "MJPEG %ux%u: %u of %u are complete JPEGs of that size (last SOF %ux%u)",
	       w, h, good, ok, sw, sh);
}

/* This process's CPU time so far, user and system, in clock ticks. */
static int cpu_ticks(unsigned long long *user, unsigned long long *sys)
{
	char buf[1024], *p;
	FILE *fp = fopen("/proc/self/stat", "re");
	size_t n;

	if (fp == NULL)
		return -1;
	n = fread(buf, 1, sizeof(buf) - 1, fp);
	fclose(fp);
	buf[n] = 0;
	/* Fields 14 and 15, counted after the ")" that ends the name. */
	p = strrchr(buf, ')');
	if (p == NULL ||
	    sscanf(p + 2, "%*c %*d %*d %*d %*d %*d %*u %*u %*u %*u %*u %llu %llu",
		   user, sys) != 2)
		return -1;
	return 0;
}

/*
 * MJPEG paced like the server's ticker: one read every 1/fps s for secs
 * seconds. The delivered rate is what a viewer would get; the CPU is this
 * process's share of the one core over the run.
 */
static void run_mjpeg_paced(unsigned int w, unsigned int h, unsigned int quality,
			    unsigned int fps, unsigned int secs)
{
	unsigned long long u0 = 0, s0 = 0, u1 = 0, s1 = 0;
	uint64_t start, next, end;
	unsigned int ok = 0, errors = 0;
	long hz = sysconf(_SC_CLK_TCK);
	double elapsed, delivered;
	int have_cpu;

	have_cpu = cpu_ticks(&u0, &s0) == 0;
	start = now_us();
	next = start;
	end = start + (uint64_t)secs * 1000000U;
	while (now_us() < end) {
		uint8_t *data = NULL;
		uint32_t size = 0;

		sleep_until(next);
		next += 1000000U / fps;
		if (api.kvmv_read_img((uint16_t)w, (uint16_t)h, 0, (uint16_t)quality,
				      &data, &size) == 0 && data != NULL) {
			ok++;
			api.free_kvmv_data(&data);
		} else {
			errors++;
		}
	}
	elapsed = (double)(now_us() - start) / 1e6;
	have_cpu = have_cpu && cpu_ticks(&u1, &s1) == 0 && hz > 0;
	delivered = ok / elapsed;
	if (have_cpu)
		note("MJPEG %ux%u q%u paced at %u fps: %u pictures in %.1f s = %.1f fps, %u errors; CPU %.1f%% (user %.1f%%, sys %.1f%%)",
		     w, h, quality, fps, ok, elapsed, delivered, errors,
		     100.0 * (double)(u1 - u0 + s1 - s0) / hz / elapsed,
		     100.0 * (double)(u1 - u0) / hz / elapsed,
		     100.0 * (double)(s1 - s0) / hz / elapsed);
	else
		note("MJPEG %ux%u q%u paced at %u fps: %u pictures in %.1f s = %.1f fps, %u errors",
		     w, h, quality, fps, ok, elapsed, delivered, errors);
	/* Holds with the JPEG unit; software JPEG manages about 4 fps at 1080p. */
	result(delivered >= fps * 0.9, "MJPEG %ux%u keeps up with %u fps (%.1f)", w, h,
	       fps, delivered);
}

struct mjpeg_load {
	volatile int stop;
	unsigned int reads, errors;
};

static void *mjpeg_loop(void *arg)
{
	struct mjpeg_load *load = arg;

	while (!load->stop) {
		uint8_t *data = NULL;
		uint32_t size = 0;
		int ret = api.kvmv_read_img(960, 540, 0, 80, &data, &size);

		if (ret == 0) {
			load->reads++;
			api.free_kvmv_data(&data);
		} else {
			load->errors++;
			usleep(10000);
		}
	}
	return NULL;
}

static void usage(const char *argv0)
{
	fprintf(stderr,
		"usage: %s [-l lib] [-n frames] [-b kbit/s] [-g gop] [-f fps] [-q jpeg quality]\n"
		"          [-o out.h264] [-j dir] [-s] [-v]\n"
		"  -s  skip the downscale phase\n"
		"  -j  save the last MJPEG picture of each size in dir\n", argv0);
}

int main(int argc, char **argv)
{
	const char *lib = "./libkvm.so";
	const char *out_path = NULL;
	unsigned int frames = 300, kbps = 4000, gop = 30, fps = 30, quality = 80;
	int skip_scale = 0;
	struct run_stats st;
	int hevc = 0;
	uint8_t *data = NULL;
	uint32_t size = 0;
	uint64_t t0;
	int opt, ret, i;

	while ((opt = getopt(argc, argv, "l:n:b:g:f:q:o:j:sv")) != -1) {
		switch (opt) {
		case 'l': lib = optarg; break;
		case 'q': quality = (unsigned int)atoi(optarg); break;
		case 'j': jpeg_dir = optarg; break;
		case 'n': frames = (unsigned int)atoi(optarg); break;
		case 'b': kbps = (unsigned int)atoi(optarg); break;
		case 'g': gop = (unsigned int)atoi(optarg); break;
		case 'f': fps = (unsigned int)atoi(optarg); break;
		case 'o': out_path = optarg; break;
		case 's': skip_scale = 1; break;
		case 'v': verbose = 1; break;
		default: usage(argv[0]); return 2;
		}
	}
	if (fps < 10 || fps > 60 || gop < 1 || gop > 100 || frames < 1) {
		usage(argv[0]);
		return 2;
	}
	if (out_path && (out_file = fopen(out_path, "wb")) == NULL) {
		perror(out_path);
		return 2;
	}

	/* 1. ABI */
	if (load(lib))
		return 1;

	/* 2. Bring-up, as main.go does it. */
	api.kvmv_init(verbose);
	ret = api.kvmv_hdmi_control(1);
	result(ret == 0, "kvmv_hdmi_control(1) = %d (the vendor library answers 255 off the PCIe board)", ret);
	t0 = now_us();
	while (!api.kvmv_hdmi_signal_active() && now_us() - t0 < 5000000U)
		usleep(100000);
	result(api.kvmv_hdmi_signal_active(), "HDMI signal active after %.1f s",
	       (double)(now_us() - t0) / 1e6);

	/*
	 * 3. Codecs. H.265 needs the WAVE420L (ironkvm-dist#55, wave420l.ko);
	 * without it an H.265 read answers the encoder error.
	 */
	{
		u8_u8_fn supported = (u8_u8_fn)dlsym(RTLD_DEFAULT, "kvmv_codec_supported");

		if (supported != NULL) {
			hevc = supported(2);
			result(supported(0) == 1 && supported(1) == 1,
			       "kvmv_codec_supported: MJPEG %u, H.264 %u, H.265 %u (want 1 1, H.265 %s)",
			       supported(0), supported(1), supported(2),
			       hevc ? "found" : "not found");
		} else {
			note("no kvmv_codec_supported (the vendor library has none)");
		}
	}
	if (!hevc) {
		ret = api.kvmv_read_video(0, 0, 2, kbps, 0, 0, &data, &size);
		result(ret == IMG_VENC_ERROR, "H.265 read without an H.265 encoder answers %d (want -2)",
		       ret);
		if (ret >= 0)
			api.free_kvmv_data(&data);
	}

	/* 3b. MJPEG: time per read at three sizes, then the frame rate at the
	 * stream rate. The library logs which encoder it uses (the JPEG unit,
	 * or software) and whether the scaler gives full range. */
	run_mjpeg(1920, 1080, quality, 30);
	run_mjpeg(1280, 720, quality, 30);
	run_mjpeg(960, 540, quality, 20);
	run_mjpeg_paced(1920, 1080, quality, fps, 10);
	run_mjpeg_paced(1280, 720, quality, fps, 10);

	/* 4. H.264 at the configured rate, as h264_source.go reads it. */
	api.set_h264_fps((uint8_t)fps);
	api.set_capture_fps((uint8_t)fps);
	api.set_h264_gop((uint8_t)gop);
	run(&st, frames, 0, 0, (uint16_t)kbps, 0, (uint8_t)fps,
	    frames * 1000U / fps + 10000U);
	report(&st, fps);
	result(st.frames == frames, "read %u of %u H.264 frames", st.frames, frames);
	result(st.first_type == TYPE_KEY, "first frame is a keyframe (type %d)", st.first_type);
	/* Delta frames the library drops while it waits for an IDR show as -1
	 * reads: a kernel without ironkvm-dist patch 0912 has seven at GOP 30. */
	result(st.errors == 0, "stream start without dropped reads (%u errors)", st.errors);
	result(st.keys > 0 && st.keys_complete == st.keys,
	       "%u of %u keyframes carry SPS, PPS and IDR", st.keys_complete, st.keys);
	if (frames > 2 * gop)
		result(st.max_key_gap <= gop + 1 && st.min_key_gap + 1 >= gop,
		       "keyframe every %u frames (seen %u..%u)", gop,
		       st.min_key_gap, st.max_key_gap);
	result(st.sps_width > 0 && st.sps_height > 0, "SPS reports %ux%u",
	       st.sps_width, st.sps_height);
	{
		double nominal = st.frames ? (double)st.bytes * 8.0 / 1000.0 /
					     ((double)st.frames / fps) : 0;
		/* Informational until #35 is settled: report, fail only far off. */
		result(nominal < kbps * 3.0, "output %.0f kbit/s against %u kbit/s (fails above 3x; see #35)",
		       nominal, kbps);
	}

	/* 5. Runtime changes: half the bitrate, double the GOP. */
	{
		unsigned int gop2 = gop * 2 > 100 ? 100 : gop * 2;
		unsigned int count = gop2 * 3;

		api.set_h264_gop((uint8_t)gop2);
		run(&st, count, 0, 0, (uint16_t)(kbps / 2), 0, (uint8_t)fps,
		    count * 1000U / fps + 10000U);
		report(&st, fps);
		result(st.first_type == TYPE_KEY && st.errors == 0,
		       "set_h264_gop gives a keyframe next (type %d, %u errors)",
		       st.first_type, st.errors);
		result(st.keys >= 2 && st.max_key_gap <= gop2 + 1 && st.min_key_gap + 1 >= gop2,
		       "keyframe every %u frames after the change (seen %u..%u)", gop2,
		       st.min_key_gap, st.max_key_gap);
		api.set_h264_gop((uint8_t)gop);
	}

	/* 5b. H.264 with an MJPEG reader alongside, as a VNC viewer or a
	 * screenshot next to a web viewer would be. A software encode runs
	 * without the pipeline lock and competes for the one core; a read
	 * through the JPEG unit holds the lock for about 10 ms. Either way
	 * H.264 should keep its frames; the MJPEG rate is reported, not judged. */
	{
		struct mjpeg_load load = { 0, 0, 0 };
		pthread_t thread;
		unsigned int count = fps * 5;

		if (pthread_create(&thread, NULL, mjpeg_loop, &load) == 0) {
			run(&st, count, 0, 0, (uint16_t)kbps, 0, (uint8_t)fps,
			    count * 1000U / fps + 10000U);
			load.stop = 1;
			pthread_join(thread, NULL);
			report(&st, fps);
			note("alongside: %u MJPEG 960x540 reads (%.1f/s), %u errors",
			     load.reads, st.elapsed_us ? load.reads * 1e6 / st.elapsed_us : 0.0,
			     load.errors);
			result(st.frames == count && st.errors == 0,
			       "H.264 with an MJPEG reader: %u of %u frames, %u errors",
			       st.frames, count, st.errors);
		}
	}

	/*
	 * 5c. H.265, a switch from H.264 and back, and an MJPEG reader beside
	 * it. The two encoders share the codec SRAM, so a switch rebuilds the
	 * pipe with the other encoder; the first frame of each must be a
	 * keyframe that carries VPS, SPS and PPS.
	 */
	if (hevc) {
		struct mjpeg_load load = { 0, 0, 0 };
		pthread_t thread;
		unsigned int count = fps * 5;

		run_codec = 2;
		run(&st, frames, 0, 0, (uint16_t)kbps, 0, (uint8_t)fps,
		    frames * 1000U / fps + 10000U);
		report(&st, fps);
		result(st.frames == frames, "read %u of %u H.265 frames", st.frames, frames);
		result(st.first_type == TYPE_KEY && st.errors == 0,
		       "H.265 after H.264 starts on a keyframe (type %d, %u errors)",
		       st.first_type, st.errors);
		result(st.keys > 0 && st.keys_complete == st.keys,
		       "%u of %u H.265 keyframes carry VPS, SPS, PPS and an IRAP picture",
		       st.keys_complete, st.keys);
		if (frames > 2 * gop)
			result(st.max_key_gap <= gop + 1 && st.min_key_gap + 1 >= gop,
			       "H.265 keyframe every %u frames (seen %u..%u)", gop,
			       st.min_key_gap, st.max_key_gap);
		result(st.sps_width > 0 && st.sps_height > 0, "H.265 SPS reports %ux%u",
		       st.sps_width, st.sps_height);
		{
			double nominal = st.frames ? (double)st.bytes * 8.0 / 1000.0 /
						     ((double)st.frames / fps) : 0;

			result(nominal < kbps * 3.0, "H.265 output %.0f kbit/s against %u kbit/s",
			       nominal, kbps);
		}

		if (pthread_create(&thread, NULL, mjpeg_loop, &load) == 0) {
			run(&st, count, 0, 0, (uint16_t)kbps, 0, (uint8_t)fps,
			    count * 1000U / fps + 10000U);
			load.stop = 1;
			pthread_join(thread, NULL);
			report(&st, fps);
			note("alongside: %u MJPEG 960x540 reads (%.1f/s), %u errors",
			     load.reads, st.elapsed_us ? load.reads * 1e6 / st.elapsed_us : 0.0,
			     load.errors);
			result(st.frames == count && st.errors == 0 && load.reads > 0,
			       "H.265 with an MJPEG reader: %u of %u frames, %u errors, %u MJPEG reads",
			       st.frames, count, st.errors, load.reads);
		}

		run_codec = 1;
		run(&st, fps * 2, 0, 0, (uint16_t)kbps, 0, (uint8_t)fps, 10000U);
		report(&st, fps);
		result(st.frames == fps * 2 && st.first_type == TYPE_KEY && st.errors == 0,
		       "H.264 after H.265: %u frames, first type %d, %u errors",
		       st.frames, st.first_type, st.errors);
	}

	/* 6. Downscale through the VPSS. */
	if (!skip_scale) {
		run(&st, 60, 1280, 720, (uint16_t)kbps, 0, (uint8_t)fps, 8000);
		report(&st, fps);
		result(st.frames == 60 && st.first_type == TYPE_KEY,
		       "1280x720 stream starts on a keyframe (%u frames, first type %d)",
		       st.frames, st.first_type);
		result(st.sps_width == 1280 && st.sps_height == 720,
		       "SPS reports %ux%u for a 1280x720 request", st.sps_width,
		       st.sps_height);
	}

	/* 7. HDMI capture off and on, as the idle timer does it. */
	ret = api.kvmv_hdmi_control(0);
	ret = api.kvmv_read_video(0, 0, 1, (uint16_t)kbps, 0, 0, &data, &size);
	result(ret == IMG_NOT_EXIST && !api.kvmv_hdmi_signal_active(),
	       "capture off: read answers %d, signal %d", ret, api.kvmv_hdmi_signal_active());
	if (ret >= 0)
		api.free_kvmv_data(&data);
	api.kvmv_hdmi_control(1);
	run(&st, 30, 0, 0, (uint16_t)kbps, 0, (uint8_t)fps, 5000);
	result(st.frames == 30 && st.first_type == TYPE_KEY,
	       "capture on again: %u frames, first type %d", st.frames, st.first_type);

	/* 8. Teardown, twice over, as a server restart would. */
	api.kvmv_deinit();
	api.kvmv_init(verbose);
	api.kvmv_hdmi_control(1);
	for (i = 0, ret = -1; i < 50 && ret < 0; i++)
		ret = api.kvmv_read_video(0, 0, 1, (uint16_t)kbps, 0, 0, &data, &size);
	result(ret == TYPE_KEY, "after deinit and init the first frame is a keyframe (%d)", ret);
	if (ret >= 0)
		api.free_kvmv_data(&data);
	api.kvmv_deinit();

	if (out_file)
		fclose(out_file);
	printf("%s: %d failed\n", failures ? "FAILED" : "OK", failures);
	return failures;
}
