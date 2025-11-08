package model

import (
	"encoding/json"
	"time"
)

// MessageType defines the type of message
type MessageType string

const (
	// PingMessage represents a ping message
	PingMessage MessageType = "ping"
	// PongMessage represents a pong message
	PongMessage MessageType = "pong"
)

// Message represents a ping-pong message exchanged between servers
type Message struct {
	Type      MessageType `json:"type"`
	Content   string      `json:"content"`
	Timestamp time.Time   `json:"timestamp"`
	Sequence  int         `json:"sequence"`
	From      string      `json:"from"`
	To        string      `json:"to"`
}

// NewPingMessage creates a new ping message
func NewPingMessage(content string, sequence int, from, to string) *Message {
	return &Message{
		Type:      PingMessage,
		Content:   content,
		Timestamp: time.Now(),
		Sequence:  sequence,
		From:      from,
		To:        to,
	}
}

// NewPongMessage creates a new pong message
func NewPongMessage(content string, sequence int, from, to string) *Message {
	return &Message{
		Type:      PongMessage,
		Content:   content,
		Timestamp: time.Now(),
		Sequence:  sequence,
		From:      from,
		To:        to,
	}
}

// ToJSON converts message to JSON bytes
func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON creates message from JSON bytes
func FromJSON(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return &msg, err
}

// String returns a string representation of the message
func (m *Message) String() string {
	return string(m.Type) + ": " + m.Content
}
