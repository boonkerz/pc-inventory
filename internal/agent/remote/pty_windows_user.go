//go:build windows

package remote

import (
	"fmt"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Dieser Pfad startet die Shell im Kontext des an der Konsole angemeldeten
// Benutzers. Dafür wird – anders als beim System-Pfad (conpty-Lib) – die
// ConPTY-Plumbing selbst aufgebaut und der Prozess mit CreateProcessAsUser unter
// dem Token der aktiven Sitzung gestartet. Voraussetzung: der Agent läuft als
// SYSTEM-Dienst (WTSQueryUserToken verlangt SeTcbPrivilege).

var (
	modKernel32                           = windows.NewLazySystemDLL("kernel32.dll")
	procCreatePseudoConsole               = modKernel32.NewProc("CreatePseudoConsole")
	procResizePseudoConsole               = modKernel32.NewProc("ResizePseudoConsole")
	procClosePseudoConsole                = modKernel32.NewProc("ClosePseudoConsole")
	procInitializeProcThreadAttributeList = modKernel32.NewProc("InitializeProcThreadAttributeList")
	procUpdateProcThreadAttribute         = modKernel32.NewProc("UpdateProcThreadAttribute")
)

const procThreadAttributePseudoConsole uintptr = 0x20016

type hpcon windows.Handle

type userPTY struct {
	hpc                          hpcon
	pi                           windows.ProcessInformation
	ptyIn, ptyOut, cmdIn, cmdOut windows.Handle
	token                        windows.Token
	envBlock                     *uint16
	closeOnce                    sync.Once
}

func packCoord(cols, rows int) uintptr {
	return uintptr((int32(rows) << 16) | (int32(cols) & 0xffff))
}

func startUserPTY(shell string) (ptySession, error) {
	sess := windows.WTSGetActiveConsoleSessionId()
	if sess == 0xFFFFFFFF {
		return nil, fmt.Errorf("keine aktive Konsolen-Sitzung (niemand angemeldet)")
	}
	var token windows.Token
	if err := windows.WTSQueryUserToken(sess, &token); err != nil {
		return nil, fmt.Errorf("token der aktiven sitzung holen (Agent muss als SYSTEM-Dienst laufen): %w", err)
	}

	var envBlock *uint16
	_ = windows.CreateEnvironmentBlock(&envBlock, token, false) // best effort

	p := &userPTY{token: token, envBlock: envBlock}
	fail := func(err error) (ptySession, error) {
		p.Close()
		return nil, err
	}

	if err := windows.CreatePipe(&p.ptyIn, &p.cmdIn, nil, 0); err != nil {
		return fail(fmt.Errorf("CreatePipe: %w", err))
	}
	if err := windows.CreatePipe(&p.cmdOut, &p.ptyOut, nil, 0); err != nil {
		return fail(fmt.Errorf("CreatePipe: %w", err))
	}

	hpc, err := createPseudoConsole(120, 30, p.ptyIn, p.ptyOut)
	if err != nil {
		return fail(err)
	}
	p.hpc = hpc

	si, err := newStartupInfoEx(hpc)
	if err != nil {
		return fail(err)
	}

	cmdLine, err := windows.UTF16PtrFromString(windowsShellCmd(shell))
	if err != nil {
		return fail(err)
	}
	flags := uint32(windows.EXTENDED_STARTUPINFO_PRESENT | windows.CREATE_UNICODE_ENVIRONMENT)
	if err := windows.CreateProcessAsUser(
		token, nil, cmdLine, nil, nil, false, flags, envBlock, nil, &si.startupInfo, &p.pi,
	); err != nil {
		return fail(fmt.Errorf("CreateProcessAsUser: %w", err))
	}
	return p, nil
}

func createPseudoConsole(cols, rows int, hIn, hOut windows.Handle) (hpcon, error) {
	if err := procCreatePseudoConsole.Find(); err != nil {
		return 0, fmt.Errorf("CreatePseudoConsole nicht verfügbar: %w", err)
	}
	var hpc hpcon
	ret, _, _ := procCreatePseudoConsole.Call(
		packCoord(cols, rows), uintptr(hIn), uintptr(hOut), 0, uintptr(unsafe.Pointer(&hpc)))
	if ret != 0 {
		return 0, fmt.Errorf("CreatePseudoConsole: status 0x%x", ret)
	}
	return hpc, nil
}

// startupInfoEx bildet STARTUPINFOEX nach. Der Slice-Header von attrList liegt
// direkt hinter startupInfo, sodass dessen Datenzeiger genau dort steht, wo Win32
// die lpAttributeList erwartet (Cb deckt StartupInfo + einen Zeiger ab).
type startupInfoEx struct {
	startupInfo windows.StartupInfo
	attrList    []byte
}

func newStartupInfoEx(hpc hpcon) (*startupInfoEx, error) {
	if err := procInitializeProcThreadAttributeList.Find(); err != nil {
		return nil, err
	}
	if err := procUpdateProcThreadAttribute.Find(); err != nil {
		return nil, err
	}
	var si startupInfoEx
	si.startupInfo.Cb = uint32(unsafe.Sizeof(windows.StartupInfo{}) + unsafe.Sizeof(uintptr(0)))
	si.startupInfo.Flags |= windows.STARTF_USESTDHANDLES

	var size uintptr
	procInitializeProcThreadAttributeList.Call(0, 1, 0, uintptr(unsafe.Pointer(&size)))
	si.attrList = make([]byte, size)
	ret, _, err := procInitializeProcThreadAttributeList.Call(
		uintptr(unsafe.Pointer(&si.attrList[0])), 1, 0, uintptr(unsafe.Pointer(&size)))
	if ret != 1 {
		return nil, fmt.Errorf("InitializeProcThreadAttributeList: %w", err)
	}
	ret, _, err = procUpdateProcThreadAttribute.Call(
		uintptr(unsafe.Pointer(&si.attrList[0])), 0,
		procThreadAttributePseudoConsole, uintptr(hpc), unsafe.Sizeof(hpc), 0, 0)
	if ret != 1 {
		return nil, fmt.Errorf("UpdateProcThreadAttribute: %w", err)
	}
	return &si, nil
}

func (p *userPTY) Read(b []byte) (int, error) {
	var n uint32
	err := windows.ReadFile(p.cmdOut, b, &n, nil)
	return int(n), err
}

func (p *userPTY) Write(b []byte) (int, error) {
	var n uint32
	err := windows.WriteFile(p.cmdIn, b, &n, nil)
	return int(n), err
}

func (p *userPTY) Resize(cols, rows int) error {
	if err := procResizePseudoConsole.Find(); err != nil {
		return err
	}
	ret, _, _ := procResizePseudoConsole.Call(uintptr(p.hpc), packCoord(cols, rows))
	if ret != 0 {
		return fmt.Errorf("ResizePseudoConsole: status 0x%x", ret)
	}
	return nil
}

func (p *userPTY) Wait() int {
	if p.pi.Process == 0 {
		return -1
	}
	windows.WaitForSingleObject(p.pi.Process, 0xFFFFFFFF) // INFINITE
	var code uint32
	_ = windows.GetExitCodeProcess(p.pi.Process, &code)
	return int(code)
}

// Close ist idempotent (Wait- und Teardown-Pfad rufen es nebenläufig).
func (p *userPTY) Close() error {
	p.closeOnce.Do(func() {
		if p.hpc != 0 {
			procClosePseudoConsole.Call(uintptr(p.hpc)) // killt den angehängten Prozess
		}
		if p.envBlock != nil {
			_ = windows.DestroyEnvironmentBlock(p.envBlock)
		}
		for _, h := range []windows.Handle{p.pi.Process, p.pi.Thread, p.ptyIn, p.ptyOut, p.cmdIn, p.cmdOut} {
			if h != 0 && h != windows.InvalidHandle {
				windows.CloseHandle(h)
			}
		}
		if p.token != 0 {
			p.token.Close()
		}
	})
	return nil
}
