package network

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/net"

	"github.com/Alice/pain_tz/internal/config"
	"github.com/Alice/pain_tz/internal/exporter"
)

// Collector collects network metrics using gopsutil.
type Collector struct {
	cfg     config.NetworkConfig
	healthy bool

	// Track previous values for delta calculation on cumulative counters.
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

// Collect gathers per-interface network counters.
func (c *Collector) Collect(ctx context.Context) error {
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

		// Calculate deltas from previous values (counters are cumulative)
		prevSentBytes := c.prevSent[nic.Name]
		prevRecvBytes := c.prevRecv[nic.Name]

		sentDelta := float64(nic.BytesSent - prevSentBytes)
		recvDelta := float64(nic.BytesRecv - prevRecvBytes)

		// For the first collection, deltas will be the full counter value.
		// This is acceptable — subsequent collections will show true deltas.

		exporter.AddNetBytesSent(nic.Name, sentDelta)
		exporter.AddNetBytesRecv(nic.Name, recvDelta)

		// Packets
		exporter.AddNetPacketsSent(nic.Name, float64(nic.PacketsSent))
		exporter.AddNetPacketsRecv(nic.Name, float64(nic.PacketsRecv))

		// Errors and drops
		exporter.AddNetErrorsSent(nic.Name, float64(nic.Errin))
		exporter.AddNetErrorsRecv(nic.Name, float64(nic.Errout))
		exporter.AddNetDropsSent(nic.Name, float64(nic.Dropin))
		exporter.AddNetDropsRecv(nic.Name, float64(nic.Dropout))

		// Update previous values
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
