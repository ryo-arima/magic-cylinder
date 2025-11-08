package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

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

	// Send initial ping to trigger the pingpong loop (supports WebTransport or /plain)
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
	u, err := url.Parse(serverURL)
	if err != nil {
		log.Printf("[Client] ❌ Invalid server URL: %v", err)
		return fmt.Errorf("invalid server URL: %w", err)
	}
	log.Printf("[Client] ✅ Server URL parsed successfully")

	// Mode decision: explicit path match (exact) for /plain, otherwise WebTransport
	cleanPath := strings.TrimSuffix(u.Path, "/")
	switch {
	case cleanPath == "/plain":
		log.Printf("[Client] Mode selected: PLAINTEXT (HTTP POST)")
		return sendPlain(u)
	case cleanPath == "/webtransport":
		log.Printf("[Client] Mode selected: WEBTRANSPORT")
		return sendWebTransportPing(u)
	default:
		log.Printf("[Client] ⚠ Unknown path '%s' -> defaulting to WEBTRANSPORT attempt", cleanPath)
		return sendWebTransportPing(u)
	}
}

func sendWebTransportPing(u *url.URL) error {
	// Create WebTransport dialer with insecure skip verify for self-signed certs
	log.Printf("[Client] Creating WebTransport dialer with TLS InsecureSkipVerify")
	dialer := &webtransport.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	serverURL := u.String()
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

	// Create and send ping message
	log.Printf("[Client] Creating ping message...")
	message := model.NewPingMessage("Initial ping from client", 1, "client", "server")
	log.Printf("[Client] Message created: %s (seq: %d)", message.Content, message.Sequence)

	data, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	_, err = stream.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to stream: %w", err)
	}
	log.Printf("[Client] ✅ Sent ping: %s (seq: %d)", message.Content, message.Sequence)

	// Read response (one-shot)
	buffer := make([]byte, 4096)
	n, err := stream.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	response, err := model.FromJSON(buffer[:n])
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	log.Printf("[Client] ✅ Received response: %s (seq: %d)", response.Content, response.Sequence)
	return nil
}

func sendPlain(u *url.URL) error {
	// Server listens with TLS only; auto-upgrade http -> https for /plain
	if u.Scheme == "http" {
		log.Printf("[Client] (plain) Upgrading scheme http -> https for TLS endpoint")
		u.Scheme = "https"
	}

	// Build message
	message := model.NewPingMessage("Initial ping from client (plain)", 1, "client", "server")
	data, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{Transport: tr}
	log.Printf("[Client] (plain) POST %s", u.String())
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	log.Printf("[Client] (plain) Response: %s", resp.Status)
	if len(body) > 0 {
		if rmsg, perr := model.FromJSON(body); perr == nil {
			log.Printf("[Client] (plain) Received response: %s (seq: %d)", rmsg.Content, rmsg.Sequence)
		} else {
			log.Printf("[Client] (plain) Body: %s", string(body))
		}
	}
	return nil
}
