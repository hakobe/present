package collector

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/hakobe/present/config"
)

func urls() []string {
	urls := []string{}

	for _, tag := range config.Tags {
		queries := url.Values{}
		queries.Add("safe", "on")
		queries.Add("mode", "rss")
		queries.Add("q", tag)
		urls = append(urls, "http://b.hatena.ne.jp/search/tag?"+queries.Encode())
	}
	return urls
}

func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

type RssFeed struct {
	XMLName    xml.Name `xml:"RDF"`
	Url        string
	Title      string      `xml:"channel>title"`
	RssEntries []*RssEntry `xml:"item"`
}

type RssEntry struct {
	XMLName  xml.Name `xml:"item"`
	RawTitle string   `xml:"title"`
	RawUrl   string   `xml:"link"`
	RawDate  string   `xml:"http://purl.org/dc/elements/1.1/ date"`
}

func (entry *RssEntry) Title() string {
	return entry.RawTitle
}

func (entry *RssEntry) Url() string {
	return entry.RawUrl
}

func (entry *RssEntry) Date() time.Time {
	t, err := time.Parse("2006-01-02T15:04:05-07:00", entry.RawDate)
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func parseRss(data []byte) (*RssFeed, error) {
	var v RssFeed
	err := xml.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func fetchRss(url string) (*RssFeed, error) {
	data, err := fetch(url)
	if err != nil {
		return nil, err
	}
	feed, err := parseRss(data)
	if err != nil {
		return nil, err
	}
	feed.Url = url

	return feed, nil
}

func Start() <-chan *RssEntry {
	ticker := time.Tick(30 * time.Minute)
	out := make(chan *RssEntry)

	go func() {
		for _ = range ticker {
			for _, url := range urls() {
				go func(url string) {
					feed, err := fetchRss(url)

					if err != nil {
						log.Print(err)
						return
					}
					for _, entry := range feed.RssEntries {
						log.Printf("Queued entry: %s\n", entry.Title())
						out <- entry
					}
				}(url)
			}
		}
	}()

	return out
}
