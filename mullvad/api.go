package mullvad

import (
  "encoding/json"
  "fmt"
  "net"
  "net/http"
  "strconv"
)

// https://api.mullvad.net/public/documentation/
// v2 has extra info (active, owned) but lacks v1 multihop_port
const (
  apiv1URL = "https://api.mullvad.net/public/relays/wireguard/v1/"
  apiv2URL = "https://api.mullvad.net/public/relays/wireguard/v2/"
)

const (
  relaySuffix = ".relays.mullvad.net"
  relayPort   = 51820
)

type apiv1Response struct {
  Countries []struct {
    Name   string `json:"name"`
    Code   string `json:"code"`
    Cities []struct {
      Name      string  `json:"name"`
      Code      string  `json:"code"`
      Latitude  float64 `json:"latitude"`
      Longitude float64 `json:"longitude"`
      Relays    []struct {
        Hostname     string `json:"hostname"` // add suffix and default port
        IPv4         string `json:"ipv4_addr_in"`
        IPv6         string `json:"ipv6_addr_in"`
        PublicKey    string `json:"public_key"`
        MultihopPort int    `json:"multihop_port"`
      } `json:"relays"`
    } `json:"cities"`
  } `json:"countries"`
}

type apiv2Response struct {
  Locations map[string]struct {
    Country   string  `json:"country"`
    City      string  `json:"city"`
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
  } `json:"locations"`
  WireGuard struct {
    PortRanges  [][]int `json:"port_ranges"`
    IPv4Gateway string  `json:"ipv4_gateway"`
    IPv6Gateway string  `json:"ipv6_gateway"`
    Relays      []struct {
      Hostname         string `json:"hostname"` // add suffix and default port
      Active           bool   `json:"active"`
      Owned            bool   `json:"owned"`
      Location         string `json:"location"`
      Provider         string `json:"provider"`
      Weight           int    `json:"weight"`
      IncludeInCountry bool   `json:"include_in_country"`
      IPv4             string `json:"ipv4_addr_in"`
      IPv6             string `json:"ipv6_addr_in"`
      PublicKey        string `json:"public_key"`
    } `json:"relays"`
  } `json:"wireguard"`
}

type relay struct {
  ID           string
  Hostname     string
  Port         int
  Owned        bool
  Country      string
  City         string
  PublicKey    string
  MultihopPort int
}

func (s *Server) fetchRelays() ([]relay, error) {
  resp1, err := http.Get(apiv1URL)
  if err != nil {
    return nil, err
  }
  defer resp1.Body.Close()
  var v1 apiv1Response
  if err := json.NewDecoder(resp1.Body).Decode(&v1); err != nil {
    return nil, err
  }
  if len(v1.Countries) == 0 {
    return nil, fmt.Errorf("empty APIv1 response")
  }
  resp2, err := http.Get(apiv2URL)
  if err != nil {
    return nil, err
  }
  defer resp2.Body.Close()
  var v2 apiv2Response
  if err := json.NewDecoder(resp2.Body).Decode(&v2); err != nil {
    return nil, err
  }
  if len(v2.WireGuard.Relays) == 0 {
    return nil, fmt.Errorf("empty APIv2 response")
  }

  multihopPort := map[string]int{}
  for _, country := range v1.Countries {
    for _, city := range country.Cities {
      for _, relay := range city.Relays {
        multihopPort[relay.Hostname] = relay.MultihopPort
      }
    }
  }

  locations := map[string]struct {
    Country string
    City    string
  }{}
  for location, e := range v2.Locations {
    locations[location] = struct {
      Country string
      City    string
    }{
      Country: e.Country,
      City:    e.City,
    }
  }

  var servers []relay
  for _, r := range v2.WireGuard.Relays {
    if !r.Active {
      continue
    }
    servers = append(servers, relay{
      ID:           r.Hostname,
      Hostname:     r.Hostname + relaySuffix,
      Port:         relayPort,
      Owned:        r.Owned,
      Country:      locations[r.Location].Country,
      City:         locations[r.Location].City,
      PublicKey:    r.PublicKey,
      MultihopPort: multihopPort[r.Hostname],
    })
  }

  return servers, nil
}

// findRelays finds the relays for this server.
// If single-hop, it returns the single relay.
// If multi-hop, it returns the entry relay then the exit.
func (s *Server) findRelays(server string) ([]relay, error) {
  host, sport, err := net.SplitHostPort(server)
  if err != nil {
    return nil, err
  }
  port, err := strconv.Atoi(sport)
  if err != nil {
    return nil, err
  }
  relays, err := s.listRelays()
  if err != nil {
    return nil, err
  }
  var entry relay
  for _, e := range relays {
    // single-hop
    if e.Hostname == host && e.Port == port {
      return []relay{e}, nil
    }
    // multi-hop
    if e.Hostname == host {
      entry = e
    }
  }
  for _, e := range relays {
    if e.MultihopPort == port {
      if entry.Hostname == "" {
        return nil, fmt.Errorf("found exit server (multihop port) but not entry server %v", host)
      }
      return []relay{entry, e}, nil
    }
  }
  return nil, fmt.Errorf("server %v not found", server)
}
