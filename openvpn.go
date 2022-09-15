package main

import (
  "bufio"
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
  "regexp"
  "strings"
  "time"
)

func openVPN() (switchable, error) {
  matches, err := filepath.Glob("/etc/openvpn/*.conf")
  if err != nil {
    return nil, err
  }
  if len(matches) == 0 {
    return nil, errNotConfigured
  }
  if len(matches) > 1 {
    return nil, fmt.Errorf("multiple /etc/openvpn/*.conf files found: %v; want 1", len(matches))
  }
  return &openVPNserver{
    config: matches[0],
  }, nil
}

type openVPNserver struct {
  config string
}

func (s *openVPNserver) Current() (string, error) {
  f, err := os.Open(s.config)
  if err != nil {
    return "", err
  }
  defer f.Close()

  var current string
  scanner := bufio.NewScanner(f)
  for scanner.Scan() {
    f := strings.Split(scanner.Text(), " ")
    if len(f) < 3 || f[0] != "remote" {
      continue
    }
    current = f[1]
  }

  if err := scanner.Err(); err != nil {
    return "", err
  }
  return current, nil
}

func (s *openVPNserver) List() ([]string, error) {
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

var disableRemoteRE = regexp.MustCompile("(?m)^(remote .*)$")

func (s *openVPNserver) Switch(server string) error {
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

  if out, err := exec.Command("invoke-rc.d", "openvpn", "stop").CombinedOutput(); err != nil {
    return fmt.Errorf("could not stop openvpn: %v - %v", err, string(out))
  }

  dev := "tun0"
  for ; ; time.Sleep(time.Second) {
    if err := exec.Command("ip", "link", "list", "dev", dev).Run(); err != nil {
      break
    }
  }

  name := strings.TrimSuffix(filepath.Base(s.config), ".conf")
  if out, err := exec.Command("invoke-rc.d", "openvpn", "start", name).CombinedOutput(); err != nil {
    return fmt.Errorf("could not start openvpn: %v - %v", err, string(out))
  }

  return nil
}

func (s *openVPNserver) Next() error {
  list, err := s.List()
  if err != nil {
    return err
  }
  current, err := s.Current()
  if err != nil {
    return err
  }
  var next string
  for i, e := range list {
    if e == current {
      next = list[(i+1)%len(list)]
      break
    }
  }
  if next == "" {
    return fmt.Errorf("could not find next server")
  }
  return s.Switch(next)
}
