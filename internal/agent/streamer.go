package agent

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/djavgira/TZ/internal/collector"
	"github.com/djavgira/TZ/pkg/convert"
)

// Streamer collects metrics synchronously and pushes them via gRPC.
// It polls all registered collectors at the configured push interval,
// converts the merged Snapshot to a protobuf MetricReport, and sends it.
type Streamer struct {
	logger *logrus.Logger
	hostID string
	reg    *collector.Registry
	client *Client
}

// NewStreamer creates a Streamer.
func NewStreamer(logger *logrus.Logger, hostID string, reg *collector.Registry, client *Client) *Streamer {
	return &Streamer{
		logger: logger,
		hostID: hostID,
		reg:    reg,
		client: client,
	}
}

// Run connects to the server and then pushes metrics at each tick.
// On send failure, it triggers a reconnect. Blocks until ctx is cancelled.
func (s *Streamer) Run(ctx context.Context, interval time.Duration) error {
	// Initial connect
	if err := s.client.Connect(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.client.CloseAndRecv()
			s.client.Close()
			return ctx.Err()
		case <-ticker.C:
			if err := s.push(ctx); err != nil {
				s.logger.WithError(err).Warn("push failed, reconnecting")
				s.client.Close()

				// Reconnect with backoff
				if err := s.client.Connect(ctx); err != nil {
					return err
				}
			}
		}
	}
}

// push collects all metrics and sends them to the server.
func (s *Streamer) push(ctx context.Context) error {
	snap, err := s.reg.CollectAll(ctx)
	if err != nil {
		return err
	}
	snap.Timestamp = time.Now()

	report := convert.SnapshotToProto(s.hostID, snap)
	return s.client.Send(report)
}
