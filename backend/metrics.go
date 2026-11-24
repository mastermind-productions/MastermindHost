package backend

import (
	"runtime"
	"sync"
	"time"
)

const metricsCap = 1800

type MetricPoint struct {
	T   int64   `json:"t"`
	CPU float64 `json:"cpu"`
	RAM uint64  `json:"ram"`
}

type metricsStore struct {
	mu   sync.Mutex
	data map[string][]MetricPoint
}

func (m *metricsStore) add(key string, p MetricPoint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data == nil {
		m.data = map[string][]MetricPoint{}
	}
	s := append(m.data[key], p)
	if len(s) > metricsCap {
		s = append([]MetricPoint(nil), s[len(s)-metricsCap:]...)
	}
	m.data[key] = s
}

func (m *metricsStore) get(key string, seconds int) []MetricPoint {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.data[key]
	cut := time.Now().Add(-time.Duration(seconds) * time.Second).UnixMilli()
	i := 0
	for i < len(s) && s[i].T < cut {
		i++
	}
	return append([]MetricPoint{}, s[i:]...)
}

func (h *Host) recordHostMetric() {
	total, avail := hostMemory()
	h.metrics.add("host", MetricPoint{T: time.Now().UnixMilli(), CPU: hostCPU(), RAM: total - avail})
}

func (h *Host) ServerMetrics(id string, seconds int) []MetricPoint {
	if seconds <= 0 || seconds > metricsCap*2 {
		seconds = metricsCap * 2
	}
	return h.metrics.get(id, seconds)
}

func (h *Host) HostMetrics(seconds int) map[string]any {
	if seconds <= 0 || seconds > metricsCap*2 {
		seconds = metricsCap * 2
	}
	total, _ := hostMemory()
	return map[string]any{"points": h.metrics.get("host", seconds), "ramTotal": total, "cores": runtime.NumCPU()}
}
