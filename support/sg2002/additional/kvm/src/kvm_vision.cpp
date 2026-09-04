/**
 * 待解决的问题:
 * // 分辨率跟随输出
 * // 自动切换分辨率
 * // HDMI分辨率问题
 * // H264输入空图,VENC容易炸问题
 * // MJPEG/H264切换时卡退
 * // 再加一条手动清空全部内存
 * // deinit时free全部内存
 * // free 错内存时会炸的问题
 */
#include "kvm_vision.h"
#include "vi_state_shared.hpp"
#include "internal/vi_state_writer.hpp"

#include <errno.h>
#include <sys/stat.h>

#define default_venc_chn        1

#define VENC_MJPEG              0
#define VENC_H264               1

#define KVMV_MAX_TRY_NUM	   	2
#define vi_min_width            32
#define vi_min_height           3
#define vi_max_width            1920
#define vi_max_height           1080

#define Farame_sample_size      76800   // 320*240

#define default_vi_width        1920
#define default_vi_height       1080
#define default_vpss_width      1920
#define default_vpss_height     1080
#define default_venc_type       VENC_MJPEG
#define default_mjpeg_qlty      60
#define default_h264_qlty       1000
#define default_h264_gop        30
// What the encoder is told until the server says otherwise. The server has
// always clamped its own setting to the same 10 to 60, and set_h264_fps
// clamps again so a caller that skips the server cannot get past it.
#define default_h264_fps        60
#define fresh_frame_discard_count 5

#define kvmv_data_buffer_size   4
#define Try_rounds_HDMI_err_res 5

#define vi_width_path           "/kvmapp/kvm/width"
#define vi_height_path          "/kvmapp/kvm/height"
#define hdmi_mode_path          "/etc/kvm/hdmi_mode"
#define hdmi_state_path         "/proc/lt_int"
// tmpfs, not the boot medium: this file is rewritten on every change of HDMI
// presence, and the value means nothing after a reboot. S95nanokvm leaves a
// symlink at /kvmapp/kvm/state for the readers that know the old path. The
// publish below renames a temporary file over this one, and a rename replaces
// a symlink rather than following it, so both paths have to name tmpfs.
#define hdmi_signal_file_path   "/tmp/kvm/state"
#define watchdog_mode_path      "/etc/kvm/watchdog"
#define watchdog_temp_path      "/tmp/watchdog"
#define watchdog_file           "/tmp/nanokvm_wd"
#define vi_state_publish_interval_ms 10000U
// The detection loop polls /proc/lt_int for HDMI edge counts. It has to be
// quick while a resolution is being chosen, and it has nothing to do once the
// input is settled. See the sleep at the end of vi_subsystem_detection.
#define vi_detection_active_poll_ms 10U
#define vi_detection_idle_poll_ms 100U

// The VPSS channel can stop delivering frames while the VI device keeps
// running. Measured on 2026-09-04: VIDevFPS held at about 50 while the
// channel's SendOK stayed frozen and its FrameRate read 0 for 27 minutes,
// with the kernel reporting "CVI_VPSS_GetChnFrame fail" and
// "jobs wait(0) work(0) done(0)" once a second for as long as something kept
// reading. The board recovered only when the HDMI source changed resolution,
// because that sets reopen_cam_flag and check_kvmv then restarts the camera.
// Nothing else recovers it, so a viewer that loses the channel sees "no HDMI"
// until the process is restarted.
//
// The counters below ask for that same restart on the board's own initiative.
// They are deliberately slow. Reopening the MMF channel is the operation that
// can exhaust the carveout heap, which is the reason fresh_frame_discard_count
// exists to avoid it after an HDMI idle, so this must not become something the
// board does often.
//
// A run of failures has to be long enough to rule out the ordinary case, which
// is a source that stopped sending. Measured in the same fault: the reader
// failed about once a second, so 30 failures is roughly half a minute of a
// channel that hands out nothing.
#define wedge_fail_threshold        30U
#define wedge_min_run_ms          2000U
#define wedge_first_cooldown_ms  10000U
#define wedge_max_cooldown_ms    60000U

#define LT6911_ADDR 	0x2B
#define LT6911_READ 	0xFF
#define LT6911_WRITE 	0x00

pthread_mutex_t vi_mutex;
pthread_mutex_t hdmi_signal_mutex = PTHREAD_MUTEX_INITIALIZER;

static char NanoKVM_edit[] = {
	0x00,0xFF,0xFF,0xFF,0xFF,0xFF,0xFF,0x00,0x41,0x0C,0x33,0xC2,0x66,0xBA,0x00,0x00,
	0x2B,0x1F,0x01,0x04,0xA5,0x50,0x22,0x78,0x3B,0xCC,0xE5,0xAB,0x51,0x48,0xA6,0x26,
	0x0C,0x50,0x54,0xBF,0xEF,0x00,0xD1,0xC0,0xB3,0x00,0x95,0x00,0x81,0x80,0x81,0x40,
	0x81,0xC0,0x01,0x01,0x01,0x01,0xD8,0x59,0x00,0x60,0xA3,0x38,0x28,0x40,0xA0,0x10,
	0x3A,0x10,0x20,0x4F,0x31,0x00,0x00,0x1A,0x00,0x00,0x00,0xFF,0x00,0x55,0x4B,0x30,
	0x32,0x31,0x34,0x33,0x30,0x34,0x37,0x37,0x31,0x38,0x00,0x00,0x00,0xFC,0x00,0x50,
	0x48,0x4C,0x20,0x33,0x34,0x32,0x45,0x32,0x0A,0x20,0x20,0x20,0x00,0x00,0x00,0xFD,
	0x00,0x30,0x4B,0x63,0x63,0x1E,0x01,0x0A,0x20,0x20,0x20,0x20,0x20,0x20,0x01,0x8E
};

using namespace maix;
using namespace maix::sys;
using namespace maix::peripheral;
i2c::I2C LT6911_i2c(4, i2c::Mode::MASTER);

typedef struct {
	uint16_t vi_width = default_vi_width;
	uint16_t vi_height = default_vi_height;
	uint16_t vpss_width = default_vpss_width;
	uint16_t vpss_height = default_vpss_height;
	uint8_t venc_type;
	uint16_t qlty;
	uint8_t stream_stop;
	uint8_t frame_detact = 0;
	uint8_t display;
    uint8_t reinit_flag = 1;
    uint8_t reopen_cam_flag = 0;
    uint8_t hdmi_cable_state = 0;
    // Read and written through __atomic_* everywhere, and it has to stay that
    // way. kvmv_deinit sets this and then joins the two threads that watch it.
    // As a plain field it is an object no thread loop can be seen to write, so
    // the compiler is entitled to read it once and keep the answer: in the
    // library shipped before 2026-08-20, GCC hoisted the test out of both loops
    // and left no load of it anywhere inside either. Neither thread could ever
    // observe the request, both joins waited for good, and kvmv_deinit never
    // reached cam->close() or mmf_deinit(). The process was then killed by its
    // own timeout with the VI channel pool and the ISP shared buffer it had
    // allocated still held: 6,516,736 bytes of the carveout per server
    // generation, which only a reboot gives back.
    uint8_t try_exit_thread = 0;
    uint8_t thread_is_running = 0;
    uint8_t Auto_res = 0;
    uint8_t hdmi_version = 0;
    uint8_t hw_version = 0;
    uint8_t hdmi_stop_flag = 0;
    uint8_t hdmi_reading_flag = 0;
    uint8_t hdmi_mode = 0;
    uint8_t hdmi_res_type = 0;
    uint8_t hdmi_res_err = 0;
    uint8_t hdmi_try_rounds = 0;
    uint8_t vi_detect_state = 0;
    uint8_t venc_auto_recyc = 0;
    uint8_t fresh_frame_count = 0;
    // Handles for the two threads kvmv_init starts, so kvmv_deinit can join
    // them. Both used to be written into one local in kvmv_init - the second
    // pthread_create overwrote the first - so neither could be waited for. That
    // left kvmv_deinit destroying vi_mutex while a thread was still using it,
    // and left kvmv_init reading thread_is_running while the old thread was
    // deciding whether to exit. Both threads break on try_exit_thread, and the
    // slower poll is 500ms, so the joins are bounded.
    pthread_t detect_thread;
    pthread_t watchdog_thread;
    uint8_t detect_thread_valid = 0;
    uint8_t watchdog_thread_valid = 0;
} kvmv_cfg_t;

typedef struct {
	uint8_t* p_img_data = NULL;
    uint8_t img_data_type = 0;
    uint32_t img_data_size = 0;
    // The allocation behind p_img_data outlives the frame in it. Capacity
    // is what was allocated, img_data_size is what the current frame uses,
    // and in_use says the slot is taken.
    uint32_t img_data_capacity = 0;
    uint8_t in_use = 0;
} kvmv_data_t;

typedef struct {
    uint8_t mmf_venc_chn;
	uint8_t enc_h264_running;
	uint8_t enc_h264_init;
    mmf_venc_cfg_t kvm_venc_cfg;
} kvm_venc_t;

camera::Camera *cam = new camera::Camera(default_vpss_width, default_vpss_height, image::FMT_YVU420SP);
// camera::Camera *cam = new camera::Camera(320, 240, image::FMT_RGB888);

kvmv_cfg_t kvmv_cfg;

kvmv_data_t kvmv_data_buffer[kvmv_data_buffer_size];

uint8_t kvmv_data_buffer_index = 0;

uint8_t debug_en = 0;
uint8_t last_vi_state_code = 0;
uint32_t last_vi_state_refresh_ms = 0;
uint8_t hdmi_capture_enabled = 0;
uint8_t hdmi_signal_active = 0;

// Set by the reader when it decides the channel is wedged, and consumed by
// vi_subsystem_detection. The restart happens on that thread because every
// other cam->restart in this library already happens there or under
// reopen_cam_flag, and because vi_mutex does not serialise the two: the
// detection thread has never taken that lock.
uint8_t vi_wedge_rebuild_request = 0;

// How the reader decides. Split out from kvmv_read_img so it can be run
// off-device: everything it needs is an argument, and it touches no hardware.
//
// "A frame" means the channel handed one out, whatever happened to it
// afterwards. An encode that fails still proves the channel is alive, which
// is the only thing these counters are about.
typedef struct {
    uint32_t fail_count;
    uint32_t first_fail_ms;
    uint32_t last_rebuild_ms;
    uint32_t cooldown_ms;
    uint8_t  rebuilt_before;
} wedge_state_t;

void wedge_note_frame(wedge_state_t *w)
{
    if(w == NULL){
        return;
    }
    w->fail_count = 0;
    w->cooldown_ms = wedge_first_cooldown_ms;
}

uint8_t wedge_note_no_frame(wedge_state_t *w, uint32_t now_ms, uint8_t vi_live)
{
    if(w == NULL){
        return 0;
    }
    // A source that stopped sending is not a wedged channel, and restarting
    // the camera cannot bring a signal back. The detection thread owns that
    // case, and it is by far the common one on a desk where the host sleeps.
    //
    // Start the run again rather than counting these. A client that keeps
    // asking across a long outage drives the counter into the hundreds, and
    // counting them would mean the first failure after the signal returned
    // rebuilt the channel on the strength of evidence gathered while there
    // was no signal at all. Every recovery seen on 2026-09-04 produced three
    // or four such failures, so that would have fired on all of them.
    if(vi_live == 0){
        w->fail_count = 0;
        return 0;
    }

    if(w->fail_count == 0){
        w->first_fail_ms = now_ms;
    }
    if(w->fail_count < 0xFFFFFFFFU){
        w->fail_count++;
    }
    if(w->fail_count < wedge_fail_threshold){
        return 0;
    }
    // Unsigned throughout, so the monotonic clock wrapping every 49 days is
    // not a special case here.
    if(now_ms - w->first_fail_ms < wedge_min_run_ms){
        return 0;
    }

    uint32_t cooldown = w->cooldown_ms == 0 ? wedge_first_cooldown_ms : w->cooldown_ms;
    if(w->rebuilt_before != 0 && now_ms - w->last_rebuild_ms < cooldown){
        return 0;
    }

    w->last_rebuild_ms = now_ms;
    if(w->rebuilt_before == 0){
        // The first retry comes quickly, because one restart often is enough
        // and a viewer is waiting.
        w->rebuilt_before = 1;
        w->cooldown_ms = wedge_first_cooldown_ms;
    } else {
        // Then back off, so a channel that cannot be recovered is not
        // reopened every ten seconds for as long as the board is up.
        w->cooldown_ms = cooldown > wedge_max_cooldown_ms / 2
            ? wedge_max_cooldown_ms
            : cooldown * 2;
    }
    return 1;
}

wedge_state_t vi_wedge = { 0, 0, 0, wedge_first_cooldown_ms, 0 };

void debug(const char *format, ...);

