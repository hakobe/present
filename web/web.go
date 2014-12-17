package web

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/hakobe/present/entries"
	slackOutgoing "github.com/hakobe/present/slack/outgoing"
)

var bind string = ":" + os.Getenv("PORT")

func Start(db *sql.DB) chan *slackOutgoing.Op {
	op := make(chan *slackOutgoing.Op, 1000)

	http.HandleFunc(
		"/hook",
		func(rw http.ResponseWriter, r *http.Request) {
			slackOutgoing.Handle(op, rw, r)
		},
	)

	http.HandleFunc(
		"/upcommings",
		func(rw http.ResponseWriter, r *http.Request) {
			es, err := entries.Upcommings(db)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
				return
			}
			tmpl, err := template.New("upcommings").Parse(`
<html>
<body>
  <ul>
	{{ range . }}
	<li><a href="{{.Url}}">{{.Title}}</a></li>
	{{ end }}
  </ul>
</body>
</html>
		`)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}

			tmpl.Execute(rw, es)
		},
	)

	go func() {
		log.Printf("Starting slack webhook on \"%s\"\n", bind)
		err := http.ListenAndServe(bind, nil)
		if err != nil {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	return op
}
