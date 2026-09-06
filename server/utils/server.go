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

		// An empty non-nil map turns HTTP/2 off. ListenAndServeTLS configures
		// it automatically when this field is nil, so the HTTPS listener spoke
		// h2 while the plain one spoke HTTP/1.1, and the video stream paid for
		// the difference.
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
