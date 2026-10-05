// NanoKVM addition to crypto/tls, applied through server/goroot-overlay
// (ironkvm-dist#68). Not part of the Go distribution.

package tls

// gatherState is a connection's gathering of application data records: while
// on, each record is built and sealed in buf, where it is sent from, and buf
// goes to the socket once it holds flushAt bytes and at EndGather.
//
// Why (ironkvm-dist#68, run sheet trial 37). Over HTTPS a 1080p MJPEG picture
// is about 270 kB, seventeen 16 kB records. Without this each record is
// copied next to its header (outBuf), sealed there, and copied again into the
// server's write batch (utils.batchConn), which gathers the records into 64 kB
// socket writes. On a riscv64 core both copies run a byte at a time, because
// the runtime's memmove does that whenever source and destination differ in
// their alignment modulo 8, and a record header is 5 bytes and a record 16406.
// The two copies took 8 to 9% of the C906 at 1080p, the batch copy alone 6 to
// 8%. Here the record is sealed straight from the caller's bytes into the
// buffer the socket is written from, so neither copy is made.
type gatherState struct {
	on      bool
	flushAt int
	buf     []byte
}

// gatherSlack is what a record adds to its payload at most: header, a TLS 1.2
// explicit nonce, content type and tag, or a CBC MAC and padding.
const gatherSlack = recordHeaderLen + 16 + 1 + 64 + 256

// BeginGather gathers the application data that Write sends from now until
// EndGather, and writes it to the socket in pieces of at least flushAt bytes.
// Every other record (an alert, a key update) first sends what is gathered,
// so the order on the wire is the order of the calls.
//
// The buffer is kept with the connection between EndGather and the next
// BeginGather; it is flushAt plus one record.
func (c *Conn) BeginGather(flushAt int) {
	if flushAt < maxPlaintext {
		flushAt = maxPlaintext
	}

	c.out.Lock()
	defer c.out.Unlock()

	c.gather.on = true
	c.gather.flushAt = flushAt
	if need := flushAt + maxPlaintext + gatherSlack; cap(c.gather.buf) < need {
		c.gather.buf = make([]byte, 0, need)
	}
}

// EndGather sends what is gathered and returns Write to sending each record
// on its own.
func (c *Conn) EndGather() error {
	c.out.Lock()
	defer c.out.Unlock()

	c.gather.on = false
	if err := c.out.err; err != nil {
		c.gather.buf = c.gather.buf[:0]
		return err
	}
	if err := c.flushGatheredLocked(); err != nil {
		return c.out.setErrorLocked(err)
	}

	return nil
}

// flushGatheredLocked writes what is gathered to the socket.
func (c *Conn) flushGatheredLocked() error {
	if len(c.gather.buf) == 0 {
		return nil
	}

	n, err := c.conn.Write(c.gather.buf)
	c.bytesSent += int64(n)
	c.gather.buf = c.gather.buf[:0]

	return err
}

// writeRecordGathered is writeRecordLocked for application data while
// gathering.
func (c *Conn) writeRecordGathered(data []byte) (int, error) {
	vers := c.vers
	if vers == 0 {
		vers = VersionTLS10
	} else if vers == VersionTLS13 {
		vers = VersionTLS12
	}

	var n int
	for len(data) > 0 {
		m := len(data)
		if maxPayload := c.maxPayloadSizeForWrite(recordTypeApplicationData); m > maxPayload {
			m = maxPayload
		}

		buf := c.gather.buf
		if cap(buf)-len(buf) < m+gatherSlack {
			// Not reached with the sizes BeginGather sets, but a record
			// must never be sealed into a buffer that append then moves.
			grown := make([]byte, len(buf), len(buf)+m+gatherSlack+c.gather.flushAt)
			copy(grown, buf)
			buf = grown
		}

		start := len(buf)
		record := buf[start : start+recordHeaderLen]
		record[0] = byte(recordTypeApplicationData)
		record[1] = byte(vers >> 8)
		record[2] = byte(vers)
		record[3] = byte(m >> 8)
		record[4] = byte(m)

		sealed, ok := c.out.sealTLS13Direct(record, data[:m])
		if !ok {
			var err error
			sealed, err = c.out.encrypt(record, data[:m], c.config.rand())
			if err != nil {
				c.gather.buf = buf[:start]
				return n, err
			}
		}
		if &sealed[0] != &buf[start : start+1][0] {
			panic("tls: internal error: gathered record moved")
		}
		c.gather.buf = buf[:start+len(sealed)]

		n += m
		data = data[m:]

		if len(c.gather.buf) >= c.gather.flushAt {
			if err := c.flushGatheredLocked(); err != nil {
				return n, err
			}
		}
	}

	return n, nil
}

// tls13Sealer is the AEAD that can seal a TLS 1.3 inner plaintext, the
// payload followed by its content type, without the two being next to each
// other in memory. The ChaCha20-Poly1305 of the vendored x/crypto has it with
// server/goroot-overlay applied.
type tls13Sealer interface {
	SealTLS13(dst, nonce, plaintext []byte, contentType byte, additionalData []byte) []byte
}

// sealTLS13Direct seals payload into record, which holds the header and has
// room after it, reading the payload where it is. It is encrypt's TLS 1.3
// AEAD case without the copy of the payload next to the header that encrypt
// makes before sealing in place. ok is false when the connection or its
// cipher is not one this applies to, and nothing has changed.
func (hc *halfConn) sealTLS13Direct(record, payload []byte) (sealed []byte, ok bool) {
	if hc.version != VersionTLS13 {
		return nil, false
	}
	x, isXor := hc.cipher.(*xorNonceAEAD)
	if !isXor {
		return nil, false
	}
	s, isSealer := x.aead.(tls13Sealer)
	if !isSealer {
		return nil, false
	}

	// As encrypt: the real content type goes inside, the outer one says
	// application data, and the header carries the sealed length.
	contentType := record[0]
	record[0] = byte(recordTypeApplicationData)
	n := len(payload) + 1 + x.Overhead()
	record[3] = byte(n >> 8)
	record[4] = byte(n)

	// As xorNonceAEAD.Seal, with the sequence number as the nonce.
	for i, b := range hc.seq {
		x.nonceMask[4+i] ^= b
	}
	sealed = s.SealTLS13(record[:recordHeaderLen], x.nonceMask[:], payload, contentType, record[:recordHeaderLen])
	for i, b := range hc.seq {
		x.nonceMask[4+i] ^= b
	}
	hc.incSeq()

	return sealed, true
}
