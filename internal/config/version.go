package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)
  type GitHubContent struct {
        Name        string `json:"name"`
        Path        string `json:"path"`
        Sha         string `json:"sha"`
        Size        int    `json:"size"`
        URL         string `json:"url"`
        HTMLURL     string `json:"html_url"`
        GitURL      string `json:"git_url"`
        DownloadURL string `json:"download_url"`
        Type        string `json:"type"`
        Content     string `json:"content"`
        Encoding    string `json:"encoding"`
  }
func GetLatestVersion() (string, error) {
	target := "https://api.github.com/repos/takashiraki/homebrew-tap/contents/myenv.rb"
	resp, err := http.Get(target)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var content GitHubContent

	if err := json.Unmarshal(body, &content); err != nil {
		return "", err
	}

	decoded, err := base64.StdEncoding.DecodeString(content.Content)

	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`version\s+"([^"]+)"`)

	matches := re.FindStringSubmatch(string(decoded))

	if len(matches) < 2 {
		return "", errors.New("version not found")
	}

	version := matches[1]

	return "v" +version, nil
}

func CheckForUpdates(currentVersion string) {
	latestVersion, err := GetLatestVersion()
	if err != nil {
		return
	}

	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(latestVersion, "v")

	if compareVersions(current, latest) < 0 {
		println("\n\033[33m⚠ Update Available:\033[0m")
		println("   Current version: \033[36m" + current + "\033[0m")
		println("   Latest version:  \033[32m" + latest + "\033[0m")
		println("\n\033[33m💡 Upgrade:\033[0m")
		println("   Run: \033[36mbrew update && brew upgrade myenv\033[0m\n")
	}
}

// compareVersions compares two semantic version strings.
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Compare each part (major, minor, patch)
	for i := 0; i < 3; i++ {
		var num1, num2 int

		if i < len(parts1) {
			num1, _ = strconv.Atoi(parts1[i])
		}

		if i < len(parts2) {
			num2, _ = strconv.Atoi(parts2[i])
		}

		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
	}

	return 0
}