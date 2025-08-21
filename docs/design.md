# glab-tui Design Document

## 🎯 Project Vision

Create a terminal-based user interface for GitLab that brings the same level of productivity and elegance that k9s brings to Kubernetes. The goal is to provide DevOps engineers, developers, and platform teams with a powerful, keyboard-driven interface for managing GitLab CI/CD pipelines and projects.

## 🏗️ Architecture Overview

### Core Components

```
┌─────────────────────────────────────────────────────────────┐
│                        glab-tui                             │
├─────────────────────────────────────────────────────────────┤
│  UI Layer (Bubble Tea + Lipgloss)                          │
│  ├── Views (Pipeline, Job, MR, Artifact, Settings)         │
│  ├── Components (Tables, Logs, Forms, Modals)              │
│  └── Themes (Dark, Light, Custom)                          │
├─────────────────────────────────────────────────────────────┤
│  Business Logic Layer                                      │
│  ├── State Management (Global App State)                   │
│  ├── Event Handling (Keyboard, Timer, API)                 │
│  ├── Data Processing (Filtering, Sorting, Caching)         │
│  └── Action Controllers (Pipeline, Job, MR Operations)     │
├─────────────────────────────────────────────────────────────┤
│  Data Layer                                                │
│  ├── GitLab API Client (REST + GraphQL)                    │
│  ├── Authentication (Token, OAuth)                         │
│  ├── Caching (In-memory, Persistent)                       │
│  └── Configuration (YAML, Environment)                     │
├─────────────────────────────────────────────────────────────┤
│  External Dependencies                                      │
│  ├── GitLab API v4 (REST)                                  │
│  ├── GitLab GraphQL API                                    │
│  ├── Local Git Repository (Optional)                       │
│  └── File System (Config, Cache, Logs)                     │
└─────────────────────────────────────────────────────────────┘
```

## 🎨 User Interface Design

### Layout Structure

The interface follows a multi-pane layout similar to k9s:

```
┌─ Header ─────────────────────────────────────────────────────┐
│ Project: group/project | View: Pipelines | Status: ●●●○○    │
├─ Navigation ─────────────────────────────────────────────────┤
│ [P]ipelines [J]obs [M]Rs [A]rtifacts [V]ars [S]ettings [?] │
├─ Main Content ──────────────────────────────────────────────┤
│                                                             │
│  Primary View (Table/List)                                  │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │ ● Pipeline #123  running  feat/branch  (2m ago)        │ │
│  │ ✓ Pipeline #122  success  main        (1h ago)        │ │
│  │ ✗ Pipeline #121  failed   fix/bug     (2h ago)        │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                             │
├─ Detail Pane ───────────────────────────────────────────────┤
│                                                             │
│  Secondary View (Details/Logs)                             │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │ Jobs for Pipeline #123:                                 │ │
│  │ ● build        running   (30s)                          │ │
│  │ ○ test         pending                                  │ │
│  │ ○ deploy       pending                                  │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                             │
├─ Status Bar ────────────────────────────────────────────────┤
│ Last refresh: 2s ago | API calls: 15/1000 | Shortcuts: ?   │
└─────────────────────────────────────────────────────────────┘
```

### View Hierarchy

1. **Dashboard View** - Project overview and health metrics
2. **Pipeline View** - List of pipelines with status and metadata
3. **Job View** - Detailed job information and execution status
4. **Log View** - Real-time log streaming with syntax highlighting
5. **Merge Request View** - MR status and associated pipelines
6. **Artifact View** - Browse and download build artifacts
7. **Variable View** - Manage CI/CD variables and secrets
8. **Settings View** - Application configuration and preferences

## 🔄 State Management

### Application State Structure

```go
type AppState struct {
    // Current context
    CurrentProject   *Project
    CurrentView      ViewType
    CurrentSelection interface{}
    
    // Data cache
    Projects    map[string]*Project
    Pipelines   map[string][]*Pipeline
    Jobs        map[string][]*Job
    MergeRequests map[string][]*MergeRequest
    
    // UI state
    Filters     FilterState
    Sorting     SortState
    Pagination  PaginationState
    
    // Configuration
    Config      *Config
    Theme       *Theme
    
    // Runtime state
    RefreshTimer time.Time
    APILimits    *APILimitState
    Errors       []Error
}
```

### Data Flow

```
User Input → Event Handler → State Update → View Refresh → UI Render
     ↑                                                         ↓
API Response ← GitLab API Client ← Background Refresh ← Timer Event
```

## 🌐 GitLab API Integration

