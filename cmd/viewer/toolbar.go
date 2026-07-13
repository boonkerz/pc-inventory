package main

import "github.com/jupiterrider/purego-sdl3/sdl"

// Overlay-Bedienleiste (oben) für den Viewer: Sondertasten, Sperren, Meldung,
// Qualität, Vollbild, Trennen. Font-basiert (font8x8), damit Buttons beschriftet
// sind. Die Leiste liegt über dem oberen Bildrand; das Remote-Bild wird darunter
// gerendert (siehe layout in main.go).

const barHeight float32 = 30

type button struct {
	id    string
	label string
	x, w  float32 // wird beim Layout gesetzt
}

// tbState hält Beschriftungen, die sich ändern (Qualität, Sperr-Status).
type toolbar struct {
	buttons []button
}

func newToolbar() *toolbar {
	return &toolbar{buttons: []button{
		{id: "sas", label: "Strg+Alt+Entf"},
		{id: "win", label: "Win"},
		{id: "alttab", label: "Alt+Tab"},
		{id: "esc", label: "Esc"},
		{id: "lock", label: "Sperren"},
		{id: "msg", label: "Meldung"},
		{id: "qual", label: "Qual: M"},
		{id: "full", label: "Vollbild"},
		{id: "disc", label: "Trennen"},
	}}
}

func (t *toolbar) setLabel(id, label string) {
	for i := range t.buttons {
		if t.buttons[i].id == id {
			t.buttons[i].label = label
		}
	}
}

const tbScale float32 = 2 // Font-Vergrößerung
const tbPad float32 = 8   // Innenabstand je Button

// layout ordnet die Buttons links nach rechts an (füllt x,w).
func (t *toolbar) layout() {
	x := float32(4)
	for i := range t.buttons {
		w := textWidth(t.buttons[i].label, tbScale) + 2*tbPad
		t.buttons[i].x = x
		t.buttons[i].w = w
		x += w + 4
	}
}

// hit liefert die Button-ID an Fensterposition (mx,my) oder "".
func (t *toolbar) hit(mx, my float32) string {
	if my > barHeight {
		return ""
	}
	for _, b := range t.buttons {
		if mx >= b.x && mx < b.x+b.w {
			return b.id
		}
	}
	return ""
}

func (t *toolbar) draw(r *sdl.Renderer, winW float32, hoverID string) {
	// Leistenhintergrund.
	sdl.SetRenderDrawColor(r, 0x1a, 0x1f, 0x28, 0xff)
	sdl.RenderFillRect(r, &sdl.FRect{X: 0, Y: 0, W: winW, H: barHeight})
	for _, b := range t.buttons {
		if b.id == hoverID {
			sdl.SetRenderDrawColor(r, 0x2e, 0x38, 0x46, 0xff)
		} else {
			sdl.SetRenderDrawColor(r, 0x26, 0x2d, 0x38, 0xff)
		}
		sdl.RenderFillRect(r, &sdl.FRect{X: b.x, Y: 3, W: b.w, H: barHeight - 6})
		ty := (barHeight - 8*tbScale) / 2
		drawText(r, b.x+tbPad, ty, tbScale, b.label, 0xe6, 0xea, 0xf0)
	}
}

// drawText zeichnet einen ASCII-String mit der 8x8-Font als Rechtecke.
func drawText(r *sdl.Renderer, x, y, scale float32, s string, cr, cg, cb uint8) {
	sdl.SetRenderDrawColor(r, cr, cg, cb, 0xff)
	rects := make([]sdl.FRect, 0, len(s)*16)
	cx := x
	for _, ch := range s {
		if ch >= 0 && ch < 128 {
			g := font8x8[ch]
			for row := 0; row < 8; row++ {
				bits := g[row]
				for col := 0; col < 8; col++ {
					if bits&(1<<uint(col)) != 0 {
						rects = append(rects, sdl.FRect{X: cx + float32(col)*scale, Y: y + float32(row)*scale, W: scale, H: scale})
					}
				}
			}
		}
		cx += 8*scale + scale
	}
	if len(rects) > 0 {
		sdl.RenderFillRects(r, rects)
	}
}

func textWidth(s string, scale float32) float32 {
	return float32(len(s)) * (8*scale + scale)
}
