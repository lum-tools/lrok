package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	githubAPI      = "https://api.github.com/repos/lum-tools/lrok/releases/latest"
	checkInterval  = 24 * time.Hour
	cacheFileName  = ".lrok_version_check"
)

// VersionCache stores the last version check
type VersionCache struct {
	LastCheck     time.Time `json:"last_check"`
	LatestVersion string    `json:"latest_version"`
}

// InstallMethod represents how lrok was installed
type InstallMethod string

const (
	InstallScript InstallMethod = "script"
	InstallNPM    InstallMethod = "npm"
	InstallPIP    InstallMethod = "pip"
	InstallManual InstallMethod = "manual"
)

// CheckForUpdate checks if a newer version is available
// Returns (hasUpdate, latestVersion, installMethod, error)
func CheckForUpdate(currentVersion string) (bool, string, InstallMethod, error) {
	// Skip check for dev builds
	if currentVersion == "dev" || currentVersion == "" {
		return false, "", InstallManual, nil
	}

	// Check cache first
	cacheFile := filepath.Join(os.TempDir(), cacheFileName)
	if cache, err := readCache(cacheFile); err == nil {
		// Use cached result if less than 24 hours old
		if time.Since(cache.LastCheck) < checkInterval {
			hasUpdate := compareVersions(currentVersion, cache.LatestVersion)
			method := detectInstallMethod()
			return hasUpdate, cache.LatestVersion, method, nil
		}
	}

	// Fetch latest version from GitHub
	latestVersion, err := fetchLatestVersion()
	if err != nil {
		// If fetch fails, don't show error to user (silent fail)
		return false, "", InstallManual, nil
	}

	// Save to cache
	cache := VersionCache{
		LastCheck:     time.Now(),
		LatestVersion: latestVersion,
	}
	_ = saveCache(cacheFile, cache) // Ignore error

	hasUpdate := compareVersions(currentVersion, latestVersion)
	method := detectInstallMethod()

	return hasUpdate, latestVersion, method, nil
}

// fetchLatestVersion fetches the latest version from GitHub API
func fetchLatestVersion() (string, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	
	resp, err := client.Get(githubAPI)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(body, &release); err != nil {
		return "", err
	}

	// Remove 'v' prefix if present
	version := strings.TrimPrefix(release.TagName, "v")
	return version, nil
}

// compareVersions returns true if latest > current
func compareVersions(current, latest string) bool {
	// Simple string comparison for now (works for semver)
	// v0.0.5 < v0.0.6, etc.
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	
	return latest > current
}

// detectInstallMethod detects how lrok was installed
func detectInstallMethod() InstallMethod {
	// Check if installed via npm
	if _, err := exec.LookPath("npm"); err == nil {
		cmd := exec.Command("npm", "list", "-g", "lrok")
		if err := cmd.Run(); err == nil {
			return InstallNPM
		}
	}

	// Check if installed via pip
	if _, err := exec.LookPath("pip"); err == nil {
		cmd := exec.Command("pip", "show", "lrok")
		if err := cmd.Run(); err == nil {
			return InstallPIP
		}
	}
	if _, err := exec.LookPath("pip3"); err == nil {
		cmd := exec.Command("pip3", "show", "lrok")
		if err := cmd.Run(); err == nil {
			return InstallPIP
		}
	}

	// Check if in ~/.local/bin (install script)
	exePath, err := os.Executable()
	if err == nil && strings.Contains(exePath, ".local/bin") {
		return InstallScript
	}

	// Default to manual
	return InstallManual
}

// GetUpdateCommand returns the appropriate update command for the install method
func GetUpdateCommand(method InstallMethod) string {
	switch method {
	case InstallScript:
		return "curl -fsSL https://platform.lum.tools/install.sh | bash"
	case InstallNPM:
		return "npm update -g lrok"
	case InstallPIP:
		return "pip install --upgrade lrok"
	default:
		return "curl -fsSL https://platform.lum.tools/install.sh | bash"
	}
}

// readCache reads the version cache file
func readCache(path string) (*VersionCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cache VersionCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

// saveCache saves the version cache file
func saveCache(path string, cache VersionCache) error {
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ShowUpdateWarning displays a formatted update warning
func ShowUpdateWarning(currentVersion, latestVersion string, method InstallMethod) {
	updateCmd := GetUpdateCommand(method)
	
	fmt.Println()
	fmt.Println("╭─────────────────────────────────────────────────────────────╮")
	fmt.Printf("│ ⚠️  Update available: %s → %s%-20s│\n", currentVersion, latestVersion, "")
	fmt.Println("│                                                             │")
	fmt.Printf("│ Run: %-54s │\n", updateCmd)
	fmt.Println("╰─────────────────────────────────────────────────────────────╯")
	fmt.Println()
}

