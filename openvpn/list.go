package openvpn

import (
  "bufio"
  "os"
  "strings"
)

// List lists available servers.
func (s *Server) List() ([]string, error) {
  f, err := os.Open(s.config)
  if err != nil {
    return nil, err
  }
  defer f.Close()

  var servers []string
  scanner := bufio.NewScanner(f)
  for scanner.Scan() {
    t := strings.TrimPrefix(scanner.Text(), ";")
    f := strings.Split(t, " ")
    if len(f) < 3 || f[0] != "remote" {
      continue
    }
    servers = append(servers, f[1])
  }

  if err := scanner.Err(); err != nil {
    return nil, err
  }
  return servers, nil
}
