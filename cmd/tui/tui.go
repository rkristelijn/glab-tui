package tui

import (
	"fmt"
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
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
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
	pipelines []core.Pipeline
	pipelineCursor int
	pipelineSelected map[int]struct{}
	
	// Job view
	jobs []core.Job
	jobCursor int
	selectedPipelineID int
	
	// Log view
	logs string
	selectedJobID int
	
	// GitLab wrapper
	gitlab *gitlab.GlabWrapper
}

func initialModel() model {
	// Create GitLab wrapper
	wrapper := gitlab.NewGlabWrapper("group/project/frontend-apps")
	
	// Try to get real data using glab
	pipelines, err := wrapper.GetProjectPipelines(12345678) // Replace with your project ID
	
	if err != nil || len(pipelines) == 0 {
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
