#include "config.h"
#include "system_state.h"
#include "vi_state_shared.hpp"
#include "net_probe.h"
#include <sys/socket.h>
#include <net/if.h>
#include <sys/ioctl.h>

using namespace maix;
using namespace maix::sys;

extern kvm_sys_state_t kvm_sys_state;
extern kvm_oled_state_t kvm_oled_state;

// read_small_file reads at most size - 1 bytes of path into buf and ends them
// with a NUL. A file that cannot be opened reads as empty and returns -1.
//
// The readers below used to fopen, seek to the end, and read ftell's answer
// into a 10-byte buffer with no check of either. A missing file crashed the
// process on the NULL stream (the runtime files in /tmp/kvm exist only once
// S95nanokvm has made them), and a sysfs attribute, whose ftell is the page
// size, could overrun the buffer.
static int read_small_file(const char *path, char *buf, size_t size)
{
	FILE *fp = fopen(path, "r");
	if(fp == NULL){
		buf[0] = 0;
		return -1;
	}
	size_t n = fread(buf, 1, size - 1, fp);
	fclose(fp);
	buf[n] = 0;
	return (int)n;
}

int get_nic_state(const char* interface_name)
{
	int sock;
	struct ifreq ifr;
	int ret = NIC_STATE_NO_EXIST;
	if ((sock = socket(AF_INET, SOCK_STREAM, 0)) < 0) {
		return ret;
	}
	strcpy(ifr.ifr_name, interface_name);
	if (ioctl(sock, SIOCGIFFLAGS, &ifr) < 0) {
		close(sock);
		return ret;
	}
	if (ifr.ifr_flags & IFF_UP) {
		if (ifr.ifr_flags & IFF_RUNNING) {
			ret = NIC_STATE_RUNNING;
		} else {
			ret = NIC_STATE_UP;
		}
	} else {
		ret = NIC_STATE_DOWN;
	}
	close(sock);
	return ret;
}

int get_ping_allow_state(void)
{
	if(access("/etc/kvm/stop_ping", F_OK) == 0) {
		kvm_sys_state.ping_allow = 0;
	} else {
		kvm_sys_state.ping_allow = 1;
	}
	return kvm_sys_state.ping_allow;
}

// net_port
int get_ip_addr(ip_addr_t ip_type)
{
	switch (ip_type){
		case ETH_IP: // eth_addr
			if(strcmp(ip_address()["eth0"].c_str(), (char*)kvm_sys_state.eth_addr) != 0){
				if(*(ip_address()["eth0"].c_str()) == 0){
					printf("can`t get ip addr\r\n");
					kvm_sys_state.eth_addr[0] = 0;
					return 0;
				} 
				for(int i = 0; i <= 15; i++)
				{
					kvm_sys_state.eth_addr[i] = *(ip_address()["eth0"].c_str() + i);
					printf("%c", kvm_sys_state.eth_addr[i]);
				}
				printf("\r\n");
			}
			return 1;
		case WiFi_IP: // wifi_addr
			if(strcmp(ip_address()["wlan0"].c_str(), (char*)kvm_sys_state.wifi_addr) != 0){
				if(*(ip_address()["wlan0"].c_str()) == 0){
					printf("can`t get ip addr\r\n");
					kvm_sys_state.wifi_addr[0] = 0;
					return 0;
				} 
				for(int i = 0; i <= 15; i++)
				{
					kvm_sys_state.wifi_addr[i] = *(ip_address()["wlan0"].c_str() + i);
					printf("%c", kvm_sys_state.wifi_addr[i]);
				}
				printf("\r\n");
			}
			return 1;
		case Tailscale_IP: // tail_addr
			if(*(ip_address()["tailscale0"].c_str()) == 0){
				printf("can`t get ip addr\r\n");
				kvm_sys_state.tail_addr[0] = 0;
				return 0;
			} 
			for(int i = 0; i <= 15; i++)
			{
				kvm_sys_state.tail_addr[i] = *(ip_address()["tailscale0"].c_str() + i);
				printf("%c", kvm_sys_state.tail_addr[i]);
			}
			printf("\r\n");
			return 1;
		case RNDIS_IP: // rndis_addr
			if(*(ip_address()["usb0"].c_str()) == 0){
				printf("can`t get ip addr\r\n");
				kvm_sys_state.rndis_addr[0] = 0;
				return 0;
			} 
			for(int i = 0; i <= 15; i++)
			{
				kvm_sys_state.rndis_addr[i] = *(ip_address()["usb0"].c_str() + i);
				printf("%c", kvm_sys_state.rndis_addr[i]);
			}
			printf("\r\n");
			return 1;
		case ETH_ROUTE: // eth_route
			if(access("/etc/kvm/gateway", F_OK) != 0){
				// 不存在gateway文件
				// Read from /proc/net/route rather than a shell pipeline of
				// ip, grep and awk: this runs every pass until a default route
				// appears, which on a port with none is every pass for ever.
				// 开机时未插入ETH: nothing is found and eth_route stays empty.
				return route_gateway("eth0", (char*)kvm_sys_state.eth_route, sizeof( kvm_sys_state.eth_route ));
			} else {
				return read_small_file("/etc/kvm/gateway", (char*)kvm_sys_state.eth_route,
					sizeof(kvm_sys_state.eth_route)) > 0;
			}
		case WiFi_ROUTE: // wifi_route
			// Every pass while wlan0 has an address and is not yet up, so no
			// shell here either.
			return route_gateway("wlan0", (char*)kvm_sys_state.wifi_route, sizeof( kvm_sys_state.wifi_route ));
	}
	return 0;
}

