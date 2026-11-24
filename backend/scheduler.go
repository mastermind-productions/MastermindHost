package backend

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Schedule struct {
	ID          string    `json:"id"`
	Enabled     bool      `json:"enabled"`
	Action      string    `json:"action"`
	Mode        string    `json:"mode"`
	Time        string    `json:"time"`
	Days        []int     `json:"days"`
	IntervalMin int       `json:"intervalMin"`
	Arg         string    `json:"arg"`
	WarnMinutes int       `json:"warnMinutes"`
	LastRun     time.Time `json:"lastRun"`
}

var scheduleActions = map[string]bool{"restart": true, "shutdown": true, "start": true, "stop": true,
	"backup": true, "update": true, "broadcast": true, "rcon": true}

func (h *Host) GetSchedules(id string) ([]Schedule, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return nil, err
	}
	if sv.Schedules == nil {
		return []Schedule{}, nil
	}
	return sv.Schedules, nil
}

func (h *Host) SaveSchedule(id string, sc Schedule) (Schedule, error) {
	if !scheduleActions[sc.Action] {
		return sc, fmt.Errorf("unknown action %q", sc.Action)
	}
	switch sc.Mode {
	case "daily":
		if _, err := time.Parse("15:04", sc.Time); err != nil {
			return sc, errors.New("time must be HH:MM (24h)")
		}
	case "interval":
		if sc.IntervalMin < 1 {
			return sc, errors.New("interval must be at least 1 minute")
		}
	default:
		return sc, errors.New("mode must be daily or interval")
	}
	if sc.WarnMinutes < 0 {
		sc.WarnMinutes = 0
	}
	err := h.store.Update(id, func(s *Server) error {
		if sc.ID == "" {
			sc.ID = newID()
			if sc.Mode == "interval" {
				sc.LastRun = time.Now()
			}
			s.Schedules = append(s.Schedules, sc)
			return nil
		}
		for i := range s.Schedules {
			if s.Schedules[i].ID == sc.ID {
				if sc.Mode == "interval" && s.Schedules[i].Mode != "interval" {
					sc.LastRun = time.Now()
				} else {
					sc.LastRun = s.Schedules[i].LastRun
				}
				s.Schedules[i] = sc
				return nil
			}
		}
		return errors.New("schedule not found")
	})
	return sc, err
}

func (h *Host) DeleteSchedule(id, scheduleID string) error {
	return h.store.Update(id, func(s *Server) error {
		for i := range s.Schedules {
			if s.Schedules[i].ID == scheduleID {
				s.Schedules = append(s.Schedules[:i], s.Schedules[i+1:]...)
				return nil
			}
		}
		return errors.New("schedule not found")
	})
}

func (h *Host) RunScheduleNow(id, scheduleID string) error {
	sv, err := h.store.Get(id)
	if err != nil {
		return err
	}
	for _, sc := range sv.Schedules {
		if sc.ID == scheduleID {
			go h.runScheduled(id, sc, time.Now())
			return nil
		}
	}
	return errors.New("schedule not found")
}

func (sc Schedule) nextDue(now time.Time) time.Time {
	if sc.Mode == "interval" {
		return sc.LastRun.Add(time.Duration(sc.IntervalMin) * time.Minute)
	}
	t, _ := time.Parse("15:04", sc.Time)
	for d := -1; d <= 7; d++ {
		day := now.AddDate(0, 0, d)
		due := time.Date(day.Year(), day.Month(), day.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
		if !due.After(sc.LastRun) || !sc.dayAllowed(due.Weekday()) {
			continue
		}
		return due
	}
	return time.Time{}
}

func (sc Schedule) dayAllowed(w time.Weekday) bool {
	if len(sc.Days) == 0 {
		return true
	}
	for _, d := range sc.Days {
		if d == int(w) {
			return true
		}
	}
	return false
}

func (h *Host) schedulerLoop() {
	for {
		time.Sleep(15 * time.Second)
		now := time.Now()
		for _, sv := range h.store.List() {
			for _, sc := range sv.Schedules {
				if !sc.Enabled {
					continue
				}
				due := sc.nextDue(now)
				if due.IsZero() {
					continue
				}
				fireAt := due.Add(-time.Duration(sc.WarnMinutes) * time.Minute)
				if now.Before(fireAt) {
					continue
				}
				if sc.Mode == "daily" && now.Sub(due) > 5*time.Minute {
					h.markRun(sv.ID, sc.ID, due)
					continue
				}
				h.markRun(sv.ID, sc.ID, due)
				go h.runScheduled(sv.ID, sc, due)
			}
		}
	}
}

func (h *Host) markRun(id, scID string, t time.Time) {
	h.store.Update(id, func(s *Server) error {
		for i := range s.Schedules {
			if s.Schedules[i].ID == scID {
				s.Schedules[i].LastRun = t
			}
		}
		return nil
	})
}

func (h *Host) runScheduled(id string, sc Schedule, due time.Time) {
	if sc.Action == "update" {
		h.CheckForUpdates()
		if inf, err := h.Get(id); err == nil && inf.Installed && !inf.UpdateAvailable {
			h.logAction(id, "scheduled update", "info", "already up to date")
			return
		}
	}
	warnable := sc.Action == "restart" || sc.Action == "shutdown" || sc.Action == "stop" || sc.Action == "update"
	if warnable && h.isRunning(id) {
		for {
			left := time.Until(due)
			if left <= 0 {
				break
			}
			mins := int((left + 30*time.Second) / time.Minute)
			if mins >= 1 && (mins <= 5 || mins%5 == 0 || mins == sc.WarnMinutes) {
				msg := sc.Arg
				if msg == "" {
					msg = "Server %s in %d minute(s)."
				}
				if strings.Contains(msg, "%") {
					msg = fmt.Sprintf(msg, sc.Action, mins)
				}
				h.Broadcast(id, msg)
			}
			sleep := time.Minute
			if left < sleep {
				sleep = left
			}
			time.Sleep(sleep)
		}
	} else if d := time.Until(due); d > 0 {
		time.Sleep(d)
	}
	var err error
	switch sc.Action {
	case "restart":
		if h.isRunning(id) {
			err = h.Restart(id)
		} else {
			err = h.Start(id)
		}
	case "shutdown":
		err = h.Shutdown(id)
	case "stop":
		err = h.Stop(id)
	case "start":
		if !h.isRunning(id) {
			err = h.Start(id)
		}
	case "backup":
		_, err = h.Backup(id)
	case "update":
		err = h.Update(id, false)
	case "broadcast":
		_, err = h.Broadcast(id, sc.Arg)
	case "rcon":
		_, err = h.RconCommand(id, sc.Arg)
	}
	h.logAction(id, "scheduled "+sc.Action, statusOf(err), errStr(err))
}
