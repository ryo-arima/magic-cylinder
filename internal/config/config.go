package config

// ServerConfig holds the configuration for a server instance
type ServerConfig struct {
	Port      string // Server port number
	CertFile  string // Path to TLS certificate file
	KeyFile   string // Path to TLS key file
	Name      string // Server name for logging
	TargetURL string // URL of the other server to echo messages to
}

// NewServerConfig creates a new server configuration
func NewServerConfig(port, name, targetURL string) *ServerConfig {
	return &ServerConfig{
		Port:      port,
		CertFile:  "certs/server.crt",
		KeyFile:   "certs/server.key",
		Name:      name,
		TargetURL: targetURL,
	}
}
