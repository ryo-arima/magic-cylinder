package main

import (
	"flag"
	"log"

	"github.com/ryo-arima/magic-cylinder/internal"
	"github.com/ryo-arima/magic-cylinder/internal/config"
)

func main() {
	// Parse command-line arguments
	port := flag.String("port", "8443", "Server port")
	name := flag.String("name", "server", "Server name")
	targetURL := flag.String("target", "", "Target server URL for echo (e.g., https://localhost:8444/webtransport)")
	delay := flag.Int("delay", 0, "Delay seconds before echoing to target (0 for no delay)")
	flag.Parse()

	log.Printf("=== %s Starting ===", *name)
	log.Printf("[Main] Command-line arguments parsed:")
	log.Printf("[Main]   - Port: %s", *port)
	log.Printf("[Main]   - Name: %s", *name)
	log.Printf("[Main]   - Target URL: %s", *targetURL)
	log.Printf("[Main]   - Delay (s): %d", *delay)

	// Initialize configuration and dependencies
	log.Printf("[Main] Initializing server configuration...")
	cfg := config.NewServerConfig(*port, *name, *targetURL)
	log.Printf("[Main] Configuration created: Port=%s, Name=%s, Target=%s", cfg.Port, cfg.Name, cfg.TargetURL)

	log.Printf("[Main] Initializing dependencies...")
	router := internal.InitializeDependencies(*targetURL, *delay)

	log.Printf("[Main] Creating server instance...")
	server := internal.NewServer(cfg.Port, cfg.CertFile, cfg.KeyFile)

	// Start the server
	log.Printf("[Main] Starting server %s...", *name)
	if err := server.Start(router); err != nil {
		log.Fatalf("[Main] Server failed to start: %v", err)
	}
}
