package tui

import (
	"fmt"
	"sort"
	"strings"

	"gpuwatch/internal/types"

	lg "github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}

	header := headStyle.Render("gpuwatch — per‑user GPU usage") + "  " + subtle.Render(m.status)
	if m.err != nil {
		header += "  " + errStyle.Render(m.err.Error())
	}

	// Show active filters
	if m.filterUser != "" || m.filterGPU != -1 || m.sortByMem {
		var filters []string
		if m.filterUser != "" {
			filters = append(filters, fmt.Sprintf("user:%s", m.filterUser))
		}
		if m.filterGPU != -1 {
			filters = append(filters, fmt.Sprintf("GPU:%d", m.filterGPU))
		}
		if m.sortByMem {
			filters = append(filters, "sorted:mem")
		}
		header += "  " + lg.NewStyle().Foreground(lg.Color("#FFA500")).Render(fmt.Sprintf("[filters: %s]", strings.Join(filters, ", ")))
	}

	body := m.renderBody()
	help := m.renderHelp()

	return header + "\n\n" + body + "\n\n" + help
}

func (m model) renderBody() string {
	left := m.renderGPUs()
	right := m.renderUsers()
	bottom := m.renderProcs()

	// layout: two columns top, then bottom full width
	row := lg.JoinHorizontal(lg.Top, left, right)
	return row + "\n" + bottom
}

func (m model) renderGPUs() string {
	snap := m.getFilteredSnapshot()
	if len(snap.GPUs) == 0 {
		return box.Width(m.width/2 - 4).Render(subtle.Render("no GPU data"))
	}
	var lines []string
	for _, g := range snap.GPUs {
		title := label.Render(fmt.Sprintf("GPU %d — %s", g.Index, g.Name))
		util := fmt.Sprintf("util %2.0f%% | mem %2.0f%% (%0.0f/%0.0f MB)", g.UtilGPU, g.UtilMem, g.MemUsedMB, g.MemTotalMB)
		therm := fmt.Sprintf("temp %2.0f°C | power %0.0f/%0.0f W", g.TempC, g.PowerDrawW, g.PowerLimitW)

		// Add alert indicators
		var alerts []string
		if g.TempC > m.config.MaxTemp {
			alerts = append(alerts, fmt.Sprintf("⚠️  HIGH TEMP %.0f°C", g.TempC))
		}
		if g.UtilMem > m.config.MaxMem {
			alerts = append(alerts, fmt.Sprintf("⚠️  HIGH MEM %.0f%%", g.UtilMem))
		}

		lines = append(lines, title)
		lines = append(lines, drawBar(g.UtilGPU, 100, 24))
		lines = append(lines, subtle.Render(util))
		lines = append(lines, subtle.Render(therm))
		if len(alerts) > 0 {
			lines = append(lines, lg.NewStyle().Foreground(lg.Color("#FF0000")).Render(strings.Join(alerts, " ")))
		}
		lines = append(lines, "")
	}
	content := strings.Join(lines, "\n")
	return box.Width(m.width/2 - 4).Render(content)
}

func (m model) renderUsers() string {
	snap := m.getFilteredSnapshot()
	if len(snap.Procs) == 0 {
		return box.Width(m.width/2 - 4).Render(subtle.Render("no running GPU processes"))
	}
	agg := make(map[string]float64)
	for _, p := range snap.Procs {
		agg[p.User] += p.UsedMemMB
	}
	var users []types.UserAgg
	for u, v := range agg {
		users = append(users, types.UserAgg{User: u, MemUsedMB: v})
	}
	sort.Slice(users, func(i, j int) bool { return users[i].MemUsedMB > users[j].MemUsedMB })

	var lines []string
	lines = append(lines, label.Render("Per‑user GPU memory (MB)"))
	max := 1.0
	for _, u := range users {
		if u.MemUsedMB > max {
			max = u.MemUsedMB
		}
	}
	for _, u := range users {
		bar := drawBar(u.MemUsedMB, max, 30)
		userLabel := u.User
		if m.filterUser == u.User {
			userLabel = "►" + userLabel
		}
		lines = append(lines, fmt.Sprintf("%-12s %s %5.0f", userLabel, bar, u.MemUsedMB))
	}
	content := strings.Join(lines, "\n")
	return box.Width(m.width/2 - 4).Render(content)
}

func (m model) renderProcs() string {
	snap := m.getFilteredSnapshot()
	var b strings.Builder
	b.WriteString(label.Render("Top GPU processes (by used MB)") + "\n")
	if len(snap.Procs) == 0 {
		b.WriteString(subtle.Render("none"))
		return box.Width(m.width - 4).Render(b.String())
	}
	procs := append([]types.GPUProcess(nil), snap.Procs...)

	// Apply sorting if enabled
	if m.sortByMem {
		sort.Slice(procs, func(i, j int) bool { return procs[i].UsedMemMB > procs[j].UsedMemMB })
	} else {
		sort.Slice(procs, func(i, j int) bool { return procs[i].UsedMemMB > procs[j].UsedMemMB })
	}

	maxN := 10
	if len(procs) < maxN {
		maxN = len(procs)
	}
	for i := 0; i < maxN; i++ {
		p := procs[i]
		b.WriteString(fmt.Sprintf("%5d  %-12s  %-22s  %6.0f MB  %s\n", p.PID, p.User, trim(p.ProcessName, 22), p.UsedMemMB, shortUUID(p.GPUUUID)))
	}
	return box.Width(m.width - 4).Render(b.String())
}

func (m model) renderHelp() string {
	if !m.showHelp {
		return subtle.Render("a: auto | r: refresh | s: save | h: history | f: filter user | g: filter GPU | c: clear | ?: help | q: quit")
	}
	return box.Width(m.width - 4).Render(strings.Join([]string{
		"Navigation & Actions:",
		"  a — Toggle auto-recording of live samples",
		"  r — Refresh once (live mode)",
		"  s — Save the current snapshot immediately",
		"  h — Toggle History mode",
		"  ←/→ — Previous/Next snapshot of the selected date",
		"  ↑/↓ — Move one day back/forward",
		"  t — Jump back to today and live mode",
		"",
		"Filters & Display:",
		"  f — Cycle through users to filter by specific user",
		"  g — Cycle through GPUs to filter by specific GPU",
		"  m — Toggle sort processes by memory usage",
		"  c — Clear all active filters",
		"",
		"  q — Quit",
	}, "\n"))
}

func drawBar(value, max float64, width int) string {
	if max <= 0 {
		max = 1
	}
	ratio := value / max
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(float64(width) * ratio)
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	return bar.Width(filled).Render(strings.Repeat(" ", filled)) + subtle.Width(width-filled).Render(strings.Repeat("·", width-filled))
}

func trim(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func shortUUID(u string) string {
	if len(u) <= 8 {
		return u
	}
	return u[len(u)-8:]
}
