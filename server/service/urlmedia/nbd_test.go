//go:build linux

package urlmedia

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"syscall"
	"testing"
	"time"
)

// socketPair is the pair the kernel gets one end of: here the test plays
// the kernel on the other end.
func socketPair(t *testing.T) (server, client net.Conn) {
	t.Helper()
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatal(err)
	}
	conn := func(fd int) net.Conn {
		f := os.NewFile(uintptr(fd), "nbd")
		c, err := net.FileConn(f)
		_ = f.Close()
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	server, client = conn(fds[0]), conn(fds[1])
	t.Cleanup(func() { _ = server.Close(); _ = client.Close() })
	return server, client
}

// fakeKernel speaks the transmission phase as the kernel's nbd driver does.
type fakeKernel struct {
	t    *testing.T
	conn net.Conn
}

func (k fakeKernel) send(cmd uint16, handle uint64, off uint64, length uint32, payload []byte) {
	k.t.Helper()
	var req [28]byte
	binary.BigEndian.PutUint32(req[0:], nbdRequestMagic)
	binary.BigEndian.PutUint16(req[4:], 0)
	binary.BigEndian.PutUint16(req[6:], cmd)
	binary.BigEndian.PutUint64(req[8:], handle)
	binary.BigEndian.PutUint64(req[16:], off)
	binary.BigEndian.PutUint32(req[24:], length)
	if _, err := k.conn.Write(append(req[:], payload...)); err != nil {
		k.t.Fatalf("send: %v", err)
	}
}

// reply reads one simple reply, and dataLen bytes after it when it is not
// an error.
func (k fakeKernel) reply(dataLen int) (errno uint32, handle uint64, data []byte) {
	k.t.Helper()
	_ = k.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var hdr [16]byte
	if _, err := io.ReadFull(k.conn, hdr[:]); err != nil {
		k.t.Fatalf("read reply: %v", err)
	}
	if m := binary.BigEndian.Uint32(hdr[0:]); m != nbdReplyMagic {
		k.t.Fatalf("reply magic %#x", m)
	}
	errno = binary.BigEndian.Uint32(hdr[4:])
	handle = binary.BigEndian.Uint64(hdr[8:])
	if errno == 0 && dataLen > 0 {
		data = make([]byte, dataLen)
		if _, err := io.ReadFull(k.conn, data); err != nil {
			k.t.Fatalf("read data: %v", err)
		}
	}
	return errno, handle, data
}

func serveInBackground(t *testing.T, conn net.Conn, b backend) chan error {
	done := make(chan error, 1)
	go func() { done <- serveNBD(context.Background(), conn, b) }()
	return done
}

func TestNBDServesReads(t *testing.T) {
	server, client := socketPair(t)
	src := &memSource{data: testImage(3 << 20)}
	done := serveInBackground(t, server, src)
	k := fakeKernel{t, client}

	k.send(nbdCmdRead, 42, 4096, 8192, nil)
	errno, handle, data := k.reply(8192)
	if errno != 0 || handle != 42 {
		t.Fatalf("errno %d handle %d", errno, handle)
	}
	if !bytes.Equal(data, src.data[4096:4096+8192]) {
		t.Fatal("wrong bytes")
	}

	// Larger than one buffer: streamed in pieces.
	k.send(nbdCmdRead, 43, 0, 3<<20, nil)
	errno, _, data = k.reply(3 << 20)
	if errno != 0 || !bytes.Equal(data, src.data) {
		t.Fatalf("large read: errno %d, equal %t", errno, bytes.Equal(data, src.data))
	}

	k.send(nbdCmdDisc, 44, 0, 0, nil)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("serve after a disconnect: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return after a disconnect")
	}
}

func TestNBDRefusesWritesWithEPERM(t *testing.T) {
	server, client := socketPair(t)
	src := &memSource{data: testImage(1 << 16)}
	serveInBackground(t, server, src)
	k := fakeKernel{t, client}

	k.send(nbdCmdWrite, 7, 0, 512, make([]byte, 512))
	if errno, handle, _ := k.reply(0); errno != nbdEPERM || handle != 7 {
		t.Fatalf("write: errno %d handle %d, want EPERM", errno, handle)
	}
	k.send(nbdCmdTrim, 8, 0, 512, nil)
	if errno, _, _ := k.reply(0); errno != nbdEPERM {
		t.Fatalf("trim: errno %d, want EPERM", errno)
	}
	// The write's payload was read off the socket: the next request parses.
	k.send(nbdCmdFlush, 9, 0, 0, nil)
	if errno, handle, _ := k.reply(0); errno != 0 || handle != 9 {
		t.Fatalf("flush: errno %d handle %d", errno, handle)
	}
	k.send(nbdCmdRead, 10, 0, 16, nil)
	if errno, _, data := k.reply(16); errno != 0 || !bytes.Equal(data, src.data[:16]) {
		t.Fatalf("read after a write: errno %d", errno)
	}
}

func TestNBDAnswersEIOWhenTheSourceFails(t *testing.T) {
	server, client := socketPair(t)
	src := &memSource{data: testImage(1 << 16), fail: errors.New("network down")}
	serveInBackground(t, server, src)
	k := fakeKernel{t, client}

	k.send(nbdCmdRead, 1, 0, 4096, nil)
	if errno, _, _ := k.reply(4096); errno != nbdEIO {
		t.Fatalf("errno %d, want EIO", errno)
	}
	src.setFail(nil)
	k.send(nbdCmdRead, 2, 0, 4096, nil)
	if errno, _, data := k.reply(4096); errno != 0 || !bytes.Equal(data, src.data[:4096]) {
		t.Fatalf("after recovery: errno %d", errno)
	}
}

func TestNBDAnswersEINVALPastTheEnd(t *testing.T) {
	server, client := socketPair(t)
	src := &memSource{data: testImage(4096)}
	serveInBackground(t, server, src)
	k := fakeKernel{t, client}

	k.send(nbdCmdRead, 1, 4000, 200, nil)
	if errno, _, _ := k.reply(200); errno != nbdEINVAL {
		t.Fatalf("errno %d, want EINVAL", errno)
	}
	k.send(99, 2, 0, 0, nil)
	if errno, _, _ := k.reply(0); errno != nbdEINVAL {
		t.Fatalf("unknown command: errno %d, want EINVAL", errno)
	}
}

func TestNBDStopsOnABadMagic(t *testing.T) {
	server, client := socketPair(t)
	done := serveInBackground(t, server, &memSource{data: testImage(4096)})

	_, _ = client.Write(make([]byte, 28))
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("serve accepted a request without the magic")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not stop on a bad magic")
	}
}

func TestNBDReturnsWhenTheSocketCloses(t *testing.T) {
	server, client := socketPair(t)
	done := serveInBackground(t, server, &memSource{data: testImage(4096)})
	_ = client.Close()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not return when the kernel's end closed")
	}
}
