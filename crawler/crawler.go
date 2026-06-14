package crawler

import (
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"webcrawler/export"
	"webcrawler/linkparser"
	"webcrawler/pagefetch"
	"webcrawler/ratelimiter"
	"webcrawler/robots"
	"webcrawler/urlresolver"
)

var skipExt = map[string]struct{}{
	".pdf":   {},
	".doc":   {},
	".docx":  {},
	".xls":   {},
	".xlsx":  {},
	".ppt":   {},
	".pptx":  {},
	".csv":   {},
	".txt":   {},
	".png":   {},
	".jpg":   {},
	".jpeg":  {},
	".gif":   {},
	".webp":  {},
	".svg":   {},
	".bmp":   {},
	".tiff":  {},
	".mp3":   {},
	".wav":   {},
	".aac":   {},
	".ogg":   {},
	".flac":  {},
	".mp4":   {},
	".mkv":   {},
	".mov":   {},
	".avi":   {},
	".wmv":   {},
	".webm":  {},
	".zip":   {},
	".rar":   {},
	".7z":    {},
	".tar":   {},
	".gz":    {},
	".msi":   {},
	".pkg":   {},
	".deb":   {},
	".rpm":   {},
	".dmg":   {},
	".exe":   {},
	".apk":   {},
	".iso":   {},
	".woff":  {},
	".woff2": {},
	".ttf":   {},
	".eot":   {},
	".ico":   {},
	".bin":   {},
	".dat":   {},
}

type CrawlJob struct {
	URL   string
	Depth int
}

type PageInfo struct {
	URL        string
	Depth      int
	StatusCode int
	CrawledAt  time.Time
	Error      error
}

type Crawler struct {
	urlQueue    chan CrawlJob
	visited     map[string]PageInfo
	rootHost    string
	mu          sync.Mutex
	wg          sync.WaitGroup
	done        chan struct{}
	robots      *robots.Handler
	rateLimiter *ratelimiter.Limiter
	Results     chan PageResult
	ResultsList []export.PageResult
	StartTime   time.Time
	errorCount  int
}

type PageResult struct {
	URL        string
	StatusCode int
	Depth      int
	LinksFound int
	Error      error
}

func New(workerCount int) *Crawler {
	return &Crawler{
		urlQueue:    make(chan CrawlJob, workerCount*1000),
		visited:     make(map[string]PageInfo),
		done:        make(chan struct{}),
		robots:      robots.New(),
		Results:     make(chan PageResult),
		rateLimiter: ratelimiter.New(time.Second),
	}
}

func (c *Crawler) Run(targetURL string, workerCount int, maxDepth int, silent bool) {
	c.StartTime = time.Now()
	c.robots.OpenLog("logs.txt")

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		fmt.Println(err)
	}

	c.rootHost = parsedURL.Host

	seedRobots := c.robots.Get(parsedURL)
	if seedRobots != nil {
		log.Printf("Loaded robots.txt, sitemaps: %v", seedRobots.Sitemaps)
	}

	c.visited[targetURL] = PageInfo{
		URL:   targetURL,
		Depth: 0,
	}
	c.wg.Add(1)
	c.urlQueue <- CrawlJob{
		URL:   targetURL,
		Depth: 0,
	}

	go func() {
		for r := range c.Results {
			log.Printf("%+v\n", r)
		}
	}()

	for workerID := range workerCount {
		go c.consumeQueue(workerID, maxDepth, silent)
	}

	go func() {
		c.wg.Wait()
		close(c.done)
		close(c.urlQueue)
		close(c.Results)
	}()
	<-c.done

	log.Printf("crawled %d pages, %d errors, took %v",
		len(c.ResultsList),
		c.errorCount,
		time.Since(c.StartTime).Round(time.Second),
	)

	c.robots.CloseLog()

}

func (c *Crawler) consumeQueue(workerID int, maxDepth int, silent bool) {

	for currentURL := range c.urlQueue {

		host, err := url.Parse(currentURL.URL)
		if err != nil {
			log.Println(err.Error())
		}

		c.rateLimiter.Wait(host.Host)

		data, statusCode, err := pagefetch.FetchPage(currentURL.URL)

		if err != nil {
			fmt.Printf(
				"fetch error: %s: %v\n",
				currentURL.URL,
				err,
			)
			c.mu.Lock()

			info := c.visited[currentURL.URL]
			info.Error = err
			info.CrawledAt = time.Now()

			c.errorCount++

			c.visited[currentURL.URL] = info

			c.mu.Unlock()

			c.wg.Done()
			continue
		}

		c.mu.Lock()

		info := c.visited[currentURL.URL]
		info.CrawledAt = time.Now()
		info.StatusCode = 200

		c.visited[currentURL.URL] = info

		c.mu.Unlock()

		links, err := linkparser.ParseLinks(data)
		if err != nil {
			fmt.Println("parse error", err)
			c.wg.Done()
			continue
		}

		if !silent {
			fmt.Printf("FETCH %s (depth: %d) at %v by worker %d\n",
				currentURL.URL,
				currentURL.Depth,
				time.Now().Format("15:04:05.000"),
				workerID,
			)
		}

		for _, link := range links {

			resolvedURL, err := urlresolver.ResolveURL(currentURL.URL, link)
			if err != nil {
				continue
			}
			// fmt.Printf("Resolve : %s \n", resolvedURL)

			parsedResolved, err := url.Parse(resolvedURL)
			if err != nil {
				continue
			}

			parsedResolved.RawQuery = ""
			parsedResolved.Fragment = ""

			if parsedResolved.Path == "/" {
				parsedResolved.Path = ""
			}

			resolvedURL = parsedResolved.String()

			if parsedResolved.Host != c.rootHost {
				continue
			}

			ext := strings.ToLower(filepath.Ext(resolvedURL))
			if _, ok := skipExt[ext]; ok {
				continue
			}

			if currentURL.Depth >= maxDepth {
				continue
			}

			if !c.robots.IsAllowed(parsedResolved) {
				continue
			}

			c.mu.Lock()
			_, exist := c.visited[resolvedURL]

			if !exist {
				c.wg.Add(1)
				c.mu.Unlock()

				select {
				case c.urlQueue <- CrawlJob{
					URL:   resolvedURL,
					Depth: currentURL.Depth + 1,
				}:
					c.mu.Lock()
					c.visited[resolvedURL] = PageInfo{
						URL:   resolvedURL,
						Depth: currentURL.Depth + 1,
					}
					c.mu.Unlock()

				case <-c.done:
					c.wg.Done()
					return
				}

				continue
			}
			c.mu.Unlock()
		}

		c.mu.Lock()
		c.ResultsList = append(c.ResultsList, export.PageResult{
			URL:        currentURL.URL,
			StatusCode: statusCode,
			Depth:      currentURL.Depth,
			LinksFound: len(links),
		})

		c.mu.Unlock()
		c.wg.Done()
	}
}
