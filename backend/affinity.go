package backend

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
)

func cpusToMask(cpus []int) (uintptr, error) {
	var m uintptr
	for _, c := range cpus {
		if c < 0 || c >= runtime.NumCPU() || c >= 64 {
			return 0, fmt.Errorf("invalid CPU index %d", c)
		}
		m |= 1 << uint(c)
	}
	if m == 0 {
		return 0, errors.New("select at least one CPU")
	}
	return m, nil
}

func maskToCPUs(m uintptr) []int {
	var out []int
	for i := 0; i < 64 && i < runtime.NumCPU(); i++ {
		if m&(1<<uint(i)) != 0 {
			out = append(out, i)
		}
	}
	return out
}

func (h *Host) applyAffinity(id string) {
	sv, err := h.store.Get(id)
	if err != nil || len(sv.Affinity) == 0 {
		return
	}
	s := h.state(id)
	h.mu.Lock()
	pid := s.pid
	h.mu.Unlock()
	if pid == 0 {
		return
	}
	mask, err := cpusToMask(sv.Affinity)
	if err != nil {
		return
	}
	err = setProcessAffinity(pid, mask)
	h.logAction(id, "cpu affinity", statusOf(err), errStr(err))
}

type AffinityInfo struct {
	Cores   int   `json:"cores"`
	Saved   []int `json:"saved"`
	Current []int `json:"current"`
	Running bool  `json:"running"`
}

func (h *Host) GetAffinity(id string) (AffinityInfo, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return AffinityInfo{}, err
	}
	info := AffinityInfo{Cores: runtime.NumCPU(), Saved: sv.Affinity}
	if info.Saved == nil {
		info.Saved = []int{}
	}
	s := h.state(id)
	h.mu.Lock()
	pid := s.pid
	h.mu.Unlock()
	info.Running = pid != 0
	info.Current = []int{}
	if pid != 0 {
		if m, err := getProcessAffinity(pid); err == nil {
			info.Current = maskToCPUs(m)
		}
	}
	return info, nil
}

func (h *Host) SetAffinity(id string, cpus []int) error {
	if len(cpus) > 0 {
		if _, err := cpusToMask(cpus); err != nil {
			return err
		}
		sort.Ints(cpus)
	}
	if err := h.store.Update(id, func(sv *Server) error { sv.Affinity = cpus; return nil }); err != nil {
		return err
	}
	s := h.state(id)
	h.mu.Lock()
	pid := s.pid
	h.mu.Unlock()
	if pid == 0 {
		return nil
	}
	if len(cpus) == 0 {
		cpus = make([]int, 0, runtime.NumCPU())
		for i := 0; i < runtime.NumCPU() && i < 64; i++ {
			cpus = append(cpus, i)
		}
	}
	mask, _ := cpusToMask(cpus)
	err := setProcessAffinity(pid, mask)
	h.logAction(id, "cpu affinity", statusOf(err), errStr(err))
	return err
}
