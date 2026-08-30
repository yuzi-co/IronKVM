#include "oled_ui.h"

using namespace maix;
using namespace maix::sys;

extern kvm_sys_state_t kvm_sys_state;
extern kvm_oled_state_t kvm_oled_state;

void kvm_init_cube_ui(void)
{
	uint8_t temp;
	char* str_temp = "192.168.1.243";

	OLED_Clear();
	// OLED_Revolve();
	OLED_ShowKVMLogo();
	OLED_ShowLogo();
	OLED_ShowKVMState(HDMI_STATE, 	0);
	OLED_ShowKVMState(HID_STATE, 	0);
	OLED_ShowKVMState(ETH_STATE, 	0);
	OLED_ShowKVMState(WIFI_STATE, 	0);
	OLED_Showline();
	OLED_ShowKVMStreamState(KVM_INIT, &temp);
	temp = 0;
	OLED_ShowKVMStreamState(KVM_ETH_IP, &temp);
	temp = KVM_RES_none;
	OLED_ShowKVMStreamState(KVM_HDMI_RES, &temp);
	temp = KVM_TYPE_none;
	OLED_ShowKVMStreamState(KVM_STEAM_TYPE, &temp);
	temp = 0;
	OLED_ShowKVMStreamState(KVM_STEAM_FPS, &temp);
	temp = 80;
	OLED_ShowKVMStreamState(KVM_JPG_QLTY, &temp);
}

void kvm_init_pcie_ui(void)
{
	OLED_Revolve();
	OLED_Showline_1();
	OLED_ShowLogo();
	OLED_ShowKVMState(HDMI_STATE, 	0);
	OLED_ShowKVMState(HID_STATE, 	0);
	OLED_ShowKVMState(ETH_STATE, 	0);
	OLED_ShowKVMState(WIFI_STATE, 	0);

	// OLED_ShowString(0, 2, "!192.168.222.197", 4);
	// OLED_ShowString(1, 2, "  192.168.2.197", 4);
	// OLED_ShowString(1, 3, "1920*1080", 4);
	// OLED_ShowString(41, 3, "  FPS", 4);

	// OLED_ShowString_AlignRight(AlignRightEND_P, 2, "192.168.0.69", 4);
	// OLED_ShowString_AlignRight(37, 3, "1920*1080", 4);
	OLED_ShowString_AlignRight(AlignRightEND_P, 3, "  FPS", 4);
	// OLED_ShowString(0, 3, "\"ABCDEFGHIJKLMMN\'", 4);
	// OLED_ShowString(5, 2, "1", 4);
}

void kvm_init_ui(void)
{
	if(kvm_hw_ver != 2){
		kvm_init_cube_ui();
	} else {
		kvm_init_pcie_ui();
	}
}

void qrcode_to_oled(QRCode *qr)
{
	char *p_oled_data;
	uint16_t count = 0;
	uint8_t bit;
	uint8_t begin_x = 2;
	uint8_t begin_y = 2;
	p_oled_data = (char *)malloc( 132 * sizeof(char));
	if(p_oled_data != NULL){
		// fill
		for(int i = 0; i < 132; i++){
			p_oled_data[i] = 0xFF;
		}
		// i | ; j -
		for(int i = 0; i < 29; i++){
			for(int j = 0; j < 29; j++){
				if(qrIsBlacke(qr, i, j)){
					// 敲黑点
					uint16_t data_index = ((i+begin_y)/8)*33+(j+begin_x);
					uint8_t mask = ~(0x01 << ((i+begin_y)%8));
					p_oled_data[data_index] &= mask;
				}
			}
		}
		OLED_Fill();
		OLED_ShowIMG(29, 0, p_oled_data, 33, 4);
		free(p_oled_data);
	}
}

int qrencode(char *string)
{
	// test: qrencode("WIFI:T:WPA2;S:NanoKVM;P:12345678;;");
	int errcode = QR_ERR_NONE;
	QRCode* p = qrInit(3, QR_EM_8BIT, 1, 4, &errcode);
	if (p == NULL) {
		printf("error\n");
		return -1;
	}
	qrAddData(p, (const qr_byte_t*)string, strlen(string));
	if (!qrFinalize(p)) {
		printf("finalize error\n");
		return -1;
	}

	qrcode_to_oled(p);
	if(string[0] == 'W'){
		OLED_ShowStringTurn(3, 1, "WiFi", 8);
		OLED_ShowStringTurn(3, 2, "AP:", 8);
	} else if(string[0] == 'h'){
		if(string[7] == '1'){
			OLED_ShowStringTurn(3, 1, "Web:", 8);
		} else if(string[8] == 'w'){
			OLED_ShowStringTurn(3, 1, "WiKi", 8);
		}
	}
	int size = 0;
	// width = height = qr_vertable[version] * mag + sep * mag * 2
	qr_byte_t * buffer = qrSymbolToBMP(p, 5, 5, &size);
	if (buffer == NULL) {
		printf("error %s", qrGetErrorInfo(p));
		return -1;
	}
	// output qrcode to file
	// ofstream f("/etc/kvm/wifi_config.bmp");
	// if (f.fail()) {
	// 	return -1;
	// }
	// f.write((const char *)buffer, size);
	// f.close();
	return 0;
}

