// Binary switchman is a small web server to switch VPN exits.
//
// Supported VPN:
//   - Mullvad: switch between servers fetched from their API, single config (`wg0.conf`)
//   - basic OpenVPN: switch between `remote` commented out with `;`, single config (`*.conf`)
//   - basic WireGuard: switch between `Endpoint` commented out with `#`, single config (`wg0.conf`)
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
)

var (
	flagMullvad   = flag.Bool("mullvad", false, "Switch Mullvad.")
	flagOpenVPN   = flag.Bool("openvpn", false, "Switch OpenVPN.")
	flagWireGuard = flag.Bool("wireguard", false, "Switch WireGuard.")
)

func main() {
	flag.Parse()

	s, err := func() (switchable, error) {
		switch {
		case *flagMullvad:
			return mullvad()
		case *flagOpenVPN:
			return openVPN()
		case *flagWireGuard:
			return wireGuard()
		default:
			return autodetect()
		}
	}()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(serve(s))
}

type switchable interface {
	// Current returns current server, if known.
	Current() (string, error)
	// List lists available servers, if possible.
	List() ([]string, error)
	// Switch switches to the specified server, if possible.
	Switch(server string) error
}

var errNotConfigured = errors.New("not configured")

func autodetect() (switchable, error) {
	for _, f := range []func() (switchable, error){
		mullvad,
		openVPN,
		wireGuard,
	} {
		if s, err := f(); err == nil {
			log.Fatal(serve(s))
		} else if err != errNotConfigured {
			log.Fatal(err)
		}
	}
	return nil, fmt.Errorf("no supported VPN configuration found")
}
