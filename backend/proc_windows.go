package backend

import (
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modPsapi                   = windows.NewLazySystemDLL("psapi.dll")
	procGetProcessMemoryInfo   = modPsapi.NewProc("GetProcessMemoryInfo")
	modKernel32                = windows.NewLazySystemDLL("kernel32.dll")
	procGlobalMemoryStatusEx   = modKernel32.NewProc("GlobalMemoryStatusEx")
	procGetSystemTimes         = modKernel32.NewProc("GetSystemTimes")
	procSetProcessAffinityMask = modKernel32.NewProc("SetProcessAffinityMask")
	procGetProcessAffinityMask = modKernel32.NewProc("GetProcessAffinityMask")
)

type processMemoryCounters struct {
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

func hideWindowAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}

func newConsoleAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: 0x00000010}
}

func processImagePath(pid uint32) string {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)
	buf := make([]uint16, windows.MAX_LONG_PATH)
	size := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(h, 0, &buf[0], &size); err != nil {
		return ""
	}
	return windows.UTF16ToString(buf[:size])
}

func findProcessByPath(exePath string) uint32 {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0
	}
	defer windows.CloseHandle(snap)
	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	want := strings.ToLower(filepath.Base(exePath))
	full := strings.ToLower(filepath.Clean(exePath))
	for err = windows.Process32First(snap, &pe); err == nil; err = windows.Process32Next(snap, &pe) {
		if strings.ToLower(windows.UTF16ToString(pe.ExeFile[:])) != want {
			continue
		}
		if strings.ToLower(filepath.Clean(processImagePath(pe.ProcessID))) == full {
			return pe.ProcessID
		}
	}
	return 0
}

func processAlive(pid uint32) bool {
	if pid == 0 {
		return false
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)
	var code uint32
	if err := windows.GetExitCodeProcess(h, &code); err != nil {
		return false
	}
	return code == 259
}

func killProcess(pid uint32) error {
	h, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, pid)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	return windows.TerminateProcess(h, 1)
}

func processTimes(pid uint32) (cpu time.Duration, start time.Time, ram uint64, ok bool) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return
	}
	defer windows.CloseHandle(h)
	var c, e, k, u windows.Filetime
	if err := windows.GetProcessTimes(h, &c, &e, &k, &u); err != nil {
		return
	}
	cpu = time.Duration((k.Nanoseconds() + u.Nanoseconds()))
	start = time.Unix(0, c.Nanoseconds())
	var pmc processMemoryCounters
	pmc.CB = uint32(unsafe.Sizeof(pmc))
	r, _, _ := procGetProcessMemoryInfo.Call(uintptr(h), uintptr(unsafe.Pointer(&pmc)), uintptr(pmc.CB))
	if r != 0 {
		ram = uint64(pmc.WorkingSetSize)
	}
	ok = true
	return
}

func hostMemory() (total, avail uint64) {
	var m memoryStatusEx
	m.Length = uint32(unsafe.Sizeof(m))
	procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&m)))
	return m.TotalPhys, m.AvailPhys
}

func systemTimes() (idle, total uint64) {
	var i, k, u windows.Filetime
	procGetSystemTimes.Call(uintptr(unsafe.Pointer(&i)), uintptr(unsafe.Pointer(&k)), uintptr(unsafe.Pointer(&u)))
	ft := func(f windows.Filetime) uint64 { return uint64(f.HighDateTime)<<32 | uint64(f.LowDateTime) }
	return ft(i), ft(k) + ft(u)
}

func diskUsage(path string) (total, free uint64) {
	p, err := windows.UTF16PtrFromString(filepath.VolumeName(path) + `\`)
	if err != nil {
		return
	}
	var avail uint64
	windows.GetDiskFreeSpaceEx(p, &avail, &total, &free)
	return
}

func openInExplorer(path string) error {
	return exec.Command("explorer.exe", path).Start()
}

func runElevated(file string) error {
	cmd := exec.Command("powershell.exe", "-NoProfile", "-Command",
		"Start-Process -FilePath cmd.exe -ArgumentList '/c \""+file+"\"' -Verb RunAs")
	cmd.SysProcAttr = hideWindowAttr()
	return cmd.Run()
}

func setProcessAffinity(pid uint32, mask uintptr) error {
	h, err := windows.OpenProcess(windows.PROCESS_SET_INFORMATION|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	r, _, e := procSetProcessAffinityMask.Call(uintptr(h), mask)
	if r == 0 {
		return e
	}
	return nil
}

func getProcessAffinity(pid uint32) (uintptr, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(h)
	var pm, sm uintptr
	r, _, e := procGetProcessAffinityMask.Call(uintptr(h), uintptr(unsafe.Pointer(&pm)), uintptr(unsafe.Pointer(&sm)))
	if r == 0 {
		return 0, e
	}
	return pm, nil
}
