package delivery

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RouteConfigurator defines the interface for registering routes on a Gin router.
type RouteConfigurator interface {
	// RegisterRoutes registers all routes on the provided Gin router instance.
	RegisterRoutes(router *gin.Engine)
}

// MiddlewareConfigurator defines the interface for registering middleware on a Gin router.
type MiddlewareConfigurator interface {
	// RegisterMiddlewares registers all middleware on the provided Gin router instance.
	RegisterMiddlewares(router *gin.Engine)
}

// HTTPServer represents an HTTP server with TLS support and graceful shutdown capabilities.
type HTTPServer struct {
	// l is the structured logger for server operations.
	l *zap.SugaredLogger
	// server is the underlying HTTP server instance.
	server *http.Server
	// certFile is the path to the TLS certificate file.
	certFile string
	// keyFile is the path to the TLS private key file.
	keyFile string
	// startTimeout is the maximum time to wait for server startup.
	startTimeout time.Duration
	// stopTimeout is the maximum time to wait for graceful shutdown.
	stopTimeout time.Duration
	// tlsEnabled indicates whether TLS encryption is enabled.
	tlsEnabled bool
}

// NewHTTPServer creates a new HTTP server instance with the provided configuration.
func NewHTTPServer(
	logger *zap.SugaredLogger,
	rc RouteConfigurator,
	mc MiddlewareConfigurator,
	addr string,
	startTimeout time.Duration,
	stopTimeout time.Duration,
	tlsEnabled bool,
	certFile string,
	keyFile string,
) *HTTPServer {
	r := gin.New()
	mc.RegisterMiddlewares(r)
	rc.RegisterRoutes(r)

	s := &HTTPServer{
		l: logger,
		server: &http.Server{
			Addr:    addr,
			Handler: r,
		},
		startTimeout: startTimeout,
		stopTimeout:  stopTimeout,
		tlsEnabled:   tlsEnabled,
		certFile:     certFile,
		keyFile:      keyFile,
	}

	return s
}

// Start starts the HTTP server and returns an error if startup fails.
func (s *HTTPServer) Start(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		if err := s.listen(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	if err := s.startCheck(ctx, errChan); err != nil {
		s.l.Errorf("Failed to start HTTP server: %v", err)
		return fmt.Errorf("HTTP server start failed: %w", err)
	}

	s.l.Infof("%s server started successfully on %s", s.getProtocol(), s.server.Addr)
	return nil
}

// Stop gracefully shuts down the HTTP server with the configured timeout.
func (s *HTTPServer) Stop(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, s.stopTimeout)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.l.Errorf("Failed to gracefully shutdown HTTP server: %v", err)
		return fmt.Errorf("failed to gracefully shutdown HTTP server: %w", err)
	}

	return nil
}

// startCheck waits for server startup completion or timeout.
func (s *HTTPServer) startCheck(ctx context.Context, errChan chan error) error {
	select {
	case err := <-errChan:
		return fmt.Errorf("server startup error: %w", err)
	case <-time.After(s.startTimeout):
		return nil
	case <-ctx.Done():
		if err := s.server.Shutdown(context.Background()); err != nil {
			s.l.Errorf("Failed to shutdown server during start cancellation: %v", err)
			return fmt.Errorf("failed to shutdown HTTP server during start cancellation: %w", err)
		}
		return fmt.Errorf("start cancelled: %w", ctx.Err())
	}
}

// init sets Gin framework to release mode for production deployments.
func init() {
	gin.SetMode(gin.ReleaseMode)
}

// listen starts the HTTP or HTTPS listener based on TLS configuration.
func (s *HTTPServer) listen() error {
	if s.tlsEnabled {
		return s.listenHTTPS()
	}
	return s.listenHTTP()
}

// getProtocol returns the protocol string (HTTP or HTTPS) based on TLS configuration.
func (s *HTTPServer) getProtocol() string {
	if s.tlsEnabled {
		return "HTTPS"
	}
	return "HTTP"
}

// listenHTTP starts the HTTP server listener.
func (s *HTTPServer) listenHTTP() error {
	if err := s.server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	return nil
}

// listenHTTPS starts the HTTPS server listener with TLS certificates.
func (s *HTTPServer) listenHTTPS() error {
	if err := s.server.ListenAndServeTLS(s.certFile, s.keyFile); err != nil {
		return fmt.Errorf("failed to start HTTPS server: %w", err)
	}
	return nil
}
