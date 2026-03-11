// Package proxy implements API proxy for LLM providers.
// Phase 4 implementation.
package proxy

// Config for the proxy server.
type Config struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
}

// Server handles API proxy requests.
type Server struct {
	config Config
}

// NewServer creates a new proxy server.
func NewServer(config Config) *Server {
	return &Server{config: config}
}