ip_addr_t show_which_ip(void)
{
	if(kvm_sys_state.wifi_state == -2) return ETH_IP;
	if(kvm_oled_state.eth_state == 3 && kvm_oled_state.wifi_state != 1) return ETH_IP;
	if(kvm_oled_state.eth_state != 3 && kvm_oled_state.wifi_state == 1) return WiFi_IP;
	static uint8_t run_count = 0;
	static ip_addr_t ip_type = ETH_IP; 
	run_count++;
	if(run_count > IP_Change_time/STATE_DELAY){
		run_count = 0;
		switch(ip_type){
			case ETH_IP:
				ip_type = WiFi_IP;
				break;
			case WiFi_IP:
				ip_type = ETH_IP;
				break;
			default:
				ip_type = ETH_IP;
		}
	}
	// printf("show_which_ip? %d\n", ip_type);
	return ip_type;
}

uint8_t ip_changed(ip_addr_t ip_type)
{
	uint8_t *kvm_sys_ip;
	uint8_t *kvm_oled_ip;
	uint8_t ret;
	if(ip_type == ETH_IP){
		kvm_sys_ip = kvm_sys_state.eth_addr;
		kvm_oled_ip = kvm_oled_state.eth_addr;
	} else if(ip_type == WiFi_IP){
		kvm_sys_ip = kvm_sys_state.wifi_addr;
		kvm_oled_ip = kvm_oled_state.wifi_addr;
	} else ret = 0;
	for (int i = 0; i < 16; i++){
		if(kvm_sys_ip[i] == 0) ret = 0;
		if(kvm_sys_ip[i] != kvm_oled_ip[i]) ret = 1;
	}
	if(ret == 1){
		memcpy(kvm_oled_ip, kvm_sys_ip, 16);
	}
	return ret;
}

uint8_t kvm_state_is_changed()
{
	if(	(kvm_sys_state.eth_state 	!= kvm_oled_state.eth_state) 	|| 
		(kvm_sys_state.wifi_state 	!= kvm_oled_state.wifi_state) 	|| 
		(kvm_sys_state.usb_state 	!= kvm_oled_state.usb_state) 	|| 
		(kvm_sys_state.hdmi_state 	!= kvm_oled_state.hdmi_state) 	|| 
	/*	(kvm_sys_state.now_fps 		!= kvm_oled_state.now_fps) 		|| */
		(kvm_sys_state.hdmi_width 	!= kvm_oled_state.hdmi_width) 	|| 
		(kvm_sys_state.hdmi_height 	!= kvm_oled_state.hdmi_height) 	|| 
		(kvm_sys_state.type 		!= kvm_oled_state.type) 		|| 
		(kvm_sys_state.qlty 		!= kvm_oled_state.qlty) )
		return 1;
	else
		return 0;
}

void kvm_eth_state_disp(ip_addr_t _ip_type, uint8_t first_disp)
{
	static ip_addr_t _ip_type_old = NULL_IP;
	uint8_t temp;
	// printf("[kvmd]eth_state = %d\n", kvm_sys_state.eth_state);
	if(	(kvm_oled_state.eth_state != kvm_sys_state.eth_state) || 
		(_ip_type_old != _ip_type) || 
		first_disp || ip_changed(ETH_IP))
	{
		// ensure eth_addr is updated
		if (kvm_sys_state.eth_state >= 2) {
			if (kvm_sys_state.eth_addr[0] != 0) 
				kvm_oled_state.eth_state = kvm_sys_state.eth_state;
		} else {
			kvm_oled_state.eth_state = kvm_sys_state.eth_state;
		}
		
		_ip_type_old = _ip_type;
		switch(kvm_oled_state.eth_state){
			case -1:
			case  0:
				temp = 0;
				OLED_ShowKVMState(ETH_STATE, 	0);
				if(_ip_type == ETH_IP)
					OLED_ShowKVMStreamState(KVM_ETH_IP, &temp);
				break;
			case  1:
				OLED_ShowKVMState(ETH_STATE, 	1);
				if(_ip_type == ETH_IP)
					OLED_ShowKVMStreamState(KVM_ETH_IP, &temp);
				break;
			case  2:
				OLED_Show_Network_Error(1);
			case  3:
				OLED_Show_Network_Error(0);
				OLED_ShowKVMState(ETH_STATE, 	1);
				if(_ip_type == ETH_IP)
					OLED_ShowKVMStreamState(KVM_ETH_IP, kvm_sys_state.eth_addr);
				break;
		}
	}
}

