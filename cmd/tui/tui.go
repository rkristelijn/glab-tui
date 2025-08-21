package tui

import (
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rkristelijn/glab-tui/internal/core"
	"github.com/rkristelijn/glab-tui/internal/gitlab"
)

var (
	// Styles
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#F25D94")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EE6FF8")).
			Bold(true)

	runningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	failedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87"))

	pendingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF87"))
)

func Run() error {
	fmt.Println("🚀 GitLab TUI - Pipeline Monitor")
	fmt.Println("⚡ Loading pipeline data...")

	// Auto-detect current project
	projectPath, err := getCurrentProjectPath()
	if err != nil {
		fmt.Printf("❌ Could not detect GitLab project: %v\n", err)
		fmt.Println("💡 Make sure you're in a GitLab repository and authenticated with 'glab auth login'")
		return err
	}

	model := initialModel(projectPath)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

type viewMode int

const (
	pipelineView viewMode = iota
	jobView
	logView
)

type model struct {
	// View state
	currentView viewMode
	projectPath string

	// Pipeline view
	pipelines        []core.Pipeline
	pipelineCursor   int
	pipelineSelected map[int]struct{}

	// Job view
	jobs               []core.Job
	jobCursor          int
	selectedPipelineID int

	// Log view
	logs          string
	selectedJobID int

	// GitLab wrapper
	gitlab *gitlab.GlabWrapper
}

func initialModel(projectPath string) model {
	// Try to get real data using the same approach as CLI
	pipelines, err := getProjectPipelinesViaGlab(projectPath)
	if err != nil {
		// Fall back to mock data
		pipelines = core.GetMockPipelines()
	}

	return model{
		currentView:      pipelineView,
		projectPath:      projectPath,
		pipelines:        pipelines,
		pipelineCursor:   0,
		pipelineSelected: make(map[int]struct{}),
		gitlab:           gitlab.NewGlabWrapper(projectPath),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// StartWithMockData starts the TUI with mock data for demo purposes
func StartWithMockData() error {
	fmt.Println("🚀 Starting GitLab TUI Demo Mode...")

	// Create model with mock data
	m := model{
		currentView:      pipelineView,
		projectPath:      "demo-project",
		pipelines:        core.GetMockPipelines(),
		pipelineCursor:   0,
		pipelineSelected: make(map[int]struct{}),
		gitlab:           nil, // No real GitLab connection in demo mode
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// StartWithRemoteProject starts the TUI with a remote GitLab project
func StartWithRemoteProject(projectPath string) error {
	fmt.Printf("🚀 Starting GitLab TUI for remote project: %s\n", projectPath)

	// Try to get pipelines for the remote project
	pipelines, err := getRemoteProjectPipelines(projectPath)
	if err != nil {
		fmt.Printf("❌ Could not load pipelines from remote project: %v\n", err)
		fmt.Println("💡 Make sure you're authenticated with 'glab auth login'")
		return err
	}

	// Create model with remote data
	m := model{
		currentView:      pipelineView,
		projectPath:      projectPath,
		pipelines:        pipelines,
		pipelineCursor:   0,
		pipelineSelected: make(map[int]struct{}),
		gitlab:           nil, // Use nil for remote mode, handle in Update()
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// getRemoteProjectPipelines gets pipelines for a remote GitLab project
func getRemoteProjectPipelines(projectPath string) ([]core.Pipeline, error) {
	fmt.Printf("📡 Fetching real pipelines for %s...\n", projectPath)

	// Use glab pipeline list with remote project
	cmd := exec.Command("glab", "pipeline", "list", "--repo", projectPath, "--limit", "20")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("❌ Failed to get pipelines: %v\n", err)
		fmt.Println("💡 Make sure you're authenticated with 'glab auth login'")
		fmt.Println("📝 Falling back to mock data for demonstration")
		return core.GetMockPipelines(), nil
	}

	// Parse glab pipeline list output
	pipelines := parseGlabPipelineOutput(string(output))
	if len(pipelines) == 0 {
		fmt.Println("📝 No pipelines found, using mock data for demonstration")
		return core.GetMockPipelines(), nil
	}

	fmt.Printf("✅ Loaded %d real pipelines from %s\n", len(pipelines), projectPath)
	return pipelines, nil
}

// parseGlabPipelineOutput parses the output from 'glab pipeline list'
func parseGlabPipelineOutput(output string) []core.Pipeline {
	var pipelines []core.Pipeline
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "ID") || strings.HasPrefix(line, "--") {
			continue
		}

		// Parse line format: ID STATUS REF CREATED
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			// Extract ID
			idStr := fields[0]
			if id, err := strconv.Atoi(idStr); err == nil {
				pipeline := core.Pipeline{
					ID:          id,
					Status:      fields[1],
					Ref:         fields[2],
					ProjectName: "remote-project",
					Jobs:        "loading...",
					Duration:    "unknown",
				}

				// Add more fields if available
				if len(fields) >= 4 {
					pipeline.Duration = fields[3]
				}

				pipelines = append(pipelines, pipeline)
			}
		}
	}

	return pipelines
}

// getRemotePipelineJobs gets jobs for a pipeline in remote project
func getRemotePipelineJobs(projectPath string, pipelineID int) ([]core.Job, error) {
	fmt.Printf("📡 Fetching jobs for pipeline %d in %s...\n", pipelineID, projectPath)

	// Use glab API to get pipeline jobs
	cmd := exec.Command("glab", "api", fmt.Sprintf("projects/%s/pipelines/%d/jobs", url.QueryEscape(projectPath), pipelineID))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs: %w", err)
	}

	// Parse jobs from API response
	jobs := parseJobsFromAPI(string(output))
	return jobs, nil
}

// getRemoteJobLogs gets logs for a job in remote project
func getRemoteJobLogs(projectPath string, jobID int) (string, error) {
	fmt.Printf("📡 Fetching logs for job %d in %s...\n", jobID, projectPath)

	// Use glab API to get job logs
	cmd := exec.Command("glab", "api", fmt.Sprintf("projects/%s/jobs/%d/trace", url.QueryEscape(projectPath), jobID))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get job logs: %w", err)
	}

	return string(output), nil
}

// parseJobsFromAPI parses jobs from GitLab API response
func parseJobsFromAPI(jsonResponse string) []core.Job {
	var jobs []core.Job

	// Simple parsing - look for job objects
	lines := strings.Split(jsonResponse, "\n")
	var currentJob core.Job
	inJob := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, `"id":`) && !strings.Contains(line, `"pipeline_id":`) {
			if id := extractNumber(line, `"id":`); id != 0 {
				currentJob.ID = id
				inJob = true
			}
		}

		if strings.Contains(line, `"name":`) {
			if name := extractString(line, `"name":`); name != "" {
				currentJob.Name = name
			}
		}

		if strings.Contains(line, `"status":`) {
			if status := extractString(line, `"status":`); status != "" {
				currentJob.Status = status
			}
		}

		if strings.Contains(line, `"stage":`) {
			if stage := extractString(line, `"stage":`); stage != "" {
				currentJob.Stage = stage
			}
		}

		// End of job object
		if strings.Contains(line, "}") && inJob {
			if currentJob.ID != 0 && currentJob.Name != "" {
				jobs = append(jobs, currentJob)
			}
			currentJob = core.Job{}
			inJob = false
		}
	}

	return jobs
}

