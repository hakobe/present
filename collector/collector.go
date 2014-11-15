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
		"http://b.hatena.ne.jp/search/tag?safe=on&q=perl&mode=rss",
		"http://b.hatena.ne.jp/search/tag?safe=on&q=ruby&mode=rss",
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
	XMLName xml.Name `xml:"item"`
	Title   string   `xml:"title"`
	Url     string   `xml:"link"`
	RawDate string   `xml:"http://purl.org/dc/elements/1.1/ date"`
}

func (entry RssEntry) Date() time.Time {
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

var lastUpdated map[string]time.Time = make(map[string]time.Time)
var lastUpdatedSem chan struct{} = make(chan struct{}, 1)

func updateLastUpdated(url string) {
	<-lastUpdatedSem
	lastUpdated[url] = time.Now()
	lastUpdatedSem <- struct{}{}
}

var started time.Time = time.Now()

func getLastUpdated(url string) time.Time {
	<-lastUpdatedSem
	t, exists := lastUpdated[url]
	lastUpdatedSem <- struct{}{}

	if exists {
		return t
	} else {
		return time.Time{}
	}
}

func Start() <-chan *RssEntry {
	ticker := time.Tick(10 * time.Second)
	out := make(chan *RssEntry)
	lastUpdatedSem <- struct{}{}

	go func() {
		for _ = range ticker {
			log.Println("tick!")
			for _, url := range urls() {
				go func(url string) {
					feed, err := fetchRss(url)
					checked := getLastUpdated(url)

					if err != nil {
						log.Print(err)
						return
					}
					anyUpdated := false
					for _, entry := range feed.RssEntries {
						if checked.Before(entry.Date()) {
							log.Printf("Queued entry: %s\n", entry.Title)
							anyUpdated = true
							out <- entry
						}
					}

					if anyUpdated {
						updateLastUpdated(url)
					}
				}(url)
			}
		}
	}()

	return out
}
