package cpu

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"

	"github.com/Alice/pain_tz/internal/config"
	"github.com/Alice/pain_tz/internal/exporter"
)

// Collector collects CPU metrics using gopsutil.
type Collector struct {
	cfg     config.CPUConfig
	healthy bool
}

// New creates a new CPU collector.
func New(cfg config.CPUConfig) *Collector {
	return &Collector{
		cfg: cfg,
	}
}

// Name returns the collector name.
func (c *Collector) Name() string { return "cpu" }

// RegisterMetrics registers CPU-related Prometheus metrics.
func (c *Collector) RegisterMetrics(reg *prometheus.Registry) error {
	return exporter.RegisterAll(reg)
}

// Collect gathers CPU metrics. For CPU, this call blocks for the configured
// interval duration because cpu.Percent() needs two samples.
func (c *Collector) Collect(ctx context.Context) error {
	// Get logical CPU count (static — only collected once)
	logicalCount, err := cpu.Counts(true)
	if err != nil {
		c.healthy = false
		return fmt.Errorf("cpu.Counts: %w", err)
	}
	exporter.SetCPUCount(float64(logicalCount))

	// cpu.Percent blocks for the given interval to compute the delta
	interval := c.cfg.Interval
	if interval <= 0 {
		interval = 1 * time.Second
	}

	perCPU := c.cfg.PerCore
	percentages, err := cpu.Percent(interval, perCPU)
	if err != nil {
		c.healthy = false
		return fmt.Errorf("cpu.Percent: %w", err)
	}

	if len(percentages) > 0 {
		exporter.SetCPUUsage(percentages[0])
	}

	// Get detailed time breakdown (non-blocking)
	times, err := cpu.Times(false)
	if err != nil {
		c.healthy = false
		return fmt.Errorf("cpu.Times: %w", err)
	}

	if len(times) > 0 {
		total := times[0].User + times[0].System + times[0].Idle + times[0].Iowait +
			times[0].Irq + times[0].Softirq + times[0].Steal
		if total > 0 {
			exporter.SetCPUUser((times[0].User / total) * 100)
			exporter.SetCPUSystem((times[0].System / total) * 100)
			exporter.SetCPUIdle((times[0].Idle / total) * 100)
			exporter.SetCPUIOWait((times[0].Iowait / total) * 100)
		}
	}

	c.healthy = true
	return nil
}

// Healthy returns nil if the last collection was successful.
func (c *Collector) Healthy() error {
	if !c.healthy {
		return fmt.Errorf("cpu collector has not completed a successful collection")
	}
	return nil
}
