package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/hakobe/present/accesslogs"
	"github.com/hakobe/present/collector"
	"github.com/hakobe/present/config"
	"github.com/hakobe/present/entries"
	slackIncoming "github.com/hakobe/present/slack/incomming"
	"github.com/hakobe/present/tags"
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
	entryUrl := entry.Url()
	if config.AccesslogUrlBase != "" {
		entryUrl = config.AccesslogUrlBase + "/entry/" + strconv.Itoa(entry.ID())
	}
	err = slackIncoming.Post(
		fmt.Sprintf("<%s|%s>", entryUrl, entry.Title()),
		trim(entry.Description(), 150),
	)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	log.Printf("Entry posted.\n")
}

func postTags(db *sql.DB) {
	tags, err := tags.All(db)
	if err != nil {
		log.Printf("Tags retrieve error: %v\n", err)
		return
	}
	log.Printf("Posting tags: %s\n", strings.Join(tags, ", "))
	err = slackIncoming.Post("Watching tags: "+strings.Join(tags, ", "), "")
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	log.Printf("Tags posted.\n")
}

func addTag(db *sql.DB, tag string) {
	err := tags.Add(db, tag)
	if err != nil {
		log.Printf("Adding tag failed: %v\n", err)
		return
	}
	log.Printf("Tag added: %s\n", tag)
	postTags(db)
}

func delTag(db *sql.DB, tag string) {
	err := tags.Del(db, tag)
	if err != nil {
		log.Printf("Deleting tag failed: %v\n", err)
		return
	}
	log.Printf("Tag deleted: %s\n", tag)
	postTags(db)
}

func postRankings(db *sql.DB) {
	rankedEntries, err := entries.Rankings(db)
	if err != nil {
		log.Printf("Rankings retrieve error: %v\n", err)
		return
	}

	tmpl, err := template.New("rankings").Parse(`
Click rankings in 24 hours
{{ range $i, $e := . }}{{$e.AccessCount}} clicks <{{$e.Entry.Url}}|{{$e.Entry.Title}}> 
{{ end }}`)
	if err != nil {
		log.Printf("Rankings template error: %v\n", err)
		return
	}

	buf := bytes.NewBufferString("")
	tmpl.Execute(buf, rankedEntries)

	err = slackIncoming.Post("", buf.String())
	if err != nil {
		log.Printf("%v\n", err)
		return
	}

	log.Printf("Rankings posted.\n")
}

func postHelp() {
	text := `
plz: Post a one url.
tags: Show all watching tags.
add <tag>: Add a watching tag.
del <tag>: Delete a watching tag.
rankings: Show clicks ranking in 24 hours.
help: Show this message.
	`
	err := slackIncoming.Post("", text)
	if err != nil {
		log.Printf("%v\n", err)
		return
	}
	log.Printf("Help posted.\n")
}

func updateToSavedTags(db *sql.DB, updateTags chan<- []string) {
	tags, err := tags.All(db)
	if err != nil {
		log.Printf("Tags retrieve error: %v\n", err)
		return
	}
	updateTags <- tags
}

func main() {
	db, err := sql.Open("mysql", config.DbDsn)

	err = entries.Prepare(db)
	if err != nil {
		log.Fatalf("DB(entries) preparation error: %v\n", err)
	}
	err = tags.Prepare(db)
	if err != nil {
		log.Fatalf("DB(tags) preparation error: %v\n", err)
	}
	err = accesslogs.Prepare(db)
	if err != nil {
		log.Fatalf("DB(accesslogs) preparation error: %v\n", err)
	}

	collectedEntries, updateTags := collector.Start()
	updateToSavedTags(db, updateTags)
	webOp := web.Start(db)

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

	go func() {
		c := time.Tick(1 * time.Minute)
		for now := range c {
			if now.Hour() == config.RankingsHour && now.Minute() == 0 {
				postRankings(db)
				log.Printf("Scheduled rankings are posted.\n")
			}
		}
	}()

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
		case o := <-webOp:
			switch o.Op {
			case "humanspeaking":
				log.Printf("Humans are speaking. Go to next sleep.\n")
			case "plz":
				postNextEntry(db)
			case "fever":
				for i := 0; i < 10; i++ {
					postNextEntry(db)
				}
			case "tags":
				postTags(db)
			case "add":
				addTag(db, o.Args[0])
				updateToSavedTags(db, updateTags)
			case "del":
				delTag(db, o.Args[0])
				updateToSavedTags(db, updateTags)
			case "help":
				postHelp()
			case "rankings":
				postRankings(db)
			}

			noopCount = 0
		}
	}
}
