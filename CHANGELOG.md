# Changelog

All notable changes to the GPU Tracker project.

## [1.1.0] - 2026-01-31

### Added

#### Command-Line Interface
- **Sampling interval configuration** (`-interval` flag): Configure custom sampling rates from 1 second to any duration
- **Database path customization** (`-db` flag): Specify custom location for SQLite database
- **Version information** (`-version` flag): Display application version

#### Operating Modes
- **One-shot mode** (`-once` flag): Sample GPU state once and exit without starting TUI
- **Continuous monitoring mode** (`-continuous` flag): Background sampling with automatic database saves
- **User listing mode** (`-list-users` flag): Quick summary of GPU users and their memory consumption

#### Export Functionality
- **JSON export** (`-export json`): Export complete snapshot data in JSON format
- **CSV export** (`-export csv`): Export data in CSV format for spreadsheet analysis
- **Output file specification** (`-output` flag): Direct export output to file or stdout
- Export includes full GPU metrics, process information, and timestamps

#### Alert System
- **Temperature threshold alerts** (`-max-temp` flag): Configurable temperature warnings (default: 90°C)
- **Memory threshold alerts** (`-max-mem` flag): Configurable memory utilization warnings (default: 95%)
- Visual alert indicators in TUI with red highlights
- Alert messages in stderr for continuous and one-shot modes
- Emoji indicators (⚠️) for better visibility

#### TUI Enhancements
- **User filtering** (key: `f`): Cycle through users to focus on specific user activity
- **GPU filtering** (key: `g`): Filter view by specific GPU in multi-GPU systems
- **Memory sorting** (key: `m`): Toggle sorting of processes by memory usage
- **Clear filters** (key: `c`): Reset all active filters to full view
- **Active filter indicators**: Visual display of currently applied filters in header
- **Highlighted filtered items**: Active filters shown with arrow indicators (►)

#### Configuration System
- New `Config` struct for passing configuration to TUI
- `NewWithConfig()` function for advanced TUI initialization
- Backward compatible with existing `New()` function

#### Documentation
- **FEATURES.md**: Comprehensive feature documentation with examples
- **QUICKSTART.md**: Quick reference guide for common operations
- **CHANGELOG.md**: Version history and changes
- Enhanced README.md with usage examples and integration guides
- Advanced use case examples (monitoring, alerting, cron jobs)
- Troubleshooting section expanded

### Changed
- Main function refactored to support multiple operating modes
- TUI model enhanced with filtering and configuration support
- View rendering updated to use filtered snapshots
- Help text expanded with new keyboard shortcuts
- Sample interval now configurable (was hardcoded to 5 seconds)
- Export functions integrated directly in main package

### Improved
- Better separation of concerns (sampling, storage, display, export)
- More flexible architecture for future extensions
- Enhanced error handling in export functions
- Clearer status messages in TUI
- More informative help overlay
- Better user experience with filtering indicators

### Technical Details
- Added `encoding/json` and `encoding/csv` imports
- Added `flag` package for CLI argument parsing
- Enhanced model struct with filter fields
- New helper functions: `getUniqueUsers()`, `getFilteredSnapshot()`
- Alert checking logic separated into `checkAlerts()` function
- Export logic in `exportToJSON()` and `exportToCSV()` functions

### Backward Compatibility
- All existing features work unchanged
- Database format unchanged - existing databases fully compatible
- Default behavior preserved when no flags specified
- Existing keyboard shortcuts unchanged
- API compatible with version 1.0.x

## [1.0.0] - Initial Release

### Features
- Live GPU monitoring via nvidia-smi
- Per-user GPU memory tracking
- Historical snapshot storage in SQLite
- Interactive Terminal UI (TUI) using Bubble Tea
- Auto-recording with 5-second intervals
- Manual snapshot saving
- History browsing by date and snapshot
- GPU utilization and temperature display
- Process tracking with PID to user mapping
- Beautiful lipgloss-styled interface
- Help overlay with keyboard shortcuts

### Core Components
- `/cmd/gpu-tracker/main.go`: Application entry point
- `/internal/sampler/nvidia.go`: NVIDIA GPU sampling logic
- `/internal/store/store.go`: SQLite database operations
- `/internal/tui/`: Terminal UI components
- `/internal/types/types.go`: Data structures
- `/internal/util/proc_unix.go`: Process to user mapping

### Initial Keyboard Shortcuts
- `a`: Toggle auto-recording
- `r`: Refresh snapshot
- `s`: Save snapshot
- `h`: Toggle history mode
- `←/→`: Navigate snapshots
- `↑/↓`: Navigate days
- `t`: Jump to today
- `q`: Quit
- `?`: Toggle help

### System Requirements
- Go 1.21+
- Linux OS
- NVIDIA GPU with drivers
- nvidia-smi utility
- gcc for CGO (SQLite)

---

## Version Numbering

This project follows [Semantic Versioning](https://semver.org/):
- MAJOR version for incompatible API changes
- MINOR version for added functionality (backward compatible)
- PATCH version for backward compatible bug fixes

## Upgrade Path

### From 1.0.x to 1.1.0
No special steps required:
1. Replace binary with new version
2. All existing data preserved
3. New features available via flags and key shortcuts
4. No configuration changes needed

## Future Plans

### Planned for 1.2.0
- AMD GPU support (ROCm)
- Web dashboard option
- Prometheus exporter
- Email notification system
- Historical data analysis tools
- Multi-host aggregation

### Under Consideration
- Intel GPU support
- Remote monitoring
- REST API server mode
- Grafana dashboard templates
- Docker container deployment
- Configuration file support
