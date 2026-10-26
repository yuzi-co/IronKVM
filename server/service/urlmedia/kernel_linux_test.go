//go:build linux

package urlmedia

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

// TestKernelNBDServesTheURL attaches a real nbd device to an image on an
// httptest server and reads it back through the kernel. It needs root and
// the nbd module, so it runs only with URLMEDIA_NBD_TEST=1.
func TestKernelNBDServesTheURL(t *testing.T) {
	if os.Getenv("URLMEDIA_NBD_TEST") != "1" || os.Geteuid() != 0 {
		t.Skip("set URLMEDIA_NBD_TEST=1 and run as root")
	}
	if err := kernelNBDAvailable(); err != nil {
		t.Fatalf("nbd: %v", err)
	}

	data := testImage(5<<20 + 512*3)
	srv := rangeServer(t, data)
	src, err := Open(context.Background(), srv.URL+"/x.iso", testOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	cache := NewCache(src, DefaultBlockSize, DefaultBlocks, DefaultReadAhead)
	defer cache.Close()

	dev, conn, err := attachNBD(src.Size())
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	served := make(chan error, 1)
	go func() {
		served <- serveNBD(context.Background(), conn, cache)
		_ = conn.Close()
	}()
	t.Logf("attached %s", dev.Path())

	f, err := os.Open(dev.Path())
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(f)
	_ = f.Close()
	if err != nil {
		t.Fatalf("read %s: %v", dev.Path(), err)
	}
	if !bytes.Equal(got, data) {
		kd := dev.(*kernelDevice)
		select {
		case <-kd.done:
			t.Logf("NBD_DO_IT returned %v", kd.doItErr)
		default:
		}
		t.Fatalf("read %d bytes through the kernel, want the %d of the image", len(got), len(data))
	}

	// Read-only: the kernel refuses to open it for writing.
	if w, err := os.OpenFile(dev.Path(), os.O_WRONLY, 0); err == nil {
		_, werr := w.Write(make([]byte, 512))
		serr := w.Sync()
		_ = w.Close()
		if werr == nil && serr == nil {
			t.Fatal("a write to the read-only device succeeded")
		}
	}

	// Signals to the process must not reach the thread in NBD_DO_IT: the
	// kernel would shut the device down. The server starts child processes
	// (SIGCHLD) and Go preempts with SIGURG.
	for i := 0; i < 20; i++ {
		_ = exec.Command("true").Run()
		_ = syscall.Kill(os.Getpid(), syscall.SIGURG)
	}
	time.Sleep(100 * time.Millisecond)
	select {
	case <-dev.(*kernelDevice).done:
		t.Fatalf("the device stopped after signals: %v", dev.(*kernelDevice).doItErr)
	default:
	}

	st := cache.Stats()
	t.Logf("cache: %+v, bytes over HTTP %d", st, src.BytesRead())

	start := time.Now()
	if err := dev.Disconnect(); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
	select {
	case err := <-served:
		if err != nil {
			t.Fatalf("serve: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("serve did not end after the disconnect")
	}
	t.Logf("disconnect took %s", time.Since(start))
}
