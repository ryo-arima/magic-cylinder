package internal

import (
	"log"
	"net/http"
	"time"

	"github.com/quic-go/webtransport-go"
	"github.com/ryo-arima/magic-cylinder/internal/controller"
	"github.com/ryo-arima/magic-cylinder/internal/repository"
)

// Router handles routing and dependency injection
type Router struct {
	commonController controller.CommonController
	commonRepository repository.CommonRepository
	targetURL        string
}

// NewRouter creates a new router with injected dependencies
func NewRouter(
	commonController controller.CommonController,
	commonRepository repository.CommonRepository,
	targetURL string,
) *Router {
	return &Router{
		commonController: commonController,
		commonRepository: commonRepository,
		targetURL:        targetURL,
	}
}

// SetupRoutes initializes routes and handlers
func (r *Router) SetupRoutes(server *webtransport.Server) {
	log.Printf("[Router] Setting up routes...")

	log.Printf("[Router] Registering /webtransport endpoint")
	http.HandleFunc("/webtransport", r.handleWebTransport(server))

	log.Printf("[Router] Registering /plain endpoint (plaintext mode)")
	http.HandleFunc("/plain", r.handlePlain)

	log.Printf("[Router] Registering /health endpoint")
	http.HandleFunc("/health", r.handleHealth)

	log.Printf("[Router] All routes registered successfully")
}

// handleWebTransport handles WebTransport connections
func (r *Router) handleWebTransport(server *webtransport.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("[Router] WebTransport request received from %s", req.RemoteAddr)
		r.commonController.HandleWebTransport(server, w, req, r.targetURL)
	}
}

// handlePlain handles plaintext HTTP POST messages at /plain
func (r *Router) handlePlain(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[Router] Plaintext request received from %s", req.RemoteAddr)
	r.commonController.HandlePlain(w, req, r.targetURL)
}

// handleHealth handles health check requests
func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	log.Printf("[Router] Health check request received from %s", req.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// InitializeDependencies creates and returns all required dependencies
func InitializeDependencies(targetURL string, delaySeconds int) *Router {
	log.Printf("[Router] Initializing dependencies with target URL: %s", targetURL)
	if delaySeconds < 0 {
		delaySeconds = 0
	}
	commonRepo := repository.NewCommonRepository(time.Duration(delaySeconds) * time.Second)
	commonController := controller.NewCommonController(commonRepo)
	log.Printf("[Router] Dependencies initialized successfully")
	return NewRouter(commonController, commonRepo, targetURL)
}
