# glab-tui 🚀

A beautiful, interactive Terminal User Interface (TUI) for GitLab CI/CD pipelines.

**Stop typing repetitive `glab` commands - see everything at once in a visual dashboard!** ✨

## ⚡ Quick Start

```bash
# The easy way
npx glab-tui

# Or download and run
curl -L https://github.com/rkristelijn/glab-tui/releases/latest/download/glab-tui-linux-amd64 -o glab-tui
chmod +x glab-tui
./glab-tui
```

## 🎯 Why glab-tui?

### **Before (CLI):**
```bash
glab pipeline list          # Check pipelines
glab pipeline ci view 123   # Check specific pipeline  
glab api jobs/456/trace     # Get job logs
# Repeat every 30 seconds... 😴
```

### **After (TUI):**
```bash
glab-tui                    # See everything at once! 🎉
```

## ✨ Features

- **🔄 Real-time monitoring** - Live pipeline status updates
- **🔥 Real-time log streaming** - Stream job logs as they happen with `--follow` ✅ **CONFIRMED WORKING**
- **🎨 Beautiful interface** - Color-coded status indicators and visual formatting
- **📊 Visual overview** - Multiple pipelines at a glance in one screen
- **⌨️ Keyboard driven** - Vim-style navigation (hjkl) for quick browsing
- **🎯 Interactive dashboard** - Navigate through pipelines, jobs, and details
- **🔍 Better UX** - No more repetitive CLI commands for status checks
- **🚀 Easy setup** - Uses your existing glab authentication

## 🎮 Usage

### **Interactive TUI (Default)**
```bash
./glab-tui                  # Start the beautiful TUI
```

### **CLI Commands**
```bash
./glab-tui pipelines        # List pipelines
./glab-tui job 12345        # Check job status
./glab-tui logs 12345       # View job logs
./glab-tui logs --follow 12345  # 🔥 Stream logs in real-time
./glab-tui help             # Show help
```

## ⌨️ Keyboard Controls

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit |
| `j/k` or `↓/↑` | Navigate up/down |
| `Enter` | Drill down (Pipeline → Jobs → Logs) |
| `Esc` | Go back |
| `r` | Refresh |
| `?` | Help |

## 🚀 Installation

### **NPX (Recommended)**
```bash
npx glab-tui                # Zero install, just works!
```

### **Download Binary**
```bash
# Linux/macOS
curl -L https://github.com/rkristelijn/glab-tui/releases/latest/download/glab-tui-linux-amd64 -o glab-tui
chmod +x glab-tui

# Or with Go
go install github.com/rkristelijn/glab-tui@latest
```

## 🔧 Requirements

- **GitLab CLI (`glab`)** - Install from [cli.gitlab.com](https://gitlab.com/gitlab-org/cli)
- **Authentication** - Run `glab auth login` first
- **Git repository** - Run from inside a GitLab project

## 📊 User Experience

| Feature | glab CLI | glab-tui |
|---------|----------|----------|
| **Visual Overview** | ❌ Plain text | ✅ Color-coded dashboard |
| **Multi-pipeline View** | ❌ One at a time | ✅ All at once |
| **Navigation** | ❌ Type commands | ✅ Keyboard shortcuts |
| **Real-time Updates** | ❌ Manual refresh | ✅ Live monitoring |

**Result: Better workflow + visual experience** 🏆

## 🎨 Interface Preview

```
┌─ GitLab TUI - my-awesome-project ──────────────────────────────────────┐
│ [P]ipelines [J]obs [L]ogs                                         [?] │
├─────────────────────────────────────────────────────────────────────────┤
│ Pipelines                                               ↻ Auto-refresh │
│ ● #1234567  running   feat/new-feature    (2m ago)  [●●●○○○] 3/6 jobs │
│ ✓ #1234566  success   main                (1h ago)  [●●●●●●] 6/6 jobs │
│ ✗ #1234565  failed    fix/bug-123         (2h ago)  [●●●✗○○] failed   │
├─────────────────────────────────────────────────────────────────────────┤
│ Jobs (Pipeline #1234567)                                               │
│ ✓ build          success   (45s)   Dependencies installed             │
│ ● test           running   (12s)   Running test suite...              │
│ ○ deploy         pending           Waiting for tests                   │
└─────────────────────────────────────────────────────────────────────────┘
```

## 💬 What Users Say

> *"Finally, a GitLab interface that doesn't make me want to cry."*  
> — **Frontend Developer**

> *"Much better overview than running glab commands repeatedly."*  
> — **DevOps Engineer**

> *"It's like k9s but for GitLab. Love the visual dashboard."*  
> — **Platform Engineer**

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md).

## 📄 License

MIT License - See [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- Inspired by [k9s](https://github.com/derailed/k9s) - The gold standard for Kubernetes TUI
- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Excellent Go TUI framework
- Powered by [GitLab CLI](https://gitlab.com/gitlab-org/cli) - Official GitLab command line tool

---

**Transform your GitLab workflow today!** 🚀
