package core

// Job represents a GitLab CI/CD job
type Job struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Stage    string `json:"stage"`
	Duration string `json:"duration,omitempty"`
}

// Pipeline represents a GitLab CI/CD pipeline
type Pipeline struct {
	ID          int    `json:"id"`
	Status      string `json:"status"`
	Ref         string `json:"ref"`
	WebURL      string `json:"web_url"`
	ProjectID   int    `json:"project_id"`
	ProjectName string `json:"-"` // Computed field
	Jobs        string `json:"-"` // Computed field
	Duration    string `json:"-"` // Computed field for display
}

// Project represents a GitLab project
type Project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	NameWithNamespace string `json:"name_with_namespace"`
	WebURL            string `json:"web_url"`
	LastActivityAt    string `json:"last_activity_at"`
	Archived          bool   `json:"archived"`
}

// MergeRequest represents a GitLab merge request
type MergeRequest struct {
	ID           int    `json:"iid"`
	Title        string `json:"title"`
	State        string `json:"state"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Author       struct {
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"author"`
	WebURL      string `json:"web_url"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	MergedAt    string `json:"merged_at,omitempty"`
	ClosedAt    string `json:"closed_at,omitempty"`
	ProjectID   int    `json:"project_id"`
	ProjectName string `json:"-"` // Computed field
	HasPipeline bool   `json:"-"` // Computed field
	PipelineID  int    `json:"-"` // Computed field
}

// PipelineService handles pipeline operations
type PipelineService struct {
	// Will contain GitLab client, config, etc.
}

// GetMockPipelines returns mock data for development (multi-project)
func GetMockPipelines() []Pipeline {
	return []Pipeline{
		{ID: 1996879423, Status: "running", Ref: "feat/zap-c3", ProjectID: 123, ProjectName: "frontend-app", Jobs: "3/8 jobs", Duration: "4m 32s"},
		{ID: 1996867272, Status: "running", Ref: "refs/merge-req/406", ProjectID: 456, ProjectName: "backend-api", Jobs: "5/8 jobs", Duration: "2m 15s"},
		{ID: 1996733511, Status: "success", Ref: "fix/supplier-bug", ProjectID: 123, ProjectName: "frontend-app", Jobs: "8/8 jobs", Duration: "6m 45s"},
		{ID: 1996723026, Status: "failed", Ref: "fix/supplier-bug", ProjectID: 789, ProjectName: "data-pipeline", Jobs: "failed", Duration: "3m 12s"},
		{ID: 1996719037, Status: "success", Ref: "main", ProjectID: 456, ProjectName: "backend-api", Jobs: "8/8 jobs", Duration: "5m 23s"},
		{ID: 1996719038, Status: "running", Ref: "feature/auth", ProjectID: 101, ProjectName: "auth-service", Jobs: "2/5 jobs", Duration: "1m 45s"},
		{ID: 1996719039, Status: "success", Ref: "main", ProjectID: 789, ProjectName: "data-pipeline", Jobs: "12/12 jobs", Duration: "8m 15s"},
	}
}

// GetMockProjects returns mock project data
func GetMockProjects() []Project {
	return []Project{
		{ID: 123, Name: "frontend-app", NameWithNamespace: "company/frontend-app", Archived: false},
		{ID: 456, Name: "backend-api", NameWithNamespace: "company/backend-api", Archived: false},
		{ID: 789, Name: "data-pipeline", NameWithNamespace: "company/data-pipeline", Archived: false},
		{ID: 101, Name: "auth-service", NameWithNamespace: "company/auth-service", Archived: false},
	}
}

// GetMockJobs returns mock job data for demo purposes
func GetMockJobs() []Job {
	return []Job{
		{ID: 11100001, Name: "build-frontend", Status: "success", Stage: "build"},
		{ID: 11100002, Name: "test-unit", Status: "success", Stage: "test"},
		{ID: 11100003, Name: "test-integration", Status: "running", Stage: "test"},
		{ID: 11100004, Name: "zap-security-scan", Status: "running", Stage: "security"},
		{ID: 11100005, Name: "build-docker", Status: "pending", Stage: "build"},
		{ID: 11100006, Name: "deploy-staging", Status: "pending", Stage: "deploy"},
		{ID: 11100007, Name: "cypress-e2e", Status: "pending", Stage: "test"},
		{ID: 11100008, Name: "deploy-production", Status: "pending", Stage: "deploy"},
	}
}

// GetMockMergeRequests returns mock merge request data for development
func GetMockMergeRequests() []MergeRequest {
	return []MergeRequest{
		{
			ID:           42,
			Title:        "feat: Add real-time log streaming support",
			State:        "opened",
			SourceBranch: "feat/log-streaming",
			TargetBranch: "main",
			Author: struct {
				Name     string `json:"name"`
				Username string `json:"username"`
			}{Name: "John Doe", Username: "johndoe"},
			WebURL:       "https://gitlab.com/company/frontend-app/-/merge_requests/42",
			CreatedAt:    "2024-08-20T10:30:00Z",
			UpdatedAt:    "2024-08-22T07:15:00Z",
			ProjectID:    123,
			ProjectName:  "frontend-app",
			HasPipeline:  true,
			PipelineID:   1996879423,
		},
		{
			ID:           41,
			Title:        "fix: Resolve child pipeline navigation bug",
			State:        "opened",
			SourceBranch: "fix/child-pipeline-nav",
			TargetBranch: "main",
			Author: struct {
				Name     string `json:"name"`
				Username string `json:"username"`
			}{Name: "Jane Smith", Username: "janesmith"},
			WebURL:       "https://gitlab.com/company/frontend-app/-/merge_requests/41",
			CreatedAt:    "2024-08-21T14:20:00Z",
			UpdatedAt:    "2024-08-22T06:45:00Z",
			ProjectID:    123,
			ProjectName:  "frontend-app",
			HasPipeline:  true,
			PipelineID:   1996867272,
		},
		{
			ID:           40,
			Title:        "docs: Update README with new features",
			State:        "merged",
			SourceBranch: "docs/update-readme",
			TargetBranch: "main",
			Author: struct {
				Name     string `json:"name"`
				Username string `json:"username"`
			}{Name: "Bob Wilson", Username: "bobwilson"},
			WebURL:       "https://gitlab.com/company/frontend-app/-/merge_requests/40",
			CreatedAt:    "2024-08-19T09:15:00Z",
			UpdatedAt:    "2024-08-20T16:30:00Z",
			MergedAt:     "2024-08-20T16:30:00Z",
			ProjectID:    123,
			ProjectName:  "frontend-app",
			HasPipeline:  false,
		},
		{
			ID:           39,
			Title:        "refactor: Improve TUI navigation performance",
			State:        "closed",
			SourceBranch: "refactor/tui-performance",
			TargetBranch: "main",
			Author: struct {
				Name     string `json:"name"`
				Username string `json:"username"`
			}{Name: "Alice Brown", Username: "alicebrown"},
			WebURL:       "https://gitlab.com/company/frontend-app/-/merge_requests/39",
			CreatedAt:    "2024-08-18T11:00:00Z",
			UpdatedAt:    "2024-08-19T08:45:00Z",
			ClosedAt:     "2024-08-19T08:45:00Z",
			ProjectID:    123,
			ProjectName:  "frontend-app",
			HasPipeline:  false,
		},
	}
}

// Future methods:
// func (ps *PipelineService) GetGroupProjects(groupID int) ([]Project, error)
// func (ps *PipelineService) GetProjectPipelines(projectID int) ([]Pipeline, error)
// func (ps *PipelineService) GetGroupPipelines(groupID int) ([]Pipeline, error)
// func (ps *PipelineService) RetryPipeline(pipelineID int) error
// func (ps *PipelineService) CancelPipeline(pipelineID int) error
