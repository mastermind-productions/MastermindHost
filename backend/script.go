package backend

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ScriptProc struct {
	Name string `json:"name"`
	Exe  string `json:"exe"`
	Port int    `json:"port"`
}

type ScriptSpec struct {
	WorkDir  string       `json:"workDir"`
	StartCmd string       `json:"startCmd"`
	StopCmd  string       `json:"stopCmd"`
	Procs    []ScriptProc `json:"procs"`
}

type ScriptComponent struct {
	Name      string `json:"name"`
	Exe       string `json:"exe"`
	Port      int    `json:"port"`
	Running   bool   `json:"running"`
	Listening bool   `json:"listening"`
	PID       uint32 `json:"pid"`
}

func (sp *ScriptSpec) exe(p ScriptProc) string {
	if filepath.IsAbs(p.Exe) {
		return filepath.Clean(p.Exe)
	}
	return filepath.Join(sp.WorkDir, p.Exe)
}

func findFirst(dir string, rels ...string) string {
	for _, r := range rels {
		if _, err := os.Stat(filepath.Join(dir, r)); err == nil {
			return filepath.Join(dir, r)
		}
	}
	return ""
}

func startCmdFor(script string) string {
	if strings.EqualFold(filepath.Ext(script), ".ps1") {
		return `powershell -NoProfile -ExecutionPolicy Bypass -File "` + script + `"`
	}
	if strings.EqualFold(filepath.Base(script), "Start-Servers.bat") {
		return `"` + script + `" -NoPause`
	}
	return `"` + script + `"`
}

func discoverScript(game, dir string) (ScriptSpec, error) {
	sp := ScriptSpec{WorkDir: dir}
	switch game {
	case "rappelz":
		sp.Procs = []ScriptProc{
			{"LogServer", `Log_bin\LogServer.exe`, 4516},
			{"ChattingLogServer", `LogChat_bin\ChattingLogServer.exe`, 4520},
			{"UploadServer", `Upload_bin\UploadServer.exe`, 4615},
			{"PatchServer", `patch_bin\PatchServer.exe`, 3516},
			{"AuthServer", `Auth_Bin\PrincessAuroraServer.exe`, 4501},
			{"GameServer", `Game_Bin\CaptainHerlockServer.exe`, 34391},
		}
	case "scions":
		matches, _ := filepath.Glob(filepath.Join(dir, "Server", "Source_V20", "*", "LS", "loginServer.exe"))
		if len(matches) == 0 {
			return sp, fmt.Errorf("loginServer.exe not found under %s\\Server\\Source_V20", dir)
		}
		root := filepath.Dir(filepath.Dir(matches[0]))
		sp.Procs = []ScriptProc{
			{"loginServer", filepath.Join(root, "LS", "loginServer.exe"), 1300},
			{"GameServer", filepath.Join(root, "GS", "GameServer.exe"), 6773},
		}
	case "hero":
		exe := findFirst(dir, "herorebirth-server.exe", `HeroRebirthFork\herorebirth-server.exe`)
		if exe == "" {
			return sp, fmt.Errorf("herorebirth-server.exe not found in %s", dir)
		}
		sp.WorkDir = filepath.Dir(exe)
		sp.Procs = []ScriptProc{{"herorebirth-server", exe, 5310}}
	default:
		return sp, fmt.Errorf("no script defaults for %q", game)
	}
	scripts := []string{"Start-Servers.bat", "Start_Servers.bat", "Start-Servers.ps1", "Start_Servers.ps1", "start.bat", "start_server.bat"}
	if game == "scions" {
		scripts = []string{"Start_Servers.ps1", "Start_Servers.bat"}
	}
	if script := findFirst(sp.WorkDir, scripts...); script != "" {
		sp.StartCmd = startCmdFor(script)
	}
	if sp.StartCmd == "" && len(sp.Procs) == 0 {
		return sp, errors.New("no start script or executables found")
	}
	return sp, nil
}

