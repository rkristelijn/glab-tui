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

// PipelineService handles pipeline operations
type PipelineService struct {
	// Will contain GitLab client, config, etc.
}

// GetMockPipelines returns mock data for development (multi-project)
func GetMockPipelines() []Pipeline {
	return []Pipeline{
		{ID: 1996879423, Status: "running", Ref: "feat/zap-c3", ProjectID: 123, ProjectName: "frontend-app", Jobs: "3/8 jobs"},
		{ID: 1996867272, Status: "running", Ref: "refs/merge-req/406", ProjectID: 456, ProjectName: "backend-api", Jobs: "5/8 jobs"},
		{ID: 1996733511, Status: "success", Ref: "fix/supplier-bug", ProjectID: 123, ProjectName: "frontend-app", Jobs: "8/8 jobs"},
		{ID: 1996723026, Status: "failed", Ref: "fix/supplier-bug", ProjectID: 789, ProjectName: "data-pipeline", Jobs: "failed"},
		{ID: 1996719037, Status: "success", Ref: "main", ProjectID: 456, ProjectName: "backend-api", Jobs: "8/8 jobs"},
		{ID: 1996719038, Status: "running", Ref: "feature/auth", ProjectID: 101, ProjectName: "auth-service", Jobs: "2/5 jobs"},
		{ID: 1996719039, Status: "success", Ref: "main", ProjectID: 789, ProjectName: "data-pipeline", Jobs: "12/12 jobs"},
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

// Future methods:
// func (ps *PipelineService) GetGroupProjects(groupID int) ([]Project, error)
// func (ps *PipelineService) GetProjectPipelines(projectID int) ([]Pipeline, error)
// func (ps *PipelineService) GetGroupPipelines(groupID int) ([]Pipeline, error)
// func (ps *PipelineService) RetryPipeline(pipelineID int) error
// func (ps *PipelineService) CancelPipeline(pipelineID int) error
