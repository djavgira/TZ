package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

// CPU metrics
var (
	cpuUsagePercent = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "cpu",
			Name:      "usage_percent",
			Help:      "Overall CPU usage as a percentage (0-100).",
		},
	)

	cpuUserPercent = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "cpu",
			Name:      "user_percent",
			Help:      "CPU time spent in user mode as a percentage.",
		},
	)

	cpuSystemPercent = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "cpu",
			Name:      "system_percent",
			Help:      "CPU time spent in system/kernel mode as a percentage.",
		},
	)

	cpuIdlePercent = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "cpu",
			Name:      "idle_percent",
			Help:      "CPU idle time as a percentage.",
		},
	)

	cpuIOWaitPercent = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "cpu",
			Name:      "iowait_percent",
			Help:      "CPU time spent waiting for I/O as a percentage.",
		},
	)

	cpuCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "cpu",
			Name:      "logical_count",
			Help:      "Number of logical CPU cores.",
		},
	)
)

// Memory metrics
var (
	memoryTotalBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "memory",
			Name:      "total_bytes",
			Help:      "Total physical memory in bytes.",
		},
	)

	memoryUsedBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "memory",
			Name:      "used_bytes",
			Help:      "Used physical memory in bytes.",
		},
	)

	memoryAvailableBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "memory",
			Name:      "available_bytes",
			Help:      "Available physical memory in bytes (includes reclaimable).",
		},
	)

	memoryUsedPercent = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "memory",
			Name:      "used_percent",
			Help:      "Physical memory usage as a percentage.",
		},
	)

	swapTotalBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "memory",
			Name:      "swap_total_bytes",
			Help:      "Total swap space in bytes.",
		},
	)

	swapUsedBytes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "memory",
			Name:      "swap_used_bytes",
			Help:      "Used swap space in bytes.",
		},
	)

	swapUsedPercent = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "memory",
			Name:      "swap_used_percent",
			Help:      "Swap usage as a percentage.",
		},
	)
)

// Disk metrics
var (
	diskTotalBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "disk",
			Name:      "total_bytes",
			Help:      "Total disk space in bytes per mount point.",
		},
		[]string{"mountpoint", "device", "fstype"},
	)

	diskUsedBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "disk",
			Name:      "used_bytes",
			Help:      "Used disk space in bytes per mount point.",
		},
		[]string{"mountpoint", "device", "fstype"},
	)

	diskFreeBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "disk",
			Name:      "free_bytes",
			Help:      "Free disk space in bytes per mount point.",
		},
		[]string{"mountpoint", "device", "fstype"},
	)

	diskUsedPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "pain_tz",
			Subsystem: "disk",
			Name:      "used_percent",
			Help:      "Disk usage as a percentage per mount point.",
		},
		[]string{"mountpoint", "device", "fstype"},
	)

	diskReadBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pain_tz",
			Subsystem: "disk",
			Name:      "read_bytes_total",
			Help:      "Total bytes read from disk.",
		},
		[]string{"device"},
	)

	diskWriteBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pain_tz",
			Subsystem: "disk",
			Name:      "write_bytes_total",
			Help:      "Total bytes written to disk.",
		},
		[]string{"device"},
	)
)

// Network metrics
var (
	netBytesSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pain_tz",
			Subsystem: "network",
			Name:      "bytes_sent_total",
			Help:      "Total bytes sent per network interface.",
		},
		[]string{"interface"},
	)

	netBytesRecv = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pain_tz",
			Subsystem: "network",
			Name:      "bytes_recv_total",
			Help:      "Total bytes received per network interface.",
		},
		[]string{"interface"},
	)

	netPacketsSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pain_tz",
			Subsystem: "network",
			Name:      "packets_sent_total",
			Help:      "Total packets sent per network interface.",
		},
		[]string{"interface"},
	)

	netPacketsRecv = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pain_tz",
			Subsystem: "network",
			Name:      "packets_recv_total",
			Help:      "Total packets received per network interface.",
		},
		[]string{"interface"},
	)

	netErrorsSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pain_tz",
			Subsystem: "network",
			Name:      "errors_sent_total",
			Help:      "Total transmit errors per network interface.",
		},
		[]string{"interface"},
	)

	netErrorsRecv = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pain_tz",
			Subsystem: "network",
			Name:      "errors_recv_total",
			Help:      "Total receive errors per network interface.",
		},
		[]string{"interface"},
	)

	netDropsSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pain_tz",
			Subsystem: "network",
			Name:      "drops_sent_total",
			Help:      "Total transmit drops per network interface.",
		},
		[]string{"interface"},
	)

	netDropsRecv = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "pain_tz",
			Subsystem: "network",
			Name:      "drops_recv_total",
			Help:      "Total receive drops per network interface.",
		},
		[]string{"interface"},
	)
)

