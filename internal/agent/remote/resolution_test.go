package remote

import "testing"

func TestPickResolution(t *testing.T) {
	modes := [][2]int{
		{3840, 2160}, {2560, 1440}, {1920, 1080}, {1600, 1200}, {1280, 1024}, {1024, 768}, {800, 600},
	}
	cases := []struct {
		name   string
		tw, th int
		wantW  int
		wantH  int
		wantOK bool
	}{
		{"exakt 16:9", 1920, 1080, 1920, 1080, true},
		{"16:9-fenster bevorzugt 16:9-modus trotz 4:3-größennähe", 1700, 956, 1920, 1080, true},
		{"4:3-fenster bevorzugt größennächsten 4:3-modus", 1180, 900, 1024, 768, true},
		{"hidpi 1440p", 2560, 1440, 2560, 1440, true},
		{"ungültig", 0, 0, 0, 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w, h, ok := pickResolution(modes, c.tw, c.th)
			if ok != c.wantOK || (ok && (w != c.wantW || h != c.wantH)) {
				t.Fatalf("pickResolution(%d,%d) = %d,%d,%v; erwartet %d,%d,%v",
					c.tw, c.th, w, h, ok, c.wantW, c.wantH, c.wantOK)
			}
		})
	}
}

func TestPickResolutionEmpty(t *testing.T) {
	if _, _, ok := pickResolution(nil, 1920, 1080); ok {
		t.Fatal("leere Modusliste sollte kein Ergebnis liefern")
	}
	// Zu kleine Modi werden ignoriert.
	if _, _, ok := pickResolution([][2]int{{320, 240}}, 1920, 1080); ok {
		t.Fatal("nur Sub-640-Modi -> kein Ergebnis")
	}
}
