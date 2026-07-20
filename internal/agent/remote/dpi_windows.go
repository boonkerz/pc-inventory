//go:build windows

package remote

import "syscall"

// DPI-Awareness für den Aufnahme-Helfer: DXGI erfasst physische Pixel, und die
// Maus-Umrechnung (pointerEvent) normiert über GetSystemMetrics(SM_*VIRTUALSCREEN).
// Ist der Prozess NICHT DPI-aware, virtualisiert Windows bei >100% Skalierung diese
// Metriken (z.B. ÷1,5 bei 150%) → Framebuffer- und Eingabe-Koordinaten driften
// auseinander. Als Per-Monitor-V2 (bzw. mind. System-DPI-aware) liefern die Metriken
// physische Pixel und passen wieder zur Aufnahme.

var (
	modUser32dpi                  = syscall.NewLazyDLL("user32.dll")
	procSetProcessDpiAwarenessCtx = modUser32dpi.NewProc("SetProcessDpiAwarenessContext")
	procSetProcessDPIAwareLegacy  = modUser32dpi.NewProc("SetProcessDPIAware")
	dpiAwarenessPerMonitorAwareV2 = ^uintptr(3) // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = (HANDLE)-4
)

// setProcessDPIAware macht den aktuellen Prozess DPI-aware. Muss früh (vor der ersten
// Bildschirm-/Metrik-Nutzung) aufgerufen werden. Best effort.
func setProcessDPIAware() {
	if procSetProcessDpiAwarenessCtx.Find() == nil {
		if r, _, _ := procSetProcessDpiAwarenessCtx.Call(dpiAwarenessPerMonitorAwareV2); r != 0 {
			return
		}
	}
	// Fallback für ältere Windows: System-DPI-aware.
	procSetProcessDPIAwareLegacy.Call()
}
