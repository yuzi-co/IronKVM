/* Measure what libkvm gives back to the ION carveout when capture stops.
 *
 *   venc-leak [--mode channel|pipeline] [--frames N] [--abort]
 *
 * The teardown fault this tool exists for is not the one that was fixed on
 * 2026-08-20. That one stopped two threads that never saw their exit flag, and
 * a stop measured against an idle pipeline now returns everything it took. An
 * idle pipeline holds no encoder, so that measurement said nothing about the
 * encoder, and a restart taken while a browser was streaming still orphans
 * 11,636,736 bytes: two extra VENC_1_ReconFrameBuf, and one extra each of
 * VCODEC_H264_FW_Buffer, VENC_1_BitStreamBuffer, VENC_1_H264_WorkBuffer,
 * ISP_SHARED_BUFFER_0 and VbPool.
 *
 * Nothing in the kernel takes those buffers back. osdrv/interdrv/v2/vcodec
 * builds with -DCVI_H26X_USE_ION_FW_BUFFER, which compiles vpu_free_buffers()
 * out of vpu_release(), and soph_sys releases bindings on close but never walks
 * its buffer list. A process that exits without destroying its encoder leaves
 * that memory allocated until the board reboots. So the question is only ever
 * "does the owner ask", and this program asks it directly.
 *
 * Two modes, because they need different things from the board:
 *
 *   channel   mmf_init, add an H.264 channel, delete it, mmf_try_deinit.
 *             This is the encoder on its own. It needs no HDMI signal, so it
 *             runs whether or not a host is awake.
 *
 *             --push N encodes N synthetic frames first. Creating a channel
 *             allocates the work, bitstream and firmware buffers, but the VPU
 *             opens its instance on the first frame, and that is what allocates
 *             VENC_1_ReconFrameBuf. A channel that never encoded therefore
 *             proves nothing about the pair of buffers that dominate the leak.
 *
 *   pipeline  kvmv_init, read frames, kvmv_deinit. This is the path the server
 *             takes. It needs a live HDMI signal: with no frames arriving,
 *             kvmv_read_img never returns, and no encoder is ever created.
 *
 * The report prints the carveout at each step, with every buffer named, so a
 * step that adds a buffer and a later step that fails to remove it are both
 * visible rather than inferred from a total.
 *
 * --abort leaves through _exit instead of tearing down, which is what the
 * server does when it is killed rather than asked to stop. It costs a
 * generation every time. Run it deliberately and read the next baseline.
 *
 * The video hardware takes one owner, and the supervisor restarts a server that
 * goes away. Stop both before running this:
 *
 *     /etc/init.d/S98supervise stop
 *     /etc/init.d/S95nanokvm stop
 *     /tmp/vlk/venc-leak --mode channel
 *     /etc/init.d/S95nanokvm start
 *     /etc/init.d/S98supervise start
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "kvm_vision.h"
#include "kvm_mmf.hpp"

#define ION_SUMMARY "/sys/kernel/debug/ion/cvi_carveout_heap_dump/summary"

#define VENC_H264 1

/* The maix format id for FMT_YVU420SP. kvm_vision.cpp passes the enumerator;
 * this passes the number, so the tool needs no MaixCDK header to build.
 */
#define MAIX_FMT_YVU420SP 8

/* The capture resolution the server runs at, and the LT6911's ceiling. */
#define VENC_WIDTH  1920
#define VENC_HEIGHT 1080

/* Print the carveout under a label. The summary names every buffer, so the
 * report shows which ones a step added and which ones the next step failed to
 * take away. A carveout that cannot be read is reported rather than skipped:
 * this whole program is the reading.
 */
static void report(const char *label)
{
	FILE *fp;
	char line[256];

	printf("\n########## %s\n", label);
	fflush(stdout);

	fp = fopen(ION_SUMMARY, "r");
	if (fp == NULL) {
		printf("  cannot read %s\n", ION_SUMMARY);
		fflush(stdout);
		return;
	}

	while (fgets(line, sizeof(line), fp) != NULL) {
		/* The free-region table below the buffers measures fragmentation
		 * rather than usage, and it is long. Stop at it.
		 */
		if (strstr(line, "minimum ion allocate unit") != NULL)
			break;
		if (line[0] == '\n')
			continue;
		printf("  %s", line);
	}

	fclose(fp);
	fflush(stdout);
}

/* Create an H.264 channel and delete it again, with no camera involved.
 *
 * These are the same calls kvm_vision makes from init_venc_h264, with the same
 * configuration. mmf_add_venc_channel is what makes the vc driver allocate
 * VENC_1_ReconFrameBuf and the rest; mmf_del_venc_channel is the only thing
 * that asks for them back.
 */
