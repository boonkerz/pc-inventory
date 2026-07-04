//go:build windows

package remote

import (
	"unsafe"

	"log/slog"

	"golang.org/x/sys/windows"
)

var (
	modWTS             = windows.NewLazySystemDLL("wtsapi32.dll")
	procWTSSendMessage = modWTS.NewProc("WTSSendMessageW")
)

const (
	mbYesNo         = 0x00000004
	mbIconQuestion  = 0x00000020
	mbSetForeground = 0x00010000
	mbTopmost       = 0x00040000
	idYes           = 6
	consentTimeout  = 30 // Sekunden
)

// confirmRemote fragt den an der Konsole angemeldeten Nutzer per Meldung
// (WTSSendMessage, aus dem Dienst heraus) um Erlaubnis. true nur bei „Ja";
// Zeitüberschreitung/„Nein"/kein Nutzer -> false.
func confirmRemote(log *slog.Logger) bool {
	sess := windows.WTSGetActiveConsoleSessionId()
	if sess == 0xFFFFFFFF {
		return false
	}
	title, _ := windows.UTF16FromString("Fernwartung")
	msg, _ := windows.UTF16FromString("Ein Techniker möchte sich mit diesem Computer verbinden. Zulassen?")
	var resp uint32
	r, _, _ := procWTSSendMessage.Call(
		0, // WTS_CURRENT_SERVER_HANDLE
		uintptr(sess),
		uintptr(unsafe.Pointer(&title[0])), uintptr(len(title)*2-2),
		uintptr(unsafe.Pointer(&msg[0])), uintptr(len(msg)*2-2),
		uintptr(mbYesNo|mbIconQuestion|mbSetForeground|mbTopmost),
		uintptr(consentTimeout),
		uintptr(unsafe.Pointer(&resp)),
		1, // bWait
	)
	if r == 0 {
		log.Warn("zustimmungs-dialog fehlgeschlagen")
		return false
	}
	return resp == idYes
}
