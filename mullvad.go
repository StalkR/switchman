package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func mullvad() (switchable, error) {
	const config = "/etc/wireguard/wg0.conf"
	if _, err := os.Stat(config); err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, errNotConfigured
	}
	s := &mullvadServer{
		config: config,
	}
	current, err := s.Current()
	if err != nil {
		return nil, errNotConfigured
	}
	fields := strings.Split(current, ":")
	if len(fields) != 2 || !strings.HasSuffix(fields[0], mullvadRelaySuffix) {
		return nil, errNotConfigured
	}
	go func() {
		for ; ; time.Sleep(24 * time.Hour) {
			api, err := mullvadFetchAPI()
			s.m.Lock()
			s.api = api
			s.apiLastError = err
			s.m.Unlock()
		}
	}()
	return s, nil
}

type mullvadServer struct {
	config string

	m            sync.Mutex // protects below
	api          *mullvadAPIResponse
	apiLastError error
}

func mullvadFetchAPI() (*mullvadAPIResponse, error) {
	resp, err := http.Get(mullvadAPI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var v mullvadAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}
	if len(v.Countries) == 0 {
		return nil, fmt.Errorf("weird, empty API response")
	}
	return &v, nil
}

const mullvadAPI = "https://api.mullvad.net/public/relays/wireguard/v1/"
const mullvadRelaySuffix = ".relays.mullvad.net"

type mullvadAPIResponse struct {
	Countries []struct {
		Name   string `json:"name"`
		Code   string `json:"code"`
		Cities []struct {
			Name      string  `json:"name"`
			Code      string  `json:"code"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Relays    []struct {
				Hostname     string `json:"hostname"` // add suffix and use port 51820
				IPv4         string `json:"ipv4_addr_in"`
				IPv6         string `json:"ipv6_addr_in"`
				PublicKey    string `json:"public_key"`
				MultihopPort int    `json:"multihop_port"`
			} `json:"relays"`
		} `json:"cities"`
	} `json:"countries"`
}

func (s *mullvadServer) Current() (string, error) {
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

type mullvadEndpoint struct {
	Endpoint  string
	PublicKey string
}

func (s *mullvadServer) listEndpoints() ([]mullvadEndpoint, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if s.apiLastError != nil {
		return nil, s.apiLastError
	}
	var servers []mullvadEndpoint
	for _, country := range s.api.Countries {
		for _, city := range country.Cities {
			for _, relay := range city.Relays {
				servers = append(servers, mullvadEndpoint{
					Endpoint:  fmt.Sprintf("%s%s:51820", relay.Hostname, mullvadRelaySuffix),
					PublicKey: relay.PublicKey,
				})
			}
		}
	}
	return servers, nil
}

func (s *mullvadServer) List() ([]string, error) {
	endpoints, err := s.listEndpoints()
	if err != nil {
		return nil, err
	}
	var servers []string
	for _, e := range endpoints {
		servers = append(servers, e.Endpoint)
	}
	return servers, nil
}

var publicKeyRE = regexp.MustCompile("(?m)^(PublicKey = .*)$")

func (s *mullvadServer) Switch(server string) error {
	found := false
	publicKey := ""
	endpoints, err := s.listEndpoints()
	if err != nil {
		return err
	}
	for _, e := range endpoints {
		if e.Endpoint == server {
			found = true
			publicKey = e.PublicKey
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
	b = endpointRE.ReplaceAll(b, []byte(fmt.Sprintf("Endpoint = %s", server)))
	b = publicKeyRE.ReplaceAll(b, []byte(fmt.Sprintf("PublicKey = %s", publicKey)))
	if err := os.WriteFile(s.config, b, 0644); err != nil {
		return err
	}

	return restartWireguard()
}
