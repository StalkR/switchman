package wireguard

import (
  "fmt"
  "os"
  "os/exec"
  "regexp"
  "time"
)

var endpointRE = regexp.MustCompile("(?m)^(Endpoint = .*)$")

// Switch switches to the specified server.
func (s *Server) Switch(server string) error {
  found := false
  list, err := s.List()
  if err != nil {
    return err
  }
  for _, n := range list {
    if n == server {
      found = true
      break
    }
  }
  if !found {
    return fmt.Errorf("server %v not found", server)
  }
  current, err := s.Current()
  if err != nil {
    return err
  }
  if server == current {
    return nil // not an error, just nothing to do
  }

  b, err := os.ReadFile(s.config)
  if err != nil {
    return err
  }
  b = endpointRE.ReplaceAll(b, []byte("#$1"))
  enableRE, err := regexp.Compile("(?m)^#(Endpoint = " + regexp.QuoteMeta(server) + ")$")
  if err != nil {
    return err
  }
  b = enableRE.ReplaceAll(b, []byte("$1"))
  if err := os.WriteFile(s.config, b, 0644); err != nil {
    return err
  }

  return restart()
}

func restart() error {
  const device = "wg0"
  // check if running before stop or it will fail
  if err := exec.Command("wg", "show", device).Run(); err == nil {
    if out, err := exec.Command("wg-quick", "down", device).CombinedOutput(); err != nil {
      return fmt.Errorf("could not stop wg: %v - %v", err, string(out))
    }
  }
  for ; ; time.Sleep(time.Second) {
    if err := exec.Command("ip", "link", "list", "dev", device).Run(); err != nil {
      break
    }
  }
  if out, err := exec.Command("wg-quick", "up", device).CombinedOutput(); err != nil {
    return fmt.Errorf("could not start wg: %v - %v", err, string(out))
  }
  return nil
}
