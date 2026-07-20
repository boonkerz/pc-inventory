//go:build windows

package remote

import (
	"fmt"
	"syscall"
	"unsafe"

	"log/slog"
)

// Adaptive Auflösung unter Windows: die physische Bildschirmauflösung des
// Primärdisplays wird für die Dauer der Sitzung auf den (zum Viewer-Fenster)
// passenden Modus umgestellt und beim Ende wieder auf den Registry-Standard
// zurückgesetzt. Reine Win32-Syscalls, keine Fremd-Tools.

var (
	modUser32res                 = syscall.NewLazyDLL("user32.dll")
	procEnumDisplaySettingsW     = modUser32res.NewProc("EnumDisplaySettingsW")
	procChangeDisplaySettingsExW = modUser32res.NewProc("ChangeDisplaySettingsExW")
)

const (
	enumCurrentSettings  = 0xFFFFFFFF
	dmPelsWidthFlag      = 0x00080000
	dmPelsHeightFlag     = 0x00100000
	dispChangeSuccessful = 0
)

// devmodeWres spiegelt DEVMODEW (Display-Variante). Reihenfolge/Größe müssen exakt
// zur Win32-Struktur passen.
type devmodeWres struct {
	dmDeviceName         [32]uint16
	dmSpecVersion        uint16
	dmDriverVersion      uint16
	dmSize               uint16
	dmDriverExtra        uint16
	dmFields             uint32
	dmPositionX          int32
	dmPositionY          int32
	dmDisplayOrientation uint32
	dmDisplayFixedOutput uint32
	dmColor              int16
	dmDuplex             int16
	dmYResolution        int16
	dmTTOption           int16
	dmCollate            int16
	dmFormName           [32]uint16
	dmLogPixels          uint16
	dmBitsPerPel         uint32
	dmPelsWidth          uint32
	dmPelsHeight         uint32
	dmDisplayFlags       uint32
	dmDisplayFrequency   uint32
	dmICMMethod          uint32
	dmICMIntent          uint32
	dmMediaType          uint32
	dmDitherType         uint32
	dmReserved1          uint32
	dmReserved2          uint32
	dmPanningWidth       uint32
	dmPanningHeight      uint32
}

func enumDisplaySettingsRes(mode uint32, dm *devmodeWres) bool {
	dm.dmSize = uint16(unsafe.Sizeof(*dm))
	r, _, _ := procEnumDisplaySettingsW.Call(0, uintptr(mode), uintptr(unsafe.Pointer(dm)))
	return r != 0
}

type winResController struct {
	log     *slog.Logger
	changed bool // haben wir die Auflösung umgestellt (dann Restore nötig)?
	curW    int
	curH    int
}

func newResController(log *slog.Logger) resController { return &winResController{log: log} }

// Set stellt den zum gewünschten (Fenster-)Format am besten passenden vom Treiber
// angebotenen Modus ein. w<=0||h<=0 stellt die ursprüngliche Auflösung wieder her.
func (c *winResController) Set(w, h int) {
	if w <= 0 || h <= 0 {
		c.Restore()
		return
	}
	tw, th, ok := closestMode(w, h)
	if !ok {
		return
	}
	if c.changed && tw == c.curW && th == c.curH {
		return // schon in dieser Auflösung
	}
	// Aktuelle Einstellungen als Basis lesen (setzt dmSize u. a. Felder korrekt).
	var dm devmodeWres
	if !enumDisplaySettingsRes(enumCurrentSettings, &dm) {
		return
	}
	if !c.changed && int(dm.dmPelsWidth) == tw && int(dm.dmPelsHeight) == th {
		return // native Auflösung passt bereits – nichts umstellen
	}
	dm.dmPelsWidth = uint32(tw)
	dm.dmPelsHeight = uint32(th)
	dm.dmFields = dmPelsWidthFlag | dmPelsHeightFlag
	// dwFlags=0 → dynamische (nicht-persistente) Umstellung; Restore via NULL-DEVMODE.
	r, _, _ := procChangeDisplaySettingsExW.Call(0, uintptr(unsafe.Pointer(&dm)), 0, 0, 0)
	if int32(r) == dispChangeSuccessful {
		c.changed = true
		c.curW, c.curH = tw, th
		c.log.Info("adaptive auflösung gesetzt", "size", fmt.Sprintf("%dx%d", tw, th))
		return
	}
	c.log.Warn("adaptive auflösung abgelehnt", "size", fmt.Sprintf("%dx%d", tw, th), "code", int32(r))
}

// Restore setzt die vor der Sitzung aktive (Registry-Standard-)Auflösung zurück.
func (c *winResController) Restore() {
	if !c.changed {
		return
	}
	// NULL-DEVMODE → Rückkehr zum in der Registry gespeicherten Standardmodus.
	procChangeDisplaySettingsExW.Call(0, 0, 0, 0, 0)
	c.changed = false
	c.curW, c.curH = 0, 0
	c.log.Info("auflösung zurückgesetzt")
}

// closestMode enumeriert die vom Treiber angebotenen Modi und wählt den best-
// passenden (Auswahllogik in pickResolution, plattformneutral).
func closestMode(tw, th int) (int, int, bool) {
	seen := map[[2]int]bool{}
	var modes [][2]int
	var dm devmodeWres
	for i := uint32(0); enumDisplaySettingsRes(i, &dm); i++ {
		mw, mh := int(dm.dmPelsWidth), int(dm.dmPelsHeight)
		if dm.dmBitsPerPel < 16 {
			continue
		}
		k := [2]int{mw, mh}
		if seen[k] {
			continue
		}
		seen[k] = true
		modes = append(modes, k)
	}
	return pickResolution(modes, tw, th)
}
