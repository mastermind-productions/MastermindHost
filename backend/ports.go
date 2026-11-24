package backend

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type portBinding struct {
	Proto string
	PID   uint32
}

type PortEntry struct {
	Label     string `json:"label"`
	Proto     string `json:"proto"`
	Port      int    `json:"port"`
	Bound     bool   `json:"bound"`
	PID       uint32 `json:"pid"`
	Process   string `json:"process"`
	Ours      bool   `json:"ours"`
	Conflicts string `json:"conflicts"`
}

type ServerPorts struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Game     string      `json:"gameName"`
	Running  bool        `json:"running"`
	Firewall bool        `json:"firewall"`
	Ports    []PortEntry `json:"ports"`
}

func scanBoundPorts() (map[string]portBinding, error) {
	out := map[string]portBinding{}
	for _, proto := range []string{"UDP", "TCP"} {
		cmd := exec.Command("netstat", "-ano", "-p", proto)
		cmd.SysProcAttr = hideWindowAttr()
		b, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(string(b), "\n") {
			f := strings.Fields(line)
			if len(f) < 4 || !strings.EqualFold(f[0], proto) {
				continue
			}
			if proto == "TCP" && (len(f) < 5 || !strings.EqualFold(f[3], "LISTENING")) {
				continue
			}
			local := f[1]
			i := strings.LastIndex(local, ":")
			if i < 0 {
				continue
			}
			port, err := strconv.Atoi(local[i+1:])
			if err != nil {
				continue
			}
			pidStr := f[len(f)-1]
			pid, _ := strconv.Atoi(pidStr)
			out[fmt.Sprintf("%s/%d", proto, port)] = portBinding{Proto: proto, PID: uint32(pid)}
		}
	}
	return out, nil
}

func (h *Host) serverPortList(sv Server) []PortEntry {
	r := h.runtime(sv)
	var list []PortEntry
	if isPal(sv) {
		list = append(list, PortEntry{Label: "Game", Proto: "UDP", Port: r.Port})
		list = append(list, PortEntry{Label: "Query (A2S)", Proto: "UDP", Port: h.queryPortFor(sv)})
		if r.RconEnabled {
			list = append(list, PortEntry{Label: "RCON", Proto: "TCP", Port: r.RconPort})
		}
		return list
	}
	list = append(list, PortEntry{Label: "Game", Proto: "UDP", Port: r.Port})
	list = append(list, PortEntry{Label: "Peer", Proto: "UDP", Port: r.Port + 1})
	list = append(list, PortEntry{Label: "Query (A2S)", Proto: "UDP", Port: h.queryPortFor(sv)})
	if r.RconEnabled {
		list = append(list, PortEntry{Label: "RCON", Proto: "TCP", Port: r.RconPort})
	}
	return list
}

func (h *Host) PortsReport() ([]ServerPorts, error) {
	bound, err := scanBoundPorts()
	if err != nil {
		return nil, err
	}
	var servers []Server
	for _, sv := range h.store.List() {
		if !isScript(sv) {
			servers = append(servers, sv)
		}
	}
	type owner struct{ name, label string }
	claims := map[string][]owner{}
	reports := make([]ServerPorts, 0, len(servers))
	for _, sv := range servers {
		sp := ServerPorts{ID: sv.ID, Name: sv.Name, Game: Games[sv.Game].Name, Running: h.isRunning(sv.ID), Ports: h.serverPortList(sv)}
		sp.Firewall, _ = h.FirewallStatus(sv.ID)
		for _, p := range sp.Ports {
			k := fmt.Sprintf("%s/%d", p.Proto, p.Port)
			claims[k] = append(claims[k], owner{sv.Name, p.Label})
		}
		reports = append(reports, sp)
	}
	for i := range reports {
		sv, _ := h.store.Get(reports[i].ID)
		s := h.state(sv.ID)
		h.mu.Lock()
		pid := s.pid
		h.mu.Unlock()
		for j := range reports[i].Ports {
			p := &reports[i].Ports[j]
			k := fmt.Sprintf("%s/%d", p.Proto, p.Port)
			if b, ok := bound[k]; ok {
				p.Bound, p.PID = true, b.PID
				p.Process = filepath.Base(processImagePath(b.PID))
				if p.Process == "." {
					p.Process = ""
				}
				p.Ours = pid != 0 && b.PID == pid
			}
			var msgs []string
			for _, o := range claims[k] {
				if o.name != reports[i].Name {
					msgs = append(msgs, "also configured by "+o.name)
				}
			}
			if p.Bound && !p.Ours {
				who := p.Process
				if who == "" {
					who = fmt.Sprintf("PID %d", p.PID)
				}
				if reports[i].Running || p.Process != "" {
					msgs = append(msgs, "in use by "+who)
				}
			}
			p.Conflicts = strings.Join(msgs, "; ")
		}
	}
	return reports, nil
}

func (h *Host) FindFreePort(proto string, start int) (int, error) {
	proto = strings.ToUpper(proto)
	if proto != "UDP" && proto != "TCP" {
		return 0, errors.New("proto must be UDP or TCP")
	}
	if start < 1024 {
		start = 1024
	}
	bound, err := scanBoundPorts()
	if err != nil {
		return 0, err
	}
	used := map[int]bool{}
	for _, sv := range h.store.List() {
		for _, p := range h.serverPortList(sv) {
			used[p.Port] = true
		}
	}
	for p := start; p <= 65535; p++ {
		if used[p] {
			continue
		}
		if _, ok := bound[fmt.Sprintf("%s/%d", proto, p)]; ok {
			continue
		}
		return p, nil
	}
	return 0, errors.New("no free port found")
}
