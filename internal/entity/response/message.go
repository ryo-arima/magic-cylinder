package response

import "github.com/ryo-arima/magic-cylinder/internal/entity/model"

// PingResponse represents a ping response
type PingResponse struct {
	Message *model.Message `json:"message"`
	Success bool           `json:"success"`
	Error   string         `json:"error,omitempty"`
}

// PongResponse represents a pong response
type PongResponse struct {
	Message *model.Message `json:"message"`
	Success bool           `json:"success"`
	Error   string         `json:"error,omitempty"`
}