int chack_net_state(ip_addr_t use_ip_type)
{
	const char* ifname;
	const uint8_t* route;
	if		(use_ip_type == ETH_ROUTE)  { ifname = "eth0";  route = kvm_sys_state.eth_route; }
	else if	(use_ip_type == WiFi_ROUTE) { ifname = "wlan0"; route = kvm_sys_state.wifi_route; }
	else return -1;	// 不支持的端口

	// Probe in-process when the gateway is an address. A name in
	// /etc/kvm/gateway, or a process that may not open a raw socket, falls
	// through to ping, which is what every probe used to cost.
	struct in_addr gateway;
	if (parse_gateway((const char*)route, sizeof(kvm_sys_state.eth_route), &gateway)) {
		int ret = icmp_probe(ifname, gateway, GATEWAY_PROBE_TIMEOUT_MS);
		if (ret >= 0) return ret;
	}

	char Cmd[100]={0};
	sprintf( Cmd,"ping -I %s -w 1 %s > /dev/null", ifname, (const char*)route);
	if(system(Cmd) == 0){	// 256：不通； = 0：通
		return 1;
	}
	return 0;
}

void patch_eth_wifi(void)
{
	// system("ip link set eth0 down");
	// system("ip link set eth0 up");
	// system("udhcpc -i eth0 &");
}

int kvm_wifi_exist()
{
	if (get_nic_state("wlan0") == NIC_STATE_NO_EXIST) return 0;
	else return 1;
}

void kvm_update_usb_state()
{
	// usb_state, hid_state, rndis_state, udisk_state
	char RW_Data[16];
	read_small_file("/sys/class/udc/4340000.usb/state", RW_Data, sizeof(RW_Data));
	if(RW_Data[0] == 'n') kvm_sys_state.usb_state = 0;
	else if(RW_Data[0] == 'c') kvm_sys_state.usb_state = 1;
	else kvm_sys_state.usb_state = -1;
	// hid_state & udisk_state (rndis_state单独处理)
	if(kvm_sys_state.usb_state == 1){
		if(access("/sys/kernel/config/usb_gadget/g0/configs/c.1/hid.GS*", F_OK) == 0) 
			kvm_sys_state.hid_state = 1;
		if(access("/sys/kernel/config/usb_gadget/g0/configs/c.1/mass_storage.disk0", F_OK) == 0) 
			kvm_sys_state.udisk_state = 1;
	} else {
		kvm_sys_state.hid_state = 0;
		kvm_sys_state.udisk_state = 0;
	}
}

