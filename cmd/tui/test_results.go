package tui

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TestSuite struct {
	Name         string     `json:"name"`
	TotalCount   int        `json:"total_count"`
	SuccessCount int        `json:"success_count"`
	FailedCount  int        `json:"failed_count"`
	ErrorCount   int        `json:"error_count"`
	SkippedCount int        `json:"skipped_count"`
	TestCases    []TestCase `json:"test_cases"`
}

type TestCase struct {
	Name      string `json:"name"`
	Classname string `json:"classname"`
	Status    string `json:"status"`
	Time      float64 `json:"execution_time"`
	Message   string `json:"message"`
}

type TestReport struct {
	TotalCount   int         `json:"total_count"`
	SuccessCount int         `json:"success_count"`
	FailedCount  int         `json:"failed_count"`
	ErrorCount   int         `json:"error_count"`
	SkippedCount int         `json:"skipped_count"`
	TestSuites   []TestSuite `json:"test_suites"`
}

type testResultsModel struct {
	testReport   TestReport
	selectedSuite int
	selectedTest  int
	viewMode     string // "suites" or "tests" or "detail"
	width        int
	height       int
}

func NewTestResultsModel() testResultsModel {
	return testResultsModel{
		viewMode: "suites",
	}
}

func (m testResultsModel) Init() tea.Cmd {
	return fetchTestResults
}

func fetchTestResults() tea.Msg {
	cmd := exec.Command("glab", "api", "projects/:id/pipelines")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get pipelines: %v", err)
	}

	var pipelines []map[string]interface{}
	if err := json.Unmarshal(output, &pipelines); err != nil {
		return fmt.Errorf("failed to parse pipelines: %v", err)
	}

	if len(pipelines) == 0 {
		return fmt.Errorf("no pipelines found")
	}

	pipelineID := fmt.Sprintf("%.0f", pipelines[0]["id"].(float64))
	
	cmd = exec.Command("glab", "api", fmt.Sprintf("projects/:id/pipelines/%s/test_report", pipelineID))
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get test results: %v", err)
	}

	var testReport TestReport
	if err := json.Unmarshal(output, &testReport); err != nil {
		return fmt.Errorf("failed to parse test results: %v", err)
	}

	return testReport
}

func (m testResultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TestReport:
		m.testReport = msg
		return m, nil

	case error:
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			if m.viewMode == "detail" {
				m.viewMode = "tests"
			} else if m.viewMode == "tests" {
				m.viewMode = "suites"
			} else {
				return m, tea.Quit
			}
		case "enter":
			if m.viewMode == "suites" && len(m.testReport.TestSuites) > 0 {
				m.viewMode = "tests"
				m.selectedTest = 0
			} else if m.viewMode == "tests" && len(m.testReport.TestSuites[m.selectedSuite].TestCases) > 0 {
				m.viewMode = "detail"
			}
		case "up", "k":
			if m.viewMode == "suites" {
				if m.selectedSuite > 0 {
					m.selectedSuite--
				}
			} else if m.viewMode == "tests" {
				if m.selectedTest > 0 {
					m.selectedTest--
				}
			}
		case "down", "j":
			if m.viewMode == "suites" {
				if m.selectedSuite < len(m.testReport.TestSuites)-1 {
					m.selectedSuite++
				}
			} else if m.viewMode == "tests" {
				if m.selectedSuite < len(m.testReport.TestSuites) && m.selectedTest < len(m.testReport.TestSuites[m.selectedSuite].TestCases)-1 {
					m.selectedTest++
				}
			}
		}
	}
	return m, nil
}

func (m testResultsModel) View() string {
	if len(m.testReport.TestSuites) == 0 {
		return "Loading test results..."
	}

	switch m.viewMode {
	case "suites":
		return m.renderSuitesView()
	case "tests":
		return m.renderTestsView()
	case "detail":
		return m.renderDetailView()
	default:
		return m.renderSuitesView()
	}
}

