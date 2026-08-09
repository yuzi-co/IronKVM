/*
 * opusbench - measure what an Opus encoder costs on the NanoKVM's C906 core.
 *
 * The question this answers: the server currently spends its audio budget on a
 * 129-tap FIR plus G.711, and we want to know what full-rate Opus would cost
 * instead. So the same harness measures both, on the same signal, on the same
 * core, and prints the ratio.
 *
 * CPU time comes from CLOCK_PROCESS_CPUTIME_ID rather than the wall clock. The
 * board runs a KVM while this runs, and wall time would count somebody else's
 * work.
 */

#include <math.h>
#include <opus.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#define RATE 48000
#define FRAME 960 /* 20 ms, the same chunk the server reads from arecord */
#define SECONDS 10
#define TOTAL_FRAMES (RATE * SECONDS / FRAME)

static double cpu_seconds(void) {
    struct timespec ts;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &ts);
    return (double)ts.tv_sec + (double)ts.tv_nsec * 1e-9;
}

/* ---------------------------------------------------------------- signals */

static uint32_t rng_state = 22222u;

static double white(void) {
    rng_state = rng_state * 1664525u + 1013904223u;
    return (double)(int32_t)rng_state / 2147483648.0;
}

/*
 * Pink noise through the Paul Kellett filter. Noise is the hard case for a
 * transform codec: nothing to predict, so the encoder spends every bit it is
 * allowed and takes the slow path through the bit allocator.
 */
static double pink(void) {
    static double b0, b1, b2, b3, b4, b5, b6;
    double w = white();

    b0 = 0.99886 * b0 + w * 0.0555179;
    b1 = 0.99332 * b1 + w * 0.0750759;
    b2 = 0.96900 * b2 + w * 0.1538520;
    b3 = 0.86650 * b3 + w * 0.3104856;
    b4 = 0.55000 * b4 + w * 0.5329522;
    b5 = -0.7616 * b5 - w * 0.0168980;

    double out = b0 + b1 + b2 + b3 + b4 + b5 + b6 + w * 0.5362;
    b6 = w * 0.115926;

    return out * 0.11;
}

/*
 * A tonal signal stands for the easy end: desktop notification sounds, music
 * with a stable harmonic structure. The encoder predicts it well.
 */
static void fill_tonal(int16_t *pcm, int frames, int channels) {
    for (int i = 0; i < frames; i++) {
        double t = (double)i / RATE;
        double v = 0.30 * sin(2 * M_PI * 220.0 * t) +
                   0.20 * sin(2 * M_PI * 440.0 * t) +
                   0.12 * sin(2 * M_PI * 880.0 * t) +
                   0.06 * sin(2 * M_PI * 1760.0 * t) +
                   0.02 * pink();

        for (int c = 0; c < channels; c++) {
            /* A little decorrelation, or stereo collapses to mono coding. */
            double s = c == 0 ? v : v * 0.9 + 0.05 * sin(2 * M_PI * 331.0 * t);
            pcm[i * channels + c] = (int16_t)(s * 26000.0);
        }
    }
}

static void fill_noise(int16_t *pcm, int frames, int channels) {
    for (int i = 0; i < frames; i++) {
        for (int c = 0; c < channels; c++) {
            double s = pink();
            if (s > 1.0) s = 1.0;
            if (s < -1.0) s = -1.0;
            pcm[i * channels + c] = (int16_t)(s * 26000.0);
        }
    }
}

/* --------------------------------------------- the pipeline in use today */

/*
 * A transcription of server/service/stream/audio/resample.go and g711.go: a
 * 129-tap Hamming-windowed sinc at 3.4 kHz, decimation by 6, then mu-law. It
 * is here so the Opus numbers have something measured to sit beside rather
 * than an estimate.
 */
#define TAPS 129
#define DECIM 6

static double fir_coeffs[TAPS];
static double fir_history[TAPS];
static int fir_next, fir_phase;

static void fir_init(void) {
    double middle = (TAPS - 1) / 2.0;
    double omega = 2 * M_PI * 3400.0 / RATE;
    double sum = 0;

    for (int i = 0; i < TAPS; i++) {
        double n = (double)i - middle;
        double value = n == 0 ? omega : sin(omega * n) / n;
        double window = 0.54 - 0.46 * cos(2 * M_PI * i / (double)(TAPS - 1));

        fir_coeffs[i] = value * window;
        sum += fir_coeffs[i];
    }

    for (int i = 0; i < TAPS; i++) fir_coeffs[i] /= sum;

    memset(fir_history, 0, sizeof(fir_history));
    fir_next = fir_phase = 0;
}

