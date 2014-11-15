package main

import (
	"fmt"
	"github.com/hakobe/hfav/admin"
	"github.com/hakobe/hfav/collector"
	slackIncoming "github.com/hakobe/hfav/slack/incomming"
)

func main() {
	go admin.ListenAndServe()
	entries := collector.StartLoop()

	for entry := range entries {
		fmt.Printf("%v\n", entry)
		err := slackIncoming.Post(fmt.Sprintf("%s - %s", entry.Title, entry.Url))
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
	}
}
