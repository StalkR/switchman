package wireguard

import (
	"os"
)

// New creates a new Server to switch a WireGuard server.
func New() (*Server, error) {
	const config = "/etc/wireguard/wg0.conf"
	if _, err := os.Stat(config); err != nil {
		return nil, err
	}
	return &Server{
		config: config,
	}, nil
}

// A Server implements the ability to switch a WireGuard server.
// It implements the Switchable interface.
type Server struct {
	config string
}
