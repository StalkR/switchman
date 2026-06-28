package mullvadapp

import (
  "fmt"
  "sort"
  "strings"
)

// List lists available relay locations.
func (s *Server) List() ([]string, error) {
  relays, err := s.listRelays()
  if err != nil {
    return nil, err
  }
  var servers []string
  for _, relay := range relays {
    servers = append(servers, relay.Location())
  }
  return servers, nil
}

func (s *Server) listRelays() ([]*relay, error) {
  // assumed already sorted
  list, err := run("mullvad", "relay", "list")
  if err != nil {
    return nil, err
  }
  var relays []*relay
  var country, city string
  for _, line := range strings.Split(list, "\n") {
    if !strings.HasPrefix(line, "\t\t") {
      continue
    }

    fields := strings.Fields(line)
    r := &relay{
      Country:  fields[0][:2],
      City:     fields[0][3:6],
      Hostname: fields[0],
      IPv4:     strings.Trim(fields[1], "(),"),
      IPv6:     strings.Trim(fields[2], "(),"),
      // shorten 'Mullvad-owned' into just 'owned'
      Ownership: strings.ReplaceAll(strings.Trim(fields[7], "()"), "Mullvad-", ""),
      HostedBy:  fields[6],
    }

    // relay entries to choose location by country or country and city
    if country != r.Country {
      country = r.Country
      relays = append(relays, &relay{Country: r.Country})
    }
    if city != r.City {
      city = r.City
      relays = append(relays, &relay{Country: r.Country, City: r.City})
    }

    // relays by hostname
    relays = append(relays, r)
  }
  sort.Slice(relays, func(i, j int) bool {
    if relays[i].Country == relays[j].Country {
      if relays[i].City == relays[j].City {
        return relays[i].Hostname < relays[j].Hostname
      }
      return relays[i].City < relays[j].City
    }
    return relays[i].Country < relays[j].Country
  })
  return relays, nil
}

type relay struct {
  Country   string
  City      string
  Hostname  string
  IPv4      string
  IPv6      string
  Ownership string
  HostedBy  string
}

func (s *relay) Location() string {
  switch {
  case s.Hostname != "":
    return s.Hostname
  case s.City != "":
    return fmt.Sprintf("%s %s", s.Country, s.City)
  default:
    return s.Country
  }
}
