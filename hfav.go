package main

import (
	"fmt"
	"github.com/hakobe/hfav/collector"
	slackIncoming "github.com/hakobe/hfav/slack/incomming"
	slackOutgoing "github.com/hakobe/hfav/slack/outgoing"
)

func main() {
	entries := collector.Start()
	webhookArrived := slackOutgoing.Start()
	buffer := make(chan *collector.RssEntry, 1000)

	go func() {
		for entry := range entries {
			buffer <- entry
		}
	}()

	for entry := range buffer {
		fmt.Printf("waiting ping\n")
		<- webhookArrived
		fmt.Printf("%v\n", entry)
		err := slackIncoming.Post(fmt.Sprintf("%s - %s", entry.Title, entry.Url))
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}
	}
}
