package urlmedia

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

// The transmission phase of the NBD protocol, the only phase there is when
// the kernel gets its socket through NBD_SET_SOCK: no handshake, no options,
// simple replies only. See the protocol document in the NBD project
// (doc/proto.md); this is written from it, not from another server.
const (
	nbdRequestMagic = 0x25609513
	nbdReplyMagic   = 0x67446698

	nbdCmdRead  = 0
	nbdCmdWrite = 1
	nbdCmdDisc  = 2
	nbdCmdFlush = 3
	nbdCmdTrim  = 4

	nbdEPERM  = 1
	nbdEIO    = 5
	nbdEINVAL = 22

	nbdRequestLen = 28
)

// maxBuffered is the largest read answered from one buffer. A larger one,
// which the kernel only sends if its queue limits allow, is streamed in
// pieces of this size; an error after the first piece can no longer be
// reported in the reply, so it closes the connection, and the kernel fails
// the request.
const maxBuffered = 1 << 20

// maxWriteDrain bounds the payload of a write read off the socket before it
// is refused, so a confused peer cannot make the server read forever.
const maxWriteDrain = 64 << 20

var errBadRequest = errors.New("nbd: bad request")

// serveNBD answers the kernel's requests on conn until it disconnects or
// conn fails. The image is read-only: writes and trims get EPERM.
func serveNBD(ctx context.Context, conn io.ReadWriter, b backend) error {
	var hdr [nbdRequestLen]byte
	// Grown to the largest read seen, up to maxBuffered.
	var buf []byte
	for {
		if _, err := io.ReadFull(conn, hdr[:]); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if binary.BigEndian.Uint32(hdr[0:]) != nbdRequestMagic {
			return errBadRequest
		}
		cmd := binary.BigEndian.Uint16(hdr[6:])
		handle := binary.BigEndian.Uint64(hdr[8:])
		off := binary.BigEndian.Uint64(hdr[16:])
		length := binary.BigEndian.Uint32(hdr[24:])

		switch cmd {
		case nbdCmdDisc:
			return nil
		case nbdCmdRead:
			if err := serveRead(ctx, conn, b, handle, off, length, &buf); err != nil {
				return err
			}
		case nbdCmdWrite:
			if length > maxWriteDrain {
				return fmt.Errorf("%w: a write of %d bytes", errBadRequest, length)
			}
			if _, err := io.CopyN(io.Discard, conn, int64(length)); err != nil {
				return err
			}
			if err := reply(conn, nbdEPERM, handle); err != nil {
				return err
			}
		case nbdCmdFlush:
			if err := reply(conn, 0, handle); err != nil {
				return err
			}
		case nbdCmdTrim:
			if err := reply(conn, nbdEPERM, handle); err != nil {
				return err
			}
		default:
			if err := reply(conn, nbdEINVAL, handle); err != nil {
				return err
			}
		}
	}
}

func reply(w io.Writer, errno uint32, handle uint64) error {
	var r [16]byte
	binary.BigEndian.PutUint32(r[0:], nbdReplyMagic)
	binary.BigEndian.PutUint32(r[4:], errno)
	binary.BigEndian.PutUint64(r[8:], handle)
	_, err := w.Write(r[:])
	return err
}

func serveRead(ctx context.Context, conn io.Writer, b backend, handle, off uint64, length uint32, buf *[]byte) error {
	size := uint64(b.Size())
	if off > size || uint64(length) > size-off {
		return reply(conn, nbdEINVAL, handle)
	}
	// The first piece is read before the reply goes out, so a read that
	// fails at once gets an EIO, which the host sees as one failed command.
	first := min(int(length), maxBuffered)
	if cap(*buf) < first {
		*buf = make([]byte, first)
	}
	p := (*buf)[:first]
	if err := b.ReadAt(ctx, p, int64(off)); err != nil {
		log.Warnf("urlmedia: read %d bytes at %d: %s", length, off, err)
		return reply(conn, nbdEIO, handle)
	}
	if err := reply(conn, 0, handle); err != nil {
		return err
	}
	if _, err := conn.Write(p); err != nil {
		return err
	}
	for done := uint64(first); done < uint64(length); {
		n := min(uint64(length)-done, maxBuffered)
		p = (*buf)[:n]
		if err := b.ReadAt(ctx, p, int64(off+done)); err != nil {
			return fmt.Errorf("read %d bytes at %d after the reply: %w", n, off+done, err)
		}
		if _, err := conn.Write(p); err != nil {
			return err
		}
		done += n
	}
	return nil
}
