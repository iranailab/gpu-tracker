package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gpuwatch/internal/sampler"
	"gpuwatch/internal/store"
	"gpuwatch/internal/tui"
	"gpuwatch/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	sampleIntervalFlag = flag.Int("interval", 5, "Sampling interval in seconds (default: 5)")
	exportFormat       = flag.String("export", "", "Export current snapshot to file (formats: json, csv)")
	exportFile         = flag.String("output", "", "Output file for export (default: stdout)")
	dbPathFlag         = flag.String("db", "", "Custom database path (default: ~/.local/share/gpuwatch/gpuwatch.db)")
	oneShotMode        = flag.Bool("once", false, "Sample once and exit (no TUI)")
	continuousMode     = flag.Bool("continuous", false, "Continuously sample and save without TUI")
	showVersion        = flag.Bool("version", false, "Show version information")
	maxTemp            = flag.Float64("max-temp", 90.0, "Alert threshold for GPU temperature (°C)")
	maxMem             = flag.Float64("max-mem", 95.0, "Alert threshold for memory usage (%)")
	listUsers          = flag.Bool("list-users", false, "List all users using GPUs and exit")
)

const version = "1.1.0"

func ensureDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(home, ".local", "share", "gpuwatch")
	if err := os.MkdirAll(p, 0o755); err != nil {
		return "", err
	}
	return p, nil
}

func exportToJSON(snap types.Snapshot, path string) error {
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	if path == "" {
		fmt.Println(string(data))
		return nil
	}
	return os.WriteFile(path, data, 0644)
}

