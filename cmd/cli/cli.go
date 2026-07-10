package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rkristelijn/glab-tui/cmd/tui"
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
	case "auto":
		// Auto mode: try TUI first, fallback to selector
		autoMode()
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
			fmt.Println("Usage: glab-tui logs [--follow] <job-id>")
			fmt.Println("  --follow, -f    Stream logs in real-time")
			os.Exit(1)
		}

		// Parse flags and job ID
		var follow bool
		var jobIDStr string

		for i := 1; i < len(args); i++ {
			arg := args[i]
			if arg == "--follow" || arg == "-f" {
				follow = true
			} else if !strings.HasPrefix(arg, "-") {
				jobIDStr = arg
				break
			}
		}

		if jobIDStr == "" {
			fmt.Println("Error: job ID required")
			fmt.Println("Usage: glab-tui logs [--follow] <job-id>")
			os.Exit(1)
		}

		if follow {
			streamJobLogs(jobIDStr)
		} else {
			showJobLogs(jobIDStr)
		}
	case "tests", "t":
		if len(args) < 2 {
			fmt.Println("Usage: glab-tui tests <pipeline-id>")
			fmt.Println("       glab-tui tests latest")
			os.Exit(1)
		}
		showTestResults(args[1])
	case "test-real":
		testRealGitLab()
	case "help", "h", "--help":
		showHelp()
	case "demo", "d":
		// Demo mode with mock data
		fmt.Println("🎯 Starting glab-tui in DEMO mode with mock data...")
		startTUIWithMockData()
	case "remote", "url":
		// Remote GitLab project mode
		if len(args) < 2 {
			fmt.Println("Usage: glab-tui remote <gitlab-url>")
			fmt.Println("Example: glab-tui remote https://gitlab.com/group/project-name")
			os.Exit(1)
		}
		startTUIWithRemoteURL(args[1])
	case "select", "s":
		// Repository selector mode
		startRepositorySelector()
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

	// Try to get current project from git context
	projectPath, err := getCurrentProjectPath()
	if err != nil {
		fmt.Printf("❌ Could not detect GitLab project: %v\n", err)
		fmt.Println("💡 Make sure you're in a GitLab repository")
		os.Exit(1)
	}

	fmt.Printf("📊 Project: %s\n", projectPath)

	// Use glab command to get job details with project context
	// First try the project-specific endpoint
	cmd := exec.Command("glab", "api", fmt.Sprintf("projects/%s/jobs/%d", strings.ReplaceAll(projectPath, "/", "%2F"), jobID))
	output, err := cmd.Output()
	if err != nil {
		// Check if it's a 404 (job not found) vs auth issue
		if strings.Contains(string(output), "404 Not Found") {
			fmt.Printf("❌ Job %d not found\n", jobID)
			fmt.Println("💡 Make sure the job ID is correct and from the current project")
		} else {
			fmt.Printf("❌ Failed to get job details: %v\n", err)
			fmt.Println("💡 Make sure you're authenticated with 'glab auth login'")
		}
		os.Exit(1)
	}

	// Parse the JSON response to extract key info
	jobInfo, err := parseJobJSON(string(output))
	if err != nil {
		fmt.Printf("❌ Failed to parse job data: %v\n", err)
		os.Exit(1)
	}

	// Display job information
	fmt.Printf("✅ Job %d details:\n", jobID)
	fmt.Printf("   Name: %s\n", jobInfo.Name)
	fmt.Printf("   Status: %s\n", jobInfo.Status)
	fmt.Printf("   Stage: %s\n", jobInfo.Stage)
	if jobInfo.Duration != "" {
		fmt.Printf("   Duration: %s\n", jobInfo.Duration)
	}
}