static bool write_hdmi_signal_file(uint8_t active)
{
    char temp_path[] = "/tmp/kvm/.state.tmp.XXXXXX";
    const char *state = active != 0 ? "1\n" : "0\n";
    const size_t state_size = 2;
    int fd = mkstemp(temp_path);
    if(fd < 0){
        debug("[hdmi] failed to create state file: %d\n", errno);
        return false;
    }

    int failure = 0;
    if(fchmod(fd, 0644) != 0){
        failure = errno;
    }

    size_t written = 0;
    while(failure == 0 && written < state_size){
        ssize_t result = write(fd, state + written, state_size - written);
        if(result > 0){
            written += (size_t)result;
        } else if(result < 0 && errno == EINTR){
            continue;
        } else {
            failure = result < 0 ? errno : EIO;
        }
    }

    if(failure == 0 && fsync(fd) != 0){
        failure = errno;
    }
    if(close(fd) != 0 && failure == 0){
        failure = errno;
    }
    if(failure == 0 && rename(temp_path, hdmi_signal_file_path) != 0){
        failure = errno;
    }
    if(failure != 0){
        unlink(temp_path);
        debug("[hdmi] failed to publish state file: %d\n", failure);
        return false;
    }
    return true;
}

static void set_hdmi_signal_state(uint8_t active)
{
    pthread_mutex_lock(&hdmi_signal_mutex);
    uint8_t enabled = __atomic_load_n(&hdmi_capture_enabled, __ATOMIC_ACQUIRE);
    uint8_t next = enabled != 0 && active != 0 ? 1 : 0;
    uint8_t previous = __atomic_exchange_n(&hdmi_signal_active, next, __ATOMIC_ACQ_REL);
    if(previous != next){
        write_hdmi_signal_file(next);
    }
    pthread_mutex_unlock(&hdmi_signal_mutex);
}

static void set_hdmi_capture_enabled(uint8_t enabled)
{
    pthread_mutex_lock(&hdmi_signal_mutex);
    __atomic_store_n(&hdmi_capture_enabled, enabled != 0 ? 1 : 0, __ATOMIC_RELEASE);
    __atomic_store_n(&hdmi_signal_active, 0, __ATOMIC_RELEASE);
    write_hdmi_signal_file(0);
    pthread_mutex_unlock(&hdmi_signal_mutex);
}

static void set_hdmi_detection_state(uint8_t active)
{
    __atomic_store_n(&kvmv_cfg.hdmi_cable_state, active != 0 ? 1 : 0, __ATOMIC_RELEASE);
    set_hdmi_signal_state(active);
}

void debug(const char *format, ...)
{
    if(debug_en){
        printf(format);
    }
}

uint8_t refresh_vi_state()
{
	// The driver read blocks for about a second; mark the deadline before it so
	// that the publication cadence measures wall-clock time, not read time plus it.
    last_vi_state_refresh_ms = vi_state_shared::monotonic_ms();
    last_vi_state_code = vi_state_shared::refresh();
    set_hdmi_detection_state(last_vi_state_code == 1 || last_vi_state_code >= 3);
    return last_vi_state_code;
}

uint8_t to_roll(int8_t _input)
{
    if(_input <                      0) return _input + kvmv_data_buffer_size;
    if(_input >= kvmv_data_buffer_size) return _input - kvmv_data_buffer_size;
    return _input;
}

int maxmin_data(int _max, int _min, int _data)
{
	if(_data > _max) return _max;
	if(_data < _min) return _min;
	return _data;
}

// Take the next free slot. The version this replaces looked at one slot,
// the one after the last, and gave up if that slot was busy: three free
// buffers next to it were not considered and the caller was told the ring
// was full. Walk the whole ring instead, starting where the old one did so
// the buffers still rotate rather than always reusing slot 0.
kvmv_data_t* get_save_buffer()
{
    for(int i = 0; i < kvmv_data_buffer_size; i++){
        kvmv_data_buffer_index = to_roll(kvmv_data_buffer_index + 1);
        kvmv_data_t* buffer = &kvmv_data_buffer[kvmv_data_buffer_index];
        if(buffer->in_use == 0){
            buffer->in_use = 1;
            buffer->img_data_type = 0;
            buffer->img_data_size = 0;
            return buffer;
        }
    }
    return NULL;
}

// Hand a slot back without freeing what it holds. Every path that takes a
// buffer and then fails has to reach this, or the slot is lost for the life
// of the process.
static void release_save_buffer(kvmv_data_t* buffer)
{
    if(buffer != NULL){
        buffer->in_use = 0;
    }
}

// Make sure the slot can hold size bytes, reusing the allocation it already
// has whenever that one is big enough.
//
// A frame is one or two hundred kilobytes at 1080p, far above the size at
// which musl serves malloc from mmap. The code this replaces allocated and
// freed one of those per frame, so every frame paid an mmap, a fault on
// each of its forty-odd pages, and a munmap, thirty times a second on the
// one core this board has.
//
// What it costs to keep them is bounded: four slots, each at the largest
// frame it has held. Capacity never shrinks, so a board that ran at 1080p
// keeps 1080p sized buffers after a change to 720p until the process ends.
static bool reserve_save_buffer(kvmv_data_t* buffer, uint32_t size)
{
    if(buffer == NULL || size == 0){
        return false;
    }
    if(buffer->p_img_data != NULL && buffer->img_data_capacity >= size){
        return true;
    }

    uint8_t* data = (uint8_t *)realloc(buffer->p_img_data, size);
    if(data == NULL){
        return false;
    }
    buffer->p_img_data = data;
    buffer->img_data_capacity = size;
    return true;
}

// ====HDMI RES==================================================

uint16_t hdmi_res_list[][2] = {
    {1920, 1080},
    {1600, 900},
    {1440, 1080},
    {1440, 900},
    {1280, 1024},
    {1280, 960},
    {1280, 800},
    {1280, 720},
    {1152, 864},
    {1024, 768},
    {800, 600},
	{640, 480},
};

uint16_t hdmi_unsupported_res_list[][2] = {
    {1680, 1050},
    {1440, 1050},
    {1400, 1050},
    {1368, 768},
    {1366, 768},
    {720, 576},
};

/* return 0 : normal res;
/* return 1 : new res;
 * return 2 : unsupport res;
 * return 3 : unknow res;
 */
uint8_t check_res(uint16_t _width, uint16_t _height)
{
    uint8_t i;
    uint8_t ret;

    for(i = 0; i < sizeof(hdmi_res_list)/4; i++){
        if(_width == hdmi_res_list[i][0] && _height == hdmi_res_list[i][1]) return NORMAL_RES;
    }
    for(i = 0; i < sizeof(hdmi_unsupported_res_list)/4; i++){
        if(_width == hdmi_unsupported_res_list[i][0] && _height == hdmi_unsupported_res_list[i][1]) return UNSUPPORT_RES;
    }
    return UNKNOWN_RES;
}

void write_res_to_file(uint16_t _width, uint16_t _height)
{
	char Cmd[100]={0};
    sprintf(Cmd, "echo %d > %s", _width, vi_width_path);
    system(Cmd);
    sprintf(Cmd, "echo %d > %s", _height, vi_height_path);
    system(Cmd);
    system("sync");
}

int set_hdmi_mode(uint8_t _hdmi_mode)
{
    if(_hdmi_mode >= 0 && _hdmi_mode <= 2){
        char Cmd[100]={0};
        sprintf(Cmd, "echo %d > %s", _hdmi_mode, hdmi_mode_path);
        system(Cmd);
        return 1;
    } else {
        debug("[kvmv] Incorrect HDMI mode.\n");
        return 0;
    }
}

int get_hdmi_mode(void)
{
    if(access(hdmi_mode_path, F_OK) == 0){
        // exist
        FILE *fp;
        uint8_t tmp8;
        uint8_t RW_Data[3] = {0};

        fp = fopen(hdmi_mode_path, "r");
        if (fp == NULL) {
            kvmv_cfg.hdmi_mode = 0;
            return 0;
        }
        fread(RW_Data, sizeof(char), sizeof(RW_Data) - 1, fp);
        fclose(fp);
        tmp8 = atoi((char*)RW_Data);
        if(tmp8 > 2) {
            tmp8 = 0;
	        char Cmd[100]={0};
            sprintf(Cmd, "echo 0 > %s", hdmi_mode_path);
            system(Cmd);
        }
        if(tmp8 != kvmv_cfg.hdmi_mode){
            kvmv_cfg.hdmi_mode = tmp8;
            debug("[kvmv] hdmi mode = %d\n", kvmv_cfg.hdmi_mode);
            return 1;
        } else {
            return 0;
        }
    }
    kvmv_cfg.hdmi_mode = 0;
    return 0;
}

uint8_t watchdog_sf_is_open(void)
{
	if(access(watchdog_mode_path, F_OK) == 0) return 1;
    else if(access(watchdog_temp_path, F_OK) == 0) return 1;
	else return 0;
}

int vision_update_watchdog()
{
    FILE *file;

    file = fopen(watchdog_file, "w");
    if (file == NULL) {
        debug("[kvmv] watchdog open error\n");
        return -1;
    }
    // fprintf(file, "%s", 'v');
    fclose(file);
    return 1;
}

int get_manual_resolution(void)
{
    uint8_t RW_Data[35];
    FILE *fp;
    int file_size;
    uint16_t tmp_width, tmp_height;
    int res = 0;

    // get res
    if(access("/kvmapp/kvm/width", F_OK) == 0){
        fp = fopen("/kvmapp/kvm/width", "r");
        fseek(fp, 0, SEEK_END);
        file_size = ftell(fp);
        fseek(fp, 0, SEEK_SET);
        fread(RW_Data, sizeof(char), file_size, fp);
        fclose(fp);
        RW_Data[file_size] = 0;
        tmp_width = atoi((char*)RW_Data);
    } else {
        tmp_width = 1920;
    }
    if(access("/kvmapp/kvm/height", F_OK) == 0){
        fp = fopen("/kvmapp/kvm/height", "r");
        fseek(fp, 0, SEEK_END);
        file_size = ftell(fp);
        fseek(fp, 0, SEEK_SET);
        fread(RW_Data, sizeof(char), file_size, fp);
        fclose(fp);
        RW_Data[file_size] = 0;
        tmp_height = atoi((char*)RW_Data);
    } else {
        tmp_height = 1080;
    }

    // res min limit
    if(tmp_width < vi_min_width){
        tmp_width = vi_min_width;
	    char Cmd[100]={0};
        sprintf(Cmd, "echo %d > %s", vi_min_width, vi_width_path);
	    system(Cmd);
    }
    if(tmp_height < vi_min_height){
        tmp_height = vi_min_height;
	    char Cmd[100]={0};
        sprintf(Cmd, "echo %d > %s", vi_min_height, vi_height_path);
	    system(Cmd);
    }
    // res max limit
    if(tmp_width > vi_max_width){
        tmp_width = vi_max_width;
	    char Cmd[100]={0};
        sprintf(Cmd, "echo %d > %s", vi_max_width, vi_width_path);
	    system(Cmd);
    }
    if(tmp_height > vi_max_height){
        tmp_height = vi_max_height;
	    char Cmd[100]={0};
        sprintf(Cmd, "echo %d > %s", vi_max_height, vi_height_path);
	    system(Cmd);
    }

    // res change ?
    if(kvmv_cfg.vi_width != tmp_width){
        kvmv_cfg.vi_width = tmp_width;
        printf("[kvmk] get new width = %d\n", kvmv_cfg.vi_width);
        res = 1;
    }
    if(kvmv_cfg.vi_height != tmp_height){
        kvmv_cfg.vi_height = tmp_height;
        printf("[kvmk] get new height = %d\n", kvmv_cfg.vi_height);
        res = 1;
    }
    return res;
}

