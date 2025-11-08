package request

import "github.com/ryo-arima/magic-cylinder/internal/entity/model"

// PingRequest represents a ping request
type PingRequest struct {
	Message *model.Message `json:"message"`
}

// PongRequest represents a pong request
type PongRequest struct {
	Message *model.Message `json:"message"`
}
