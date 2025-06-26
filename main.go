/*
Binary switchman is a small web server to switch VPN exits.

Supported VPNs:
  - Mullvad: switch between servers fetched from their API, single config (`wg0.conf`)
  - basic OpenVPN: switch between `remote` commented out with `;`, single config (`*.conf`)
  - basic WireGuard: switch between `Endpoint` commented out with `#`, single config (`wg0.conf`)
*/
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/StalkR/switchman/mullvad"
	"github.com/StalkR/switchman/openvpn"
	"github.com/StalkR/switchman/wireguard"
)

var (
	flagListen = flag.String("listen", ":81", "Port to listen on for HTTP requests.")

	flagMullvad   = flag.Bool("mullvad", false, "Switch Mullvad.")
	flagOpenVPN   = flag.Bool("openvpn", false, "Switch OpenVPN.")
	flagWireGuard = flag.Bool("wireguard", false, "Switch WireGuard.")
)

func main() {
	flag.Parse()

	s, err := func() (Switchable, error) {
		switch {
		case *flagMullvad:
			return mullvad.New()
		case *flagOpenVPN:
			return openvpn.New()
		case *flagWireGuard:
			return wireguard.New()
		default:
			return autodetect()
		}
	}()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(serve(s, *flagListen))
}

// A Switchable implements support for a VPN that can be switched servers.
// Optionally, it can also implement Indexable to provide a custom index.
type Switchable interface {
	// Current returns the current server.
	Current() (string, error)
	// List lists available servers.
	List() ([]string, error)
	// Switch switches to the specified server.
	Switch(server string) error
}

var errNotConfigured = errors.New("not configured")

func autodetect() (Switchable, error) {
	for _, f := range []func() (Switchable, error){
		func() (Switchable, error) { return mullvad.New() },
		func() (Switchable, error) { return openvpn.New() },
		func() (Switchable, error) { return wireguard.New() },
	} {
		if s, err := f(); err == nil {
			return s, nil
		}
	}
	return nil, fmt.Errorf("no supported VPN found")
}
