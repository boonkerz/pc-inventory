package remote

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"log/slog"
)

// Eigener, minimaler RFB/VNC-Server (view-only): überträgt den Bildschirm als
// Raw-Encoding an einen RFB-Client (noVNC). Keine Fremd-Software, keine Bundles –
// der „VNC-Server" ist der Agent selbst. Eingaben (Maus/Tastatur) folgen später.
//
// Ablauf (RFB 3.8, Security = None): ProtocolVersion → Security → ClientInit →
// ServerInit → Schleife aus FramebufferUpdateRequest → FramebufferUpdate.

// screenSource liefert den aktuellen Bildschirminhalt als 32-bpp-Pixel in der
// Reihenfolge B,G,R,X (little-endian 0x00RRGGBB) – passend zum unten deklarierten
// Pixelformat.
type screenSource interface {
	Bounds() (w, h int)
	Capture() ([]byte, error) // Länge = w*h*4
	Close() error
}

// inputSink nimmt Maus-/Tastatureingaben entgegen. Quellen, die das umsetzen können
// (Windows), spielen sie ins System ein; sonst bleibt die Sitzung view-only.
// x,y sind Framebuffer-Koordinaten; keysym ist ein X11-Keysym (RFB).
type inputSink interface {
	Pointer(buttonMask, x, y int)
	Key(down bool, keysym uint32)
}

// rfbPixelFormat: 32 bpp, Tiefe 24, little-endian, TrueColor, Shifts R=16/G=8/B=0.
func rfbPixelFormat() []byte {
	return []byte{
		32, 24, 0, 1, // bpp, depth, big-endian=0, true-colour=1
		0, 255, 0, 255, 0, 255, // red/green/blue max (je 2 Byte BE)
		16, 8, 0, // red/green/blue shift
		0, 0, 0, // padding
	}
}

func rfbServe(ctx context.Context, conn io.ReadWriter, src screenSource, log *slog.Logger) error {
	w, h := src.Bounds()

	// 1. ProtocolVersion
	if _, err := conn.Write([]byte("RFB 003.008\n")); err != nil {
		return err
	}
	if err := skip(conn, 12); err != nil {
		return err
	}
	// 2. Security: nur „None" (1). Tunnel + RBAC sind die Sicherheitsgrenze.
	if _, err := conn.Write([]byte{1, 1}); err != nil {
		return err
	}
	if err := skip(conn, 1); err != nil { // gewählter Security-Typ
		return err
	}
	if _, err := conn.Write([]byte{0, 0, 0, 0}); err != nil { // SecurityResult OK
		return err
	}
	// 3. ClientInit (shared-flag) ignorieren
	if err := skip(conn, 1); err != nil {
		return err
	}
	// 4. ServerInit
	si := make([]byte, 0, 32)
	si = be16(si, w)
	si = be16(si, h)
	si = append(si, rfbPixelFormat()...)
	name := []byte("PC-Inventory")
	si = be32(si, len(name))
	si = append(si, name...)
	if _, err := conn.Write(si); err != nil {
		return err
	}

	var lastFrame time.Time
	var prev []byte // letztes gesendetes Vollbild (für Dirty-Rectangle-Diff)
	hdr := make([]byte, 1)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if _, err := io.ReadFull(conn, hdr); err != nil {
			return err
		}
		switch hdr[0] {
		case 0: // SetPixelFormat (3 padding + 16 pf) – wir bleiben bei unserem Format
			if err := skip(conn, 19); err != nil {
				return err
			}
		case 2: // SetEncodings (1 padding + 2 count + count*4)
			b := make([]byte, 3)
			if _, err := io.ReadFull(conn, b); err != nil {
				return err
			}
			if err := skip(conn, int(binary.BigEndian.Uint16(b[1:]))*4); err != nil {
				return err
			}
		case 3: // FramebufferUpdateRequest: incremental(1) + x,y,w,h (je 2)
			req := make([]byte, 9)
			if _, err := io.ReadFull(conn, req); err != nil {
				return err
			}
			// grobe Ratenbegrenzung (~15 fps), damit wir nicht durchdrehen
			if d := time.Since(lastFrame); d < 66*time.Millisecond {
				time.Sleep(66*time.Millisecond - d)
			}
			lastFrame = time.Now()
			np, err := sendFrame(conn, src, prev, w, h, req[0] != 0)
			if err != nil {
				return err
			}
			prev = np
		case 4: // KeyEvent: down-flag(1) + padding(2) + keysym(4)
			b := make([]byte, 7)
			if _, err := io.ReadFull(conn, b); err != nil {
				return err
			}
			if in, ok := src.(inputSink); ok {
				in.Key(b[0] != 0, binary.BigEndian.Uint32(b[3:]))
			}
		case 5: // PointerEvent: button-mask(1) + x(2) + y(2)
			b := make([]byte, 5)
			if _, err := io.ReadFull(conn, b); err != nil {
				return err
			}
			if in, ok := src.(inputSink); ok {
				in.Pointer(int(b[0]), int(binary.BigEndian.Uint16(b[1:])), int(binary.BigEndian.Uint16(b[3:])))
			}
		case 6: // ClientCutText (1+3 padding... eig. 3 padding + 4 len + len)
			b := make([]byte, 7)
			if _, err := io.ReadFull(conn, b); err != nil {
				return err
			}
			if err := skip(conn, int(binary.BigEndian.Uint32(b[3:]))); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unbekannte RFB-Nachricht %d", hdr[0])
		}
	}
}

