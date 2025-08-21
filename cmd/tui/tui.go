package tui

import (
	"fmt"
	"os"
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
	// Check for SPEED MODE challenge
	if len(os.Args) > 1 && os.Args[1] == "speed" {
		fmt.Println("🔥 SPEED CHALLENGE MODE ACTIVATED!")
		fmt.Println("⚡ Monitoring Pipeline Q's challenge targets...")

		p := tea.NewProgram(NewSpeedMode(), tea.WithAltScreen())
		_, err := p.Run()
		return err
	}

	// ENHANCED TUI MODE (the new default!)
	fmt.Println("🚀 ENHANCED GITLAB TUI - DOMINATION MODE!")
	fmt.Println("⚡ Connecting to GitLab API...")

	// Auto-detect current project
	projectPath, err := getCurrentProjectPath()
	if err != nil {
		fmt.Printf("❌ Could not detect GitLab project: %v\n", err)
		fmt.Println("💡 Make sure you're in a GitLab repository and authenticated with 'glab auth login'")
		return err
	}

	model, err := NewEnhancedTUI(projectPath)
	if err != nil {
		fmt.Printf("❌ Failed to initialize enhanced TUI: %v\n", err)
		fmt.Println("💡 Make sure you're authenticated with 'glab auth login'")
		return err
	}

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

func initialModel() model {
	// Try to get current project from git context
	projectPath, err := getCurrentProjectPath()
	if err != nil {
		// Fall back to mock data
		return model{
			currentView:      pipelineView,
			pipelines:        core.GetMockPipelines(),
			pipelineCursor:   0,
			pipelineSelected: make(map[int]struct{}),
			gitlab:           gitlab.NewGlabWrapper("mock-project"),
		}
	}

	// Create GitLab wrapper with detected project
	wrapper := gitlab.NewGlabWrapper(projectPath)

	// Try to get real data using the same approach as CLI
	pipelines, err := getProjectPipelinesViaGlab(projectPath)
	if err != nil {
		// Fall back to mock data
		pipelines = core.GetMockPipelines()
	}

	return model{
		currentView:      pipelineView,
		pipelines:        pipelines,
		pipelineCursor:   0,
		pipelineSelected: make(map[int]struct{}),
		gitlab:           wrapper,
	}
}

// Add the same helper functions from CLI
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
		case " ":
			// Toggle selection (only in pipeline view)
			if m.currentView == pipelineView {
				_, ok := m.pipelineSelected[m.pipelineCursor]
				if ok {
					delete(m.pipelineSelected, m.pipelineCursor)
				} else {
					m.pipelineSelected[m.pipelineCursor] = struct{}{}
				}
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	// Title bar
	title := titleStyle.Render("🚀 GitLab TUI - glab-tui")

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
	s += header + "                                                     ↻ Auto-refresh\n\n"

	// Pipeline list
	for i, pipeline := range m.pipelines {
		cursor := "  "
		if m.pipelineCursor == i {
			cursor = "> "
		}

		checked := " "
		if _, ok := m.pipelineSelected[i]; ok {
			checked = "x"
		}

		status := getStatusIcon(pipeline.Status)
		statusStyled := getStyledStatus(pipeline.Status, status)

		line := fmt.Sprintf("%s[%s] %s #%d  %-10s %-15s %-20s %s",
			cursor, checked, statusStyled, pipeline.ID, pipeline.Status, pipeline.ProjectName, pipeline.Ref, pipeline.Jobs)

		if m.pipelineCursor == i {
			line = selectedStyle.Render(line)
		}

		s += line + "\n"
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("Press Enter to view jobs, j/k to navigate, space to select, q to quit")
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

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("Press Enter to view logs, Esc to go back, j/k to navigate")
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

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("Press Esc to go back, logs auto-refresh for running jobs")
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
