package controller

import (
	"net/http"

	"github.com/quic-go/webtransport-go"
	"github.com/ryo-arima/magic-cylinder/internal/entity/model"
)

// CommonController defines the interface for controller operations
type CommonController interface {
	HandleWebTransport(server *webtransport.Server, w http.ResponseWriter, r *http.Request, targetURL string)
	HandlePing(message *model.Message) (*model.Message, error)
	HandlePong(message *model.Message) (*model.Message, error)
}
