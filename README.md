# Web Crawler

A concurrent web crawler written in Go with rate limiting, robots.txt compliance, and structured data export.

## Features

- **Concurrent crawling** -- configurable worker pool (1-25 goroutines)
- **Per-host rate limiting** -- 1 request per second per host to avoid overwhelming servers
- **robots.txt compliance** -- respects `robots.txt` directives with in-memory caching per host
- **Depth-limited crawling** -- control how many link levels to follow from the seed URL
- **Same-domain restriction** -- stays within the seed host, no off-site crawling
- **URL normalization** -- lowercases host/scheme, strips fragments, query strings, and trailing slashes
- **Binary and asset filtering** -- automatically skips common file extensions (images, fonts, archives, executables, documents)
- **Unsupported scheme filtering** -- ignores `mailto:`, `tel:`, `javascript:`, and fragment-only links
- **Structured export** -- outputs crawl results in JSON, CSV, or both
- **Duplicate detection** -- visited URL tracking prevents re-crawling the same page
- **Graceful shutdown** -- workers exit cleanly when the queue is exhausted

## Installation

```bash
git clone https://github.com/your-username/webcrawler.git
cd webcrawler
go mod download
```

## Usage

```bash
go run main.go -url <target-url> [flags]
```

### Flags

| Flag     | Type   | Default | Description                              |
|----------|--------|---------|------------------------------------------|
| `-url`   | string | (required) | Seed URL to start crawling from        |
| `-depth` | int    | `3`     | Maximum crawl depth (minimum 1)          |
| `-worker`| int    | `5`     | Number of concurrent workers (1-25)      |
| `-silent`| bool   | `false` | Suppress per-page fetch log output       |
| `-export`| string | `both`  | Export format: `json`, `csv`, or `both`  |

### Examples

Crawl `example.com` with default settings (depth 3, 5 workers):
```bash
go run main.go -url example.com
```

Deep crawl with 20 workers and JSON output:
```bash
go run main.go -url https://example.com -depth 5 -worker 20 -export json
```

Quiet mode with CSV output:
```bash
go run main.go -url example.com -silent -export csv
```

### Output Files

| File                | Description                                     |
|---------------------|-------------------------------------------------|
| `<host>_data.json`  | Crawl results in JSON format with indentation   |
| `<host>_data.csv`   | Crawl results in CSV format with header row     |
| `logs.txt`          | URLs blocked by robots.txt directives           |

## Architecture

```
webcrawler/
├── main.go                # Entry point and CLI flag parsing
├── crawler/
│   └── crawler.go         # Core engine: worker pool, URL queue, visited tracking
├── pagefetch/
│   └── pagefetch.go       # HTTP GET with custom headers and timeout
├── linkparser/
│   └── linkparser.go      # HTML parsing and <a href> extraction
├── urlresolver/
│   └── urlresolver.go     # Relative-to-absolute URL resolution and normalization
├── ratelimiter/
│   └── ratelimiter.go     # Per-host request rate limiter
├── robots/
│   └── robots.go          # robots.txt fetching, caching, and compliance checks
├── export/
│   └── export.go          # JSON and CSV export for crawl results
└── go.mod                 # Module definition and dependencies
```

### Package Responsibilities

| Package        | Responsibility                                                |
|----------------|---------------------------------------------------------------|
| `crawler`      | Orchestrates worker goroutines, manages the URL queue and visited map, collects results, and logs summary statistics |
| `pagefetch`    | Performs HTTP GET requests with Chrome user-agent headers and a 10-second timeout; returns body and status code |
| `linkparser`   | Parses HTML content using `golang.org/x/net/html`, walks the DOM tree, and collects all `href` attribute values from `<a>` elements |
| `urlresolver`  | Resolves a relative `href` against a base URL to produce an absolute URL; normalizes scheme, host, path; filters unsupported schemes |
| `ratelimiter`  | Maintains per-host timestamps and enforces a configurable minimum interval between requests to the same host |
| `robots`       | Fetches and parses `robots.txt` per host, caches results in memory, and checks whether a given URL path is allowed by the rules. Writes blocked URLs to a log file |
| `export`       | Serializes the crawl result list to JSON (pretty-printed) and CSV (with headers) |

## Data Flow

1. **CLI** (`main.go`) parses flags and constructs a `Crawler` instance
2. **Seed URL** is normalized (scheme prepended if missing) and enqueued at depth 0
3. **Worker goroutines** consume from the URL queue:
   a. Apply per-host rate limiting
   b. Fetch the page via `pagefetch`
   c. Parse links via `linkparser`
   d. For each link: resolve, normalize, filter by host/ext/depth, check robots.txt
   e. Enqueue new URLs and record results
4. **Shutdown** triggers when the queue drains and all workers finish; a completion signal unblocks the main goroutine
5. **Export** writes the accumulated results to the requested file formats
6. **Summary** prints total pages crawled, error count, and elapsed time

## Crawl Result Schema

Each crawled page produces a result record with the following fields:

```json
{
  "url": "https://example.com/page",
  "status_code": 200,
  "depth": 1,
  "links_found": 42,
  "error": null
}
```

## Dependencies

| Dependency               | Purpose                     |
|--------------------------|-----------------------------|
| `golang.org/x/net/html`  | HTML tokenization and parsing |
| `github.com/temoto/robotstxt` | robots.txt parsing       |

## License

MIT