void kvm_wifi_state_disp(ip_addr_t _ip_type, uint8_t first_disp)
{
	static ip_addr_t _ip_type_old = NULL_IP;
	uint8_t temp;
	// printf("[kvmd]wifi_state = %d\n", kvm_sys_state.wifi_state);
	if(	(kvm_oled_state.wifi_state != kvm_sys_state.wifi_state) || 
		(_ip_type_old != _ip_type) || 
		first_disp || ip_changed(WiFi_IP))
	{
		kvm_oled_state.wifi_state = kvm_sys_state.wifi_state;
		_ip_type_old = _ip_type;
		switch(kvm_oled_state.wifi_state){
			case -2:
				OLED_ShowKVMState(WIFI_STATE, 	-1);
				break;
			case -1:
			case  0:
				temp = 0;
				OLED_ShowKVMState(WIFI_STATE, 	0);
				if(_ip_type == WiFi_IP)
					OLED_ShowKVMStreamState(KVM_WIFI_IP, &temp);
				break;
			case  1:
				OLED_ShowKVMState(WIFI_STATE, 	1);
				if(_ip_type == WiFi_IP){
					OLED_ShowKVMStreamState(KVM_WIFI_IP, kvm_sys_state.wifi_addr);
				}
				break;
		}
	}
}

void kvm_usb_state_disp(uint8_t first_disp)
{
	if((kvm_oled_state.usb_state != kvm_sys_state.usb_state) || first_disp){
		kvm_oled_state.usb_state = kvm_sys_state.usb_state;
		switch(kvm_oled_state.usb_state){
			case -1:
			case  0:
				OLED_ShowKVMState(HID_STATE, 	0);
				break;
			case  1:
				OLED_ShowKVMState(HID_STATE, 	1);
				break;
		}
	}
}

void kvm_hdmi_state_disp(uint8_t first_disp)
{
	if((kvm_oled_state.hdmi_state != kvm_sys_state.hdmi_state) || first_disp){
		kvm_oled_state.hdmi_state = kvm_sys_state.hdmi_state;
		switch(kvm_oled_state.hdmi_state){
			case -1:
			case  0:
				OLED_ShowKVMState(HDMI_STATE, 	0);
				break;
			case  1:
				OLED_ShowKVMState(HDMI_STATE, 	1);
				break;
		}
	}
}

void kvm_fps_disp(uint8_t first_disp)
{
	if((kvm_oled_state.now_fps != kvm_sys_state.now_fps) || first_disp){
		kvm_oled_state.now_fps = kvm_sys_state.now_fps;
		OLED_ShowKVMStreamState(KVM_STEAM_FPS, &kvm_oled_state.now_fps);
	}
}

void kvm_res_disp(uint8_t first_disp)
{
	if(	(kvm_oled_state.hdmi_width != kvm_sys_state.hdmi_width) || \
		(kvm_oled_state.hdmi_height != kvm_sys_state.hdmi_height) || \
		first_disp ){
		kvm_oled_state.hdmi_width = kvm_sys_state.hdmi_width;
		kvm_oled_state.hdmi_height = kvm_sys_state.hdmi_height;
		OLED_Show_Res(kvm_sys_state.hdmi_width, kvm_sys_state.hdmi_height);
	}
}

void kvm_type_disp(uint8_t first_disp)
{
	if(	(kvm_oled_state.type != kvm_sys_state.type) || first_disp){
		kvm_oled_state.type = kvm_sys_state.type;
		OLED_ShowKVMStreamState(KVM_STEAM_TYPE, &kvm_oled_state.type);
	}
}

void kvm_qlty_disp(uint8_t first_disp)
{
	if(	(kvm_oled_state.qlty != kvm_sys_state.qlty) || first_disp){
		kvm_oled_state.qlty = kvm_sys_state.qlty;
		OLED_ShowKVMStreamState(KVM_JPG_QLTY, &kvm_sys_state.qlty);
	}
}

void kvm_main_disp(uint8_t first_disp)
{
	if(first_disp){
		OLED_Clear();
		kvm_init_ui();
	}
}

