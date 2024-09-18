package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"
)

func GetEndpointURL(network, endpoint string) (string, error) {
	endpoints := GetConfig(fmt.Sprintf("endpoints.%s", network))
	netEndpoints := endpoints.(map[string]interface{})
	url, ok := netEndpoints[endpoint].(string)
	if !ok {
		return "", fmt.Errorf("endpoint %s not found for network %s", endpoint, network)
	}
	return url, nil
}

func MakeGetRequest(network, endpoint, additionalPath string, params map[string]string, result interface{}) error {
	baseURL, err := GetEndpointURL(network, endpoint)
	if err != nil {
		return err
	}

	fullURL := fmt.Sprintf("%s%s", baseURL, additionalPath)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	query := req.URL.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode JSON response: %v", err)
	}

	return nil
}

type InitiaRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type InitiaVersionWithDownloadURL map[string]string

func ListInitiaReleases() InitiaVersionWithDownloadURL {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var os, arch string
	switch goos {
	case "darwin":
		os = "Darwin"
	case "linux":
		os = "Linux"
	default:
		panic(fmt.Errorf("unsupported OS: %s", goos))
	}
	switch goarch {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	default:
		panic(fmt.Errorf("unsupported architecture: %s", goarch))
	}

	url := "https://api.github.com/repos/initia-labs/initia/releases"
	resp, err := http.Get(url)
	if err != nil {
		panic(fmt.Errorf("failed to fetch releases: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Errorf("failed to read response body: %v", err))
	}

	var releases []InitiaRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		panic(fmt.Errorf("failed to unmarshal JSON: %v", err))
	}

	searchString := fmt.Sprintf("%s_%s.tar.gz", os, arch)
	versions := make(InitiaVersionWithDownloadURL)

	for _, release := range releases {
		for _, asset := range release.Assets {
			if strings.Contains(asset.BrowserDownloadURL, searchString) {
				versions[release.TagName] = asset.BrowserDownloadURL
			}
		}
	}

	return versions
}
