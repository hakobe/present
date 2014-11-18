package collector

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func urls() []string {
	return []string{
		"http://b.hatena.ne.jp/search/tag?safe=on&q=aws&mode=rss",
		"http://b.hatena.ne.jp/search/tag?safe=on&q=docker&mode=rss",
		"http://b.hatena.ne.jp/search/tag?safe=on&q=linux&mode=rss",
		"http://b.hatena.ne.jp/search/tag?safe=on&q=http&mode=rss",
		"http://b.hatena.ne.jp/search/tag?safe=on&q=perl&mode=rss",
		"http://b.hatena.ne.jp/search/tag?safe=on&q=ruby&mode=rss",
		"http://b.hatena.ne.jp/search/tag?safe=on&q=vim&mode=rss",
		"http://b.hatena.ne.jp/search/tag?safe=on&q=emacs&mode=rss",
		"http://b.hatena.ne.jp/search/tag?safe=on&q=javascript&mode=rss",
		"http://b.hatena.ne.jp/search/tag?safe=on&q=golang&mode=rss",
		"http://b.hatena.ne.jp/search/tag?safe=on&q=scala&mode=rss",
		"http://b.hatena.ne.jp/search/tag?safe=on&q=java&mode=rss",
	}
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

var seenUrls map[string]bool = make(map[string]bool)
var seenUrlsSem chan struct{} = make(chan struct{}, 1)

func setSeen(url string) {
	<-seenUrlsSem
	seenUrls[url] = true
	seenUrlsSem <- struct{}{}
}

func hasSeen(url string) bool {
	<-seenUrlsSem
	t, exists := seenUrls[url]
	seenUrlsSem <- struct{}{}

	if exists && t {
		return true
	} else {
		return false
	}
}

func Start() <-chan *RssEntry {
	ticker := time.Tick(30 * time.Minute)
	out := make(chan *RssEntry)
	seenUrlsSem <- struct{}{}

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
						if !hasSeen(entry.Url()) {
							log.Printf("Queued entry: %s\n", entry.Title())
							setSeen(entry.Url())
							out <- entry
						}
					}
				}(url)
			}
		}
	}()

	return out
}
