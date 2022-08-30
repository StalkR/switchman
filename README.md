# Switchman #

[![Build Status][build-img]][build] [![Godoc][godoc-img]][godoc]

[build]: https://github.com/StalkR/switchman/actions/workflows/build.yml
[build-img]: https://github.com/StalkR/switchman/actions/workflows/build.yml/badge.svg
[godoc]: https://godoc.org/github.com/StalkR/switchman
[godoc-img]: https://godoc.org/github.com/StalkR/switchman?status.png

A small web server to switch OpenVPN/WireGuard servers (remote/endpoints), so
it can change the corresponding exits.

It listens on TCP IPv4/IPv6 at the specified port.

Example:

    $ go run . -listen :81

# Setup #

Install go package, create Debian package, install:

    $ go get -u github.com/StalkR/switchman
    $ cd $GOPATH/src/github.com/StalkR/switchman
    $ fakeroot debian/rules clean binary
    $ sudo dpkg -i ../switchman_1-1_amd64.deb

Configure in `/etc/default/switchman` and start with `/etc/init.d/switchman start`.

# License #

[Apache License, version 2.0](http://www.apache.org/licenses/LICENSE-2.0).

# Bugs, feature requests, questions #

Create a [new issue](https://github.com/StalkR/switchman/issues/new).
