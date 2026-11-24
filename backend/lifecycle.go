package backend

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func (h *Host) usedPorts(exceptID string) map[int]bool {
	used := map[int]bool{}
	for _, sv := range h.store.List() {
		if sv.ID == exceptID {
			continue
		}
		if isScript(sv) {
			continue
		}
		for _, p := range h.runtime(sv).ports() {
			used[p] = true
		}
	}
	return used
}

func (h *Host) Create(name, game string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("server name is required")
	}
	g, err := gameDef(game)
	if err != nil {
		return "", err
	}
	root := h.store.Settings().ServersRoot
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	dir := uniqueDir(root, name, "")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	used := h.usedPorts("")
	port := g.DefaultPort
	for used[port] || used[port+1] {
		port += 2
	}
	rcon := g.DefaultRcon
	for used[rcon] {
		rcon++
	}
	query := 27015
	for used[query] || query == rcon {
		query++
	}
	pass := randomPassword(12)
	cmd := defaultCmd(g, port, query, rcon, pass)
	sv := &Server{
		ID: newID(), Name: name, Game: g.ID, Dir: dir, Cmd: cmd, CreatedAt: time.Now(),
		Backup:   BackupSettings{Keep: 20, AutoRemove: true},
		ModNames: map[string]string{},
	}
	if err := h.store.Add(sv); err != nil {
		return "", err
	}
	gusPath := filepath.Join(dir, Games[game].ConfigFiles[1].RelPath)
	os.MkdirAll(filepath.Dir(gusPath), 0o755)
	os.WriteFile(gusPath, []byte(initialConfig(g, name, rcon, pass)), 0o644)
	h.state(sv.ID)
	h.logAction(sv.ID, "create", "success", "")
	h.emit(sv.ID)
	if h.latestBuild(g.AppID, "") == "" {
		go func(id string) {
			if h.fetchLatest(g.AppID) == nil {
				h.emit(id)
			}
		}(sv.ID)
	}
	return sv.ID, nil
}

func (h *Host) Import(name, dir, game string) (string, error) {
	g, err := gameDef(game)
	if err != nil {
		return "", err
	}
	if g.Kind == "script" {
		return h.importScript(name, dir, g)
	}
	if _, err := os.Stat(filepath.Join(dir, g.Exe)); err != nil {
		return "", fmt.Errorf("%s not found in %s", g.ProcessName, dir)
	}
	cmd := defaultCmd(g, g.DefaultPort, 27015, g.DefaultRcon, randomPassword(12))
	if b, err := os.ReadFile(filepath.Join(filepath.Dir(filepath.Join(dir, g.Exe)), "start_server.bat")); err == nil {
		if c := extractCmdFromBat(string(b), g.ProcessName); c != "" {
			cmd = c
		}
	}
	if strings.TrimSpace(name) == "" {
		name = filepath.Base(dir)
	}
	sv := &Server{ID: newID(), Name: name, Game: g.ID, Dir: dir, Cmd: cmd, CreatedAt: time.Now(),
		Backup: BackupSettings{Keep: 20, AutoRemove: true}, ModNames: map[string]string{}}
	if err := h.store.Add(sv); err != nil {
		return "", err
	}
	h.logAction(sv.ID, "import", "success", "")
	if g.ID == "arksa" {
		go h.ResolveModNames(sv.ID)
	}
	return sv.ID, nil
}

func extractCmdFromBat(bat, exe string) string {
	for _, line := range strings.Split(bat, "\n") {
		i := strings.Index(strings.ToLower(line), strings.ToLower(exe))
		if i < 0 {
			continue
		}
		rest := strings.TrimSpace(line[i+len(exe):])
		return strings.TrimSpace(strings.TrimPrefix(rest, `"`))
	}
	return ""
}

