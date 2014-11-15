package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hakobe/present/collector"
	slackIncoming "github.com/hakobe/present/slack/incomming"
	slackOutgoing "github.com/hakobe/present/slack/outgoing"
)

func getRecentlyCollectedEntry( collectedEntries chan *collector.CollectedEntry ) *collector.RssEntry {
	var entry *collector.RssEntry
	for collectedEntry := range collectedEntries {
		if collectedEntry.CollectedAt.After(time.Now().Add( -3 * time.Hour )) {
			entry = collectedEntry.Entry
			break
		}
	}
	return entry
}

func main() {
	collectedEntries := collector.Start()
	webhookArrived := slackOutgoing.Start()

	buffer := make(chan *collector.CollectedEntry, 1000)
	go func() {
		for collectedEntry := range collectedEntries {
			buffer <- collectedEntry
		}
	}()

	minWait := 3 * 60
	maxWait := 20 * 60
	wait := 5 * 60
	for {
		select {
		case <-time.After(time.Duration(wait) * time.Second):
			log.Println("Getting next entry...")
			entry := getRecentlyCollectedEntry(buffer)

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
