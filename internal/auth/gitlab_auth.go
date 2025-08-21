package auth

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GitLabAuth handles GitLab authentication
type GitLabAuth struct {
	token   string
	baseURL string
}

// NewGitLabAuth creates a new GitLab authentication handler
func NewGitLabAuth() (*GitLabAuth, error) {
	auth := &GitLabAuth{
		baseURL: "https://gitlab.com",
	}

	// Try to get token from glab CLI first
	if token, err := auth.getTokenFromGlab(); err == nil && token != "" {
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

// getTokenFromGlab gets the token from glab CLI
func (g *GitLabAuth) getTokenFromGlab() (string, error) {
	cmd := exec.Command("glab", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	token := strings.TrimSpace(string(output))
	if token == "" {
		return "", fmt.Errorf("empty token from glab")
	}

	return token, nil
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
