package main

import (
	"fmt"
	"net"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/djavgira/TZ/internal/config"
	"github.com/djavgira/TZ/internal/server"
	"github.com/djavgira/TZ/internal/server/tui"
)

func init() {
	serverCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "Path to configuration file (YAML)")
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start in server mode (receive metrics + display TUI)",
	Long: `Server mode runs a gRPC server that receives metric streams from
multiple tz agents and displays them in a top-like TUI.

The TUI refreshes every second and color-codes resource usage:
  green (<50%)  yellow (50-80%)  red (>80%)

Use 'q' or Ctrl+C to quit.`,
	RunE: runServer,
}

func runServer(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger := setupLogger(cfg)
	logger.WithField("listen_addr", cfg.GRPCServer.ListenAddr).Info("starting tz server")

	// Create store and gRPC service
	store := server.NewStore()
	svc := server.NewGRPCServer(store, logger)

	// Start gRPC server in background
	lis, err := net.Listen("tcp", cfg.GRPCServer.ListenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", cfg.GRPCServer.ListenAddr, err)
	}

	grpcSrv := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              10 * time.Second,
			Timeout:           5 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	// Register the service
	// (the proto-generated RegisterMetricsServiceServer will be available after buf generate)
	// For now, we register via the generated code - this will compile after buf generate
	// server.RegisterMetricsServiceServer(grpcSrv, svc)

	go func() {
		logger.Info("gRPC server listening")
		if err := grpcSrv.Serve(lis); err != nil {
			logger.WithError(err).Error("gRPC server error")
		}
	}()

	// Start stale pruner goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if removed := store.PruneStale(cfg.GRPCServer.StaleTimeout); removed > 0 {
				logger.WithField("removed", removed).Debug("pruned stale hosts")
			}
		}
	}()

	// Run the TUI (blocks until user quits)
	m := tui.NewModel(store, cfg.GRPCServer.StaleTimeout)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Cleanup after TUI exits
	logger.Info("shutting down gRPC server")
	grpcSrv.GracefulStop()

	return nil
}