uint8_t auto_try_res()
{
    char Cmd[100]={0};
    uint8_t err_code;
    uint8_t auto_trying_times = 0;

    for (auto_trying_times = 0; auto_trying_times < sizeof(hdmi_res_list)/4; auto_trying_times++){
        err_code = refresh_vi_state();
        switch(err_code){
        case 0:
            // shouldn't be possible to run here
            cam->restart(default_vpss_width, default_vpss_height, image::FMT_YVU420SP);
            printf("[kvmv] VI not init\n");
            break;
        case 1:
            printf("[kvmv] VI subsystem is normal\n");
            return 1;
            break;
        case 2:
            // Nothing readable on the port. That is not a resolution problem,
            // so walking the list against it cannot help: return, and let the
            // detection thread come back later.
            //
            // This used to decrement the counter and sleep a second here
            // instead. The sleep keeps the core free, but the loop still sits
            // on the same index for as long as the input stays unreadable, and
            // it still prints once a second into the tmpfs that the restart
            // path needs 36MB of.
            //
            // The counter is uint8_t, so at index 0 the decrement wraps to 255
            // and the loop's increment wraps it back to 0. There is no exit
            // from the loop at all.
            //
            // It also blocks shutdown. try_exit_thread is only tested at the
            // top of the detection thread's own loop, which this never returns
            // to, so a deinit that joins the thread waits on an unplugged
            // cable.
            //
            // Returning hands the wait to that loop, which sleeps for the same
            // second and tests try_exit_thread every time round.
            //
            // debug rather than printf, matching the same condition in mode 2
            // below. debug_en is 0 unless someone asks for it, so a disconnected
            // input costs nothing to report.
            debug("[kvmv] Cannot obtain HDMI input\n");
            return 2;
        case 3: // width too small
        case 4: // width too large
        case 5: // height too small
        case 6: // height too large
            // CSI abnormal due to resolution error
            // The test list is short; sequential testing can be performed
            printf("[kvmv] Trying %d * %d res ..\n", hdmi_res_list[auto_trying_times][0], hdmi_res_list[auto_trying_times][1]);
            sprintf(Cmd, "echo %d > %s", hdmi_res_list[auto_trying_times][0], vi_width_path);
            system(Cmd);
            sprintf(Cmd, "echo %d > %s", hdmi_res_list[auto_trying_times][1], vi_height_path);
            system(Cmd);

            kvmv_cfg.vi_width = hdmi_res_list[auto_trying_times][0];
            kvmv_cfg.vi_height = hdmi_res_list[auto_trying_times][1];
            printf("[kvmv] restart cam...\n");
            cam->restart(default_vpss_width, default_vpss_height, image::FMT_YVU420SP);
            time::sleep_ms(50);
            break;
        case 7: // Unknown reason
            printf("[kvmv] CSI abnormal: Unknown reason\n");
            break;
        }
    }
    err_code = refresh_vi_state();
    if (err_code == 1 || err_code == 2) return err_code;
    return 0;
}

/* return :
 * 0 : error
 * 1 : out of mem
 * 2 : normal
*/
uint8_t chack_ion()
{
    // cat /sys/kernel/debug/ion/cvi_carveout_heap_dump/summary | grep "usage rate:" | awk -F '[:%]' '{print $2}'
	uint8_t RW_Data[10];
    uint8_t ATOI_Data[3] = {0};
    uint8_t ion_usage_rate;
    char Cmd[150]={0};
    // sprintf( Cmd, "cat /sys/kernel/debug/ion/cvi_carveout_heap_dump/summary | grep \"usage rate:\" | awk -F '[:%]' '{print $2}'");
    sprintf( Cmd, "cat /sys/kernel/debug/ion/cvi_carveout_heap_dump/summary | grep \"usage rate:\" | awk '{print $2}'");
    FILE* fp = popen( Cmd, "r" );
    if ( NULL == fp )
    {
        pclose(fp);
        return 0;
    }
    fgets((char*)RW_Data, 8, fp);
    pclose(fp);
    RW_Data[8] = 0;
    if (RW_Data[6] == '&') return 1;
    else {
        ATOI_Data[0] = RW_Data[5];
        ATOI_Data[1] = RW_Data[6];
    }
    ion_usage_rate = atoi((char*)ATOI_Data);

    if(ion_usage_rate >= 95) return 1;
    else return 2;
}

void lt6911_enable()
{
	uint8_t buf[2];

    if(kvmv_cfg.hdmi_version == 3){
        return;
    }

	buf[0] = 0xff;
	buf[1] = 0x80;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

	buf[0] = 0xee;
	buf[1] = 0x01;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

    if(kvmv_cfg.hdmi_version == 1){
        // disable watchdog
        buf[0] = 0x10;
        buf[1] = 0x00;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 2);
    }
}

void lt6911_disable()
{
	uint8_t buf[2];

    if(kvmv_cfg.hdmi_version == 3){
        return;
    }

	buf[0] = 0xff;
	buf[1] = 0x80;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

	buf[0] = 0xee;
	buf[1] = 0x00;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 2);
}

void lt6911_get_hdmi_errer()
{
	uint8_t buf[6];

	buf[0] = 0xff;
	buf[1] = 0xC0;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

	buf[0] = 0x20;
	buf[1] = 0x01;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

	time::sleep_ms(100);

	buf[0] = 0x24;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 1);

	maix::Bytes *dat = LT6911_i2c.readfrom(LT6911_ADDR, 6);

	buf[0] = 0x20;
	buf[1] = 0x07;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

	for(int i = 0; i < 6; i++){
		buf[i] = (uint8_t)dat->data[i];
	}
	delete dat;

	debug("hdmi_errer_code = %x, %x, %x, %x, %x, %x\n", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5]);
}

uint8_t lt6911_get_hdmi_res()
{
	uint8_t buf[2];
	uint8_t revbuf[4];
	uint16_t Vactive;
	uint16_t Hactive;


    if(kvmv_cfg.hdmi_version == 0){
        // LT6911C
        buf[0] = 0xff;
        buf[1] = 0xd2;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

        buf[0] = 0x83;
        buf[1] = 0x11;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

        time::sleep_ms(5);

        // Vactive
        buf[0] = 0x96;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 1);
        maix::Bytes *dat0 = LT6911_i2c.readfrom(LT6911_ADDR, 2);

        // Hactive
        buf[0] = 0x8b;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 1);
        maix::Bytes *dat1 = LT6911_i2c.readfrom(LT6911_ADDR, 2);

        revbuf[0] = (uint8_t)dat0->data[0];
        revbuf[1] = (uint8_t)dat0->data[1];
        revbuf[2] = (uint8_t)dat1->data[0];
        revbuf[3] = (uint8_t)dat1->data[1];

        Vactive = (revbuf[0] << 8)|revbuf[1];
        Hactive = (revbuf[2] << 8)|revbuf[3];
        Hactive *= 2;

        debug("[hdmi]HDMI res modification event\n");
        debug("[hdmi]new res: %d * %d\n", Hactive, Vactive);

        delete dat0;
        delete dat1;

        if (Vactive != 0 && Hactive != 0){
            return 1;
        } else {
            // system("echo 0 > %s", vi_height_path);
            // system("echo 0 > %s", vi_width_path);
            return 0;
        }
    } else if (kvmv_cfg.hdmi_version == 1){
        // LT6911UXC
        buf[0] = 0xff;
        buf[1] = 0x86;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

        buf[0] = 0xff;
        buf[1] = 0x86;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

        // HDMI signal disappear/stable
        buf[0] = 0xA3;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 1);
        maix::Bytes *dat0 = LT6911_i2c.readfrom(LT6911_ADDR, 1);

        revbuf[0] = (uint8_t)dat0->data[0];
        delete dat0;

        debug("[hdmi]HDMI-UXC res modification event\n");

        if(revbuf[0] == 0x55)       return 1;
        else if(revbuf[0] == 0x88)  return 0;
        else return 0;
    } else if (kvmv_cfg.hdmi_version == 3){
        // LT6911D
        buf[0] = 0xff;
        buf[1] = 0xe0;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

        // Hactive
        buf[0] = 0x8c;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 1);
        maix::Bytes *dat0 = LT6911_i2c.readfrom(LT6911_ADDR, 2);

        // Vactive
        buf[0] = 0x8e;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 1);
        maix::Bytes *dat1 = LT6911_i2c.readfrom(LT6911_ADDR, 2);

        revbuf[0] = (uint8_t)dat0->data[0];
        revbuf[1] = (uint8_t)dat0->data[1];
        revbuf[2] = (uint8_t)dat1->data[0];
        revbuf[3] = (uint8_t)dat1->data[1];

        Hactive = ((revbuf[0] << 8)|revbuf[1]) * 2;
        Vactive = (revbuf[2] << 8)|revbuf[3];

        debug("[hdmi]HDMI-D res modification event\n");
        debug("[hdmi]new res: %d * %d\n", Hactive, Vactive);

        delete dat0;
        delete dat1;

        if(Hactive != 0 || Vactive != 0) return 1;
        else return 0;
    } else {
        return 0;
    }
}

void lt6911_get_hdmi_clk()
{
	uint8_t buf[2];
	uint8_t revbuf[3];
	uint32_t clk;

	buf[0] = 0xff;
	buf[1] = 0xa0;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

	buf[0] = 0x34;
	buf[1] = 0x0b;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

	time::sleep_ms(50);

	// clk
	buf[0] = 0xff;
	buf[1] = 0xb8;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

	buf[0] = 0xb1;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 1);
	maix::Bytes *dat0 = LT6911_i2c.readfrom(LT6911_ADDR, 3);

	revbuf[0] = (uint8_t)dat0->data[0];
	revbuf[1] = (uint8_t)dat0->data[1];
	revbuf[2] = (uint8_t)dat0->data[2];
	revbuf[0] &= 0x07;

	clk = revbuf[0];
	clk <<= 8;
	clk |= revbuf[1];
	clk <<= 8;
	clk |= revbuf[2];

	debug("[hdmi]HDMI CLK = %d\n", clk);

	delete dat0;
}

uint8_t lt6911_get_csi_res(uint16_t *p_width, uint16_t *p_height)
{
	uint8_t buf[2];
	uint8_t revbuf[4];
    uint8_t res_type;
	static uint16_t old_Vactive;
	static uint16_t old_Hactive;
	uint16_t Vactive;
	uint16_t Hactive;

    if(kvmv_cfg.hdmi_version == 0){
        // LT6911C
        buf[0] = 0xff;
        buf[1] = 0xc2;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

        // Vactive
        buf[0] = 0x06;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 1);
        maix::Bytes *dat0 = LT6911_i2c.readfrom(LT6911_ADDR, 2);

        // Hactive
        buf[0] = 0x38;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 1);
        maix::Bytes *dat1 = LT6911_i2c.readfrom(LT6911_ADDR, 2);

        revbuf[0] = (uint8_t)dat0->data[0];
        revbuf[1] = (uint8_t)dat0->data[1];
        revbuf[2] = (uint8_t)dat1->data[0];
        revbuf[3] = (uint8_t)dat1->data[1];

        delete dat0;
        delete dat1;

        Vactive = (revbuf[0] << 8)|revbuf[1];
        Hactive = (revbuf[2] << 8)|revbuf[3];
    } else if(kvmv_cfg.hdmi_version == 1) {
        // LT6911UXC
        debug("[hdmi]UXC get csi res\n");
        buf[0] = 0xff;
        buf[1] = 0x85;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

        // Vactive
        buf[0] = 0xF0;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 1);
        maix::Bytes *dat0 = LT6911_i2c.readfrom(LT6911_ADDR, 2);

        // Hactive
        buf[0] = 0xEA;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 1);
        maix::Bytes *dat1 = LT6911_i2c.readfrom(LT6911_ADDR, 2);

        revbuf[0] = (uint8_t)dat0->data[0];
        revbuf[1] = (uint8_t)dat0->data[1];
        revbuf[2] = (uint8_t)dat1->data[0];
        revbuf[3] = (uint8_t)dat1->data[1];

        delete dat0;
        delete dat1;

        Vactive = (revbuf[0] << 8)|revbuf[1];
        Hactive = (revbuf[2] << 8)|revbuf[3];
    } else if(kvmv_cfg.hdmi_version == 3) {
        // LT6911D
        debug("[hdmi]D get csi res\n");
        buf[0] = 0xff;
        buf[1] = 0xe0;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 2);

        // Vactive
        buf[0] = 0x8e;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 1);
        maix::Bytes *dat0 = LT6911_i2c.readfrom(LT6911_ADDR, 2);

        // Hactive
        buf[0] = 0x8c;
        LT6911_i2c.writeto(LT6911_ADDR, buf, 1);
        maix::Bytes *dat1 = LT6911_i2c.readfrom(LT6911_ADDR, 2);

        revbuf[0] = (uint8_t)dat0->data[0];
        revbuf[1] = (uint8_t)dat0->data[1];
        revbuf[2] = (uint8_t)dat1->data[0];
        revbuf[3] = (uint8_t)dat1->data[1];

        delete dat0;
        delete dat1;

        Vactive = (revbuf[0] << 8)|revbuf[1];
        Hactive = ((revbuf[2] << 8)|revbuf[3]) * 2;
    } else {
        return UNKNOWN_RES;
    }

    res_type = check_res(Hactive, Vactive);
    if(res_type == NORMAL_RES)
    {
        if(old_Hactive != Hactive || old_Vactive != Vactive){
            old_Hactive = Hactive;
            old_Vactive = Vactive;
            // p_kvm_cfg->width = Hactive;
            // p_kvm_cfg->height = Vactive;
            *p_width = Hactive;
            *p_height = Vactive;
            res_type = NEW_RES;
        }
    }

    switch (res_type)
    {
    case NORMAL_RES:
        printf("[hdmi] get res : %d * %d\n", Hactive, Vactive);
        break;
    case NEW_RES:
        printf("[hdmi] get new res : %d * %d\n", Hactive, Vactive);
        write_res_to_file(Hactive, Vactive);
        break;
    case UNSUPPORT_RES:
        printf("[hdmi] get unsupport res : %d * %d\n", Hactive, Vactive);
        break;
    case UNKNOWN_RES:
        printf("[hdmi] get unknown res : %d * %d\n", Hactive, Vactive);
        break;
    }

	return res_type;
}

