package main

import (
	"sync"

	"github.com/ebitengine/purego"
)

// Das purego-sdl3-Binding registriert SDL_GetClipboardText, aber SDL_SetClipboardText
// ist dort auskommentiert. Wir holen das bereits von SDL geladene Bibliotheks-Handle
// und registrieren die Funktion selbst (cgo-frei, keine zusätzliche Abhängigkeit).

var (
	clipOnce        sync.Once
	sdlSetClipboard func(text string) bool
)

func initClipboard() {
	h, err := sdlLibHandle()
	if err != nil {
		return
	}
	defer func() { _ = recover() }() // fehlt das Symbol, bleibt Set-Clipboard einfach ohne Funktion
	purego.RegisterLibFunc(&sdlSetClipboard, h, "SDL_SetClipboardText")
}

// setClipboardText legt Text in die lokale Zwischenablage (no-op, falls die
// Funktion nicht registriert werden konnte).
func setClipboardText(s string) {
	clipOnce.Do(initClipboard)
	if sdlSetClipboard != nil {
		sdlSetClipboard(s)
	}
}
