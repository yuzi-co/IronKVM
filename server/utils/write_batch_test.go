package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
)

// recordingConn keeps every write it is given.
type recordingConn struct {
	net.Conn
	mu     sync.Mutex
	writes [][]byte
}

func (c *recordingConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writes = append(c.writes, append([]byte(nil), p...))
	return len(p), nil
}

func (c *recordingConn) joined() []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	return bytes.Join(c.writes, nil)
}

func TestBatchConnPassesWritesThroughOutsideABatch(t *testing.T) {
	raw := &recordingConn{}
	conn := &batchConn{Conn: raw}

	conn.Write([]byte("one"))
	conn.Write([]byte("two"))

	if len(raw.writes) != 2 {
		t.Fatalf("%d socket writes, want 2: nothing should wait outside a batch", len(raw.writes))
	}
}

func TestBatchConnGathersABatch(t *testing.T) {
	raw := &recordingConn{}
	conn := &batchConn{Conn: raw}

	conn.begin()
	for i := 0; i < 3; i++ {
		if n, err := conn.Write(bytes.Repeat([]byte{byte(i)}, 1000)); n != 1000 || err != nil {
			t.Fatalf("write %d: %d, %v", i, n, err)
		}
	}
	if len(raw.writes) != 0 {
		t.Fatalf("%d socket writes before the batch ended", len(raw.writes))
	}
	if err := conn.end(); err != nil {
		t.Fatal(err)
	}
	if len(raw.writes) != 1 || len(raw.writes[0]) != 3000 {
		t.Fatalf("got %d socket writes, want one of 3000 bytes", len(raw.writes))
	}

	// After the batch, writes pass through again.
	conn.Write([]byte("after"))
	if len(raw.writes) != 2 {
		t.Fatalf("%d socket writes, want 2", len(raw.writes))
	}
}

func TestBatchConnFlushesWhenFull(t *testing.T) {
	raw := &recordingConn{}
	conn := &batchConn{Conn: raw}

	record := make([]byte, 16384)
	rand.Read(record)
	var want []byte

	conn.begin()
	for i := 0; i < 20; i++ {
		conn.Write(record)
		want = append(want, record...)
	}
	conn.end()

	for i, w := range raw.writes {
		if len(w) > batchFlushSize+tlsRecordMax {
			t.Errorf("socket write %d holds %d bytes, more than one flush and a record", i, len(w))
		}
	}
	if len(raw.writes) < 4 || len(raw.writes) > 6 {
		t.Errorf("%d socket writes for 320 kB, want about 5", len(raw.writes))
	}
	if !bytes.Equal(raw.joined(), want) {
		t.Error("the bytes on the socket differ from the bytes written")
	}
}

// countingListener counts the socket writes on every connection it accepts.
type countingListener struct {
	net.Listener
	writes *atomic.Int64
}

type countingConn struct {
	net.Conn
	writes *atomic.Int64
}

func (c countingConn) Write(p []byte) (int, error) {
	c.writes.Add(1)
	return c.Conn.Write(p)
}

func (l countingListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return countingConn{conn, l.writes}, nil
}

// socketWritesFor serves one 320 kB response over TLS and reports how many
// socket writes it took, with and without a batch around it.
func socketWritesFor(t *testing.T, batch bool) int64 {
	t.Helper()

	body := make([]byte, 320<<10)
	rand.Read(body)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var writes atomic.Int64

	server := NewServer("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		end := noBatch
		if batch {
			end = BeginWriteBatch(r)
		}
		w.Write(body)
		w.(http.Flusher).Flush()
		if err := end(); err != nil {
			t.Error(err)
		}
	}))
	server.TLSConfig = &tls.Config{Certificates: []tls.Certificate{selfSignedCert(t)}}
	go func() {
		_ = server.ServeTLS(batchListener{countingListener{listener, &writes}}, "", "")
	}()
	defer server.Close()

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	defer client.CloseIdleConnections()

	// Warm the connection up so the handshake is not in the count.
	resp, err := client.Get("https://" + listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	before := writes.Load()
	resp, err = client.Get("https://" + listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !bytes.Equal(got, body) {
		t.Fatalf("received %d bytes that differ from the %d sent", len(got), len(body))
	}

	return writes.Load() - before
}

func TestBeginWriteBatchGathersTLSRecords(t *testing.T) {
	without := socketWritesFor(t, false)
	with := socketWritesFor(t, true)

	// 320 kB is twenty full records, each its own write without a batch.
	if without < 20 {
		t.Errorf("%d socket writes without a batch; expected one per record", without)
	}
	if with > 7 {
		t.Errorf("%d socket writes with a batch, want about 5 (64 kB each)", with)
	}
}

func TestBeginWriteBatchOnPlainHTTPDoesNothing(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	if err := BeginWriteBatch(req)(); err != nil {
		t.Fatal(err)
	}
}
