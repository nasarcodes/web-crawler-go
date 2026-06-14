package urlresolver

import (
	"fmt"
	"net/url"
	"strings"
)

func ResolveURL(base string, href string) (string, error) {

	if href == "" {
		return "", fmt.Errorf("empty href")
	}

	loweredHref := strings.ToLower(strings.TrimSpace(href))

	switch {
	case strings.HasPrefix(loweredHref, "mailto:"),
		strings.HasPrefix(loweredHref, "tel:"),
		strings.HasPrefix(loweredHref, "javascript:"),
		strings.HasPrefix(loweredHref, "#"):
		return "", fmt.Errorf("unsupported scheme: %s", href)
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	refURL, err := url.Parse(href)
	if err != nil {
		return "", err
	}

	resolvedURL := baseURL.ResolveReference(refURL)

	resolvedURL.Scheme = strings.ToLower(resolvedURL.Scheme)
	resolvedURL.Host = strings.ToLower(resolvedURL.Host)

	resolvedURL.Fragment = ""

	if len(resolvedURL.Path) > 1 {
		resolvedURL.Path = strings.TrimSuffix(resolvedURL.Path, "/")
	}
	return resolvedURL.String(), err
}