void lt6911_write_reg(uint8_t reg, uint8_t val)
{
	uint8_t buf[2];
	buf[0] = reg;
	buf[1] = val;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 2);
}

void lt6911_read_reg(uint8_t reg)
{
	uint8_t buf[16];
	buf[0] = reg;
	LT6911_i2c.writeto(LT6911_ADDR, buf, 1);

	maix::Bytes *dat = LT6911_i2c.readfrom(LT6911_ADDR, 16);

	for(int i = 0; i < 16; i++){
		buf[i] = (uint8_t)dat->data[i];
	}

	delete dat;

	debug("[hdmi]%3x %3x %3x %3x %3x %3x %3x %3x |%3x %3x %3x %3x %3x %3x %3x %3x \n", \
			buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6], buf[7], \
			buf[8], buf[9], buf[10], buf[11], buf[12], buf[13], buf[14], buf[15]);
}

uint8_t lt6911_read_one_reg(uint8_t reg)
{
	uint8_t ret;
	ret = reg;
	LT6911_i2c.writeto(LT6911_ADDR, &ret, 1);

	maix::Bytes *dat = LT6911_i2c.readfrom(LT6911_ADDR, 1);

	ret = (uint8_t)dat->data[0];

	delete dat;
	return ret;
}

void lt6911_write_edid(void)
{
	uint8_t i, j;
	uint8_t buf[2];
	lt6911_enable();

	// to 90
	lt6911_write_reg(0xff, 0x90);
	buf[0] = lt6911_read_one_reg(0x02);
	buf[0] &= 0xDF;
	lt6911_write_reg(0x02, buf[0]);
	buf[0] |= 0x20;
	lt6911_write_reg(0x02, buf[0]);

	// to 80
	lt6911_write_reg(0xff, 0x80);
	// wren enable
	lt6911_write_reg(0x5A, 0x86);
	lt6911_write_reg(0x5A, 0x82);

	for(i = 0; i < 16; i++){
		// 写wren命令(为一个pulse，此时不需要考虑wrrd_mode，spi_paddr[1:0]的值)
		lt6911_write_reg(0x5A, 0x86);
		lt6911_write_reg(0x5A, 0x82);

		// 配置spi_len[3:0]= 15，可配置，spi内部加1，即配置一次写入16个字节
		lt6911_write_reg(0x5E, 0xEF);
		lt6911_write_reg(0x5A, 0xA2);
		lt6911_write_reg(0x5A, 0x82);

		lt6911_write_reg(0x58, 0x01);

		if(i < 8) {
			for(j = 0; j < 16; j++){
				lt6911_write_reg(0x59, NanoKVM_edit[i*16+j]);
			}
		} else {
			for(j = 0; j < 16; j++){
				lt6911_write_reg(0x59, 0x00);
			}
		}

		// 把fifo数据写到flash(当wrrd_mode = 1，spi_paddr= 2’b10,addr[23:0](地址在写入过程中需要保持不变)准备好的情况下，给一个spi_sta的pulse，就开始写入flash)
		lt6911_write_reg(0x5B, 0x00);
		lt6911_write_reg(0x5C, 0x21);
		lt6911_write_reg(0x5D, i*16);
		lt6911_write_reg(0x5E, 0xE0);
		lt6911_write_reg(0x5A, 0x92);
		lt6911_write_reg(0x5A, 0x82);

	}

	lt6911_write_reg(0x5A, 0x8A);
	lt6911_write_reg(0x5A, 0x82);

	lt6911_disable();
	debug("[hdmi]lt6911_write_edid OK\n");
}

void lt6911_read_edid(void)
{
	uint8_t i, j;
	lt6911_enable();

	// to 80
	lt6911_write_reg(0xff, 0x80);
	lt6911_write_reg(0xEE, 0x01);

	// configure parameter
	lt6911_write_reg(0x5A, 0x80);
	lt6911_write_reg(0x5E, 0xC0);
	lt6911_write_reg(0x58, 0x00);
	lt6911_write_reg(0x59, 0x51);
	lt6911_write_reg(0x5A, 0x90);
	time::sleep_ms(1);
	lt6911_write_reg(0x5A, 0x80);

	for(i = 0; i < 16; i++){
		// to 90
		lt6911_write_reg(0xff, 0x90);
		// fifo rst_n
		lt6911_write_reg(0x02, 0xdf);
		time::sleep_ms(1);
		lt6911_write_reg(0x02, 0xff);

		// wren
		// to 80
		lt6911_write_reg(0xff, 0x80);
		// fifo rst_n
		lt6911_write_reg(0x5A, 0x84);
		time::sleep_ms(1);
		lt6911_write_reg(0x5A, 0x80);

		// flash to fifo
		lt6911_write_reg(0x5E, 0x6F);
		lt6911_write_reg(0x5A, 0xA0);
		time::sleep_ms(1);
		lt6911_write_reg(0x5A, 0x80);
		lt6911_write_reg(0x5B, 0x00);
		lt6911_write_reg(0x5C, 0x21);
		lt6911_write_reg(0x5D, 16*i);
		lt6911_write_reg(0x5A, 0x90);
		time::sleep_ms(5);
		lt6911_write_reg(0x5A, 0x80);
		lt6911_write_reg(0x58, 0x01);

		// for(j = 0; j < 16; j++){
		debug("[hdmi] EDID: ");
		lt6911_read_reg(0x5F);
		// }
		time::sleep_ms(10);
	}

	// wrdi
	lt6911_write_reg(0x5A, 0x88);
	time::sleep_ms(1);
	lt6911_write_reg(0x5A, 0x80);

	lt6911_disable();
	debug("[hdmi]lt6911_read_edid OK\n");
}

void lt6911_read_fw(void)
{
	uint32_t i, j;
	uint8_t buf[3];
	lt6911_enable();

	// to 80
	lt6911_write_reg(0xff, 0x80);
	lt6911_write_reg(0xEE, 0x01);

	// configure parameter
	lt6911_write_reg(0x5A, 0x80);
	lt6911_write_reg(0x5E, 0xC0);
	lt6911_write_reg(0x58, 0x00);
	lt6911_write_reg(0x59, 0x51);
	lt6911_write_reg(0x5A, 0x90);
	time::sleep_ms(1);
	lt6911_write_reg(0x5A, 0x80);

	for(i = 0x1; i < 50000; i++){
		buf[0] = ((i*16) & 0xFF0000) >> 16;
		buf[1] = ((i*16) & 0xFF00) >> 8;
		buf[2] = ((i*16) & 0xFF);
		// to 90
		lt6911_write_reg(0xff, 0x90);
		// fifo rst_n
		lt6911_write_reg(0x02, 0xdf);
		time::sleep_ms(1);
		lt6911_write_reg(0x02, 0xff);

		// wren
		// to 80
		lt6911_write_reg(0xff, 0x80);
		// fifo rst_n
		lt6911_write_reg(0x5A, 0x84);
		time::sleep_ms(1);
		lt6911_write_reg(0x5A, 0x80);

		// flash to fifo
		lt6911_write_reg(0x5E, 0x6F);
		lt6911_write_reg(0x5A, 0xA0);
		time::sleep_ms(1);
		lt6911_write_reg(0x5A, 0x80);
		lt6911_write_reg(0x5B, buf[0]);
		lt6911_write_reg(0x5C, buf[1]);
		lt6911_write_reg(0x5D, buf[2]);
		lt6911_write_reg(0x5A, 0x90);
		time::sleep_ms(5);
		lt6911_write_reg(0x5A, 0x80);
		lt6911_write_reg(0x58, 0x01);

		// for(j = 0; j < 16; j++){
		debug("[hdmi]REG %6x: ", i*16);
		lt6911_read_reg(0x5F);
		// }
		time::sleep_ms(10);
	}

	// wrdi
	lt6911_write_reg(0x5A, 0x88);
	time::sleep_ms(1);
	lt6911_write_reg(0x5A, 0x80);

	lt6911_disable();
	debug("[hdmi]lt6911_read_edid OK\n");
}

// ==============================================================

void* watchdog_sf_feed(void * arg)
{
    while(true)
    {
        if(__atomic_load_n(&kvmv_cfg.try_exit_thread, __ATOMIC_ACQUIRE) == 1)
            break;
        time::sleep_ms(500);
        if (watchdog_sf_is_open()){
            if (chack_ion() == 1){
                debug("[kvmv] Ion memory is full reboot now!\n");
                system("reboot");
            }
            // debug("[kvmv] watchdog_sf_feed now!\n");
            vision_update_watchdog();
        }
    }

    // The return is not decoration. Both of libkvm's threads are declared void*
    // and neither used to return anything, so control fell off the end of a
    // function that owes a value - undefined behaviour, and GCC is entitled to
    // assume it never happens. It concluded that the break above could not be
    // reached and deleted the test that leads to it. In the library shipped
    // before 2026-08-20 the loop held no reference to the exit flag at all, and
    // making the flag atomic on its own did not bring the test back: the load
    // appeared, and its result was still never examined.
    //
    // So kvmv_deinit's join waited on a thread with no way out, the teardown
    // never reached mmf_deinit, and every stop of the server left a VI channel
    // pool and an ISP shared buffer behind.
    return NULL;
}

void get_hdmi_version()
{
	FILE *fp;
	uint8_t RW_Data[2];
    system("/kvmapp/system/init.d/S15kvmhwd get_hdmi_version");
	if(access("/etc/kvm/hdmi_version", F_OK) == 0){
        fp = fopen("/etc/kvm/hdmi_version", "r");
        fread(RW_Data, sizeof(char), 2, fp);
        fclose(fp);
        if(RW_Data[0] == 'u'){
            // 6911uxc
            if(RW_Data[1] == 'e'){
                kvmv_cfg.hdmi_version = 2;
                debug("[hdmi]HDMI-UE exist!\n");
                set_hdmi_mode(1);
            } else if(RW_Data[1] == 'x') {
                kvmv_cfg.hdmi_version = 1;
                debug("[hdmi]HDMI-UX exist!\n");
            } else {
                kvmv_cfg.hdmi_version = 1;
                debug("[hdmi]Incomplete version number, set to 'ux'\n");
            }
        } else if(RW_Data[0] == 'd'){
            // 6911d
            kvmv_cfg.hdmi_version = 3;
            debug("[hdmi]HDMI-D exist!\n");
        } else {
            // 6911c
            kvmv_cfg.hdmi_version = 0;
            debug("[hdmi]HDMI-C exist!\n");
        }
        RW_Data[0] = 0;
    } else {
        kvmv_cfg.hdmi_version = 0;
    }
}

