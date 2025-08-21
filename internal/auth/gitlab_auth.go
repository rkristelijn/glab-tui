package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GitLabAuth handles GitLab authentication
type GitLabAuth struct {
	token   string
	baseURL string
}

// GlabConfig represents the glab CLI config structure
type GlabConfig struct {
	Hosts map[string]struct {
		Token string `yaml:"token"`
	} `yaml:"hosts"`
}

// NewGitLabAuth creates a new GitLab authentication handler
func NewGitLabAuth() (*GitLabAuth, error) {
	auth := &GitLabAuth{
		baseURL: "https://gitlab.com",
	}

	// Try to get token from glab config file
	if token, err := auth.getTokenFromGlabConfig(); err == nil && token != "" {
		auth.token = token
		return auth, nil
	}

	// Try environment variables
	if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		auth.token = token
		return auth, nil
	}

	if token := os.Getenv("GLAB_TOKEN"); token != "" {
		auth.token = token
		return auth, nil
	}

	return nil, fmt.Errorf("no GitLab token found - run 'glab auth login' first")
}

// getTokenFromGlabConfig reads the token from glab CLI config file
func (g *GitLabAuth) getTokenFromGlabConfig() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(homeDir, ".config", "glab-cli", "config.yml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	// Parse manually to handle !!null tags
	content := string(data)
	lines := strings.Split(content, "\n")

	inGitLabHost := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check if we're in the gitlab.com section
		if strings.Contains(line, "gitlab.com:") {
			inGitLabHost = true
			continue
		}

		// If we're in gitlab.com section and find token line
		if inGitLabHost && strings.HasPrefix(line, "token:") {
			// Extract token value, handling !!null prefix
			tokenPart := strings.TrimPrefix(line, "token:")
			tokenPart = strings.TrimSpace(tokenPart)

			if strings.HasPrefix(tokenPart, "!!null ") {
				token := strings.TrimPrefix(tokenPart, "!!null ")
				token = strings.TrimSpace(token)
				if token != "" {
					return token, nil
				}
			}
		}

		// If we hit another host section, we're done with gitlab.com
		if inGitLabHost && strings.HasSuffix(line, ":") && !strings.HasPrefix(line, " ") {
			break
		}
	}

	return "", fmt.Errorf("no token found for gitlab.com in glab config")
}

// GetToken returns the GitLab token
func (g *GitLabAuth) GetToken() string {
	return g.token
}

// GetBaseURL returns the GitLab base URL
func (g *GitLabAuth) GetBaseURL() string {
	return g.baseURL
}

// GetAuthHeader returns the authorization header value
func (g *GitLabAuth) GetAuthHeader() string {
	return fmt.Sprintf("Bearer %s", g.token)
}

// IsAuthenticated checks if we have a valid token
func (g *GitLabAuth) IsAuthenticated() bool {
	return g.token != ""
}
