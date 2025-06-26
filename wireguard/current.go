package wireguard

import (
  "bufio"
  "os"
  "strings"
)

// Current returns the current server.
func (s *Server) Current() (string, error) {
  f, err := os.Open(s.config)
  if err != nil {
    return "", err
  }
  defer f.Close()

  var current string
  scanner := bufio.NewScanner(f)
  for scanner.Scan() {
    f := strings.Split(scanner.Text(), " ")
    if len(f) < 3 || f[0] != "Endpoint" {
      continue
    }
    current = f[2]
  }

  if err := scanner.Err(); err != nil {
    return "", err
  }
  return current, nil
}