void* vi_subsystem_detection(void * arg)
{
	uint64_t __attribute__((unused)) int_time;

	FILE *fp;
	uint8_t RW_Data[2];
    uint8_t file_size;
    uint8_t tmp8;
    uint8_t rising_times = 0;
    uint8_t falling_times = 0;
	uint8_t cam_need_restart = 0;
    uint8_t auto_change_mode = 0;
    kvmv_cfg.thread_is_running = 1;
	if(access("/proc/lt_int", F_OK) != 0){
		time::sleep_ms(10);
		debug("[hdmi]/proc/lt_int not ok\n");
	}

    get_hdmi_version();

    // Refresh on elapsed time, not loop iterations whose work varies widely.
    last_vi_state_refresh_ms = vi_state_shared::monotonic_ms() - vi_state_publish_interval_ms;
    while(true)
    {
        if(__atomic_load_n(&kvmv_cfg.try_exit_thread, __ATOMIC_ACQUIRE) == 1)
            break;

        uint8_t get_new_hdmi_mode = get_hdmi_mode();
        uint8_t try_res;
        uint8_t err_code;
        uint8_t vi_state_refreshed = 0;
        uint8_t vi_state_due = vi_state_shared::monotonic_ms() - last_vi_state_refresh_ms >= vi_state_publish_interval_ms;
        uint8_t manual_vi_init_now =
            (kvmv_cfg.hdmi_mode == 2 &&
             (get_new_hdmi_mode == 1 || kvmv_cfg.vi_detect_state == 0));
        uint8_t defer_vi_state_refresh =
            (kvmv_cfg.hdmi_mode == 1 &&
             (kvmv_cfg.vi_detect_state == 1 || get_new_hdmi_mode == 1)) ||
            (kvmv_cfg.hdmi_mode == 2 &&
             (manual_vi_init_now || kvmv_cfg.vi_detect_state == 1));

        if (vi_state_due && !defer_vi_state_refresh) {
            refresh_vi_state();
            vi_state_refreshed = 1;
        }

        switch (kvmv_cfg.hdmi_mode){
        case 0:
            // Switching to Mode 0 requires restarting HDMI (effective only for PCIe version)
            // Handling of automatic detection situations
            if(get_new_hdmi_mode == 1){
                kvmv_cfg.vi_detect_state = 0;
                // reset hdmi_state
                char Cmd[100]={0};
                sprintf(Cmd, "echo 0 > %s", hdmi_state_path);
                system(Cmd);
                // reset hdmi
                kvmv_hdmi_control(0);
                time::sleep_ms(10);
                kvmv_hdmi_control(1);
            }
            // Initialize VI with I2C information after HDMI insertion
            fp = fopen(hdmi_state_path, "r+");
            if(fp != NULL){
                // fseek(fp, 0, SEEK_END);
                // file_size = ftell(fp);
                // fseek(fp, 0, SEEK_SET);
                fread(RW_Data, sizeof(char), 2, fp);
                tmp8 = atoi((char*)RW_Data);
                // debug("[hdmi]UXC tmp8 = %d\n", tmp8);
                if(tmp8 != 0){
                    // reset hdmi_state
                    fputs("0", fp);
                    // count edge ints
                    rising_times = tmp8%10;   // RISING times
                    falling_times = tmp8/10;   // FALLING times

                    if(kvmv_cfg.hdmi_stop_flag != 1){
                        kvmv_cfg.hdmi_reading_flag = 1;
                        if(kvmv_cfg.hdmi_version == 0){
                            // LT6911C
                            if(rising_times != 0){
                                lt6911_enable();
                                if(lt6911_get_hdmi_res()){
                                    // hdmi get res
                                    debug("[hdmi] C HDMI cable insertion!\n");
                                    set_hdmi_detection_state(1);
                                    kvmv_cfg.hdmi_res_type = lt6911_get_csi_res(&kvmv_cfg.vi_width, &kvmv_cfg.vi_height);
                                    if (kvmv_cfg.hdmi_res_type == NEW_RES) kvmv_cfg.reopen_cam_flag = 1;
                                    else if (kvmv_cfg.hdmi_res_type == UNKNOWN_RES){
                                        /* Move HDMI resolution modification directly
                                            to mode1 to solve deadlock problem */
                                        auto_change_mode = 1;
                                        set_hdmi_mode(1);
                                    }
                                } else {
                                    // HDMI res = 0*0/x*0
                                    debug("[hdmi] C HDMI cable unplugged!\n");
                                    set_hdmi_detection_state(0);
                                }
                                lt6911_disable();
                            }
                        } else if (kvmv_cfg.hdmi_version == 1){
                            // LT6911UXC
                                debug("[hdmi] UXC int\n");
                            if(falling_times != 0){
                                debug("[hdmi] UXC int && \n");
                                lt6911_enable();
                                if(lt6911_get_hdmi_res()){
                                    // hdmi get res
                                    debug("[hdmi] UXC HDMI cable insertion!\n");
                                    set_hdmi_detection_state(1);
                                    kvmv_cfg.hdmi_res_type = lt6911_get_csi_res(&kvmv_cfg.vi_width, &kvmv_cfg.vi_height);
                                    if (kvmv_cfg.hdmi_res_type == NEW_RES) kvmv_cfg.reopen_cam_flag = 1;
                                    else if (kvmv_cfg.hdmi_res_type == UNKNOWN_RES){
                                        /* Move HDMI resolution modification directly
                                            to mode1 to solve deadlock problem */
                                        auto_change_mode = 1;
                                        set_hdmi_mode(1);
                                    }
                                } else {
                                    // HDMI res = 0*0/x*0
                                    debug("[hdmi] UXC HDMI cable unplugged!\n");
                                    set_hdmi_detection_state(0);
                                }
                                lt6911_disable();
                            }
                        } else if (kvmv_cfg.hdmi_version == 3){
                            // LT6911D
                                debug("[hdmi] D int\n");
                            if(falling_times != 0){
                                debug("[hdmi] D int && \n");
                                lt6911_enable();
                                if(lt6911_get_hdmi_res()){
                                    // hdmi get res
                                    debug("[hdmi] D HDMI cable insertion!\n");
                                    set_hdmi_detection_state(1);
                                    kvmv_cfg.hdmi_res_type = lt6911_get_csi_res(&kvmv_cfg.vi_width, &kvmv_cfg.vi_height);
                                    if (kvmv_cfg.hdmi_res_type == NEW_RES) kvmv_cfg.reopen_cam_flag = 1;
                                    else if (kvmv_cfg.hdmi_res_type == UNKNOWN_RES){
                                        /* Move HDMI resolution modification directly
                                            to mode1 to solve deadlock problem */
                                        auto_change_mode = 1;
                                        set_hdmi_mode(1);
                                    }
                                } else {
                                    // HDMI res = 0*0/x*0
                                    debug("[hdmi] D HDMI cable unplugged!\n");
                                    set_hdmi_detection_state(0);
                                }
                                lt6911_disable();
                            }
                        } else {
                            debug("[hdmi] Chip not supported for reading \n");
                        }
                        kvmv_cfg.hdmi_reading_flag = 0;
                    }
                }
                fclose(fp);
            }
            break;
        /* Mode 1 & 2 will disable API access to the image during detection,
           and will output -4: Modifying image resolution, please wait.*/
        case 1: // Automatically trying common resolutions
            /* kvmv_cfg.vi_detect_state :
             * 0: HDMI standard mode, detection program does not interfere with the camera
             * 1: Preparing / Testing in progress
             * 2: Test completed: Suitable resolution found,
             */

            if(get_new_hdmi_mode == 1){
                kvmv_cfg.vi_detect_state = 1;
            }
            if(kvmv_cfg.vi_detect_state == 1){
                try_res = auto_try_res();
                if (try_res == 1) {
                    kvmv_cfg.hdmi_res_err = NORMAL_RES;
                    kvmv_cfg.hdmi_try_rounds = 0;
                    kvmv_cfg.vi_detect_state = 2;
                } else if (try_res == 0) {
                    /*  There may be a situation where the correct resolution cannot be recognized
                        after one round of detection. By default, Try_rounds_HDMI_err_res rounds will be detected,
                        and if it cannot be detected, jump to the next mode */
                    kvmv_cfg.hdmi_res_err = ERROR_RES;
                    kvmv_cfg.hdmi_try_rounds++;
                    if(kvmv_cfg.hdmi_try_rounds >= Try_rounds_HDMI_err_res){
                        kvmv_cfg.hdmi_try_rounds = 0;
                        kvmv_cfg.vi_detect_state = 1;
                        // printf("[kvmv] Suitable resolution not found, switching to manual input mode automatically\n");
                        if(auto_change_mode == 1){
                            auto_change_mode = 0;
                            set_hdmi_mode(0);
                        } else {
                            set_hdmi_mode(2);
                        }
                    }
                } else if (try_res == 2) {
                    kvmv_cfg.hdmi_try_rounds = 0;
                    // Cannot obtain HDMI input / No signal on HDMI.
                    //
                    // Wait before looking again. auto_try_res now returns as
                    // soon as it finds the input unreadable, and this branch of
                    // the thread has no delay of its own, so without this the
                    // spin that used to live inside auto_try_res would simply
                    // move out here. One second matches the interval the
                    // vi_detect_state == 2 branch below already uses.
                    time::sleep_ms(1000);
                }
            } else if (kvmv_cfg.vi_detect_state == 2){
                // Low-frequency detection of HDMI status, no log output
                if (vi_state_refreshed && last_vi_state_code != 1) {
                    kvmv_cfg.vi_detect_state = 1;
                }
            } else {
                kvmv_cfg.vi_detect_state = 1;
            }
            break;
        case 2:
            // Manually initialize VI.
            if (manual_vi_init_now) {
                // Initialize immediately when entering manual mode instead of
                // waiting for the next periodic state refresh.
                kvmv_cfg.vi_detect_state = 1;
            }
            if(manual_vi_init_now || vi_state_due){
                if (kvmv_cfg.vi_detect_state == 1){
                    // detect_res
                    if (get_manual_resolution()) {
                        debug("[kvmv] restart cam...\n");
                        cam->restart(default_vpss_width, default_vpss_height, image::FMT_YVU420SP);
                    }

                    // Sample after a manual resolution change.
                    err_code = refresh_vi_state();
                    switch(err_code){
                    case 0:
                        debug("[kvmv] VI not init\n");
                        break;
                    case 1:
                        debug("[kvmv] HDMI and CSI status are normal\n");
                        kvmv_cfg.vi_detect_state = 2;
                        break;
                    case 2:
                        debug("[kvmv] HDMI abnormal\n");
                        break;
                    case 3:
                        debug("[kvmv] CSI abnormal: width too small\n");
                        break;
                    case 4:
                        debug("[kvmv] CSI abnormal: width too large\n");
                        break;
                    case 5:
                        debug("[kvmv] CSI abnormal: height too small\n");
                        break;
                    case 6:
                        debug("[kvmv] CSI abnormal: height too large\n");
                        break;
                    case 7:
                        debug("[kvmv] CSI abnormal: Unknown reason\n");
                        break;
                    }
                } else if (kvmv_cfg.vi_detect_state == 2){
                    // detection of HDMI status, no log output
                    err_code = last_vi_state_code;
                    if (err_code != 1) kvmv_cfg.vi_detect_state = 1;
                } else {
                    kvmv_cfg.vi_detect_state = 1;
                }
            }
            break;

        default:
            debug("Non-existent hdmi state = %d\n", kvmv_cfg.hdmi_mode);
            break;
        }

		// Poll fast only while the resolution is still in question. In mode 0
		// the board reads a settled input, and in modes 1 and 2 a detect state
		// of 2 means the resolution is found, so a further 90ms of latency on
		// an HDMI edge costs nothing a person can see.
		//
		// Measured on this fork, 2026-09-04, mode 0 with no HDMI signal and no
		// viewer: the thread took 57 ticks over 30 seconds, which is 1.9% of
		// the only core. The 10ms sleep is the whole of it, because the loop
		// opens /proc/lt_int on every pass. Upstream reports about 15% for the
		// same change; that figure does not reproduce here and it likely
		// includes the mode 1 spin this fork already stopped with the 1000ms
		// sleep in the try_res == 2 branch above.
		//
		// One thing this widens: the loop reads two characters of /proc/lt_int
		// and splits them into a falling and a rising count of one digit each.
		// Ten edges of one kind between two polls therefore read as zero. A
		// bouncing connector is ten times more likely to reach that in 100ms
		// than in 10ms. The driver is kernel side and not in this tree, so
		// whether it saturates or wraps is unverified.
		// The rebuild the reader asked for, run here so that every
		// cam->restart this library performs happens on one thread. Doing
		// it on the reader's own thread would not have excluded this one:
		// vi_mutex is taken in exactly one place, kvmv_read_img, and this
		// thread has never held it.
		//
		// The request is dropped rather than kept whenever a transition is
		// already in flight, because that transition rebuilds the channel
		// by itself. If the channel is still wedged afterwards the reader
		// asks again, one cooldown later.
		if (__atomic_load_n(&vi_wedge_rebuild_request, __ATOMIC_ACQUIRE) != 0) {
			__atomic_store_n(&vi_wedge_rebuild_request, 0, __ATOMIC_RELEASE);
			if (kvmv_cfg.vi_detect_state != 1 &&
					kvmv_cfg.reopen_cam_flag == 0 &&
					kvmv_cfg.hdmi_reading_flag == 0 &&
					kvmv_cfg.hdmi_stop_flag != 1) {
				printf("[kvmv]rebuilding the VI channel, the VPSS channel handed out nothing\n");
				cam->restart(default_vpss_width, default_vpss_height, image::FMT_YVU420SP);
			} else {
				// Worth a line of its own. A rebuild that never happens because
				// something else was already transitioning looks exactly like a
				// rebuild that happened and did not help.
				printf("[kvmv]a rebuild was asked for while the channel was already in transition, standing down\n");
			}
		}

		const uint32_t poll_interval_ms =
			(kvmv_cfg.hdmi_mode == 0 || kvmv_cfg.vi_detect_state == 2)
				? vi_detection_idle_poll_ms
				: vi_detection_active_poll_ms;
		time::sleep_ms(poll_interval_ms);
    }
    kvmv_cfg.thread_is_running = 0;

    // See watchdog_sf_feed: without this the compiler treats the break out of
    // the loop above as unreachable and removes the test that reaches it.
    return NULL;
}

