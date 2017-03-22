package main

import (
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
)

func addUrl(feedUrl string) error {

	_, err := http.Head(feedUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "podcasts: %v\n\n", err)
		return err
	}

	// get the podcast list file
	usr, _ := user.Current()
	feedPath := filepath.Join(usr.HomeDir, ".podcasts", "feeds.txt")

	// if the file doesn't exist, create it with the feedURL string.
	if _, err := os.Stat(feedPath); os.IsNotExist(err) {
		err = writeText(feedUrl, feedPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "podcasts: %v\n\n", err)
			return err
		}
		return nil
	}

	err = appendText("\n"+feedUrl, feedPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "podcasts: %v\n\n", err)
		return err
	}

	return nil
}
