package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Alice/pain_tz/internal/collector"
	cpucol "github.com/Alice/pain_tz/internal/collector/cpu"
	diskcol "github.com/Alice/pain_tz/internal/collector/disk"
	memcol "github.com/Alice/pain_tz/internal/collector/memory"
	netcol "github.com/Alice/pain_tz/internal/collector/network"
	"github.com/Alice/pain_tz/internal/config"
	"github.com/Alice/pain_tz/internal/server"
	"github.com/Alice/pain_tz/pkg/convert"
	"github.com/Alice/pain_tz/pkg/metrics"
	"github.com/Alice/pain_tz/pkg/version"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "pain_tz",
	Short: "pain_tz — lightweight Linux monitoring agent",
	Long: `pain_tz is a lightweight Linux monitoring agent that collects
CPU, memory, disk, and network metrics.

Three modes:
  serve   Standalone Prometheus exporter (port 9100)
  agent   gRPC client — pushes metrics to a central server
  server  gRPC server + TUI — displays metrics from all agents`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Info())
		fmt.Println("\nAvailable modes: serve, agent, server")
		cmd.Help()
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Standalone Prometheus exporter mode",
	Long:  "Start the pain_tz agent in standalone mode, collecting metrics and exposing them via HTTP /metrics for Prometheus scraping.",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Path to configuration file (YAML)")
	rootCmd.AddCommand(serveCmd)
}

// --- Shared helpers ---

func setupLogger(cfg *config.Config) *logrus.Logger {
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	if cfg.Logging.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	if cfg.Logging.Output == "stdout" {
		logger.SetOutput(os.Stdout)
	} else {
		f, err := os.OpenFile(cfg.Logging.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.SetOutput(os.Stdout)
			logger.WithError(err).Warn("failed to open log file, falling back to stdout")
		} else {
			logger.SetOutput(f)
		}
	}
	return logger
}

func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		cancel()
	}()
	return ctx, cancel
}

// --- serve command ---

func runServe(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger := setupLogger(cfg)
	logger.WithFields(logrus.Fields{
		"version":    version.Version,
		"commit":     version.Commit,
		"build_date": version.BuildDate,
		"host_id":    cfg.Agent.HostID,
	}).Info("starting pain_tz (serve mode)")

	// serveModeHandler bridges collector snapshots to Prometheus exporters
	serveModeHandler := func(name string, snap *metrics.Snapshot) {
		convert.SnapshotToPrometheus(snap)
	}

	promReg := prometheus.NewRegistry()
	reg := collector.NewRegistry(serveModeHandler)

	// Register enabled collectors
	if cfg.Collectors.CPU.Enabled {
		c := cpucol.New(cfg.Collectors.CPU)
		if err := reg.Register(c, cfg.Collectors.CPU.Interval); err != nil {
			return fmt.Errorf("register cpu: %w", err)
		}
		if mc, ok := c.(collector.MetricsCollector); ok {
			mc.RegisterMetrics(promReg)
		}
	}

	if cfg.Collectors.Memory.Enabled {
		c := memcol.New(cfg.Collectors.Memory)
		if err := reg.Register(c, cfg.Collectors.Memory.Interval); err != nil {
			return fmt.Errorf("register memory: %w", err)
		}
		if mc, ok := c.(collector.MetricsCollector); ok {
			mc.RegisterMetrics(promReg)
		}
	}

	if cfg.Collectors.Disk.Enabled {
		c := diskcol.New(cfg.Collectors.Disk)
		if err := reg.Register(c, cfg.Collectors.Disk.Interval); err != nil {
			return fmt.Errorf("register disk: %w", err)
		}
		if mc, ok := c.(collector.MetricsCollector); ok {
			mc.RegisterMetrics(promReg)
		}
	}

	if cfg.Collectors.Network.Enabled {
		c := netcol.New(cfg.Collectors.Network)
		if err := reg.Register(c, cfg.Collectors.Network.Interval); err != nil {
			return fmt.Errorf("register network: %w", err)
		}
		if mc, ok := c.(collector.MetricsCollector); ok {
			mc.RegisterMetrics(promReg)
		}
	}

	logger.WithField("collectors", reg.Count()).Info("collectors registered")

	// Start async collectors
	ctx, cancel := signalContext()
	defer cancel()
	reg.StartAll(ctx)

	// Start HTTP server
	srv := server.New(cfg.Server, logger, promReg, reg, reg)
	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutting down (serve mode)")
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	reg.StopAll()
	if err := srv.Shutdown(cfg.Server.ShutdownTimeout); err != nil {
		logger.WithError(err).Error("server shutdown error")
	}
	logger.Info("agent shut down gracefully")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
