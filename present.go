package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/hakobe/present/collector"
	"github.com/hakobe/present/config"
	"github.com/hakobe/present/entries"
	slackIncoming "github.com/hakobe/present/slack/incomming"
	slackOutgoing "github.com/hakobe/present/slack/outgoing"
)

func main() {
	db, err := sql.Open("mysql", config.DbDsn)
	err = entries.Prepare(db)
	if err != nil {
		log.Fatalf("db error: %v\n", err)
	}

	collectedEntries := collector.Start()
	webhookArrived := slackOutgoing.Start()

	go func() {
		for entry := range collectedEntries {
			err = entries.Add(db, entry)
			if err != nil {
				log.Printf("SQL error: %v\n", err)
				continue
			}
		}
	}()
	entries.StartCleaner(db)

	wait := 20 * 60
	for {
		select {
		case <-time.After(time.Duration(wait) * time.Second):
			entry, err := entries.Next(db)
			if err != nil {
				log.Printf("No entries can be retrieved: %v\n", err)
				continue
			}
			log.Printf("Posting entry: %s\n", entry.Title())
			err = slackIncoming.Post(fmt.Sprintf("%s - %s", entry.Title(), entry.Url()))
			if err != nil {
				log.Printf("%v\n", err)
				continue
			}
			log.Printf("Entry posted. Please wait %ds for next.\n", wait)
		case <-webhookArrived:
			log.Printf("Webhook arrived. Please wait %ds for next.\n", wait)
		}
	}
}
