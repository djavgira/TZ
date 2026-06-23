// Package metrics defines pure-Go data structures for system metrics.
// These types have zero external dependencies and serve as the central
// contract shared by collectors, converters, and the server store.
package metrics

import "time"

// Snapshot represents a single complete collection cycle from a host.
type Snapshot struct {
	Timestamp time.Time
	CPU       CPUMetrics
	Memory    MemoryMetrics
	Disks     []DiskMetrics
	Networks  []NetworkMetrics
}

// CPUMetrics holds CPU-related metrics.
type CPUMetrics struct {
	UsagePercent  float64
	UserPercent   float64
	SystemPercent float64
	IdlePercent   float64
	IOWaitPercent float64
	LogicalCount  uint32
}

// MemoryMetrics holds physical memory and swap metrics.
type MemoryMetrics struct {
	TotalBytes      uint64
	UsedBytes       uint64
	AvailableBytes  uint64
	UsedPercent     float64
	SwapTotalBytes  uint64
	SwapUsedBytes   uint64
	SwapUsedPercent float64
}

// DiskMetrics holds per-mountpoint disk usage and IO counters.
type DiskMetrics struct {
	Mountpoint  string
	Device      string
	FSType      string
	TotalBytes  uint64
	UsedBytes   uint64
	FreeBytes   uint64
	UsedPercent float64
	ReadBytes   uint64
	WriteBytes  uint64
}

// NetworkMetrics holds per-interface network counters.
type NetworkMetrics struct {
	Interface   string
	BytesSent   uint64
	BytesRecv   uint64
	PacketsSent uint64
	PacketsRecv uint64
	ErrorsSent  uint64
	ErrorsRecv  uint64
	DropsSent   uint64
	DropsRecv   uint64
}