static int16_t fir_output(void) {
    double sum = 0;
    int index = fir_next;

    for (int i = 0; i < TAPS; i++) {
        sum += fir_coeffs[i] * fir_history[index];
        index = (index + 1) % TAPS;
    }

    if (sum > 32767) return 32767;
    if (sum < -32768) return -32768;

    return (int16_t)sum;
}

static unsigned char ulaw(int16_t sample) {
    const int BIAS = 0x84;
    int sign = (sample >> 8) & 0x80;

    if (sign) sample = (int16_t)-sample;
    if (sample > 32635) sample = 32635;

    sample = (int16_t)(sample + BIAS);

    int exponent = 7;
    for (int mask = 0x4000; (sample & mask) == 0 && exponent > 0; exponent--, mask >>= 1) {
    }

    int mantissa = (sample >> (exponent + 3)) & 0x0F;

    return (unsigned char)~(sign | (exponent << 4) | mantissa);
}

/*
 * The checksum is not diagnostic, it is load-bearing. Without a consumer for
 * the filtered samples the compiler can see that fir_output()'s result is dead
 * and delete the whole 129-tap loop, which is the part being measured.
 */
static volatile unsigned long g711_checksum;

static void bench_g711(const int16_t *pcm, int channels) {
    unsigned char out[FRAME / DECIM + 8];
    double start, elapsed;
    long bytes = 0;
    unsigned long sum = 0;

    fir_init();
    start = cpu_seconds();

    for (int f = 0; f < TOTAL_FRAMES; f++) {
        const int16_t *in = pcm + (size_t)f * FRAME * channels;
        int n = 0;

        for (int i = 0; i < FRAME; i++) {
            double mono = channels == 2
                              ? ((double)in[i * 2] + (double)in[i * 2 + 1]) / 2.0
                              : (double)in[i];

            fir_history[fir_next] = mono;
            fir_next = (fir_next + 1) % TAPS;

            if (++fir_phase < DECIM) continue;
            fir_phase = 0;

            out[n] = ulaw(fir_output());
            sum += out[n];
            n++;
        }

        bytes += n;
    }

    elapsed = cpu_seconds() - start;
    g711_checksum = sum;

    printf("%-22s %-8s %6s %5s %8.3f %8.2f %9.1f\n", "g711-fir(today)",
           channels == 2 ? "stereo" : "mono", "-", "-", elapsed,
           100.0 * elapsed / SECONDS, (double)bytes * 8 / SECONDS / 1000);
    printf("   (checksum %lu, proving the filter ran)\n", g711_checksum);
    fflush(stdout);
}

/* ------------------------------------------------------------------ opus */

static void bench_opus(const int16_t *pcm, int channels, int bitrate,
                       int complexity, int application, const char *appname,
                       const char *signame, int with_decode) {
    int err = 0;
    OpusEncoder *enc = opus_encoder_create(RATE, channels, application, &err);
    if (err != OPUS_OK || !enc) {
        printf("encoder create failed: %s\n", opus_strerror(err));
        return;
    }

    opus_encoder_ctl(enc, OPUS_SET_BITRATE(bitrate));
    opus_encoder_ctl(enc, OPUS_SET_COMPLEXITY(complexity));

    unsigned char packet[4000];
    /* Warm the encoder so the first frames' setup is not in the measurement. */
    for (int f = 0; f < 10; f++) {
        opus_encode(enc, pcm + (size_t)f * FRAME * channels, FRAME, packet, sizeof(packet));
    }

    long bytes = 0;
    double start = cpu_seconds();

    for (int f = 0; f < TOTAL_FRAMES; f++) {
        int n = opus_encode(enc, pcm + (size_t)f * FRAME * channels, FRAME, packet,
                            sizeof(packet));
        if (n < 0) {
            printf("encode failed: %s\n", opus_strerror(n));
            opus_encoder_destroy(enc);
            return;
        }
        bytes += n;
    }

    double elapsed = cpu_seconds() - start;

    char label[64];
    snprintf(label, sizeof(label), "opus/%s/%s", appname, signame);

    char rate_label[16];
    snprintf(rate_label, sizeof(rate_label), "%dk", bitrate / 1000);

    char cx_label[16];
    snprintf(cx_label, sizeof(cx_label), "%d", complexity);

    printf("%-22s %-8s %6s %5s %8.3f %8.2f %9.1f\n", label,
           channels == 2 ? "stereo" : "mono", rate_label, cx_label, elapsed,
           100.0 * elapsed / SECONDS, (double)bytes * 8 / SECONDS / 1000);
    fflush(stdout);

    if (with_decode) {
        /* Re-encode into a held buffer so decode is measured on its own. */
        static unsigned char store[TOTAL_FRAMES][1500];
        static int sizes[TOTAL_FRAMES];

        opus_encoder_ctl(enc, OPUS_RESET_STATE);
        for (int f = 0; f < TOTAL_FRAMES; f++) {
            sizes[f] = opus_encode(enc, pcm + (size_t)f * FRAME * channels, FRAME,
                                   store[f], sizeof(store[f]));
        }

        OpusDecoder *dec = opus_decoder_create(RATE, channels, &err);
        if (err == OPUS_OK && dec) {
            int16_t out[FRAME * 2];

            start = cpu_seconds();
            for (int f = 0; f < TOTAL_FRAMES; f++) {
                opus_decode(dec, store[f], sizes[f], out, FRAME, 0);
            }
            elapsed = cpu_seconds() - start;

            snprintf(label, sizeof(label), "opus-DECODE/%s", signame);
            printf("%-22s %-8s %6s %5s %8.3f %8.2f %9s\n", label,
                   channels == 2 ? "stereo" : "mono", rate_label, cx_label, elapsed,
                   100.0 * elapsed / SECONDS, "-");
            fflush(stdout);

            opus_decoder_destroy(dec);
        }
    }

    opus_encoder_destroy(enc);
}

