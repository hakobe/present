package outgoing

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hakobe/present/config"
)

type Op struct {
	Op   string
	Args []string
}

func isNameMatched(name string) bool {
	matched := false
	for _, n := range config.Names {
		if name == n {
			matched = true
			break
		}
	}
	return matched
}

func Handle(op chan *Op, rw http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		userId := r.FormValue("user_id")
		text := strings.TrimSpace(r.FormValue("text"))
		texts := regexp.MustCompile("\\s+").Split(text, -1)

		if userId != "USLACKBOT" {
			if len(texts) > 1 && isNameMatched(texts[0]) {
				switch texts[1] {
				case "plz":
					op <- &Op{"plz", nil}
				case "fever":
					op <- &Op{"fever", nil}
				case "tags":
					op <- &Op{"tags", nil}
				case "add":
					op <- &Op{"add", []string{strings.Join(texts[2:], " ")}}
				case "del":
					op <- &Op{"del", []string{strings.Join(texts[2:], " ")}}
				case "help":
					op <- &Op{"help", nil}
				}
			} else {
				op <- &Op{"humanspeaking", nil}
			}
		}
	}
	fmt.Fprintf(rw, "ok")
}