void kvm_oled_clear(uint8_t subpage_changed)
{
	if(subpage_changed){
		OLED_Clear();
	}
}


// --- the compact roaming page ---
//
// An OLED pixel dims with the light it has emitted, so a status screen that
// fills the panel and never moves wears its own layout into the glass. On a
// board measured on 2026-08-30 the ghost was still legible with every pixel
// lit, and it read TYPE: H264 while the live screen said MJPG: the wear records
// what the panel showed for the longest, not what it shows now.
//
// The page below answers that in two ways at once. It draws the same facts in
// 84x24 rather than 128x64, so about a third of the pixels carry the load, and
// it moves the block around the panel so no pixel carries it for long.
//
// Moving needs the block to be small first, and both halves of the move are
// done in the drawing rather than in the controller.
//
// The controller can shift vertically by itself with command 0xD3, which costs
// no redraw at all, and the first version of this used it. On the panel the
// block came apart: the shift wraps, and it wraps in the direction opposite to
// the one assumed here, so the bottom line reappeared at the top. Choosing the
// first page instead addresses display RAM directly and cannot wrap, at the
// price of moving in steps of a whole page.
//
// So vertical travel is six positions of 8 rows and horizontal travel is a
// column bias, and a move redraws three pages. At one move per period that is
// nothing: the same three pages are redrawn whenever a value changes anyway.

// The 6x8 font, which is the middle of the three this panel has. 16 columns of
// it is 96 pixels, leaving 32 columns of travel.
//
// The 4 pixel font fits more and was tried first. It is too small to read at
// the distance the panel is actually looked at, and its table stops at ']': no
// lower case, so "Middle" and "1920x1080" could not even be spelled. The 8x16
// font is the other way. Three lines of it would fill 48 of the 64 rows and all
// 128 columns, leaving nothing to move through.
#define COMPACT_FONT    8
#define COMPACT_CHAR_W  6
#define COMPACT_COLS    16
#define COMPACT_W       (COMPACT_COLS * COMPACT_CHAR_W)

// The address is drawn in the 8x16 font instead, because it is the reason
// somebody looks at this panel: it is what they type into a browser. Everything
// else on the screen describes a session they can only start once they have it.
//
// That font is 8 pixels wide and two pages tall, so the address line is as wide
// as 8 times its length and the block is four pages rather than three.
#define COMPACT_ADDR_FONT 16
#define COMPACT_ADDR_CHAR_W 8
#define COMPACT_PAGES   4
#define COMPACT_PAGE_MAX (8 - COMPACT_PAGES)

// The OLED thread runs on a 1000 ms loop, so a period is in passes and in
// seconds at the same time. /etc/kvm/oled_move overrides it.
#define COMPACT_PERIOD_DEFAULT  600
#define COMPACT_PERIOD_MIN      10

// The steps are coprime with the travel they walk, so the block visits many
// positions before it repeats rather than pacing one line.
#define COMPACT_DX_STEP 13
#define COMPACT_PAGE_STEP 5

static uint8_t compact_dx = 0;
static uint8_t compact_page = 0;
static uint16_t compact_move_pass = 0;
static uint32_t compact_move_index = 0;

// The address this page last drew. kvm_state_is_changed does not compare
// addresses, because the full page detects those through ip_changed in its own
// per-field draws, and this page has no per-field draws to hang that on.
static char compact_last_addr[24] = {0};

// The drive currently written to the panel. OLED_Init sets it once, and this
// page re-reads the file so the level can be judged on the panel itself rather
// than through a build. Reading is cheap: the file is a few bytes and the
// kernel has it cached, and it is only read every COMPACT_DRIVE_PASSES passes.
#define COMPACT_DRIVE_PASSES 5
static uint8_t compact_drive = 0;
static uint8_t compact_drive_pass = 0;

// compact_period reads the move period once per pass. Reading it every time
// rather than caching it lets an operator retune the panel without restarting
// the process, and the file is small and in the page cache.
static void compact_remember(void);

static uint16_t compact_period(void)
{
	FILE *fp;
	char buf[16] = {0};
	long value;
	char *end;

	fp = fopen("/etc/kvm/oled_move", "r");
	if(fp == NULL){
		return COMPACT_PERIOD_DEFAULT;
	}
	if(fgets(buf, sizeof(buf), fp) == NULL){
		fclose(fp);
		return COMPACT_PERIOD_DEFAULT;
	}
	fclose(fp);

	value = strtol(buf, &end, 0);
	if(end == buf || value < COMPACT_PERIOD_MIN){
		return COMPACT_PERIOD_DEFAULT;
	}
	if(value > 65535){
		return 65535;
	}
	return (uint16_t)value;
}

