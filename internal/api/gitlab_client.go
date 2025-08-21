package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rkristelijn/glab-tui/internal/auth"
)

// GitLabClient handles GitLab API requests
type GitLabClient struct {
	auth       *auth.GitLabAuth
	httpClient *http.Client
	baseURL    string
}

// Pipeline represents a GitLab pipeline
type Pipeline struct {
	ID        int       `json:"id"`
	IID       int       `json:"iid"`
	ProjectID int       `json:"project_id"`
	Status    string    `json:"status"`
	Ref       string    `json:"ref"`
	SHA       string    `json:"sha"`
	WebURL    string    `json:"web_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Job represents a GitLab job
type Job struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	Stage      string     `json:"stage"`
	CreatedAt  time.Time  `json:"created_at"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	Duration   *float64   `json:"duration"`
	WebURL     string     `json:"web_url"`
	Pipeline   struct {
		ID int `json:"id"`
	} `json:"pipeline"`
}

// NewGitLabClient creates a new GitLab API client
func NewGitLabClient() (*GitLabClient, error) {
	auth, err := auth.NewGitLabAuth()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize auth: %w", err)
	}

	return &GitLabClient{
		auth: auth,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: auth.GetBaseURL(),
	}, nil
}

// GetPipelines gets pipelines for a project
func (c *GitLabClient) GetPipelines(projectPath string, limit int) ([]Pipeline, error) {
	// Convert project path to project ID or use path directly
	projectID := url.PathEscape(projectPath)

	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/pipelines", c.baseURL, projectID)
	if limit > 0 {
		apiURL += fmt.Sprintf("?per_page=%d", limit)
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.auth.GetAuthHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var pipelines []Pipeline
	if err := json.NewDecoder(resp.Body).Decode(&pipelines); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return pipelines, nil
}

// GetJobs gets jobs for a pipeline
func (c *GitLabClient) GetJobs(projectPath string, pipelineID int) ([]Job, error) {
	projectID := url.PathEscape(projectPath)

	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/pipelines/%d/jobs", c.baseURL, projectID, pipelineID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.auth.GetAuthHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var jobs []Job
	if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return jobs, nil
}

// GetJob gets a specific job
func (c *GitLabClient) GetJob(projectPath string, jobID int) (*Job, error) {
	projectID := url.PathEscape(projectPath)

	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/jobs/%d", c.baseURL, projectID, jobID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.auth.GetAuthHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var job Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &job, nil
}

// GetJobLogs gets logs for a specific job
func (c *GitLabClient) GetJobLogs(projectPath string, jobID int) (string, error) {
	projectID := url.PathEscape(projectPath)

	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/jobs/%d/trace", c.baseURL, projectID, jobID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.auth.GetAuthHeader())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// TestConnection tests the GitLab API connection
func (c *GitLabClient) TestConnection() error {
	apiURL := fmt.Sprintf("%s/api/v4/user", c.baseURL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.auth.GetAuthHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
