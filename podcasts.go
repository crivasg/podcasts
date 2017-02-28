package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/xml"
	"flag"
	"fmt"
	"go/doc"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var numOfDays = flag.Int("days", 1, "Number of days back to download an episode")
var outputFile = flag.String("output", ``, "Path of the output file")

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
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	PubDate       string `xml:"pubDate"`
	Items         []Item `xml:"item"`
	LastBuildDate string `xml:"lastBuildDate"`
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
		curr_line := strings.Trim(scanner.Text(), "\t ")
		match, _ := regexp.MatchString("^htt(p|ps)://", curr_line)
		if match == true {
			lines = append(lines, curr_line)
		}
	}
	return lines, scanner.Err()
}

// writetext writes the lines to the given file.
func writeText(text string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	fmt.Fprintln(w, text)
	return w.Flush()
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
	//path, _ := os.Getwd()
	//directories := strings.Split(path, "/")
	feed_path := filepath.Join(usr.HomeDir, ".podcasts", "feeds.txt")
	//fmt.Println(path)
	//fmt.Println(feed_path)

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

func podcast_fetch(url string, dirname string, days int, ch chan<- string) {

	start := time.Now()

	now := time.Now().UTC()
	channel, err := GetPodcastData(url)
	printed_channel := 0

	//resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprint(err) // send to channel ch
		return
	}

	feed_array := []string{"## " + channel.Title, "## URL: " + channel.Link,
		"## " + channel.LastBuildDate}
	for _, item := range channel.Items {

		parsed, t1_err := ParseTime(item.PubDate)
		if t1_err != nil {
			continue
		}

		parsed = parsed.UTC()
		diff := now.Sub(parsed)

		if diff.Hours() > float64(days)*24.0 {
			break
		}

		if printed_channel == 0 {
			//fmt.Println(channel)
			feed_array = append(feed_array, channel.String())
			printed_channel = 1
		}

		//fmt.Println(item)
		feed_array = append(feed_array, item.String(), "#")

		for _, encl := range item.Enclosures {

			filename, err := GetFileName(encl.String())
			if err != nil {
				filename = ""
				continue
			}
			//fmt.Println("wget -O " + filename + " " + encl.String())
			//fmt.Println("#")
			feed_array = append(feed_array, "wget -O "+filename+" "+encl.String(), "#")

		}

	}

	feed_array = append(feed_array, "#\n")

	h := sha1.New()
	h.Write([]byte(url))
	bs := h.Sum(nil)
	filename := fmt.Sprintf("%x.feed", bs)
	filepath := filepath.Join(dirname, filename)

	w, err := os.Create(filepath)
	if err != nil {
		ch <- fmt.Sprintf("couldn't create %s: %v\n", filepath, err)
		return
	}

	defer w.Close()
	text := strings.Join(feed_array, "\n")

	nbytes, err1 := w.WriteString(text)

	if err1 != nil {
		ch <- fmt.Sprintf("while reading %s: %v\n", url, err1)
		return
	}

	secs := time.Since(start).Seconds()
	ch <- fmt.Sprintf("%9.2fs : %-8d : %10x : %-25s : %s", secs, nbytes, bs[0:10], channel.Title, url)

}

func deleteFiles(path string, f os.FileInfo, err error) error {
	if filepath.Ext(path) == ".feed" {
		//fmt.Printf("%s\n", path)
		err_os := os.Remove(path)
		if err_os != nil {
			return err_os
		}
	}
	return nil
}

func mergeDataOfFiles(folder string, extension string) string {

	// https://golang.org/pkg/io/ioutil/#ReadAll
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return ""
	}
	//feed_files := []string{}
	feed_text := "#\n# --------------------------------------- PODCASTS ---------------------------------------\n#\n"
	for _, file := range files {
		match, _ := regexp.MatchString(".feed$", file.Name())
		if match == true {

			//http://stackoverflow.com/questions/36111777/golang-how-to-read-a-text-file
			//http://stackoverflow.com/questions/13078314/combine-absolute-path-and-relative-path-to-get-a-new-absolute-path
			b, err := ioutil.ReadFile(filepath.Join(folder, file.Name())) // just pass the file name
			if err != nil {
				fmt.Print(err)
				continue
			}

			match, _ := regexp.MatchString("wget\\s-O", string(b))

			if match == true {
				feed_text += string(b) + "\n"
			}
		}
	}

	return feed_text

}

func main() {

	flag.Parse()

	feed_list, feed_path, _ := GetFeedList()
	feed_data_folder := filepath.Dir(feed_path)

	// delete the .feed files if they exist
	//feedExtensions := []string{".feed"}
	err_walker := filepath.Walk(feed_data_folder, deleteFiles)
	if err_walker != nil {
		log.Fatal(err_walker)
	}

	start := time.Now()
	ch := make(chan string)
	fmt.Printf("%10s : %8s : %20s : %-25s : %s\n", "secs", "nbytes", "sha1", "Title", "URL")

	for _, url := range feed_list {
		go podcast_fetch(url, feed_data_folder, *numOfDays, ch) // start a goroutine
	}

	for range feed_list {
		fmt.Println(<-ch)
	}

	fmt.Printf("\n%6.2fs elapsed\n\n", time.Since(start).Seconds())

	feed_text := mergeDataOfFiles(feed_data_folder, ".feed")

	if len(*outputFile) == 0 {
		fmt.Printf(feed_text)
	} else {
		writeText(feed_text, *outputFile)
	}

	// delete the .feed files if they exist
	//feedExtensions := []string{".feed"}
	// clean up before finishing...
	err_walker = filepath.Walk(feed_data_folder, deleteFiles)
	if err_walker != nil {
		log.Fatal(err_walker)
	}

}

// http://siongui.github.io/2015/03/03/go-parse-web-feed-rss-atom/
// https://github.com/jbub/podcasts