// compact_pair lays one line out as a fixed field, left flush and right flush,
// padded with spaces. The padding is what erases the previous value: a shorter
// string drawn over a longer one would otherwise leave its tail behind.
static void compact_pair(char *out, const char *left, const char *right)
{
	size_t l = strlen(left);
	size_t r = strlen(right);

	memset(out, ' ', COMPACT_COLS);
	out[COMPACT_COLS] = '\0';

	if(l > COMPACT_COLS) l = COMPACT_COLS;
	memcpy(out, left, l);

	// The left side wins a collision. The pair that can collide is an address
	// and a frame rate, and an address with a digit missing is worse than no
	// frame rate at all: 192.168.100.100 is 15 of the 16 columns by itself.
	if(r + l + 1 > COMPACT_COLS){
		return;
	}
	memcpy(out + COMPACT_COLS - r, right, r);
}

// compact_flags is one character per subsystem, and a dot for down. Four
// characters carry what four icons used to, and they cost 16 pixels.
//
// W is blank rather than a dot when the board has no wireless at all, because
// wifi_state -2 means the hardware is absent and a dot would read as a fault.
static void compact_flags(char *out)
{
	out[0] = (kvm_sys_state.eth_state >= 1) ? 'E' : '.';
	if(kvm_sys_state.wifi_state == -2){
		out[1] = ' ';
	} else {
		out[1] = (kvm_sys_state.wifi_state == 1) ? 'W' : '.';
	}
	out[2] = (kvm_sys_state.usb_state == 1) ? 'H' : '.';
	out[3] = (kvm_sys_state.hdmi_state == 1) ? 'V' : '.';
	out[4] = '\0';
}

// compact_address prefers the wired address, because that is the one a rack
// cable reaches. The full page alternates between the two every five seconds;
// this one does not, since an address that changes under the reader is worth
// less than one that stays still.
static void compact_address(char *out, size_t n)
{
	if(kvm_sys_state.eth_state == 3 && kvm_sys_state.eth_addr[0] != 0){
		snprintf(out, n, "%s", (char *)kvm_sys_state.eth_addr);
		return;
	}
	if(kvm_sys_state.wifi_state == 1 && kvm_sys_state.wifi_addr[0] != 0){
		snprintf(out, n, "%s", (char *)kvm_sys_state.wifi_addr);
		return;
	}
	snprintf(out, n, "--");
}

static const char *compact_type(void)
{
	switch(kvm_sys_state.type){
		case KVM_TYPE_MJPG: return "MJPG";
		case KVM_TYPE_H264: return "H264";
		default:            return "--";
	}
}

static const char *compact_quality(void)
{
	switch(kvm_sys_state.qlty){
		case 1:  return "LOW";
		case 2:  return "Middle";
		case 3:  return "HIGH";
		case 4:  return "EXTRA";
		default: return "--";
	}
}

// compact_draw writes all three lines. The whole block is redrawn rather than
// the field that changed, because it is 63 characters of a 4 pixel font: about
// 250 I2C writes, against the 1024 that one OLED_Clear costs. Partial updates
// bought their complexity when a redraw meant the whole panel.
// compact_block_w is how wide the drawn block actually is, which decides how
// far it can travel sideways. The address line sets it whenever the address is
// long, so the travel shrinks on a 192.168.100.100 and opens up on a 10.0.0.5
// rather than being fixed at the worst case.
static uint8_t compact_block_w(const char *address)
{
	uint8_t addr_w = (uint8_t)(strlen(address) * COMPACT_ADDR_CHAR_W);

	return (addr_w > COMPACT_W) ? addr_w : COMPACT_W;
}

static void compact_draw(void)
{
	char line[COMPACT_COLS + 1];
	char address[24] = {0};
	char flags[8] = {0};
	char res[16] = {0};
	char detail[COMPACT_COLS + 1] = {0};

	// Clear the block's own pages first. The address is drawn in a wider font
	// and is not padded to a fixed width, so a shorter one would otherwise
	// leave the tail of the last behind, and 10.0.0.5 over 10.0.0.222 reads as
	// a working address that is not this board's.
	OLED_Clear_Pages(compact_page, COMPACT_PAGES);

	compact_address(address, sizeof(address));
	OLED_ShowString(compact_dx, compact_page + 0, address, COMPACT_ADDR_FONT);

	if(kvm_sys_state.hdmi_width != 0 || kvm_sys_state.hdmi_height != 0){
		snprintf(res, sizeof(res), "%dx%d",
			kvm_sys_state.hdmi_width, kvm_sys_state.hdmi_height);
	} else {
		snprintf(res, sizeof(res), "--");
	}
	compact_flags(flags);
	compact_pair(line, flags, res);
	OLED_ShowString(compact_dx, compact_page + 2, line, COMPACT_FONT);

	// The frame rate rides with the quality, so a stream that is running says
	// so on the line that describes it.
	if(kvm_sys_state.now_fps > 0){
		snprintf(detail, sizeof(detail), "%s %dF",
			compact_quality(), kvm_sys_state.now_fps);
	} else {
		snprintf(detail, sizeof(detail), "%s", compact_quality());
	}
	compact_pair(line, compact_type(), detail);
	OLED_ShowString(compact_dx, compact_page + 3, line, COMPACT_FONT);

	compact_remember();
}

