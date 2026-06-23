package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/djavgira/TZ/internal/agent"
	"github.com/djavgira/TZ/internal/collector"
	cpucol "github.com/djavgira/TZ/internal/collector/cpu"
	diskcol "github.com/djavgira/TZ/internal/collector/disk"
	memcol "github.com/djavgira/TZ/internal/collector/memory"
	netcol "github.com/djavgira/TZ/internal/collector/network"
	"github.com/djavgira/TZ/internal/config"
)

func init() {
	agentCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Path to configuration file (YAML)")
	rootCmd.AddCommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start in agent mode (collect metrics + push to server via gRPC)",
	Long: `Agent mode collects CPU, memory, disk, and network metrics and pushes
them to a central tz server via gRPC streaming.

The agent connects to the server address specified in the config and
auto-reconnects with exponential backoff on connection loss.`,
	RunE: runAgent,
}

func runAgent(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger := setupLogger(cfg)
	logger.WithFields(logrus.Fields{
		"host_id":     cfg.Agent.HostID,
		"server_addr": cfg.GRPCClient.ServerAddr,
	}).Info("starting tz agent")

	// Build collector registry (agent mode: no async goroutines, we collect synchronously)
	reg := collector.NewRegistry(nil) // nil handler — agent mode uses CollectAll

	if cfg.Collectors.CPU.Enabled {
		reg.Register(cpucol.New(cfg.Collectors.CPU), cfg.Collectors.CPU.Interval)
	}
	if cfg.Collectors.Memory.Enabled {
		reg.Register(memcol.New(cfg.Collectors.Memory), cfg.Collectors.Memory.Interval)
	}
	if cfg.Collectors.Disk.Enabled {
		reg.Register(diskcol.New(cfg.Collectors.Disk), cfg.Collectors.Disk.Interval)
	}
	if cfg.Collectors.Network.Enabled {
		reg.Register(netcol.New(cfg.Collectors.Network), cfg.Collectors.Network.Interval)
	}

	logger.WithField("collectors", reg.Count()).Info("collectors registered")

	// gRPC client
	client := agent.NewClient(
		cfg.GRPCClient.ServerAddr,
		cfg.GRPCClient.Insecure,
		cfg.GRPCClient.ReconnectBackoffMax,
		logger,
	)

	// Streamer
	streamer := agent.NewStreamer(logger, cfg.Agent.HostID, reg, client)

	// Run with signal-aware context
	ctx, cancel := signalContext()
	defer cancel()

	if err := streamer.Run(ctx, cfg.GRPCClient.PushInterval); err != nil {
		return fmt.Errorf("streamer error: %w", err)
	}

	logger.Info("agent shut down")
	return nil
}
