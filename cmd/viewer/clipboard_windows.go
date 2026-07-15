//go:build windows

package main

import "golang.org/x/sys/windows"

// sdlLibHandle liefert das Handle der bereits geladenen SDL3.dll (LoadLibrary ist
// referenzgezählt und liefert dasselbe Modul).
func sdlLibHandle() (uintptr, error) {
	h, err := windows.LoadLibrary("SDL3.dll")
	if err != nil {
		return 0, err
	}
	return uintptr(h), nil
}
