package link

import (
	"errors"
	"io"
	"strings"

	"golang.org/x/net/html"
)

type Link struct {
	Href string
	Text string
}

func ParseHtml(r io.Reader) ([]Link, error) {
	doc, err := html.Parse(r)
	links := []Link{}

	if err != nil {
		return links, err
	}

	nodes := linkNodes(doc)
	for _, n := range nodes {
		lnk, err := buildLink(n)
		if err != nil {
			return links, err
		}
		links = append(links, lnk)
	}

	return links, nil
}

func buildLink(n *html.Node) (Link, error) {
	href := ""
	text := []string{}

	for _, a := range n.Attr {
		if a.Key == "href" {
			href = a.Val
			break
		}
	}

	for c := range n.Descendants() {
		if c.Type == html.TextNode && c.Data != "" {
			text = append(text, strings.TrimSpace(c.Data))
		}
	}

	if href == "" {
		return Link{}, errors.New("Invalid link node, empty href")
	}

	return Link{
		Href: href,
		Text: strings.Join(text, " "),
	}, nil
}

func linkNodes(root *html.Node) []*html.Node {
	stack := []*html.Node{root}
	lns := []*html.Node{}

	for len(stack) > 0 {
		currentNode := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if currentNode.Type == html.ElementNode && currentNode.Data == "a" {
			lns = append(lns, currentNode)
			continue
		}
		for c := currentNode.LastChild; c != nil; c = c.PrevSibling {
			stack = append(stack, c)
		}
	}
	return lns
}
