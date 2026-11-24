package backend

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

func (h *Host) rconCreds(sv Server) (int, string) {
	r := h.runtime(sv)
	return r.RconPort, r.RconPass
}

func (h *Host) rcon(sv Server, command string, timeout time.Duration) (string, error) {
	port, pass := h.rconCreds(sv)
	if pass == "" {
		if isPal(sv) {
			return "", fmt.Errorf("AdminPassword is not set in PalWorldSettings.ini")
		}
		return "", fmt.Errorf("ServerAdminPassword is not set")
	}
	return RconExec("127.0.0.1", port, pass, command, timeout)
}

func (h *Host) RconCommand(id, command string) (string, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return "", err
	}
	out, err := h.rcon(sv, command, 15*time.Second)
	h.logAction(id, "rcon: "+command, statusOf(err), errStr(err))
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(out) == "" {
		out = "(no response)"
	}
	return out, nil
}

type Player struct {
	Index string `json:"index"`
	Name  string `json:"name"`
	ID    string `json:"id"`
}

func parsePlayers(out string) []Player {
	var list []Player
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		dot := strings.Index(line, ". ")
		comma := strings.LastIndex(line, ",")
		if dot <= 0 || comma <= dot {
			continue
		}
		list = append(list, Player{Index: line[:dot], Name: strings.TrimSpace(line[dot+2 : comma]), ID: strings.TrimSpace(line[comma+1:])})
	}
	return list
}

func (h *Host) Players(id string) ([]Player, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return nil, err
	}
	out, err := h.rcon(sv, playersCmd(sv), 10*time.Second)
	if err != nil {
		return nil, err
	}
	p := parsePlayersFor(sv, out)
	if p == nil {
		p = []Player{}
	}
	return p, nil
}

func (h *Host) PlayerKick(id, playerID string) (string, error) {
	return h.RconCommand(id, "KickPlayer "+playerID)
}

func (h *Host) PlayerBan(id, playerID string) (string, error) {
	return h.RconCommand(id, "BanPlayer "+playerID)
}

func (h *Host) PlayerUnban(id, playerID string) (string, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return "", err
	}
	return h.RconCommand(id, unbanCmd(sv)+playerID)
}

func (h *Host) Broadcast(id, message string) (string, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return "", err
	}
	return h.RconCommand(id, broadcastCmd(sv, message))
}

func (h *Host) monitorLoop() {
	tick := 0
	for {
		time.Sleep(2 * time.Second)
		tick++
		h.recordHostMetric()
		for _, sv := range h.store.List() {
			h.monitorOne(sv, tick)
		}
	}
}

func (h *Host) monitorOne(sv Server, tick int) {
	if isScript(sv) {
		h.monitorScript(sv)
		return
	}
	s := h.state(sv.ID)
	h.mu.Lock()
	pid, busy, status := s.pid, s.busy, s.status
	h.mu.Unlock()
	if busy {
		return
	}
	if pid == 0 {
		if status == "" || tick%5 == 0 {
			if h.installed(sv) {
				if found := findProcessByPath(h.exePath(sv)); found != 0 {
					_, start, _, _ := processTimes(found)
					h.mu.Lock()
					s.pid, s.startTime, s.userStop = found, start, false
					h.mu.Unlock()
					h.applyAffinity(sv.ID)
					h.setStatus(sv.ID, StatusStarting)
					go h.watchAdopted(sv.ID, found)
					return
				}
			}
			idle := h.idleStatus(sv)
			if status != StatusStopping && status != StatusShuttingDown && status != StatusRestarting {
				h.setStatus(sv.ID, idle)
			}
		}
		return
	}
	cpu, _, ram, ok := processTimes(pid)
	if !ok {
		return
	}
	h.trackServerWindow(sv, pid)
	now := time.Now()
	h.mu.Lock()
	if !s.lastSample.IsZero() {
		wall := now.Sub(s.lastSample)
		if wall > 0 {
			s.cpu = float64(cpu-s.lastCPU) / float64(wall) / float64(runtime.NumCPU()) * 100
		}
	}
	s.lastCPU, s.lastSample, s.ram = cpu, now, ram
	h.metrics.add(sv.ID, MetricPoint{T: now.UnixMilli(), CPU: s.cpu, RAM: ram})
	needRcon := now.Sub(s.lastRcon) > 10*time.Second
	if needRcon {
		s.lastRcon = now
	}
	started := s.startTime
	h.mu.Unlock()
	if !needRcon || (status != StatusStarting && status != StatusRunning && status != StatusOnline) {
		h.emit(sv.ID)
		return
	}
	out, err := h.rcon(sv, playersCmd(sv), 4*time.Second)
	if err == nil {
		h.mu.Lock()
		s.players = len(parsePlayersFor(sv, out))
		h.mu.Unlock()
		h.setStatus(sv.ID, StatusOnline)
	} else if status == StatusOnline || time.Since(started) > 20*time.Minute {
		h.setStatus(sv.ID, StatusRunning)
	}
	h.emit(sv.ID)
}

func (h *Host) watchAdopted(id string, pid uint32) {
	for processAlive(pid) {
		time.Sleep(2 * time.Second)
	}
	h.handleExit(id, pid)
}
