package incomming

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
)

func webhookUrl() string {
	return os.Getenv("PRESENT_SLACK_INCOMMING_URL")
}

type payload struct {
	Text string `json:"text"`
}

func Post(message string) error {
	p, err := json.Marshal(payload{Text: message})
	if err != nil {
		return err
	}
	_, err = http.PostForm(webhookUrl(), url.Values{
		"payload": []string{string(p)},
	})
	if err != nil {
		return err
	}
	return nil
}
