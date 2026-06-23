package network

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/net"

	"github.com/djavgira/TZ/internal/config"
	"github.com/djavgira/TZ/internal/exporter"
	"github.com/djavgira/TZ/pkg/metrics"
)

// Collector collects network metrics using gopsutil.
type Collector struct {
	cfg     config.NetworkConfig
	healthy bool

	// Track previous cumulative counter values for delta calculation.
	mu       sync.Mutex
	prevSent map[string]uint64
	prevRecv map[string]uint64
}

// New creates a new network collector.
func New(cfg config.NetworkConfig) *Collector {
	return &Collector{
		cfg:      cfg,
		prevSent: make(map[string]uint64),
		prevRecv: make(map[string]uint64),
	}
}

// Name returns the collector name.
func (c *Collector) Name() string { return "network" }

// RegisterMetrics registers network-related Prometheus metrics.
func (c *Collector) RegisterMetrics(reg *prometheus.Registry) error {
	return exporter.RegisterAll(reg)
}

// Collect gathers per-interface network counters into snap.Networks.
// Deltas are computed from the previous collection's cumulative values.
func (c *Collector) Collect(ctx context.Context, snap *metrics.Snapshot) error {
	counters, err := net.IOCounters(true) // per NIC
	if err != nil {
		c.healthy = false
		return fmt.Errorf("net.IOCounters: %w", err)
	}

	denySet := toSet(c.cfg.InterfaceDenylist)
	allowSet := toSet(c.cfg.Interfaces)

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, nic := range counters {
		// Apply filters
		if denySet[nic.Name] {
			continue
		}
		if len(allowSet) > 0 && !allowSet[nic.Name] {
			continue
		}

		// Calculate deltas from previous values
		prevSentBytes := c.prevSent[nic.Name]
		prevRecvBytes := c.prevRecv[nic.Name]

		snap.Networks = append(snap.Networks, metrics.NetworkMetrics{
			Interface:   nic.Name,
			BytesSent:   nic.BytesSent - prevSentBytes,
			BytesRecv:   nic.BytesRecv - prevRecvBytes,
			PacketsSent: nic.PacketsSent,
			PacketsRecv: nic.PacketsRecv,
			ErrorsSent:  nic.Errin,
			ErrorsRecv:  nic.Errout,
			DropsSent:   nic.Dropin,
			DropsRecv:   nic.Dropout,
		})

		// Update previous values for next delta calculation
		c.prevSent[nic.Name] = nic.BytesSent
		c.prevRecv[nic.Name] = nic.BytesRecv
	}

	c.healthy = true
	return nil
}

// Healthy returns nil if the last collection was successful.
func (c *Collector) Healthy() error {
	if !c.healthy {
		return fmt.Errorf("network collector has not completed a successful collection")
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
