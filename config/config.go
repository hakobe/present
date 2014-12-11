package config

import (
	"os"
	"strings"
)

var DbDsn string = os.Getenv("PRESENT_DB_DSN")

var SlackIncomingWebhookUrl string = os.Getenv("PRESENT_SLACK_INCOMMING_URL")

var Name string
var Tags []string

func init() {
	tags := os.Getenv("PRESENT_TAGS")
	if tags != "" {
		Tags = strings.Split(tags, ",")
	}

	name := os.Getenv("PRESENT_NAME")
	if name != "" {
		Name = "present"
	}
}
