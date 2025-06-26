package openvpn

import (
  "fmt"
  "path/filepath"
)

// New creates a new Server to switch an OpenVPN server.
func New() (*Server, error) {
  const configPattern = "/etc/openvpn/*.conf"
  matches, err := filepath.Glob(configPattern)
  if err != nil {
    return nil, err
  }
  if len(matches) == 0 || len(matches) > 1 {
    return nil, fmt.Errorf("found %v %v files; want 1", len(matches), configPattern)
  }
  return &Server{
    config: matches[0],
  }, nil
}

// A Server implements the ability to switch an OpenVPN server.
// It implements the Switchable interface.
type Server struct {
  config string
}
