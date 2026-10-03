package utils

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
)

// batchFlushSize is how much a batch holds before it goes to the socket
// anyway. Measured on the board (ironkvm-dist#60), 64 kB socket writes cost
// the kernel as little as one write per 300 kB picture did, so a larger batch
// buys nothing and only holds more memory per viewer.
const batchFlushSize = 64 << 10

// batchConn sits between a TLS connection and its TCP socket and, while a
// batch is open, gathers the TLS records into larger socket writes.
//
// crypto/tls hands every record to the socket on its own, and a record holds
// at most 16 kB, so a 300 kB MJPEG picture went out as about twenty writes.
// On the C906 with the 5.10 kernel those cost the kernel about twice the
// system time of the same bytes in large writes: measured at 1080p, gathering
// them took MJPEG over HTTPS from 17.4 to 21.5 fps, with system time down from
// 30% to 16% of the core. On the 7.2 kernel small writes cost less and the
// gain is smaller (17.2 to 18.7 fps), as the copy into the batch costs about
// what it saves.
//
// Outside a batch every write passes straight through, so the handshake,
// ordinary responses and websockets behave exactly as before. Only a handler
// that opens a batch (BeginWriteBatch) changes anything, and only for its own
// connection.
type batchConn struct {
	net.Conn

	// mu keeps the socket writes in order. A write from another goroutine
	// (a TLS close alert, say) that arrives during a batch joins it rather
	// than overtaking the records ahead of it.
	mu    sync.Mutex
	open  bool
	batch []byte
}

func (c *batchConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.open {
		return c.Conn.Write(p)
	}

	c.batch = append(c.batch, p...)
	if len(c.batch) >= batchFlushSize {
		if err := c.flushLocked(); err != nil {
			return 0, err
		}
	}

	return len(p), nil
}

func (c *batchConn) begin() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.open = true
	if c.batch == nil {
		// One flush's worth plus a record, so a batch never regrows.
		c.batch = make([]byte, 0, batchFlushSize+tlsRecordMax)
	}
}

func (c *batchConn) end() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.open = false

	return c.flushLocked()
}

func (c *batchConn) flushLocked() error {
	if len(c.batch) == 0 {
		return nil
	}

	_, err := c.Conn.Write(c.batch)
	c.batch = c.batch[:0]

	return err
}

// tlsRecordMax is the largest record crypto/tls writes: 16 kB of payload plus
// header, content type and AEAD tag.
const tlsRecordMax = 16384 + 5 + 1 + 16

// batchListener wraps every accepted connection in a batchConn.
type batchListener struct {
	net.Listener
}

func (l batchListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return &batchConn{Conn: conn}, nil
}

type connContextKey struct{}

// rememberConn is the server's ConnContext: it puts the connection in every
// request's context, where BeginWriteBatch looks for it.
func rememberConn(ctx context.Context, conn net.Conn) context.Context {
	return context.WithValue(ctx, connContextKey{}, conn)
}

// BeginWriteBatch gathers what the handler writes to r's connection from now
// until the returned function is called, and that function sends it. The
// handler must flush its response before calling it, or the last bytes are
// still in net/http's buffer and miss the batch.
//
// It only does anything on a TLS listener started by ListenAndServeTLS. On a
// plain listener a large write already reaches the socket in one piece, and
// there the returned function does nothing.
func BeginWriteBatch(r *http.Request) (end func() error) {
	conn, _ := r.Context().Value(connContextKey{}).(*tls.Conn)
	if conn == nil {
		return noBatch
	}

	batch, ok := conn.NetConn().(*batchConn)
	if !ok {
		return noBatch
	}

	batch.begin()

	return batch.end
}

func noBatch() error {
	return nil
}

// ListenAndServeTLS is server.ListenAndServeTLS with the connections wrapped
// for BeginWriteBatch.
func ListenAndServeTLS(server *http.Server, certFile, keyFile string) error {
	addr := server.Addr
	if addr == "" {
		addr = ":https"
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return server.ServeTLS(batchListener{listener}, certFile, keyFile)
}
