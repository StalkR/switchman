# Switchman

[![Build Status][build-img]][build] [![Godoc][godoc-img]][godoc]

[build]: https://github.com/StalkR/switchman/actions/workflows/build.yml
[build-img]: https://github.com/StalkR/switchman/actions/workflows/build.yml/badge.svg
[godoc]: https://godoc.org/github.com/StalkR/switchman
[godoc-img]: https://godoc.org/github.com/StalkR/switchman?status.png

switchman is a small web server to switch VPN exits. Supported VPNs:

- Mullvad: switch between servers fetched from their API, single config (`wg0.conf`)
- basic OpenVPN: switch between `remote` commented out with `;`, single config (`*.conf`)
- basic WireGuard: switch between `Endpoint` commented out with `#`, single config (`wg0.conf`)

It listens on TCP IPv4/IPv6 at the specified port.

Example:

    $ go run . -listen :81

# Setup

Clone this repo, create Debian package, install:

    $ git clone github.com/StalkR/switchman
    $ cd switchman
    $ fakeroot debian/rules clean binary
    $ sudo dpkg -i ../switchman_1-1_amd64.deb

Configure in `/etc/default/switchman` and start with `/etc/init.d/switchman start`.

# License

[Apache License, version 2.0](http://www.apache.org/licenses/LICENSE-2.0).

# Bugs, feature requests, questions

Create a [new issue](https://github.com/StalkR/switchman/issues/new).
