package auth

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
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

	var config GlabConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", err
	}

	if host, exists := config.Hosts["gitlab.com"]; exists {
		return host.Token, nil
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
