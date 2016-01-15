package collector

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

func feedUrl(tag string) string {
	queries := url.Values{}
	queries.Add("safe", "on")
	queries.Add("mode", "rss")
	queries.Add("users", "5")
	queries.Add("q", tag)
	return "http://b.hatena.ne.jp/search/tag?" + queries.Encode()
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
	XMLName        xml.Name `xml:"item"`
	RawTitle       string   `xml:"title"`
	RawUrl         string   `xml:"link"`
	RawDescription string   `xml:"description"`
	RawDate        string   `xml:"http://purl.org/dc/elements/1.1/ date"`
	tag            string
}

func (entry *RssEntry) ID() int {
	return -1
}

func (entry *RssEntry) Title() string {
	return entry.RawTitle
}

func (entry *RssEntry) Url() string {
	return entry.RawUrl
}

func (entry *RssEntry) Description() string {
	return entry.RawDescription
}

func (entry *RssEntry) Date() time.Time {
	t, err := time.Parse("2006-01-02T15:04:05-07:00", entry.RawDate)
	if err != nil {
		log.Fatal(err)
	}
	return t
}

func (entry *RssEntry) Tag() string {
	return entry.tag
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

func Start() (<-chan *RssEntry, chan<- []string) {
	ticker := time.Tick(10 * time.Minute)
	out := make(chan *RssEntry)
	newTags := make(chan []string)

	collect := func(tags []string, out chan *RssEntry) {
		for _, tag := range tags {
			go func(tag string) {
				feed, err := fetchRss(feedUrl(tag))

				if err != nil {
					log.Print(err)
					return
				}
				for _, entry := range feed.RssEntries {
					log.Printf("Queued entry: %s\n", entry.Title())
					entry.tag = tag
					out <- entry
				}
			}(tag)
		}
	}

	go func() {
		tags := make([]string, 0)
		for {
			select {
			case <-ticker:
				collect(tags, out)
			case ts := <-newTags:
				tags = ts
				log.Printf("New tags: %s\n", tags)
				collect(tags, out)
			}
		}
	}()

	return out, newTags
}
