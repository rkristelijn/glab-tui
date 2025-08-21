# glab-tui 🚀

A k9s-inspired Terminal User Interface (TUI) for GitLab CI/CD pipelines and project management.

## 🎯 Vision

Bring the power and elegance of k9s to GitLab! Monitor pipelines, jobs, merge requests, and more in a beautiful, keyboard-driven terminal interface.

## ✨ Features (Planned)

### 🔄 Pipeline Management
- **Real-time pipeline monitoring** - Watch pipelines as they run
- **Multi-project support** - Switch between GitLab projects seamlessly  
- **Pipeline history** - Browse recent and historical pipeline runs
- **Job details** - Drill down into individual job execution
- **Live log streaming** - Follow job logs in real-time with syntax highlighting

### 🔍 Advanced Views
- **Dashboard overview** - Project health at a glance
- **Merge request tracking** - Monitor MR pipelines and status
- **Resource usage** - Runner utilization and performance metrics
- **Artifact browser** - Download and inspect build artifacts
- **Variable inspector** - View and manage CI/CD variables

### ⌨️ Keyboard-Driven Interface
- **Vim-style navigation** - `hjkl` movement, `/` for search
- **Context-aware hotkeys** - Different actions per view
- **Quick actions** - Retry jobs, cancel pipelines, trigger builds
- **Multi-selection** - Bulk operations on jobs/pipelines

## 🎨 Interface Preview

```
┌─ GitLab TUI - theapsgroup/agility/frontend-apps ─────────────────────────────┐
│ [P]ipelines [J]obs [M]Rs [A]rtifacts [V]ariables [S]ettings            [?] │
├───────────────────────────────────────────────────────────────────────────────┤
│ Pipelines                                                     ↻ Auto-refresh │
│ ● #1996879423  running   feat/zap-c3        (2m ago)  [●●●○○○○○] 3/8 jobs   │
│ ● #1996867272  running   refs/merge-req/406 (8m ago)  [●●●●●○○○] 5/8 jobs   │
│ ✓ #1996733511  success   fix/supplier-...   (1h ago)  [●●●●●●●●] 8/8 jobs   │
│ ✗ #1996723026  failed    fix/supplier-...   (1h ago)  [●●●●●✗○○] failed     │
│ ○ #1996719037  success   main               (1h ago)  [●●●●●●●●] 8/8 jobs   │
├───────────────────────────────────────────────────────────────────────────────┤
│ Jobs (Pipeline #1996879423)                                                  │
│ ✓ npm-preparation        success   (45s)   Node.js dependencies installed   │
│ ● nx-mono-repo-affected  running   (12s)   Analyzing affected projects...   │
│ ○ cloudflare-deploy      pending           Waiting for dependencies         │
│ ○ zap-security-scan      pending           Security scan queued             │
│ ○ cypress-e2e            pending           E2E tests queued                 │
├───────────────────────────────────────────────────────────────────────────────┤
│ Logs (nx-mono-repo-affected)                                        [Follow] │
│ 2025-08-21T16:30:15Z │ 🔍 Analyzing project dependencies...                │
│ 2025-08-21T16:30:16Z │ ✅ Found 3 affected projects: internal-demo-app...  │
│ 2025-08-21T16:30:17Z │ 🏗️  Building dependency graph...                    │
│ 2025-08-21T16:30:18Z │ 📦 Preparing build artifacts...                     │
└───────────────────────────────────────────────────────────────────────────────┘
```

## 🛠️ Technology Stack

- **Language**: Go (for performance and single-binary distribution)
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **GitLab API**: REST API v4 + GraphQL for advanced queries
- **Configuration**: YAML-based config with GitLab token management

## 🚀 Installation (Future)

```bash
# Homebrew (planned)
brew install glab-tui

# Go install
go install github.com/username/glab-tui@latest

# Binary releases
curl -L https://github.com/username/glab-tui/releases/latest/download/glab-tui-linux-amd64 -o glab-tui
```

## ⌨️ Keybindings

### Global
- `q` / `Ctrl+C` - Quit
- `?` - Help/keybindings
- `r` - Refresh current view
- `/` - Search/filter
- `Esc` - Clear search/go back

### Navigation  
- `h/j/k/l` - Navigate (vim-style)
- `g/G` - Go to top/bottom
- `Ctrl+u/d` - Page up/down
- `Tab` - Switch between panes

### Pipeline View
- `Enter` - View pipeline jobs
- `Space` - Toggle pipeline selection
- `c` - Cancel selected pipeline(s)
- `t` - Trigger new pipeline
- `d` - Delete pipeline

### Job View
- `Enter` - View job logs
- `r` - Retry job
- `c` - Cancel job
- `a` - Download artifacts
- `f` - Follow logs (live tail)

## 🔧 Configuration

```yaml
# ~/.config/glab-tui/config.yaml
gitlab:
  url: "https://gitlab.com"
  token: "glpat-xxxxxxxxxxxxxxxxxxxx"
  
ui:
  refresh_interval: 5s
  theme: "dark"
  vim_mode: true
  
projects:
  - "theapsgroup/agility/frontend-apps"
  - "theapsgroup/platform/standard-components"
  
filters:
  default_branch_only: false
  show_system_pipelines: false
  max_pipelines: 50
```

## 🎯 Use Cases

### DevOps Engineers
- Monitor multiple project pipelines simultaneously
- Quick troubleshooting of failed jobs
- Bulk operations on pipeline management

### Developers  
- Track MR pipeline status during code review
- Debug CI/CD issues with live log streaming
- Quick access to build artifacts and test results

### Platform Teams
- Monitor runner utilization across projects
- Manage CI/CD variables and configurations
- Analyze pipeline performance and bottlenecks

## 🤝 Contributing

This project is in the conceptual phase! We're looking for:

- **Go developers** - Core TUI development
- **GitLab API experts** - Efficient data fetching strategies  
- **UX designers** - Interface design and user workflows
- **DevOps practitioners** - Real-world use case validation

## 📋 Roadmap

### Phase 1: MVP
- [ ] Basic pipeline listing and status
- [ ] Job details and log viewing
- [ ] Single project support
- [ ] Core navigation and keybindings

### Phase 2: Enhanced Features
- [ ] Multi-project support
- [ ] Live log streaming
- [ ] Pipeline actions (retry, cancel)
- [ ] Search and filtering

### Phase 3: Advanced Features
- [ ] Merge request integration
- [ ] Artifact management
- [ ] Variable management
- [ ] Performance metrics

### Phase 4: Power User Features
- [ ] Custom dashboards
- [ ] Notification system
- [ ] Plugin architecture
- [ ] Team collaboration features

## 🎨 Inspiration

- **k9s** - The gold standard for Kubernetes TUI
- **lazygit** - Excellent Git TUI with intuitive interface
- **htop** - Clean system monitoring interface
- **GitLab Web UI** - Feature completeness reference

## 📄 License

MIT License - See [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- The k9s team for showing how powerful TUI applications can be
- GitLab for providing comprehensive APIs
- The Charm team for excellent Go TUI libraries

---

**Status**: 🚧 Conceptual phase - Looking for contributors!

*Born from the frustration of constantly refreshing GitLab pipeline pages* 😅
