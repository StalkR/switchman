package mullvad

import (
  "html/template"
  "io"
  "sort"
)

// TODO: multihop server selection
var indexTmpl = template.Must(template.New("").Parse(`
Current server: {{.Current}}
{{if eq (len .CurrentRelays) 2}}
<ul>
  <li>{{$ := index .CurrentRelays 0}}entry: {{$.Country}}, {{$.City}}, {{if $.Owned}}owned{{else}}rented{{end}}</li>
  <li>{{$ := index .CurrentRelays 1}}exit: {{$.Country}}, {{$.City}}, {{if $.Owned}}owned{{else}}rented{{end}}</li>
</ul>
{{end}}
{{if eq (len .CurrentRelays) 1}}
<ul>
  <li>{{$ := index .CurrentRelays 0}}{{$.Country}}, {{$.City}}, {{if $.Owned}}owned{{else}}rented{{end}}
</ul>
{{end}}
<br>
Servers ({{len .Relays}}):
<br>
<table>
  <thead>
    <tr>
      <th>Country</th>
      <th>City</th>
      <th>Owned</th>
      <th>Switch</th>
    </tr>
  </thead>
  <tbody>
{{range .Relays}}
  <tr>
    <td>{{.Country}}</td>
    <td>{{.City}}</td>
    <td>{{if .Owned}}owned{{else}}rented{{end}}</td>
    <td><a href="switch?server={{.Hostname}}:{{.Port}}">switch</a></td>
  </tr>
{{end}}
  </tbody>
</table>
<br>
`))

// Index writes an HTML index page to switch the Server.
func (s *Server) Index(w io.Writer) error {
  current, err := s.Current()
  if err != nil {
    return err
  }
  currentRelays, err := s.findRelays(current)
  if err != nil {
    currentRelays = nil
  }
  relays, err := s.listRelays()
  if err != nil {
    return err
  }
  sort.Slice(relays, func(i, j int) bool {
    if relays[i].Country == relays[j].Country {
      if relays[i].City == relays[j].City {
        return relays[i].Hostname < relays[j].Hostname
      }
      return relays[i].City < relays[j].City
    }
    return relays[i].Country < relays[j].Country
  })
  return indexTmpl.Execute(w, struct {
    Current       string
    CurrentRelays []relay
    Relays        []relay
  }{
    Current:       current,
    CurrentRelays: currentRelays,
    Relays:        relays,
  })
}