static int run_channel(int abort_without_deinit, int push_frames, int chn,
		       int with_jpeg, int refcount_deinits, int server_stop)
{
	mmf_venc_cfg_t cfg;
	int ret;
	int i;

	report("baseline, nothing running");

	if (server_stop) {
		/* Start and stop the way the server does, so the sequence under
		 * test is the one that leaked: kvmv_init, both encoders alive at
		 * once, kvmv_deinit. The synthetic frames below stand in for the
		 * camera, and the encoder cannot tell the difference.
		 */
		kvmv_init(0);
		report("after kvmv_init");
	} else if (mmf_init() != 0) {
		printf("  mmf_init failed\n");
		return 1;
	} else {
		report("after mmf_init");
	}

	if (with_jpeg) {
		/* MaixCDK's Image::to_jpeg reaches mmf_enc_jpg_init, so the MJPEG
		 * delivery path takes this reference whether or not the server
		 * knows it. Nothing gives it back: mmf_enc_jpg_deinit runs from
		 * _mmf_deinit, and from the mode switch in kvm_vision only when
		 * venc_auto_recyc is set, which nothing sets.
		 */
		ret = mmf_enc_jpg_init(0, VENC_WIDTH, VENC_HEIGHT,
				       mmf_invert_format_to_mmf(MAIX_FMT_YVU420SP),
				       80);
		printf("  mmf_enc_jpg_init returned %d\n", ret);
		report("after mmf_enc_jpg_init");
	}

	memset(&cfg, 0, sizeof(cfg));
	cfg.type = 2; /* 1, h265; 2, h264 */
	cfg.w = VENC_WIDTH;
	cfg.h = VENC_HEIGHT;
	cfg.fmt = mmf_invert_format_to_mmf(MAIX_FMT_YVU420SP);
	cfg.jpg_quality = 0;
	cfg.gop = 30;
	cfg.intput_fps = 60;
	cfg.output_fps = 60;
	cfg.bitrate = 3000;

	ret = mmf_add_venc_channel(chn, &cfg);
	printf("  mmf_add_venc_channel(%d) returned %d\n", chn, ret);
	report("after mmf_add_venc_channel");

	if (push_frames > 0) {
		/* NV21 is one luma plane and one interleaved chroma plane at half
		 * the height, so the frame is w * h * 3 / 2. The content does not
		 * matter here: the encoder allocates the same buffers for a grey
		 * frame as for a desktop.
		 */
		size_t len = (size_t)cfg.w * cfg.h * 3 / 2;
		uint8_t *frame = (uint8_t *)malloc(len);
		int pushed = 0, popped = 0;
		int i;

		if (frame == NULL) {
			printf("  cannot allocate a %zu byte frame\n", len);
			return 1;
		}
		memset(frame, 0x80, len);

		for (i = 0; i < push_frames; i++) {
			mmf_stream_t stream;

			if (mmf_venc_push(chn, frame, cfg.w, cfg.h, cfg.fmt) == 0)
				pushed++;
			if (mmf_venc_pop(chn, &stream) == 0) {
				popped++;
				mmf_venc_free(chn);
			}
		}
		free(frame);
		printf("  pushed %d, popped %d of %d frame(s)\n",
		       pushed, popped, push_frames);
		report("after encoding");
	}

	if (abort_without_deinit) {
		printf("\n########## leaving without any teardown\n");
		fflush(stdout);
		_exit(0);
	}

	if (server_stop) {
		/* No mmf_del_venc_channel by hand. Whether the encoder comes back
		 * is exactly what kvmv_deinit is being asked here.
		 */
		kvmv_deinit();
		report("after kvmv_deinit");
		return 0;
	}

	if (refcount_deinits > 0) {
		/* Do not destroy the channel by hand. mmf_deinit is refcounted,
		 * and _mmf_deinit - the only caller of mmf_del_venc_channel_all
		 * - runs on the call that takes the count to zero and on no
		 * other. Counting the calls it takes is the measurement.
		 */
		for (i = 0; i < refcount_deinits; i++) {
			char label[64];

			mmf_deinit();
			snprintf(label, sizeof(label),
				 "after mmf_deinit #%d", i + 1);
			report(label);
		}
		return 0;
	}

	ret = mmf_del_venc_channel(chn);
	printf("  mmf_del_venc_channel(%d) returned %d\n", chn, ret);
	report("after mmf_del_venc_channel");

	/* force, because this process holds the only reference and a refcounted
	 * deinit that decrements to one would tear nothing down and say nothing.
	 */
	mmf_try_deinit(true);
	report("after mmf_try_deinit(force)");

	return 0;
}

