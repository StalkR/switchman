package openvpn

import (
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
  "regexp"
  "strings"
  "time"
)

var disableRemoteRE = regexp.MustCompile("(?m)^(remote .*)$")

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
  b = disableRemoteRE.ReplaceAll(b, []byte(";$1"))
  enableRE, err := regexp.Compile("(?m)^;(remote " + regexp.QuoteMeta(server) + " .*)$")
  if err != nil {
    return err
  }
  b = enableRE.ReplaceAll(b, []byte("$1"))
  if err := os.WriteFile(s.config, b, 0644); err != nil {
    return err
  }

  return restartOpenVPN(s.config)
}

const device = "tun0"

func restartOpenVPN(config string) error {
  if out, err := exec.Command("invoke-rc.d", "openvpn", "stop").CombinedOutput(); err != nil {
    return fmt.Errorf("could not stop openvpn: %v - %v", err, string(out))
  }
  for ; ; time.Sleep(time.Second) {
    if err := exec.Command("ip", "link", "list", "dev", device).Run(); err != nil {
      break
    }
  }
  name := strings.TrimSuffix(filepath.Base(config), ".conf")
  if out, err := exec.Command("invoke-rc.d", "openvpn", "start", name).CombinedOutput(); err != nil {
    return fmt.Errorf("could not start openvpn: %v - %v", err, string(out))
  }
  return nil
}
