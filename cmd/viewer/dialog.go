//go:build linux && cgo

package main

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

// connectDialog zeigt ein kleines Fenster zum Verbinden ohne Kommandozeile: der
// Startcode (aus dem Browser-Button „Kopieren") wird per Strg+V eingefügt – die
// Zwischenablage wird beim Öffnen automatisch geprüft – und mit Enter/Klick auf den
// grünen Button verbunden. Bewusst font-frei: alle Texte stehen in der Titelleiste
// (vom Fenstermanager gerendert), der Body gibt nur Status-Feedback per Farbe.
// Rückgabe (nil, nil) = vom Nutzer abgebrochen.
func connectDialog() (*launchConfig, error) {
	const w, h = 520, 230
	win, err := sdl.CreateWindow("PC-Inventory Fernsteuerung",
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, w, h, sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}
	defer win.Destroy()
	ren, err := sdl.CreateRenderer(win, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, err
	}
	defer ren.Destroy()

	var cfg *launchConfig
	buf := ""
	setTitle := func() {
		switch {
		case cfg != nil:
			win.SetTitle("Bereit: " + cfg.Device + " — Enter/Klick zum Verbinden   (Esc = Abbruch)")
		case buf != "":
			win.SetTitle("Ungültiger Startcode — Strg+V zum Einfügen   (Esc = Abbruch)")
		default:
			win.SetTitle("Startcode einfügen: Strg+V, dann Enter/Klick zum Verbinden   (Esc = Abbruch)")
		}
	}
	tryDecode := func(s string) {
		buf = strings.TrimSpace(s)
		if buf == "" {
			cfg = nil
		} else if c, derr := decodeLaunchCode(buf); derr == nil {
			cfg = c
		} else {
			cfg = nil
		}
		setTitle()
	}
	if clip, cerr := sdl.GetClipboardText(); cerr == nil {
		tryDecode(clip) // Browser-„Kopieren" direkt davor → sofort „Bereit"
	} else {
		setTitle()
	}

	btn := sdl.Rect{X: (w - 220) / 2, Y: 150, W: 220, H: 56}
	for {
		for ev := sdl.PollEvent(); ev != nil; ev = sdl.PollEvent() {
			switch e := ev.(type) {
			case *sdl.QuitEvent:
				return nil, nil
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_CLOSE {
					return nil, nil
				}
			case *sdl.KeyboardEvent:
				if e.Type != sdl.KEYDOWN {
					break
				}
				switch {
				case e.Keysym.Sym == sdl.K_ESCAPE:
					return nil, nil
				case e.Keysym.Sym == sdl.K_RETURN || e.Keysym.Sym == sdl.K_KP_ENTER:
					if cfg != nil {
						return cfg, nil
					}
				case e.Keysym.Sym == sdl.K_v && e.Keysym.Mod&uint16(sdl.KMOD_CTRL) != 0:
					if clip, cerr := sdl.GetClipboardText(); cerr == nil {
						tryDecode(clip)
					}
				}
			case *sdl.MouseButtonEvent:
				if e.Type == sdl.MOUSEBUTTONDOWN && e.Button == sdl.BUTTON_LEFT && cfg != nil &&
					int32(e.X) >= btn.X && int32(e.X) < btn.X+btn.W &&
					int32(e.Y) >= btn.Y && int32(e.Y) < btn.Y+btn.H {
					return cfg, nil
				}
			}
		}

		ren.SetDrawColor(0x14, 0x18, 0x20, 0xff)
		ren.Clear()
		// „Eingabefeld" – füllt sich grün, wenn ein gültiger Code geladen ist.
		field := sdl.Rect{X: 40, Y: 64, W: w - 80, H: 40}
		ren.SetDrawColor(0x0b, 0x0e, 0x14, 0xff)
		ren.FillRect(&field)
		ren.SetDrawColor(0x33, 0x3a, 0x46, 0xff)
		ren.DrawRect(&field)
		if len(buf) > 0 {
			barW := int32(len(buf))
			if barW > field.W-8 {
				barW = field.W - 8
			}
			if cfg != nil {
				ren.SetDrawColor(0x2e, 0x7d, 0x32, 0xff)
			} else {
				ren.SetDrawColor(0x8a, 0x3b, 0x3b, 0xff)
			}
			ren.FillRect(&sdl.Rect{X: field.X + 4, Y: field.Y + 12, W: barW, H: 16})
		}
		// Connect-Button (grün = bereit) mit Play-Dreieck.
		if cfg != nil {
			ren.SetDrawColor(0x2e, 0x7d, 0x32, 0xff)
		} else {
			ren.SetDrawColor(0x30, 0x36, 0x40, 0xff)
		}
		ren.FillRect(&btn)
		drawPlay(ren, btn, cfg != nil)
		ren.Present()
		sdl.Delay(16)
	}
}

// drawPlay zeichnet ein nach rechts zeigendes Play-Dreieck mittig in r.
func drawPlay(ren *sdl.Renderer, r sdl.Rect, active bool) {
	if active {
		ren.SetDrawColor(0xff, 0xff, 0xff, 0xff)
	} else {
		ren.SetDrawColor(0x66, 0x6c, 0x78, 0xff)
	}
	const size = int32(16)
	cx, cy := r.X+r.W/2-size/2, r.Y+r.H/2
	for dy := -size; dy <= size; dy++ {
		wln := size - abs32(dy)
		ren.DrawLine(cx, cy+dy, cx+wln, cy+dy)
	}
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}
