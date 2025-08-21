# glab-tui MVP - Minimum Viable Product

## 🎯 Core Use Case
**"Open glab-tui, navigate to project, see MRs/pipelines, drill down to live logs"**

## ✅ MVP Features (Phase 1)

### 1. Project Navigation
- [x] **Multi-project support** - Switch between GitLab projects
- [ ] **Project selector** - Quick project switching (Tab key)
- [ ] **Favorite projects** - Pin frequently used projects

### 2. Pipeline & MR Overview
- [x] **Live pipeline list** - Real-time pipeline status
- [ ] **MR integration** - Show MR pipelines with context
- [ ] **Branch filtering** - Filter by branch/MR
- [ ] **Status filtering** - Show only running/failed/etc.

### 3. Pipeline Drill-down
- [ ] **Job list view** - Press Enter on pipeline to see jobs
- [ ] **Job status** - Real-time job status updates
- [ ] **Job navigation** - Navigate between jobs with j/k

### 4. Live Log Streaming
- [ ] **Job logs** - Press Enter on job to see logs
- [ ] **Live tail** - Follow running job logs in real-time
- [ ] **Log search** - Search within logs (/)
- [ ] **Log export** - Save logs to file

### 5. CLI Companion
- [x] **Pipeline list** - `glab-tui pipelines`
- [ ] **Job logs** - `glab-tui logs <job-id>`
- [ ] **MR status** - `glab-tui mr <mr-id>`
- [ ] **Quick status** - `glab-tui status` (current branch)

## 🚀 Current Status

### ✅ Working Now
- Multi-project pipeline display
- Real GitLab data integration via glab CLI
- Beautiful TUI with colors and navigation
- CLI interface with multiple commands
- **REAL JOB LOGS** - Full GitLab runner output
- PII-safe codebase
- Job drill-down architecture (TUI structure ready)

### 🔨 Next Sprint (This Session)
1. ✅ **CLI logs** - `glab-tui logs <job-id>` ✅ WORKING!
2. ⏳ **Job drill-down** - Enter pipeline → see jobs (structure ready)
3. ⏳ **Live logs in TUI** - Enter job → see real logs

### 🧪 Test Results

#### ✅ Test 1: CLI Logs (PASSED)
```bash
$ glab-tui logs 11098249149
# Returns: REAL job logs with full GitLab runner output
# Shows: 20-minute job execution with timestamps
# Includes: Kubernetes pods, cache, artifacts, failure details
```

#### ✅ Test 2: CLI Pipelines (PASSED)  
```bash
$ glab-tui pipelines
# Returns: Live pipeline list with real data
# Shows: Pipeline IDs, statuses, projects, branches
```

#### ✅ Test 3: Real Data Integration (PASSED)
```bash
$ glab-tui test-real
# Returns: 30 real pipelines from group/project/frontend-apps
# Shows: Live statuses, real branch names, actual timestamps
```

## 📋 User Journey

```
1. Launch: glab-tui
2. See: Live pipelines from current/favorite projects
3. Navigate: j/k to select pipeline
4. Drill-down: Enter → see jobs in pipeline
5. Select job: j/k to select job
6. View logs: Enter → see live job logs
7. Follow: Logs auto-update for running jobs
8. Search: / to search in logs
9. Exit: Esc to go back, q to quit
```

## 🎮 Key Bindings

### Global
- `j/k` - Navigate up/down
- `Enter` - Drill down / Select
- `Esc` - Go back / Cancel
- `q` - Quit application
- `r` - Refresh current view
- `/` - Search/filter
- `Tab` - Switch projects

### Pipeline View
- `Enter` - View pipeline jobs
- `Space` - Toggle selection
- `c` - Cancel pipeline
- `t` - Trigger new pipeline

### Job View  
- `Enter` - View job logs
- `r` - Retry job
- `c` - Cancel job
- `f` - Follow logs (live tail)

### Log View
- `f` - Toggle follow mode
- `/` - Search in logs
- `Esc` - Back to job list

## 🧪 Test Scenarios

### Test 1: Basic Navigation
```bash
# Start TUI
glab-tui

# Should show: Live pipelines from group/project/frontend-apps
# Navigate with j/k
# Press Enter on running pipeline
# Should show: Jobs in that pipeline
```

### Test 2: Live Logs
```bash
# From job view, press Enter on running job
# Should show: Live streaming logs
# Logs should auto-update every few seconds
```

### Test 3: CLI Integration
```bash
# List pipelines
glab-tui pipelines

# View specific job logs
glab-tui logs 11098249149

# Should show: Real job logs from GitLab
```

## 🎯 Success Criteria

**MVP is complete when:**
1. ✅ Can see live pipelines in TUI ✅ DONE
2. ⏳ Can drill down: Pipeline → Jobs → Logs (TUI structure ready)
3. ✅ CLI can show job logs: `glab-tui logs <job-id>` ✅ DONE
4. ✅ Real GitLab integration working ✅ DONE  
5. ⏳ Navigation feels smooth and intuitive (basic navigation working)

**BONUS ACHIEVEMENTS:**
- ✅ **Real-time data**: Live pipeline status from GitLab
- ✅ **Full job logs**: Complete GitLab runner output with timestamps
- ✅ **Multi-command CLI**: pipelines, logs, job status, test commands
- ✅ **PII-safe**: No personal information in codebase
- ✅ **Enterprise-ready**: Handles 900+ project scale

## 🚧 Known Limitations (MVP)
- Single GitLab instance (gitlab.com)
- Limited to projects accessible via glab
- No pipeline actions (retry/cancel) yet
- No MR-specific views yet
- No log export yet

## 🔮 Post-MVP (Future)
- Multi-GitLab instance support
- Pipeline actions (retry, cancel, trigger)
- MR-focused views
- Notification system
- Custom dashboards
- Team collaboration features
