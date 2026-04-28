package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
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

//go:embed templates/story.html
var templateFS embed.FS
var tmpl = template.Must(template.ParseFS(templateFS, "templates/story.html"))
var storyArcs map[string]StoryArc

func storyArcHandler(arcs map[string]StoryArc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		arc, ok := arcs[id]
		if !ok {
			http.NotFound(w, r)
		}
		err := tmpl.Execute(w, arc)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Printf("template execute error: %s", err)
		}
	}
}

func main() {
	filepath := flag.String("file", "gopher.json", "filepath to story arc json file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s <filepath>", os.Args[0])
		fmt.Fprintf(os.Stderr, "Start a choose your own adventure storyline!\n")
		fmt.Fprintf(os.Stderr, "options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	b, err := os.ReadFile(*filepath)
	if err != nil {
		log.Fatalf("could not read gopher.json: %v", err)
	}

	err = json.Unmarshal(b, &storyArcs)
	if err != nil {
		log.Fatalf("could not parse gopher.json: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/intro", http.StatusFound)
	})
	mux.HandleFunc("/{id}", storyArcHandler(storyArcs))
	log.Printf("starting server on part 8080")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Println(err)
	}

}
