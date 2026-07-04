package remote

import "time"

// syntheticSource erzeugt ein bewegtes Testbild. Dient zum Verifizieren des
// RFB-Servers/Tunnels und als Übergangs-Quelle auf Plattformen ohne echte Aufnahme.
type syntheticSource struct {
	w, h int
	buf  []byte
}

func newSyntheticSource() *syntheticSource {
	s := &syntheticSource{w: 1280, h: 720}
	s.buf = make([]byte, s.w*s.h*4)
	return s
}

func (s *syntheticSource) Bounds() (int, int) { return s.w, s.h }
func (s *syntheticSource) Close() error       { return nil }

func (s *syntheticSource) Capture() ([]byte, error) {
	t := int(time.Now().UnixMilli() / 40)
	for y := 0; y < s.h; y++ {
		row := y * s.w * 4
		for x := 0; x < s.w; x++ {
			i := row + x*4
			s.buf[i+0] = byte(x + t) // B
			s.buf[i+1] = byte(y + t) // G
			s.buf[i+2] = byte(x ^ y) // R
			s.buf[i+3] = 0
		}
	}
	return s.buf, nil
}
