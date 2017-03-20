package main

import (
	"fmt"
	"net/http"
	"os"
)

func addUrl(feedUrl string) error {

	fmt.Fprintf(os.Stdout, "podcasts: %s\n\n", feedUrl)

	res, err := http.Get(feedUrl)
	if err != nil {
		return err
	}

    // http://stackoverflow.com/a/16785343
	finalURL := res.Request.URL.String()

	return nil
}
