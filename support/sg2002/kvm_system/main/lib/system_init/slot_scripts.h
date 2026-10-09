#ifndef SLOT_SCRIPTS_H_
#define SLOT_SCRIPTS_H_

// Which boot scripts an application update may change on this slot.
//
// new_app_init copies boot scripts from /kvmapp/system/init.d into /etc/init.d
// after every tarball update, and deletes S30wifi on a board without wlan0.
// That is right for Sipeed's own firmware, but an IronKVM slot image chooses its
// own boot scripts: the vendor image removes S00kmod, because its hook runner
// loads the same modules, and keeps IronKVM's S30wifi, which says the board has
// no radio and exits 0. The mainline image refuses more. ironkvm-dist run
// sheet, trial 70: an update put S00kmod back and deleted S30wifi on slot A.
//
// Every IronKVM image installs the list of scripts it refuses at
// SLOT_REFUSE_LIST, one name per line, '#' starting a comment. install.sh reads
// the same file. A slot without it is a stock Sipeed board, and there nothing
// changes: every copy and the S30wifi delete run as before.
//
// This file has no MaixCDK dependency on purpose, so a host test can build it
// alone (tools/kvm_system/test-slot-scripts.sh).

#define SLOT_REFUSE_LIST	"/usr/share/ironkvm/init.d.refuse"

// Whether the list names the script. A missing list refuses nothing.
int slot_refuses(const char *list, const char *name);

// Copies src_dir/name into dest_dir with cp -f, unless the list refuses it,
// which is reported on stdout. Answers 1 when it ran the copy, 0 when not.
int install_boot_script(const char *list, const char *src_dir,
			const char *dest_dir, const char *name);

enum s30wifi_action {
	S30WIFI_INSTALL,	// cp -f the package's S30wifi
	S30WIFI_KEEP,		// leave /etc/init.d/S30wifi as the image made it
	S30WIFI_REMOVE,		// rm -f /etc/init.d/S30wifi
};

// What new_app_init does with S30wifi. vendor is whether the kernel is the
// vendor 5.10 one, wifi whether wlan0 exists. Without a list this is the stock
// rule, install on a vendor kernel with wlan0 and remove otherwise. With one, a
// refused S30wifi is removed and any other is installed where stock would
// install it and kept where stock would remove it.
enum s30wifi_action s30wifi_action(const char *list, int vendor, int wifi);

#endif
