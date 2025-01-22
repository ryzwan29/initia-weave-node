package cosmosutils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/initia-labs/weave/client"
)

const (
	PolkachuBaseURL string = "https://polkachu.com"

	PolkachuSnapshotURL  string = "https://www.polkachu.com/%s/snapshots"
	PolkachuStateSyncURL        = PolkachuBaseURL + "/%s/state_sync"
	PolkachuPeersURL            = PolkachuBaseURL + "/%s/peers"

	SnapshotFileExtension string = ".tar.lz4"
)

func FetchPolkachuSnapshotDownloadURL(chainSlug string) (string, error) {
	httpClient := client.NewHTTPClient()
	body, err := httpClient.Get(fmt.Sprintf(PolkachuSnapshotURL, chainSlug), "", nil, nil)
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

func FetchPolkachuStateSyncURL(chainSlug string) (string, error) {
	httpClient := client.NewHTTPClient()
	body, err := httpClient.Get(fmt.Sprintf(PolkachuStateSyncURL, chainSlug), "", nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var stateSyncURL string
	doc.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		href, exists := s.Attr("href")
		if exists && isLikelyStateSyncURL(href) {
			stateSyncURL = href
			return false
		}
		return true
	})

	if stateSyncURL == "" {
		return "", fmt.Errorf("no state sync URL found")
	}

	return stateSyncURL, nil
}

func isLikelyStateSyncURL(href string) bool {
	return strings.Contains(href, "rpc") && strings.Contains(href, "polkachu.com")
}

func FetchPolkachuStateSyncPeers(chainSlug string) (string, error) {
	httpClient := client.NewHTTPClient()
	body, err := httpClient.Get(fmt.Sprintf(PolkachuPeersURL, chainSlug), "", nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var stateSyncPeer string
	titleFound := false
	doc.Find("body *").Each(func(i int, s *goquery.Selection) {
		text := s.Text()

		if strings.Contains(text, "Polkachu State-Sync Peer") {
			titleFound = true
			return
		}

		if titleFound {
			s.Find("a[data-cfemail]").Each(func(j int, link *goquery.Selection) {
				dataEmail := link.AttrOr("data-cfemail", "")
				if dataEmail != "" {
					decodedEmail, err := decodeCFEmail(dataEmail)
					if err != nil {
						panic(err)
					}

					portText := strings.TrimSpace(link.Parent().Text())
					portText = strings.ReplaceAll(portText, "[email protected]", "")
					portText = strings.TrimSpace(strings.TrimPrefix(portText, ":"))

					if strings.Contains(portText, ":") {
						parts := strings.Split(portText, ":")
						portText = strings.TrimSpace(parts[len(parts)-1])
					}

					stateSyncPeer = fmt.Sprintf("%s:%s", decodedEmail, portText)
				}
			})
		}
	})

	if stateSyncPeer == "" {
		return "", fmt.Errorf("no state-sync peer found")
	}

	return stateSyncPeer, nil
}

func FetchPolkachuPersistentPeers(chainSlug string) (string, error) {
	httpClient := client.NewHTTPClient()
	body, err := httpClient.Get(fmt.Sprintf(PolkachuPeersURL, chainSlug), "", nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var peers []string
	doc.Find("pre code").Each(func(i int, s *goquery.Selection) {
		emailMatches := s.Find("a[data-cfemail]")
		emailMatches.Each(func(j int, link *goquery.Selection) {
			dataEmail := link.AttrOr("data-cfemail", "")
			if dataEmail != "" {
				decodedEmail, err := decodeCFEmail(dataEmail)
				if err != nil {
					panic(err)
				}

				portText := strings.TrimSpace(link.Parent().Text())
				portText = strings.ReplaceAll(portText, "[email protected]", "")
				portText = strings.TrimSpace(strings.TrimPrefix(portText, ":"))
				if strings.Contains(portText, ":") {
					parts := strings.Split(portText, ":")
					portText = strings.TrimSpace(parts[len(parts)-1])
				}

				peers = append(peers, fmt.Sprintf("%s:%s", decodedEmail, portText))
			}
		})
	})

	if len(peers) == 0 {
		return "", fmt.Errorf("no peers found")
	}

	return strings.Join(peers, ","), nil
}

func decodeCFEmail(encoded string) (string, error) {
	if len(encoded) < 2 {
		return "", fmt.Errorf("encoded email is too short")
	}

	r, err := strconv.ParseInt(encoded[:2], 16, 0)
	if err != nil {
		return "", fmt.Errorf("failed to decode prefix: %w", err)
	}

	var email string
	for i := 2; i < len(encoded); i += 2 {
		b, err := strconv.ParseInt(encoded[i:i+2], 16, 0)
		if err != nil {
			return "", fmt.Errorf("failed to decode email character: %w", err)
		}
		email += string(rune(b ^ r))
	}

	return email, nil
}
