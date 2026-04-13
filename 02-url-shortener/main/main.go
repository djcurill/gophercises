package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	urlshort "github.com/djcurill/gophercises/02-url-shortener"
)

func main() {
	yamlFile := flag.String("yaml", "", "yaml file path")
	jsonFile := flag.String("json", "", "json file path")
	flag.Parse()

	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := urlshort.MapHandler(pathsToUrls, mux)

	// Build the YAMLHandler using the mapHandler as the
	// fallback
	if *yamlFile != "" {
		b, err := os.ReadFile(*yamlFile)
		if err != nil {
			panic(err)
		}
		yamlHandler, err := urlshort.YAMLHandler(b, mapHandler)
		if err != nil {
			panic(err)
		}
		fmt.Println("Starting the server using YAML Handler on :8080")
		http.ListenAndServe(":8080", yamlHandler)
	} else if *jsonFile != "" {
		b, err := os.ReadFile(*jsonFile)
		if err != nil {
			panic(err)
		}
		jsonHandler, err := urlshort.JSONHandler(b, mapHandler)
		if err != nil {
			panic(err)
		}
		fmt.Println("Starting the server using JSON Handler on :8080")
		http.ListenAndServe(":8080", jsonHandler)
	} else {
		fmt.Println("starting the server on :8080")
		http.ListenAndServe(":8080", mapHandler)
	}

}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
