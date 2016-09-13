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
	"time"
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

	return fmt.Sprintf("##\n# %s\n# %s\n%s\n##", c.Title, c.Link, buf.String())
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

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func GenerateFeedListFile(path string) error {
	s := make([]string, 2)
	s[0] = "http://feeds.5by5.tv/master"
	s[1] = "http://feed.thisamericanlife.org/talpodcast"

	err1 := writeLines(s, path)
	return err1
}

func GetFeedList() ([]string, string, error) {

	usr, _ := user.Current()
	path, _ := os.Getwd()
	directories := strings.Split(path, "/")
	feed_path := filepath.Join(usr.HomeDir, "."+directories[len(directories)-1], "feeds.txt")

	if _, err := os.Stat(feed_path); os.IsNotExist(err) {
		// http://stackoverflow.com/a/12518877
		GenerateFeedListFile(feed_path)
	}

	lines, err := readLines(feed_path)

	return lines, feed_path, err

}

// https://github.com/jteeuwen/go-pkg-rss/blob/master/timeparser.go
func ParseTime(formatted string) (time.Time, error) {
	var layouts = [...]string{
		"Mon, _2 Jan 2006 15:04:05 MST",
		"Mon, _2 Jan 2006 15:04:05 -0700",
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		"Mon, 2, Jan 2006 15:4",
		"02 Jan 2006 15:04:05 MST",
	}
	var t time.Time
	var err error
	formatted = strings.TrimSpace(formatted)
	for _, layout := range layouts {
		t, err = time.Parse(layout, formatted)
		if !t.IsZero() {
			break
		}
	}
	return t, err
}

func main() {

	t := time.Now()
	fmt.Println(t.Format(time.RFC1123))

	feed_list, feed_path, _ := GetFeedList()
	fmt.Println(feed_path)

	for _, feed_url := range feed_list {

		channel, err := GetPodcastData(feed_url)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(channel)

		for _, item := range channel.Items {
			fmt.Println(item)
			parsed, t1_err := ParseTime(item.PubDate)
			if t1_err != nil {
				continue
			}

			fmt.Println("--- ", parsed.String())
			fmt.Println("--- ", parsed.Location())

			for _, encl := range item.Enclosures {

				filename, err := GetFileName(encl.String())
				if err != nil {
					filename = ""
					continue
				}
				fmt.Println("wget -O " + filename + " " + encl.String())
				fmt.Println("#")

			}
		}
	}

}

// http://siongui.github.io/2015/03/03/go-parse-web-feed-rss-atom/
// https://github.com/jbub/podcasts
