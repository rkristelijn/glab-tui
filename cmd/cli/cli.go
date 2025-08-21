package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rkristelijn/glab-tui/internal/config"
	"github.com/rkristelijn/glab-tui/internal/core"
	"github.com/rkristelijn/glab-tui/internal/gitlab"
)

func Run(args []string) {
	if len(args) == 0 {
		showHelp()
		return
	}

	command := args[0]

	switch command {
	case "pipelines", "p":
		listPipelines()
	case "job", "j":
		if len(args) < 2 {
			fmt.Println("Usage: glab-tui job <job-id>")
			os.Exit(1)
		}
		checkJob(args[1])
	case "logs", "l":
		if len(args) < 2 {
			fmt.Println("Usage: glab-tui logs <job-id>")
			os.Exit(1)
		}
		showJobLogs(args[1])
	case "test-real":
		testRealGitLab()
	case "help", "h", "--help":
		showHelp()
	case "version", "v", "--version":
		fmt.Println("glab-tui v0.1.0")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showHelp()
		os.Exit(1)
	}
}

func listPipelines() {
	// First try to get current project from git context
	projectPath, err := getCurrentProjectPath()
	if err != nil {
		fmt.Printf("Warning: Could not detect GitLab project: %v\n", err)
		// Fall back to mock data
		pipelines := core.GetMockPipelines()
		displayPipelines(pipelines, "Mock Data - No Project Context")
		return
	}

	// Try using glab command directly first (most reliable)
	if pipelines, err := getProjectPipelinesViaGlab(projectPath); err == nil {
		displayPipelines(pipelines, fmt.Sprintf("Real Data via glab - %s", projectPath))
		return
	} else {
		fmt.Printf("glab command failed: %v\n", err)
	}

	// Fallback to direct API calls
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		// Fall back to mock data
		pipelines := core.GetMockPipelines()
		displayPipelines(pipelines, "Mock Data - Config Error")
		return
	}

	if cfg.GitLab.Token == "" || cfg.GitLab.Token == "your-token-here" {
		fmt.Println("❌ No GitLab token found!")
		fmt.Println("Please set GITLAB_TOKEN in your .env file")
		// Fall back to mock data
		pipelines := core.GetMockPipelines()
		displayPipelines(pipelines, "Mock Data - No Token")
		return
	}

	client, err := gitlab.NewClient(cfg)
	if err != nil {
		fmt.Printf("Failed to create GitLab client: %v\n", err)
		// Fall back to mock data
		pipelines := core.GetMockPipelines()
		displayPipelines(pipelines, "Mock Data - Client Error")
		return
	}

	// Get project details first
	project, err := client.GetProjectByPath(projectPath)
	if err != nil {
		fmt.Printf("Failed to get project %s: %v\n", projectPath, err)
		// Fall back to mock data
		pipelines := core.GetMockPipelines()
		displayPipelines(pipelines, "Mock Data - Project Not Found")
		return
	}

	// Extract project ID (this is a bit hacky, we need to improve the client interface)
	projectID := extractProjectID(project)
	if projectID == 0 {
		fmt.Printf("Could not extract project ID from project data\n")
		// Fall back to mock data
		pipelines := core.GetMockPipelines()
		displayPipelines(pipelines, "Mock Data - No Project ID")
		return
	}

	// Get real pipeline data
	pipelines, err := client.GetProjectPipelines(projectID)
	if err != nil {
		fmt.Printf("Failed to get pipelines: %v\n", err)
		// Fall back to mock data
		mockPipelines := core.GetMockPipelines()
		displayPipelines(mockPipelines, "Mock Data - API Error")
		return
	}

	displayPipelines(pipelines, fmt.Sprintf("Real Data via API - %s", projectPath))
}

func getProjectPipelinesViaGlab(projectPath string) ([]core.Pipeline, error) {
	// Use glab command to get pipeline data
	cmd := exec.Command("glab", "pipeline", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("glab command failed: %w", err)
	}

	// Parse the text output
	pipelines, err := parseGlabPipelineText(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse glab output: %w", err)
	}

	return pipelines, nil
}

