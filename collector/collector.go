package collector

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func urls() []string {
	return []string{
		"http://b.hatena.ne.jp/hakobe932/rss",
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
	XMLName xml.Name `xml:"RDF"`
	Url string
	Title string `xml:"channel>title"`
	RssEntries []RssEntry `xml:"item"`
}

type RssEntry struct {
	XMLName xml.Name `xml:"item"`
	Title string `xml:"title"`
	Url string `xml:"link"`
	RawDate string `xml:"http://purl.org/dc/elements/1.1/ date"`
}

func (entry RssEntry) Date() time.Time {
	t, err := time.Parse("2006-01-02T15:04:05-07:00", entry.RawDate)
	if err != nil {
		fmt.Println(err)
	}
	return t
}

func parseRss(data []byte) (*RssFeed, error) {
	var v RssFeed
	fmt.Printf("%s", data)
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

func StartLoop() <-chan *RssEntry {
	ticker := time.Tick(5 * time.Second)
	out := make(chan *RssEntry)

	go func() {
		for _ = range ticker {
			for _, url := range urls() {
				go func(url string) { feed, err := fetchRss(url)
					if err != nil {
						fmt.Print(err)
						return
					}
					for i, entry := range feed.RssEntries {
						fmt.Printf("%v\n", i)
						out <- &entry
					}
				}(url)
			}
		}
	}()

	return out
}