int sync_vi_res()
{
    int res = 0;
    uint8_t RW_Data[35];
    FILE *fp;
    int file_size;
    uint16_t tmp16;

    // vi_width:
    if (access(vi_width_path, F_OK) != 0){
        kvmv_cfg.vi_width = default_vi_width;
        kvmv_cfg.vi_height = default_vi_height;
        res = -1;
        return res;
    } else {
        fp = fopen(vi_width_path, "r");
        fseek(fp, 0, SEEK_END);
        file_size = ftell(fp);
        fseek(fp, 0, SEEK_SET);
        fread(RW_Data, sizeof(char), file_size, fp);
        fclose(fp);
        RW_Data[file_size] = 0;
        tmp16 = atoi((char*)RW_Data);
        if(tmp16 != kvmv_cfg.vi_width){
            kvmv_cfg.vi_width = tmp16;
            debug("[hdmi] Get new HDMI width = %d\r\n", kvmv_cfg.vi_width);
            res = 1;
        }
    }
    // vi_height:
    if (access(vi_height_path, F_OK) != 0){
        kvmv_cfg.vi_height = default_vi_height;
        res = -1;
        return res;
    } else {
        fp = fopen(vi_height_path, "r");
        fseek(fp, 0, SEEK_END);
        file_size = ftell(fp);
        fseek(fp, 0, SEEK_SET);
        fread(RW_Data, sizeof(char), file_size, fp);
        fclose(fp);
        RW_Data[file_size] = 0;
        tmp16 = atoi((char*)RW_Data);
        if(tmp16 != kvmv_cfg.vi_height){
            kvmv_cfg.vi_height = tmp16;
            debug("[hdmi] Get new HDMI height = %d\r\n", kvmv_cfg.vi_height);
            res = 1;
        }
    }
    return res;
}

// Sample the frame and report whether it differs from the one sampled
// before it.
//
// This took an image::Image, which meant the caller had to map and copy
// the whole frame before it could ask. It reads a strided sample of the
// luma plane and nothing else, so it now takes the pixels where the
// hardware left them.
//
// A frame it cannot read counts as changed. Answering "unchanged" for a
// frame nobody looked at would stop the stream on a screen that is still
// moving; being wrong the other way costs one encoded frame.
//
// size comes from the hardware now, so it carries the stride padding that
// data_size() did not. The two agree at 1080p, where the stride equals the
// width. Where they disagree the size test below reports one change, and
// the next sample is against a like-for-like figure.
uint8_t frame_changed(const uint8_t *data, int size)
{
    static int raw_size = 0;
    static uint8_t Farame_sample[Farame_sample_size] = {0};
    uint8_t ret = 0;
    uint8_t sample_byte;
    if(data == NULL || size <= 0){
        return 1;
    }
    if(size != raw_size){
        raw_size = size;
        ret = 1;
    }
    int Detection_Pixel_Interval = raw_size/(Farame_sample_size*1.5);
    for(int i = 0; i < Farame_sample_size; i ++){
        // printf("[kvmv] i = %d\n", i);
        if(i >= raw_size){
            ret = 0;
        }
        sample_byte = *(data+(i*Detection_Pixel_Interval));
        if(sample_byte != Farame_sample[i]){
            Farame_sample[i] = sample_byte;
            ret = 1;
        }
    }
    return ret;
}

// Encode a frame the hardware still holds, straight into a save slot. This
// is the JPEG counterpart of frame_to_h264, and like that one it owns the
// VI frame from the moment it is called: every exit here releases it.
//
// The JPEG encoder lives on channel 0, which is where image::Image::to_jpeg
// puts it as well.
static int8_t frame_to_jpeg(int vi_ch, kvmv_data_t* dump_to, uint16_t quality)
{
    int ret = mmf_enc_jpg_push_vi_with_quality(0, vi_ch, quality);
    if (ret != 0) {
        release_save_buffer(dump_to);
        mmf_vi_frame_release(vi_ch);
        return IMG_VENC_ERROR;
    }

    uint8_t *data = NULL;
    int data_size = 0;
    ret = mmf_enc_jpg_pop(0, &data, &data_size);
    if (ret != 0 || data == NULL || data_size <= 0) {
        mmf_enc_jpg_deinit(0);
        release_save_buffer(dump_to);
        mmf_vi_frame_release(vi_ch);
        return IMG_VENC_ERROR;
    }

    if (!reserve_save_buffer(dump_to, (uint32_t)data_size)) {
        dump_to->img_data_size = 0;
        dump_to->img_data_type = 0;
        if (mmf_enc_jpg_free(0) != 0) {
            mmf_enc_jpg_deinit(0);
        }
        release_save_buffer(dump_to);
        mmf_vi_frame_release(vi_ch);
        return IMG_BUFFER_FULL;
    }

    memcpy(dump_to->p_img_data, data, data_size);
    dump_to->img_data_size = data_size;
    dump_to->img_data_type = VENC_MJPEG;
    // The encoder is done with the input frame once the output is out, so
    // the release below is in the right order behind this one.
    if (mmf_enc_jpg_free(0) != 0) {
        mmf_enc_jpg_deinit(0);
        dump_to->img_data_size = 0;
        dump_to->img_data_type = 0;
        release_save_buffer(dump_to);
        mmf_vi_frame_release(vi_ch);
        return IMG_VENC_ERROR;
    }
    mmf_vi_frame_release(vi_ch);
    return IMG_MJPEG_TYPE;
}

uint8_t kvmvenc_gop = default_h264_gop;
uint8_t kvmvenc_fps = default_h264_fps;
// 0 means the channel hands out every frame the source produces, which is
// what it did before this existed. A server too old to call
// set_capture_fps therefore behaves exactly as it used to.
uint8_t kvmvi_dst_fps = 0;
uint8_t kvmvi_fps_pending = 0;
kvm_venc_t kvm_venc;
mmf_venc_cfg_t cfg;
void init_venc_h264(uint16_t _width, uint16_t _height, uint16_t _qlty)
{
    cfg.type = 2; //1, h265, 2, h264
    cfg.w = _width;
    cfg.h = _height;
    cfg.fmt = mmf_invert_format_to_mmf(image::Format::FMT_YVU420SP);
    cfg.jpg_quality = 0;       // unused
    cfg.gop = kvmvenc_gop;
    // The rate controller divides the bitrate by the frame rate it is given
    // to decide what one frame may cost. Told 60 while the capture loop runs
    // at 30, it spent half the configured bitrate per frame and the stream
    // came out at about half the rate that was asked for.
    cfg.intput_fps = kvmvenc_fps;
    cfg.output_fps = kvmvenc_fps;
    cfg.bitrate = _qlty;  // 码率

    kvm_venc.mmf_venc_chn = default_venc_chn;
    kvm_venc.enc_h264_init = 0;
    kvm_venc.kvm_venc_cfg = cfg;

	// if(mmf_vdec_is_used(kvm_venc.mmf_venc_chn)){
		mmf_del_venc_channel(kvm_venc.mmf_venc_chn);
	// }
    if (0 != mmf_add_venc_channel(kvm_venc.mmf_venc_chn, &kvm_venc.kvm_venc_cfg)) {
        err::check_raise(err::ERR_RUNTIME, "mmf venc init failed!");
    }
    kvm_venc.enc_h264_init = 1;

	// init_kvm_h264_stream(&kvm_h264_stream, mmf_stream_buf);
	// init_h264_stream_struct(&kvm_h264_stream);
}

int h264_stream_dump(kvmv_data_t* dump_to, mmf_stream_t* dump_from)
{
    static int8_t I_Frame_index = -1;
    if(dump_to == NULL || dump_from == NULL){
        return IMG_VENC_ERROR;
    }
    // debug("[kvmv]dump_from->count = %d\n", dump_from->count);
    if (dump_from->count == 3) {
        // Neither branch used to test its allocation, and memory is what
        // this board runs out of first.
        for(int i = 0; i < dump_from->count; i++){
            if(dump_from->data[i] == NULL || dump_from->data_size[i] <= 0){
                return IMG_VENC_ERROR;
            }
        }
        uint64_t total_size = (uint64_t)dump_from->data_size[0] +
                              (uint64_t)dump_from->data_size[1] +
                              (uint64_t)dump_from->data_size[2];
        if(total_size > UINT32_MAX || !reserve_save_buffer(dump_to, (uint32_t)total_size)){
            return IMG_BUFFER_FULL;
        }
        dump_to->img_data_size = (uint32_t)total_size;
        dump_to->img_data_type = IMG_H264_TYPE_IF;
        memcpy(dump_to->p_img_data, dump_from->data[0], dump_from->data_size[0]);
        memcpy(dump_to->p_img_data+dump_from->data_size[0], dump_from->data[1], dump_from->data_size[1]);
        memcpy(dump_to->p_img_data+dump_from->data_size[0]+dump_from->data_size[1], dump_from->data[2], dump_from->data_size[2]);

        debug("[kvmv]SPS size = %d\n", dump_from->data_size[0]);
        debug("[kvmv]PPS size = %d\n", dump_from->data_size[1]);
        debug("[kvmv]I-Frame size = %d\n", dump_from->data_size[2]);

        return IMG_H264_TYPE_IF;

    } else if (dump_from->count == 1) {
        // debug("[kvmv]dump P-Frame\r\n");
        I_Frame_index = -1;
        if(dump_from->data[0] == NULL || dump_from->data_size[0] <= 0){
            return IMG_VENC_ERROR;
        }
        if(!reserve_save_buffer(dump_to, (uint32_t)dump_from->data_size[0])){
            return IMG_BUFFER_FULL;
        }
        dump_to->img_data_size = dump_from->data_size[0];
        dump_to->img_data_type = IMG_H264_TYPE_PF;
        memcpy(dump_to->p_img_data, dump_from->data[0], dump_from->data_size[0]);
        return IMG_H264_TYPE_PF;
    } else {
        debug("[kvmv]venc error!\r\n");
        return IMG_VENC_ERROR;
    }
}

void set_h264_gop(uint8_t _gop)
{
    kvm_venc.enc_h264_init = 0; // call
    kvmvenc_gop = maxmin_data(100, 1, (int)_gop);
    debug("[kvmv] set_h264_gop = %d\n", kvmvenc_gop);
}

// Change the frame rate the encoder is configured for. The next frame
// rebuilds the channel, which is what init_venc_h264 does when
// enc_h264_init is clear.
//
// Unlike set_h264_gop this returns early when nothing changed. The server
// calls it whenever a stream starts, and tearing the encoder down and
// building it again to arrive at the value it already had would cost a
// keyframe every time.
void set_h264_fps(uint8_t _fps)
{
    uint8_t fps = maxmin_data(60, 10, (int)_fps);
    if (fps == kvmvenc_fps) {
        return;
    }

    kvmvenc_fps = fps;
    kvm_venc.enc_h264_init = 0;
    debug("[kvmv] set_h264_fps = %d\n", kvmvenc_fps);
}

// Tell the capture channel how many frames a second anyone is going to
// take from it.
//
// The VPSS channel came up asking for 60 in and 60 out. Frame rate control
// only drops when the destination is below the source, so it dropped
// nothing, and no setting of any kind could make it drop anything. The
// stream loop reads at the configured rate, which defaults to 30.
//
// Measured on 2026-09-04 at 1080p, from the CHN OUTPUT RESOLUTION block of
// /proc/cvitek/vpss: the channel hands out 31 frames a second with the rate
// set to 30, against a source delivering about 50. Each frame it no longer
// hands out is 3.1MB of NV21 that is not written, so roughly 59MB/s of
// memory bandwidth. None of it shows in a CPU figure.
//
// Read that block and not VIDevFPS. VIDevFPS is the rate the VI device
// delivers, upstream of this control, and it does not move when the channel
// rate changes. It cannot show the effect of this either way.
//
// The channel is not touched from here. A resolution change rebuilds it
// from the detector thread, and CVI_VPSS_SetChnAttr from this thread while
// that is in flight is a race worth not having. The next read applies it
// under vi_mutex instead, which is one frame away.
void set_capture_fps(uint8_t _fps)
{
    uint8_t fps = maxmin_data(60, 10, (int)_fps);
    if (fps == kvmvi_dst_fps) {
        return;
    }

    kvmvi_dst_fps = fps;
    kvmvi_fps_pending = 1;
    debug("[kvmv] set_capture_fps = %d\n", kvmvi_dst_fps);
}

void set_frame_detact(uint8_t _frame_detact)
{
    uint8_t frame_detact = maxmin_data(100, 0, (int)_frame_detact);
    kvmv_cfg.frame_detact = frame_detact;
    debug("[kvmv] set_frame_detact = %d\n", kvmv_cfg.frame_detact);
    kvmv_cfg.stream_stop = 0;
}