func (h *Host) isRunning(id string) bool {
	s := h.state(id)
	script := false
	if sv, err := h.store.Get(id); err == nil {
		script = isScript(sv)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if script {
		return s.compUp > 0 || s.launching
	}
	return s.pid != 0
}

func (h *Host) SetName(id, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("server name is required")
	}
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	if isScript(sv) {
		err = h.store.Update(id, func(s *Server) error { s.Name = name; return nil })
		h.logAction(id, "rename", "success", "")
		h.emit(id)
		return err
	}
	newDir := uniqueDir(filepath.Dir(sv.Dir), name, sv.Dir)
	if !strings.EqualFold(newDir, sv.Dir) || filepath.Base(newDir) != filepath.Base(sv.Dir) {
		if h.isRunning(id) || h.state(id).busy {
			return errors.New("stop the server before renaming - its folder is in use")
		}
		if _, err := os.Stat(sv.Dir); err == nil {
			if err := os.Rename(sv.Dir, newDir); err != nil {
				return fmt.Errorf("could not rename folder: %w", err)
			}
		}
		if sv.Backup.Dir == "" {
			root := h.store.Settings().BackupsRoot
			oldB, newB := filepath.Join(root, filepath.Base(sv.Dir)), filepath.Join(root, filepath.Base(newDir))
			if _, err := os.Stat(newB); os.IsNotExist(err) {
				os.Rename(oldB, newB)
			}
		}
	}
	err = h.store.Update(id, func(s *Server) error { s.Name = name; s.Dir = newDir; return nil })
	h.logAction(id, "rename", "success", "")
	h.emit(id)
	return err
}

func (h *Host) SetForceRunning(id string, on bool) error {
	return h.store.Update(id, func(s *Server) error { s.ForceRunning = on; return nil })
}

func (h *Host) Remove(id string) error {
	if sv, err := h.store.Get(id); err == nil && isScript(sv) && h.scriptUp(sv) > 0 {
		return errors.New("stop the server first")
	}
	if h.isRunning(id) {
		return errors.New("stop the server first")
	}
	return h.store.Remove(id)
}

func (h *Host) Delete(id string) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	if isScript(sv) {
		return errors.New("script servers live in their own folders - use Remove from list instead")
	}
	if h.isRunning(id) || h.state(id).busy {
		return errors.New("stop the server first")
	}
	if err := os.RemoveAll(sv.Dir); err != nil {
		return err
	}
	return h.store.Remove(id)
}

func (h *Host) writeStartBat(sv Server) {
	exe := h.exePath(sv)
	bat := fmt.Sprintf("@echo off\r\ncd /d \"%s\"\r\nstart \"\" \"%s\" %s\r\n", filepath.Dir(exe), Games[sv.Game].ProcessName, sv.Cmd)
	os.WriteFile(filepath.Join(filepath.Dir(exe), "start_server.bat"), []byte(bat), 0o644)
}

func (h *Host) Start(id string) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	if isScript(sv) {
		return h.startScript(sv)
	}
	if !h.installed(sv) {
		return errors.New("server is not installed")
	}
	if h.isRunning(id) {
		return errors.New("server is already running")
	}
	h.writeStartBat(sv)
	exe := h.exePath(sv)
	cmd := exec.Command(exe)
	cmd.Dir = filepath.Dir(exe)
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: `"` + exe + `" ` + sv.Cmd}
	if err := cmd.Start(); err != nil {
		h.logAction(id, "start", "error", err.Error())
		return err
	}
	pid := uint32(cmd.Process.Pid)
	s := h.state(id)
	h.mu.Lock()
	s.pid, s.startTime, s.userStop, s.lastCPU, s.lastSample, s.players = pid, time.Now(), false, 0, time.Time{}, 0
	h.mu.Unlock()
	h.applyAffinity(id)
	h.setStatus(id, StatusStarting)
	h.logAction(id, "start", "success", "")
	go func() {
		cmd.Wait()
		h.handleExit(id, pid)
	}()
	return nil
}

func (h *Host) handleExit(id string, pid uint32) {
	s := h.state(id)
	h.mu.Lock()
	if s.pid != pid {
		h.mu.Unlock()
		return
	}
	s.pid, s.startTime, s.cpu, s.ram, s.players = 0, time.Time{}, 0, 0, 0
	crashed := !s.userStop && !s.busy
	h.mu.Unlock()
	if !s.busy {
		h.setStatus(id, StatusOffline)
	}
	if crashed {
		h.logAction(id, "crash", "error", "server process exited unexpectedly")
		if sv, err := h.store.Get(id); err == nil && sv.ForceRunning {
			delay, ok := h.crashBackoff(id)
			if !ok {
				h.logAction(id, "auto-restart", "error", "crash loop detected (5 crashes in 15 minutes); auto-restart suspended; start manually once fixed")
				return
			}
			go func() {
				time.Sleep(delay)
				if !h.isRunning(id) && !h.state(id).busy {
					h.logAction(id, "auto-restart", "info", "force running enabled")
					h.Start(id)
				}
			}()
		}
	}
}

