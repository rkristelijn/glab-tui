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
- **🔥 Real-time log streaming** - Stream job logs as they happen with auto-refresh ✅ **CONFIRMED WORKING**
- **🔗 Child pipeline support** - Navigate through mono-repo child pipelines seamlessly
- **🎨 Beautiful interface** - Color-coded status indicators and visual formatting
- **📊 Visual overview** - Multiple pipelines at a glance in one screen
- **⌨️ Vim-style navigation** - Complete keyboard shortcuts (hjkl, gg, G, Ctrl+U/D)
- **🔍 Advanced log search** - Search through logs with highlighting and navigation
- **🎯 Interactive dashboard** - Navigate through pipelines, jobs, and child pipelines
- **🔍 Better UX** - No more repetitive CLI commands for status checks
- **🚀 Easy setup** - Uses your existing glab authentication

## 🎮 Usage

### **Interactive TUI (Default)**
```bash
./glab-tui                  # Start the beautiful TUI
```

**🎯 TUI Quick Start:**
- **First time?** Press `r` to refresh and load pipelines
- **Navigate:** Use ↑/↓ arrows or `j`/`k` (vim-style)
- **Drill down:** Press `Enter` to go: Pipelines → Jobs → Logs
- **Child pipelines:** Press `Enter` on 🔗 entries to navigate to child pipeline jobs
- **Real-time logs:** Press `l` on any job for live streaming
- **Search logs:** Press `/` to search, `n` for next match
- **Go back:** Press `Esc` to return to previous view
- **Quit:** Press `q` or `Ctrl+C`

**🔥 Pro tip:** Navigate to a running job and press `l` for real-time log streaming!

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
| `g` / `G` | Go to first/last item |
| `Ctrl+U/D` | Page up/down |
| `Enter` | Drill down (Pipeline → Jobs → Logs) |
| `Esc` | Go back |
| `r` | Refresh |
| `/` | Search (in logs) |
| `n` | Next search match |
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
| **Navigation** | ❌ Type commands | ✅ Vim-style shortcuts |
| **Real-time Updates** | ❌ Manual refresh | ✅ Live monitoring |
| **Child Pipelines** | ❌ Not visible | ✅ Full navigation support |
| **Log Search** | ❌ Basic grep | ✅ Interactive search with highlighting |

**Result: Better workflow + visual experience** 🏆

## 🎨 Interface Preview

```
┌─ GitLab TUI - agility/frontend-apps | 10 pipelines (3 running) ─────────┐
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
│ 🔗 ● portal #1997363518    running   Child pipeline                    │
│ 🔗 ● internal-demo #1997363434  running   Child pipeline               │
│ ○ deploy         pending           Waiting for tests                   │
└─────────────────────────────────────────────────────────────────────────┘
```

## 🔗 Child Pipeline Support

Perfect for **NX mono-repo** workflows:

```
🔧 Jobs (Pipeline #1997353757)
📊 12 total | ✅ 2 success | 🔄 6 running | ❌ 0 failed

  ✓ nx-mono-repo-affected     success      mono-repo
  ● 🔗 ● internal-demo-application #1997363434    running      child-pipeline
  ● 🔗 ● portal #1997363518                       running      child-pipeline
  ● 🔗 ● request-for-quote #1997363593            running      child-pipeline
  ● 🔗 ● social-networker #1997363731             running      child-pipeline
```

- **Navigate to child pipelines** with Enter key
- **See real app names** extracted from job logs
- **Full drill-down support** for mono-repo workflows

## 💬 What Users Say

> *"Finally, a GitLab interface that doesn't make me want to cry."*  
> — **Frontend Developer**

> *"Much better overview than running glab commands repeatedly."*  
> — **DevOps Engineer**

> *"It's like k9s but for GitLab. Love the visual dashboard and child pipeline support."*  
> — **Platform Engineer**

> *"The real-time log streaming and search is a game changer for debugging."*  
> — **Senior Developer**

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
