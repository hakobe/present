package outgoing

import (
	"fmt"
	"net/http"
	"os"
)

var bind string = ":" + os.Getenv("PRESENT_SLACK_OUTGOING_PORT")

func Start() chan string {
	out := make(chan string, 1000)

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		out <- "ping"
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
