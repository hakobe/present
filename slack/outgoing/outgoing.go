package outgoing

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var bind string = ":" + os.Getenv("PORT")

func Start() chan struct{} {
	out := make(chan struct{}, 1000)

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			userId := r.FormValue("user_id")
			if userId != "USLACKBOT" {
				log.Println("ping!!")
				out <- struct{}{}
			}
		}
		fmt.Fprintf(rw, "ok")
	})

	go func() {
		log.Printf("Starting slack webhook on \"%s\"\n", bind)
		err := http.ListenAndServe(bind, nil)
		if err != nil {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	return out
}