func (h *Host) importScript(name, dir string, g GameDef) (string, error) {
	dir = strings.TrimSpace(dir)
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return "", fmt.Errorf("folder not found: %s", dir)
	}
	sp, err := discoverScript(g.ID, dir)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(name) == "" {
		name = g.Name
	}
	sv := &Server{ID: newID(), Name: strings.TrimSpace(name), Game: g.ID, Dir: dir, CreatedAt: time.Now(),
		Backup: BackupSettings{Keep: 20, AutoRemove: true}, ModNames: map[string]string{}, Script: &sp}
	if err := h.store.Add(sv); err != nil {
		return "", err
	}
	h.logAction(sv.ID, "import", "success", "")
	return sv.ID, nil
}

func (h *Host) GetScript(id string) (ScriptSpec, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return ScriptSpec{}, err
	}
	if sv.Script == nil {
		return ScriptSpec{}, errors.New("not a script server")
	}
	out := *sv.Script
	out.Procs = append([]ScriptProc{}, out.Procs...)
	return out, nil
}

func (h *Host) SetScript(id string, spec ScriptSpec) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	if !isScript(sv) {
		return errors.New("not a script server")
	}
	spec.WorkDir = strings.TrimSpace(spec.WorkDir)
	spec.StartCmd = strings.TrimSpace(spec.StartCmd)
	spec.StopCmd = strings.TrimSpace(spec.StopCmd)
	if spec.WorkDir == "" {
		return errors.New("working folder is required")
	}
	var procs []ScriptProc
	for _, p := range spec.Procs {
		p.Name, p.Exe = strings.TrimSpace(p.Name), strings.TrimSpace(p.Exe)
		if p.Exe == "" {
			continue
		}
		if p.Name == "" {
			p.Name = strings.TrimSuffix(filepath.Base(p.Exe), filepath.Ext(p.Exe))
		}
		procs = append(procs, p)
	}
	spec.Procs = procs
	if spec.StartCmd == "" && len(procs) == 0 {
		return errors.New("a start command or at least one executable is required")
	}
	if err := h.store.Update(id, func(s *Server) error { s.Script = &spec; return nil }); err != nil {
		return err
	}
	h.emit(id)
	return nil
}

func (h *Host) scriptInstalled(sv Server) bool {
	sp := sv.Script
	if sp == nil {
		return false
	}
	if st, err := os.Stat(sp.WorkDir); err != nil || !st.IsDir() {
		return false
	}
	for _, p := range sp.Procs {
		if _, err := os.Stat(sp.exe(p)); err == nil {
			return true
		}
	}
	return sp.StartCmd != ""
}

func (h *Host) scriptPIDs(sv Server) []uint32 {
	if sv.Script == nil {
		return nil
	}
	out := make([]uint32, len(sv.Script.Procs))
	for i, p := range sv.Script.Procs {
		out[i] = findProcessByPath(sv.Script.exe(p))
	}
	return out
}

func (h *Host) scriptUp(sv Server) int {
	n := 0
	for _, pid := range h.scriptPIDs(sv) {
		if pid != 0 {
			n++
		}
	}
	return n
}

var listenCache struct {
	sync.Mutex
	at    time.Time
	ports map[int]bool
}

func listeningPorts() map[int]bool {
	listenCache.Lock()
	defer listenCache.Unlock()
	if listenCache.ports != nil && time.Since(listenCache.at) < 4*time.Second {
		return listenCache.ports
	}
	m := map[int]bool{}
	if b, err := scanBoundPorts(); err == nil {
		for k := range b {
			var proto string
			var port int
			if i := strings.Index(k, "/"); i > 0 {
				proto = k[:i]
				fmt.Sscan(k[i+1:], &port)
			}
			if proto != "" {
				m[port] = true
			}
		}
	}
	listenCache.ports, listenCache.at = m, time.Now()
	return m
}

func (h *Host) ScriptComponents(id string) ([]ScriptComponent, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return nil, err
	}
	if sv.Script == nil {
		return nil, errors.New("not a script server")
	}
	pids := h.scriptPIDs(sv)
	var listen map[int]bool
	out := make([]ScriptComponent, 0, len(pids))
	for i, p := range sv.Script.Procs {
		c := ScriptComponent{Name: p.Name, Exe: sv.Script.exe(p), Port: p.Port, PID: pids[i], Running: pids[i] != 0}
		if c.Running && p.Port > 0 {
			if listen == nil {
				listen = listeningPorts()
			}
			c.Listening = listen[p.Port]
		}
		out = append(out, c)
	}
	return out, nil
}

