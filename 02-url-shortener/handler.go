package urlshort

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"gopkg.in/yaml.v3"
)

type UrlMapping struct {
	Path string `yaml:"path" json:"path"`
	Url  string `yaml:"url" json:"url"`
}

type Format int

const (
	FormatYAML Format = iota
	FormatJSON
)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	f := func(w http.ResponseWriter, r *http.Request) {
		shortUrl, ok := pathsToUrls[string(r.URL.Path)]
		if ok {
			log.Printf("%s -> %s", r.URL.Path, shortUrl)
			http.Redirect(w, r, shortUrl, http.StatusFound)
		} else {
			log.Println("url not found in map, fallingback")
			fallback.ServeHTTP(w, r)
		}
	}
	return http.HandlerFunc(f)
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//   - path: /some-path
//     url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	urlMappings, err := BuildMap(yml, FormatYAML)
	if err != nil {
		return nil, err
	}
	return MapHandler(urlMappings, fallback), nil
}

func JSONHandler(json []byte, fallback http.Handler) (http.HandlerFunc, error) {
	urlMappings, err := BuildMap(json, FormatJSON)
	if err != nil {
		return nil, err
	}
	return MapHandler(urlMappings, fallback), nil
}

func BuildMap(b []byte, format Format) (map[string]string, error) {
	var data []UrlMapping
	var err error
	urlMappings := make(map[string]string)

	switch format {
	case FormatJSON:
		err = json.Unmarshal(b, &data)
	case FormatYAML:
		err = yaml.Unmarshal(b, &data)
	default:
		err = errors.New("unrecognized file format provided")
	}

	if err != nil {
		return urlMappings, err
	}

	for _, urlMap := range data {
		urlMappings[urlMap.Path] = urlMap.Url
	}
	return urlMappings, nil

}
