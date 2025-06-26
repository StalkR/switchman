package mullvad

import (
  "fmt"
  "os"
  "strings"
  "sync"
  "time"
)

func New() (*server, error) {
  const config = "/etc/wireguard/wg0.conf"
  if _, err := os.Stat(config); err != nil {
    return nil, err
  }
  s := &server{
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

type server struct {
  config string

  m      sync.Mutex // protects below
  relays []relay
  error  error
}

func (s *server) periodicallyFetchEndpoints() {
  for ; ; time.Sleep(24 * time.Hour) {
    relays, err := s.fetchRelays()
    s.m.Lock()
    s.relays = relays
    s.error = err
    s.m.Unlock()
  }
}

func (s *server) listRelays() ([]relay, error) {
  s.m.Lock()
  defer s.m.Unlock()
  return s.relays, s.error
}