func (h *Host) scriptInfo(sv Server) ServerInfo {
	g := Games[sv.Game]
	s := h.state(sv.ID)
	h.mu.Lock()
	status, cpu, ram, start := s.status, s.cpu, s.ram, s.startTime
	h.mu.Unlock()
	if status == "" {
		status = h.idleStatus(sv)
	}
	inf := ServerInfo{
		ID: sv.ID, Name: sv.Name, Game: sv.Game, GameName: g.Name, Dir: sv.Dir,
		CreatedAt: sv.CreatedAt.Format(time.RFC3339), Status: status, Installed: h.installed(sv),
		CPU: cpu, RAM: ram, Group: g.Group, Script: true, ForceRunning: sv.ForceRunning,
	}
	if sv.Script != nil {
		inf.CompTotal = len(sv.Script.Procs)
		if len(sv.Script.Procs) > 0 {
			inf.Port = sv.Script.Procs[len(sv.Script.Procs)-1].Port
		}
	}
	h.mu.Lock()
	inf.CompUp = s.compUp
	h.mu.Unlock()
	if !start.IsZero() {
		inf.StartTime = start.UnixMilli()
	}
	return inf
}

func (h *Host) monitorScript(sv Server) {
	if sv.Script == nil {
		return
	}
	s := h.state(sv.ID)
	h.mu.Lock()
	busy, status, launching, started := s.busy, s.status, s.launching, s.startTime
	h.mu.Unlock()
	if busy {
		return
	}
	pids := h.scriptPIDs(sv)
	up := 0
	var cpuSum time.Duration
	var ramSum uint64
	var earliest time.Time
	for _, pid := range pids {
		if pid == 0 {
			continue
		}
		up++
		if c, st, r, ok := processTimes(pid); ok {
			cpuSum += c
			ramSum += r
			if earliest.IsZero() || st.Before(earliest) {
				earliest = st
			}
		}
	}
	now := time.Now()
	h.mu.Lock()
	s.compUp = up
	if up > 0 {
		if !s.lastSample.IsZero() && cpuSum >= s.lastCPU {
			if wall := now.Sub(s.lastSample); wall > 0 {
				s.cpu = float64(cpuSum-s.lastCPU) / float64(wall) / float64(runtime.NumCPU()) * 100
			}
		}
		s.lastCPU, s.lastSample, s.ram = cpuSum, now, ramSum
		if !earliest.IsZero() {
			s.startTime = earliest
		}
		h.metrics.add(sv.ID, MetricPoint{T: now.UnixMilli(), CPU: s.cpu, RAM: ramSum})
	} else {
		s.cpu, s.ram, s.lastSample, s.lastCPU = 0, 0, time.Time{}, 0
		if !launching {
			s.startTime = time.Time{}
		}
	}
	h.mu.Unlock()
	if up == 0 {
		if launching || (status == StatusStarting && time.Since(started) < 15*time.Second) {
			return
		}
		h.setStatus(sv.ID, h.idleStatus(sv))
		return
	}
	total := len(pids)
	allListen := true
	var listen map[int]bool
	for i, p := range sv.Script.Procs {
		if p.Port > 0 && pids[i] != 0 {
			if listen == nil {
				listen = listeningPorts()
			}
			if !listen[p.Port] {
				allListen = false
			}
		}
	}
	switch {
	case up == total && allListen:
		h.setStatus(sv.ID, StatusOnline)
	case launching || (status == StatusStarting && time.Since(started) < 10*time.Minute):
		h.setStatus(sv.ID, StatusStarting)
	default:
		h.setStatus(sv.ID, StatusRunning)
	}
	h.emit(sv.ID)
}

func shellCmd(command, dir string) *exec.Cmd {
	cmd := exec.Command("cmd.exe")
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000, CmdLine: `cmd.exe /S /C "` + command + `"`}
	return cmd
}