int8_t frame_to_h264(uint8_t *data, int width, int height, int format, int vi_ch,
    kvmv_data_t* ret_stream, uint16_t _qlty)
{
	uint64_t __attribute__((unused)) start_time;
	uint64_t __attribute__((unused)) frame_time;
	int8_t ret = 0;
	static uint8_t P_Frame_Count = 0;
		// start_time = time::time_ms();
		// log::info("getimg: %d \r\n", (int)(time::time_ms()));
    mmf_stream_t _stream = {0};
    if(kvm_venc.enc_h264_init != 1 || width != kvm_venc.kvm_venc_cfg.w || height != kvm_venc.kvm_venc_cfg.h || _qlty != kvm_venc.kvm_venc_cfg.bitrate){
		debug("[kvmv]init_venc_h264	enc_h264_init = %d; width = %d | %d height = %d | %d \n",
				kvm_venc.enc_h264_init,
				width, kvm_venc.kvm_venc_cfg.w,
				height, kvm_venc.kvm_venc_cfg.h);

		init_venc_h264(width, height, _qlty);
        debug("[kvmv]init_venc_h264 finish enc_h264_init = %d; width = %d | %d height = %d | %d \n",
				kvm_venc.enc_h264_init,
				width, kvm_venc.kvm_venc_cfg.w,
				height, kvm_venc.kvm_venc_cfg.h);
        // if(kvm_venc.enc_h264_init == 1){
		// 	init_venc_h264(width, height, _qlty);
		// } else {
		// 	init_venc_h264(default_vpss_width, default_vpss_height, default_h264_qlty);
		// }
    }
	// log::info("init(): %d \r\n", (int)(time::time_ms() - start_time));
    int push_ret = vi_ch >= 0
        ? mmf_venc_push_vi(kvm_venc.mmf_venc_chn, vi_ch)
        : mmf_venc_push(kvm_venc.mmf_venc_chn, data, width, height, format);
    if (push_ret) {
        mmf_del_venc_channel(kvm_venc.mmf_venc_chn);
        // The only exit from here that mmf_venc_free does not cover. It
        // returns early while the channel is not running, and the channel
        // never started, so the frame would be held until the VB pool ran
        // out.
        if (vi_ch >= 0) {
            mmf_vi_frame_release(vi_ch);
        }
        kvm_venc.enc_h264_init = 0;
        // rtmp->unlock();
		debug("[kvmv]mmf venc push failed!\n");
        // err::check_raise(err::ERR_RUNTIME, "mmf venc push failed!\r\n");
        return -1;
    }
	// log::info("push(): %d \r\n", (int)(time::time_ms() - start_time));
    if (mmf_venc_pop(kvm_venc.mmf_venc_chn, &_stream)) {
        // log::error("mmf_venc_pop failed\n");
        //
        // This exit releases the VI frame, although nothing here calls
        // mmf_vi_frame_release. The push above succeeded, so the encoder
        // channel is running, and mmf_venc_free releases every held VI frame
        // whenever it finds the channel in that state. The push failure path
        // a few lines up has to release by hand for the opposite reason: it
        // returns while the channel never started, so mmf_venc_free returns
        // without releasing anything there.
        //
        // The asymmetry is deliberate. A release added here would be a
        // second release of a block from a pool of two.
        mmf_venc_free(kvm_venc.mmf_venc_chn);
        mmf_del_venc_channel(kvm_venc.mmf_venc_chn);
        kvm_venc.enc_h264_init = 0;
		debug("[kvmv]mmf venc pop failed!\n");
        // rtmp->unlock();
        return -1;
    }
	// log::info("pop(): %d \r\n", (int)(time::time_ms() - start_time));
    ret = h264_stream_dump(ret_stream, &_stream);
    mmf_venc_free(kvm_venc.mmf_venc_chn);
	// log::info("dump(): %d \r\n", (int)(time::time_ms() - start_time));

	// debug("[kvmv]_stream.data[0][4] = %d;\n", _stream.data[0][4]);
	debug("[kvmv]Frame size = %d;\n", ret_stream->img_data_size);

    if(ret == IMG_H264_TYPE_IF){
		debug("[kvmv]================ GOP = %d ================\n", kvm_venc.kvm_venc_cfg.gop);
		debug("[kvmv]SPS; PPS; I-Frame, I-Frame size = %d\n", ret_stream->img_data_size);
		P_Frame_Count = 0;

    } else if(ret == IMG_H264_TYPE_PF){
		debug("[kvmv]P-Frame size = %d, P-count = %d\n", ret_stream->img_data_size, P_Frame_Count);
		P_Frame_Count++;
    }
    debug("[kvmv]dump ret = %d\n", ret);
    return ret;
}

void kvmv_init(uint8_t _debug_info_en)
{
    pthread_mutex_init(&vi_mutex, NULL);
    if(_debug_info_en == 0) debug_en = 0;
    else                    debug_en = 1;

    // debug("[kvmv]kvmv_init - 1\r\n");

    cam->hmirror(1);
    cam->vflip(1);
    cam->restart(default_vpss_width, default_vpss_height, image::FMT_YVU420SP);
    for(int i = 0; i < kvmv_data_buffer_size; i++){
        kvmv_data_buffer[i].p_img_data = NULL;
        kvmv_data_buffer[i].img_data_capacity = 0;
        kvmv_data_buffer[i].in_use = 0;
    }

    __atomic_store_n(&kvmv_cfg.try_exit_thread, 0, __ATOMIC_RELEASE);
    // debug("[kvmv]kvmv_init - 2\r\n");

    if(kvmv_cfg.thread_is_running == 1){
        debug("[kvmv]thread is running!\r\n");
    } else {
        if (0 != pthread_create(&kvmv_cfg.detect_thread, NULL, vi_subsystem_detection, NULL)) {
            debug("[kvmv]create vi_subsystem_detection thread failed!\r\n");
            // return -1;
        } else {
            kvmv_cfg.detect_thread_valid = 1;
        }

        if (0 != pthread_create(&kvmv_cfg.watchdog_thread, NULL, watchdog_sf_feed, NULL)) {
            debug("[kvmv]create watchdog_sf_feed thread failed!\r\n");
            // return -1;
        } else {
            kvmv_cfg.watchdog_thread_valid = 1;
        }
    }
    // debug("[kvmv]kvmv_init - 3\r\n");
}

uint8_t check_kvmv(uint8_t _try_num)
{
    if(__atomic_load_n(&kvmv_cfg.hdmi_cable_state, __ATOMIC_ACQUIRE) == 0){
        debug("[kvmv]HDMI Cable not exist!\n");
        return 0;
    }
    if(_try_num >= KVMV_MAX_TRY_NUM){
        debug("[kvmv]try_num >= KVMV_MAX_TRY_NUM!\n");
        return 0;
    }
    // if(sync_vi_res() != 0){
    if(kvmv_cfg.reopen_cam_flag == 1){
        // vi size changed
        kvmv_cfg.reopen_cam_flag = 0;
        // cam->open(kvmv_cfg.vpss_width, kvmv_cfg.vpss_height, image::FMT_YVU420SP, 3);
        cam->restart(default_vpss_width, default_vpss_height, image::FMT_YVU420SP);
        debug("[kvmv]vi size changed, try again\n");
        return 1;
    }
    debug("[kvmv]just try again\n");
    return 1;
}

void set_venc_auto_recyc(uint8_t _enable)
{
    if(_enable) kvmv_cfg.venc_auto_recyc = 1;
    else kvmv_cfg.venc_auto_recyc = 0;
}

/**********************************************************************************
 * @name    kvmv_read_img
 * @author  Sipeed BuGu
 * @date    2024/10/25
 * @version R1.0
 * @brief   Acquire the encoded image with auto init
 * @param	_width				@input: 	Output image width
 * @param	_height				@input: 	Output image height
 * @param	_type				@input: 	Encode type
 * @param	_qlty				@input: 	MJPEG: (50-100) | H264:  (500-10000)
 * @param	_pp_kvm_data		@output: 	Encode data
 * @param	_p_kvmv_data_size	@output: 	Encode data size
 * @return
        -7: HDMI INPUT RES ERROR
        -6: Unsupported resolution, please modify it in the host settings.
        -5: Retrieving image, please wait
        -4: Modifying image resolution, please wait
        -3: img buffer full
        -2: VENC Errorl
        -1: No images were acquired
         0: Acquire MJPEG encoded images
         1: Acquire H264 encoded images(SPS)[Deprecated]
         2: Acquire H264 encoded images(PPS)[Deprecated]
         3: Acquire H264 encoded images(I)
         4: Acquire H264 encoded images(P)
         5: IMG not changed
 **********************************************************************************/
