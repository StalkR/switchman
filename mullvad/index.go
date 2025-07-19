package mullvad

import (
  "html/template"
  "io"
  "sort"
)

var indexTmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width" />
  <title>switchman</title>
  <script>
  document.addEventListener('DOMContentLoaded', () => {
    const elEntry = document.getElementById('entry');
    const elExit = document.getElementById('exit');
    const elSwitch = document.getElementById('switch');
    function updateSwitch(event) {
      if (!elEntry.value || !elExit.value) return;
      elSwitch.href = 'switch?server=' + encodeURIComponent(elEntry.value + ':' + elExit.value);
    }
    elEntry.addEventListener('change', updateSwitch);
    elExit.addEventListener('change', updateSwitch);
  });
  </script>
</head>
<body>
<p>Current server: {{.Current}}</p>
{{if eq (len .CurrentRelays) 2}}
<ul>
  <li>{{with index .CurrentRelays 0}}entry: {{.ID}} ({{.Country}}, {{.City}}, {{if .Owned}}owned{{else}}rented{{end}}{{end}})</li>
  <li>{{with index .CurrentRelays 1}}exit: {{.ID}} ({{.Country}}, {{.City}}, {{if .Owned}}owned{{else}}rented{{end}}{{end}})</li>
</ul>
{{end}}
{{if eq (len .CurrentRelays) 1}}
<ul>
  <li>{{with index .CurrentRelays 0}}{{.ID}} ({{.Country}}, {{.City}}, {{if .Owned}}owned{{else}}rented{{end}}){{end}}</li>
</ul>
{{end}}
{{if .LastError}}
<p>Error fetching server list: {{.LastError}}</p>
{{end}}
<form>
  Multihop:
  entry <select id="entry" name="entry">
    <option value="">-</option>
    {{range .Relays}}
    <option value="{{.Hostname}}">{{.ID}} ({{.Country}}, {{.City}}, {{if .Owned}}owned{{else}}rented{{end}})</option>
    {{end}}
  </select>
  exit <select id="exit" name="exit">
    <option value="">-</option>
    {{range .Relays}}
    <option value="{{.MultihopPort}}">{{.ID}} ({{.Country}}, {{.City}}, {{if .Owned}}owned{{else}}rented{{end}})</option>
    {{end}}
  </select>
  <a id="switch" href="#">switch</a>
</form>
<p>
Servers ({{len .Relays}})
</p>
<table>
  <thead>
    <tr>
      <th align="left">ID</th>
      <th align="left">Country</th>
      <th align="left">City</th>
      <th align="left">Ownership</th>
      <th align="left">Active</th>
      <th align="left">Switch</th>
    </tr>
  </thead>
  <tbody>
    {{range .Relays}}
    <tr>
      <td>{{.ID}}</td>
      <td>{{.Country}}</td>
      <td>{{.City}}</td>
      <td>{{if .Owned}}owned{{else}}rented{{end}}</td>
      <td>{{if .Active}}active{{else}}<span style="color: red;">inactive</span>{{end}}</td>
      <td><a href="switch?server={{.Hostname}}:{{.Port}}">switch</a></td>
    </tr>
    {{end}}
  </tbody>
</table>
</body>
</html>`))

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
  relays, lastError := s.listRelays()
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
    LastError     error
  }{
    Current:       current,
    CurrentRelays: currentRelays,
    Relays:        relays,
    LastError:     lastError,
  })
}