func (h *Host) startScript(sv Server) error {
	sp := sv.Script
	if sp == nil {
		return errors.New("no launch configuration")
	}
	if !h.scriptInstalled(sv) {
		return fmt.Errorf("server files not found in %s", sp.WorkDir)
	}
	pids := h.scriptPIDs(sv)
	allUp := len(pids) > 0
	for _, p := range pids {
		if p == 0 {
			allUp = false
		}
	}
	if allUp {
		return errors.New("server is already running")
	}
	s := h.state(sv.ID)
	h.mu.Lock()
	if s.launching {
		h.mu.Unlock()
		return errors.New("server is already starting")
	}
	s.launching, s.userStop, s.startTime = true, false, time.Now()
	h.mu.Unlock()
	h.setStatus(sv.ID, StatusStarting)
	h.logAction(sv.ID, "start", "success", "")
	go func() {
		err := h.runLaunch(sv)
		h.mu.Lock()
		s.launching = false
		h.mu.Unlock()
		if err != nil {
			h.logAction(sv.ID, "start", "error", err.Error())
		}
		h.emit(sv.ID)
	}()
	return nil
}

func (h *Host) runLaunch(sv Server) error {
	sp := sv.Script
	if sp.StartCmd != "" {
		cmd := shellCmd(sp.StartCmd, sp.WorkDir)
		if err := cmd.Start(); err != nil {
			return err
		}
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("start command failed: %v", err)
		}
		return nil
	}
	for _, p := range sp.Procs {
		exe := sp.exe(p)
		if findProcessByPath(exe) != 0 {
			continue
		}
		cmd := exec.Command(exe)
		cmd.Dir = filepath.Dir(exe)
		if sp.WorkDir != "" {
			if _, err := os.Stat(sp.WorkDir); err == nil && filepath.IsAbs(p.Exe) == false {
				cmd.Dir = sp.WorkDir
			}
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000010, CmdLine: `"` + exe + `"`}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("%s: %v", p.Name, err)
		}
		cmd.Process.Release()
		if p.Port > 0 {
			deadline := time.Now().Add(60 * time.Second)
			for time.Now().Before(deadline) && !listeningPortsFresh()[p.Port] {
				time.Sleep(time.Second)
			}
		}
	}
	return nil
}

func listeningPortsFresh() map[int]bool {
	listenCache.Lock()
	listenCache.ports = nil
	listenCache.Unlock()
	return listeningPorts()
}

func closeGracefully(pid uint32, wait time.Duration) error {
	if !processAlive(pid) {
		return nil
	}
	cmd := exec.Command("taskkill", "/PID", fmt.Sprint(pid))
	cmd.SysProcAttr = hideWindowAttr()
	if cmd.Run() == nil {
		deadline := time.Now().Add(wait)
		for time.Now().Before(deadline) {
			if !processAlive(pid) {
				return nil
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
	if err := killProcess(pid); err != nil && processAlive(pid) {
		return err
	}
	return nil
}

func (h *Host) stopScript(sv Server) error {
	sp := sv.Script
	if sp == nil {
		return errors.New("no launch configuration")
	}
	if sp.StopCmd != "" {
		cmd := shellCmd(sp.StopCmd, sp.WorkDir)
		if err := cmd.Start(); err == nil {
			done := make(chan struct{})
			go func() { cmd.Wait(); close(done) }()
			select {
			case <-done:
			case <-time.After(3 * time.Minute):
				cmd.Process.Kill()
			}
		}
	}
	var errs []string
	for i := len(sp.Procs) - 1; i >= 0; i-- {
		pid := findProcessByPath(sp.exe(sp.Procs[i]))
		if pid == 0 {
			continue
		}
		if err := closeGracefully(pid, 15*time.Second); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", sp.Procs[i].Name, err))
		}
	}
	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

func (h *Host) scriptStopOp(sv Server, action string) error {
	s := h.state(sv.ID)
	h.mu.Lock()
	s.userStop, s.busy = true, true
	h.mu.Unlock()
	h.setStatus(sv.ID, StatusStopping)
	err := h.stopScript(sv)
	h.mu.Lock()
	s.busy = false
	if err == nil {
		s.cpu, s.ram, s.compUp, s.startTime, s.lastSample = 0, 0, 0, time.Time{}, time.Time{}
	}
	h.mu.Unlock()
	h.logAction(sv.ID, action, statusOf(err), errStr(err))
	if err != nil {
		h.setStatus(sv.ID, StatusRunning)
		return err
	}
	h.setStatus(sv.ID, StatusOffline)
	return nil
}

func (h *Host) scriptRestart(sv Server, action string) error {
	if err := h.scriptStopOp(sv, action); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	return h.startScript(sv)
}
