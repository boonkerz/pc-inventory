//go:build windows

package remote

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procOpenClipboard    = modUser32.NewProc("OpenClipboard")
	procCloseClipboard   = modUser32.NewProc("CloseClipboard")
	procEmptyClipboard   = modUser32.NewProc("EmptyClipboard")
	procGetClipboardData = modUser32.NewProc("GetClipboardData")
	procSetClipboardData = modUser32.NewProc("SetClipboardData")
	procGlobalAlloc      = modKernel32.NewProc("GlobalAlloc")
	procGlobalFree       = modKernel32.NewProc("GlobalFree")
	procGlobalLock       = modKernel32.NewProc("GlobalLock")
	procGlobalUnlock     = modKernel32.NewProc("GlobalUnlock")
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

func getClipboardText() string {
	if r, _, _ := procOpenClipboard.Call(0); r == 0 {
		return ""
	}
	defer procCloseClipboard.Call()
	h, _, _ := procGetClipboardData.Call(cfUnicodeText)
	if h == 0 {
		return ""
	}
	p, _, _ := procGlobalLock.Call(h)
	if p == 0 {
		return ""
	}
	defer procGlobalUnlock.Call(h)
	return windows.UTF16PtrToString((*uint16)(unsafe.Pointer(p)))
}

func setClipboardText(s string) {
	u16, err := windows.UTF16FromString(s)
	if err != nil {
		return
	}
	h, _, _ := procGlobalAlloc.Call(gmemMoveable, uintptr(len(u16)*2))
	if h == 0 {
		return
	}
	p, _, _ := procGlobalLock.Call(h)
	if p == 0 {
		procGlobalFree.Call(h)
		return
	}
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(p)), len(u16)), u16)
	procGlobalUnlock.Call(h)

	if r, _, _ := procOpenClipboard.Call(0); r == 0 {
		procGlobalFree.Call(h)
		return
	}
	defer procCloseClipboard.Call()
	procEmptyClipboard.Call()
	if r, _, _ := procSetClipboardData.Call(cfUnicodeText, h); r == 0 {
		procGlobalFree.Call(h) // fehlgeschlagen → Speicher gehört noch uns
	}
}

// winClipboard verfolgt den zuletzt gesehenen/gesetzten Inhalt, um Echos und
// unnötige Übertragungen zu vermeiden. Wird in gdiSource/dxgiSource eingebettet.
type winClipboard struct{ last string }

func (c *winClipboard) GetClipboard() (string, bool) {
	t := getClipboardText()
	if t == c.last {
		return "", false
	}
	c.last = t
	return t, true
}

func (c *winClipboard) SetClipboard(text string) {
	setClipboardText(text)
	c.last = text
}
