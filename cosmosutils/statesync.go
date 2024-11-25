package cosmosutils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/initia-labs/weave/client"
)

const (
	PolkachuStateSyncURL      string = "https://www.polkachu.com/%s/state_sync"
	PolkachuStateSyncPeersURL string = "https://www.polkachu.com/%s/peers"
)

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
	body, err := httpClient.Get(fmt.Sprintf(PolkachuStateSyncPeersURL, chainSlug), "", nil, nil)
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

type BlockResponse struct {
	Result struct {
		Block struct {
			Header struct {
				Height string `json:"height"`
			} `json:"header"`
		} `json:"block"`
	} `json:"result"`
}

type HashResponse struct {
	Result struct {
		BlockID struct {
			Hash string `json:"hash"`
		} `json:"block_id"`
	} `json:"result"`
}

type StateSyncInfo struct {
	TrustHeight int
	TrustHash   string
}

func GetStateSyncInfo(url string) (*StateSyncInfo, error) {
	httpClient := client.NewHTTPClient()
	var latestBlock BlockResponse
	_, err := httpClient.Get(url, "/block", nil, &latestBlock)
	if err != nil {
		return nil, fmt.Errorf("Error fetching latest block height: %v\n", err)
	}

	latestHeight, err := strconv.Atoi(latestBlock.Result.Block.Header.Height)
	if err != nil {
		return nil, fmt.Errorf("Error converting block height to integer: %v\n", err)
	}
	blockHeight := latestHeight - 2000

	var trustHashResp HashResponse
	_, err = httpClient.Get(url, "/block", map[string]string{"height": strconv.Itoa(blockHeight)}, &trustHashResp)
	if err != nil {
		return nil, fmt.Errorf("Error fetching trust hash: %v\n", err)
	}

	return &StateSyncInfo{
		TrustHeight: blockHeight,
		TrustHash:   trustHashResp.Result.BlockID.Hash,
	}, nil
}
