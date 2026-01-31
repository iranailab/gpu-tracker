# gpuwatch

![logo](./assets/logo.png)

**gpuwatch** is a beautiful and modular terminal-based application written in Go that monitors **GPU usage per user**, displays it live in a TUI (Terminal User Interface), and saves usage history for browsing by date—all inside the terminal.  
Powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and SQLite.

---

## Features

- **Live GPU Monitoring:**  
  See detailed GPU stats and per-user memory usage in real time (via `nvidia-smi`).

- **User Breakdown:**  
  Aggregates running processes on each GPU and maps them to users.

- **Historical Browsing:**  
  All usage snapshots are stored; navigate through any day and view every snapshot.

- **Flexible Export:**  
  Export snapshots to JSON or CSV format for analysis and reporting.

- **Advanced Filtering:**  
  Filter by specific user or GPU, clear view for focused monitoring.

- **Alert Thresholds:**  
  Configurable temperature and memory usage alerts with visual indicators.

- **Multiple Modes:**  
  TUI mode, one-shot sampling, continuous background monitoring, or export mode.

- **Customizable Sampling:**  
  Configure sampling intervals and database location via command-line flags.

- **Elegant Terminal UI:**  
  Beautiful, colorful, and informative display with keybindings for productivity.

- **Portable and Modular:**  
  Clean architecture—easy to extend and modify for your cluster or desktop setup.

---

## screenshot

![gpuwatch tui preview](./assets/screenshot.png)


## Installation

### Prerequisites

- **Go 1.21+** (recommended: Go 1.22 or newer)
- Linux (tested), with NVIDIA drivers and `nvidia-smi` available in PATH
- `gcc` (for go-sqlite3, if not present: `sudo apt install build-essential`)
- Optional: color-capable terminal (for best UI experience)

### Build

```bash
git clone hhttps://github.com/iranailab/gpu-tracker
cd gpu-tracker

go mod tidy
go build -o gpuwatch ./cmd/gpu-tracker
````

### Run

```bash
./gpuwatch
```

> On first run, the app creates its database in `~/.local/share/gpuwatch/gpuwatch.db`.

---

## Usage

### Running the TUI (Default Mode)

```bash
./gpuwatch
```

### Command-Line Options

```bash
./gpuwatch [OPTIONS]
```

**Available options:**

| Flag | Description | Default |
|------|-------------|----------|
| `-interval` | Sampling interval in seconds | 5 |
| `-db` | Custom database path | `~/.local/share/gpuwatch/gpuwatch.db` |
| `-once` | Sample once and exit (no TUI) | false |
| `-continuous` | Continuously sample and save without TUI | false |
| `-export` | Export format: `json` or `csv` | - |
| `-output` | Output file for export (default: stdout) | - |
| `-list-users` | List all users using GPUs and exit | false |
| `-max-temp` | Alert threshold for GPU temperature (°C) | 90.0 |
| `-max-mem` | Alert threshold for memory usage (%) | 95.0 |
| `-version` | Show version information | false |

### Usage Examples

**1. Basic TUI mode with default settings:**
```bash
./gpuwatch
```

**2. Custom sampling interval (10 seconds):**
```bash
./gpuwatch -interval 10
```

**3. One-shot sampling (sample once and display):**
```bash
./gpuwatch -once
```

**4. Continuous background monitoring:**
```bash
./gpuwatch -continuous -interval 30
```

**5. Export current snapshot to JSON:**
```bash
./gpuwatch -export json -output snapshot.json
```

**6. Export current snapshot to CSV:**
```bash
./gpuwatch -export csv -output snapshot.csv
```

**7. Export to stdout (pipe to other tools):**
```bash
./gpuwatch -export json | jq '.GPUs[0].Name'
```

**8. List users currently using GPUs:**
```bash
./gpuwatch -list-users
```

**9. Custom alert thresholds:**
```bash
./gpuwatch -max-temp 80 -max-mem 90
```

**10. Custom database location:**
```bash
./gpuwatch -db /path/to/custom/gpuwatch.db
```

### TUI Key Bindings

**Navigation & Actions:**

| Key     | Action                                 |
| ------- | -------------------------------------- |
| `a`     | Toggle auto-recording (live, configurable interval) |
| `r`     | Refresh snapshot once                  |
| `s`     | Save a snapshot manually               |
| `h`     | Toggle History mode                    |
| `← / →` | Prev/Next snapshot (in History)        |
| `↑ / ↓` | Prev/Next day (in History)             |
| `t`     | Jump to today/live mode                |
| `q`     | Quit                                   |
| `?`     | Toggle help overlay                    |

**Filters & Display:**

| Key     | Action                                 |
| ------- | -------------------------------------- |
| `f`     | Cycle through users to filter          |
| `g`     | Cycle through GPUs to filter           |
| `m`     | Toggle sort by memory usage            |
| `c`     | Clear all active filters               |

---

## How It Works

* **Sampling:**
  The app runs `nvidia-smi` to capture GPU/process stats. For each process, it maps PID → UID (via `/proc/<pid>/status`) → username (`/etc/passwd`).
  
* **History:**
  Snapshots are saved to SQLite on disk. Auto-recording can be toggled or snapshots saved manually.
  
* **Browsing:**
  Switch to history mode and browse by day/snapshot, all within the TUI.
  
* **Filtering:**
  Filter the view by specific users or GPUs to focus on relevant data. Use keyboard shortcuts to cycle through available filters.
  
* **Alerts:**
  Visual indicators appear when GPU temperature or memory usage exceeds configured thresholds. Alerts are also shown in continuous mode.
  
* **Export:**
  Export snapshots to JSON or CSV format for integration with other tools, reporting, or analysis.
  
* **Modes:**
  - **TUI Mode (default):** Interactive terminal UI with real-time updates
  - **One-shot Mode:** Sample once and display/export
  - **Continuous Mode:** Background monitoring that saves snapshots automatically
  - **List Mode:** Quick overview of current GPU users
  
* **Extensible:**
  Sampler and database logic are separated—add support for AMD (ROCm), NVML, or other GPUs easily.

---

## Project Structure

```
.
├── cmd/
│   └── gpuwatch/       # App entry point (main.go)
├── internal/
│   ├── sampler/        # GPU/process sampling logic
│   ├── store/          # SQLite storage abstraction
│   ├── tui/            # TUI (Bubble Tea) code
│   ├── types/          # Shared types & models
│   └── util/           # Helpers (e.g., PID->User)
├── go.mod
├── go.sum
└── README.md
```

---

## Advanced Use Cases

### Integration with Monitoring Systems

**Prometheus/Grafana Integration:**
```bash
# Export to JSON and parse with jq
./gpuwatch -export json | jq -r '.GPUs[] | "\(.Name) \(.UtilGPU)"'
```

**Alerting Script:**
```bash
#!/bin/bash
# Check GPU usage and send alerts
./gpuwatch -once -max-temp 85 -max-mem 90 2>&1 | grep "ALERT" && \
  echo "GPU alert detected!" | mail -s "GPU Alert" admin@example.com
