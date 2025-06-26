package wireguard

import (
	"os"
)

func New() (*server, error) {
	const config = "/etc/wireguard/wg0.conf"
	if _, err := os.Stat(config); err != nil {
		return nil, err
	}
	return &server{
		config: config,
	}, nil
}

type server struct {
	config string
}
