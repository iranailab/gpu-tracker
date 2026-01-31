# GPU Tracker - New Features Summary

This document describes all the new features and options added to the GPU tracker application.

## Version 1.1.0 - New Features

### 1. Command-Line Flags

The application now supports comprehensive command-line configuration:

#### Basic Options
- **`-interval <seconds>`**: Set custom sampling interval (default: 5 seconds)
  - Example: `./gpuwatch -interval 10`
  
- **`-db <path>`**: Specify custom database location
  - Example: `./gpuwatch -db /custom/path/gpuwatch.db`
  
- **`-version`**: Display version information
  - Example: `./gpuwatch -version`

#### Operation Modes
- **`-once`**: Sample once and exit without starting the TUI
  - Use case: Quick status check or scripting
  - Example: `./gpuwatch -once`
  
- **`-continuous`**: Continuous background monitoring mode
  - Samples at regular intervals and saves to database
  - No TUI, runs until stopped with Ctrl+C
  - Example: `./gpuwatch -continuous -interval 30`

- **`-list-users`**: List all users currently using GPUs
  - Quick summary of user memory consumption
  - Example: `./gpuwatch -list-users`

#### Export Options
- **`-export <format>`**: Export snapshot data (formats: `json`, `csv`)
  - JSON: Full structured data export
  - CSV: Tabular format for spreadsheets
  - Example: `./gpuwatch -export json`
  
- **`-output <file>`**: Specify output file for exports
  - If omitted, prints to stdout
  - Example: `./gpuwatch -export csv -output report.csv`

#### Alert Thresholds
- **`-max-temp <degrees>`**: GPU temperature alert threshold (default: 90°C)
  - Visual alerts in TUI when exceeded
  - Example: `./gpuwatch -max-temp 85`
  
- **`-max-mem <percent>`**: Memory utilization alert threshold (default: 95%)
  - Visual alerts in TUI when exceeded
  - Example: `./gpuwatch -max-mem 90`

### 2. Export Functionality

#### JSON Export
Exports complete snapshot data in JSON format including:
- Full GPU information (utilization, memory, temperature, power)
- Process details (PID, name, user, memory usage)
- Timestamp information

**Example:**
```bash
./gpuwatch -export json -output snapshot.json
```

**Use cases:**
- Integration with monitoring systems
- Data analysis with jq or Python
- API consumption

#### CSV Export
Exports data in CSV format suitable for spreadsheets:
- Columns: Timestamp, GPU Index, GPU Name, Utilization %, Memory %, Temperature, Power, PID, Process, User, Memory MB
- Easy to import into Excel, Google Sheets, or process with awk/sed

**Example:**
```bash
./gpuwatch -export csv -output report.csv
```

### 3. TUI Filtering and Display Options

#### Filter by User (Key: `f`)
- Cycle through users to filter view by specific user
- Shows only processes belonging to the selected user
- Indicator shows active user filter
- Press `f` multiple times to cycle through all users
- Clear with `c`

#### Filter by GPU (Key: `g`)
- Cycle through GPUs to focus on specific GPU
- Shows only the selected GPU and its processes
- Useful for multi-GPU systems
- Press `g` multiple times to cycle through all GPUs
- Clear with `c`

#### Sort by Memory (Key: `m`)
- Toggle sorting of processes by memory usage
- Helps identify memory-intensive processes
- Works in combination with other filters

#### Clear Filters (Key: `c`)
- Resets all active filters
- Returns to full system view

### 4. Alert System

#### Visual Indicators
- **Temperature Alerts**: Red warning when GPU temperature exceeds threshold
- **Memory Alerts**: Red warning when memory utilization exceeds threshold
- Displayed directly in TUI next to GPU stats
- Example: `⚠️  HIGH TEMP 92°C`

#### Command-Line Alerts
In one-shot and continuous modes, alerts are printed to stderr:
```
⚠️  ALERT: GPU 0 (NVIDIA GeForce RTX 3090) temperature 91.0°C exceeds threshold 90.0°C
```

