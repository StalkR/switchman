package mullvadapp

import (
  "fmt"
  "regexp"
)

var (
  locationRE            = regexp.MustCompile(`^\s*Location:\s*(.*)$`)
  countryRE             = regexp.MustCompile(`^country ([a-z]*)$`)
  cityCountryRE         = regexp.MustCompile(`^city ([a-z]*), ([a-z]*)$`)
  cityCountryHostnameRE = regexp.MustCompile(`^city ([a-z]*), ([a-z]*), hostname ([-a-z0-9])$`)
)

// Current returns the current relay.
// It can be a country location, or a country and city location, or the relay hostname.
func (s *Server) Current() (string, error) {
  relayOptions, err := run("mullvad", "relay", "get")
  if err != nil {
    return "", err
  }
  m := locationRE.FindStringSubmatch(relayOptions)
  if m == nil {
    return "", fmt.Errorf("could not parse relay options")
  }
  location := m[1]
  switch {
  case countryRE.MatchString(location):
    m = countryRE.FindStringSubmatch(location)
    return m[1], nil // country

  case cityCountryRE.MatchString(location):
    m = cityCountryRE.FindStringSubmatch(location)
    return fmt.Sprintf("%s %s", m[2], m[1]), nil // country city

  case cityCountryHostnameRE.MatchString(location):
    m = cityCountryHostnameRE.FindStringSubmatch(location)
    return m[3], nil // hostname
  }
  return "", fmt.Errorf("could not parse relay location")
}
