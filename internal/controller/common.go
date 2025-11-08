package controller

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/quic-go/webtransport-go"
	"github.com/ryo-arima/magic-cylinder/internal/entity/model"
	"github.com/ryo-arima/magic-cylinder/internal/repository"
)

// commonController implements the CommonController interface
type commonController struct {
	repo repository.CommonRepository
}

// NewCommonController creates a new controller instance with repository dependency
func NewCommonController(repo repository.CommonRepository) CommonController {
	return &commonController{
		repo: repo,
	}
}

// HandleWebTransport handles incoming WebTransport connection requests
func (c *commonController) HandleWebTransport(server *webtransport.Server, w http.ResponseWriter, r *http.Request, targetURL string) {
	log.Printf("[Controller] ============================================")
	log.Printf("[Controller] New WebTransport connection request")
	log.Printf("[Controller]   Remote Address: %s", r.RemoteAddr)
	log.Printf("[Controller]   Method: %s", r.Method)
	log.Printf("[Controller]   URL: %s", r.URL.String())
	log.Printf("[Controller]   Protocol: %s", r.Proto)
	log.Printf("[Controller] ============================================")

	conn, err := server.Upgrade(w, r)
	if err != nil {
		log.Printf("[Controller] ❌ Failed to upgrade to WebTransport: %v", err)
		log.Printf("[Controller]   Error details: %T", err)
		http.Error(w, "Failed to upgrade", http.StatusInternalServerError)
		return
	}

	log.Printf("[Controller] ✅ WebTransport connection established successfully")
	log.Printf("[Controller]   Connection ID: %p", conn)
	log.Printf("[Controller]   Target URL for echo: %s", targetURL)

	go c.handleConnection(conn, targetURL)
}

// HandlePlain handles plaintext POST /plain requests by reading a JSON message,
// generating the next message via repository, replying with JSON, and echoing
// to the target using plaintext or WebTransport depending on target URL scheme.
func (c *commonController) HandlePlain(w http.ResponseWriter, r *http.Request, targetURL string) {
	log.Printf("[Controller] (plain) ============================================")
	log.Printf("[Controller] (plain) New plaintext request")
	log.Printf("[Controller] (plain)   Remote Address: %s", r.RemoteAddr)
	log.Printf("[Controller] (plain)   Method: %s", r.Method)
	log.Printf("[Controller] (plain)   URL: %s", r.URL.String())

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[Controller] (plain) ❌ Failed to read body: %v", err)
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	msg, err := model.FromJSON(body)
	if err != nil {
		log.Printf("[Controller] (plain) ❌ Failed to parse JSON: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	log.Printf("[Controller] (plain)[RAW] %s", msg.Content)

	var resp *model.Message
	if msg.Type == model.PingMessage {
		log.Printf("[Controller] (plain) Routing to HandlePing...")
		resp, err = c.HandlePing(msg)
	} else {
		log.Printf("[Controller] (plain) Routing to HandlePong...")
		resp, err = c.HandlePong(msg)
	}
	if err != nil {
		log.Printf("[Controller] (plain) ❌ Handler failed: %v", err)
		http.Error(w, "handler error", http.StatusInternalServerError)
		return
	}

	data, err := resp.ToJSON()
	if err != nil {
		log.Printf("[Controller] (plain) ❌ Failed to marshal response: %v", err)
		http.Error(w, "marshal error", http.StatusInternalServerError)
		return
	}
	log.Printf("[Controller] (plain)[RAW] %s", resp.Content)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		log.Printf("[Controller] (plain) ❌ Failed to write response: %v", err)
	}

	if targetURL != "" {
		log.Printf("[Controller] (plain) Triggering echo to target: %s", targetURL)
		go func() {
			var echoErr error
			if strings.HasPrefix(targetURL, "http://") || strings.HasPrefix(targetURL, "https://") {
				echoErr = c.repo.SendPlainEchoToTarget(targetURL, resp)
			} else {
				echoErr = c.repo.SendEchoToTarget(targetURL, resp)
			}
			if echoErr != nil {
				log.Printf("[Controller] (plain) ❌ Echo to target %s failed: %v", targetURL, echoErr)
			} else {
				log.Printf("[Controller] (plain) ✅ Echo to target %s completed successfully", targetURL)
			}
		}()
	}
}

// handleConnection manages the lifecycle of a WebTransport connection
func (c *commonController) handleConnection(conn *webtransport.Session, targetURL string) {
	log.Printf("[Controller] Starting connection handler goroutine")
	log.Printf("[Controller]   Connection: %p", conn)

	defer func() {
		log.Printf("[Controller] Closing connection: %p", conn)
		conn.CloseWithError(0, "connection closed")
		log.Printf("[Controller] Connection closed successfully")
	}()

	log.Printf("[Controller] Waiting for incoming streams...")

	for {
		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Printf("[Controller] ❌ Failed to accept stream: %v", err)
			log.Printf("[Controller]   Error type: %T", err)
			return
		}

		log.Printf("[Controller] ✅ Stream accepted successfully: %d", stream.StreamID())
		go c.handleStream(stream, targetURL)
	}
}

