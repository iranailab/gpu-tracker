package tui

import (
	"fmt"
	"strings"
	"time"

	"gpuwatch/internal/sampler"
	"gpuwatch/internal/store"
	"gpuwatch/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

type Config struct {
	SampleInterval time.Duration
	MaxTemp        float64
	MaxMem         float64
}

type model struct {
	db         *store.DB
	config     Config
	live       bool
	autoRecord bool
	width      int
	height     int

	curr   types.Snapshot // live or currently viewed snapshot
	status string
	err    error

	// history
	historyDate time.Time
	metas       []store.SnapshotMeta
	index       int // index into metas for current snapshot

	showHelp bool

	// filters
	filterUser string
	filterGPU  int // -1 means all GPUs
	sortByMem  bool
}

type (
	refreshMsg struct{ snap types.Snapshot }
	savedMsg   struct{ id int64 }
	metasMsg   struct{ metas []store.SnapshotMeta }
	errorMsg   struct{ err error }
)

func New(db *store.DB) model {
	return NewWithConfig(db, Config{
		SampleInterval: 5 * time.Second,
		MaxTemp:        90.0,
		MaxMem:         95.0,
	})
}

func NewWithConfig(db *store.DB, config Config) model {
	loc := time.Now().Location()
	return model{
		db:          db,
		config:      config,
		live:        true,
		autoRecord:  true,
		historyDate: time.Now().In(loc),
		filterGPU:   -1, // show all GPUs by default
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.refreshOnce(), m.tickIfNeeded())
}

func (m model) tickIfNeeded() tea.Cmd {
	if m.live && m.autoRecord {
		return tea.Tick(m.config.SampleInterval, func(time.Time) tea.Msg { return m.doSample() })
	}
	return nil
}

func (m model) doSample() tea.Msg {
	s, err := sampler.Sample()
	if err != nil {
		return errorMsg{err}
	}
	// Save when auto record
	if m.autoRecord {
		id, err := m.db.SaveSnapshot(s)
		if err != nil {
			return errorMsg{err}
		}
		s.ID = id
		return refreshMsg{snap: s}
	}
	return refreshMsg{snap: s}
}

func (m model) refreshOnce() tea.Cmd {
	return func() tea.Msg {
		s, err := sampler.Sample()
		if err != nil {
			return errorMsg{err}
		}
		return refreshMsg{snap: s}
	}
}

func (m model) loadMetasCmd(day time.Time) tea.Cmd {
	return func() tea.Msg {
		metas, err := m.db.ListSnapshotsByDate(day)
		if err != nil {
			return errorMsg{err}
		}
		return metasMsg{metas: metas}
	}
}

func (m model) loadByMetaCmd(idx int) tea.Cmd {
	if idx < 0 || idx >= len(m.metas) {
		return nil
	}
	id := m.metas[idx].ID
	return func() tea.Msg {
		s, err := m.db.LoadSnapshot(id)
		if err != nil {
			return errorMsg{err}
		}
		return refreshMsg{snap: s}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case refreshMsg:
		m.curr = msg.snap
		if m.live {
			m.status = fmt.Sprintf("LIVE %s | autosave:%v", m.curr.TS.Format("15:04:05"), m.autoRecord)
		} else {
			m.status = fmt.Sprintf("HISTORY %s (%d/%d)", m.curr.TS.Format("2006-01-02 15:04:05"), m.index+1, len(m.metas))
		}
		m.err = nil
		return m, m.tickIfNeeded()
	case errorMsg:
		m.err = msg.err
		m.status = "error"
		return m, m.tickIfNeeded()
	case savedMsg:
		m.status = fmt.Sprintf("saved snapshot #%d", msg.id)
		return m, nil
	case metasMsg:
		m.metas = msg.metas
		m.index = 0
		if len(m.metas) == 0 {
			m.curr = types.Snapshot{}
			m.status = "no snapshots on this date"
			return m, nil
		}
		return m, m.loadByMetaCmd(m.index)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "a": // toggle auto-recording
			m.autoRecord = !m.autoRecord
			return m, m.tickIfNeeded()
		case "r": // refresh once
			if m.live {
				return m, m.refreshOnce()
			}
			return m, nil
		case "s": // save snapshot immediately
			if m.live {
				return m, func() tea.Msg {
					if m.curr.TS.IsZero() {
						return errorMsg{fmt.Errorf("no current snapshot")}
					}
					id, err := m.db.SaveSnapshot(m.curr)
					if err != nil {
						return errorMsg{err}
					}
					return savedMsg{id: id}
				}
			}
			return m, nil
		case "h": // toggle history mode
			m.live = !m.live
			if m.live {
				return m, m.refreshOnce()
			}
			// entering history: load today metas
			m.historyDate = time.Now()
			return m, m.loadMetasCmd(m.historyDate)
		case "t": // today/live
			m.live = true
			return m, m.refreshOnce()
		case "left":
			if !m.live && len(m.metas) > 0 {
				if m.index > 0 {
					m.index--
				}
				return m, m.loadByMetaCmd(m.index)
			}
		case "right":
			if !m.live && len(m.metas) > 0 {
				if m.index < len(m.metas)-1 {
					m.index++
				}
				return m, m.loadByMetaCmd(m.index)
			}
		case "up":
			if !m.live {
				m.historyDate = m.historyDate.AddDate(0, 0, -1)
				return m, m.loadMetasCmd(m.historyDate)
			}
		case "down":
			if !m.live {
				m.historyDate = m.historyDate.AddDate(0, 0, 1)
				return m, m.loadMetasCmd(m.historyDate)
			}
		case "f": // filter by user
			// Cycle through users in current snapshot
			users := m.getUniqueUsers()
			if len(users) == 0 {
				m.filterUser = ""
			} else if m.filterUser == "" {
				m.filterUser = users[0]
			} else {
				found := false
				for i, u := range users {
					if u == m.filterUser && i < len(users)-1 {
						m.filterUser = users[i+1]
						found = true
						break
					}
				}
				if !found {
					m.filterUser = "" // clear filter
				}
			}
			return m, nil
		case "g": // filter by GPU
			// Cycle through GPUs
			if len(m.curr.GPUs) == 0 {
				m.filterGPU = -1
			} else if m.filterGPU == -1 {
				m.filterGPU = m.curr.GPUs[0].Index
			} else {
				found := false
				for i, gpu := range m.curr.GPUs {
					if gpu.Index == m.filterGPU && i < len(m.curr.GPUs)-1 {
						m.filterGPU = m.curr.GPUs[i+1].Index
						found = true
						break
					}
				}
				if !found {
					m.filterGPU = -1 // show all
				}
			}
			return m, nil
		case "m": // toggle sort by memory
			m.sortByMem = !m.sortByMem
			return m, nil
		case "c": // clear all filters
			m.filterUser = ""
			m.filterGPU = -1
			m.sortByMem = false
			return m, nil
		}
	}
	return m, nil
}

