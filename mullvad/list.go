package mullvad

import (
  "fmt"
)

// List lists available servers.
func (s *Server) List() ([]string, error) {
  relays, err := s.listRelays()
  if err != nil {
    return nil, err
  }
  var servers []string
  for _, e := range relays {
    servers = append(servers, fmt.Sprintf("%s:%d", e.Hostname, e.Port))
  }
  return servers, nil
}
