package config

import (
	"os"
	"strings"
	"strconv"
)

var DbDsn string = os.Getenv("PRESENT_DB_DSN")

var SlackIncomingWebhookUrl string = os.Getenv("PRESENT_SLACK_INCOMMING_URL")

var Name string
var Tags []string
var Wait int = 15 * 60

func init() {
	tags := os.Getenv("PRESENT_TAGS")
	if tags != "" {
		Tags = strings.Split(tags, ",")
	}

	name := os.Getenv("PRESENT_NAME")
	if name != "" {
		Name = name
	}

	wait := os.Getenv("PRESENT_WAIT")
	if w, err := strconv.Atoi(wait); err == nil {
		Wait = w
	}
}
