package app

import (
	"fmt"
	"io"
	"time"
)

type timingReport struct {
	order     []string
	durations map[string]time.Duration
}

func newTimingReport() *timingReport {
	return &timingReport{
		order:     make([]string, 0),
		durations: make(map[string]time.Duration),
	}
}

func (r *timingReport) Add(name string, duration time.Duration) {
	if r == nil {
		return
	}
	if _, exists := r.durations[name]; !exists {
		r.order = append(r.order, name)
	}
	r.durations[name] += duration
}

func (r *timingReport) Merge(prefix string, other *timingReport) {
	if r == nil || other == nil {
		return
	}
	for _, name := range other.order {
		r.Add(prefix+name, other.durations[name])
	}
}

func (r *timingReport) Print(out io.Writer) {
	if r == nil || out == nil || len(r.order) == 0 {
		return
	}
	fmt.Fprintln(out, "timings:")
	for _, name := range r.order {
		fmt.Fprintf(out, "  %s: %s\n", name, formatTimingDuration(r.durations[name]))
	}
}

func formatTimingDuration(duration time.Duration) string {
	switch {
	case duration >= time.Second:
		return duration.Round(time.Millisecond).String()
	case duration >= time.Millisecond:
		return duration.Round(time.Microsecond).String()
	case duration >= time.Microsecond:
		return duration.Round(time.Microsecond).String()
	default:
		return duration.String()
	}
}

func measureTiming(report *timingReport, name string, fn func() error) error {
	start := time.Now()
	err := fn()
	report.Add(name, time.Since(start))
	return err
}
