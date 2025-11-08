package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/quic-go/webtransport-go"
	"github.com/ryo-arima/magic-cylinder/internal/entity/model"
)

func main() {
	// Parse command-line arguments
	serverURL := flag.String("server", "https://localhost:8443/webtransport", "Server URL to connect")
	flag.Parse()

	log.Printf("============================================")
	log.Printf("[Client] Starting WebTransport Client")
	log.Printf("[Client] Target server: %s", *serverURL)
	log.Printf("============================================")

	// Send initial ping to trigger the pingpong loop
	err := sendPing(*serverURL)
	if err != nil {
		log.Fatalf("[Client] ❌ Failed to send ping: %v", err)
	}

	log.Printf("============================================")
	log.Printf("[Client] ✅ Initial ping completed successfully")
	log.Printf("[Client] Server will continue pingpong loop")
	log.Printf("============================================")
}

// sendPing sends an initial ping message to the server
func sendPing(serverURL string) error {
	log.Printf("[Client] Parsing server URL: %s", serverURL)
	_, err := url.Parse(serverURL)
	if err != nil {
		log.Printf("[Client] ❌ Invalid server URL: %v", err)
		return fmt.Errorf("invalid server URL: %w", err)
	}
	log.Printf("[Client] ✅ Server URL parsed successfully")

	// Create WebTransport dialer with insecure skip verify for self-signed certs
	log.Printf("[Client] Creating WebTransport dialer with TLS InsecureSkipVerify")
	dialer := &webtransport.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// Dial the server
	log.Printf("[Client] Dialing server at %s...", serverURL)
	_, conn, err := dialer.Dial(context.Background(), serverURL, nil)
	if err != nil {
		log.Printf("[Client] ❌ Failed to dial server: %v", err)
		log.Printf("[Client]   Error type: %T", err)
		return fmt.Errorf("failed to dial server: %w", err)
	}
	defer conn.CloseWithError(0, "client disconnect")
	log.Printf("[Client] ✅ Connected to server successfully")

	// Open a stream
	log.Printf("[Client] Opening stream...")
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		log.Printf("[Client] ❌ Failed to open stream: %v", err)
		return fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()
	log.Printf("[Client] ✅ Stream opened: %d", stream.StreamID())

	log.Printf("[Client] ✅ Stream opened: %d", stream.StreamID())

	// Create and send ping message
	log.Printf("[Client] Creating ping message...")
	message := model.NewPingMessage(
		"Initial ping from client",
		1,
		"client",
		"server",
	)
	log.Printf("[Client] Message created: %s (seq: %d)", message.Content, message.Sequence)

	log.Printf("[Client] Marshalling message to JSON...")
	data, err := message.ToJSON()
	if err != nil {
		log.Printf("[Client] ❌ Failed to marshal message: %v", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	log.Printf("[Client] ✅ Message marshalled, size: %d bytes", len(data))

	log.Printf("[Client] Writing message to stream...")
	_, err = stream.Write(data)
	if err != nil {
		log.Printf("[Client] ❌ Failed to write to stream: %v", err)
		return fmt.Errorf("failed to write to stream: %w", err)
	}
	log.Printf("[Client] ✅ Sent ping: %s (seq: %d)", message.Content, message.Sequence)

	// Read response
	log.Printf("[Client] Reading response from server...")
	buffer := make([]byte, 4096)
	n, err := stream.Read(buffer)
	if err != nil {
		log.Printf("[Client] ❌ Failed to read response: %v", err)
		return fmt.Errorf("failed to read response: %w", err)
	}
	log.Printf("[Client] ✅ Response received, size: %d bytes", n)

	response, err := model.FromJSON(buffer[:n])
	if err != nil {
		log.Printf("[Client] ❌ Failed to parse response: %v", err)
		return fmt.Errorf("failed to parse response: %w", err)
	}

	log.Printf("[Client] ✅ Received response: %s (seq: %d)", response.Content, response.Sequence)
	log.Printf("[Client] Closing stream and connection (client's job is done)")

	return nil
}