func exportToCSV(snap types.Snapshot, path string) error {
	var w *csv.Writer
	if path == "" {
		w = csv.NewWriter(os.Stdout)
	} else {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		w = csv.NewWriter(f)
	}
	defer w.Flush()

	// Write header
	if err := w.Write([]string{"Timestamp", "GPU Index", "GPU Name", "GPU Util %", "Mem Util %", "Mem Used MB", "Mem Total MB", "Temp C", "Power W", "PID", "Process", "User", "Proc Mem MB"}); err != nil {
		return err
	}

	ts := snap.TS.Format(time.RFC3339)
	for _, gpu := range snap.GPUs {
		// Find processes for this GPU
		hasProc := false
		for _, proc := range snap.Procs {
			if proc.GPUUUID == gpu.UUID {
				hasProc = true
				if err := w.Write([]string{
					ts,
					fmt.Sprintf("%d", gpu.Index),
					gpu.Name,
					fmt.Sprintf("%.1f", gpu.UtilGPU),
					fmt.Sprintf("%.1f", gpu.UtilMem),
					fmt.Sprintf("%.1f", gpu.MemUsedMB),
					fmt.Sprintf("%.1f", gpu.MemTotalMB),
					fmt.Sprintf("%.1f", gpu.TempC),
					fmt.Sprintf("%.1f", gpu.PowerDrawW),
					fmt.Sprintf("%d", proc.PID),
					proc.ProcessName,
					proc.User,
					fmt.Sprintf("%.1f", proc.UsedMemMB),
				}); err != nil {
					return err
				}
			}
		}
		if !hasProc {
			if err := w.Write([]string{
				ts,
				fmt.Sprintf("%d", gpu.Index),
				gpu.Name,
				fmt.Sprintf("%.1f", gpu.UtilGPU),
				fmt.Sprintf("%.1f", gpu.UtilMem),
				fmt.Sprintf("%.1f", gpu.MemUsedMB),
				fmt.Sprintf("%.1f", gpu.MemTotalMB),
				fmt.Sprintf("%.1f", gpu.TempC),
				fmt.Sprintf("%.1f", gpu.PowerDrawW),
				"", "", "", "",
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func listUsersMode(snap types.Snapshot) {
	userMemMap := make(map[string]float64)
	for _, proc := range snap.Procs {
		userMemMap[proc.User] += proc.UsedMemMB
	}
	fmt.Println("Users currently using GPUs:")
	fmt.Println("User\t\tMemory (MB)")
	fmt.Println("----\t\t-----------")
	for user, mem := range userMemMap {
		fmt.Printf("%s\t\t%.1f\n", user, mem)
	}
}

func checkAlerts(snap types.Snapshot, maxTemp, maxMem float64) {
	for _, gpu := range snap.GPUs {
		if gpu.TempC > maxTemp {
			fmt.Fprintf(os.Stderr, "⚠️  ALERT: GPU %d (%s) temperature %.1f°C exceeds threshold %.1f°C\n",
				gpu.Index, gpu.Name, gpu.TempC, maxTemp)
		}
		if gpu.UtilMem > maxMem {
			fmt.Fprintf(os.Stderr, "⚠️  ALERT: GPU %d (%s) memory utilization %.1f%% exceeds threshold %.1f%%\n",
				gpu.Index, gpu.Name, gpu.UtilMem, maxMem)
		}
	}
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("gpuwatch version %s\n", version)
		return
	}

	// Get database path
	var dbPath string
	if *dbPathFlag != "" {
		dbPath = *dbPathFlag
	} else {
		dataDir, err := ensureDataDir()
		if err != nil {
			log.Fatal(err)
		}
		dbPath = filepath.Join(dataDir, "gpuwatch.db")
	}

	// One-shot mode: sample once and optionally export
	if *oneShotMode || *listUsers || *exportFormat != "" {
		snap, err := sampler.Sample()
		if err != nil {
			log.Fatalf("Failed to sample: %v", err)
		}

		checkAlerts(snap, *maxTemp, *maxMem)

		if *listUsers {
			listUsersMode(snap)
			return
		}

		if *exportFormat != "" {
			switch *exportFormat {
			case "json":
				if err := exportToJSON(snap, *exportFile); err != nil {
					log.Fatalf("Export failed: %v", err)
				}
			case "csv":
				if err := exportToCSV(snap, *exportFile); err != nil {
					log.Fatalf("Export failed: %v", err)
				}
			default:
				log.Fatalf("Unknown export format: %s (supported: json, csv)", *exportFormat)
			}
			return
		}

		// Just print snapshot
		fmt.Printf("Snapshot at %s\n", snap.TS.Format(time.RFC3339))
		for _, gpu := range snap.GPUs {
			fmt.Printf("GPU %d: %s - Util: %.1f%%, Mem: %.1f%%, Temp: %.1f°C\n",
				gpu.Index, gpu.Name, gpu.UtilGPU, gpu.UtilMem, gpu.TempC)
		}
		return
	}

	// Continuous mode: sample and save without TUI
	if *continuousMode {
		db, err := store.Open(dbPath)
		if err != nil {
			log.Fatalf("open db: %v", err)
		}
		defer db.Close()

		fmt.Printf("Continuous mode: sampling every %d seconds (Ctrl+C to stop)\n", *sampleIntervalFlag)
		ticker := time.NewTicker(time.Duration(*sampleIntervalFlag) * time.Second)
		defer ticker.Stop()

		for {
			snap, err := sampler.Sample()
			if err != nil {
				log.Printf("Sample error: %v", err)
				continue
			}
			checkAlerts(snap, *maxTemp, *maxMem)
			id, err := db.SaveSnapshot(snap)
			if err != nil {
				log.Printf("Save error: %v", err)
			} else {
				fmt.Printf("[%s] Saved snapshot #%d\n", snap.TS.Format("15:04:05"), id)
			}
			<-ticker.C
		}
	}

	// Normal TUI mode
	db, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	sampleInterval := time.Duration(*sampleIntervalFlag) * time.Second
	m := tui.NewWithConfig(db, tui.Config{
		SampleInterval: sampleInterval,
		MaxTemp:        *maxTemp,
		MaxMem:         *maxMem,
	})
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
