package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/djcurill/sitemap"
)

func main() {
	siteUrl := flag.String("url", "", "website url")
	flag.Parse()

	if *siteUrl == "" {
		fmt.Fprintf(os.Stderr, "sitemap generator requires an input url")
		os.Exit(1)
	}

	urlSet, err := sitemap.GenerateUrlSet(*siteUrl)
	if err != nil {
		panic(err)
	}

	fmt.Println("URL Set generation complete")
	fmt.Println(urlSet)
}
