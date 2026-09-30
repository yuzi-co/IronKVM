#ifndef NET_PROBE_H_
#define NET_PROBE_H_

// Gateway reachability without a shell.
//
// kvm_system used to run "ping -I <if> -w 1 <gateway>" through system() on
// every pass of its 1000 ms state loop: a fork of sh, a fork of busybox, and a
// second of blocking, once a second, on a single-core board. The OLED only
// shows the verdict as an icon, so it does not need an answer every second.
//
// This file has no MaixCDK dependency on purpose, so a host test can build it
// alone (tools/oled/test-net-probe.sh).

#include <stddef.h>
#include <stdint.h>
#include <netinet/in.h>

// How often a gateway is probed while nothing about the link has changed. A
// change of link or address forces the next probe, so this only bounds how
// long a gateway that stops or starts answering takes to show.
#define GATEWAY_PROBE_INTERVAL_MS	10000U

// The time a probe waits for its reply. "ping -w 1" waited the same.
#define GATEWAY_PROBE_TIMEOUT_MS	1000

typedef struct {
	uint8_t  valid;		// 0: the next check is due, whatever the time
	uint32_t last_ms;	// when the last probe ran, on the monotonic clock
} probe_timer_t;

// Milliseconds on the monotonic clock. Wraps after 49 days; the arithmetic in
// probe_due is unsigned, so the wrap is harmless.
uint32_t probe_now_ms(void);

// 1 when a probe is due: never probed, invalidated, or the interval has passed.
int probe_due(const probe_timer_t *t, uint32_t now_ms, uint32_t interval_ms);
void probe_mark(probe_timer_t *t, uint32_t now_ms);
void probe_invalidate(probe_timer_t *t);

// Read an IPv4 address from the start of s, at most len bytes, stopping at the
// first byte that cannot be part of one. s need not be terminated: the route
// buffers are 16 bytes and a 15-character address with its newline fills one.
// 1 on success.
int parse_gateway(const char *s, size_t len, struct in_addr *out);

// Find the default gateway of ifname in a /proc/net/route dump, the way
// "ip route | grep '^default' | grep <ifname> | awk '{print $3}'" did: the
// first default route through that interface that names a gateway. Writes it
// dotted, followed by a space where that pipeline's newline used to be, into
// out, truncated to out_len - 1 bytes and terminated. That is what fgets into
// the 16-byte route buffer produced, so a 15-character address loses its
// space, as it did before. 1 when one was found.
int route_gateway_from_table(const char *table, const char *ifname, char *out, size_t out_len);

// The same, reading /proc/net/route. 0 when it cannot be read.
int route_gateway(const char *ifname, char *out, size_t out_len);

// Send one ICMP echo to dst out of ifname and wait up to timeout_ms for the
// reply. 1: it answered. 0: it did not. -1: this process cannot probe (no raw
// socket, or the interface cannot be bound), and the caller should fall back.
int icmp_probe(const char *ifname, struct in_addr dst, int timeout_ms);

#endif // NET_PROBE_H_
