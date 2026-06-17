package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Alice/pain_tz/internal/agent"
	"github.com/Alice/pain_tz/internal/config"
	"github.com/Alice/pain_tz/pkg/version"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "pain_tz",
		Short: "pain_tz — lightweight Linux monitoring agent",
		Long: `pain_tz is a lightweight Linux monitoring agent that collects
CPU, memory, disk, and network metrics and exposes them via Prometheus.

Deploy as a single static binary, managed by systemd.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Info())
			cmd.Help()
		},
	}

	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the monitoring agent",
		Long:  "Start the pain_tz agent, begin collecting metrics, and serve the HTTP /metrics endpoint.",
		RunE:  runServe,
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Path to configuration file (YAML)")
}

func runServe(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
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
			return fmt.Errorf("failed to open log file %s: %w", cfg.Logging.FilePath, err)
		}
		defer f.Close()
		logger.SetOutput(f)
	}

	logger.WithFields(logrus.Fields{
		"version":    version.Version,
		"commit":     version.Commit,
		"build_date": version.BuildDate,
		"host_id":    cfg.Agent.HostID,
	}).Info("starting pain_tz agent")

	// Create agent
	agt, err := agent.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		logger.WithField("signal", sig.String()).Info("received shutdown signal")
		cancel()
	}()

	// Run the agent (blocks until context is cancelled)
	if err := agt.Run(ctx); err != nil {
		return fmt.Errorf("agent error: %w", err)
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
