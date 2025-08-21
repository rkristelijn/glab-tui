package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/rkristelijn/glab-tui/internal/core"
	"github.com/rkristelijn/glab-tui/internal/config"
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
	pipelines := core.GetMockPipelines()
	
	fmt.Println("GitLab Pipelines (Multi-Project):")
	fmt.Println("ID          Status    Project         Ref                   Jobs")
	fmt.Println("─────────────────────────────────────────────────────────────────────")
	
	for _, p := range pipelines {
		status := getStatusIcon(p.Status)
		fmt.Printf("%-10d  %s %-8s %-15s %-20s %s\n", 
			p.ID, status, p.Status, p.ProjectName, p.Ref, p.Jobs)
	}
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

	fmt.Printf("Checking job %d in group/project/frontend-apps...\n", jobID)
	
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
	
	wrapper := gitlab.NewGlabWrapper("group/project/frontend-apps")
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
    glab-tui [COMMAND]          Run CLI command

COMMANDS:
    pipelines, p               List pipelines
    job, j <job-id>           Check specific job status
    logs, l <job-id>          Show job logs
    test-real                 Test GitLab API connection
    help, h                   Show this help
    version, v                Show version

EXAMPLES:
    glab-tui                          # Start TUI
    glab-tui pipelines                # List pipelines in CLI
    glab-tui job 11098249149         # Check specific job
    glab-tui logs 11098249149        # Show job logs
    glab-tui test-real               # Test GitLab connection
    glab-tui help                    # Show help`)
}
