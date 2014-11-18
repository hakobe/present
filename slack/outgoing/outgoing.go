package outgoing

import (
	"fmt"
	"net/http"
)

func Handle(op chan string, rw http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		userId := r.FormValue("user_id")
		if userId != "USLACKBOT" {
			op <- "slackoutgoing"
		}
	}
	fmt.Fprintf(rw, "ok")
}
