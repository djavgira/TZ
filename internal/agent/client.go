package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	pain_tzv1 "github.com/Alice/pain_tz/pkg/proto/pain_tz/v1"
)

// Client manages a gRPC connection to the server with auto-reconnect.
type Client struct {
	addr      string
	logger    *logrus.Logger
	backoff   time.Duration
	maxBackoff time.Duration
	insecure  bool

	conn   *grpc.ClientConn
	stream pain_tzv1.MetricsService_StreamMetricsClient
}

// NewClient creates a gRPC client with the given configuration.
func NewClient(addr string, insecure bool, maxBackoff time.Duration, logger *logrus.Logger) *Client {
	return &Client{
		addr:       addr,
		logger:     logger,
		backoff:    1 * time.Second,
		maxBackoff: maxBackoff,
		insecure:   insecure,
	}
}

// Connect dials the server and opens a streaming RPC. Retries with
// exponential backoff until the context is cancelled.
func (c *Client) Connect(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		opts := []grpc.DialOption{
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:    10 * time.Second,
				Timeout: 5 * time.Second,
			}),
		}

		if c.insecure {
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}

		conn, err := grpc.DialContext(ctx, c.addr, opts...)
		if err != nil {
			c.logger.WithError(err).WithField("backoff", c.backoff).Warn("gRPC dial failed, retrying")
			c.sleep(ctx)
			continue
		}

		c.conn = conn
		c.backoff = 1 * time.Second // reset on success
		break
	}

	// Open the stream
	client := pain_tzv1.NewMetricsServiceClient(c.conn)
	stream, err := client.StreamMetrics(ctx)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("StreamMetrics: %w", err)
	}
	c.stream = stream

	c.logger.WithField("addr", c.addr).Info("connected to gRPC server")
	return nil
}

// Send pushes a MetricReport on the stream. Returns an error if the stream
// is broken (caller should reconnect).
func (c *Client) Send(report *pain_tzv1.MetricReport) error {
	if c.stream == nil {
		return fmt.Errorf("not connected")
	}
	return c.stream.Send(report)
}

// CloseAndRecv gracefully closes the send direction and waits for the server
// acknowledgement.
func (c *Client) CloseAndRecv() error {
	if c.stream == nil {
		return nil
	}
	_, err := c.stream.CloseAndRecv()
	return err
}

// Close tears down the gRPC connection.
func (c *Client) Close() {
	if c.stream != nil {
		c.CloseAndRecv()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) sleep(ctx context.Context) {
	timer := time.NewTimer(c.backoff)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}

	c.backoff *= 2
	if c.backoff > c.maxBackoff {
		c.backoff = c.maxBackoff
	}
}
