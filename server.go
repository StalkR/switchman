package main

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"sort"
	"strings"
)

func serve(s Switchable, listen string) error {
	srv := &server{s}
	http.HandleFunc("/", srv.handleIndex)
	http.HandleFunc("/switch", srv.handleSwitch)
	http.HandleFunc("/next", srv.handleNext)
	return http.ListenAndServe(listen, nil)
}

type server struct {
	Switchable
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf8")
	if err := index(w, s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

var indexTmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width" />
  <title>switchman</title>
</head>
<body>
Current server: {{.Current}}
<br>
Servers ({{len .Servers}}):
<br>
<ul>
{{range .Servers}}<li><a href="switch?server={{.}}">{{.}}</a></li>{{end}}
</ul>
</body>
</html>`))

// Indexable allows implementations to provide a custom index page instead of
// the default showing a list of servers.
type Indexable interface {
	// Index produces an HTML index.
	Index(w io.Writer) error
}

func index(w io.Writer, s *server) error {
	if i, ok := s.Switchable.(Indexable); ok {
		return i.Index(w)
	}
	current, err := s.Current()
	if err != nil {
		return err
	}
	servers, err := s.List()
	if err != nil {
		return err
	}
	sort.Strings(servers)
	return indexTmpl.Execute(w, struct {
		Current string
		Servers []string
	}{
		Current: current,
		Servers: servers,
	})
}

// note: no xsrf protection
func (s *server) handleSwitch(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/switch" {
		http.NotFound(w, r)
		return
	}
	if err := s.Switch(r.URL.Query().Get("server")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !acceptsHTML(r) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "ok\n")
		return
	}
	fmt.Fprint(w, "<script>window.location=document.referrer;</script>")
}

// note: no xsrf protection
func (s *server) handleNext(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/next" {
		http.NotFound(w, r)
		return
	}
	if err := next(s); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !acceptsHTML(r) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "ok\n")
		return
	}
	fmt.Fprint(w, "<script>window.location=document.referrer;</script>")
}

func next(s Switchable) error {
	current, err := s.Current()
	if err != nil {
		return err
	}
	servers, err := s.List()
	if err != nil {
		return err
	}
	var next string
	for i, e := range servers {
		if e == current {
			next = servers[(i+1)%len(servers)]
			break
		}
	}
	if next == "" {
		return fmt.Errorf("could not find next server")
	}
	return s.Switch(next)
}

func acceptsHTML(r *http.Request) bool {
	for _, e := range strings.Split(r.Header.Get("Accept"), ",") {
		if len(e) == 0 {
			continue
		}
		media := strings.Split(e, ";")[0]
		if strings.ToLower(media) == "text/html" {
			return true
		}
	}
	return false
}