void kvm_update_hdmi_state()
{
	const uint32_t shared_state_ttl_ms = 30000U;
	const uint32_t fallback_state_ttl_ms = 10000U;
	static uint8_t check_times = 4;
	static vi_state_shared::State fallback_state = {};
	static uint32_t fallback_updated_ms = 0;
	static uint32_t fallback_attempted_ms = 0;
	static bool fallback_valid = false;
	static bool fallback_attempted = false;
	static bool shared_failure_logged = false;
	static bool fallback_failure_logged = false;
	if(++check_times > 5){
		check_times = 0;
		vi_state_shared::State state = {};
		vi_state_shared::ReadStatus status = vi_state_shared::read_state(&state, shared_state_ttl_ms);
		if (status == vi_state_shared::READ_OK) {
			kvm_sys_state.hdmi_state = state.fps == 0 ? 0 : 1;
			if (shared_failure_logged) {
				fprintf(stderr, "[kvm_system] VI shared state recovered\n");
			}
			shared_failure_logged = false;
			fallback_failure_logged = false;
			return;
		}

		if (!shared_failure_logged) {
			fprintf(stderr, "[kvm_system] VI shared state %s; using direct fallback\n",
				vi_state_shared::read_status_name(status));
			shared_failure_logged = true;
		}

		uint32_t now = vi_state_shared::monotonic_ms();
		if (!fallback_attempted || now - fallback_attempted_ms >= fallback_state_ttl_ms) {
			fallback_attempted_ms = now;
			fallback_attempted = true;
			uint32_t fields = vi_state_shared::FIELD_NONE;
			vi_state_shared::State direct_state = {};
			vi_state_shared::ProcReadStatus direct_status =
				vi_state_shared::read_proc_state(&direct_state, &fields);
			if (direct_status == vi_state_shared::PROC_READ_OK &&
				(fields & vi_state_shared::FIELD_FPS) != 0U) {
				fallback_state = direct_state;
				fallback_updated_ms = now;
				fallback_valid = true;
				if (fallback_failure_logged) {
					fprintf(stderr, "[kvm_system] direct VI fallback recovered\n");
				}
				fallback_failure_logged = false;
			} else {
				fallback_valid = false;
				if (!fallback_failure_logged) {
					fprintf(stderr, "[kvm_system] direct VI fallback unavailable\n");
					fallback_failure_logged = true;
				}
			}
		}

		if (fallback_valid && now - fallback_updated_ms <= fallback_state_ttl_ms) {
			kvm_sys_state.hdmi_state = fallback_state.fps == 0 ? 0 : 1;
		} else {
			kvm_sys_state.hdmi_state = -1;
		}
	}
}

void kvm_update_stream_fps(void)
{
	uint8_t RW_Data[16];

	// FPS
	read_small_file("/kvmapp/kvm/now_fps", (char*)RW_Data, sizeof(RW_Data));
	kvm_sys_state.now_fps = atoi((char*)RW_Data);
}

void kvm_update_stream_type(void)
{
	uint8_t RW_Data[16];

	// type
	read_small_file("/kvmapp/kvm/type", (char*)RW_Data, sizeof(RW_Data));
	if(RW_Data[0] == 'm') 		kvm_sys_state.type = KVM_TYPE_MJPG;
	else if(RW_Data[0] == 'h') 	kvm_sys_state.type = KVM_TYPE_H264;
	else 						kvm_sys_state.type = KVM_TYPE_none;
}

