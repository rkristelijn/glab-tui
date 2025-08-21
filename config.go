package main

import (
	"os"
	"path/filepath"
)

type Config struct {
	GitLab GitLabConfig `yaml:"gitlab"`
	UI     UIConfig     `yaml:"ui"`
}

type GitLabConfig struct {
	URL   string `yaml:"url"`
	Token string `yaml:"token"`
}

type UIConfig struct {
	RefreshInterval string `yaml:"refresh_interval"`
	Theme           string `yaml:"theme"`
	VimMode         bool   `yaml:"vim_mode"`
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "glab-tui", "config.yaml")
}

func getDefaultConfig() Config {
	return Config{
		GitLab: GitLabConfig{
			URL:   "https://gitlab.com",
			Token: os.Getenv("GITLAB_TOKEN"),
		},
		UI: UIConfig{
			RefreshInterval: "5s",
			Theme:           "dark",
			VimMode:         true,
		},
	}
}
