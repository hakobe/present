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
	"github.com/hakobe/present/web"
)

func trim(s string, l int) string {
	r := []rune(s)
	res := string(r)
	if len(r) > l && l >= 3 {
		res = string(r[0:(l-3)]) + "..."
	}
	return res
}

func postNextEntry(db *sql.DB) {
	entry, err := entries.Next(db)
	if err != nil {
		log.Printf("No entries can be retrieved: %v\n", err)
		return
	}
	log.Printf("Posting entry: %s\n", entry.Title())
	err = slackIncoming.Post(
		fmt.Sprintf("<%s|%s>", entry.Url(), entry.Title()),
		trim(entry.Description(), 150),
	)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	log.Printf("Entry posted.\n")
}

func main() {
	db, err := sql.Open("mysql", config.DbDsn)
	err = entries.Prepare(db)
	if err != nil {
		log.Fatalf("db error: %v\n", err)
	}

	collectedEntries := collector.Start()
	webOp := web.Start()

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

	noopCount := 0
	wait := config.Wait
	for {
		var timer <-chan time.Time
		if config.NoopLimit == 0 || noopCount < config.NoopLimit {
			timer = time.After(time.Duration(wait) * time.Second)
		} else {
			timer = make(chan time.Time) // block. leak?
			log.Printf("NoopLimit(%d) reached, going to long sleep...\n", config.NoopLimit)
		}
		select {
		case <-timer:
			postNextEntry(db)

			noopCount += 1
		case op := <-webOp:
			if op == "slack-humanspeaking" {
				log.Printf("Humans are speaking. Go to next sleep.\n")
			} else if op == "postnext" || op == "slack-plz" {
				postNextEntry(db)
			} else if op == "slack-fever" {
				for i := 0; i < 10; i++ {
					postNextEntry(db)
				}
			}

			noopCount = 0
		}
	}
}
