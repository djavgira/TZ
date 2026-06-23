package collector

import (
	"context"

	"github.com/Alice/pain_tz/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector is the base interface every data collector must implement.
// Collect gathers metrics and populates the provided Snapshot struct.
type Collector interface {
	// Name returns a unique identifier for this collector, e.g. "cpu", "memory".
	Name() string

	// Collect performs one collection cycle, writing results into snap.
	// The context carries a deadline equal to the collector's configured interval.
	Collect(ctx context.Context, snap *metrics.Snapshot) error
}

// MetricsCollector extends Collector for Prometheus integration (serve mode).
type MetricsCollector interface {
	Collector

	// RegisterMetrics registers all prometheus.Collector descriptors with
	// the provided registry. Called once during initialization.
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
