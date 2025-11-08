package repository

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync"

	"github.com/quic-go/webtransport-go"
	"github.com/ryo-arima/magic-cylinder/internal/entity/model"
)

// commonRepository implements the CommonRepository interface
type commonRepository struct {
	sequence int        // Current message sequence number
	mu       sync.Mutex // Mutex for thread-safe sequence operations
}

// NewCommonRepository creates a new repository instance
func NewCommonRepository() CommonRepository {
	return &commonRepository{
		sequence: 0,
	}
}

// ProcessPing processes a ping message and generates a pong response
func (r *commonRepository) ProcessPing(message *model.Message) (*model.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("[Repository] ==========================================")
	log.Printf("[Repository] ProcessPing started")
	log.Printf("[Repository]   Input: %s (seq: %d, from: %s)", message.Content, message.Sequence, message.From)

	r.sequence++
	response := model.NewPongMessage(
		fmt.Sprintf("Pong response to: %s", message.Content),
		r.sequence,
		"repository",
		message.From,
	)

	log.Printf("[Repository] ✅ Pong generated successfully")
	log.Printf("[Repository]   Output: %s (seq: %d, to: %s)", response.Content, response.Sequence, response.To)
	log.Printf("[Repository] ==========================================")

	return response, nil
}

// ProcessPong processes a pong message and generates a ping response
func (r *commonRepository) ProcessPong(message *model.Message) (*model.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Printf("[Repository] ==========================================")
	log.Printf("[Repository] ProcessPong started")
	log.Printf("[Repository]   Input: %s (seq: %d, from: %s)", message.Content, message.Sequence, message.From)

	r.sequence++
	response := model.NewPingMessage(
		fmt.Sprintf("Ping response to: %s", message.Content),
		r.sequence,
		"repository",
		message.From,
	)

	log.Printf("[Repository] ✅ Ping generated successfully")
	log.Printf("[Repository]   Output: %s (seq: %d, to: %s)", response.Content, response.Sequence, response.To)
	log.Printf("[Repository] ==========================================")

	return response, nil
}

// GetSequence returns the current sequence number
func (r *commonRepository) GetSequence() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.sequence
}

// IncrementSequence increments and returns the new sequence number
func (r *commonRepository) IncrementSequence() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sequence++
	return r.sequence
}

// SendEchoToTarget sends a message echo to the target server URL
func (r *commonRepository) SendEchoToTarget(targetURL string, message *model.Message) error {
	log.Printf("[Repository] ==========================================")
	log.Printf("[Repository] SendEchoToTarget started")
	log.Printf("[Repository]   Target URL: %s", targetURL)
	log.Printf("[Repository]   Message: %s (seq: %d)", message.Content, message.Sequence)
	log.Printf("[Repository] Creating NEW connection to target server")

	dialer := &webtransport.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	log.Printf("[Repository] Dialing target server...")
	_, conn, err := dialer.Dial(context.Background(), targetURL, nil)
	if err != nil {
		log.Printf("[Repository] ❌ Failed to dial target: %v", err)
		return fmt.Errorf("failed to dial target: %w", err)
	}
	defer func() {
		conn.CloseWithError(0, "echo complete")
		log.Printf("[Repository] Connection to target closed")
	}()
	log.Printf("[Repository] ✅ Connected to target successfully")

	log.Printf("[Repository] Opening stream to target...")
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		log.Printf("[Repository] ❌ Failed to open stream: %v", err)
		return fmt.Errorf("failed to open stream: %w", err)
	}
	defer stream.Close()
	log.Printf("[Repository] ✅ Stream opened: %d", stream.StreamID())

	log.Printf("[Repository] Marshalling message to JSON...")
	data, err := message.ToJSON()
	if err != nil {
		log.Printf("[Repository] ❌ Failed to marshal message: %v", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	log.Printf("[Repository] ✅ Message marshalled, size: %d bytes", len(data))

	log.Printf("[Repository] Writing message to target stream...")
	_, err = stream.Write(data)
	if err != nil {
		log.Printf("[Repository] ❌ Failed to write to stream: %v", err)
		return fmt.Errorf("failed to write to stream: %w", err)
	}
	log.Printf("[Repository] ✅ Message written to target: %s (seq: %d)", message.Content, message.Sequence)

	// Read response from target server
	log.Printf("[Repository] Reading response from target...")
	buffer := make([]byte, 4096)
	n, err := stream.Read(buffer)
	if err != nil {
		log.Printf("[Repository] ❌ Failed to read response: %v", err)
		return fmt.Errorf("failed to read response: %w", err)
	}
	log.Printf("[Repository] ✅ Response received, size: %d bytes", n)

	response, err := model.FromJSON(buffer[:n])
	if err != nil {
		log.Printf("[Repository] ❌ Failed to parse response: %v", err)
		return fmt.Errorf("failed to parse response: %w", err)
	}

	log.Printf("[Repository] ✅ Echo response received: %s (seq: %d)", response.Content, response.Sequence)
	log.Printf("[Repository] Connection will be closed after this function returns")

	// Connection will be closed by defer statements
	// The response is already logged, no need to process it further in this connection
	log.Printf("[Repository] ==========================================")

	return nil
}