/* Ask for frames until enough of them arrive.
 *
 * kvmv_read_img returns a negative code while the HDMI input is still being
 * negotiated. Those are not failures, so the loop counts successes rather than
 * stopping at the first refusal. It cannot bound its own runtime: with no
 * frames at all the call does not return, because the library loops inside.
 * That is why --mode channel exists.
 */
static int pump(int type, int frames, int width, int height, int quality)
{
	int got = 0;
	int tries = 0;
	int last = 0;

	while (got < frames && tries < frames * 40 + 200) {
		uint8_t *data = NULL;
		uint32_t size = 0;
		int ret;

		tries++;
		ret = kvmv_read_img((uint16_t)width, (uint16_t)height,
				    (uint8_t)type, (uint16_t)quality,
				    &data, &size);
		last = ret;
		if (ret >= 0 && data != NULL) {
			got++;
			free_kvmv_data(&data);
		} else {
			usleep(20000);
		}
	}

	printf("  %d frame(s) in %d call(s), last return %d\n",
	       got, tries, last);
	fflush(stdout);
	return got;
}

static int run_pipeline(int frames, int abort_without_deinit, int cycles)
{
	report("baseline, nothing running");

	/* kvmv_deinit forces the MMF refcount to zero, so a later kvmv_init has
	 * to bring the pipeline back from nothing. The idle timeout stops
	 * capture at runtime, not only at exit, so a start that did not work
	 * would leave the board with a black screen until it was restarted.
	 * Cycling with no frames tests exactly that and needs no HDMI signal.
	 */
	if (cycles > 0) {
		int c;

		for (c = 0; c < cycles; c++) {
			char label[64];

			kvmv_init(0);
			snprintf(label, sizeof(label), "cycle %d, after kvmv_init", c + 1);
			report(label);

			kvmv_deinit();
			snprintf(label, sizeof(label), "cycle %d, after kvmv_deinit", c + 1);
			report(label);
		}
		return 0;
	}

	kvmv_init(0);
	report("after kvmv_init");

	pump(VENC_H264, frames, 1920, 1080, 3000);
	report("after h264 frames");

	if (abort_without_deinit) {
		printf("\n########## leaving without kvmv_deinit\n");
		fflush(stdout);
		_exit(0);
	}

	kvmv_deinit();
	report("after kvmv_deinit");

	return 0;
}

int main(int argc, char **argv)
{
	int channel_mode = 1;
	int frames = 60;
	int push_frames = 0;
	/* The server encodes on channel 1, and the leaked buffers are named
	 * VENC_1_*. Match it, so a report from this tool and a report from a
	 * live board name the same things.
	 */
	int chn = 1;
	int with_jpeg = 0;
	int refcount_deinits = 0;
	int cycles = 0;
	int server_stop = 0;
	int abort_without_deinit = 0;
	int i;

	for (i = 1; i < argc; i++) {
		if (strcmp(argv[i], "--mode") == 0 && i + 1 < argc) {
			i++;
			if (strcmp(argv[i], "channel") == 0) {
				channel_mode = 1;
			} else if (strcmp(argv[i], "pipeline") == 0) {
				channel_mode = 0;
			} else {
				printf("unknown mode: %s\n", argv[i]);
				return 2;
			}
		} else if (strcmp(argv[i], "--frames") == 0 && i + 1 < argc) {
			frames = atoi(argv[++i]);
		} else if (strcmp(argv[i], "--push") == 0 && i + 1 < argc) {
			push_frames = atoi(argv[++i]);
		} else if (strcmp(argv[i], "--chn") == 0 && i + 1 < argc) {
			chn = atoi(argv[++i]);
		} else if (strcmp(argv[i], "--jpg") == 0) {
			with_jpeg = 1;
		} else if (strcmp(argv[i], "--deinits") == 0 && i + 1 < argc) {
			refcount_deinits = atoi(argv[++i]);
		} else if (strcmp(argv[i], "--cycles") == 0 && i + 1 < argc) {
			cycles = atoi(argv[++i]);
		} else if (strcmp(argv[i], "--server-stop") == 0) {
			server_stop = 1;
		} else if (strcmp(argv[i], "--abort") == 0) {
			abort_without_deinit = 1;
		} else {
			printf("usage: venc-leak [--mode channel|pipeline] "
			       "[--push N] [--chn N] [--frames N] [--abort]\n");
			return 2;
		}
	}

	/* libkvm writes its own progress with printf, and stdout here is a pipe
	 * or a file. Without this the report interleaves with the library by
	 * buffer boundary rather than by line, and the steps stop lining up.
	 */
	setvbuf(stdout, NULL, _IOLBF, 0);

	if (channel_mode)
		return run_channel(abort_without_deinit, push_frames, chn,
				   with_jpeg, refcount_deinits, server_stop);
	return run_pipeline(frames, abort_without_deinit, cycles);
}
