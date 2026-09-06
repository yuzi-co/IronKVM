package utils

import (
	"crypto/tls"
	"net/http"
	"time"
)

// readHeaderTimeout bounds how long a client may take to send its request
// headers. Without it a handful of sockets trickling one byte at a time ties
// up the whole server (slowloris).
const readHeaderTimeout = 15 * time.Second

// idleTimeout closes keep-alive connections that go quiet between requests.
const idleTimeout = 2 * time.Minute

// NewServer builds an http.Server with the timeouts that gin's Run helpers
// leave at zero.
//
// ReadTimeout and WriteTimeout stay unset on purpose: this server streams MJPEG
// responses that never end and accepts multi-gigabyte image uploads, both of
// which a whole-request deadline would cut off. Websocket connections hijack
// the socket, so these timeouts do not apply to them either way.
func NewServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		IdleTimeout:       idleTimeout,

		// An empty non-nil map turns HTTP/2 off, and AllowHTTP2 below is what
		// puts it back. ListenAndServeTLS configures h2 automatically when this
		// field is nil, so the HTTPS listener spoke it while the plain one
		// spoke HTTP/1.1, and the video stream paid for the difference.
		//
		// The board serves one video stream to one viewer over a single
		// connection. HTTP/2 exists to multiplex many requests over one
		// connection, and this workload has nothing to multiplex, so its frame
		// scheduler, per-stream write queues and flow-control accounting are
		// cost with no matching benefit. The C906 is already saturated by the
		// record layer: measured 2026-09-06 at 1080p, MJPEG carried 10.5MB/s
		// over HTTP/1.1 at 69% of the core and 3.5MB/s over h2 with TLS at 97%,
		// which cut the delivered frame rate from 30 to about 10.
		//
		// The field is only consulted for TLS connections, so setting it here
		// costs the plain listeners nothing.
		//
		// It also removes a hazard. Go's HTTP/2 server does not implement the
		// extended CONNECT of RFC 8441, so a websocket cannot run on an h2
		// connection, and keyboard and mouse ride the websocket at /api/ws.
		TLSNextProto: map[string]func(*http.Server, *tls.Conn, http.Handler){},
	}
}

// AllowHTTP2 offers HTTP/2 on a server that has not started yet. It is the
// `http2: true` setting in server.yaml, and it is off by default because on
// this hardware it is not a trade-off with two sides.
//
// Nothing here multiplexes. The board serves one video stream to one viewer
// over one connection, so h2's frame scheduler, per-stream write queues and
// flow-control accounting are cost against no benefit, and the C906 has none
// spare: measured 2026-09-06 at 1080p, MJPEG carried 10.5MB/s over HTTP/1.1 at
// 69% of the core and 3.5MB/s over h2 with TLS at 97%, which took the delivered
// frame rate from 30 to about 10.
//
// It also takes the keyboard and the mouse away. Go's HTTP/2 server does not
// implement the extended CONNECT of RFC 8441, so a websocket cannot run on an
// h2 connection, and /api/ws is where HID rides. The page still loads, the
// video still plays, and nothing anywhere reports an error: this is the failure
// that reads as "HTTPS breaks the keyboard".
//
// The setting exists for the deployment this does not describe, such as a proxy
// in front of the board that requires h2 to the origin. Anyone turning it on
// should expect to lose HID unless something else terminates the connection.
func AllowHTTP2(server *http.Server) {
	server.TLSNextProto = nil
}
