package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the top-level configuration for the pain_tz agent.
type Config struct {
	Agent      AgentConfig      `mapstructure:"agent"`
	Collectors CollectorsConfig `mapstructure:"collectors"`
	Server     ServerConfig     `mapstructure:"server"`
	GRPCClient GRPCClientConfig `mapstructure:"grpc_client"`
	GRPCServer GRPCServerConfig `mapstructure:"grpc_server"`
	Logging    LoggingConfig    `mapstructure:"logging"`
}

// AgentConfig holds agent-level settings.
type AgentConfig struct {
	HostID string `mapstructure:"host_id"`
}

// CollectorsConfig groups all collector-specific configurations.
type CollectorsConfig struct {
	CPU     CPUConfig     `mapstructure:"cpu"`
	Memory  MemoryConfig  `mapstructure:"memory"`
	Disk    DiskConfig    `mapstructure:"disk"`
	Network NetworkConfig `mapstructure:"network"`
}

// CPUConfig holds configuration for the CPU collector.
type CPUConfig struct {
	Enabled  bool          `mapstructure:"enabled"`
	Interval time.Duration `mapstructure:"interval"`
	PerCore  bool          `mapstructure:"per_core"`
}

// MemoryConfig holds configuration for the memory collector.
type MemoryConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	Interval    time.Duration `mapstructure:"interval"`
	IncludeSwap bool          `mapstructure:"include_swap"`
}

// DiskConfig holds configuration for the disk collector.
type DiskConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	Interval       time.Duration `mapstructure:"interval"`
	MountPoints    []string      `mapstructure:"mount_points"`
	IncludeIO      bool          `mapstructure:"include_io"`
	FSTypeDenylist []string      `mapstructure:"fs_type_denylist"`
}

// NetworkConfig holds configuration for the network collector.
type NetworkConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	Interval          time.Duration `mapstructure:"interval"`
	Interfaces        []string      `mapstructure:"interfaces"`
	InterfaceDenylist []string      `mapstructure:"interface_denylist"`
}

// ServerConfig holds configuration for the Prometheus HTTP server (serve mode).
type ServerConfig struct {
	ListenAddr      string        `mapstructure:"listen_addr"`
	MetricsPath     string        `mapstructure:"metrics_path"`
	HealthPath      string        `mapstructure:"health_path"`
	ReadinessPath   string        `mapstructure:"readiness_path"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	MaxHeaderBytes  int           `mapstructure:"max_header_bytes"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
}

// GRPCClientConfig holds configuration for the agent-mode gRPC client.
type GRPCClientConfig struct {
	// ServerAddr is the gRPC server address (e.g., "server.laptop:9090").
	ServerAddr string `mapstructure:"server_addr"`
	// PushInterval is how often metrics are pushed to the server.
	PushInterval time.Duration `mapstructure:"push_interval"`
	// ReconnectBackoffMax is the maximum backoff duration between reconnection attempts.
	ReconnectBackoffMax time.Duration `mapstructure:"reconnect_backoff_max"`
	// Insecure disables TLS (for LAN/private network use).
	Insecure bool `mapstructure:"insecure"`
}

// GRPCServerConfig holds configuration for the server-mode gRPC server.
type GRPCServerConfig struct {
	// ListenAddr is the gRPC server listen address.
	ListenAddr string `mapstructure:"listen_addr"`
	// StaleTimeout is how long without data before a host is considered stale in the TUI.
	StaleTimeout time.Duration `mapstructure:"stale_timeout"`
}

// LoggingConfig holds configuration for logging.
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
	MaxBackups int    `mapstructure:"max_backups"`
}

// DefaultConfig returns a Config with sensible defaults for all modes.
func DefaultConfig() *Config {
	return &Config{
		Agent: AgentConfig{
			HostID: "",
		},
		Collectors: CollectorsConfig{
			CPU: CPUConfig{
				Enabled:  true,
				Interval: 10 * time.Second,
				PerCore:  false,
			},
			Memory: MemoryConfig{
				Enabled:     true,
				Interval:    15 * time.Second,
				IncludeSwap: true,
			},
			Disk: DiskConfig{
				Enabled:     true,
				Interval:    30 * time.Second,
				MountPoints: []string{},
				IncludeIO:   true,
				FSTypeDenylist: []string{
					"tmpfs", "devtmpfs", "squashfs", "overlay",
				},
			},
			Network: NetworkConfig{
				Enabled:           true,
				Interval:          15 * time.Second,
				Interfaces:        []string{},
				InterfaceDenylist: []string{"lo"},
			},
		},
		Server: ServerConfig{
			ListenAddr:      ":9100",
			MetricsPath:     "/metrics",
			HealthPath:      "/health",
			ReadinessPath:   "/ready",
			ShutdownTimeout: 10 * time.Second,
			MaxHeaderBytes:  1 << 20, // 1 MiB
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     120 * time.Second,
		},
		GRPCClient: GRPCClientConfig{
			ServerAddr:          "localhost:9090",
			PushInterval:        5 * time.Second,
			ReconnectBackoffMax: 30 * time.Second,
			Insecure:            true,
		},
		GRPCServer: GRPCServerConfig{
			ListenAddr:   ":9090",
			StaleTimeout: 30 * time.Second,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			FilePath:   "/var/log/pain_tz/agent.log",
			MaxSizeMB:  100,
			MaxAgeDays: 30,
			MaxBackups: 10,
		},
	}
}

// Load reads configuration from the specified path, falling back to defaults
// and environment variables (PAIN_TZ_ prefix).
func Load(path string) (*Config, error) {
	v := viper.New()

	// Set defaults
	cfg := DefaultConfig()
	v.SetDefault("agent", cfg.Agent)
	v.SetDefault("collectors", cfg.Collectors)
	v.SetDefault("server", cfg.Server)
	v.SetDefault("grpc_client", cfg.GRPCClient)
	v.SetDefault("grpc_server", cfg.GRPCServer)
	v.SetDefault("logging", cfg.Logging)

	// Read from config file if provided
	if path != "" {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
			}
		}
	}

	// Environment variable overrides
	v.SetEnvPrefix("PAIN_TZ")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	result := &Config{}
	if err := v.Unmarshal(result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Fill host_id from system hostname if not set
	if result.Agent.HostID == "" {
		hostname, err := os.Hostname()
		if err == nil {
			result.Agent.HostID = hostname
		}
	}

	return result, nil
}
