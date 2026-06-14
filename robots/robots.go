package robots

import (
	"net/url"
	"os"
	"sync"

	"webcrawler/pagefetch"

	"github.com/temoto/robotstxt"
)

type Handler struct {
	cache   map[string]*robotstxt.RobotsData
	mu      sync.RWMutex
	logFile *os.File
	logMu   sync.Mutex
}

func New() *Handler {
	return &Handler{
		cache: make(map[string]*robotstxt.RobotsData),
	}
}

func (h *Handler) OpenLog(path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	h.logFile = f
	return nil
}

func (h *Handler) CloseLog() {
	if h.logFile != nil {
		h.logFile.Close()
	}
}

func (h *Handler) Get(u *url.URL) *robotstxt.RobotsData {
	host := u.Host

	h.mu.RLock()
	data, exists := h.cache[host]
	h.mu.RUnlock()

	if exists {
		return data
	}

	robotsURL := u.Scheme + "://" + u.Host + "/robots.txt"

	body, _, err := pagefetch.FetchPage(robotsURL)
	if err != nil {
		return nil
	}

	parsed, err := robotstxt.FromBytes([]byte(body))
	if err != nil {
		return nil
	}

	h.mu.Lock()
	h.cache[host] = parsed
	h.mu.Unlock()

	return parsed
}

func (h *Handler) FetchSeed(u *url.URL) *robotstxt.RobotsData {
	robotsURL := u.Scheme + "://" + u.Host + "/robots.txt"

	data, _, err := pagefetch.FetchPage(robotsURL)
	if err != nil {
		return nil
	}

	parsed, err := robotstxt.FromBytes([]byte(data))
	if err != nil {
		return nil
	}

	return parsed
}

func (h *Handler) IsAllowed(u *url.URL) bool {
	robots := h.Get(u)

	if robots == nil {
		return true
	}

	group := robots.FindGroup("*")

	if !group.Test(u.Path) {
		h.logMu.Lock()
		if h.logFile != nil {
			h.logFile.WriteString(u.String() + "\n")
		}
		h.logMu.Unlock()
		return false
	}

	return true
}