// sendFrame überträgt bei incremental nur den geänderten Bereich (Dirty-Rectangle),
// sonst das Vollbild – jeweils als Raw-Rechteck. Liefert das aktualisierte
// „prev"-Vollbild für den nächsten Diff zurück.
func sendFrame(conn io.Writer, src screenSource, prev []byte, w, h int, incremental bool) ([]byte, error) {
	cur, err := src.Capture()
	if err != nil || len(cur) < w*h*4 {
		_, werr := conn.Write([]byte{0, 0, 0, 0}) // leeres Update
		return prev, werr
	}
	cur = cur[:w*h*4]

	x0, y0, x1, y1 := 0, 0, w-1, h-1
	if incremental && prev != nil {
		x0, y0, x1, y1 = diffRect(prev, cur, w, h)
		if x1 < x0 { // keine Änderung
			_, werr := conn.Write([]byte{0, 0, 0, 0})
			return prev, werr
		}
	}
	rw, rh := x1-x0+1, y1-y0+1

	msg := make([]byte, 0, 16)
	msg = append(msg, 0, 0) // FramebufferUpdate, padding
	msg = be16(msg, 1)      // 1 Rechteck
	msg = be16(msg, x0)
	msg = be16(msg, y0)
	msg = be16(msg, rw)
	msg = be16(msg, rh)
	msg = be32(msg, 0) // Encoding 0 = Raw
	if _, err := conn.Write(msg); err != nil {
		return prev, err
	}
	// Zeilen des Teilrechtecks senden; Chunks < WS-Nachrichtenlimit.
	stride := w * 4
	buf := make([]byte, 0, 256*1024)
	flush := func() error {
		if len(buf) == 0 {
			return nil
		}
		_, err := conn.Write(buf)
		buf = buf[:0]
		return err
	}
	for y := y0; y <= y1; y++ {
		row := cur[y*stride+x0*4 : y*stride+(x1+1)*4]
		if len(buf)+len(row) > 256*1024 {
			if err := flush(); err != nil {
				return prev, err
			}
		}
		buf = append(buf, row...)
	}
	if err := flush(); err != nil {
		return prev, err
	}

	// prev auf das aktuelle Vollbild aktualisieren.
	if len(prev) != len(cur) {
		prev = make([]byte, len(cur))
	}
	copy(prev, cur)
	return prev, nil
}

// diffRect liefert die Bounding-Box (x0,y0)-(x1,y1) der geänderten Pixel zwischen
// prev und cur. Bei keiner Änderung ist x1 < x0.
func diffRect(prev, cur []byte, w, h int) (int, int, int, int) {
	stride := w * 4
	x0, y0, x1, y1 := w, h, -1, -1
	for y := 0; y < h; y++ {
		rs := y * stride
		if bytes.Equal(prev[rs:rs+stride], cur[rs:rs+stride]) {
			continue
		}
		if y < y0 {
			y0 = y
		}
		y1 = y
		// linke Kante
		lx := 0
		for lx < w && prev[rs+lx*4] == cur[rs+lx*4] && prev[rs+lx*4+1] == cur[rs+lx*4+1] &&
			prev[rs+lx*4+2] == cur[rs+lx*4+2] {
			lx++
		}
		if lx < x0 {
			x0 = lx
		}
		// rechte Kante
		rx := w - 1
		for rx > lx && prev[rs+rx*4] == cur[rs+rx*4] && prev[rs+rx*4+1] == cur[rs+rx*4+1] &&
			prev[rs+rx*4+2] == cur[rs+rx*4+2] {
			rx--
		}
		if rx > x1 {
			x1 = rx
		}
	}
	return x0, y0, x1, y1
}

func be16(b []byte, v int) []byte { return append(b, byte(v>>8), byte(v)) }
func be32(b []byte, v int) []byte { return append(b, byte(v>>24), byte(v>>16), byte(v>>8), byte(v)) }

// skip liest n Bytes und verwirft sie.
func skip(r io.Reader, n int) error {
	if n <= 0 {
		return nil
	}
	_, err := io.CopyN(io.Discard, r, int64(n))
	return err
}
