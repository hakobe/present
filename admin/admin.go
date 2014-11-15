package admin

import (
	"fmt"
	"net/http"
)

func ListenAndServe() {
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "hello")
	})
	http.Handle("/static/", http.FileServer(http.Dir("public")))
	http.ListenAndServe(":8080", nil)
}
