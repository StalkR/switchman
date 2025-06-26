package mullvad

import (
  "fmt"
  "os"
  "strings"
  "sync"
  "time"
)

// New creates a new Server to switch a mullvad WireGuard server.
func New() (*Server, error) {
  const config = "/etc/wireguard/wg0.conf"
  if _, err := os.Stat(config); err != nil {
    return nil, err
  }
  s := &Server{
    config: config,
  }
  current, err := s.Current()
  if err != nil {
    return nil, err
  }
  fields := strings.Split(current, ":")
  if len(fields) != 2 || !strings.HasSuffix(fields[0], relaySuffix) {
    return nil, fmt.Errorf("not mullvad")
  }
  go s.periodicallyFetchEndpoints()
  return s, nil
}

// A Server implements the ability to switch a mullvad WireGuard server.
// It implements the Switchable and Indexable interfaces.
type Server struct {
  config string

  m      sync.Mutex // protects below
  relays []relay
  error  error
}

func (s *Server) periodicallyFetchEndpoints() {
  for ; ; time.Sleep(24 * time.Hour) {
    relays, err := s.fetchRelays()
    s.m.Lock()
    s.relays = relays
    s.error = err
    s.m.Unlock()
  }
}

func (s *Server) listRelays() ([]relay, error) {
  s.m.Lock()
  defer s.m.Unlock()
  return s.relays, s.error
}
