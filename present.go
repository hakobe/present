package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hakobe/present/collector"
	slackIncoming "github.com/hakobe/present/slack/incomming"
	slackOutgoing "github.com/hakobe/present/slack/outgoing"
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

	minWait := 3
	maxWait := 20
	wait := 5
	for {
		select {
		case <-time.After(time.Duration(wait) * time.Second):
			entry := <-buffer
			log.Printf("Posting entry: %s\n", entry.Title)
			err := slackIncoming.Post(fmt.Sprintf("%s - %s", entry.Title, entry.Url))
			if err != nil {
				log.Printf("%v\n", err)
				continue
			}
			if wait > minWait {
				wait = int(float32(wait) / 1.2)
			}
			if wait <= minWait {
				wait = minWait
			}
			log.Printf("Entry posted. Please wait %ds.\n", wait)
		case <-webhookArrived:
			if wait < maxWait {
				wait = int(float32(wait) * 1.2)
			}
			if wait >= maxWait {
				wait = maxWait
			}
			log.Printf("Webhook arrived. Please wait %ds.\n", wait)
		}
	}
}
