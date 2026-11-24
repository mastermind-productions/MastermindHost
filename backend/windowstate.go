package backend

import (
	"context"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

type WindowRect struct {
	X, Y, W, H int
	Max        bool
}

var (
	modUser32                    = windows.NewLazySystemDLL("user32.dll")
	procEnumWindows              = modUser32.NewProc("EnumWindows")
	procGetWindowThreadProcessID = modUser32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible          = modUser32.NewProc("IsWindowVisible")
	procGetWindowRect            = modUser32.NewProc("GetWindowRect")
	procSetWindowPos             = modUser32.NewProc("SetWindowPos")
	procIsIconic                 = modUser32.NewProc("IsIconic")
	procIsZoomed                 = modUser32.NewProc("IsZoomed")
	procShowWindow               = modUser32.NewProc("ShowWindow")
	procGetSystemMetrics         = modUser32.NewProc("GetSystemMetrics")
)

type winRect struct{ L, T, R, B int32 }

func mainWindowOf(pid uint32) uintptr {
	var best uintptr
	var bestArea int32
	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		var p uint32
		procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&p)))
		if p != pid {
			return 1
		}
		if v, _, _ := procIsWindowVisible.Call(hwnd); v == 0 {
			return 1
		}
		var r winRect
		procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
		if a := (r.R - r.L) * (r.B - r.T); a > bestArea {
			best, bestArea = hwnd, a
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return best
}

func readWindow(hwnd uintptr) (WindowRect, bool) {
	if v, _, _ := procIsIconic.Call(hwnd); v != 0 {
		return WindowRect{}, false
	}
	var r winRect
	if v, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r))); v == 0 {
		return WindowRect{}, false
	}
	z, _, _ := procIsZoomed.Call(hwnd)
	return WindowRect{X: int(r.L), Y: int(r.T), W: int(r.R - r.L), H: int(r.B - r.T), Max: z != 0}, true
}

func rectOnScreen(r WindowRect) bool {
	vx, _, _ := procGetSystemMetrics.Call(76)
	vy, _, _ := procGetSystemMetrics.Call(77)
	vw, _, _ := procGetSystemMetrics.Call(78)
	vh, _, _ := procGetSystemMetrics.Call(79)
	return r.W >= 200 && r.H >= 120 &&
		r.X+r.W > int(int32(vx))+80 && r.X < int(int32(vx))+int(int32(vw))-80 &&
		r.Y+60 > int(int32(vy)) && r.Y < int(int32(vy))+int(int32(vh))-60
}

func placeWindow(hwnd uintptr, r WindowRect) {
	procSetWindowPos.Call(hwnd, 0, uintptr(r.X), uintptr(r.Y), uintptr(r.W), uintptr(r.H), 0x0004|0x0010)
	if r.Max {
		procShowWindow.Call(hwnd, 3)
	}
}

func (h *Host) trackServerWindow(sv Server, pid uint32) {
	s := h.state(sv.ID)
	hwnd := mainWindowOf(pid)
	if hwnd == 0 {
		return
	}
	h.mu.Lock()
	restored := s.winPid == pid
	h.mu.Unlock()
	if !restored {
		if sv.WindowRect != nil && rectOnScreen(*sv.WindowRect) {
			placeWindow(hwnd, *sv.WindowRect)
		}
		h.mu.Lock()
		s.winPid = pid
		h.mu.Unlock()
		return
	}
	cur, ok := readWindow(hwnd)
	if !ok || !rectOnScreen(cur) {
		return
	}
	if sv.WindowRect != nil && *sv.WindowRect == cur {
		return
	}
	h.store.Update(sv.ID, func(x *Server) error { c := cur; x.WindowRect = &c; return nil })
}

func (h *Host) mainWindowSave(ctx context.Context) {
	if ctx == nil || runtime.WindowIsMinimised(ctx) {
		return
	}
	w, hh := runtime.WindowGetSize(ctx)
	x, y := runtime.WindowGetPosition(ctx)
	cur := WindowRect{X: x, Y: y, W: w, H: hh, Max: runtime.WindowIsMaximised(ctx)}
	st := h.store.Settings()
	if st.Window != nil && *st.Window == cur {
		return
	}
	if cur.Max && st.Window != nil {
		cur.X, cur.Y, cur.W, cur.H = st.Window.X, st.Window.Y, st.Window.W, st.Window.H
	}
	if !cur.Max && !rectOnScreen(cur) {
		return
	}
	h.store.SetWindow(cur)
}

func (h *Host) mainWindowRestore(ctx context.Context) {
	st := h.store.Settings()
	if st.Window == nil {
		return
	}
	r := *st.Window
	if !rectOnScreen(r) {
		return
	}
	runtime.WindowSetSize(ctx, r.W, r.H)
	runtime.WindowSetPosition(ctx, r.X, r.Y)
	if r.Max {
		runtime.WindowMaximise(ctx)
	}
}

func (h *Host) windowSaveLoop() {
	for {
		time.Sleep(3 * time.Second)
		h.mainWindowSave(h.ctx)
	}
}

func (h *Host) BeforeClose(ctx context.Context) bool {
	h.mainWindowSave(ctx)
	return false
}
