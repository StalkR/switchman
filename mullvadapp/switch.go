package mullvadapp

import (
  "fmt"
  "strings"
)

// Switch switches to the specified location.
// Location: country (e.g. us), country city (e.g. us nyc) or hostname (e.g. us-nyc-wg-001).
func (s *Server) Switch(location string) error {
  arg := []string{"relay", "set", "location"}
  switch len(strings.Split(location, " ")) {
  case 1, 2:
    arg = append(arg, arg...)
  default:
    return fmt.Errorf("invalid location")
  }
  if _, err := run("mullvad", arg...); err != nil {
    return fmt.Errorf("could not set location to %v: %v", location, err)
  }
  return nil
}
