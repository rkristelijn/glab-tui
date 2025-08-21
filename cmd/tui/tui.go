package tui

import (
	"fmt"
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
					jobs, err := m.gitlab.GetPipelineJobs(selectedPipeline.ID)
					if err == nil {
						m.jobs = jobs
						m.jobCursor = 0
						m.selectedPipelineID = selectedPipeline.ID
						m.currentView = jobView
					}
				}
			case jobView:
				// Enter job -> show logs
				if m.jobCursor < len(m.jobs) {
					selectedJob := m.jobs[m.jobCursor]
					logs, err := m.gitlab.GetJobLogs(selectedJob.ID)
					if err == nil {
						m.logs = logs
						m.selectedJobID = selectedJob.ID
						m.currentView = logView
					}
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
	header := headerStyle.Render("Pipelines")

	s := title + "\n"
	s += header + "                                                     [r] Refresh\n\n"

	// Pipeline list
	for i, pipeline := range m.pipelines {
		cursor := "  "
		if m.pipelineCursor == i {
			cursor = "> "
		}

		status := getStatusIcon(pipeline.Status)
		statusStyled := getStyledStatus(pipeline.Status, status)

		line := fmt.Sprintf("%s%s #%d  %-10s %-15s %-20s %s",
			cursor, statusStyled, pipeline.ID, pipeline.Status, pipeline.ProjectName, pipeline.Ref, pipeline.Jobs)

		if m.pipelineCursor == i {
			line = selectedStyle.Render(line)
		}

		s += line + "\n"
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("Enter: view jobs | j/k: navigate | r: refresh | q: quit")
	return s
}

func (m model) renderJobView(title string) string {
	header := headerStyle.Render(fmt.Sprintf("Jobs (Pipeline #%d)", m.selectedPipelineID))

	s := title + "\n"
	s += header + "\n\n"

	// Job list
	for i, job := range m.jobs {
		cursor := "  "
		if m.jobCursor == i {
			cursor = "> "
		}

		status := getStatusIcon(job.Status)
		statusStyled := getStyledStatus(job.Status, status)

		line := fmt.Sprintf("%s%s %-25s %-10s %-15s",
			cursor, statusStyled, job.Name, job.Status, job.Stage)

		if m.jobCursor == i {
			line = selectedStyle.Render(line)
		}

		s += line + "\n"
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("Enter: view logs | Esc: back | j/k: navigate")
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
