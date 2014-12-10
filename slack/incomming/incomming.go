package incomming

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/hakobe/present/config"
)

var webhookUrl string = os.Getenv("PRESENT_SLACK_INCOMMING_URL")

type field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type attachment struct {
	Fallback string   `json:"fallback"`
	Pretext  string   `json:"pretext"`
	Color    string   `json:"color"`
	Fields   []*field `json:"fields"`
}

type payload struct {
	Attachments []*attachment `json:"attachments"`
}

func Post(title string, description string) error {
	p, err := json.Marshal(&payload{
		Attachments: []*attachment{
			&attachment{
				Fallback: title,
				Pretext:  title,
				Fields: []*field{
					&field{
						Title: "",
						Value: description,
						Short: false,
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	_, err = http.PostForm(config.SlackIncomingWebhookUrl, url.Values{
		"payload": []string{string(p)},
	})
	if err != nil {
		return err
	}
	return nil
}
