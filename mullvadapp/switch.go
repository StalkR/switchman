package mullvadapp

import (
  "fmt"
  "strings"
)

// Switch switches to the specified location.
// Location can be:
// - country (e.g. us), 1 argument
// - country and city (e.g. us nyc), 2 arguments
// - hostname (e.g. us-nyc-wg-001), 1 argument
func (s *Server) Switch(location string) error {
  cmd := []string{"relay", "set", "location"}
  switch args := strings.Split(location, " "); len(args) {
  case 1, 2:
    cmd = append(cmd, args...)
  default:
    return fmt.Errorf("invalid location")
  }
  if _, err := run("mullvad", cmd...); err != nil {
    return fmt.Errorf("could not set location to %v: %v", location, err)
  }
  return nil
}
