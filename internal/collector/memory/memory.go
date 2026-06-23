package memory

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/djavgira/TZ/internal/config"
	"github.com/djavgira/TZ/internal/exporter"
	"github.com/djavgira/TZ/pkg/metrics"
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

// Collect gathers memory metrics into snap.Memory.
func (c *Collector) Collect(ctx context.Context, snap *metrics.Snapshot) error {
	v, err := mem.VirtualMemory()
	if err != nil {
		c.healthy = false
		return fmt.Errorf("mem.VirtualMemory: %w", err)
	}

	snap.Memory.TotalBytes = v.Total
	snap.Memory.UsedBytes = v.Used
	snap.Memory.AvailableBytes = v.Available
	snap.Memory.UsedPercent = v.UsedPercent

	// Swap metrics
	if c.cfg.IncludeSwap {
		s, err := mem.SwapMemory()
		if err != nil {
			// Swap may not be available; set to zero
			snap.Memory.SwapTotalBytes = 0
			snap.Memory.SwapUsedBytes = 0
			snap.Memory.SwapUsedPercent = 0
		} else {
			snap.Memory.SwapTotalBytes = s.Total
			snap.Memory.SwapUsedBytes = s.Used
			snap.Memory.SwapUsedPercent = s.UsedPercent
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
