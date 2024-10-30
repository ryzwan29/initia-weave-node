package utils

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	PolkachuSnapshotURL   string = "https://www.polkachu.com/testnets/%s/snapshots"
	SnapshotFileExtension string = ".tar.lz4"
)

func FetchPolkachuSnapshotDownloadURL(chainSlug string) (string, error) {
	client := NewHTTPClient()
	body, err := client.Get(fmt.Sprintf(PolkachuSnapshotURL, chainSlug), "", nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var downloadURL string
	doc.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, exists := s.Attr("href")
		if exists && isSnapshotURL(href) {
			downloadURL = href
			return false
		}
		return true
	})

	if downloadURL == "" {
		return "", fmt.Errorf("no download URL found")
	}

	return downloadURL, nil
}

func isSnapshotURL(href string) bool {
	return href != "" && href[len(href)-len(SnapshotFileExtension):] == SnapshotFileExtension
}
