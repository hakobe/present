package config

import (
	"os"
)

var DbDsn string = os.Getenv("PRESENT_DB_DSN")

var SlackIncomingWebhookUrl string = os.Getenv("PRESENT_SLACK_INCOMMING_URL")
