package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/Alice/pain_tz/internal/collector"
	cpucol "github.com/Alice/pain_tz/internal/collector/cpu"
	diskcol "github.com/Alice/pain_tz/internal/collector/disk"
	memcol "github.com/Alice/pain_tz/internal/collector/memory"
	netcol "github.com/Alice/pain_tz/internal/collector/network"
	"github.com/Alice/pain_tz/internal/config"
	"github.com/Alice/pain_tz/internal/server"
)

// Agent is the top-level orchestrator.
type Agent struct {
	cfg    *config.Config
	logger *logrus.Logger

	reg    *collector.Registry
	promReg *prometheus.Registry
	srv    *server.Server
}

// New creates an Agent from configuration and logger.
func New(cfg *config.Config, logger *logrus.Logger) (*Agent, error) {
	a := &Agent{
		cfg:     cfg,
		logger:  logger,
		reg:     collector.NewRegistry(),
		promReg: prometheus.NewRegistry(),
	}

	// Register enabled collectors
	if cfg.Collectors.CPU.Enabled {
		c := cpucol.New(cfg.Collectors.CPU)
		if err := a.registerCollector(c, cfg.Collectors.CPU.Interval); err != nil {
			return nil, err
		}
	}

	if cfg.Collectors.Memory.Enabled {
		c := memcol.New(cfg.Collectors.Memory)
		if err := a.registerCollector(c, cfg.Collectors.Memory.Interval); err != nil {
			return nil, err
		}
	}

	if cfg.Collectors.Disk.Enabled {
		c := diskcol.New(cfg.Collectors.Disk)
		if err := a.registerCollector(c, cfg.Collectors.Disk.Interval); err != nil {
			return nil, err
		}
	}

	if cfg.Collectors.Network.Enabled {
		c := netcol.New(cfg.Collectors.Network)
		if err := a.registerCollector(c, cfg.Collectors.Network.Interval); err != nil {
			return nil, err
		}
	}

	logger.WithField("collectors", a.reg.Count()).Info("collectors registered")

	// Create HTTP server
	a.srv = server.New(cfg.Server, logger, a.promReg, a, a)

	return a, nil
}

// registerCollector registers a single collector and its metrics.
func (a *Agent) registerCollector(c collector.Collector, interval time.Duration) error {
	if err := a.reg.Register(c, interval); err != nil {
		return fmt.Errorf("register %s: %w", c.Name(), err)
	}

	if mc, ok := c.(collector.MetricsCollector); ok {
		if err := mc.RegisterMetrics(a.promReg); err != nil {
			return fmt.Errorf("register metrics %s: %w", c.Name(), err)
		}
	}

	return nil
}

// Run starts all collectors and the HTTP server. Blocks until ctx is cancelled.
func (a *Agent) Run(ctx context.Context) error {
	// Start all collectors
	a.reg.StartAll(ctx)

	// Start HTTP server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := a.srv.Start(); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	// Wait for context cancellation (SIGTERM/SIGINT) or server error
	select {
	case <-ctx.Done():
		a.logger.Info("context cancelled, initiating shutdown")
	case err := <-errCh:
		if err != nil {
			a.logger.WithError(err).Error("server error")
			return err
		}
	}

	// Graceful shutdown
	return a.Shutdown()
}

// Shutdown performs a graceful shutdown.
func (a *Agent) Shutdown() error {
	// Stop all collectors
	a.logger.Info("stopping collectors")
	a.reg.StopAll()

	// Shutdown HTTP server with timeout
	timeout := a.cfg.Server.ShutdownTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	if err := a.srv.Shutdown(timeout); err != nil {
		a.logger.WithError(err).Error("server shutdown error")
		return err
	}

	a.logger.Info("agent shutdown complete")
	return nil
}

// IsReady implements server.ReadyChecker.
func (a *Agent) IsReady() bool {
	return a.reg.AllReady()
}

// HealthStatuses implements server.HealthChecker.
func (a *Agent) HealthStatuses() map[string]error {
	return a.reg.Statuses()
}
