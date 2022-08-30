// Binary switchman is a small web server to switch OpenVPN/WireGuard servers.
package main

import (
	"errors"
	"flag"
	"log"
)

func main() {
	flag.Parse()

	for _, f := range []func() (switchable, error){
		openVPN,
		wireGuard,
	} {
		if s, err := f(); err == nil {
			log.Fatal(serve(s))
		} else if err != errNotConfigured {
			log.Fatal(err)
		}
	}
	log.Fatal("nothing configured")
}

type switchable interface {
	// Current returns current server, if known.
	Current() (string, error)
	// List lists available servers, if possible.
	List() ([]string, error)
	// Switch switches to the specified server, if possible.
	Switch(server string) error
	// Next switches to the next server.
	Next() error
}

var errNotConfigured = errors.New("not configured")