// Simple JSON parser for job info
func parseJobJSON(jsonStr string) (core.Job, error) {
	// For now, use a simple approach to extract key fields
	// In a production version, we'd use proper JSON parsing

	job := core.Job{}

	// Extract name
	if nameStart := strings.Index(jsonStr, `"name":"`); nameStart != -1 {
		nameStart += 8
		if nameEnd := strings.Index(jsonStr[nameStart:], `"`); nameEnd != -1 {
			job.Name = jsonStr[nameStart : nameStart+nameEnd]
		}
	}

	// Extract status
	if statusStart := strings.Index(jsonStr, `"status":"`); statusStart != -1 {
		statusStart += 10
		if statusEnd := strings.Index(jsonStr[statusStart:], `"`); statusEnd != -1 {
			job.Status = jsonStr[statusStart : statusStart+statusEnd]
		}
	}

	// Extract stage
	if stageStart := strings.Index(jsonStr, `"stage":"`); stageStart != -1 {
		stageStart += 9
		if stageEnd := strings.Index(jsonStr[stageStart:], `"`); stageEnd != -1 {
			job.Stage = jsonStr[stageStart : stageStart+stageEnd]
		}
	}

	// Extract duration if available
	if durationStart := strings.Index(jsonStr, `"duration":`); durationStart != -1 {
		durationStart += 11
		if durationEnd := strings.Index(jsonStr[durationStart:], `,`); durationEnd != -1 {
			durationStr := jsonStr[durationStart : durationStart+durationEnd]
			if durationStr != "null" {
				job.Duration = durationStr + "s"
			}
		}
	}

	return job, nil
}

