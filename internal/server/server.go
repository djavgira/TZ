package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/Alice/pain_tz/internal/config"
)

// ReadyChecker is implemented by the agent to report readiness.
type ReadyChecker interface {
	IsReady() bool
}

// HealthChecker is implemented by the agent to report health.
type HealthChecker interface {
	HealthStatuses() map[string]error
}

// Server is the HTTP server exposing /metrics, /health, and /ready endpoints.
type Server struct {
	srv    *http.Server
	logger *logrus.Logger

	readyChecker  ReadyChecker
	healthChecker HealthChecker
}

// New creates a new HTTP server.
func New(
	cfg config.ServerConfig,
	logger *logrus.Logger,
	promReg *prometheus.Registry,
	readyChecker ReadyChecker,
	healthChecker HealthChecker,
) *Server {
	mux := http.NewServeMux()

	s := &Server{
		srv: &http.Server{
			Addr:           cfg.ListenAddr,
			Handler:        nil, // set below
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
			IdleTimeout:    cfg.IdleTimeout,
			MaxHeaderBytes: cfg.MaxHeaderBytes,
		},
		logger:        logger,
		readyChecker:  readyChecker,
		healthChecker: healthChecker,
	}

	// /metrics endpoint
	mux.Handle(cfg.MetricsPath, promhttp.HandlerFor(promReg, promhttp.HandlerOpts{
		ErrorLog:      logger.WithField("component", "promhttp"),
		EnableOpenMetrics: false,
	}))

	// /health endpoint
	mux.HandleFunc(cfg.HealthPath, s.handleHealth)

	// /ready endpoint
	mux.HandleFunc(cfg.ReadinessPath, s.handleReady)

	// Apply middleware
	handler := Chain(mux,
		RequestLogging(logger),
		PanicRecovery(logger),
	)
	s.srv.Handler = handler

	return s
}

// Start begins listening and serving. Returns an error if the server fails
// to start (e.g., port already bound).
func (s *Server) Start() error {
	s.logger.WithField("addr", s.srv.Addr).Info("starting http server")
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server.ListenAndServe: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the HTTP server with the configured timeout.
func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	s.logger.Info("shutting down http server")
	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server.Shutdown: %w", err)
	}
	return nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	statuses := s.healthChecker.HealthStatuses()

	healthy := len(statuses) == 0
	statusCode := http.StatusOK
	if !healthy {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := healthResponse{
		Status: map[bool]string{true: "healthy", false: "unhealthy"}[healthy],
	}
	if !healthy {
		resp.Checks = make(map[string]string, len(statuses))
		for name, err := range statuses {
			resp.Checks[name] = err.Error()
		}
	}

	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	ready := s.readyChecker.IsReady()
	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(readyResponse{
		Ready: ready,
	})
}

type healthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

type readyResponse struct {
	Ready bool `json:"ready"`
}
