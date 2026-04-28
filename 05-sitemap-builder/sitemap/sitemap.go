package sitemap

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/djcurill/sitemap/link"
)

var urlSetTemplateString = `<?xml version="1.0" encoding="UTF-8"?>
<urlset>
	{{range .Urls}}
	<url>
		<loc>{{.}}</loc>
	</url>
	{{end}}
</urlset>
`
var urlSetTemplate = template.Must(template.New("").Parse(urlSetTemplateString))

func getLinks(pageUrl *url.URL) ([]link.Link, error) {
	res, err := http.Get(pageUrl.String())
	if err != nil {
		return []link.Link{}, err
	}
	ct := res.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		return []link.Link{}, nil
	}
	links, err := link.ParseHtml(res.Body)
	if err != nil {
		return []link.Link{}, err
	}
	return links, nil
}

func GenerateUrlSet(rootUrl string) (string, error) {

	startURL, err := url.Parse(rootUrl)
	if err != nil {
		fmt.Printf("error parsing root URL: %s\n", err)
		return "", err
	}

	sm := SiteMap{Urls: []string{}, seen: map[string]bool{}, Host: startURL}
	err = sm.Crawl(startURL)
	if err != nil {
		fmt.Printf("error crawling site URLs: %s\n", err)
		return "", nil
	}

	urlSet, err := executeTemplate(&sm)

	return urlSet, err
}

type SiteMap struct {
	Urls []string
	seen map[string]bool
	Host *url.URL
}

func (p *SiteMap) Crawl(pageUrl *url.URL) error {

	stringUrl := pageUrl.String()
	if p.seen[stringUrl] {
		return nil
	}
	p.seen[stringUrl] = true

	links, err := getLinks(pageUrl)
	if err != nil {
		return err
	}

	p.Urls = append(p.Urls, stringUrl)

	for _, ln := range links {
		hrefUrl, err := p.formatHref(ln.Href)
		if err != nil {
			return err
		}
		if hrefUrl.Host != p.Host.Host {
			continue
		}
		err = p.Crawl(hrefUrl)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *SiteMap) formatHref(href string) (*url.URL, error) {
	hrefUrl, err := url.Parse(href)
	if err != nil {
		return &url.URL{}, nil
	}

	scheme := hrefUrl.Scheme
	if scheme == "" {
		scheme = p.Host.Scheme
	}
	host := hrefUrl.Host
	if host == "" {
		host = p.Host.Host
	}
	return &url.URL{Host: host, Scheme: scheme, Path: hrefUrl.Path}, nil
}

func executeTemplate(sm *SiteMap) (string, error) {
	var b bytes.Buffer
	err := urlSetTemplate.Execute(&b, sm)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
