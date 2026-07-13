//go:build windows

package remote

import (
	"log/slog"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modSas      = windows.NewLazySystemDLL("sas.dll")
	procSendSAS = modSas.NewProc("SendSAS")

	modWtsapi           = windows.NewLazySystemDLL("wtsapi32.dll")
	procWTSSendMessageW = modWtsapi.NewProc("WTSSendMessageW")

	procWTSGetActiveConsoleSessionId = modKernel32.NewProc("WTSGetActiveConsoleSessionId")
	procBlockInput                   = modUser32.NewProc("BlockInput")
)

// agentSendSAS löst eine echte Strg+Alt+Entf-Sequenz aus. Setzt zuvor best-effort die
// Richtlinie SoftwareSASGeneration=1, damit ein SYSTEM-Dienst die SAS erzeugen darf.
func agentSendSAS(log *slog.Logger) {
	_ = exec.Command("reg", "add",
		`HKLM\SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System`,
		"/v", "SoftwareSASGeneration", "/t", "REG_DWORD", "/d", "1", "/f").Run()
	if r, _, err := procSendSAS.Call(0); r == 0 { // AsUser = FALSE
		log.Warn("SendSAS fehlgeschlagen (Richtlinie/Berechtigung?)", "err", err)
	}
}

// agentShowMessage zeigt eine Meldung in der aktiven Konsolen-Sitzung (vom Dienst aus).
func agentShowMessage(log *slog.Logger, text string) {
	if text == "" {
		return
	}
	sess, _, _ := procWTSGetActiveConsoleSessionId.Call()
	title := utf16Bytes("Fernwartung")
	msg := utf16Bytes(text)
	var resp uint32
	procWTSSendMessageW.Call(
		0, sess, // WTS_CURRENT_SERVER_HANDLE, aktive Sitzung
		uintptr(unsafe.Pointer(&title[0])), uintptr(len(title)-2), // Länge in Bytes ohne NUL
		uintptr(unsafe.Pointer(&msg[0])), uintptr(len(msg)-2),
		0x40, // MB_OK | MB_ICONINFORMATION
		0,    // kein Timeout
		uintptr(unsafe.Pointer(&resp)), 0)
}

func utf16Bytes(s string) []byte {
	u, _ := syscall.UTF16FromString(s) // inkl. NUL
	b := make([]byte, len(u)*2)
	for i, v := range u {
		b[i*2] = byte(v)
		b[i*2+1] = byte(v >> 8)
	}
	return b
}

// blockInput sperrt/entsperrt lokale Maus/Tastatur. Wirkt nur, wenn im interaktiven
// Sitzungskontext ausgeführt (interaktiver Agent bzw. __capture-Helfer).
func blockInput(on bool) {
	v := uintptr(0)
	if on {
		v = 1
	}
	procBlockInput.Call(v)
}

// gdiSource (interaktiver Agent) sperrt direkt.
func (s *gdiSource) BlockInput(on bool) { blockInput(on) }
