#include "config.h"
#include "system_init.h"
#include "screen_defaults.h"

#include <errno.h>
#include <sys/stat.h>
#include <sys/utsname.h>

using namespace maix;
using namespace maix::sys;

extern kvm_sys_state_t kvm_sys_state;
extern kvm_oled_state_t kvm_oled_state;

// The state file lives in tmpfs. /kvmapp is the boot medium, and a file that
// is rewritten whenever HDMI presence changes wears it for no gain: the value
// describes this moment and means nothing after a reboot. S95nanokvm leaves a
// symlink at the old path, which a reader follows and this writer must not
// replace, so the temporary file is made beside the target rather than beside
// the link.
static void write_hdmi_state(uint8_t active)
{
    char temp_path[] = "/tmp/kvm/.state.init.XXXXXX";
    const char *state = active != 0 ? "1\n" : "0\n";
    int fd = mkstemp(temp_path);
    if(fd < 0){
        printf("failed to create HDMI state file: %d\n", errno);
        return;
    }

    int failure = 0;
    if(fchmod(fd, 0644) != 0){
        failure = errno;
    }
    if(failure == 0 && write(fd, state, 2) != 2){
        failure = errno != 0 ? errno : EIO;
    }
    if(failure == 0 && fsync(fd) != 0){
        failure = errno;
    }
    if(close(fd) != 0 && failure == 0){
        failure = errno;
    }
    if(failure == 0 && rename(temp_path, "/tmp/kvm/state") != 0){
        failure = errno;
    }
    if(failure != 0){
        unlink(temp_path);
        printf("failed to publish HDMI state file: %d\n", failure);
    }
}

uint8_t get_hdmi_version()
{
	FILE *fp;
	uint8_t RW_Data[2];
    system("/kvmapp/system/init.d/S15kvmhwd get_hdmi_version");
	if(access("/etc/kvm/hdmi_version", F_OK) == 0){
        fp = fopen("/etc/kvm/hdmi_version", "r");
        fread(RW_Data, sizeof(char), 2, fp);
        fclose(fp);
        if(RW_Data[0] == 'u'){
            // 6911uxc / 6911uxe
            if(RW_Data[1] == 'e'){
                return 2;
            } else if(RW_Data[1] == 'x') {
                return 1;
            } else {
                return 1;
            }
        } else if(RW_Data[0] == 'd'){
            // 6911d
            return 3;
        } else {
            // 6911c
            return 0;
        }
    } else {
		return 0;
    }
}

void Production_testing_patch(void)
{	
	// Product UE version detecte
	if(get_hdmi_version() == 2){
		printf("ue_patch_state = 1;\n");
		kvm_oled_state.ue_patch_state = 1;
	} else {
		printf("ue_patch_state = 0;\n");
		kvm_oled_state.ue_patch_state = 0;
	}

	// New products default to disabling mDNS functionality
	system("rm -f /etc/init.d/S50ssdpd");
	system("sync");
}

// Whether the running kernel is Sipeed's vendor 5.10 kernel. Part of
// new_app_init is written for that kernel alone: S00kmod loads modules from
// /mnt/system/ko, S15kvmhwd and S30wifi drive vendor GPIO numbers and drivers,
// and soph_saradc.ko and soph_mipi_rx.ko are 5.10 modules. The IronKVM
// mainline slot runs Linux 7.x, has none of those modules, and refuses the
// same scripts on an application update (ironkvm-dist
// devices/sipeed-nanokvm/flavour.d/mainline/init.d.refuse). There, the module
// check below finds no vendor soph_mipi_rx.ko, copies one in and reboots, on
// every update.
//
// The kernel's own release is the test, not a file an image installs, so it
// holds on a stock Sipeed image, on either IronKVM slot, and on a slot whose
// IronKVM files are missing. If uname fails the answer is yes, which is the
// behaviour this function had before it asked.
static int vendor_kernel(char *release, size_t len)
{
	struct utsname uts;
	if(uname(&uts) != 0){
		snprintf(release, len, "unknown");
		return 1;
	}
	snprintf(release, len, "%s", uts.release);
	return strncmp(uts.release, "5.10", 4) == 0;
}