int main(void) {
    size_t samples = (size_t)RATE * SECONDS + FRAME * 16;
    int16_t *stereo = malloc(samples * 2 * sizeof(int16_t));
    int16_t *mono = malloc(samples * sizeof(int16_t));
    int16_t *stereo_noise = malloc(samples * 2 * sizeof(int16_t));

    if (!stereo || !mono || !stereo_noise) {
        fprintf(stderr, "out of memory\n");
        return 1;
    }

    fill_tonal(stereo, (int)samples, 2);
    fill_tonal(mono, (int)samples, 1);
    fill_noise(stereo_noise, (int)samples, 2);

    printf("libopus %s, %d s of 48 kHz audio per row, 20 ms frames\n",
           opus_get_version_string(), SECONDS);
    printf("%-22s %-8s %6s %5s %8s %8s %9s\n", "what", "chans", "rate", "cx",
           "cpu_s", "core_%", "kbit/s");
    printf("--------------------------------------------------------------------------------\n");

    bench_g711(stereo, 2);

    /* Complexity sweep at the setting a KVM would ship: stereo, 96 kbit/s. */
    int complexities[] = {0, 1, 3, 5, 8, 10};
    for (size_t i = 0; i < sizeof(complexities) / sizeof(*complexities); i++) {
        bench_opus(stereo, 2, 96000, complexities[i], OPUS_APPLICATION_AUDIO, "audio",
                   "tone", 0);
    }

    /* The same sweep on noise, which is the expensive signal. */
    for (size_t i = 0; i < sizeof(complexities) / sizeof(*complexities); i++) {
        bench_opus(stereo_noise, 2, 96000, complexities[i], OPUS_APPLICATION_AUDIO,
                   "audio", "pink", 0);
    }

    /* Mono halves the channel work. Does it halve the cost? */
    bench_opus(mono, 1, 64000, 0, OPUS_APPLICATION_AUDIO, "audio", "tone", 0);
    bench_opus(mono, 1, 64000, 3, OPUS_APPLICATION_AUDIO, "audio", "tone", 0);
    bench_opus(mono, 1, 64000, 5, OPUS_APPLICATION_AUDIO, "audio", "tone", 0);
    bench_opus(mono, 1, 64000, 10, OPUS_APPLICATION_AUDIO, "audio", "tone", 0);

    /* Bitrate barely moves the encoder; confirm rather than assume. */
    bench_opus(stereo, 2, 32000, 5, OPUS_APPLICATION_AUDIO, "audio", "tone", 0);
    bench_opus(stereo, 2, 64000, 5, OPUS_APPLICATION_AUDIO, "audio", "tone", 0);
    bench_opus(stereo, 2, 128000, 5, OPUS_APPLICATION_AUDIO, "audio", "tone", 0);

    /* VOIP picks SILK more often, which is a different cost curve. */
    bench_opus(stereo, 2, 96000, 5, OPUS_APPLICATION_VOIP, "voip", "tone", 0);
    bench_opus(mono, 1, 32000, 5, OPUS_APPLICATION_VOIP, "voip", "tone", 0);

    /* Decode is the microphone direction. */
    bench_opus(stereo, 2, 96000, 5, OPUS_APPLICATION_AUDIO, "audio", "tone", 1);

    free(stereo);
    free(mono);
    free(stereo_noise);

    return 0;
}
