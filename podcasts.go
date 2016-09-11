package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"go/doc"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	rssXmlns        = "http://www.itunes.com/dtds/podcast-1.0.dtd"
	rssVersion      = "2.0"
	PARAGRAPH_WIDTH = 82
)

type Rss2 struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr,omitempty"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Items       []Item `xml:"item"`
}

func (c Channel) String() string {

	desc := strings.TrimSpace(StripTags(c.Description))

	var buf bytes.Buffer
	doc.ToText(&buf, strings.TrimSpace(desc), "# ", "", PARAGRAPH_WIDTH)
	//fmt.Println(buf.String())

	return fmt.Sprintf("##\n# %s\n# %s\n# %s\n##", c.Title, c.Link, buf.String())
}

type Item struct {
	Title       string      `xml:"title"`
	Link        string      `xml:"link"`
	Guid        string      `xml:"guid"`
	PubDate     string      `xml:"pubDate"`
	Author      string      `xml:"author"`
	Description string      `xml:"description"`
	Enclosures  []Enclosure `xml:"enclosure"`
}

func (i Item) String() string {

	desc := strings.TrimSpace(StripTags(i.Description))

	var buf bytes.Buffer
	doc.ToText(&buf, desc, "# ", "", PARAGRAPH_WIDTH)

	desc = buf.String()
	return fmt.Sprintf("# Title: %s\n# PubDate: %s\n# GUID: %s\n#%s", strings.TrimSpace(i.Title),
		i.PubDate, strings.TrimSpace(i.Guid), strings.TrimSpace(desc))
}

type Enclosure struct {
	Url    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

func (e Enclosure) String() string {

	encl, err := StripUrl(e.Url)
	if err != nil {
		return fmt.Sprintf("%s", e.Url)
	}
	return fmt.Sprintf("%s", encl)
}

func GetFileName(uu string) (string, error) {

	u, err := url.Parse(uu)
	if err != nil {
		return "", err
	}

	slice1 := strings.Split(u.Path, "/")
	return slice1[len(slice1)-1], nil

}

func StripUrl(uu string) (string, error) {

	u, err := url.Parse(uu)
	if err != nil {
		return "", err
	}

	result := u.Scheme + "://" + u.Host + u.Path
	return result, nil
}

func GetPodcastData(feed_url string) (Channel, error) {

	res, err := http.Get(feed_url)
	if err != nil {
		return Channel{}, err
	}

	if res.StatusCode != http.StatusOK {
		return Channel{}, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Channel{}, err
	}

	var feed Rss2
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		return Channel{}, err
	}

	return feed.Channel, nil
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func GetFeedList() ([]string, string, error) {

	usr, _ := user.Current()
	path, _ := os.Getwd()
	directories := strings.Split(path, "/")
	feed_path := filepath.Join(usr.HomeDir, "."+directories[len(directories)-1], "feeds.txt")

	lines, err := readLines(feed_path)

	return lines, feed_path, err

}

func main() {

	feed_list, feed_path, _ := GetFeedList()

	for _, feed_url := range feed_list {

		channel, err := GetPodcastData(feed_url)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(channel)

		for _, item := range channel.Items {
			fmt.Println(item)
			for _, encl := range item.Enclosures {

				filename, err := GetFileName(encl.String())
				if err != nil {
					filename = ""
				}
				fmt.Println("wget -O " + filename + " " + encl.String())
				fmt.Println("#")

			}
		}
	}

}

// http://siongui.github.io/2015/03/03/go-parse-web-feed-rss-atom/
// https://github.com/jbub/podcasts
