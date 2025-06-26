package mullvad

import (
  "fmt"
  "os"
  "os/exec"
  "regexp"
  "time"
)

var (
  endpointRE  = regexp.MustCompile("(?m)^(Endpoint = .*)$")
  publicKeyRE = regexp.MustCompile("(?m)^(PublicKey = .*)$")
)

// Switch switches to the specified server.
func (s *Server) Switch(server string) error {
  current, err := s.Current()
  if err != nil {
    return err
  }
  if server == current {
    return nil // not an error, just nothing to do
  }
  relays, err := s.findRelays(server)
  if err != nil {
    return err
  }
  publicKey := relays[len(relays)-1].PublicKey

  b, err := os.ReadFile(s.config)
  if err != nil {
    return err
  }
  b = endpointRE.ReplaceAll(b, []byte(fmt.Sprintf("Endpoint = %s", server)))
  b = publicKeyRE.ReplaceAll(b, []byte(fmt.Sprintf("PublicKey = %s", publicKey)))
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
