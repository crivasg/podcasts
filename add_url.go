package main

import (
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
)

func addUrl(feedUrl string) error {

	fmt.Fprintf(os.Stdout, "podcasts: %s\n\n", feedUrl)

	res, err := http.Get(feedUrl)
	if err != nil {
		return err
	}

	// http://stackoverflow.com/a/16785343
	finalURL := res.Request.URL.String()
	fmt.Fprintf(os.Stdout, "podcasts: %s\n\n", finalURL)

	// get the podcast list file
	usr, _ := user.Current()
	feedPath := filepath.Join(usr.HomeDir, ".podcasts", "feeds.txt")

	// if the file doesn't exist, create it with the feedURL string.
	if _, err := os.Stat(feedPath); os.IsNotExist(err) {
		err = writeText(feedUrl, feedPath)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}
