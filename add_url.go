package main

import (
	"fmt"
	"os"
)

func addUrl(feed_url string) error {

	fmt.Fprintf(os.Stdout, "podcasts: %v\n", feed_url)

	return nil
}
