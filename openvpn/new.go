package openvpn

import (
  "fmt"
  "path/filepath"
)

func New() (*server, error) {
  const configPattern = "/etc/openvpn/*.conf"
  matches, err := filepath.Glob(configPattern)
  if err != nil {
    return nil, err
  }
  if len(matches) == 0 || len(matches) > 1 {
    return nil, fmt.Errorf("found %v %v files; want 1", len(matches), configPattern)
  }
  return &server{
    config: matches[0],
  }, nil
}

type server struct {
  config string
}
