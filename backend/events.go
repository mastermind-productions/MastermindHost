package backend

import (
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Event struct {
	Time    time.Time `json:"time"`
	Action  string    `json:"action"`
	Result  string    `json:"result"`
	Message string    `json:"message"`
}

const maxEvents = 300

func (h *Host) logAction(id, action, result, message string) {
	h.store.Update(id, func(s *Server) error {
		s.Events = append(s.Events, Event{Time: time.Now(), Action: action, Result: result, Message: message})
		if len(s.Events) > maxEvents {
			s.Events = s.Events[len(s.Events)-maxEvents:]
		}
		return nil
	})
	if h.ctx != nil {
		wruntime.EventsEmit(h.ctx, "gs-event", id)
	}
}

func (h *Host) GetEvents(id string) ([]Event, error) {
	sv, err := h.store.Get(id)
	if err != nil {
		return nil, err
	}
	out := make([]Event, len(sv.Events))
	for i, e := range sv.Events {
		out[len(sv.Events)-1-i] = e
	}
	return out, nil
}

func (h *Host) ClearEvents(id string) error {
	return h.store.Update(id, func(s *Server) error { s.Events = nil; return nil })
}