// allMetrics is the complete list of all exported metric descriptors.
var allMetrics = []prometheus.Collector{
	// CPU
	cpuUsagePercent,
	cpuUserPercent,
	cpuSystemPercent,
	cpuIdlePercent,
	cpuIOWaitPercent,
	cpuCount,

	// Memory
	memoryTotalBytes,
	memoryUsedBytes,
	memoryAvailableBytes,
	memoryUsedPercent,
	swapTotalBytes,
	swapUsedBytes,
	swapUsedPercent,

	// Disk
	diskTotalBytes,
	diskUsedBytes,
	diskFreeBytes,
	diskUsedPercent,
	diskReadBytes,
	diskWriteBytes,

	// Network
	netBytesSent,
	netBytesRecv,
	netPacketsSent,
	netPacketsRecv,
	netErrorsSent,
	netErrorsRecv,
	netDropsSent,
	netDropsRecv,
}

// RegisterAll registers all metric descriptors with the given prometheus registry.
func RegisterAll(reg *prometheus.Registry) error {
	for _, m := range allMetrics {
		if err := reg.Register(m); err != nil {
			return err
		}
	}
	return nil
}

// --- CPU metric setters ---

func SetCPUUsage(v float64)          { cpuUsagePercent.Set(v) }
func SetCPUUser(v float64)           { cpuUserPercent.Set(v) }
func SetCPUSystem(v float64)         { cpuSystemPercent.Set(v) }
func SetCPUIdle(v float64)           { cpuIdlePercent.Set(v) }
func SetCPUIOWait(v float64)         { cpuIOWaitPercent.Set(v) }
func SetCPUCount(v float64)          { cpuCount.Set(v) }

// --- Memory metric setters ---

func SetMemoryTotal(v float64)      { memoryTotalBytes.Set(v) }
func SetMemoryUsed(v float64)       { memoryUsedBytes.Set(v) }
func SetMemoryAvailable(v float64)  { memoryAvailableBytes.Set(v) }
func SetMemoryUsedPct(v float64)    { memoryUsedPercent.Set(v) }
func SetSwapTotal(v float64)        { swapTotalBytes.Set(v) }
func SetSwapUsed(v float64)         { swapUsedBytes.Set(v) }
func SetSwapUsedPct(v float64)      { swapUsedPercent.Set(v) }

// --- Disk metric setters ---

func SetDiskTotal(mountpoint, device, fstype string, v float64) {
	diskTotalBytes.WithLabelValues(mountpoint, device, fstype).Set(v)
}
func SetDiskUsed(mountpoint, device, fstype string, v float64) {
	diskUsedBytes.WithLabelValues(mountpoint, device, fstype).Set(v)
}
func SetDiskFree(mountpoint, device, fstype string, v float64) {
	diskFreeBytes.WithLabelValues(mountpoint, device, fstype).Set(v)
}
func SetDiskUsedPct(mountpoint, device, fstype string, v float64) {
	diskUsedPercent.WithLabelValues(mountpoint, device, fstype).Set(v)
}
func AddDiskReadBytes(device string, v float64) {
	diskReadBytes.WithLabelValues(device).Add(v)
}
func AddDiskWriteBytes(device string, v float64) {
	diskWriteBytes.WithLabelValues(device).Add(v)
}

// --- Network metric setters ---

func AddNetBytesSent(iface string, v float64)   { netBytesSent.WithLabelValues(iface).Add(v) }
func AddNetBytesRecv(iface string, v float64)   { netBytesRecv.WithLabelValues(iface).Add(v) }
func AddNetPacketsSent(iface string, v float64) { netPacketsSent.WithLabelValues(iface).Add(v) }
func AddNetPacketsRecv(iface string, v float64) { netPacketsRecv.WithLabelValues(iface).Add(v) }
func AddNetErrorsSent(iface string, v float64)  { netErrorsSent.WithLabelValues(iface).Add(v) }
func AddNetErrorsRecv(iface string, v float64)  { netErrorsRecv.WithLabelValues(iface).Add(v) }
func AddNetDropsSent(iface string, v float64)   { netDropsSent.WithLabelValues(iface).Add(v) }
func AddNetDropsRecv(iface string, v float64)   { netDropsRecv.WithLabelValues(iface).Add(v) }
