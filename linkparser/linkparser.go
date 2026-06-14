package linkparser

import (
	"strings"

	"golang.org/x/net/html"
)

func traverseLinks(node *html.Node, collectedLinks *[]string) {
	if node == nil {
		return
	}
	if node.Type == html.ElementNode && node.Data == "a" {
		for _, attr := range node.Attr {
			if attr.Key == "href" {
				*collectedLinks = append(*collectedLinks, attr.Val)
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		traverseLinks(child, collectedLinks)
	}
}

func ParseLinks(htmlContent string) ([]string, error) {
	collectedLinks := make([]string, 0)

	root, err := html.Parse(strings.NewReader(htmlContent))

	if err != nil {
		return nil, err
	}

	traverseLinks(root, &collectedLinks)

	return collectedLinks, nil
}