func parseGlabPipelineText(output string) ([]core.Pipeline, error) {
	lines := strings.Split(output, "\n")
	var pipelines []core.Pipeline

	// Extract project name from header line like "Showing 30 pipelines on group/project"
	projectName := "current-project"
	for _, line := range lines {
		if strings.HasPrefix(line, "Showing") && strings.Contains(line, " on ") {
			parts := strings.Split(line, " on ")
			if len(parts) > 1 {
				fullPath := strings.TrimSuffix(parts[1], ". (Page 1)")
				// Extract just the project name from the full path
				pathParts := strings.Split(fullPath, "/")
				if len(pathParts) > 0 {
					projectName = pathParts[len(pathParts)-1]
				}
			}
			break
		}
	}

	// Skip header lines and parse pipeline data
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Showing") || strings.HasPrefix(line, "State") {
			continue
		}

		// Parse lines like: "(running) • #1997196243	(#6866)	refs/merge-requests/406/head	(less than a minute ago)"
		if strings.Contains(line, "#") {
			pipeline := parsePipelineLine(line, projectName)
			if pipeline.ID != 0 {
				pipelines = append(pipelines, pipeline)
			}
		}
	}

	return pipelines, nil
}

func parsePipelineLine(line, projectName string) core.Pipeline {
	// Extract status
	var status string
	if strings.Contains(line, "(running)") {
		status = "running"
	} else if strings.Contains(line, "(success)") {
		status = "success"
	} else if strings.Contains(line, "(failed)") {
		status = "failed"
	} else if strings.Contains(line, "(waiting_for_resource)") {
		status = "waiting_for_resource"
	} else {
		status = "unknown"
	}

	// Extract pipeline ID
	var pipelineID int
	if idx := strings.Index(line, "#"); idx != -1 {
		idStr := ""
		for i := idx + 1; i < len(line) && (line[i] >= '0' && line[i] <= '9'); i++ {
			idStr += string(line[i])
		}
		if id, err := strconv.Atoi(idStr); err == nil {
			pipelineID = id
		}
	}

	// Extract ref (simplified) - it's the 3rd tab-separated field
	parts := strings.Split(line, "\t")
	ref := "unknown"
	if len(parts) >= 3 {
		ref = strings.TrimSpace(parts[2])
		// Clean up long merge request refs
		if strings.HasPrefix(ref, "refs/merge-requests/") {
			ref = "MR-" + strings.Split(ref, "/")[2]
		}
	}

	// Determine job status based on pipeline status
	jobs := "pending"
	switch status {
	case "running":
		jobs = "in progress"
	case "success":
		jobs = "completed"
	case "failed":
		jobs = "failed"
	case "waiting_for_resource":
		jobs = "queued"
	}

	return core.Pipeline{
		ID:          pipelineID,
		Status:      status,
		Ref:         ref,
		ProjectName: projectName,
		Jobs:        jobs,
	}
}

func displayPipelines(pipelines []core.Pipeline, source string) {
	fmt.Printf("GitLab Pipelines (%s):\n", source)
	fmt.Println("ID          Status    Project         Ref                   Jobs")
	fmt.Println("─────────────────────────────────────────────────────────────────────")

	for _, p := range pipelines {
		status := getStatusIcon(p.Status)
		fmt.Printf("%-10d  %s %-8s %-15s %-20s %s\n",
			p.ID, status, p.Status, p.ProjectName, p.Ref, p.Jobs)
	}
}

func getCurrentProjectPath() (string, error) {
	// Get the remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))

	// Parse GitLab project path from URL
	if strings.Contains(remoteURL, "gitlab") {
		// Extract project path from GitLab URL
		if strings.HasPrefix(remoteURL, "git@") {
			// SSH format: git@gitlab.com:group/project.git
			parts := strings.Split(remoteURL, ":")
			if len(parts) >= 2 {
				return strings.TrimSuffix(parts[1], ".git"), nil
			}
		} else if strings.HasPrefix(remoteURL, "https://") {
			// HTTPS format: https://gitlab.com/group/project.git
			parts := strings.Split(remoteURL, "/")
			if len(parts) >= 4 {
				projectPath := strings.Join(parts[3:], "/")
				return strings.TrimSuffix(projectPath, ".git"), nil
			}
		}
	}

	return "", fmt.Errorf("not a GitLab repository or unsupported URL format: %s", remoteURL)
}

