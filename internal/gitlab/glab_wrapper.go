package gitlab

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rkristelijn/glab-tui/internal/core"
)

// GlabWrapper uses the glab CLI to interact with GitLab
type GlabWrapper struct {
	projectPath string
}

func NewGlabWrapper(projectPath string) *GlabWrapper {
	return &GlabWrapper{
		projectPath: projectPath,
	}
}

// GetProjectPipelines fetches pipelines using glab CLI
func (g *GlabWrapper) GetProjectPipelines(projectID int) ([]core.Pipeline, error) {
	cmd := exec.Command("glab", "pipeline", "list", "-R", g.projectPath, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run glab command: %w", err)
	}

	var glabPipelines []struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
		Ref    string `json:"ref"`
		WebURL string `json:"web_url"`
	}

	if err := json.Unmarshal(output, &glabPipelines); err != nil {
		return nil, fmt.Errorf("failed to parse glab output: %w", err)
	}

	var pipelines []core.Pipeline
	for _, p := range glabPipelines {
		pipelines = append(pipelines, core.Pipeline{
			ID:          p.ID,
			Status:      p.Status,
			Ref:         p.Ref,
			WebURL:      p.WebURL,
			ProjectID:   projectID,
			ProjectName: "frontend-apps",
			Jobs:        "loading...",
		})
	}

	return pipelines, nil
}

// GetJob fetches job info using glab CLI
func (g *GlabWrapper) GetJob(projectID, jobID int) (interface{}, error) {
	// Use glab to get job info - this is tricky since glab doesn't have direct job info
	// For now, let's return a mock job with the ID
	return &struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	}{
		ID:     jobID,
		Status: "unknown", // We'd need to parse pipeline info to get job status
	}, nil
}

// GetProject returns project info
func (g *GlabWrapper) GetProject(projectID int) (interface{}, error) {
	cmd := exec.Command("glab", "project", "view", g.projectPath, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get project info: %w", err)
	}

	var project struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal(output, &project); err != nil {
		return nil, fmt.Errorf("failed to parse project info: %w", err)
	}

	return &project, nil
}

// GetProjectByPath returns project info by path
func (g *GlabWrapper) GetProjectByPath(path string) (interface{}, error) {
	return g.GetProject(0) // Use the configured project
}

// GetPipelineJobs fetches jobs for a specific pipeline using glab CLI
func (g *GlabWrapper) GetPipelineJobs(pipelineID int) ([]core.Job, error) {
	// Use glab to get pipeline details with jobs
	cmd := exec.Command("glab", "ci", "view", strconv.Itoa(pipelineID), "-R", g.projectPath)
	_, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline jobs: %w", err)
	}

	// For now, return mock jobs - we'll parse the real output later
	return []core.Job{
		{ID: 1001, Name: "npm-preparation", Status: "success", Stage: "prepare"},
		{ID: 1002, Name: "nx-mono-repo-affected", Status: "running", Stage: "build"},
		{ID: 1003, Name: "cloudflare-deploy", Status: "pending", Stage: "deploy"},
		{ID: 1004, Name: "zap-security-scan", Status: "pending", Stage: "test"},
		{ID: 1005, Name: "cypress-e2e", Status: "pending", Stage: "test"},
	}, nil
}

// GetJobLogs fetches logs for a specific job using glab CLI
func (g *GlabWrapper) GetJobLogs(jobID int) (string, error) {
	// Use glab to get job logs
	cmd := exec.Command("glab", "ci", "trace", strconv.Itoa(jobID), "-R", g.projectPath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get job logs: %w", err)
	}

	return string(output), nil
}
func ParseGlabPipelineList(output string) ([]core.Pipeline, error) {
	lines := strings.Split(output, "\n")
	var pipelines []core.Pipeline

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Showing") || strings.HasPrefix(line, "State") {
			continue
		}

		// Parse line like: "(running) • #1996941196	(#6785)	feat/zap-c3	(about 4 minutes ago)"
		if strings.Contains(line, "•") && strings.Contains(line, "#") {
			parts := strings.Split(line, "\t")
			if len(parts) >= 3 {
				// Extract status
				status := "unknown"
				if strings.Contains(parts[0], "(") && strings.Contains(parts[0], ")") {
					start := strings.Index(parts[0], "(") + 1
					end := strings.Index(parts[0], ")")
					if end > start {
						status = parts[0][start:end]
					}
				}

				// Extract pipeline ID
				pipelineID := 0
				if strings.Contains(parts[0], "#") {
					idStr := strings.Split(parts[0], "#")[1]
					idStr = strings.TrimSpace(idStr)
					if id, err := strconv.Atoi(idStr); err == nil {
						pipelineID = id
					}
				}

				// Extract ref
				ref := strings.TrimSpace(parts[2])

				if pipelineID > 0 {
					pipelines = append(pipelines, core.Pipeline{
						ID:          pipelineID,
						Status:      status,
						Ref:         ref,
						ProjectName: "frontend-apps",
						Jobs:        "loading...",
					})
				}
			}
		}
	}

	return pipelines, nil
}
