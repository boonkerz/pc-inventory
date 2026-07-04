package remote

import (
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
		case 3: // FramebufferUpdateRequest
			if err := skip(conn, 9); err != nil {
				return err
			}
			// grobe Ratenbegrenzung (~15 fps), damit wir nicht durchdrehen
			if d := time.Since(lastFrame); d < 66*time.Millisecond {
				time.Sleep(66*time.Millisecond - d)
			}
			lastFrame = time.Now()
			if err := sendFrame(conn, src, w, h); err != nil {
				return err
			}
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

// sendFrame überträgt den kompletten Bildschirm als ein Raw-Rechteck.
func sendFrame(conn io.Writer, src screenSource, w, h int) error {
	px, err := src.Capture()
	if err != nil || len(px) < w*h*4 {
		// leeres Update senden, damit der Client nicht hängt
		empty := []byte{0, 0, 0, 0} // msg-type 0, padding, num-rects=0
		_, werr := conn.Write(empty)
		return werr
	}
	msg := make([]byte, 0, 4+12)
	msg = append(msg, 0, 0) // FramebufferUpdate, padding
	msg = be16(msg, 1)      // 1 Rechteck
	msg = be16(msg, 0)      // x
	msg = be16(msg, 0)      // y
	msg = be16(msg, w)
	msg = be16(msg, h)
	msg = be32(msg, 0) // Encoding 0 = Raw
	if _, err := conn.Write(msg); err != nil {
		return err
	}
	// In Chunks schreiben: der WS-Relay begrenzt die Nachrichtengröße (< 1 MB).
	data := px[:w*h*4]
	const chunk = 200 * 1024
	for len(data) > 0 {
		n := chunk
		if n > len(data) {
			n = len(data)
		}
		if _, err := conn.Write(data[:n]); err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
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
