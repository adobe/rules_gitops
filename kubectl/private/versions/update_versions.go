// update_versions downloads kubectl release information from dl.k8s.io
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
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const (
	dlBaseURL  = "https://dl.k8s.io/release/"
	outputFile = "kubectl/private/versions/versions.bzl"
)

// Platform combinations to fetch
var platforms = []struct {
	os   string
	arch string
}{
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"linux", "amd64"},
	{"linux", "arm64"},
}

// VersionInfo holds platform -> sha256 mappings for a version
type VersionInfo struct {
	Platforms map[string]string // platform (e.g., "darwin_arm64") -> sha256
}

func main() {
	log.SetFlags(0)

	// Discover stable versions
	versions := discoverStableVersions()
	if len(versions) == 0 {
		log.Fatalf("No stable versions found")
	}

	log.Printf("Found %d stable versions: %v", len(versions), versions)

	// Fetch SHA256 checksums for each version and platform
	versionInfos := fetchAllChecksums(versions)

	// Find latest version
	latestVersion := findLatestVersion(versions)

	// Generate versions.bzl content
	content := generateBzl(versionInfos, latestVersion)

	// Determine output path
	outputPath := getOutputPath()

	// Write the file
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	log.Printf("Successfully wrote %s with %d versions", outputPath, len(versionInfos))
}

// discoverStableVersions probes stable-X.Y.txt endpoints to find available versions
func discoverStableVersions() []string {
	var versions []string

	// First, get the absolute latest stable to determine the max minor version
	latestStable := fetchStableVersion("stable.txt")
	if latestStable == "" {
		log.Printf("Warning: could not fetch stable.txt")
		return versions
	}
	versions = append(versions, latestStable)

	// Parse the minor version from latest stable (e.g., "1.35.0" -> 35)
	maxMinor := parseMinorVersion(latestStable)
	if maxMinor == 0 {
		log.Printf("Warning: could not parse minor version from %s", latestStable)
		return versions
	}

	log.Printf("Latest stable is %s (minor version %d)", latestStable, maxMinor)

	// Probe stable-1.X.txt for minor versions from 20 up to the latest
	// Kubernetes typically maintains ~3-4 minor versions, but we fetch more for completeness
	for minor := 20; minor <= maxMinor; minor++ {
		filename := fmt.Sprintf("stable-1.%d.txt", minor)
		if v := fetchStableVersion(filename); v != "" {
			// Avoid duplicates
			found := false
			for _, existing := range versions {
				if existing == v {
					found = true
					break
				}
			}
			if !found {
				versions = append(versions, v)
			}
		}
	}

	return versions
}

// parseMinorVersion extracts the minor version number from a version string like "1.35.0"
func parseMinorVersion(version string) int {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0
	}
	var minor int
	fmt.Sscanf(parts[1], "%d", &minor)
	return minor
}

// fetchStableVersion fetches a stable version string from dl.k8s.io
func fetchStableVersion(filename string) string {
	url := dlBaseURL + filename

	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	version := strings.TrimSpace(string(body))
	// Validate it looks like a version (starts with 'v')
	if !strings.HasPrefix(version, "v") {
		return ""
	}

	// Strip the 'v' prefix for consistency
	return strings.TrimPrefix(version, "v")
}

// fetchAllChecksums fetches SHA256 checksums for all versions and platforms
func fetchAllChecksums(versions []string) map[string]*VersionInfo {
	result := make(map[string]*VersionInfo)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Use a semaphore to limit concurrent requests
	sem := make(chan struct{}, 10)

	for _, version := range versions {
		for _, p := range platforms {
			wg.Add(1)
			go func(version, os, arch string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				sha256, err := fetchSHA256(version, os, arch)
				if err != nil {
					log.Printf("Warning: failed to fetch sha256 for %s %s/%s: %v", version, os, arch, err)
					return
				}

				platform := fmt.Sprintf("%s_%s", os, arch)

				mu.Lock()
				if result[version] == nil {
					result[version] = &VersionInfo{
						Platforms: make(map[string]string),
					}
				}
				result[version].Platforms[platform] = sha256
				mu.Unlock()
			}(version, p.os, p.arch)
		}
	}

	wg.Wait()
	return result
}

// fetchSHA256 fetches the SHA256 checksum for a specific version/os/arch
func fetchSHA256(version, os, arch string) (string, error) {
	url := fmt.Sprintf("%sv%s/bin/%s/%s/kubectl.sha256", dlBaseURL, version, os, arch)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// The sha256 file may contain just the hash, or "hash  filename" format
	content := strings.TrimSpace(string(body))
	parts := strings.Fields(content)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty sha256 file")
	}

	sha256 := parts[0]

	// Validate it looks like a SHA256 (64 hex chars)
	if len(sha256) != 64 {
		return "", fmt.Errorf("invalid sha256 length: %d", len(sha256))
	}

	return sha256, nil
}

// findLatestVersion returns the latest version from a list of versions
func findLatestVersion(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	latest := versions[0]
	for _, v := range versions[1:] {
		if compareVersions(v, latest) > 0 {
			latest = v
		}
	}
	return latest
}

// compareVersions compares two semver-like version strings
// Returns positive if v1 > v2, negative if v1 < v2, zero if equal
func compareVersions(v1, v2 string) int {
	if v2 == "" {
		return 1
	}

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		var n1, n2 int
		fmt.Sscanf(parts1[i], "%d", &n1)
		fmt.Sscanf(parts2[i], "%d", &n2)
		if n1 != n2 {
			return n1 - n2
		}
	}
	return len(parts1) - len(parts2)
}

func generateBzl(versions map[string]*VersionInfo, latestVersion string) string {
	var sb strings.Builder

	// Sort versions for consistent output (reverse to put newer versions first)
	versionList := make([]string, 0, len(versions))
	for v := range versions {
		versionList = append(versionList, v)
	}
	sort.Slice(versionList, func(i, j int) bool {
		return compareVersions(versionList[i], versionList[j]) > 0
	})

	sb.WriteString(`"""Generated by update_versions.go - do not edit manually."""

`)
	sb.WriteString(fmt.Sprintf("LATEST_KUBECTL_VERSION = %q\n\n", latestVersion))
	sb.WriteString("VERSIONS = {\n")

	for _, version := range versionList {
		info := versions[version]
		sb.WriteString(fmt.Sprintf("    %q: {\n", version))

		// Sort platforms for consistent output
		platformKeys := make([]string, 0, len(info.Platforms))
		for p := range info.Platforms {
			platformKeys = append(platformKeys, p)
		}
		sort.Strings(platformKeys)

		for _, platform := range platformKeys {
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
