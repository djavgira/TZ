package collector

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// Collector is the base interface every data collector must implement.
type Collector interface {
	// Name returns a unique identifier for this collector, e.g. "cpu", "memory".
	Name() string

	// Collect performs one collection cycle. It must be safe for concurrent use.
	// The context carries a deadline equal to the collector's configured interval.
	Collect(ctx context.Context) error
}

// MetricsCollector extends Collector for Prometheus integration.
// Collectors that expose metrics implement this interface to register their
// descriptors with the provided prometheus.Registry exactly once before the
// first collection cycle.
type MetricsCollector interface {
	Collector

	// RegisterMetrics registers all prometheus.Collector descriptors (gauges,
	// counters, histograms) with the provided registry. Called once during
	// initialization, before Start().
	// Must be idempotent — safe to call multiple times.
	RegisterMetrics(reg *prometheus.Registry) error
}

// StatusProvider is an optional interface for collectors that can report
// their health status (used by /health and /ready endpoints).
type StatusProvider interface {
	// Healthy returns nil if the collector is functioning correctly,
	// or an error describing the problem.
	Healthy() error
}
