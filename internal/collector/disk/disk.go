package disk

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/disk"

	"github.com/Alice/pain_tz/internal/config"
	"github.com/Alice/pain_tz/internal/exporter"
)

// Collector collects disk metrics using gopsutil.
type Collector struct {
	cfg     config.DiskConfig
	healthy bool
}

// New creates a new disk collector.
func New(cfg config.DiskConfig) *Collector {
	return &Collector{cfg: cfg}
}

// Name returns the collector name.
func (c *Collector) Name() string { return "disk" }

// RegisterMetrics registers disk-related Prometheus metrics.
func (c *Collector) RegisterMetrics(reg *prometheus.Registry) error {
	return exporter.RegisterAll(reg)
}

// Collect gathers disk usage and optional IO metrics.
func (c *Collector) Collect(ctx context.Context) error {
	partitions, err := disk.Partitions(false)
	if err != nil {
		c.healthy = false
		return fmt.Errorf("disk.Partitions: %w", err)
	}

	denySet := toSet(c.cfg.FSTypeDenylist)
	includeSet := toSet(c.cfg.MountPoints)

	for _, p := range partitions {
		// Skip denied filesystem types
		if denySet[p.Fstype] {
			continue
		}

		// If mount_points is specified, only include those
		if len(includeSet) > 0 && !includeSet[p.Mountpoint] {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			// Some mount points may be inaccessible; skip
			continue
		}

		exporter.SetDiskTotal(p.Mountpoint, p.Device, p.Fstype, float64(usage.Total))
		exporter.SetDiskUsed(p.Mountpoint, p.Device, p.Fstype, float64(usage.Used))
		exporter.SetDiskFree(p.Mountpoint, p.Device, p.Fstype, float64(usage.Free))
		exporter.SetDiskUsedPct(p.Mountpoint, p.Device, p.Fstype, usage.UsedPercent)
	}

	// IO counters (cumulative; use Add to update counters)
	if c.cfg.IncludeIO {
		ioCounters, err := disk.IOCounters()
		if err == nil {
			for name, io := range ioCounters {
				exporter.AddDiskReadBytes(name, float64(io.ReadBytes))
				exporter.AddDiskWriteBytes(name, float64(io.WriteBytes))
			}
		}
	}

	c.healthy = true
	return nil
}

// Healthy returns nil if the last collection was successful.
func (c *Collector) Healthy() error {
	if !c.healthy {
		return fmt.Errorf("disk collector has not completed a successful collection")
	}
	return nil
}

func toSet(slice []string) map[string]bool {
	set := make(map[string]bool, len(slice))
	for _, s := range slice {
		set[s] = true
	}
	return set
}
