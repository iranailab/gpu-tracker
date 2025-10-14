package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gpuwatch/internal/store"
	"gpuwatch/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

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

func main() {
	dataDir, err := ensureDataDir()
	if err != nil {
		log.Fatal(err)
	}
	dbPath := filepath.Join(dataDir, "gpuwatch.db")
	db, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	m := tui.New(db)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	_ = time.Second // keep import of time for future flags
}
