package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type BinaryRelease struct {
	TagName     string `json:"tag_name"`
	PublishedAt string `json:"published_at"`
	Assets      []struct {
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

func fetchReleases(url string) []BinaryRelease {
	client := NewHTTPClient()
	var releases []BinaryRelease
	_, err := client.Get(url, "", nil, &releases)
	if err != nil {
		panic(fmt.Errorf("failed to fetch releases: %v", err))
	}

	return releases
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
	releases := fetchReleases(url)
	return mapReleasesToVersions(releases)
}

func GetLatestMinitiaVersion(vm string) (string, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/initia-labs/mini%s/releases", vm)
	releases := fetchReleases(url)

	if len(releases) < 1 {
		return "", "", fmt.Errorf("no releases found")
	}

	os, arch := getOSArch()
	searchString := fmt.Sprintf("%s_%s.tar.gz", os, arch)

	var latestRelease *BinaryRelease
	var downloadURL string

	for _, release := range releases {
		for _, asset := range release.Assets {
			if strings.Contains(asset.BrowserDownloadURL, searchString) {
				if latestRelease == nil || compareDates(latestRelease.PublishedAt, release.PublishedAt) {
					latestRelease = &release
					downloadURL = asset.BrowserDownloadURL
				}
				break
			}
		}
	}

	if latestRelease == nil {
		return "", "", fmt.Errorf("no compatible release found for %s_%s", os, arch)
	}

	return latestRelease.TagName, downloadURL, nil
}

// SortVersions sorts the versions based on semantic versioning, including pre-release handling
func SortVersions(versions BinaryVersionWithDownloadURL) []string {
	var versionTags []string
	for tag := range versions {
		versionTags = append(versionTags, tag)
	}

	// Sort based on major, minor, patch, and pre-release metadata
	sort.Slice(versionTags, func(i, j int) bool {
		return CompareSemVer(versionTags[i], versionTags[j])
	})

	return versionTags
}

func compareDates(d1, d2 string) bool {
	const layout = time.RFC3339

	t1, err1 := time.Parse(layout, d1)
	t2, err2 := time.Parse(layout, d2)

	if err1 != nil && err2 != nil {
		return d1 < d2 // fallback
	} else if err1 != nil {
		return false
	} else if err2 != nil {
		return true
	}

	return t1.Before(t2)
}

// compareSemVer compares two semantic version strings and returns true if v1 should be ordered before v2
func CompareSemVer(v1, v2 string) bool {
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

func GetOPInitVersions() (BinaryVersionWithDownloadURL, string) {
	versions := ListBinaryReleases("https://api.github.com/repos/initia-labs/opinit-bots/releases")
	userHome, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	binaryPath := filepath.Join(userHome, WeaveDataDirectory, OPinitAppName)
	currentVersion, _ := GetBinaryVersion(binaryPath)

	return versions, currentVersion
}
