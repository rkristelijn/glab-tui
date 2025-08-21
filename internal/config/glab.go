package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadGlabToken reads the GitLab token from glab's config file
// Uses simple text parsing instead of YAML due to the !!null format
func LoadGlabToken() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(home, ".config", "glab-cli", "config.yml")
	file, err := os.Open(configPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inGitlabHost := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Look for gitlab.com host section
		if strings.Contains(line, "gitlab.com:") {
			inGitlabHost = true
			continue
		}

		// If we're in the gitlab.com section and find a token line
		if inGitlabHost && strings.Contains(line, "token:") {
			// Extract token after "token: !!null "
			parts := strings.Split(line, "token:")
			if len(parts) > 1 {
				tokenPart := strings.TrimSpace(parts[1])
				// Remove the !!null prefix if present
				tokenPart = strings.TrimPrefix(tokenPart, "!!null ")
				return tokenPart, nil
			}
		}

		// If we hit another host section, we're done with gitlab.com
		if inGitlabHost && strings.HasSuffix(line, ":") && !strings.Contains(line, "token:") {
			break
		}
	}

	return "", scanner.Err()
}
