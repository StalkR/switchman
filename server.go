package main

import (
	"flag"
	"fmt"
	"net/http"
	"text/template"
)

var flagListen = flag.String("listen", ":81", "Port to listen on for HTTP requests.")

func serve(s switchable) error {
	srv := &server{switchable: s}
	http.HandleFunc("/", srv.handleIndex)
	http.HandleFunc("/list", srv.handleList)
	http.HandleFunc("/switch", srv.handleSwitch)
	http.HandleFunc("/next", srv.handleNext)
	return http.ListenAndServe(*flagListen, nil)
}

type server struct {
	switchable
}

var indexTmpl = template.Must(template.New("").Parse(`
Current server: {{.Current}}
<br>
Servers ({{len .Servers}}):
<br>
<ul>
{{range .Servers}}<li><a href="switch?server={{.}}">{{.}}</a></li>
{{end}}
</ul>
`))

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	current, err := s.Current()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	list, err := s.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := indexTmpl.Execute(w, struct {
		Current string
		Servers []string
	}{
		Current: current,
		Servers: list,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *server) handleList(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/list" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	list, err := s.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, e := range list {
		fmt.Fprintf(w, "%v\n", e)
	}
}

// note: no xsrf protection
func (s *server) handleSwitch(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/switch" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	if err := s.Switch(r.URL.Query().Get("server")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "ok\n")
}

// note: no xsrf protection
func (s *server) handleNext(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/next" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	if err := s.Next(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "ok\n")
}
