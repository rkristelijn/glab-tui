package tui

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rkristelijn/glab-tui/internal/gitlab"
)

// SpeedModeModel - BLAZING FAST pipeline monitoring
type SpeedModeModel struct {
	pipelines   []string
	statuses    map[string]string
	lastUpdate  time.Time
	updateCount int
	wrapper     *gitlab.GlabWrapper
	autoRefresh bool
	refreshRate time.Duration
}

// NewSpeedMode creates a new speed monitoring model
func NewSpeedMode() SpeedModeModel {
	// Latest jobs Pipeline Q is monitoring + challenge targets
	latestJobs := []string{
		"11099442002", "11099441955", "11099441920", // Latest jobs
		"11099439598", "11099439587", "11099434107", // Pipeline Q's targets
	}

	return SpeedModeModel{
		pipelines:   latestJobs,
		statuses:    make(map[string]string),
		wrapper:     gitlab.NewGlabWrapper("theapsgroup/agility/frontend-apps"),
		autoRefresh: true,
		refreshRate: 2 * time.Second, // FASTER than Pipeline Q's 30 seconds!
	}
}

type speedUpdateMsg struct {
	pipeline string
	status   string
}

type speedTickMsg time.Time

// Init initializes the speed mode
func (m SpeedModeModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchAllPipelines(),
		tea.Tick(m.refreshRate, func(t time.Time) tea.Msg {
			return speedTickMsg(t)
		}),
	)
}

// Update handles speed mode updates
func (m SpeedModeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			return m, m.fetchAllPipelines()
		case "space":
			m.autoRefresh = !m.autoRefresh
			if m.autoRefresh {
				return m, tea.Tick(m.refreshRate, func(t time.Time) tea.Msg {
					return speedTickMsg(t)
				})
			}
		}

	case speedUpdateMsg:
		m.statuses[msg.pipeline] = msg.status
		m.updateCount++

	case speedTickMsg:
		if m.autoRefresh {
			return m, tea.Batch(
				m.fetchAllPipelines(),
				tea.Tick(m.refreshRate, func(t time.Time) tea.Msg {
					return speedTickMsg(t)
				}),
			)
		}
	}

	m.lastUpdate = time.Now()
	return m, nil
}

// View renders the speed mode interface
func (m SpeedModeModel) View() string {
	var b strings.Builder

	// Header with challenge info
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6B35")).
		Background(lipgloss.Color("#1A1A1A")).
		Padding(0, 1)

	b.WriteString(headerStyle.Render("🔥 SPEED CHALLENGE MODE - TUI vs CLI 🔥"))
	b.WriteString("\n\n")

	// Performance metrics
	elapsed := time.Since(m.lastUpdate)
	metricsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))

	b.WriteString(metricsStyle.Render(fmt.Sprintf(
		"⚡ SPEED STATS: %d pipelines | %d updates | %.1fs refresh | Last: %s ago",
		len(m.pipelines), m.updateCount, m.refreshRate.Seconds(), elapsed.Truncate(time.Second),
	)))
	b.WriteString("\n\n")

	// Pipeline status grid - FAST DISPLAY
	b.WriteString("📊 PIPELINE BATTLE STATUS:\n")
	b.WriteString("─────────────────────────────────────────────────────────────\n")

	for _, pipeline := range m.pipelines {
		status := m.statuses[pipeline]
		if status == "" {
			status = "fetching..."
		}

		// Color coding for speed
		var statusStyle lipgloss.Style
		switch {
		case strings.Contains(status, "success"):
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
		case strings.Contains(status, "running"):
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
		case strings.Contains(status, "failed"):
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		default:
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		}

		// Special marking for parent pipeline
		marker := "●"
		if pipeline == "1997047926" {
			marker = "🎯" // Parent pipeline
		}

		b.WriteString(fmt.Sprintf("%s #%s %s\n",
			marker, pipeline, statusStyle.Render(status)))
	}

	b.WriteString("─────────────────────────────────────────────────────────────\n")

	// Challenge status
	challengeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF1493"))

	runningCount := 0
	for _, status := range m.statuses {
		if strings.Contains(status, "running") {
			runningCount++
		}
	}

	if runningCount > 0 {
		b.WriteString(challengeStyle.Render(fmt.Sprintf(
			"🥊 CHALLENGE STATUS: %d pipelines still running - TUI MONITORING AT LIGHT SPEED!",
			runningCount,
		)))
	} else {
		b.WriteString(challengeStyle.Render("🏆 CHALLENGE COMPLETE - TUI WINS! All pipelines monitored!"))
	}

	b.WriteString("\n\n")

	// Controls
	controlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	b.WriteString(controlStyle.Render("Controls: [r]efresh [space]toggle auto-refresh [q]uit"))

	return b.String()
}

// fetchAllPipelines fetches status for all challenge pipelines
func (m SpeedModeModel) fetchAllPipelines() tea.Cmd {
	var cmds []tea.Cmd

	for _, pipeline := range m.pipelines {
		cmds = append(cmds, m.fetchPipelineStatus(pipeline))
	}

	return tea.Batch(cmds...)
}

// fetchPipelineStatus fetches status for a single job
func (m SpeedModeModel) fetchPipelineStatus(jobID string) tea.Cmd {
	return func() tea.Msg {
		// Use glab API to get job status
		cmd := exec.Command("glab", "api", fmt.Sprintf("jobs/%s", jobID), "-R", "theapsgroup/agility/frontend-apps")
		output, err := cmd.Output()
		if err != nil {
			return speedUpdateMsg{pipeline: jobID, status: "error: " + err.Error()}
		}

		// Parse JSON response
		var job struct {
			Name   string `json:"name"`
			Status string `json:"status"`
			Stage  string `json:"stage"`
		}

		if err := json.Unmarshal(output, &job); err != nil {
			return speedUpdateMsg{pipeline: jobID, status: "parse error"}
		}

		status := fmt.Sprintf("%s - %s (%s)", job.Name, job.Status, job.Stage)
		return speedUpdateMsg{pipeline: jobID, status: status}
	}
}
