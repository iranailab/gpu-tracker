# GPU Tracker - Quick Reference

## Installation
```bash
git clone https://github.com/iranailab/gpu-tracker
cd gpu-tracker
go mod tidy
go build -o gpuwatch ./cmd/gpu-tracker
```

## Quick Start
```bash
# Basic TUI mode
./gpuwatch

# Sample once and see results
./gpuwatch -once

# List users using GPUs
./gpuwatch -list-users

# Export to JSON
./gpuwatch -export json

# Export to CSV file
./gpuwatch -export csv -output report.csv
```

## Command-Line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-interval` | int | 5 | Sampling interval (seconds) |
| `-db` | string | `~/.local/share/gpuwatch/gpuwatch.db` | Database path |
| `-once` | bool | false | Sample once and exit |
| `-continuous` | bool | false | Continuous monitoring mode |
| `-export` | string | - | Export format: json, csv |
| `-output` | string | stdout | Export output file |
| `-list-users` | bool | false | List GPU users and exit |
| `-max-temp` | float | 90.0 | Temperature alert (°C) |
| `-max-mem` | float | 95.0 | Memory alert (%) |
| `-version` | bool | false | Show version |

## TUI Keyboard Shortcuts

### Basic Controls
| Key | Action |
|-----|--------|
| `q` | Quit |
| `?` | Toggle help |
| `a` | Toggle auto-record |
| `r` | Refresh now |
| `s` | Save snapshot |

### Navigation
| Key | Action |
|-----|--------|
| `h` | Toggle history mode |
| `t` | Jump to today/live |
| `←` | Previous snapshot |
| `→` | Next snapshot |
| `↑` | Previous day |
| `↓` | Next day |

### Filters (NEW)
| Key | Action |
|-----|--------|
| `f` | Cycle user filter |
| `g` | Cycle GPU filter |
| `m` | Sort by memory |
| `c` | Clear filters |

## Common Use Cases

### Monitor in Real-Time
```bash
./gpuwatch
# Press 'a' to enable auto-recording
# Press 'f' to filter by user
# Press 'g' to filter by GPU
```

### Background Monitoring
```bash
# Sample every 30 seconds
./gpuwatch -continuous -interval 30 &

# Or with log file
./gpuwatch -continuous -interval 60 >> gpu.log 2>&1 &
```

### Generate Reports
```bash
# Daily CSV report
./gpuwatch -export csv -output "gpu-$(date +%F).csv"

# JSON for API
./gpuwatch -export json | curl -X POST https://api.example.com/metrics -d @-
```

### Check Specific User
```bash
# List all users first
./gpuwatch -list-users

# Then monitor in TUI and press 'f' to cycle to that user
./gpuwatch
```

### Alert on High Temperature
```bash
#!/bin/bash
./gpuwatch -once -max-temp 80 2>&1 | grep -q "ALERT" && \
  echo "High GPU temp detected!" | mail -s "GPU Alert" you@example.com
```

### Data Analysis
```bash
# Extract GPU names
./gpuwatch -export json | jq -r '.GPUs[].Name'

# Total memory usage
./gpuwatch -export csv | awk -F',' 'NR>1 {sum+=$13} END {print sum " MB"}'

# Users sorted by memory
./gpuwatch -export json | jq -r '.Procs[] | "\(.User) \(.UsedMemMB)"' | sort -k2 -rn
```

### Cron Jobs
```bash
# Every 5 minutes, check and alert
*/5 * * * * /usr/local/bin/gpuwatch -once -max-temp 85 -max-mem 90 2>&1 | logger -t gpuwatch

# Hourly CSV snapshot
0 * * * * /usr/local/bin/gpuwatch -export csv -output "/data/gpu-$(date +\%H).csv"

# Daily database backup
0 0 * * * cp ~/.local/share/gpuwatch/gpuwatch.db ~/backups/gpuwatch-$(date +\%F).db
```

## Troubleshooting

### nvidia-smi not found
```bash
# Test nvidia-smi
nvidia-smi -L

# Add to PATH if needed
export PATH=$PATH:/usr/local/cuda/bin
```

### Permission denied
```bash
# Run as regular user, not root
./gpuwatch

# If needed, check file permissions
ls -l gpuwatch
chmod +x gpuwatch
```

### Database locked
```bash
# Only one instance can write
# Use different database for multiple instances
./gpuwatch -db /tmp/gpuwatch1.db &
./gpuwatch -db /tmp/gpuwatch2.db &
```

### High CPU usage
```bash
# Increase sampling interval
./gpuwatch -interval 30

# Or in continuous mode
./gpuwatch -continuous -interval 60
```

## Tips & Tricks

1. **Combine filters**: Use `f` and `g` together to see specific user on specific GPU
2. **Export for analysis**: Regular CSV exports make trend analysis easy
3. **Background monitoring**: Run with `-continuous` on server startup
4. **Custom intervals**: Match sampling to your workload (fast for dev, slow for production)
5. **Alert tuning**: Adjust `-max-temp` and `-max-mem` based on your hardware
6. **History browsing**: Use `h` to enter history, then arrow keys to explore
7. **Quick checks**: Use `-list-users` for fast overview without TUI
8. **Scripting**: Use `-once` with `-export json` for automation

## Environment

### Recommended
- Go 1.22+
- Linux with NVIDIA GPU
- nvidia-smi in PATH
- Terminal with 24-bit color support

### Minimum
- Go 1.21+
- Linux with NVIDIA drivers
- nvidia-smi accessible
- Any terminal

## Files & Locations

```
~/.local/share/gpuwatch/
  └── gpuwatch.db          # Default database

./gpuwatch                 # Binary
./FEATURES.md             # Detailed feature documentation
./README.md               # Full documentation
```

## Getting Help

1. Press `?` in TUI for keyboard shortcuts
2. Run `./gpuwatch -h` for flag help
3. Check FEATURES.md for detailed examples
4. See README.md for full documentation

## Version

Current: **1.1.0**

Run `./gpuwatch -version` to check your version.
