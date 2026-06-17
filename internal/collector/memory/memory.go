package memory

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/Alice/pain_tz/internal/config"
	"github.com/Alice/pain_tz/internal/exporter"
)

// Collector collects memory metrics using gopsutil.
type Collector struct {
	cfg     config.MemoryConfig
	healthy bool
}

// New creates a new memory collector.
func New(cfg config.MemoryConfig) *Collector {
	return &Collector{cfg: cfg}
}

// Name returns the collector name.
func (c *Collector) Name() string { return "memory" }

// RegisterMetrics registers memory-related Prometheus metrics.
func (c *Collector) RegisterMetrics(reg *prometheus.Registry) error {
	return exporter.RegisterAll(reg)
}

// Collect gathers memory metrics.
func (c *Collector) Collect(ctx context.Context) error {
	v, err := mem.VirtualMemory()
	if err != nil {
		c.healthy = false
		return fmt.Errorf("mem.VirtualMemory: %w", err)
	}

	exporter.SetMemoryTotal(float64(v.Total))
	exporter.SetMemoryUsed(float64(v.Used))
	exporter.SetMemoryAvailable(float64(v.Available))
	exporter.SetMemoryUsedPct(v.UsedPercent)

	// Swap metrics
	if c.cfg.IncludeSwap {
		s, err := mem.SwapMemory()
		if err != nil {
			// Swap may not be available on all systems; not fatal
			exporter.SetSwapTotal(0)
			exporter.SetSwapUsed(0)
			exporter.SetSwapUsedPct(0)
		} else {
			exporter.SetSwapTotal(float64(s.Total))
			exporter.SetSwapUsed(float64(s.Used))
			exporter.SetSwapUsedPct(s.UsedPercent)
		}
	}

	c.healthy = true
	return nil
}

// Healthy returns nil if the last collection was successful.
func (c *Collector) Healthy() error {
	if !c.healthy {
		return fmt.Errorf("memory collector has not completed a successful collection")
	}
	return nil
}
