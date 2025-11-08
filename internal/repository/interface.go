package repository

import "github.com/ryo-arima/magic-cylinder/internal/entity/model"

// CommonRepository defines the interface for repository operations
type CommonRepository interface {
	// ProcessPing processes a ping message and returns a pong response
	ProcessPing(message *model.Message) (*model.Message, error)
	// ProcessPong processes a pong message and returns a ping response
	ProcessPong(message *model.Message) (*model.Message, error)
	// GetSequence returns the current sequence number
	GetSequence() int
	// IncrementSequence increments and returns the new sequence number
	IncrementSequence() int
	// SendEchoToTarget sends a message echo to the target server URL
	SendEchoToTarget(targetURL string, message *model.Message) error
	// SendPlainEchoToTarget sends a message echo to the target over HTTP (plaintext mode)
	SendPlainEchoToTarget(targetURL string, message *model.Message) error
}
