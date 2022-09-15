package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func wireGuard() (switchable, error) {
	const config = "/etc/wireguard/wg0.conf"
	if _, err := os.Stat(config); err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, errNotConfigured
	}
	return &wireGuardServer{
		config: config,
	}, nil
}

type wireGuardServer struct {
	config string
}

func (s *wireGuardServer) Current() (string, error) {
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

func (s *wireGuardServer) List() ([]string, error) {
	f, err := os.Open(s.config)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var servers []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := strings.TrimPrefix(scanner.Text(), "#")
		f := strings.Split(t, " ")
		if len(f) < 3 || f[0] != "Endpoint" {
			continue
		}
		servers = append(servers, f[2])
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return servers, nil
}

var disableEndpointRE = regexp.MustCompile("(?m)^(Endpoint = .*)$")

func (s *wireGuardServer) Switch(server string) error {
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
	b = disableEndpointRE.ReplaceAll(b, []byte("#$1"))
	enableRE, err := regexp.Compile("(?m)^#(Endpoint = " + regexp.QuoteMeta(server) + ")$")
	if err != nil {
		return err
	}
	b = enableRE.ReplaceAll(b, []byte("$1"))
	if err := os.WriteFile(s.config, b, 0644); err != nil {
		return err
	}

	if out, err := exec.Command("wg-quick", "down", "wg0").CombinedOutput(); err != nil {
		return fmt.Errorf("could not stop wg: %v - %v", err, string(out))
	}

	for ; ; time.Sleep(time.Second) {
		if err := exec.Command("ip", "link", "list", "dev", "wg0").Run(); err != nil {
			break
		}
	}

	if out, err := exec.Command("wg-quick", "up", "wg0").CombinedOutput(); err != nil {
		return fmt.Errorf("could not start wg: %v - %v", err, string(out))
	}

	return nil
}

func (s *wireGuardServer) Next() error {
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
