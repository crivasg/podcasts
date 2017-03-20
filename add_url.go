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

	_, err := http.Get(feedUrl)
	if err != nil {
		return err
	}

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

	err = appendText("\n"+feedUrl, feedPath)
	if err != nil {
		return err
	}

	return nil
}