void new_app_init(void)
{
	char release[65];
	int vendor = vendor_kernel(release, sizeof(release));
	if(!vendor){
		printf("new_app_init: kernel %s is not the vendor 5.10 kernel; skipping S00kmod, S15kvmhwd, S30wifi, soph_saradc and soph_mipi_rx.ko\n", release);
	}

	// Update the necessary scripts
	system("rm -f /boot/logo.jpeg");
	system("cp -f /kvmapp/system/update-nanokvm.py /etc/kvm/");
	system("rm -f /etc/init.d/S02udisk");
	if(vendor){
		system("cp -f /kvmapp/system/init.d/S00kmod /etc/init.d/");
	}
	system("cp -f /kvmapp/system/init.d/S01fs /etc/init.d/");
	system("cp -f /kvmapp/system/init.d/S03usbdev /etc/init.d/");
	if(vendor){
		system("cp -f /kvmapp/system/init.d/S15kvmhwd /etc/init.d/");
	}
	system("cp -f /kvmapp/system/init.d/S30eth /etc/init.d/");
	system("cp -f /kvmapp/system/init.d/S50sshd /etc/init.d/");
	if(vendor && kvm_wifi_exist()) {
		system("cp -f /kvmapp/system/init.d/S30wifi /etc/init.d/");
	} else {
		system("rm -f /etc/init.d/S30wifi");
	}

	// if exit /etc/init.d/S98tailscaled then cp -f /kvmapp/system/init.d/S98tailscaled /etc/init.d/
	if(access("/etc/init.d/S98tailscaled", F_OK) == 0){
		system("cp -f /kvmapp/system/init.d/S98tailscaled /etc/init.d/");
	}

	// rmmod soph_saradc
	if(vendor){
		system("rmmod soph_saradc");
		system("rm -f /mnt/system/ko/soph_saradc.ko");
	}

	// PCIe Patch
	// system("cp /kvmapp/system/init.d/S95nanokvm /etc/init.d/");
	if(access("/kvmapp/jpg_stream/dl_lib/libmaixcam_lib.so", F_OK) != 0){
		system("cp -f /kvmapp/system/init.d/S95nanokvm /etc/init.d/");
	}

	// Remove unnecessary components to speed up boot time
	system("rm -f /etc/init.d/S04backlight");
	system("rm -f /etc/init.d/S05tp");
	system("rm -f /etc/init.d/S40bluetoothd");
	system("rm -f /etc/init.d/S50ssdpd");
	system("rm -f /etc/init.d/S99*");
	
	// Add necessary configuration files for program execution
	system("mkdir /kvmapp/kvm");
	system("mkdir /etc/kvm");
	system("echo 0 > /kvmapp/kvm/now_fps");
	// The screen settings are the owner's, kept on /data behind the
	// /kvmapp/kvm links: seed only the ones that are missing or unusable.
	// See screen_defaults.h.
	seed_screen_defaults("/kvmapp/kvm");
	write_hdmi_state(0);
	system("touch /etc/kvm/frame_detact");

	// rm jpg_stream & kvm_stream
	system("rm -rf /kvmapp/jpg_stream");
	system("rm -f /kvmapp/kvm_system/kvm_stream");	// Cannot delete temporarily; the old production test script uses this file to determine if the download is complete

	system("rm -f /kvmapp/kvm_new_app");
	system("sync");
	// system("/etc/init.d/S95nanokvm restart");

	// update ko
	if(vendor){
		FILE *fp;
		uint8_t RW_Data_0[30];
		uint8_t RW_Data_1[30];
		fp = popen("md5sum /mnt/system/ko/soph_mipi_rx.ko | grep 086ed01749188975afaa40fb569374f8 | awk '{print $2}'", "r");
		if ( NULL == fp )
		{
			pclose(fp);
			// return;
		}
		fgets((char*)RW_Data_0, 10, fp);
		pclose(fp);
		fp = popen("md5sum /mnt/system/ko/soph_mipi_rx.ko | grep 69be7eeded3777f750480a5dd5a1aa26 | awk '{print $2}'", "r");
		if ( NULL == fp )
		{
			pclose(fp);
			// return;
		}
		fgets((char*)RW_Data_1, 10, fp);
		pclose(fp);

		int8_t hdmi_ver = -1;

		if(access("/etc/kvm/hdmi_version", F_OK) == 0){
			uint8_t RW_Data[2];
			FILE *fp = fopen("/etc/kvm/hdmi_version", "r");
			fread(RW_Data, sizeof(char), 2, fp);
			fclose(fp);
			if(RW_Data[0] == 'c') hdmi_ver = 1;
			else if(RW_Data[0] == 'u') hdmi_ver = 2;
			else if(RW_Data[0] == 'd') hdmi_ver = 2;
		}

		// system("/etc/init.d/S03usbdev stop_start");
		create_temp_watchdog();

		if(hdmi_ver == 2){
			if(RW_Data_1[0] != '/'){
				system("cp /kvmapp/system/ko/soph_mipi_rx.ko /mnt/system/ko/soph_mipi_rx.ko");
				system("sync");
				system("reboot");
			} else {
				system("sync");
				system("/etc/init.d/S15kvmhwd start");
				system("/etc/init.d/S95nanokvm restart");
			}
		} else {
			if((RW_Data_0[0] != '/') && (RW_Data_1[0] != '/')){
				system("cp /kvmapp/system/ko/soph_mipi_rx.ko /mnt/system/ko/soph_mipi_rx.ko");
				system("sync");
				system("reboot");
			} else {
				system("sync");
				system("/etc/init.d/S15kvmhwd start");
				system("/etc/init.d/S95nanokvm restart");
			}
		}
	} else {
		// The rest of the vendor branch without the module or S15kvmhwd:
		// the same watchdog file, and the restart that lets S95nanokvm
		// finish the migration.
		create_temp_watchdog();
		system("sync");
		system("/etc/init.d/S95nanokvm restart");
	}

	if(access("/root/old/kvm_new_img", F_OK) == 0){
		system("rm -r /root/old");
	}	
	
	system("sync");
	system("killall NanoKVM-Server");
	system("rm -r /tmp/server");
	system("cp -r /kvmapp/server /tmp/");
	system("/tmp/server/NanoKVM-Server &");
}

void build_complete_resolv(void)
{
	FILE *fp = NULL;
	// fp = fopen("/boot/resolv.conf", "w+");
	fp = fopen("/boot/resolv.conf", "w");
	// 阿里: 223.5.5.5
	// 腾讯: 119.29.29.29
	fprintf(fp, "nameserver 192.168.0.1\nnameserver 8.8.4.4\nnameserver 8.8.8.8\nnameserver 114.114.114.114\nnameserver 119.29.29.29\nnameserver 223.5.5.5");
	fclose(fp);
	system("rm -rf /etc/resolv.conf");
	system("cp -vf /etc/resolv.conf /etc/resolv.conf.old");
	system("cp -vf /boot/resolv.conf /etc/resolv.conf");
}

void new_img_init(void)
{
	build_complete_resolv();
	system("rm /kvmapp/kvm_new_img");
}
