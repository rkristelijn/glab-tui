package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	GitLab GitLabConfig
	UI     UIConfig
}

type GitLabConfig struct {
	URL            string
	Token          string
	ProjectID      int    // Single project ID for testing
	GroupID        int
	GroupPath      string
	ProjectIDs     []int
	ProjectPattern string
	MaxProjects    int
	ShowArchived   bool
	MinActivityDays int
}

type UIConfig struct {
	RefreshInterval        time.Duration
	MaxPipelinesPerProject int
}

func Load() (*Config, error) {
	// Load .env file if it exists
	loadEnvFile()

	groupID, _ := strconv.Atoi(getEnv("GITLAB_GROUP_ID", "0"))
	projectID, _ := strconv.Atoi(getEnv("GITLAB_PROJECT_ID", "0"))
	maxProjects, _ := strconv.Atoi(getEnv("MAX_PROJECTS", "50"))
	minActivityDays, _ := strconv.Atoi(getEnv("MIN_ACTIVITY_DAYS", "30"))
	showArchived, _ := strconv.ParseBool(getEnv("SHOW_ARCHIVED", "false"))
	refreshInterval, _ := time.ParseDuration(getEnv("REFRESH_INTERVAL", "5s"))
	maxPipelinesPerProject, _ := strconv.Atoi(getEnv("MAX_PIPELINES_PER_PROJECT", "10"))

	// Parse project IDs if provided
	var projectIDs []int
	if projectIDsStr := getEnv("GITLAB_PROJECT_IDS", ""); projectIDsStr != "" {
		for _, idStr := range strings.Split(projectIDsStr, ",") {
			if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				projectIDs = append(projectIDs, id)
			}
		}
	}

	// Try to get token from .env first, then from glab config
	token := getEnv("GITLAB_TOKEN", "")
	if token == "" || token == "your-token-here" {
		if glabToken, err := LoadGlabToken(); err == nil && glabToken != "" {
			token = glabToken
		}
	}

	return &Config{
		GitLab: GitLabConfig{
			URL:             getEnv("GITLAB_URL", "https://gitlab.com"),
			Token:           token,
			ProjectID:       projectID,
			GroupID:         groupID,
			GroupPath:       getEnv("GITLAB_GROUP_PATH", ""),
			ProjectIDs:      projectIDs,
			ProjectPattern:  getEnv("GITLAB_PROJECT_PATTERN", ""),
			MaxProjects:     maxProjects,
			ShowArchived:    showArchived,
			MinActivityDays: minActivityDays,
		},
		UI: UIConfig{
			RefreshInterval:        refreshInterval,
			MaxPipelinesPerProject: maxPipelinesPerProject,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func loadEnvFile() {
	// Simple .env file loader
	if file, err := os.Open(".env"); err == nil {
		defer file.Close()
		
		// Read line by line and set environment variables
		// This is a basic implementation - you could use a library like godotenv
		buf := make([]byte, 1024)
		if n, err := file.Read(buf); err == nil {
			content := string(buf[:n])
			lines := splitLines(content)
			
			for _, line := range lines {
				if len(line) > 0 && line[0] != '#' {
					if parts := splitKeyValue(line); len(parts) == 2 {
						os.Setenv(parts[0], parts[1])
					}
				}
			}
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	
	for i, c := range s {
		if c == '\n' {
			if start < i {
				lines = append(lines, s[start:i])
			}
			start = i + 1
		}
	}
	
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	
	return lines
}

func splitKeyValue(line string) []string {
	for i, c := range line {
		if c == '=' {
			key := line[:i]
			value := line[i+1:]
			// Remove quotes if present
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
			return []string{key, value}
		}
	}
	return nil
}
