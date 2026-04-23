package mullvadapp

import (
  "fmt"
  "os/exec"
)

// New creates a new Server to switch mullvad via app cli.
func New() (*Server, error) {
  if _, err := exec.LookPath("mullvad"); err != nil {
    return nil, fmt.Errorf("mullvad binary not found in PATH")
  }
  return &Server{}, nil
}

// A Server implements the ability to switch mullvad via app cli.
// It implements the Switchable and Indexable interfaces.
type Server struct{}

func run(name string, arg ...string) (string, error) {
  b, err := exec.Command(name, arg...).CombinedOutput()
  return string(b), err
}
