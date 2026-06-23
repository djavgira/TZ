package convert

import (
	"github.com/djavgira/TZ/internal/exporter"
	"github.com/djavgira/TZ/pkg/metrics"
)

// SnapshotToPrometheus converts a metrics.Snapshot into Prometheus gauge/counter
// updates via the exporter package. This adapter preserves backward compatibility
// with the "tz serve" (Prometheus exporter) mode.
// The function is safe to call concurrently.
func SnapshotToPrometheus(snap *metrics.Snapshot) {
	// CPU
	exporter.SetCPUUsage(snap.CPU.UsagePercent)
	exporter.SetCPUUser(snap.CPU.UserPercent)
	exporter.SetCPUSystem(snap.CPU.SystemPercent)
	exporter.SetCPUIdle(snap.CPU.IdlePercent)
	exporter.SetCPUIOWait(snap.CPU.IOWaitPercent)
	exporter.SetCPUCount(float64(snap.CPU.LogicalCount))

	// Memory
	exporter.SetMemoryTotal(float64(snap.Memory.TotalBytes))
	exporter.SetMemoryUsed(float64(snap.Memory.UsedBytes))
	exporter.SetMemoryAvailable(float64(snap.Memory.AvailableBytes))
	exporter.SetMemoryUsedPct(snap.Memory.UsedPercent)
	exporter.SetSwapTotal(float64(snap.Memory.SwapTotalBytes))
	exporter.SetSwapUsed(float64(snap.Memory.SwapUsedBytes))
	exporter.SetSwapUsedPct(snap.Memory.SwapUsedPercent)

	// Disk
	for _, d := range snap.Disks {
		exporter.SetDiskTotal(d.Mountpoint, d.Device, d.FSType, float64(d.TotalBytes))
		exporter.SetDiskUsed(d.Mountpoint, d.Device, d.FSType, float64(d.UsedBytes))
		exporter.SetDiskFree(d.Mountpoint, d.Device, d.FSType, float64(d.FreeBytes))
		exporter.SetDiskUsedPct(d.Mountpoint, d.Device, d.FSType, d.UsedPercent)
		exporter.AddDiskReadBytes(d.Device, float64(d.ReadBytes))
		exporter.AddDiskWriteBytes(d.Device, float64(d.WriteBytes))
	}

	// Network
	for _, n := range snap.Networks {
		exporter.AddNetBytesSent(n.Interface, float64(n.BytesSent))
		exporter.AddNetBytesRecv(n.Interface, float64(n.BytesRecv))
		exporter.AddNetPacketsSent(n.Interface, float64(n.PacketsSent))
		exporter.AddNetPacketsRecv(n.Interface, float64(n.PacketsRecv))
		exporter.AddNetErrorsSent(n.Interface, float64(n.ErrorsSent))
		exporter.AddNetErrorsRecv(n.Interface, float64(n.ErrorsRecv))
		exporter.AddNetDropsSent(n.Interface, float64(n.DropsSent))
		exporter.AddNetDropsRecv(n.Interface, float64(n.DropsRecv))
	}
}
