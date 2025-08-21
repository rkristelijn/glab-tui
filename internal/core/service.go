package core

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
	"github.com/rkristelijn/glab-tui/internal/config"
)

type Service struct {
	config *config.Config
	gitlab GitLabClient
}

type GitLabClient interface {
	GetJob(projectID, jobID int) (interface{}, error)
	GetProjectPipelines(projectID int) ([]Pipeline, error)
	GetProject(projectID int) (interface{}, error)
	GetProjectByPath(path string) (interface{}, error)
}

func NewService(cfg *config.Config, gitlabClient GitLabClient) *Service {
	return &Service{
		config: cfg,
		gitlab: gitlabClient,
	}
}

// GetPipelines returns pipelines based on configuration
func (s *Service) GetPipelines() ([]Pipeline, error) {
	// For now, just use mock data - we'll improve this integration later
	return GetMockPipelines(), nil
}

// GetJobStatus gets the status of a specific job
func (s *Service) GetJobStatus(jobID int) (string, error) {
	if s.config.GitLab.ProjectID == 0 {
		return "", fmt.Errorf("no project ID configured")
	}

	job, err := s.gitlab.GetJob(s.config.GitLab.ProjectID, jobID)
	if err != nil {
		return "", fmt.Errorf("failed to get job %d: %w", jobID, err)
	}

	return extractJobStatus(job), nil
}

// Helper functions to extract data from interface{}
func extractProjectID(project interface{}) int {
	if p, ok := project.(*gitlab.Project); ok {
		return p.ID
	}
	return 0
}

func extractJobStatus(job interface{}) string {
	if j, ok := job.(*gitlab.Job); ok {
		return j.Status
	}
	return "unknown"
}
