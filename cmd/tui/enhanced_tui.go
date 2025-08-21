package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rkristelijn/glab-tui/internal/api"
)

// EnhancedTUIModel - The DOMINATING TUI model
type EnhancedTUIModel struct {
	client      *api.GitLabClient
	pipelines   []api.Pipeline
	jobs        map[int][]api.Job
	currentView string
	projectPath string
	lastUpdate  time.Time
	loading     bool
	error       string
	autoRefresh bool
}

// NewEnhancedTUI creates the ENHANCED TUI that will DOMINATE
func NewEnhancedTUI(projectPath string) (EnhancedTUIModel, error) {
	client, err := api.NewGitLabClient()
	if err != nil {
		return EnhancedTUIModel{}, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	// Test connection immediately
	if err := client.TestConnection(); err != nil {
		return EnhancedTUIModel{}, fmt.Errorf("GitLab connection failed: %w", err)
	}

	return EnhancedTUIModel{
		client:      client,
		jobs:        make(map[int][]api.Job),
		currentView: "pipelines",
		projectPath: projectPath,
		autoRefresh: true,
	}, nil
}

type enhancedUpdateMsg struct {
	pipelines []api.Pipeline
	jobs      map[int][]api.Job
	error     string
}

type enhancedTickMsg time.Time

// Init initializes the enhanced TUI
func (m EnhancedTUIModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchData(),
		tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return enhancedTickMsg(t)
		}),
	)
}

// Update handles enhanced TUI updates
func (m EnhancedTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, m.fetchData()
		case "space":
			m.autoRefresh = !m.autoRefresh
			if m.autoRefresh {
				return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
					return enhancedTickMsg(t)
				})
			}
		}

	case enhancedUpdateMsg:
		m.pipelines = msg.pipelines
		m.jobs = msg.jobs
		m.error = msg.error
		m.loading = false
		m.lastUpdate = time.Now()

	case enhancedTickMsg:
		if m.autoRefresh {
			return m, tea.Batch(
				m.fetchData(),
				tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
					return enhancedTickMsg(t)
				}),
			)
		}
	}

	return m, nil
}

// View renders the ENHANCED TUI
func (m EnhancedTUIModel) View() string {
	var b strings.Builder

	// Header - DOMINATION MODE
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6B35")).
		Background(lipgloss.Color("#1A1A1A")).
		Padding(0, 1)

	b.WriteString(headerStyle.Render("🚀 GLAB-TUI ENHANCED - GITLAB DOMINATION MODE"))
	b.WriteString("\n\n")

	// Connection status
	if m.error != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		b.WriteString(errorStyle.Render(fmt.Sprintf("❌ Error: %s", m.error)))
		b.WriteString("\n\n")
	} else {
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
		b.WriteString(successStyle.Render("✅ Connected to GitLab API"))
		b.WriteString("\n")
	}

	// Project info
	projectStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#87CEEB"))
	b.WriteString(projectStyle.Render(fmt.Sprintf("📁 Project: %s", m.projectPath)))
	b.WriteString("\n")

	// Last update info
	elapsed := time.Since(m.lastUpdate)
	updateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	if m.loading {
		b.WriteString(updateStyle.Render("⏳ Loading..."))
	} else {
		b.WriteString(updateStyle.Render(fmt.Sprintf("🔄 Last update: %s ago", elapsed.Truncate(time.Second))))
	}
	b.WriteString("\n\n")

	// Pipelines section
	b.WriteString("📊 PIPELINES (ALL OF THEM!):\n")
	b.WriteString("─────────────────────────────────────────────────────────────\n")

	if len(m.pipelines) == 0 {
		b.WriteString("No pipelines found or still loading...\n")
	} else {
		for i, pipeline := range m.pipelines {
			if i >= 10 { // Show top 10
				break
			}

			// Status styling
			var statusStyle lipgloss.Style
			switch pipeline.Status {
			case "success":
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
			case "running":
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
			case "failed":
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
			case "pending":
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
			default:
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
			}

			// Time formatting
			timeAgo := time.Since(pipeline.CreatedAt).Truncate(time.Minute)

			b.WriteString(fmt.Sprintf("● #%d %s %s (%s ago)\n",
				pipeline.ID,
				statusStyle.Render(pipeline.Status),
				pipeline.Ref,
				timeAgo,
			))

			// Show jobs for this pipeline if we have them
			if jobs, exists := m.jobs[pipeline.ID]; exists && len(jobs) > 0 {
				for _, job := range jobs {
					jobStatusStyle := statusStyle // Use same color as pipeline
					b.WriteString(fmt.Sprintf("  └─ %s: %s (%s)\n",
						job.Name,
						jobStatusStyle.Render(job.Status),
						job.Stage,
					))
				}
			}
		}
	}

	b.WriteString("─────────────────────────────────────────────────────────────\n")

	// Status info
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	if m.autoRefresh {
		b.WriteString(statusStyle.Render("🔄 Auto-refresh: ON (3s interval)"))
	} else {
		b.WriteString("🔄 Auto-refresh: OFF")
	}
	b.WriteString("\n\n")

	// Controls
	controlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	b.WriteString(controlStyle.Render("Controls: [r]efresh [space]toggle auto-refresh [q]uit"))

	return b.String()
}

// fetchData fetches all the data we need
func (m EnhancedTUIModel) fetchData() tea.Cmd {
	return func() tea.Msg {
		// Get pipelines
		pipelines, err := m.client.GetPipelines(m.projectPath, 20)
		if err != nil {
			return enhancedUpdateMsg{error: err.Error()}
		}

		// Get jobs for each pipeline (top 5 pipelines only to avoid rate limits)
		jobs := make(map[int][]api.Job)
		for i, pipeline := range pipelines {
			if i >= 5 { // Limit to top 5 pipelines for performance
				break
			}

			pipelineJobs, err := m.client.GetJobs(m.projectPath, pipeline.ID)
			if err == nil { // Don't fail if jobs can't be fetched
				jobs[pipeline.ID] = pipelineJobs
			}
		}

		return enhancedUpdateMsg{
			pipelines: pipelines,
			jobs:      jobs,
		}
	}
}
