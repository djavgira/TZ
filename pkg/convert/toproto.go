package convert

import (
	"time"

	"github.com/djavgira/TZ/pkg/metrics"
	"github.com/djavgira/TZ/pkg/proto/tz/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SnapshotToProto converts a metrics.Snapshot into a protobuf MetricReport.
func SnapshotToProto(hostID string, snap *metrics.Snapshot) *tzv1.MetricReport {
	ts := snap.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	report := &tzv1.MetricReport{
		HostId:    hostID,
		Timestamp: timestamppb.New(ts),
		Cpu: &tzv1.CPU{
			UsagePercent:  snap.CPU.UsagePercent,
			UserPercent:   snap.CPU.UserPercent,
			SystemPercent: snap.CPU.SystemPercent,
			IdlePercent:   snap.CPU.IdlePercent,
			IowaitPercent: snap.CPU.IOWaitPercent,
			LogicalCount:  snap.CPU.LogicalCount,
		},
		Memory: &tzv1.Memory{
			TotalBytes:      snap.Memory.TotalBytes,
			UsedBytes:       snap.Memory.UsedBytes,
			AvailableBytes:  snap.Memory.AvailableBytes,
			UsedPercent:     snap.Memory.UsedPercent,
			SwapTotalBytes:  snap.Memory.SwapTotalBytes,
			SwapUsedBytes:   snap.Memory.SwapUsedBytes,
			SwapUsedPercent: snap.Memory.SwapUsedPercent,
		},
	}

	for _, d := range snap.Disks {
		report.Disks = append(report.Disks, &tzv1.Disk{
			Mountpoint:  d.Mountpoint,
			Device:      d.Device,
			Fstype:      d.FSType,
			TotalBytes:  d.TotalBytes,
			UsedBytes:   d.UsedBytes,
			FreeBytes:   d.FreeBytes,
			UsedPercent: d.UsedPercent,
			ReadBytes:   d.ReadBytes,
			WriteBytes:  d.WriteBytes,
		})
	}

	for _, n := range snap.Networks {
		report.Networks = append(report.Networks, &tzv1.Network{
			InterfaceName: n.Interface,
			BytesSent:     n.BytesSent,
			BytesRecv:     n.BytesRecv,
			PacketsSent:   n.PacketsSent,
			PacketsRecv:   n.PacketsRecv,
			ErrorsSent:    n.ErrorsSent,
			ErrorsRecv:    n.ErrorsRecv,
			DropsSent:     n.DropsSent,
			DropsRecv:     n.DropsRecv,
		})
	}

	return report
}