// compact_remember copies what was just drawn into kvm_oled_state, which is
// what kvm_state_is_changed compares against. The full page does this inside
// each per-field draw. Without it here the first change would leave the two
// structures apart for good, and the page would redraw every second forever.
static void compact_remember(void)
{
	kvm_oled_state.eth_state    = kvm_sys_state.eth_state;
	kvm_oled_state.wifi_state   = kvm_sys_state.wifi_state;
	kvm_oled_state.usb_state    = kvm_sys_state.usb_state;
	kvm_oled_state.hdmi_state   = kvm_sys_state.hdmi_state;
	kvm_oled_state.hdmi_width   = kvm_sys_state.hdmi_width;
	kvm_oled_state.hdmi_height  = kvm_sys_state.hdmi_height;
	kvm_oled_state.type         = kvm_sys_state.type;
	kvm_oled_state.qlty         = kvm_sys_state.qlty;
	kvm_oled_state.now_fps      = kvm_sys_state.now_fps;

	compact_address(compact_last_addr, sizeof(compact_last_addr));
}

// compact_changed adds the address to what kvm_state_is_changed already covers.
static uint8_t compact_changed(void)
{
	char address[24] = {0};

	compact_address(address, sizeof(address));
	if(strcmp(address, compact_last_addr) != 0){
		return 1;
	}

	return kvm_state_is_changed();
}

// compact_move picks the next position. The vertical half is a controller
// register and needs no redraw; the horizontal half is a bias in the drawing,
// so the three pages are blanked first and then drawn again at the new column.
static void compact_move(void)
{
	char address[24] = {0};
	uint8_t dx_max;

	// The pages the block is leaving. compact_draw clears the ones it is
	// arriving at, so between them nothing of the old position survives.
	OLED_Clear_Pages(compact_page, COMPACT_PAGES);

	compact_address(address, sizeof(address));
	dx_max = (uint8_t)(128 - compact_block_w(address));

	compact_move_index++;
	compact_dx = (uint8_t)((compact_move_index * COMPACT_DX_STEP) % (dx_max + 1));
	compact_page = (uint8_t)((compact_move_index * COMPACT_PAGE_STEP) % (COMPACT_PAGE_MAX + 1));

	compact_draw();
}

// compact_page_wanted decides which main page runs.
//
// The compact page is the default on the boards it was drawn for, and
// /etc/kvm/oled_classic brings the full page back without a new binary. That
// escape hatch matters because this replaces the screen an operator knows, and
// a file is cheaper to try than a rebuild.
//
// The PCIe board keeps its own layout either way. It draws through
// kvm_init_pcie_ui with the panel rotated and a different address, and none of
// that has been on a bench here.
static uint8_t compact_page_wanted(void)
{
	if(kvm_hw_ver == 2){
		return 0;
	}
	if(access("/etc/kvm/oled_classic", F_OK) == 0){
		return 0;
	}
	return 1;
}

// kvm_compact_ui_disp is the page. It draws when something it shows has
// changed, and it moves on its own schedule.
void kvm_compact_ui_disp(uint8_t first_disp)
{
	if(first_disp){
		// A previous run of this code may have left a display offset behind,
		// and OLED_Init sets it to zero only when the process restarts. Put it
		// back explicitly, because everything below assumes RAM rows and screen
		// rows are the same thing.
		OLED_Set_Offset(0);
		OLED_Clear();
		compact_move_pass = 0;
		compact_dx = 0;
		compact_page = 0;
		compact_drive = oled_drive_level();
		compact_drive_pass = 0;
		compact_draw();
		return;
	}

	// The drive first, because a panel nobody can read is worse than a panel
	// that ages. This is the lever that was set too low on 2026-08-30: the
	// block is small and it moves now, and those two carry the wear on their
	// own. See tools/oled/README.md.
	compact_drive_pass++;
	if(compact_drive_pass >= COMPACT_DRIVE_PASSES){
		uint8_t drive = oled_drive_level();

		compact_drive_pass = 0;
		if(drive != compact_drive){
			compact_drive = drive;
			OLED_Set_Contrast(drive);
		}
	}

	compact_move_pass++;
	if(compact_move_pass >= compact_period()){
		compact_move_pass = 0;
		compact_move();
		return;
	}

	// now_fps is deliberately outside kvm_state_is_changed, because it moves
	// every second while a viewer watches and would redraw the block at 1 Hz
	// for a number nobody is reading that closely. It reaches the panel on the
	// next change of anything else, or on the next move.
	if(compact_changed()){
		compact_draw();
	}
}
// --- end of the compact roaming page ---

