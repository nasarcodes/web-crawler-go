package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type PageResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Depth      int    `json:"depth"`
	LinksFound int    `json:"links_found"`
	Error      error  `json:"error,omitempty"`
}

func ExportCSV(data []PageResult, host string) error {
	filename := host + "_data.csv"
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		return fmt.Errorf("create file: %w\n", err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"url", "status_code", "depth", "links_found", "error"}); err != nil {
		return fmt.Errorf("write header: %w\n", err)
	}
	for _, r := range data {
		errStr := ""
		if r.Error != nil {
			errStr = r.Error.Error()
		}
		row := []string{
			r.URL,
			strconv.Itoa(r.StatusCode),
			strconv.Itoa(r.Depth),
			strconv.Itoa(r.LinksFound),
			errStr,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("Write row: %w\n", err)
		}
	}
	return writer.Error()
}

func ExportJSON(data []PageResult, host string) error {
	fileName := host + "_data.json"
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")

	return enc.Encode(data)
}