```

**Cron Job for Regular Sampling:**
```bash
# Add to crontab: sample every 5 minutes
*/5 * * * * /path/to/gpuwatch -continuous -interval 300 >> /var/log/gpuwatch.log 2>&1
```

### Data Analysis

**Export historical data for analysis:**
```bash
# Export current state to CSV
./gpuwatch -export csv -output daily_report.csv

# Process with standard tools
cat daily_report.csv | awk -F',' '{sum+=$7} END {print "Total GPU Memory: " sum " MB"}'
```

**Monitor specific user:**
```bash
# Run TUI and filter by user immediately
# Press 'f' to cycle through users, or use export mode:
./gpuwatch -list-users
```

---

## Troubleshooting

* **Go version too old:**
  See your Go version with `go version`. For Go < 1.21, [download a new version here](https://go.dev/dl/).

* **Permission errors:**
  Make sure you own all files (use `chown`) and build/run as your regular user, not root.

* **GLIBC errors on run:**
  Build and run the binary on the same Linux distribution.
  
* **nvidia-smi not found:**
  Ensure NVIDIA drivers are installed and `nvidia-smi` is in your PATH. Test with: `nvidia-smi -L`
  
* **Database locked errors:**
  If running multiple instances, ensure only one instance writes to the database, or use different database paths with `-db` flag.
  
* **High CPU usage in continuous mode:**
  Increase the sampling interval: `./gpuwatch -continuous -interval 60` (samples every 60 seconds)

* **Export returns empty data:**
  Ensure GPUs are available and nvidia-smi is working. Try `./gpuwatch -once` first to verify sampling works.

---

## License

MIT License.
See [LICENSE](./LICENSE) for details.

---

## Credits

* [Bubble Tea](https://github.com/charmbracelet/bubbletea)
* [Lip Gloss](https://github.com/charmbracelet/lipgloss)
* [go-sqlite3](https://github.com/mattn/go-sqlite3)

---

## Author

Developed by [Alireza Parvaresh](https://github.com/parvvaresh)
Contributions welcome!
---

