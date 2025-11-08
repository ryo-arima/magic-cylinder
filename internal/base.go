package internal

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quic-go/webtransport-go"
)

// Server represents a WebTransport server instance
type Server struct {
	server     *webtransport.Server
	httpServer *http.Server
	port       string
	certFile   string
	keyFile    string
}

// NewServer creates a new WebTransport server
func NewServer(port, certFile, keyFile string) *Server {
	return &Server{
		port:     port,
		certFile: certFile,
		keyFile:  keyFile,
	}
}

// Start starts the WebTransport server with the given router
func (s *Server) Start(router *Router) error {
	log.Printf("[Server] Loading TLS certificates from %s and %s", s.certFile, s.keyFile)
	cert, err := tls.LoadX509KeyPair(s.certFile, s.keyFile)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %w", err)
	}
	log.Printf("[Server] TLS certificates loaded successfully")

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	log.Printf("[Server] Initializing WebTransport server")
	s.server = &webtransport.Server{}

	log.Printf("[Server] Setting up routes")
	router.SetupRoutes(s.server)

	s.httpServer = &http.Server{
		Addr:      ":" + s.port,
		TLSConfig: tlsConfig,
	}

	log.Printf("[Server] =====================================")
	log.Printf("[Server] Server starting on port %s", s.port)
	log.Printf("[Server] WebTransport endpoint: https://localhost:%s/webtransport", s.port)
	log.Printf("[Server] Health check endpoint: https://localhost:%s/health", s.port)
	log.Printf("[Server] =====================================")

	go func() {
		log.Printf("[Server] Starting HTTPS/HTTP3 listener on :%s", s.port)
		if err := s.httpServer.ListenAndServeTLS(s.certFile, s.keyFile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Server] Server failed to start: %v", err)
		}
	}()

	return s.waitForShutdown()
}

// waitForShutdown waits for interrupt signal and gracefully shuts down the server
func (s *Server) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("[Server] Shutdown signal received, initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("[Server] Forced shutdown due to error: %v", err)
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Printf("[Server] Server exited successfully")
	return nil
}