### API Strategy

**Primary**: REST API v4 for standard operations
**Secondary**: GraphQL for complex queries and bulk operations

### Key Endpoints

```go
// Pipeline Operations
GET    /projects/:id/pipelines
GET    /projects/:id/pipelines/:pipeline_id
POST   /projects/:id/pipeline
DELETE /projects/:id/pipelines/:pipeline_id

// Job Operations  
GET    /projects/:id/pipelines/:pipeline_id/jobs
GET    /projects/:id/jobs/:job_id
POST   /projects/:id/jobs/:job_id/retry
POST   /projects/:id/jobs/:job_id/cancel
GET    /projects/:id/jobs/:job_id/trace

// Merge Request Operations
GET    /projects/:id/merge_requests
GET    /projects/:id/merge_requests/:merge_request_iid/pipelines

// Artifact Operations
GET    /projects/:id/jobs/:job_id/artifacts
GET    /projects/:id/jobs/:job_id/artifacts/:artifact_path
```

### Caching Strategy

- **Hot Data**: Pipelines, jobs (5-30 second cache)
- **Warm Data**: Project metadata, variables (5-minute cache)  
- **Cold Data**: Historical data, artifacts (1-hour cache)
- **Persistent Cache**: User preferences, project bookmarks

### Rate Limiting

- Respect GitLab API rate limits (300 requests/minute for gitlab.com)
- Implement exponential backoff for failed requests
- Batch requests where possible using GraphQL
- Show API usage in status bar

## ⌨️ Keyboard Interface Design

### Navigation Philosophy

Follow vim-style navigation with context-aware shortcuts:

- **Consistent**: Same keys do similar things across views
- **Discoverable**: Help system shows available shortcuts
- **Efficient**: Common actions have single-key shortcuts
- **Contextual**: Different views expose relevant actions

### Key Mapping Strategy

```go
type KeyMap struct {
    // Global shortcuts (available everywhere)
    Quit        key.Binding // q, Ctrl+C
    Help        key.Binding // ?
    Refresh     key.Binding // r
    Search      key.Binding // /
    
    // Navigation (vim-style)
    Up          key.Binding // k, ↑
    Down        key.Binding // j, ↓
    Left        key.Binding // h, ←
    Right       key.Binding // l, →
    PageUp      key.Binding // Ctrl+u
    PageDown    key.Binding // Ctrl+d
    Top         key.Binding // g
    Bottom      key.Binding // G
    
    // View switching
    Pipelines   key.Binding // p
    Jobs        key.Binding // j
    MRs         key.Binding // m
    Artifacts   key.Binding // a
    Variables   key.Binding // v
    Settings    key.Binding // s
    
    // Actions (context-dependent)
    Select      key.Binding // Enter, Space
    Cancel      key.Binding // c
    Retry       key.Binding // r (in job context)
    Delete      key.Binding // d
    Download    key.Binding // d (in artifact context)
    Follow      key.Binding // f (in log context)
}
```

## 🎨 Visual Design System

### Color Scheme

