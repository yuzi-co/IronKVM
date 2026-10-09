#ifndef SCREEN_DEFAULTS_H_
#define SCREEN_DEFAULTS_H_

// The screen settings an application update seeds, and only where none is kept.
//
// new_app_init used to write all four on every tarball update:
//
//	echo 30 > /kvmapp/kvm/fps
//	echo 2000 > /kvmapp/kvm/qlty
//	echo 720 > /kvmapp/kvm/res
//	echo h264 > /kvmapp/kvm/type
//
// S95nanokvm makes each of those paths a link to /etc/kvm/screen, on /data,
// before kvm_system starts, so the redirect followed the link and every update
// reset the owner's saved stream type, frame rate, bitrate and resolution
// (ironkvm-dist run sheet, trial 68). A file that holds a usable value is now
// left alone; one that is missing, empty or unusable gets the default.
//
// The bitrate default is 3000 kbit/s, the web's "High" step and the server's
// DefaultBitRate (server/common/screen.go), not Sipeed's 2000.
//
// This file has no MaixCDK dependency on purpose, so a host test can build it
// alone (tools/kvm_system/test-screen-defaults.sh).

#define SCREEN_DEFAULT_FPS	"30"
#define SCREEN_DEFAULT_QLTY	"3000"
#define SCREEN_DEFAULT_RES	"720"
#define SCREEN_DEFAULT_TYPE	"h264"

// Whether the text a settings file holds is a value its reader can use.
// name is the file's name: fps, qlty, res or type.
int screen_setting_valid(const char *name, const char *text);

// Writes the default into each of fps, qlty, res and type under dir whose
// file is missing, empty or not valid, and leaves the others alone. A link is
// followed, as the shell redirect it replaces followed it, so a link to a
// shared file that does not exist yet creates that file. Each file written is
// reported on stdout. Answers the number written, or -1 if a write failed.
int seed_screen_defaults(const char *dir);

#endif
