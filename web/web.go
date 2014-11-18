package web

import (
	"fmt"
	"log"
	"net/http"
	"os"

	slackOutgoing "github.com/hakobe/present/slack/outgoing"
)

var bind string = ":" + os.Getenv("PORT")

func Start() chan string {
	op := make(chan string, 1000)

	http.HandleFunc(
		"/hook",
		func(rw http.ResponseWriter, r *http.Request) {
			slackOutgoing.Handle(op, rw, r)
		},
	)

	http.HandleFunc(
		"/postnext",
		func(rw http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				op <- "postnext"
				fmt.Fprint(rw, "ok")
			} else {
				rw.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprint(rw, "405 Method Not Allowed")
			}
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