func (m model) getUniqueUsers() []string {
	seen := make(map[string]bool)
	var users []string
	for _, proc := range m.curr.Procs {
		if !seen[proc.User] {
			seen[proc.User] = true
			users = append(users, proc.User)
		}
	}
	return users
}

func (m model) getFilteredSnapshot() types.Snapshot {
	snap := m.curr

	// Apply GPU filter
	if m.filterGPU != -1 {
		var filteredGPUs []types.GPU
		for _, gpu := range snap.GPUs {
			if gpu.Index == m.filterGPU {
				filteredGPUs = append(filteredGPUs, gpu)
			}
		}
		snap.GPUs = filteredGPUs

		// Filter processes for this GPU
		var filteredProcs []types.GPUProcess
		for _, proc := range snap.Procs {
			for _, gpu := range filteredGPUs {
				if proc.GPUUUID == gpu.UUID {
					filteredProcs = append(filteredProcs, proc)
					break
				}
			}
		}
		snap.Procs = filteredProcs
	}

	// Apply user filter
	if m.filterUser != "" {
		var filteredProcs []types.GPUProcess
		for _, proc := range snap.Procs {
			if strings.EqualFold(proc.User, m.filterUser) {
				filteredProcs = append(filteredProcs, proc)
			}
		}
		snap.Procs = filteredProcs
	}

	return snap
}