int kvmv_read_img(uint16_t _width, uint16_t _height, uint8_t _type, uint16_t _qlty, uint8_t** _pp_kvm_data, uint32_t* _p_kvmv_data_size)
{
    *_pp_kvm_data = NULL;
    *_p_kvmv_data_size = 0;
    static uint8_t frame_undetact_count = 0;
	// uint64_t __attribute__((unused)) start_time = time::time_ms();
    debug("[kvmv]kvmv_read_img type = %d...\n", _type);
    struct timespec ts;
    clock_gettime(CLOCK_REALTIME, &ts);
    ts.tv_sec += 1;
    // pthread_mutex_lock(&vi_mutex);         // Add lock
    int mutex_res = pthread_mutex_timedlock(&vi_mutex, &ts);
    if(mutex_res != 0){
        return -5;
    }
    if (kvmv_cfg.hdmi_res_err == ERROR_RES){
        pthread_mutex_unlock(&vi_mutex);
        return -7;
    }
    if (kvmv_cfg.hdmi_res_type == UNSUPPORT_RES){
        pthread_mutex_unlock(&vi_mutex);
        return -6;
    }
    if (kvmv_cfg.vi_detect_state == 1){
        pthread_mutex_unlock(&vi_mutex);
        return -4;
    }
    uint8_t try_num = 0;
    do {
        if(kvmv_cfg.vpss_width != _width || kvmv_cfg.vpss_height != _height){
            if(_width == 0 || _height == 0){
                // Follow the HDMI output
                kvmv_cfg.vpss_width = kvmv_cfg.vi_width;
                kvmv_cfg.vpss_height = kvmv_cfg.vi_height;

                if(kvmv_cfg.Auto_res == 0){
                    kvmv_cfg.Auto_res = 1;
                    cam->set_resolution(kvmv_cfg.vpss_width, kvmv_cfg.vpss_height);
                    kvmv_cfg.reinit_flag = 1;
                }

            } else {
                kvmv_cfg.Auto_res = 0;
                // Set the output
                kvmv_cfg.vpss_width = _width;
                kvmv_cfg.vpss_height = _height;

                cam->set_resolution(kvmv_cfg.vpss_width, kvmv_cfg.vpss_height);
                kvmv_cfg.reinit_flag = 1;
            }
        }

        //
        if (kvmv_cfg.reinit_flag == 1) {
            cam->hmirror(1);
            cam->vflip(1);
			kvmv_cfg.reinit_flag = 0;
        }

        // mmf_vi_set_chn_fps records the rate even when no channel is open,
        // and a rebuild reads it back, so a resolution change does not
        // quietly return the board to 60. That is what makes this safe and
        // not the lock: vi_mutex is taken here and nowhere else, so it does
        // not order this against the detection thread's own rebuilds.
        if (kvmvi_fps_pending != 0) {
            kvmvi_fps_pending = 0;
            mmf_vi_set_chn_fps(cam->get_channel(), (int)kvmvi_dst_fps);
        }
        // debug("[kvmv]befor read img: %d \r\n", (int)(time::time_ms() - start_time));
        //
        // No path here reads a pixel unless it is about to sample one.
        // cam->read() maps the frame and invalidates the cache over all of
        // it, which is 3.1MB at 1080p in NV12, then copies the whole frame
        // into an image::Image, and to_jpeg copied it a third time into the
        // encoder. Take the frame where the hardware left it and let the
        // encoder read it there.
        //
        // MJPEG used to be held out of this whenever frame detect was on,
        // because frame_changed() reads pixels. It reads them on one frame
        // in frame_detact, so that held out the other fifty-nine for
        // nothing. Measured on 2026-09-04 at 1080p: 95.2% of core against
        // 91.9% with detection off, while delivering 12.7% fewer frames,
        // which is 18.6% more of the core per frame that reached a viewer.
        // The map now happens on the frame that gets sampled, and there
        // only.
        //
        int native_vi_ch = cam->get_channel();
        int native_len = 0;
        int native_width = 0;
        int native_height = 0;
        int native_format = 0;
        if (mmf_vi_frame_pop_native(native_vi_ch, &native_len, &native_width,
                &native_height, &native_format) != 0) {
            native_vi_ch = -1;
        } else {
            // The channel is alive. Whatever the encoder makes of this frame
            // is a separate question, so this is the only place that says so.
            wedge_note_frame(&vi_wedge);
        }
        // debug("[kvmv]read img: %d \r\n", (int)(time::time_ms() - start_time));

        if(native_vi_ch >= 0){
            if(kvmv_cfg.fresh_frame_count != 0){
                // Do not restart VI after HDMI idle. Reopening the MMF
                // channel can exhaust the carveout heap when the detector
                // thread is also transitioning. Consume queued frames
                // instead; the camera buffer contains at most three frames.
                kvmv_cfg.fresh_frame_count--;
                mmf_vi_frame_release(native_vi_ch);
                continue;
            }

            // frame detect
            if(_type == VENC_MJPEG && kvmv_cfg.frame_detact != 0){
                if(kvmv_cfg.stream_stop == 0){
                    frame_undetact_count++;
                } else {
                    frame_undetact_count = kvmv_cfg.frame_detact;
                }

                if(frame_undetact_count == kvmv_cfg.frame_detact){
                    frame_undetact_count = 0;

                    // The one place that reads the frame, so the one
                    // place that maps it. The encoder below takes the
                    // same frame either way: it is sent by physical
                    // address and does not care whether a mapping exists.
                    const uint8_t *pixels =
                        (const uint8_t *)mmf_vi_frame_map(native_vi_ch);
                    if(frame_changed(pixels, native_len) == 0){
                        debug("[kvmv]frame not changed...\n");
                        kvmv_cfg.stream_stop = 1;
                        mmf_vi_frame_release(native_vi_ch);
                        pthread_mutex_unlock(&vi_mutex);
                        return 5;
                    } else {
                        kvmv_cfg.stream_stop = 0;
                    }
                }
            }
        } else {
            debug("[kvmv]can`t get img...\n");
            continue;
            // pthread_mutex_unlock(&vi_mutex);
            // return IMG_NOT_EXIST;
        }
        // debug("[kvmv]cheak img null?: %d \r\n", (int)(time::time_ms() - start_time));

        // img exist
        // Encode
        if(kvmv_cfg.venc_type == VENC_MJPEG && kvmv_cfg.venc_type != _type){
            if(kvmv_cfg.venc_auto_recyc == 1){
                mmf_enc_jpg_deinit(0);
            }
            kvm_venc.enc_h264_init = 1;
        }
        if(kvmv_cfg.venc_type == VENC_H264 && kvmv_cfg.venc_type != _type){
            if(kvmv_cfg.venc_auto_recyc == 1){
                mmf_del_venc_channel(kvm_venc.mmf_venc_chn);
            }
            kvm_venc.enc_h264_init = 0;
        }

        kvmv_cfg.venc_type = _type;

        if(kvmv_cfg.venc_type == VENC_MJPEG){
            kvmv_data_t* p_kvmv_data = get_save_buffer();
            if(p_kvmv_data == NULL){
                // buffer full
                mmf_vi_frame_release(native_vi_ch);
                debug("[kvmv]jpg buffer full\n");
                pthread_mutex_unlock(&vi_mutex);
                return IMG_BUFFER_FULL;
            }
            int jpeg_ret = frame_to_jpeg(native_vi_ch, p_kvmv_data,
                maxmin_data(99, 51, (int)_qlty));
            if (jpeg_ret < 0) {
                pthread_mutex_unlock(&vi_mutex);
                return jpeg_ret;
            }
            *_pp_kvm_data = p_kvmv_data->p_img_data;
            *_p_kvmv_data_size = p_kvmv_data->img_data_size;
            pthread_mutex_unlock(&vi_mutex);
            return jpeg_ret;
        } else if (kvmv_cfg.venc_type == VENC_H264){
            int ret;
            kvmv_data_t* p_kvmv_data = get_save_buffer();
            if(p_kvmv_data == NULL){
                // buffer full
                mmf_vi_frame_release(native_vi_ch);
                *_pp_kvm_data = NULL;
                pthread_mutex_unlock(&vi_mutex);
                return IMG_BUFFER_FULL;
            }
            // debug("[kvmv]get_save_buffer: %d \r\n", (int)(time::time_ms() - start_time));
            ret = frame_to_h264(NULL, native_width, native_height, native_format,
                native_vi_ch, p_kvmv_data, maxmin_data(10000, 500, (int)_qlty));
            // debug("[kvmv]venc frame_to_h264: %d \r\n", (int)(time::time_ms() - start_time));
            if(ret < 0){
                release_save_buffer(p_kvmv_data);
                *_pp_kvm_data = NULL;
                *_p_kvmv_data_size = 0;
                pthread_mutex_unlock(&vi_mutex);
                return ret;
            }
            *_pp_kvm_data = p_kvmv_data->p_img_data;
            *_p_kvmv_data_size = p_kvmv_data->img_data_size;
            pthread_mutex_unlock(&vi_mutex);
            return ret;
        } else {
            // Unreachable while the callers only ask for the two types
            // they have always asked for, and it stops being harmless
            // the moment that changes. The frame taken above is a block
            // from a pool of two, so falling through to the retry while
            // still holding it stalls the pipeline rather than leaking a
            // little heap, which is what the same fall-through cost when
            // the frame was an image::Image.
            mmf_vi_frame_release(native_vi_ch);
            debug("[kvmv]unknown encode type %d\n", (int)_type);
            pthread_mutex_unlock(&vi_mutex);
            return IMG_VENC_ERROR;
        }
    } while (check_kvmv(try_num++));
    // debug("[kvmv]return: %d \r\n", (int)(time::time_ms() - start_time));
    //
    // Nothing came out of the channel for this whole call.
    // kvmv_hdmi_signal_active is what separates a wedged channel from a
    // source that stopped: it is 0 exactly when the VI device reports no
    // frames of its own.
    if (wedge_note_no_frame(&vi_wedge, vi_state_shared::monotonic_ms(),
            kvmv_hdmi_signal_active()) != 0) {
        // Not debug(). This fires at most once per cooldown and only in a
        // fault that otherwise leaves the board blank with no explanation,
        // so it has to reach the log of a device nobody is debugging.
        // debug() would also print the format string raw: it takes its
        // arguments and passes only the format to printf.
        printf("[kvmv]channel handed out nothing %u times with the VI device live, asking for a rebuild\n",
            (unsigned int)vi_wedge.fail_count);
        __atomic_store_n(&vi_wedge_rebuild_request, 1, __ATOMIC_RELEASE);
    }
    *_pp_kvm_data = NULL;
    pthread_mutex_unlock(&vi_mutex);
    return IMG_NOT_EXIST;
}

// Give the slot back. The allocation stays where it is, ready for the next
// frame that fits in it.
//
// This runs on the reader thread while a second reader can be inside
// kvmv_read_img holding vi_mutex, and it deliberately takes no lock. The
// slot named here is in_use and belongs to this caller, so nothing else can
// reallocate it or claim it; the pointers of the other slots are only ever
// compared, and a live allocation cannot equal another live allocation.
// The one ordering that matters is reading the type before clearing in_use:
// the other way round, a reader could claim the slot in between and reset
// the type to 0 under us. Locking here would instead put a frame release
// behind whatever encoder call the other reader is waiting on, which can be
// a full second.
int free_kvmv_data(uint8_t ** _pp_kvm_data)
{
    if(_pp_kvm_data == NULL || *_pp_kvm_data == NULL){
        return IMG_NOT_EXIST;
    }
        // debug("[kvmv]free_kvmv_data - 1\r\n");
    for(int i = 0; i < kvmv_data_buffer_size; i++){
        if(*_pp_kvm_data == kvmv_data_buffer[i].p_img_data){
            // debug("[kvmv]free buffer : %d\n", *_pp_kvm_data);
            uint8_t _type = kvmv_data_buffer[i].img_data_type;
            kvmv_data_buffer[i].in_use = 0;
            return _type;
        }
    }
    return IMG_NOT_EXIST;
}

void free_all_kvmv_data()
{
    for(int i = 0; i < kvmv_data_buffer_size; i++){
        if(kvmv_data_buffer[i].p_img_data != NULL){
            free(kvmv_data_buffer[i].p_img_data);
            kvmv_data_buffer[i].p_img_data = NULL;
        }
        kvmv_data_buffer[i].img_data_capacity = 0;
        kvmv_data_buffer[i].in_use = 0;
    }
}

void kvmv_deinit()
{
    set_hdmi_capture_enabled(0);

    // Order matters, and it used to be wrong. vi_mutex was destroyed first,
    // before the threads were even told to stop, so a thread could be holding
    // or waiting on it at that moment - and kvmv_read_img takes the same mutex.
    // Destroying a locked mutex is undefined, and so is locking a destroyed one.
    //
    // Ask both threads to stop, wait for them, and only then take the pipeline
    // apart. Both break on try_exit_thread and the slower one polls every 500ms,
    // so this returns in well under a second.
    //
    // That last sentence was false for as long as the flag was a plain field:
    // the compiler read it once per thread and neither loop ever looked again,
    // so both joins here waited for a change they could not see. The store and
    // the two tests are atomic now, and the note beside the field records what
    // it cost.
    //
    // Waiting also removes a second fault. vi_subsystem_detection clears
    // thread_is_running as it leaves, and kvmv_init reads that flag to decide
    // whether to start the threads again. Without the join, an init that arrived
    // while the old thread was still exiting would skip creating a new one and
    // then watch the old one die, leaving the board with no HDMI detection until
    // the process restarted.
    __atomic_store_n(&kvmv_cfg.try_exit_thread, 1, __ATOMIC_RELEASE);

    if (kvmv_cfg.detect_thread_valid) {
        pthread_join(kvmv_cfg.detect_thread, NULL);
        kvmv_cfg.detect_thread_valid = 0;
    }
    if (kvmv_cfg.watchdog_thread_valid) {
        pthread_join(kvmv_cfg.watchdog_thread, NULL);
        kvmv_cfg.watchdog_thread_valid = 0;
    }

    pthread_mutex_destroy(&vi_mutex);
    cam->close();

    // Force, because the refcount this used to decrement does not come back to
    // zero once a browser has asked for MJPEG.
    //
    // mmf_deinit is mmf_try_deinit(false): it decrements mmf_used_cnt and does
    // the actual teardown only on the call that reaches zero. _mmf_deinit is
    // that teardown, and it is the only caller of mmf_del_venc_channel_all, so
    // it is the only thing that ever gives the H.264 encoder back.
    //
    // MJPEG is served through mmf_enc_jpg_push_vi_with_quality, which reaches
    // mmf_enc_jpg_init, which takes a reference. Nothing returns it: mmf_enc_jpg_deinit runs from
    // _mmf_deinit, and from the mode switch below only when venc_auto_recyc is
    // set, which nothing sets. So a session that ever served one JPEG frame
    // leaves the count one too high, every later stop decrements to one, and
    // the encoder stays allocated. Measured at 11,636,736 bytes for the whole
    // channel plus one ISP_SHARED_BUFFER_0, and the carveout is 75MB.
    //
    // Nothing else recovers it. The vcodec driver builds with
    // CVI_H26X_USE_ION_FW_BUFFER, which compiles vpu_free_buffers out of
    // vpu_release, and soph_sys drops bindings when its fd closes but never
    // walks its buffer list. What this call does not release stays allocated
    // until the board reboots.
    //
    // Forcing is right rather than merely convenient: kvm_vision is the only
    // user of MMF in this process, so at this point the count describes
    // references nobody holds. kvmv_init opens the camera again, which calls
    // mmf_init again, so a later start is unaffected - and it has to be,
    // because the idle timeout stops capture at runtime, not only at exit.
    //
    // tools/vidiag/venc-leak.cpp measures each step of this on the device.
    mmf_try_deinit(true);
    free_all_kvmv_data();
}

uint8_t kvmv_hdmi_control(uint8_t _en)
{
    if(_en == 0){
        set_hdmi_capture_enabled(0);
    }

    if(kvmv_cfg.hw_version == 0){

        FILE *fp;
	    uint8_t RW_Data[2];
        if(access("/etc/kvm/hw", F_OK) == 0){
            fp = fopen("/etc/kvm/hw", "r");
            fread(RW_Data, sizeof(char), 1, fp);
            fclose(fp);
            switch(RW_Data[0]){
                case 'a':
                case 'b':
                case 'p':
                    kvmv_cfg.hw_version = RW_Data[0];
                    break;
                default :
                    kvmv_cfg.hw_version = 'a';
            }
        }
    }
    if(kvmv_cfg.hw_version != 'p'){
        debug("[kvmv]Hardware not support!\n");
        if(_en != 0){
            set_hdmi_capture_enabled(1);
        }
        return -1;
    }
    if(access("/sys/class/gpio/gpio451/value", F_OK) != 0){
        system("echo 451 > /sys/class/gpio/export");
        system("echo out > /sys/class/gpio/gpio451/direction");
    }
    if(_en == 0){
        kvmv_cfg.hdmi_stop_flag = 1;
        while(kvmv_cfg.hdmi_reading_flag == 1) time::sleep_ms(10);
        system("echo 0 > /sys/class/gpio/gpio451/value");
        return 0;
    } else {
        kvmv_cfg.hdmi_stop_flag = 0;
        system("echo 1 > /sys/class/gpio/gpio451/value");

        // Keep the existing VI channel and consume queued frames after idle.
        // Reopening it here or from kvmv_read_img can exhaust the carveout
        // heap while the HDMI detector is transitioning.
        kvmv_cfg.fresh_frame_count = fresh_frame_discard_count;
        set_hdmi_capture_enabled(1);
        return 0;
    }
    return -1;
}

uint8_t kvmv_hdmi_signal_active()
{
    uint8_t enabled = __atomic_load_n(&hdmi_capture_enabled, __ATOMIC_ACQUIRE);
    uint8_t active = __atomic_load_n(&hdmi_signal_active, __ATOMIC_ACQUIRE);
    return enabled != 0 && active != 0 ? 1 : 0;
}
