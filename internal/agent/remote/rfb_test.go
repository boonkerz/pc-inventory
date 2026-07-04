package remote

import (
	"bytes"
	"context"
	"encoding/binary"
	"image/jpeg"
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
	cli.Write([]byte{1}) // gewählter Typ
	readN(4)             // SecurityResult
	cli.Write([]byte{1}) // ClientInit shared
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

// TestRFBTightJPEG prüft, dass ein großer Bereich als gültiges Tight-JPEG
// gesendet wird (Format + dekodierbares JPEG), wenn der Client Tight anbietet.
func TestRFBTightJPEG(t *testing.T) {
	srv, cli := net.Pipe()
	defer cli.Close()
	go func() { _ = rfbServe(context.Background(), srv, newSyntheticSource(), slog.Default()); srv.Close() }()

	_ = cli.SetDeadline(time.Now().Add(5 * time.Second))
	rd := func(n int) []byte { b := make([]byte, n); io.ReadFull(cli, b); return b }
	rd(12)
	cli.Write([]byte("RFB 003.008\n"))
	rd(2)
	cli.Write([]byte{1})
	rd(4)
	cli.Write([]byte{1})
	w := int(binary.BigEndian.Uint16(rd(2)))
	h := int(binary.BigEndian.Uint16(rd(2)))
	rd(16)
	rd(int(binary.BigEndian.Uint32(rd(4))))

	// SetEncodings: Tight(7) + Raw(0)
	se := []byte{2, 0}
	se = binary.BigEndian.AppendUint16(se, 2)
	se = binary.BigEndian.AppendUint32(se, 7)
	se = binary.BigEndian.AppendUint32(se, 0)
	cli.Write(se)

	// Voll-Update anfordern (nicht-incremental).
	req := []byte{3, 0, 0, 0, 0, 0}
	req = binary.BigEndian.AppendUint16(req, uint16(w))
	req = binary.BigEndian.AppendUint16(req, uint16(h))
	cli.Write(req)

	if upd := rd(4); upd[0] != 0 || binary.BigEndian.Uint16(upd[2:]) != 1 {
		t.Fatalf("kein Update: %v", upd)
	}
	rect := rd(12)
	if enc := binary.BigEndian.Uint32(rect[8:]); enc != 7 {
		t.Fatalf("erwartete Tight(7), bekam %d", enc)
	}
	if ctrl := rd(1); ctrl[0] != 0x90 {
		t.Fatalf("erwartete JPEG-Control 0x90, bekam 0x%x", ctrl[0])
	}
	// kompakte Länge lesen
	n := 0
	shift := 0
	for {
		b := rd(1)[0]
		n |= int(b&0x7f) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	jpg := rd(n)
	img, err := jpeg.Decode(bytesReader(jpg))
	if err != nil {
		t.Fatalf("JPEG dekodieren: %v", err)
	}
	if img.Bounds().Dx() != w || img.Bounds().Dy() != h {
		t.Fatalf("JPEG-Größe %v, erwartet %dx%d", img.Bounds(), w, h)
	}
}

func bytesReader(b []byte) *bytes.Reader { return bytes.NewReader(b) }
