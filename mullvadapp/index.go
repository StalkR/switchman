package mullvadapp

import (
  "html/template"
  "io"
)

var indexTmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width" />
  <title>switchman</title>
</head>
<body>
<p>Status</p><pre>{{.Status}}</pre>
<p>Version</p><pre>{{.Version}}</pre>
<p>Relay options</p><pre>{{.RelayOptions}}</pre>
<p>
Relays ({{len .Relays}})
</p>
<table>
  <thead>
    <tr>
      <th align="left">Country</th>
      <th align="left">City</th>
      <th align="left">Hostname</th>
      <th align="left">IPv4</th>
      <th align="left">IPv6</th>
      <th align="left">Ownership</th>
      <th align="left">Switch</th>
    </tr>
  </thead>
  <tbody>
    {{range .Relays}}
    <tr>
      <td>{{.Country}}</td>
      <td>{{.City}}</td>
      <td>{{.Hostname}}</td>
      <td>{{.IPv4}}</td>
      <td>{{.IPv6}}</td>
      <td>{{.Ownership}}</td>
      <td><a href="switch?server={{.Location}}">switch</a></td>
    </tr>
    {{end}}
  </tbody>
</table>
</body>
</html>`))

// Index writes an HTML index page to switch the Server.
func (s *Server) Index(w io.Writer) error {
  status, err := run("mullvad", "status", "-v")
  if err != nil {
    return err
  }
  version, err := run("mullvad", "version")
  if err != nil {
    return err
  }
  relayOptions, err := run("mullvad", "relay", "get")
  if err != nil {
    return err
  }

  relays, err := s.listRelays()
  return indexTmpl.Execute(w, struct {
    Status       string
    Version      string
    RelayOptions string
    Relays       []*relay
  }{
    Status:       status,
    Version:      version,
    RelayOptions: relayOptions,
    Relays:       relays,
  })
}
