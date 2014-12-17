package web

import (
	"log"
	"net/http"
	"os"

	slackOutgoing "github.com/hakobe/present/slack/outgoing"
)

var bind string = ":" + os.Getenv("PORT")

func Start() chan *slackOutgoing.Op {
	op := make(chan *slackOutgoing.Op, 1000)

	http.HandleFunc(
		"/hook",
		func(rw http.ResponseWriter, r *http.Request) {
			slackOutgoing.Handle(op, rw, r)
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