void kvm_main_ui_disp(uint8_t first_disp, uint8_t subpage_changed)
{
	ip_addr_t now_ip_type;
	// if(kvm_oled_state.sub_page == 0)
	// if(kvm_oled_state.oled_sleep_state == 1){

	// Any operation will update the OLED sleep time
	if (kvm_state_is_changed())
		oled_auto_sleep_time_update();

	if(kvm_oled_state.sub_page == 1){
		// main page (oled sleep)
		kvm_oled_clear(first_disp || subpage_changed);
		kvm_oled_state.oled_sleep_state = 1;
	} else if(compact_page_wanted()){
		kvm_oled_state.oled_sleep_state = 0;
		kvm_compact_ui_disp(first_disp || subpage_changed);
	} else {
		// main page
		kvm_oled_state.oled_sleep_state = 0;
		now_ip_type = show_which_ip();
		kvm_main_disp(first_disp || subpage_changed);
		kvm_eth_state_disp(now_ip_type, first_disp || subpage_changed);
		kvm_wifi_state_disp(now_ip_type, first_disp || subpage_changed);
		kvm_usb_state_disp(first_disp || subpage_changed);
		kvm_hdmi_state_disp(first_disp || subpage_changed);
		kvm_fps_disp(first_disp || subpage_changed);
		kvm_res_disp(first_disp || subpage_changed);
		kvm_type_disp(first_disp || subpage_changed);
		kvm_qlty_disp(first_disp || subpage_changed);
	}
}

uint8_t show_which_page()
{
	static uint8_t run_count = 0xfe;
	static uint8_t show_type = 0;	// 0 不动 1 QRcode; 2 text
	if(sta_connect_ap()){
		if(ssid_pass_ok()){
			// 
		}
	}
	run_count++;
	if(run_count > QR_Change_time/STATE_DELAY){
		run_count = 0;
		if(show_type == 1) show_type = 2;
		else show_type = 1;
		return show_type;
	}
	return 0;
}

void show_text_wifi_config(char *ap_ssid)
{
	OLED_Clear();
	OLED_ShowString(0, 0, "SSID:", 8);
	OLED_ShowString_AlignRight(63, 1, "NanoKVM", 8);
	OLED_ShowString(0, 2, "PASS:", 8);
	OLED_ShowString_AlignRight(63, 3, ap_ssid, 8);
}

void show_wifi_config_ip(void)
{
	OLED_Clear();
	get_ip_addr(WiFi_IP);
	OLED_ShowString(1, 0, "Config URL", 8);
	OLED_ShowString_AlignRight(63, 1, "----------------", 4);
	// OLED_ShowString_AlignRight(63, 2, (char*)kvm_sys_state.wifi_addr, 4);
	static char wifi_addr_with_path[30];
	static char wifi_addr_with_key[30];
	sprintf(wifi_addr_with_path, "%s/#/", kvm_sys_state.wifi_addr);
	sprintf(wifi_addr_with_key, "WIFI?P=%s", kvm_sys_state.wifi_ap_pass);
	OLED_ShowString_AlignRight(63, 2, wifi_addr_with_path, 4);
	OLED_ShowString_AlignRight(63, 3, wifi_addr_with_key, 4);
}

void show_wifi_config_QR(void)
{
	static char cmd[70];
	OLED_Clear();
	get_ip_addr(WiFi_IP);
	sprintf(cmd, "http://%s/#/WIFI?P=%s", kvm_sys_state.wifi_addr, kvm_sys_state.wifi_ap_pass);
	qrencode(cmd);
}

void show_wifi_starting(void)
{
	OLED_Clear();
	OLED_ShowString(0, 1, "WiFi AP is", 8);
	OLED_ShowString(0, 2, "Starting..", 8);
}