func (h *Host) waitExit(id string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		s := h.state(id)
		h.mu.Lock()
		pid := s.pid
		h.mu.Unlock()
		if pid == 0 || !processAlive(pid) {
			if pid != 0 {
				h.handleExit(id, pid)
			}
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

func (h *Host) stopInternal(id string) error {
	s := h.state(id)
	h.mu.Lock()
	pid := s.pid
	s.userStop = true
	h.mu.Unlock()
	if pid == 0 {
		return nil
	}
	if err := killProcess(pid); err != nil && processAlive(pid) {
		return err
	}
	h.waitExit(id, 20*time.Second)
	return nil
}

func (h *Host) Stop(id string) error {
	if sv, err := h.store.Get(id); err == nil && isScript(sv) {
		return h.scriptStopOp(sv, "stop")
	}
	h.setStatus(id, StatusStopping)
	err := h.stopInternal(id)
	h.logAction(id, "stop", statusOf(err), errStr(err))
	h.setStatus(id, StatusOffline)
	return err
}

func (h *Host) shutdownInternal(id string) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	if !h.isRunning(id) {
		return nil
	}
	s := h.state(id)
	h.mu.Lock()
	s.userStop = true
	h.mu.Unlock()
	if _, err := h.rcon(sv, saveCmd(sv), 60*time.Second); err != nil {
		h.mu.Lock()
		s.userStop = false
		h.mu.Unlock()
		return fmt.Errorf("RCON unavailable (%v) - use Stop to kill the process", err)
	}
	time.Sleep(3 * time.Second)
	h.rcon(sv, "DoExit", 15*time.Second)
	if !h.waitExit(id, 3*time.Minute) {
		return h.stopInternal(id)
	}
	return nil
}

func (h *Host) Shutdown(id string) error {
	if sv, err := h.store.Get(id); err == nil && isScript(sv) {
		return h.scriptStopOp(sv, "shutdown")
	}
	prev := h.getStatus(id)
	h.setStatus(id, StatusShuttingDown)
	err := h.shutdownInternal(id)
	h.logAction(id, "shutdown", statusOf(err), errStr(err))
	if err != nil {
		h.setStatus(id, prev)
		return err
	}
	h.setStatus(id, StatusOffline)
	return nil
}

func (h *Host) Restart(id string) error {
	if sv, err := h.store.Get(id); err == nil && isScript(sv) {
		return h.scriptRestart(sv, "restart")
	}
	h.setStatus(id, StatusRestarting)
	err := h.shutdownInternal(id)
	if err == nil {
		h.setStatus(id, StatusOffline)
		err = h.Start(id)
	} else {
		h.setStatus(id, StatusOnline)
	}
	h.logAction(id, "restart", statusOf(err), errStr(err))
	return err
}

func (h *Host) Reboot(id string) error {
	if sv, err := h.store.Get(id); err == nil && isScript(sv) {
		return h.scriptRestart(sv, "reboot")
	}
	h.setStatus(id, StatusRestarting)
	h.stopInternal(id)
	h.setStatus(id, StatusOffline)
	err := h.Start(id)
	h.logAction(id, "reboot", statusOf(err), errStr(err))
	return err
}

func statusOf(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

func errStr(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func (h *Host) crashBackoff(id string) (time.Duration, bool) {
	s := h.state(id)
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now()
	kept := s.crashes[:0]
	for _, t := range s.crashes {
		if now.Sub(t) < 15*time.Minute {
			kept = append(kept, t)
		}
	}
	s.crashes = append(kept, now)
	if len(s.crashes) > 5 {
		return 0, false
	}
	return time.Duration(10<<uint(len(s.crashes)-1)) * time.Second, true
}