void kvm_update_stream_qlty(void)
{
	uint8_t RW_Data[16];
	uint16_t tmp16;

	// QLTY
	read_small_file("/kvmapp/kvm/qlty", (char*)RW_Data, sizeof(RW_Data));
	tmp16 = atoi((char*)RW_Data);
	if(kvm_sys_state.type == KVM_TYPE_MJPG){
		if(tmp16 < 60) 						 	kvm_sys_state.qlty = 1;
		else if(tmp16 >= 60 && tmp16 < 75) 	 	kvm_sys_state.qlty = 2;
		else if(tmp16 >= 75 && tmp16 < 90) 	 	kvm_sys_state.qlty = 3;
		else if(tmp16 >= 90 && tmp16 <= 100) 	kvm_sys_state.qlty = 4;
		else 									kvm_sys_state.qlty = 4;
	} else {
		if(tmp16 < 1500) 						kvm_sys_state.qlty = 1;
		else if(tmp16 >= 1500 && tmp16 < 2500) 	kvm_sys_state.qlty = 2;
		else if(tmp16 >= 2500 && tmp16 < 3500) 	kvm_sys_state.qlty = 3;
		else if(tmp16 >= 3500 && tmp16 <= 5000) kvm_sys_state.qlty = 4;
		else 									kvm_sys_state.qlty = 4;
	}
}

void kvm_update_hdmi_res(void)
{
	uint8_t RW_Data[16];
	// HDMI width
	read_small_file("/kvmapp/kvm/width", (char*)RW_Data, sizeof(RW_Data));
	kvm_sys_state.hdmi_width = atoi((char*)RW_Data);
	// HDMI height
	read_small_file("/kvmapp/kvm/height", (char*)RW_Data, sizeof(RW_Data));
	kvm_sys_state.hdmi_height = atoi((char*)RW_Data);
}

void kvm_update_eth_state(void)
{	
	static uint8_t nic_state = 0;
	// The gateway is probed every GATEWAY_PROBE_INTERVAL_MS, and eth_state
	// keeps the last verdict in between. Anything else that writes eth_state
	// invalidates the timer, so the verdict is back on the next pass, as it
	// was when every pass probed. NIC_STATE_RUNNING is IFF_RUNNING, which is
	// the carrier, so a port with no cable is never probed.
	static probe_timer_t probe = {};
	nic_state = get_nic_state("eth0");

	if(nic_state == NIC_STATE_RUNNING){
		// Get IP
		if(strcmp(ip_address()["eth0"].c_str(), (char*)kvm_sys_state.eth_addr) != 0){
			probe_invalidate(&probe);
			if(get_ip_addr(ETH_IP)){
				kvm_sys_state.eth_state = 2;
			} else {
				kvm_sys_state.eth_state = 1;
				return;
			}
		}
		if(kvm_sys_state.ping_allow){
			// ping route
			if(kvm_sys_state.eth_route[0] == 0){
				get_ip_addr(ETH_ROUTE);
			} else {
				uint32_t now = probe_now_ms();
				if(probe_due(&probe, now, GATEWAY_PROBE_INTERVAL_MS)){
					probe_mark(&probe, now);
					if(chack_net_state(ETH_ROUTE)){
						// Ping successful
						kvm_sys_state.eth_state = 3;
					} else {
						kvm_sys_state.eth_state = 2;
					}
				}
			}
		} else {
			// Consider the network to be connected
			kvm_sys_state.eth_state = 3;
		}

	} else {
		probe_invalidate(&probe);
		kvm_sys_state.eth_state = 0;
		patch_eth_wifi();
	}
}

// The server reads this file to tell whether WiFi is up. It is written every
// pass, as it was by "echo N > ..." through a shell, but without the fork. It
// is a link into tmpfs (S95nanokvm), and fopen follows the link as the shell's
// redirect did.
static void write_wifi_state_file(int state)
{
	FILE *fp = fopen("/kvmapp/kvm/wifi_state", "w");
	if (fp == NULL) return;
	fprintf(fp, "%d\n", state);
	fclose(fp);
}

