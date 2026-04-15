package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
)

type StoryArc struct {
	Title      string            `json:"title"`
	Paragraphs []string          `json:"story"`
	Options    []StoryArcOptions `json:"options"`
}

type StoryArcOptions struct {
	Text string `json:"text"`
	Arc  string `json:"arc"`
}

func main() {
	var storyArcs map[string]StoryArc
	b, err := os.ReadFile("gopher.json")
	if err != nil {
		log.Fatalf("could not read gopher.json: %v", err)
	}

	err = json.Unmarshal(b, &storyArcs)
	if err != nil {
		log.Fatalf("could not parse gopher.json: %v", err)
	}

	tmpl := template.Must(template.ParseFiles("templates/story.html"))

	err = tmpl.Execute(os.Stdout, storyArcs["intro"])
	if err != nil {
		fmt.Println(err)
	}

}
