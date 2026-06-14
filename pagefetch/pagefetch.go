package pagefetch

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func FetchPage(targetURL string) (string, int, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
		"Connection":      "keep-alive",
		"Cache-Control":   "no-cache",
	}

	req, err := http.NewRequest("GET", targetURL, nil)

	if err != nil {
		return "", -1, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", -1, err
	}
	if resp.StatusCode != http.StatusOK {
		return "", resp.StatusCode, fmt.Errorf("status %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", resp.StatusCode, err
	}

	return string(body), resp.StatusCode, err
}
