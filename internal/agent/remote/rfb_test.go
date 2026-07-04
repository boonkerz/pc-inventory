package remote

import (
	"context"
	"encoding/binary"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"
)

func TestRFBServeHandshakeAndFrame(t *testing.T) {
	srv, cli := net.Pipe()
	defer cli.Close()

	go func() {
		_ = rfbServe(context.Background(), srv, newSyntheticSource(), slog.Default())
		srv.Close()
	}()

	_ = cli.SetDeadline(time.Now().Add(5 * time.Second))
	readN := func(n int) []byte {
		b := make([]byte, n)
		if _, err := io.ReadFull(cli, b); err != nil {
			t.Fatalf("read %d: %v", n, err)
		}
		return b
	}

	// ProtocolVersion
	if got := string(readN(12)); got != "RFB 003.008\n" {
		t.Fatalf("version: %q", got)
	}
	if _, err := cli.Write([]byte("RFB 003.008\n")); err != nil {
		t.Fatal(err)
	}
	// Security: [count=1, type=1]
	sec := readN(2)
	if sec[0] != 1 || sec[1] != 1 {
		t.Fatalf("security: %v", sec)
	}
	cli.Write([]byte{1})    // gewählter Typ
	readN(4)                // SecurityResult
	cli.Write([]byte{1})    // ClientInit shared
	// ServerInit
	w := int(binary.BigEndian.Uint16(readN(2)))
	h := int(binary.BigEndian.Uint16(readN(2)))
	if w != 1280 || h != 720 {
		t.Fatalf("dim %dx%d", w, h)
	}
	readN(16) // pixel format
	nameLen := int(binary.BigEndian.Uint32(readN(4)))
	readN(nameLen)

	// FramebufferUpdateRequest
	req := []byte{3, 0}
	req = append(req, 0, 0, 0, 0) // x,y
	req = binary.BigEndian.AppendUint16(req, uint16(w))
	req = binary.BigEndian.AppendUint16(req, uint16(h))
	if _, err := cli.Write(req); err != nil {
		t.Fatal(err)
	}
	// FramebufferUpdate
	upd := readN(4)
	if upd[0] != 0 {
		t.Fatalf("kein FramebufferUpdate: %d", upd[0])
	}
	if nr := binary.BigEndian.Uint16(upd[2:]); nr != 1 {
		t.Fatalf("num-rects %d", nr)
	}
	rect := readN(12)
	rw := int(binary.BigEndian.Uint16(rect[4:]))
	rh := int(binary.BigEndian.Uint16(rect[6:]))
	enc := binary.BigEndian.Uint32(rect[8:])
	if rw != w || rh != h || enc != 0 {
		t.Fatalf("rect %dx%d enc=%d", rw, rh, enc)
	}
	// Pixel lesen (w*h*4)
	if _, err := io.ReadFull(cli, make([]byte, w*h*4)); err != nil {
		t.Fatalf("pixel: %v", err)
	}
}
