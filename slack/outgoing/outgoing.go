package outgoing

import (
	"fmt"
	"net/http"
	"os"
)

var bind string = ":" + os.Getenv("PRESENT_SLACK_OUTGOING_PORT")

func Start() chan struct{} {
	out := make(chan struct{}, 1000)

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			userId := r.FormValue("user_id")
			if userId != "USLACKBOT" {
				out <- struct{}{}
			}
		}
		fmt.Fprintf(rw, "ok")
	})

	go func() {
		fmt.Printf("Starting slack webhook on \"%s\"\n", bind)
		err := http.ListenAndServe(bind, nil)
		if err != nil {
			fmt.Printf("ListenAndServe: %v", err)
		}
	}()

	return out
}
