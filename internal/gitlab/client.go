package gitlab

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
	"github.com/rkristelijn/glab-tui/internal/config"
	"github.com/rkristelijn/glab-tui/internal/core"
)

type Client struct {
	client *gitlab.Client
	config *config.Config
}

func NewClient(cfg *config.Config) (*Client, error) {
	client, err := gitlab.NewClient(cfg.GitLab.Token, gitlab.WithBaseURL(cfg.GitLab.URL))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// GetJob fetches a specific job by ID
func (c *Client) GetJob(projectID, jobID int) (interface{}, error) {
	job, _, err := c.client.Jobs.GetJob(projectID, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job %d from project %d: %w", jobID, projectID, err)
	}
	return job, nil
}

// GetProjectPipelines fetches recent pipelines for a project
func (c *Client) GetProjectPipelines(projectID int) ([]core.Pipeline, error) {
	opts := &gitlab.ListProjectPipelinesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: c.config.UI.MaxPipelinesPerProject,
			Page:    1,
		},
		OrderBy: gitlab.String("updated_at"),
		Sort:    gitlab.String("desc"),
	}

	pipelines, _, err := c.client.Pipelines.ListProjectPipelines(projectID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipelines for project %d: %w", projectID, err)
	}

	var result []core.Pipeline
	for _, p := range pipelines {
		result = append(result, core.Pipeline{
			ID:        p.ID,
			Status:    p.Status,
			Ref:       p.Ref,
			WebURL:    p.WebURL,
			ProjectID: projectID,
			Jobs:      "loading...", // We'd need another API call to get job count
		})
	}

	return result, nil
}

// GetProject fetches project details
func (c *Client) GetProject(projectID int) (interface{}, error) {
	project, _, err := c.client.Projects.GetProject(projectID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %d: %w", projectID, err)
	}
	return project, nil
}

// GetProjectByPath fetches project by path (e.g., "group/project/frontend-apps")
func (c *Client) GetProjectByPath(path string) (interface{}, error) {
	project, _, err := c.client.Projects.GetProject(path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", path, err)
	}
	return project, nil
}