void show_wifi_connecting(void)
{
	OLED_Clear();
	OLED_ShowString(0, 1, "WiFi", 8);
	OLED_ShowString(0, 2, "Connect...", 8);
}

void kvm_wifi_config_ui_disp(uint8_t first_disp, uint8_t subpage_changed)
{
	static char cmd[70];
	// printf("[kvmd]kvm_wifi_config_ui_disp %d | %d\n", first_disp, subpage_changed);
	if(first_disp) kvm_start_wifi_config_process(); // 略有不妥,就这样吧
	if(first_disp || subpage_changed){
		switch(kvm_oled_state.sub_page){
			case 0: // wifi ap is starting
				show_wifi_starting();
				break;
			case 1: // QRcode
				printf("WIFI:T:WPA2;S:NanoKVM;P:%s;;\n", kvm_sys_state.wifi_ap_pass);
				sprintf(cmd, "WIFI:T:WPA2;S:NanoKVM;P:%s;;", kvm_sys_state.wifi_ap_pass);
				qrencode(cmd);
				break;
			case 2: // Textcode
				show_text_wifi_config(kvm_sys_state.wifi_ap_pass);
				break;
			case 3: // open IP with QR Code
				show_wifi_config_QR();
				break;
			case 4: // open IP
				show_wifi_config_ip();
				break;
			case 5: // Connecting...
				show_wifi_connecting();
				break;
		}
	}
	// switch(show_which_page())
}

// oled_auto_sleep_time_update restarts the inactivity countdown.
//
// The clock is time::ticks_ms(), which MaixCDK reads from CLOCK_MONOTONIC.
// time::time_ms() is the wall clock, and this board carries no RTC battery, so
// it starts every boot near the epoch and jumps to the real date as soon as
// NTP answers. The jump is about 1.7e12 ms. A countdown that started before
// the sync is over by any measure after it, and the panel blanks with the
// operator watching it. A step backwards is worse: the subtraction below is
// unsigned, so it wraps to an enormous number rather than to zero.
void oled_auto_sleep_time_update(void)
{
	kvm_oled_state.oled_sleep_start = time::ticks_ms();
}

void oled_auto_sleep(void)
{
	uint16_t tmp16;
	uint8_t sleep_close_signal = 0;
	FILE *fp;
	int file_size;
	char RW_Data[10] = {0};
	if(access("/etc/kvm/oled_sleep", F_OK) == 0){
        fp = fopen("/etc/kvm/oled_sleep", "r");
		fseek(fp, 0, SEEK_END);
		file_size = ftell(fp); 
		fseek(fp, 0, SEEK_SET);
		if(file_size >= (int)sizeof(RW_Data)){
			file_size = sizeof(RW_Data) - 1;
		}
        fread(RW_Data, sizeof(char), file_size, fp);
		RW_Data[file_size] = '\0';
        fclose(fp);
		if(file_size != 0){
			tmp16 = atoi(RW_Data);
		} else {
			tmp16 = OLED_SLEEP_DELAY_DEFAULT;
		}
		if(tmp16 != kvm_oled_state.oled_sleep_param){
			// printf("/etc/kvm/oled_sleep = %d\n", tmp16);
			kvm_oled_state.oled_sleep_param = tmp16;
			if(kvm_oled_state.oled_sleep_param < OLED_SLEEP_DELAY_MIN){
				sleep_close_signal = 1;
			} else {
				// printf("oled_auto_sleep_time_update\n");
				oled_auto_sleep_time_update();
			}
		}
    } else {
		if(kvm_oled_state.oled_sleep_param != 0){
        	kvm_oled_state.oled_sleep_param = 0;
			sleep_close_signal = 1;
		}	
    }
	
	if(kvm_oled_state.page == 0){
		if(kvm_oled_state.oled_sleep_param < OLED_SLEEP_DELAY_MIN){
			if(sleep_close_signal == 1){
				kvm_sys_state.sub_page = 0;
			}
		} else {
			if((time::ticks_ms() - kvm_oled_state.oled_sleep_start)/1000 >= kvm_oled_state.oled_sleep_param){
				// kvm_oled_state.oled_sleep_state = 1;
				kvm_sys_state.sub_page = 1;
			} else {
				kvm_sys_state.sub_page = 0;
			}
		}
	}
}

void kvm_show_UE(void)
{
	if(kvm_hw_ver != 2){
		OLED_ShowString(0, 0, "HDMI: UE", 16);
	} else {
		OLED_Revolve();
		OLED_ShowString(0, 0, "HDMI: UE", 16);
		// OLED_ShowString_AlignRight(AlignRightEND_P, 3, "  FPS", 4);
		// kvm_init_pcie_ui();
	}
}