// extractNumber extracts a number from a JSON line
func extractNumber(line, key string) int {
	parts := strings.Split(line, key)
	if len(parts) > 1 {
		numPart := strings.TrimSpace(parts[1])
		numPart = strings.Split(numPart, ",")[0]
		numPart = strings.TrimSpace(numPart)
		if num, err := strconv.Atoi(numPart); err == nil {
			return num
		}
	}
	return 0
}

// extractString extracts a string from a JSON line
func extractString(line, key string) string {
	parts := strings.Split(line, key)
	if len(parts) > 1 {
		strPart := strings.TrimSpace(parts[1])
		strPart = strings.TrimPrefix(strPart, `"`)
		strPart = strings.Split(strPart, `"`)[0]
		return strPart
	}
	return ""
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			// Go back to previous view
			switch m.currentView {
			case jobView:
				m.currentView = pipelineView
			case logView:
				m.currentView = jobView
			}
		case "r":
			// Refresh pipelines
			if m.currentView == pipelineView {
				pipelines, err := getProjectPipelinesViaGlab(m.projectPath)
				if err == nil {
					m.pipelines = pipelines
				}
			}
		case "enter":
			// Drill down to next view
			switch m.currentView {
			case pipelineView:
				// Enter pipeline -> show jobs
				if m.pipelineCursor < len(m.pipelines) {
					selectedPipeline := m.pipelines[m.pipelineCursor]

					// Handle demo mode (when gitlab is nil)
					if m.gitlab == nil {
						// Check if this is remote mode (has real project path)
						if strings.Contains(m.projectPath, "/") {
							// Remote mode - fetch real jobs
							jobs, err := getRemotePipelineJobs(m.projectPath, selectedPipeline.ID)
							if err == nil {
								m.jobs = jobs
							} else {
								// Fallback to mock jobs
								m.jobs = core.GetMockJobs()
							}
						} else {
							// Demo mode - use mock jobs
							m.jobs = core.GetMockJobs()
						}
						m.jobCursor = 0
						m.selectedPipelineID = selectedPipeline.ID
						m.currentView = jobView
					} else {
						// Real GitLab mode
						jobs, err := m.gitlab.GetPipelineJobs(selectedPipeline.ID)
						if err == nil {
							m.jobs = jobs
							m.jobCursor = 0
							m.selectedPipelineID = selectedPipeline.ID
							m.currentView = jobView
						}
					}
				}
			case jobView:
				// Enter job -> show logs
				if m.jobCursor < len(m.jobs) {
					selectedJob := m.jobs[m.jobCursor]

					// Handle demo mode (when gitlab is nil)
					if m.gitlab == nil {
						// Check if this is remote mode
						if strings.Contains(m.projectPath, "/") {
							// Remote mode - fetch real logs
							logs, err := getRemoteJobLogs(m.projectPath, selectedJob.ID)
							if err == nil {
								m.logs = logs
							} else {
								m.logs = fmt.Sprintf("❌ Failed to get logs for job %d: %v\n\n💡 This is a remote project.\nTry using: glab-tui logs %d", selectedJob.ID, err, selectedJob.ID)
							}
						} else {
							// Demo mode - use mock logs
							m.logs = "🎯 Demo Mode - Mock Job Logs\n\n" +
								"📋 Job: " + selectedJob.Name + "\n" +
								"📊 Status: " + selectedJob.Status + "\n" +
								"🏗️  Stage: " + selectedJob.Stage + "\n\n" +
								"Sample log output:\n" +
								"[INFO] Starting job execution...\n" +
								"[INFO] Installing dependencies...\n" +
								"[INFO] Running tests...\n" +
								"[SUCCESS] All tests passed!\n" +
								"[INFO] Job completed successfully\n\n" +
								"💡 In real GitLab projects, you'd see actual job logs here.\n" +
								"🔥 Use 'l' key for real-time streaming in live projects!"
						}
						m.selectedJobID = selectedJob.ID
						m.currentView = logView
					} else {
						// Real GitLab mode
						logs, err := m.gitlab.GetJobLogs(selectedJob.ID)
						if err == nil {
							m.logs = logs
							m.selectedJobID = selectedJob.ID
							m.currentView = logView
						}
					}
				}
			}
		case "l":
			// Quick logs --follow for selected job
			if m.currentView == jobView && m.jobCursor < len(m.jobs) {
				selectedJob := m.jobs[m.jobCursor]

				// Handle demo mode
				if m.gitlab == nil {
					// Show demo message instead of trying to stream
					m.logs = "🎯 Demo Mode - Real-time Streaming Preview\n\n" +
						"📋 Job: " + selectedJob.Name + "\n" +
						"🔥 In a real GitLab project, this would start:\n" +
						"   glab-tui logs --follow " + strconv.Itoa(selectedJob.ID) + "\n\n" +
						"🔄 Live streaming features:\n" +
						"   • Real-time log updates every 2 seconds\n" +
						"   • Auto-completion detection\n" +
						"   • Graceful Ctrl+C exit\n" +
						"   • Live job status monitoring\n\n" +
						"💡 Try this in a real GitLab repository to see live streaming!\n" +
						"📝 Example: cd /path/to/gitlab/project && glab-tui"
					m.selectedJobID = selectedJob.ID
					m.currentView = logView
				} else {
					// Real GitLab mode - exit TUI and start streaming
					return m, tea.Sequence(
						tea.Quit,
						tea.ExecProcess(exec.Command("glab-tui", "logs", "--follow", strconv.Itoa(selectedJob.ID)), nil),
					)
				}
			}
		case "up", "k":
			switch m.currentView {
			case pipelineView:
				if m.pipelineCursor > 0 {
					m.pipelineCursor--
				}
			case jobView:
				if m.jobCursor > 0 {
					m.jobCursor--
				}
			}
		case "down", "j":
			switch m.currentView {
			case pipelineView:
				if m.pipelineCursor < len(m.pipelines)-1 {
					m.pipelineCursor++
				}
			case jobView:
				if m.jobCursor < len(m.jobs)-1 {
					m.jobCursor++
				}
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	// Title bar
	title := titleStyle.Render("🚀 GitLab TUI - " + getProjectName(m.projectPath))

	switch m.currentView {
	case pipelineView:
		return m.renderPipelineView(title)
	case jobView:
		return m.renderJobView(title)
	case logView:
		return m.renderLogView(title)
	default:
		return m.renderPipelineView(title)
	}
}

func (m model) renderPipelineView(title string) string {
	header := headerStyle.Render("🔄 Live Pipelines")

	// Show welcome message if no pipelines
	if len(m.pipelines) == 0 {
		return m.renderWelcomeScreen(title)
	}

	// Count running pipelines
	runningCount := 0
	for _, pipeline := range m.pipelines {
		if pipeline.Status == "running" {
			runningCount++
		}
	}

	statusLine := fmt.Sprintf("📊 %d total | 🔄 %d running | [r] Refresh | [Enter] View Jobs",
		len(m.pipelines), runningCount)

	s := title + "\n"
	s += header + "                                                     \n"
	s += lipgloss.NewStyle().Faint(true).Render(statusLine) + "\n\n"

	// Pipeline list with enhanced visualization
	for i, pipeline := range m.pipelines {
		cursor := "  "
		if m.pipelineCursor == i {
			cursor = "▶ "
		}

		// Enhanced status visualization
		status := getStatusIcon(pipeline.Status)
		statusStyled := getStyledStatus(pipeline.Status, status)

		// Progress bar for running pipelines
		progressBar := ""
		if pipeline.Status == "running" {
			progressBar = getProgressBar(pipeline.Jobs)
		}

		// Duration formatting
		duration := ""
		if pipeline.Status == "running" {
			duration = "⏱️  running..."
		} else {
			duration = fmt.Sprintf("⏱️  %s", pipeline.Duration)
		}

		// Enhanced line format with better spacing
		line := fmt.Sprintf("%s%s #%-8d %-12s %-20s %-15s %s %s",
			cursor,
			statusStyled,
			pipeline.ID,
			pipeline.Status,
			truncateString(pipeline.Ref, 20),
			truncateString(pipeline.ProjectName, 15),
			progressBar,
			duration)

		if m.pipelineCursor == i {
			line = selectedStyle.Render(line)
		}

		s += line + "\n"
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("Navigation: ↑/↓ or j/k | Enter: view jobs | r: refresh | q: quit")
	return s
}

// Welcome screen with instructions
func (m model) renderWelcomeScreen(title string) string {
	welcomeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Margin(1, 0)

	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Margin(0, 2)

	s := title + "\n\n"

	s += welcomeStyle.Render("🚀 Welcome to GitLab TUI!") + "\n\n"

	s += instructionStyle.Render("📋 Quick Start Guide:") + "\n"
	s += "  • Press 'r' to refresh and load pipelines\n"
	s += "  • Use ↑/↓ or j/k to navigate\n"
	s += "  • Press Enter to drill down: Pipelines → Jobs → Logs\n"
	s += "  • Press 'l' on a job to stream logs in real-time\n"
	s += "  • Press Esc to go back, 'q' to quit\n\n"

	s += instructionStyle.Render("🔥 New Feature - Real-time Log Streaming:") + "\n"
	s += "  • Navigate to a job and press 'l' for live streaming\n"
	s += "  • Or use CLI: glab-tui logs --follow <job-id>\n\n"

	s += instructionStyle.Render("🎯 Current Status:") + "\n"
	if len(m.pipelines) == 0 {
		s += "  • No pipelines loaded yet\n"
		s += "  • Press 'r' to refresh and load from GitLab\n"
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("Press 'r' to refresh | 'q' to quit")

	return s
}

// Helper function to create progress bar for running pipelines
func getProgressBar(jobs string) string {
	if jobs == "" {
		return "⬜⬜⬜⬜⬜"
	}

	// Simple progress visualization
	// In real implementation, this would parse actual job status
	return "🟩🟩🟨⬜⬜" // Example: 2 done, 1 running, 2 pending
}

// Helper function to truncate strings with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func (m model) renderJobView(title string) string {
	header := headerStyle.Render(fmt.Sprintf("🔧 Jobs (Pipeline #%d)", m.selectedPipelineID))

	// Count job statuses
	runningJobs := 0
	successJobs := 0
	failedJobs := 0
	for _, job := range m.jobs {
		switch job.Status {
		case "running":
			runningJobs++
		case "success":
			successJobs++
		case "failed":
			failedJobs++
		}
	}

	statusLine := fmt.Sprintf("📊 %d total | ✅ %d success | 🔄 %d running | ❌ %d failed",
		len(m.jobs), successJobs, runningJobs, failedJobs)

	s := title + "\n"
	s += header + "\n"
	s += lipgloss.NewStyle().Faint(true).Render(statusLine) + "\n\n"

	// Enhanced job list
	for i, job := range m.jobs {
		cursor := "  "
		if m.jobCursor == i {
			cursor = "▶ "
		}

		status := getStatusIcon(job.Status)
		statusStyled := getStyledStatus(job.Status, status)

		// Duration or status indicator
		statusInfo := ""
		if job.Status == "running" {
			statusInfo = "🔄 running..."
		} else if job.Status == "success" {
			statusInfo = "✅ completed"
		} else if job.Status == "failed" {
			statusInfo = "❌ failed"
		} else {
			statusInfo = fmt.Sprintf("⏸️  %s", job.Status)
		}

		line := fmt.Sprintf("%s%s %-30s %-12s %-15s %s",
			cursor,
			statusStyled,
			truncateString(job.Name, 30),
			job.Status,
			job.Stage,
			statusInfo)

		if m.jobCursor == i {
			line = selectedStyle.Render(line)
		}

		s += line + "\n"
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("Navigation: ↑/↓ or j/k | Enter: view logs | Esc: back to pipelines | l: logs --follow")
	return s
}

func (m model) renderLogView(title string) string {
	header := headerStyle.Render(fmt.Sprintf("Logs (Job #%d)", m.selectedJobID))

	s := title + "\n"
	s += header + "\n\n"

	// Show logs (truncated for display)
	logLines := strings.Split(m.logs, "\n")
	maxLines := 20 // Show last 20 lines
	startLine := 0
	if len(logLines) > maxLines {
		startLine = len(logLines) - maxLines
	}

	for i := startLine; i < len(logLines) && i < startLine+maxLines; i++ {
		s += logLines[i] + "\n"
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("Esc: back to jobs")
	return s
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

func getStyledStatus(status, icon string) string {
	switch status {
	case "running":
		return runningStyle.Render(icon)
	case "success":
		return successStyle.Render(icon)
	case "failed":
		return failedStyle.Render(icon)
	default:
		return pendingStyle.Render(icon)
	}
}

// Helper functions (same as CLI)
func getCurrentProjectPath() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))

	if strings.Contains(remoteURL, "gitlab") {
		if strings.HasPrefix(remoteURL, "git@") {
			parts := strings.Split(remoteURL, ":")
			if len(parts) >= 2 {
				return strings.TrimSuffix(parts[1], ".git"), nil
			}
		} else if strings.HasPrefix(remoteURL, "https://") {
			parts := strings.Split(remoteURL, "/")
			if len(parts) >= 4 {
				projectPath := strings.Join(parts[3:], "/")
				return strings.TrimSuffix(projectPath, ".git"), nil
			}
		}
	}

	return "", fmt.Errorf("not a GitLab repository or unsupported URL format: %s", remoteURL)
}

func getProjectPipelinesViaGlab(projectPath string) ([]core.Pipeline, error) {
	cmd := exec.Command("glab", "pipeline", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("glab command failed: %w", err)
	}

	return parseGlabPipelineText(string(output))
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

func getProjectName(projectPath string) string {
	parts := strings.Split(projectPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "project"
}
