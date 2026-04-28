package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/djcurill/cyoa"
)

func main() {
	filepath := flag.String("file", "gopher.json", "JSON file containing story")
	flag.Parse()

	f, err := os.Open(*filepath)
	if err != nil {
		log.Fatalf("unable to open file %s: %s", *filepath, err)
	}

	story, err := cyoa.JsonStory(f)
	if err != nil {
		log.Fatalf("error parsing json file: %s", err)
	}

	handler := cyoa.NewHandler(story)
	fmt.Println("starting server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
