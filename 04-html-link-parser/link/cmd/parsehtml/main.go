package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/djcurill/link"
)

func main() {
	filepath := flag.String("file", "", "html file path")
	flag.Parse()

	if *filepath == "" {
		fmt.Fprintln(os.Stderr, "error: -file flag is required")
		flag.Usage()
		os.Exit(2)
	}

	f, err := os.Open(*filepath)
	if err != nil {
		log.Fatalf("error reading file %s: %v", *filepath, err)
	}

	links, err := link.ParseHtml(f)
	if err != nil {
		log.Fatalf("An error occured parsing html: %s", err)
	}
	fmt.Fprintln(os.Stdout, "successfully parsed html links!")
	fmt.Fprintln(os.Stdout, links)
}
