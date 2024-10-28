package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

func GetEndpointURL(network, endpoint string) (string, error) {
	endpoints := GetConfig(fmt.Sprintf("constants.endpoints.%s", network))
	netEndpoints := endpoints.(map[string]interface{})
	url, ok := netEndpoints[endpoint].(string)
	if !ok {
		return "", fmt.Errorf("endpoint %s not found for network %s", endpoint, network)
	}
	return url, nil
}

func MakeGetRequestUsingConfig(network, endpoint, additionalPath string, params map[string]string, result interface{}) error {
	// TODO: Refactor this to a more generic function
	baseURL, err := GetEndpointURL(network, endpoint)
	if err != nil {
		return err
	}
	return makeGetRequest(baseURL, additionalPath, params, result)
}

func MakeGetRequestUsingURL(overrideURL, additionalPath string, params map[string]string, result interface{}) error {
	return makeGetRequest(overrideURL, additionalPath, params, result)
}

func makeGetRequest(baseURL, additionalPath string, params map[string]string, result interface{}) error {
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

type BinaryRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type BinaryVersionWithDownloadURL map[string]string

func getOSArch() (os, arch string) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

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

	return os, arch
}

func fetchReleases(url string) BinaryVersionWithDownloadURL {
	resp, err := http.Get(url)
	if err != nil {
		panic(fmt.Errorf("failed to fetch releases: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}

	var releases []BinaryRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		panic(fmt.Errorf("failed to unmarshal releases JSON: %v", err))
	}

	return mapReleasesToVersions(releases)
}

func mapReleasesToVersions(releases []BinaryRelease) BinaryVersionWithDownloadURL {
	versions := make(BinaryVersionWithDownloadURL)
	os, arch := getOSArch()
	searchString := fmt.Sprintf("%s_%s.tar.gz", os, arch)

	for _, release := range releases {
		for _, asset := range release.Assets {
			if strings.Contains(asset.BrowserDownloadURL, searchString) {
				versions[release.TagName] = asset.BrowserDownloadURL
			}
		}
	}

	return versions
}

func ListBinaryReleases(url string) BinaryVersionWithDownloadURL {
	return fetchReleases(url)
}

func GetLatestMinitiaVersion(vm string) (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/initia-labs/mini%s/releases", vm)
	versions := fetchReleases(url)

	if len(versions) < 1 {
		return "", "", fmt.Errorf("unexpected case: no binary found")
	}

	keys := make([]string, 0, len(versions))
	for k := range versions {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return compareSemVer(keys[i], keys[j])
	})

	version := keys[0]
	downloadURL := versions[version]

	return version, downloadURL, nil
}

// SortVersions sorts the versions based on semantic versioning, including pre-release handling
func SortVersions(versions BinaryVersionWithDownloadURL) []string {
	var versionTags []string
	for tag := range versions {
		versionTags = append(versionTags, tag)
	}

	// Sort based on major, minor, patch, and pre-release metadata
	sort.Slice(versionTags, func(i, j int) bool {
		return compareSemVer(versionTags[i], versionTags[j])
	})

	return versionTags
}

// compareSemVer compares two semantic version strings and returns true if v1 should be ordered before v2
func compareSemVer(v1, v2 string) bool {
	// Trim "v" prefix
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	v1Main, v1Pre := splitVersion(v1)
	v2Main, v2Pre := splitVersion(v2)

	// Compare the main (major, minor, patch) versions
	v1MainParts := strings.Split(v1Main, ".")
	v2MainParts := strings.Split(v2Main, ".")
	for i := 0; i < 3; i++ {
		v1Part, _ := strconv.Atoi(v1MainParts[i])
		v2Part, _ := strconv.Atoi(v2MainParts[i])

		if v1Part != v2Part {
			return v1Part > v2Part
		}
	}

	// Compare pre-release parts if main versions are equal
	// A pre-release version is always ordered lower than the normal version
	if v1Pre == "" && v2Pre != "" {
		return true
	}
	if v1Pre != "" && v2Pre == "" {
		return false
	}
	return v1Pre > v2Pre
}

// splitVersion separates the main version (e.g., "0.4.11") from the pre-release (e.g., "Binarytion.1")
func splitVersion(version string) (mainVersion, preRelease string) {
	if strings.Contains(version, "-") {
		parts := strings.SplitN(version, "-", 2)
		return parts[0], parts[1]
	}
	return version, ""
}
