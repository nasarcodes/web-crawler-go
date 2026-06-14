package main

import (
	"flag"
	"log"
	"strings"

	"webcrawler/crawler"
	"webcrawler/export"
)

func main() {
	urlFlag := flag.String("url", "", "URL to crawl")
	depthFlag := flag.Int("depth", 3, "Max Depth")
	workerFlag := flag.Int("worker", 5, "Max Workers")
	silentFlag := flag.Bool("silent", false, "Suppress fetch logs")
	exportFlag := flag.String("export", "both", "Choose export file format")
	flag.Parse()

	if *urlFlag == "" {
		log.Fatal("Usage: go run main.go -url <url>")
	}

	if *depthFlag < 1 {
		log.Fatal("Depth can't be less than 1")
	}

	if *workerFlag < 1 || *workerFlag > 25 {
		log.Fatal("Worker should be in range 1-25")
	}

	targetURL := *urlFlag
	maxDepth := *depthFlag
	totalWorkers := *workerFlag

	if !(strings.HasPrefix(targetURL, "https://")) && !(strings.HasPrefix(targetURL, "http://")) {
		targetURL = "https://" + targetURL
	}

	c := crawler.New(totalWorkers)
	c.Run(targetURL, totalWorkers, maxDepth, *silentFlag)

	targetURL = strings.Replace(targetURL, "https://", "", 1)

	targetURL = strings.Replace(targetURL, "http://", "", 1)

	switch *exportFlag {

	case "json":
		errJSON := export.ExportJSON(c.ResultsList, targetURL)
		if errJSON != nil {
			log.Printf("\n%v\n", errJSON)
		}
	case "csv":
		errCSV := export.ExportCSV(c.ResultsList, targetURL)
		if errCSV != nil {
			log.Printf("\n%v\n", errCSV)
		}

	default:
		errCSV := export.ExportCSV(c.ResultsList, targetURL)
		if errCSV != nil {
			log.Printf("\n%v\n", errCSV)
		}

		errJSON := export.ExportJSON(c.ResultsList, targetURL)
		if errJSON != nil {
			log.Printf("\n%v\n", errJSON)
		}
	}
}