void kvm_update_wifi_state(void)
{	
	// No WiFi module (check for existence?) -> Module exists & not connected (check if connected) ->
	// The gateway is probed at most every GATEWAY_PROBE_INTERVAL_MS, as on
	// eth0: in state 0 to decide the link is up, in state 1 that it is still
	// up. wifi_state holds between probes. The first probe is not delayed.
	static probe_timer_t probe = {};
	if(kvm_sys_state.wifi_state == -2) return;
	switch (kvm_sys_state.wifi_state){
		case -1:
		// Initial default value.
			if (kvm_wifi_exist()) {
				kvm_sys_state.wifi_state = 0;
				system("touch /etc/kvm/wifi_exist");
			}
			else {
				kvm_sys_state.wifi_state = -2; // WiFi module does not exist, exiting directly.
				system("rm /etc/kvm/wifi_exist");
				return;
			}
			// break;	// Start checking the connection directly.
		case 0:
		// WiFi is available but not connected.
			write_wifi_state_file(0);
			if (get_ip_addr(WiFi_IP) && get_ip_addr(WiFi_ROUTE)){
				// IP+Route has been acquired
				if(kvm_sys_state.ping_allow){
					uint32_t now = probe_now_ms();
					if (probe_due(&probe, now, GATEWAY_PROBE_INTERVAL_MS)){
						probe_mark(&probe, now);
						if (chack_net_state(WiFi_ROUTE)){
							// Ping successful
							kvm_sys_state.wifi_state = 1;
						}
					}
				} else {
					// Consider the network to be connected
					kvm_sys_state.wifi_state = 1;
				}
			}
			break;
		case 1:
		// Connected to the network & continuously checking if it can ping successfully.
			write_wifi_state_file(1);
			get_ip_addr(WiFi_IP);
			if(kvm_sys_state.ping_allow){
				uint32_t now = probe_now_ms();
				if (kvm_sys_state.wifi_route[0] != 0 && probe_due(&probe, now, GATEWAY_PROBE_INTERVAL_MS)){
					probe_mark(&probe, now);
					if (chack_net_state(WiFi_ROUTE) == 0){
						// Ping successful
						kvm_sys_state.wifi_state = 0;
					}
				}
			}
		// default:
		// 	kvm_sys_state.wifi_state = -1;
	}
}

void kvm_update_rndis_state(void)
{
	if (get_nic_state("usb0") == NIC_STATE_RUNNING) {
		if(kvm_sys_state.rndis_state != 1) {
			if (get_ip_addr(RNDIS_IP)) {
				kvm_sys_state.rndis_state = 1;
			}
		}
	}
	else kvm_sys_state.rndis_state = 0;
}

void kvm_update_tailscale_state(void)
{
	if (get_nic_state("tailscale0") == NIC_STATE_RUNNING) {
		if(kvm_sys_state.tail_state != 1){
			if (get_ip_addr(Tailscale_IP)) {
				kvm_sys_state.tail_state = 1;
			}
		}
	}
	else kvm_sys_state.tail_state = 0;
}

//============================================================================

uint8_t ion_free_space(void)
{
	//cat /sys/kernel/debug/ion/cvi_carveout_heap_dump/summary | grep "usage rate:" | awk '{print $2}'

	return 0;
}

int create_temp_watchdog(void) 
{
    FILE *file;

    file = fopen(watchdog_temp_path, "w");
    if (file == NULL) {
        printf("[kvmv] Temp watchdog create error\n");
        return -1;
    }
    // fprintf(file, "%s", 'v');
    fclose(file);
    return 1;
}

void rm_temp_watchdog(void)
{
	if(access(watchdog_temp_path, F_OK) == 0) {
		remove(watchdog_temp_path);
	}
}

void auto_remove_temp_watchdog(void)
{
	static uint8_t run_times = 0;
	static uint8_t temp_watchdog_removed = 0;
	if(temp_watchdog_removed) return;
	if(run_times++ >= RM_Watchdog_times){
		run_times = 0;
		temp_watchdog_removed = 1;
		rm_temp_watchdog();
	}
}

uint8_t watchdog_sf_is_open(void)
{
	if(access(watchdog_mode_path, F_OK) == 0) return 1;
	if(access(watchdog_temp_path, F_OK) == 0) return 1;
	else return 0;
}

int check_watchdog() 
{
	if(access(watchdog_file, F_OK) == 0) {
		if (remove(watchdog_file) == 0) {
			return 1;
		} else {
			return -1;
		}
	} else return 0;
}