**Status Colors**:
- 🟢 Success: Green (#28a745)
- 🔴 Failed: Red (#dc3545)  
- 🟡 Warning: Yellow (#ffc107)
- 🔵 Running: Blue (#007bff)
- ⚪ Pending: Gray (#6c757d)
- 🟠 Manual: Orange (#fd7e14)

**UI Colors**:
- Primary: Blue (#007bff)
- Secondary: Gray (#6c757d)
- Background: Dark (#1a1a1a) / Light (#ffffff)
- Text: Light Gray (#f8f9fa) / Dark Gray (#212529)
- Border: Medium Gray (#495057)

### Typography

- **Headers**: Bold, larger font
- **Body**: Regular weight, readable size
- **Code**: Monospace font for logs and technical data
- **Status**: Color-coded with symbols (●○✓✗)

### Layout Principles

- **Consistent spacing**: 1-2 character padding
- **Clear hierarchy**: Headers, sections, content
- **Responsive**: Adapt to terminal size
- **Accessible**: High contrast, clear indicators

## 🔧 Configuration System

### Configuration File Structure

```yaml
# ~/.config/glab-tui/config.yaml
gitlab:
  # GitLab instance configuration
  url: "https://gitlab.com"
  token: "glpat-xxxxxxxxxxxxxxxxxxxx"
  api_version: "v4"
  timeout: 30s
  
ui:
  # Interface preferences
  theme: "dark"              # dark, light, auto
  refresh_interval: 5s       # Auto-refresh frequency
  vim_mode: true            # Enable vim-style navigation
  show_help: true           # Show help hints
  page_size: 20             # Items per page
  
projects:
  # Project bookmarks
  - name: "Frontend Apps"
    path: "group/project/frontend-apps"
    default: true
  - name: "Platform Components"  
    path: "group/platform/standard-components"
    
filters:
  # Default filtering options
  default_branch_only: false
  show_system_pipelines: false
  max_pipelines: 50
  max_jobs: 100
  
keybindings:
  # Custom key mappings
  quit: ["q", "ctrl+c"]
  refresh: ["r", "F5"]
  help: ["?", "F1"]
  
cache:
  # Caching configuration
  enabled: true
  ttl: 300s                 # 5 minutes default TTL
  max_size: 100MB
  persist: true
```

### Environment Variables

```bash
# Authentication
GITLAB_TOKEN=glpat-xxxxxxxxxxxxxxxxxxxx
GITLAB_URL=https://gitlab.example.com

# Behavior
GLAB_TUI_REFRESH_INTERVAL=5s
GLAB_TUI_THEME=dark
GLAB_TUI_VIM_MODE=true

# Debug
GLAB_TUI_DEBUG=true
GLAB_TUI_LOG_LEVEL=info
```

## 🚀 Performance Considerations

### Optimization Strategies

1. **Lazy Loading**: Load data only when needed
2. **Virtual Scrolling**: Handle large lists efficiently  
3. **Debounced Refresh**: Avoid excessive API calls
4. **Background Updates**: Refresh data without blocking UI
5. **Efficient Rendering**: Only redraw changed components

### Memory Management

- Implement LRU cache for API responses
- Limit in-memory data retention
- Clean up unused goroutines and timers
- Monitor memory usage and provide diagnostics

### Network Efficiency

- Batch API requests where possible
- Use HTTP/2 connection pooling
- Implement request deduplication
- Cache static data (project metadata, user info)

## 🧪 Testing Strategy

### Unit Tests
- Core business logic
- API client functionality
- State management
- Configuration parsing

### Integration Tests  
- GitLab API interactions
- End-to-end user workflows
- Configuration loading
- Error handling scenarios

### Manual Testing
- Cross-platform compatibility
- Different terminal sizes
- Various GitLab instances
- Performance under load

## 📦 Distribution Strategy

### Build Targets
- Linux (amd64, arm64)
- macOS (amd64, arm64)  
- Windows (amd64)

### Distribution Channels
- **GitHub Releases**: Binary downloads
- **Homebrew**: macOS/Linux package manager
- **Go Install**: Direct from source
- **Docker**: Containerized version
- **Package Managers**: apt, yum, pacman (future)

### Versioning
- Semantic versioning (v1.2.3)
- Release notes with changelog
- Migration guides for breaking changes

## 🔮 Future Enhancements

### Phase 2 Features
- **Multi-instance support**: Connect to multiple GitLab instances
- **Team dashboards**: Shared views for team collaboration
- **Custom queries**: Save and reuse complex filters
- **Notification system**: Desktop notifications for events

### Phase 3 Features  
- **Plugin architecture**: Extensible functionality
- **Scripting support**: Automate common workflows
- **Integration APIs**: Connect with other tools
- **Advanced analytics**: Pipeline performance insights

### Phase 4 Features
- **Web interface**: Optional web UI for sharing
- **Mobile companion**: Basic mobile app
- **AI assistance**: Intelligent troubleshooting
- **Enterprise features**: SSO, audit logs, compliance

## 🤝 Development Guidelines

### Code Organization
```
glab-tui/
├── cmd/                    # CLI entry points
├── internal/
│   ├── ui/                 # UI components and views
│   ├── api/                # GitLab API client
│   ├── config/             # Configuration management
│   ├── cache/              # Caching layer
│   └── state/              # State management
├── pkg/                    # Public packages
├── docs/                   # Documentation
├── examples/               # Example configurations
└── scripts/                # Build and deployment scripts
```

### Development Workflow
1. **Feature branches**: One feature per branch
2. **Code review**: All changes reviewed before merge
3. **Testing**: Comprehensive test coverage
4. **Documentation**: Keep docs updated with changes
5. **Releases**: Regular, predictable release schedule

### Contributing Guidelines
- Follow Go best practices and conventions
- Write tests for new functionality
- Update documentation for user-facing changes
- Use conventional commit messages
- Respect existing code style and patterns

---

This design document serves as the foundation for building glab-tui. It should evolve as we learn more about user needs and technical constraints during development.
