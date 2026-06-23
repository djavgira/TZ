package convert

import (
	"time"

	"github.com/Alice/pain_tz/pkg/metrics"
	"github.com/Alice/pain_tz/pkg/proto/pain_tz/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SnapshotToProto converts a metrics.Snapshot into a protobuf MetricReport.
func SnapshotToProto(hostID string, snap *metrics.Snapshot) *pain_tzv1.MetricReport {
	ts := snap.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	report := &pain_tzv1.MetricReport{
		HostId:    hostID,
		Timestamp: timestamppb.New(ts),
		Cpu: &pain_tzv1.CPU{
			UsagePercent:  snap.CPU.UsagePercent,
			UserPercent:   snap.CPU.UserPercent,
			SystemPercent: snap.CPU.SystemPercent,
			IdlePercent:   snap.CPU.IdlePercent,
			IowaitPercent: snap.CPU.IOWaitPercent,
			LogicalCount:  snap.CPU.LogicalCount,
		},
		Memory: &pain_tzv1.Memory{
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
		report.Disks = append(report.Disks, &pain_tzv1.Disk{
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
		report.Networks = append(report.Networks, &pain_tzv1.Network{
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