func extractProjectID(project interface{}) int {
	// This is a temporary hack - we need to improve the client interface
	// For now, try to extract ID from the project data
	if p, ok := project.(map[string]interface{}); ok {
		if id, exists := p["id"]; exists {
			if idFloat, ok := id.(float64); ok {
				return int(idFloat)
			}
		}
	}
	return 0
}

func getStatusIcon(status string) string {
	switch status {
	case "running":
		return "●"
	case "success":
		return "✓"
	case "failed":
		return "✗"
	default:
		return "○"
	}
}

func checkJob(jobIDStr string) {
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		fmt.Printf("Invalid job ID: %s\n", jobIDStr)
		os.Exit(1)
	}

	fmt.Printf("Checking job %d...\n", jobID)

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if cfg.GitLab.Token == "" || cfg.GitLab.Token == "your-token-here" {
		fmt.Println("❌ No GitLab token found!")
		fmt.Println("Please set GITLAB_TOKEN in your .env file")
		fmt.Println("Get your token from: https://gitlab.com/-/profile/personal_access_tokens")
		os.Exit(1)
	}

	client, err := gitlab.NewClient(cfg)
	if err != nil {
		fmt.Printf("Failed to create GitLab client: %v\n", err)
		os.Exit(1)
	}

	service := core.NewService(cfg, client)
	status, err := service.GetJobStatus(jobID)
	if err != nil {
		fmt.Printf("❌ Failed to get job status: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Job %d status: %s\n", jobID, status)
}

func showJobLogs(jobIDStr string) {
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		fmt.Printf("Invalid job ID: %s\n", jobIDStr)
		os.Exit(1)
	}

	fmt.Printf("Fetching logs for job %d...\n", jobID)

	// Auto-detect current project
	projectPath, err := getCurrentProjectPath()
	if err != nil {
		fmt.Printf("❌ Could not detect GitLab project: %v\n", err)
		os.Exit(1)
	}

	wrapper := gitlab.NewGlabWrapper(projectPath)
	logs, err := wrapper.GetJobLogs(jobID)
	if err != nil {
		fmt.Printf("❌ Failed to get job logs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📋 Job %d logs:\n", jobID)
	fmt.Println("─────────────────────────────────────────────────")
	fmt.Println(logs)
}

func testRealGitLab() {
	fmt.Println("Testing real GitLab connection using glab...")

	// Test glab command directly
	cmd := exec.Command("glab", "pipeline", "list", "-R", "group/project/frontend-apps")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("❌ Failed to run glab command: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ glab command successful!")
	fmt.Println("📋 Raw pipeline data:")
	fmt.Println(string(output))

	// Parse the output
	pipelines, err := gitlab.ParseGlabPipelineList(string(output))
	if err != nil {
		fmt.Printf("❌ Failed to parse pipeline data: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Parsed %d pipelines!\n", len(pipelines))
	for i, p := range pipelines {
		if i < 5 { // Show first 5
			fmt.Printf("  Pipeline #%d: %s (%s)\n", p.ID, p.Status, p.Ref)
		}
	}
}

func showHelp() {
	fmt.Println(`glab-tui - GitLab TUI and CLI

USAGE:
    glab-tui                    Start interactive TUI (default)
    glab-tui speed              🔥 SPEED CHALLENGE MODE
    glab-tui [COMMAND]          Run CLI command

COMMANDS:
    pipelines, p               List pipelines
    job, j <job-id>           Check specific job status
    logs, l <job-id>          Show job logs
    test-real                 Test GitLab API connection
    speed                     🔥 Speed challenge mode
    help, h                   Show this help
    version, v                Show version

EXAMPLES:
    glab-tui                          # Start TUI
    glab-tui speed                    # 🔥 CHALLENGE MODE
    glab-tui pipelines                # List pipelines in CLI
    glab-tui job 11098249149         # Check specific job
    glab-tui logs 11098249149        # Show job logs
    glab-tui test-real               # Test GitLab connection
    glab-tui help                    # Show help`)
}