func streamJobLogs(jobIDStr string) {
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		fmt.Printf("Invalid job ID: %s\n", jobIDStr)
		os.Exit(1)
	}

	// Auto-detect current project
	projectPath, err := getCurrentProjectPath()
	if err != nil {
		fmt.Printf("❌ Could not detect GitLab project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔄 Streaming logs for job %d (Ctrl+C to exit)...\n", jobID)
	fmt.Println("─────────────────────────────────────────────────")

	wrapper := gitlab.NewGlabWrapper(projectPath)

	// Track last log position to avoid duplicates
	var lastLogSize int64 = 0

	// Setup signal handling for graceful exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create ticker for polling
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Initial log fetch
	logs, err := wrapper.GetJobLogs(jobID)
	if err != nil {
		fmt.Printf("❌ Failed to get job logs: %v\n", err)
		os.Exit(1)
	}

	if logs != "" {
		fmt.Print(logs)
		lastLogSize = int64(len(logs))
	}

	// Check initial job status
	jobStatus, err := wrapper.GetJobStatus(jobID)
	if err == nil && (jobStatus == "success" || jobStatus == "failed" || jobStatus == "canceled") {
		fmt.Printf("\n─────────────────────────────────────────────────\n")
		fmt.Printf("✅ Job %d completed with status: %s\n", jobID, jobStatus)
		return
	}

	for {
		select {
		case <-sigChan:
			fmt.Printf("\n─────────────────────────────────────────────────\n")
			fmt.Printf("🛑 Log streaming stopped\n")
			return

		case <-ticker.C:
			// Get current logs
			currentLogs, err := wrapper.GetJobLogs(jobID)
			if err != nil {
				fmt.Printf("\n❌ Error fetching logs: %v\n", err)
				continue
			}

			// Check if there are new logs
			currentSize := int64(len(currentLogs))
			if currentSize > lastLogSize {
				// Print only new content
				newContent := currentLogs[lastLogSize:]
				fmt.Print(newContent)
				lastLogSize = currentSize
			}

			// Check job status
			jobStatus, err := wrapper.GetJobStatus(jobID)
			if err == nil && (jobStatus == "success" || jobStatus == "failed" || jobStatus == "canceled") {
				fmt.Printf("\n─────────────────────────────────────────────────\n")
				fmt.Printf("✅ Job %d completed with status: %s\n", jobID, jobStatus)
				return
			}
		}
	}
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
	cmd := exec.Command("glab", "pipeline", "list", "-R", "group/project-name")
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
    glab-tui                    🔍 Repository selector (default)
    glab-tui [COMMAND]          Run CLI command

COMMANDS:
    select, s                  🔍 Show repository selector (default)
    pipelines, p               List pipelines
    job, j <job-id>           Check specific job status
    logs, l [--follow] <job-id>  Show job logs
        --follow, -f          🔥 Stream logs in real-time
    tests, t <pipeline-id>    📊 Show test results from pipeline
    demo, d                   🎯 Demo mode with mock data (for non-GitLab repos)
    remote, url <gitlab-url>  🌐 Connect to remote GitLab project
    test-real                 Test GitLab API connection
    speed                     🔥 Speed challenge mode
    help, h                   Show this help
    version, v                Show version

EXAMPLES:
    glab-tui                          # 🔍 Repository selector (default)
    glab-tui select                   # 🔍 Choose from your repositories
    glab-tui demo                     # 🎯 Demo mode (works anywhere!)
    glab-tui remote https://gitlab.com/group/project-name  # 🌐 Remote project
    glab-tui url https://gitlab.com/group/project  # 🌐 Remote project (short)
    glab-tui speed                    # 🔥 CHALLENGE MODE
    glab-tui pipelines                # List pipelines in CLI
    glab-tui job 11098249149         # Check specific job
    glab-tui logs 11098249149        # Show job logs (static)
    glab-tui logs --follow 11098249149  # 🔥 Stream logs in real-time
    glab-tui logs -f 11098249149     # 🔥 Stream logs (short flag)
    glab-tui tests latest            # 📊 Show test results from latest pipeline
    glab-tui tests 2027582415        # 📊 Show test results from specific pipeline
    glab-tui test-real               # Test GitLab connection
    glab-tui help                    # Show help

CONFIGURATION:
    Create a .env file to set your default GitLab group:
    DEFAULT_GROUP=your-org  # Group level for repository selection
    DEFAULT_GROUP=your-org/your-project  # Specific project

REPOSITORY SELECTOR:
    The repository selector supports search:
    /search-term    Filter repositories (e.g., '/social', '/frontend')
    clear           Clear current search
    1-20            Select repository by number
    tui             Start local TUI (if in GitLab repo)
    q               Quit`)
}

// startTUIWithMockData starts TUI with mock data for demo purposes
func startTUIWithMockData() {
	fmt.Println("🎯 Demo Mode: Using mock GitLab data for demonstration")
	fmt.Println("📝 This shows how glab-tui works with real GitLab projects")
	fmt.Println("")

	// Import TUI package and start with mock data
	tui.StartWithMockData()
}

// startTUIWithRemoteURL starts TUI with remote GitLab URL
func startTUIWithRemoteURL(gitlabURL string) {
	fmt.Printf("🌐 Connecting to remote GitLab: %s\n", gitlabURL)
	fmt.Println("📡 Loading pipeline data from remote repository...")
	fmt.Println("")

	// Parse GitLab URL to extract project path
	projectPath := parseGitLabURL(gitlabURL)
	if projectPath == "" {
		fmt.Printf("❌ Invalid GitLab URL format: %s\n", gitlabURL)
		fmt.Println("💡 Expected format: https://gitlab.com/group/project")
		os.Exit(1)
	}

	// Start TUI with remote project
	tui.StartWithRemoteProject(projectPath)
}

// parseGitLabURL extracts project path from GitLab URL
func parseGitLabURL(url string) string {
	// Remove common suffixes
	url = strings.TrimSuffix(url, "/-/pipelines")
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimSuffix(url, "/")

	// Extract project path from URL
	if strings.Contains(url, "gitlab.com/") {
		parts := strings.Split(url, "gitlab.com/")
		if len(parts) == 2 {
			return parts[1]
		}
	}

	return ""
}

// startRepositorySelector shows a list of available GitLab repositories
func startRepositorySelector() {
	fmt.Println("🔍 Repository Selector")
	fmt.Println("═══════════════════════")
	
	// Check if we have a .env file with default group
	if defaultGroup := getDefaultGroup(); defaultGroup != "" {
		fmt.Printf("📁 Default group: %s\n", defaultGroup)
		
		// Check if it's a specific project (contains multiple slashes) or just a group
		if strings.Count(defaultGroup, "/") >= 2 {
			// It's a specific project like "theapsgroup/social-networker/Services/social-networker"
			fmt.Print("Use default project? [Y/n]: ")
			var input string
			fmt.Scanln(&input)
			
			if input == "" || strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
				fmt.Printf("🚀 Starting TUI for project: %s\n", defaultGroup)
				tui.StartWithGroup(defaultGroup)
				return
			}
		} else {
			// It's just a group like "theapsgroup" - show repositories from that group
			fmt.Printf("🔍 Fetching repositories from group: %s\n", defaultGroup)
			showGroupRepositories(defaultGroup)
			return
		}
	}
	
	fmt.Println("🔍 Fetching your GitLab repositories...")
	
	// Get user's repositories using glab
	cmd := exec.Command("glab", "repo", "list", "--per-page", "20")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("❌ Failed to get repositories: %v\n", err)
		fmt.Println("💡 Make sure you're authenticated with 'glab auth login'")
		
		// Offer to create .env file
		fmt.Print("\nWould you like to create a .env file with your GitLab settings? [y/N]: ")
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			createEnvFile()
		}
		os.Exit(1)
	}

	repos := parseRepositoryList(string(output))
	if len(repos) == 0 {
		fmt.Println("❌ No repositories found")
		fmt.Println("💡 Try: glab auth login")
		os.Exit(1)
	}

	// Show repositories with search functionality
	showRepositoriesWithSearch(repos)
}

// showRepositoriesWithSearch displays repositories with search capability
func showRepositoriesWithSearch(repos []string) {
	filteredRepos := repos
	searchTerm := ""

	for {
		// Clear screen and show header
		fmt.Print("\033[2J\033[H") // Clear screen
		fmt.Println("🔍 Repository Selector")
		fmt.Println("═══════════════════════")
		
		if searchTerm != "" {
			fmt.Printf("🔍 Search: '%s' (showing %d/%d repos)\n", searchTerm, len(filteredRepos), len(repos))
		} else {
			fmt.Printf("📋 Found %d repositories:\n", len(repos))
		}
		fmt.Println()

		// Show filtered repositories
		for i, repo := range filteredRepos {
			if i >= 20 { // Limit display to 20
				fmt.Printf("... and %d more (use search to filter)\n", len(filteredRepos)-20)
				break
			}
			fmt.Printf("%2d. %s\n", i+1, repo)
		}
		
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  1-20: Select repository")
		fmt.Println("  /search-term: Filter repositories (e.g., '/social', '/supplier')")
		fmt.Println("  clear: Clear search")
		fmt.Println("  all: Show all accessible repositories")
		fmt.Println("  tui: Start local TUI")
		fmt.Println("  q: Quit")
		fmt.Print("\nEnter command: ")
		
		var input string
		fmt.Scanln(&input)
		
		if input == "q" || input == "quit" {
			fmt.Println("👋 Goodbye!")
			return
		}
		
		if input == "tui" {
			fmt.Println("🚀 Starting local TUI...")
			if err := tui.Run(); err != nil {
				fmt.Printf("TUI error: %v\n", err)
				os.Exit(1)
			}
			return
		}
		
		if input == "clear" {
			searchTerm = ""
			filteredRepos = repos
			continue
		}
		
		if input == "all" {
			// Get all accessible repositories
			fmt.Println("🔍 Fetching all accessible repositories...")
			cmd := exec.Command("glab", "repo", "list", "--all", "--per-page", "100")
			output, err := cmd.Output()
			if err != nil {
				fmt.Printf("❌ Failed to get all repositories: %v\n", err)
				fmt.Print("Press Enter to continue...")
				fmt.Scanln()
				continue
			}
			allRepos := parseRepositoryList(string(output))
			if len(allRepos) > 0 {
				repos = allRepos
				filteredRepos = allRepos
				searchTerm = ""
			}
			continue
		}
		
		if strings.HasPrefix(input, "/") {
			// Search command
			searchTerm = strings.TrimPrefix(input, "/")
			filteredRepos = filterRepositories(repos, searchTerm)
			if len(filteredRepos) == 0 {
				fmt.Printf("❌ No repositories found matching '%s'\n", searchTerm)
				fmt.Print("Press Enter to continue...")
				fmt.Scanln()
			}
			continue
		}
		
		// Handle search without / prefix for convenience
		if !strings.Contains(input, " ") && len(input) > 2 {
			// Looks like a search term, try searching
			searchTerm = input
			filteredRepos = filterRepositories(repos, searchTerm)
			if len(filteredRepos) == 0 {
				fmt.Printf("❌ No repositories found matching '%s'\n", searchTerm)
				fmt.Printf("💡 Tip: Use '/search-term' format or try 'all' to see more repositories\n")
				fmt.Print("Press Enter to continue...")
				fmt.Scanln()
			}
			continue
		}
		
		// Try to parse as number
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(filteredRepos) || choice > 20 {
			fmt.Println("❌ Invalid selection")
			fmt.Print("Press Enter to continue...")
			fmt.Scanln()
			continue
		}
		
		selectedRepo := filteredRepos[choice-1]
		fmt.Printf("🚀 Starting TUI for: %s\n", selectedRepo)
		
		// Ask if they want to save this as default
		fmt.Print("Save as default? [y/N]: ")
		var saveInput string
		fmt.Scanln(&saveInput)
		if strings.ToLower(saveInput) == "y" || strings.ToLower(saveInput) == "yes" {
			saveDefaultGroup(selectedRepo)
		}
		
		// Start TUI with selected repository
		tui.StartWithRemoteProject(selectedRepo)
		return
	}
}

// filterRepositories filters repositories based on search term
func filterRepositories(repos []string, searchTerm string) []string {
	if searchTerm == "" {
		return repos
	}
	
	var filtered []string
	searchLower := strings.ToLower(searchTerm)
	
	for _, repo := range repos {
		if strings.Contains(strings.ToLower(repo), searchLower) {
			filtered = append(filtered, repo)
		}
	}
	
	return filtered
}

// showGroupRepositories shows repositories from a specific GitLab group
func showGroupRepositories(groupName string) {
	fmt.Printf("🔍 Fetching repositories from group: %s...\n", groupName)
	
	// Get repositories from specific group using glab
	cmd := exec.Command("glab", "repo", "list", "--group", groupName, "--per-page", "50")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("❌ Failed to get repositories from group %s: %v\n", groupName, err)
		fmt.Println("💡 Make sure you have access to this group and are authenticated")
		
		// Fallback to all repositories
		fmt.Println("🔄 Falling back to all your repositories...")
		cmd = exec.Command("glab", "repo", "list", "--per-page", "50")
		output, err = cmd.Output()
		if err != nil {
			fmt.Printf("❌ Failed to get repositories: %v\n", err)
			fmt.Println("💡 Make sure you're authenticated with 'glab auth login'")
			os.Exit(1)
		}
	}

	repos := parseRepositoryList(string(output))
	if len(repos) == 0 {
		fmt.Printf("❌ No repositories found in group %s\n", groupName)
		fmt.Println("💡 Try: glab auth login")
		os.Exit(1)
	}

	// Show repositories with search functionality
	showRepositoriesWithSearch(repos)
}

// parseRepositoryList parses the output of 'glab repo list'
func parseRepositoryList(output string) []string {
	lines := strings.Split(output, "\n")
	var repos []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Showing") {
			continue
		}
		
		// Extract repository path from line like "group/project-name"
		if strings.Contains(line, "/") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				repos = append(repos, parts[0])
			}
		}
	}
	
	return repos
}
type TestSuite struct {
	Name         string     `json:"name"`
	TotalCount   int        `json:"total_count"`
	SuccessCount int        `json:"success_count"`
	FailedCount  int        `json:"failed_count"`
	ErrorCount   int        `json:"error_count"`
	SkippedCount int        `json:"skipped_count"`
	TestCases    []TestCase `json:"test_cases"`
}

// TestCase represents a test case from GitLab API
type TestCase struct {
	Name      string `json:"name"`
	Classname string `json:"classname"`
	Status    string `json:"status"`
	Time      float64 `json:"execution_time"`
}

// TestReport represents the full test report from GitLab API
type TestReport struct {
	TotalCount   int         `json:"total_count"`
	SuccessCount int         `json:"success_count"`
	FailedCount  int         `json:"failed_count"`
	ErrorCount   int         `json:"error_count"`
	SkippedCount int         `json:"skipped_count"`
	TestSuites   []TestSuite `json:"test_suites"`
}

func showTestResults(pipelineID string) {
	fmt.Printf("📊 Fetching test results for pipeline %s...\n", pipelineID)

	// Handle "latest" keyword
	if pipelineID == "latest" {
		cmd := exec.Command("glab", "api", "projects/:id/pipelines")
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("❌ Failed to get latest pipeline: %v\n", err)
			os.Exit(1)
		}

		var pipelines []map[string]interface{}
		if err := json.Unmarshal(output, &pipelines); err != nil {
			fmt.Printf("❌ Failed to parse pipelines: %v\n", err)
			os.Exit(1)
		}

		if len(pipelines) == 0 {
			fmt.Println("❌ No pipelines found")
			os.Exit(1)
		}

		pipelineID = fmt.Sprintf("%.0f", pipelines[0]["id"].(float64))
		fmt.Printf("📊 Using latest pipeline: %s\n", pipelineID)
	}

	// Get test report
	cmd := exec.Command("glab", "api", fmt.Sprintf("projects/:id/pipelines/%s/test_report", pipelineID))
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("❌ Failed to get test results: %v\n", err)
		fmt.Println("💡 Make sure the pipeline has test results and you have access")
		os.Exit(1)
	}

	var testReport TestReport
	if err := json.Unmarshal(output, &testReport); err != nil {
		fmt.Printf("❌ Failed to parse test results: %v\n", err)
		os.Exit(1)
	}

	// Display summary
	fmt.Printf("\n📊 Test Results Summary (Pipeline %s)\n", pipelineID)
	fmt.Printf("═══════════════════════════════════════════════════════════\n")
	fmt.Printf("Total Tests: %d | ✅ Success: %d | ❌ Failed: %d | ⚠️ Errors: %d | ⏭️ Skipped: %d\n",
		testReport.TotalCount, testReport.SuccessCount, testReport.FailedCount, 
		testReport.ErrorCount, testReport.SkippedCount)
	
	if testReport.TotalCount > 0 {
		successRate := (testReport.SuccessCount * 100) / testReport.TotalCount
		fmt.Printf("Success Rate: %d%%\n", successRate)
	}
	fmt.Println()

	// Display test suites
	for _, suite := range testReport.TestSuites {
		if suite.TotalCount == 0 {
			continue // Skip empty suites
		}

		fmt.Printf("## 🔍 %s\n", suite.Name)
		fmt.Printf("Tests: %d | ✅ %d | ❌ %d | ⚠️ %d | ⏭️ %d",
			suite.TotalCount, suite.SuccessCount, suite.FailedCount, 
			suite.ErrorCount, suite.SkippedCount)
		
		if suite.TotalCount > 0 {
			successRate := (suite.SuccessCount * 100) / suite.TotalCount
			fmt.Printf(" | Success: %d%%", successRate)
		}
		fmt.Println()

		// Show first few test cases
		maxCases := 5
		if len(suite.TestCases) < maxCases {
			maxCases = len(suite.TestCases)
		}

		for i := 0; i < maxCases; i++ {
			testCase := suite.TestCases[i]
			var icon string
			switch testCase.Status {
			case "success":
				icon = "✅"
			case "failed":
				icon = "❌"
			case "error":
				icon = "⚠️ "
			case "skipped":
				icon = "⏭️"
			default:
				icon = "❓"
			}
			fmt.Printf("  %s %s (%s)\n", icon, testCase.Name, testCase.Classname)
		}

		if len(suite.TestCases) > maxCases {
			fmt.Printf("  ... and %d more tests\n", len(suite.TestCases)-maxCases)
		}
		fmt.Println()
	}

	if len(testReport.TestSuites) == 0 {
		fmt.Println("ℹ️  No test suites found in this pipeline")
		fmt.Println("💡 Make sure your jobs generate JUnit XML reports")
	}
}

// getDefaultGroup reads the default group from .env file
func getDefaultGroup() string {
	file, err := os.Open(".env")
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "DEFAULT_GROUP=") {
			return strings.TrimPrefix(line, "DEFAULT_GROUP=")
		}
	}
	return ""
}

// saveDefaultGroup saves the default group to .env file
func saveDefaultGroup(group string) {
	// Read existing .env content
	var lines []string
	file, err := os.Open(".env")
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "DEFAULT_GROUP=") {
				lines = append(lines, line)
			}
		}
		file.Close()
	}

	// Add the new default group
	lines = append(lines, fmt.Sprintf("DEFAULT_GROUP=%s", group))

	// Write back to .env
	file, err = os.Create(".env")
	if err != nil {
		fmt.Printf("❌ Failed to save default group: %v\n", err)
		return
	}
	defer file.Close()

	for _, line := range lines {
		fmt.Fprintln(file, line)
	}

	fmt.Printf("✅ Saved %s as default group in .env\n", group)
}

// createEnvFile creates a new .env file with basic settings
func createEnvFile() {
	fmt.Println("📝 Creating .env file...")
	
	file, err := os.Create(".env")
	if err != nil {
		fmt.Printf("❌ Failed to create .env file: %v\n", err)
		return
	}
	defer file.Close()

	envContent := `# GitLab Configuration
GITLAB_URL=https://gitlab.com
GITLAB_TOKEN=your-token-here

# Default group/project (set by glab-tui select)
# DEFAULT_GROUP=your-org/your-project

# UI Configuration
REFRESH_INTERVAL=5s
MAX_PIPELINES_PER_PROJECT=10
`

	fmt.Fprint(file, envContent)
	fmt.Println("✅ Created .env file")
	fmt.Println("💡 Set your GITLAB_TOKEN and run 'glab auth login'")
}

// autoMode tries to start TUI with local repo, falls back to selector
func autoMode() {
	// Try to detect current GitLab project
	projectPath, err := getCurrentProjectPath()
	if err != nil {
		// Not in a GitLab repo, show better error and go to selector
		if strings.Contains(err.Error(), "github.com") {
			fmt.Println("ℹ️  You're in a GitHub repository, not GitLab")
		} else if strings.Contains(err.Error(), "failed to get git remote") {
			fmt.Println("ℹ️  Not in a Git repository")
		} else {
			fmt.Printf("ℹ️  Not in a GitLab repository: %v\n", err)
		}
		fmt.Println("🔍 Switching to repository selector...")
		startRepositorySelector()
		return
	}

	// We're in a GitLab repo, start TUI directly
	fmt.Printf("🚀 Starting GitLab TUI for: %s\n", projectPath)
	if err := tui.StartWithRemoteProject(projectPath); err != nil {
		fmt.Printf("TUI error: %v\n", err)
		os.Exit(1)
	}
}