// handleStream processes an individual stream within a WebTransport connection
func (c *commonController) handleStream(stream *webtransport.Stream, targetURL string) {
	defer stream.Close()

	log.Printf("[Controller] ==========================================")
	log.Printf("[Controller] Processing new stream: %d", stream.StreamID())

	buffer := make([]byte, 4096)
	n, err := stream.Read(buffer)
	if err != nil && err != io.EOF {
		log.Printf("[Controller] ❌ Failed to read from stream %d: %v", stream.StreamID(), err)
		return
	}

	log.Printf("[Controller] Read %d bytes from stream %d", n, stream.StreamID())

	message, err := model.FromJSON(buffer[:n])
	if err != nil {
		log.Printf("[Controller] ❌ Failed to parse message from stream %d: %v", stream.StreamID(), err)
		log.Printf("[Controller]   Raw data: %s", string(buffer[:n]))
		return
	}

	log.Printf("[Controller] ✅ Received %s message on stream %d", message.Type, stream.StreamID())
	log.Printf("[Controller]   Content: %s", message.Content)
	log.Printf("[Controller]   Sequence: %d", message.Sequence)
	log.Printf("[Controller]   From: %s", message.From)
	log.Printf("[Controller]   To: %s", message.To)
	log.Printf("[Controller][RAW] %s", message.Content)

	var response *model.Message
	if message.Type == model.PingMessage {
		log.Printf("[Controller] Routing to HandlePing...")
		response, err = c.HandlePing(message)
	} else {
		log.Printf("[Controller] Routing to HandlePong...")
		response, err = c.HandlePong(message)
	}

	if err != nil {
		log.Printf("[Controller] ❌ Failed to handle message: %v", err)
		return
	}

	responseData, err := response.ToJSON()
	if err != nil {
		log.Printf("[Controller] ❌ Failed to marshal response: %v", err)
		return
	}

	_, err = stream.Write(responseData)
	if err != nil {
		log.Printf("[Controller] ❌ Failed to write response to stream %d: %v", stream.StreamID(), err)
		return
	}

	log.Printf("[Controller] ✅ Sent %s message on stream %d", response.Type, stream.StreamID())
	log.Printf("[Controller]   Content: %s", response.Content)
	log.Printf("[Controller]   Sequence: %d", response.Sequence)
	log.Printf("[Controller]   From: %s", response.From)
	log.Printf("[Controller]   To: %s", response.To)
	log.Printf("[Controller][RAW] %s", response.Content)

	// Echo message to target server if targetURL is provided
	if targetURL != "" {
		log.Printf("[Controller] Triggering echo to target: %s", targetURL)
		log.Printf("[Controller] Note: Echo will create a NEW connection to target")
		go func() {
			var echoErr error
			if strings.HasPrefix(targetURL, "http://") || strings.HasPrefix(targetURL, "https://") {
				echoErr = c.repo.SendPlainEchoToTarget(targetURL, response)
			} else {
				echoErr = c.repo.SendEchoToTarget(targetURL, response)
			}
			if echoErr != nil {
				log.Printf("[Controller] ❌ Echo to target %s failed: %v", targetURL, echoErr)
			} else {
				log.Printf("[Controller] ✅ Echo to target %s completed successfully", targetURL)
			}
		}()
	} else {
		log.Printf("[Controller] No target URL configured, skipping echo")
	}

	log.Printf("[Controller] Stream %d processing complete (current connection will remain open for more streams)", stream.StreamID())
	log.Printf("[Controller] ==========================================")
}

// HandlePing processes a ping message
func (c *commonController) HandlePing(message *model.Message) (*model.Message, error) {
	log.Printf("[Controller] HandlePing called for message seq: %d", message.Sequence)
	response, err := c.repo.ProcessPing(message)
	if err != nil {
		log.Printf("[Controller] ❌ ProcessPing failed: %v", err)
		return nil, fmt.Errorf("failed to process ping: %w", err)
	}
	log.Printf("[Controller] ✅ ProcessPing successful, generated pong seq: %d", response.Sequence)
	return response, nil
}

// HandlePong processes a pong message
func (c *commonController) HandlePong(message *model.Message) (*model.Message, error) {
	log.Printf("[Controller] HandlePong called for message seq: %d", message.Sequence)
	response, err := c.repo.ProcessPong(message)
	if err != nil {
		log.Printf("[Controller] ❌ ProcessPong failed: %v", err)
		return nil, fmt.Errorf("failed to process pong: %w", err)
	}
	log.Printf("[Controller] ✅ ProcessPong successful, generated ping seq: %d", response.Sequence)
	return response, nil
}