func (m testResultsModel) renderSuitesView() string {
	var b strings.Builder
	
	b.WriteString(titleStyle.Render("📊 Test Results - Test Suites"))
	b.WriteString("\n\n")
	
	// Summary
	successRate := 0
	if m.testReport.TotalCount > 0 {
		successRate = (m.testReport.SuccessCount * 100) / m.testReport.TotalCount
	}
	
	summary := fmt.Sprintf("Total: %d | ✅ %d | ❌ %d | ⚠️ %d | ⏭️ %d | Success: %d%%",
		m.testReport.TotalCount, m.testReport.SuccessCount, m.testReport.FailedCount,
		m.testReport.ErrorCount, m.testReport.SkippedCount, successRate)
	b.WriteString(headerStyle.Render(summary))
	b.WriteString("\n\n")

	// Test suites table
	for i, suite := range m.testReport.TestSuites {
		if suite.TotalCount == 0 {
			continue
		}
		
		cursor := "  "
		if i == m.selectedSuite {
			cursor = "▶ "
		}
		
		suiteSuccessRate := 0
		if suite.TotalCount > 0 {
			suiteSuccessRate = (suite.SuccessCount * 100) / suite.TotalCount
		}
		
		line := fmt.Sprintf("%s%-30s | %3d | %3d | %3d | %3d | %3d | %3d%%",
			cursor, truncate(suite.Name, 30), suite.TotalCount, suite.SuccessCount,
			suite.FailedCount, suite.ErrorCount, suite.SkippedCount, suiteSuccessRate)
		
		if i == m.selectedSuite {
			line = lipgloss.NewStyle().Background(lipgloss.Color("#44475a")).Render(line)
		}
		
		b.WriteString(line)
		b.WriteString("\n")
	}
	
	b.WriteString("\n")
	b.WriteString("Controls: ↑/↓ navigate | Enter select | q quit")
	
	return b.String()
}

func (m testResultsModel) renderTestsView() string {
	if m.selectedSuite >= len(m.testReport.TestSuites) {
		return "Invalid suite selected"
	}
	
	suite := m.testReport.TestSuites[m.selectedSuite]
	var b strings.Builder
	
	b.WriteString(titleStyle.Render(fmt.Sprintf("📊 Test Cases - %s", suite.Name)))
	b.WriteString("\n\n")
	
	// Suite summary
	suiteSuccessRate := 0
	if suite.TotalCount > 0 {
		suiteSuccessRate = (suite.SuccessCount * 100) / suite.TotalCount
	}
	
	summary := fmt.Sprintf("Tests: %d | ✅ %d | ❌ %d | ⚠️ %d | ⏭️ %d | Success: %d%%",
		suite.TotalCount, suite.SuccessCount, suite.FailedCount,
		suite.ErrorCount, suite.SkippedCount, suiteSuccessRate)
	b.WriteString(headerStyle.Render(summary))
	b.WriteString("\n\n")

	// Test cases
	for i, testCase := range suite.TestCases {
		cursor := "  "
		if i == m.selectedTest {
			cursor = "▶ "
		}
		
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
		
		line := fmt.Sprintf("%s%s %-50s (%s)",
			cursor, icon, truncate(testCase.Name, 50), truncate(testCase.Classname, 30))
		
		if i == m.selectedTest {
			line = lipgloss.NewStyle().Background(lipgloss.Color("#44475a")).Render(line)
		}
		
		b.WriteString(line)
		b.WriteString("\n")
	}
	
	b.WriteString("\n")
	b.WriteString("Controls: ↑/↓ navigate | Enter details | Esc back | q quit")
	
	return b.String()
}

func (m testResultsModel) renderDetailView() string {
	if m.selectedSuite >= len(m.testReport.TestSuites) || 
	   m.selectedTest >= len(m.testReport.TestSuites[m.selectedSuite].TestCases) {
		return "Invalid test selected"
	}
	
	suite := m.testReport.TestSuites[m.selectedSuite]
	testCase := suite.TestCases[m.selectedTest]
	var b strings.Builder
	
	b.WriteString(titleStyle.Render("📊 Test Case Details"))
	b.WriteString("\n\n")
	
	b.WriteString(fmt.Sprintf("Name: %s\n", testCase.Name))
	b.WriteString(fmt.Sprintf("Class: %s\n", testCase.Classname))
	b.WriteString(fmt.Sprintf("Status: %s\n", testCase.Status))
	b.WriteString(fmt.Sprintf("Time: %.3fs\n", testCase.Time))
	
	if testCase.Message != "" {
		b.WriteString(fmt.Sprintf("\nMessage:\n%s\n", testCase.Message))
	}
	
	b.WriteString("\n")
	b.WriteString("Controls: Esc back | q quit")
	
	return b.String()
}

func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
