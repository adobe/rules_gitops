// update_versions downloads kustomize release information from GitHub
// and generates a versions.bzl file with SHA256 digests for each platform.
//
// Usage:
//
//	go run update_versions.go
//
// The script expects BUILD_WORKSPACE_DIRECTORY to be set, or writes to
// the default location relative to this file's directory.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	releasesURL = "https://api.github.com/repos/kubernetes-sigs/kustomize/releases"
	outputFile  = "kustomize/private/versions/versions.bzl"
)

// Asset represents a GitHub release asset
type Asset struct {
	Name              string `json:"name"`
	Digest            string `json:"digest"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Release represents a GitHub release
type Release struct {
	TagName     string  `json:"tag_name"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

// VersionInfo holds platform -> sha256 mappings for a version
type VersionInfo struct {
	Platforms map[string]string
}

func main() {
	log.SetFlags(0)

	// Fetch releases from GitHub
	releases, err := fetchReleases()
	if err != nil {
		log.Fatalf("Failed to fetch releases: %v", err)
	}

	// Parse releases and extract version info
	versions, latestVersion := parseReleases(releases)

	// Generate versions.bzl content
	content := generateBzl(versions, latestVersion)

	// Determine output path
	outputPath := getOutputPath()

	// Write the file
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	log.Printf("Successfully wrote %s with %d versions", outputPath, len(versions))
}

func fetchReleases() ([]Release, error) {
	// GitHub API may paginate, so we need to fetch all pages
	var allReleases []Release
	url := releasesURL + "?per_page=100"

	for url != "" {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "rules_gitops-version-updater")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
		}

		var releases []Release
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}
		allReleases = append(allReleases, releases...)

		// Check for pagination
		url = getNextPageURL(resp.Header.Get("Link"))
	}

	return allReleases, nil
}

func getNextPageURL(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}
	// Parse Link header: <url>; rel="next", <url>; rel="last"
	for _, part := range strings.Split(linkHeader, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, `rel="next"`) {
			// Extract URL between < and >
			start := strings.Index(part, "<")
			end := strings.Index(part, ">")
			if start >= 0 && end > start {
				return part[start+1 : end]
			}
		}
	}
	return ""
}

func parseReleases(releases []Release) (map[string]*VersionInfo, string) {
	versions := make(map[string]*VersionInfo)

	// Regex to match kustomize version tags
	versionRe := regexp.MustCompile(`^kustomize/v(\d+\.\d+\.\d+)$`)
	// Regex to extract platform from asset name
	assetRe := regexp.MustCompile(`^kustomize_v[\d.]+_([a-z]+_[a-z0-9]+)\.tar\.gz$`)

	// Track latest version by publish date
	latestVersion := ""
	latestPublishedAt := ""

	for _, release := range releases {
		matches := versionRe.FindStringSubmatch(release.TagName)
		if matches == nil {
			continue // Not a kustomize release
		}
		version := matches[1]

		versionInfo := &VersionInfo{
			Platforms: make(map[string]string),
		}

		for _, asset := range release.Assets {
			assetMatches := assetRe.FindStringSubmatch(asset.Name)
			if assetMatches == nil {
				continue
			}
			platform := assetMatches[1]

			// Extract SHA256 from digest (format: "sha256:...")
			sha256 := strings.TrimPrefix(asset.Digest, "sha256:")
			if sha256 == "" {
				continue
			}

			versionInfo.Platforms[platform] = sha256
		}

		if len(versionInfo.Platforms) > 0 {
			versions[version] = versionInfo

			// Track latest by publish date (ISO 8601 format is lexicographically sortable)
			if release.PublishedAt > latestPublishedAt {
				latestPublishedAt = release.PublishedAt
				latestVersion = version
			}
		}
	}

	return versions, latestVersion
}

func generateBzl(versions map[string]*VersionInfo, latestVersion string) string {
	var sb strings.Builder

	// Sort versions for consistent output (reverse alphabetical puts newer versions first)
	versionList := make([]string, 0, len(versions))
	for v := range versions {
		versionList = append(versionList, v)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(versionList)))

	sb.WriteString(`"""Generated by update_versions.go - do not edit manually."""

`)
	sb.WriteString(fmt.Sprintf("LATEST_KUSTOMIZE_VERSION = %q\n\n", latestVersion))
	sb.WriteString("VERSIONS = {\n")

	for _, version := range versionList {
		info := versions[version]
		sb.WriteString(fmt.Sprintf("    %q: {\n", version))

		// Sort platforms for consistent output
		platforms := make([]string, 0, len(info.Platforms))
		for p := range info.Platforms {
			platforms = append(platforms, p)
		}
		sort.Strings(platforms)

		for _, platform := range platforms {
			sha256 := info.Platforms[platform]
			sb.WriteString(fmt.Sprintf("        %q: %q,\n", platform, sha256))
		}
		sb.WriteString("    },\n")
	}

	sb.WriteString("}\n")

	return sb.String()
}

func getOutputPath() string {
	// Check for BUILD_WORKSPACE_DIRECTORY environment variable
	if wsDir := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); wsDir != "" {
		return filepath.Join(wsDir, outputFile)
	}

	// Fallback: find workspace root by looking for MODULE.bazel or WORKSPACE
	// Start from executable directory and walk up
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	dir := filepath.Dir(exePath)

	// Also try current working directory
	cwd, _ := os.Getwd()
	for _, startDir := range []string{dir, cwd} {
		d := startDir
		for d != "/" && d != "." {
			if _, err := os.Stat(filepath.Join(d, "MODULE.bazel")); err == nil {
				return filepath.Join(d, outputFile)
			}
			if _, err := os.Stat(filepath.Join(d, "WORKSPACE")); err == nil {
				return filepath.Join(d, outputFile)
			}
			d = filepath.Dir(d)
		}
	}

	log.Fatalf("Could not find workspace root. Set BUILD_WORKSPACE_DIRECTORY or run from within the workspace.")
	return ""
}
