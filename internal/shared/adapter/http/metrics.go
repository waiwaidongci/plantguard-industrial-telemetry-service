package http

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Metrics struct {
	mu       sync.RWMutex
	started  time.Time
	counters map[string]int64
	labels   map[string]map[string]int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		started:  time.Now(),
		counters: map[string]int64{},
		labels:   map[string]map[string]int64{},
	}
}

func (m *Metrics) Inc(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters[name]++
}

func (m *Metrics) IncLabel(name, label string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.labels[name] == nil {
		m.labels[name] = map[string]int64{}
	}
	m.labels[name][label]++
}

func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		m.mu.RLock()
		defer m.mu.RUnlock()
		_, _ = w.Write([]byte("# HELP plantguard_uptime_seconds Time since the process started.\n"))
		_, _ = w.Write([]byte("# TYPE plantguard_uptime_seconds gauge\n"))
		_, _ = w.Write([]byte("plantguard_uptime_seconds " + strconv.FormatFloat(time.Since(m.started).Seconds(), 'f', 3, 64) + "\n"))
		names := make([]string, 0, len(m.counters))
		for name := range m.counters {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			_, _ = w.Write([]byte("# TYPE " + name + " counter\n"))
			_, _ = w.Write([]byte(name + " " + strconv.FormatInt(m.counters[name], 10) + "\n"))
		}
		labelNames := make([]string, 0, len(m.labels))
		for name := range m.labels {
			labelNames = append(labelNames, name)
		}
		sort.Strings(labelNames)
		for _, name := range labelNames {
			_, _ = w.Write([]byte("# TYPE " + name + " counter\n"))
			labels := make([]string, 0, len(m.labels[name]))
			for label := range m.labels[name] {
				labels = append(labels, label)
			}
			sort.Strings(labels)
			for _, label := range labels {
				_, _ = w.Write([]byte(name + `{label="` + label + `"} ` + strconv.FormatInt(m.labels[name][label], 10) + "\n"))
			}
		}
	})
}
