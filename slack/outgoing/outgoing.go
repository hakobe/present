package outgoing

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/hakobe/present/config"
)

func Handle(op chan string, rw http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		userId := r.FormValue("user_id")
		text := strings.TrimSpace(r.FormValue("text"))
		if userId != "USLACKBOT" {
			if strings.Index(text, config.Name) == 0 && strings.Index(text, "plz") != -1 {
				op <- "slack-plz"
			} else if strings.Index(text, config.Name) == 0 && strings.Index(text, "fever") != -1 {
				op <- "slack-fever"
			} else {
				op <- "slack-humanspeaking"
			}
		}
	}
	fmt.Fprintf(rw, "ok")
}