This allows for easy integration with monitoring scripts and email alerts.

### 5. Multiple Operating Modes

#### Interactive TUI Mode (Default)
- Full terminal user interface
- Real-time updates
- History browsing
- Filtering and sorting

**Usage:**
```bash
./gpuwatch
```

#### One-Shot Mode
- Sample once and display
- Perfect for scripting
- Can be combined with export

**Usage:**
```bash
./gpuwatch -once
./gpuwatch -once -export json
```

#### Continuous Mode
- Background monitoring
- Automatic database saves
- No TUI overhead
- Alert notifications to stderr

**Usage:**
```bash
./gpuwatch -continuous -interval 60 >> /var/log/gpuwatch.log 2>&1
```

#### List Users Mode
- Quick summary of GPU users
- Shows memory usage per user
- Fast and lightweight

**Usage:**
```bash
./gpuwatch -list-users
```

## Integration Examples

### Cron Job for Regular Monitoring
```bash
# Sample every 10 minutes
*/10 * * * * /usr/local/bin/gpuwatch -continuous -interval 600 >> /var/log/gpuwatch.log 2>&1
```

### Email Alerts on High Temperature
```bash
#!/bin/bash
./gpuwatch -once -max-temp 85 2>&1 | grep "ALERT" && \
  echo "GPU temperature alert!" | mail -s "GPU Alert" admin@example.com
```

### Daily CSV Report
```bash
#!/bin/bash
DATE=$(date +%Y-%m-%d)
./gpuwatch -export csv -output "/reports/gpu-report-${DATE}.csv"
```

### JSON API Integration
```bash
# Get current GPU data as JSON
curl -X POST https://api.example.com/gpu-metrics \
  -H "Content-Type: application/json" \
  -d "$(./gpuwatch -export json)"
```

### Process with jq
```bash
# Get GPU 0 temperature
./gpuwatch -export json | jq '.GPUs[0].TempC'

# Get all users and their memory usage
./gpuwatch -export json | jq '.Procs[] | {user: .User, mem: .UsedMemMB}'

# Check if any GPU is over 80% memory
./gpuwatch -export json | jq '.GPUs[] | select(.UtilMem > 80)'
```

## Updated Key Bindings

### Navigation & Actions
- `a` - Toggle auto-recording
- `r` - Refresh snapshot once
- `s` - Save snapshot manually
- `h` - Toggle History mode
- `←/→` - Previous/Next snapshot (in History)
- `↑/↓` - Previous/Next day (in History)
- `t` - Jump to today/live mode
- `q` - Quit
- `?` - Toggle help overlay

### NEW: Filters & Display
- `f` - Cycle through users to filter
- `g` - Cycle through GPUs to filter
- `m` - Toggle sort by memory usage
- `c` - Clear all active filters

## Configuration Best Practices

### Development/Testing
```bash
# Fast sampling, low thresholds
./gpuwatch -interval 2 -max-temp 70 -max-mem 80
```

### Production Monitoring
```bash
# Moderate sampling, reasonable thresholds
./gpuwatch -interval 30 -max-temp 85 -max-mem 90
```

### Long-term Data Collection
```bash
# Slow sampling, save resources
./gpuwatch -continuous -interval 300 -db /data/gpuwatch/history.db
```

### Emergency Debugging
```bash
# Very fast sampling, strict alerts
./gpuwatch -interval 1 -max-temp 75 -max-mem 85
```

## Upgrade Notes

### Breaking Changes
None - All new features are opt-in via flags or keyboard shortcuts.

### Default Behavior
- Default sampling interval remains 5 seconds
- Alert thresholds default to 90°C and 95%
- TUI mode is still the default when no flags are specified
- Database location unchanged: `~/.local/share/gpuwatch/gpuwatch.db`

### Compatibility
- Existing database files work without modification
- Old keyboard shortcuts remain unchanged
- New shortcuts don't conflict with existing ones
