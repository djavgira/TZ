package server

import (
	"io"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	tzv1 "github.com/djavgira/TZ/pkg/proto/tz/v1"
)

// GRPCServer implements the MetricsService gRPC server.
type GRPCServer struct {
	tzv1.UnimplementedMetricsServiceServer
	store  *Store
	logger *logrus.Logger
}

// NewGRPCServer creates a gRPC server backed by the given store.
func NewGRPCServer(store *Store, logger *logrus.Logger) *GRPCServer {
	return &GRPCServer{
		store:  store,
		logger: logger,
	}
}

// StreamMetrics receives a stream of MetricReports from a single agent.
// Each report is merged into the store. When the stream ends, an Empty
// response is returned.
func (s *GRPCServer) StreamMetrics(stream tzv1.MetricsService_StreamMetricsServer) error {
	var hostID string

	for {
		report, err := stream.Recv()
		if err == io.EOF {
			s.logger.WithField("host_id", hostID).Info("agent stream closed normally")
			return stream.SendAndClose(&emptypb.Empty{})
		}
		if err != nil {
			s.logger.WithError(err).WithField("host_id", hostID).Warn("agent stream error")
			return status.Errorf(codes.Internal, "stream error: %v", err)
		}

		hostID = report.HostId
		s.store.Update(report)

		s.logger.WithFields(logrus.Fields{
			"host_id":      hostID,
			"cpu_percent":  report.Cpu.UsagePercent,
			"mem_percent":  report.Memory.UsedPercent,
			"disk_count":   len(report.Disks),
			"iface_count":  len(report.Networks),
		}).Debug("received metrics")
	}
}

// Store returns the backing store (used by TUI and tests).
func (s *GRPCServer) Store() *Store {
	return s.store
}
