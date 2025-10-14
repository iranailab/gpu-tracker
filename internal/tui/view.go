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
	if len(m.curr.GPUs) == 0 {
		return box.Width(m.width/2 - 4).Render(subtle.Render("no GPU data"))
	}
	var lines []string
	for _, g := range m.curr.GPUs {
		title := label.Render(fmt.Sprintf("GPU %d — %s", g.Index, g.Name))
		util := fmt.Sprintf("util %2.0f%% | mem %2.0f%% (%0.0f/%0.0f MB)", g.UtilGPU, g.UtilMem, g.MemUsedMB, g.MemTotalMB)
		therm := fmt.Sprintf("temp %2.0f°C | power %0.0f/%0.0f W", g.TempC, g.PowerDrawW, g.PowerLimitW)
		lines = append(lines, title)
		lines = append(lines, drawBar(g.UtilGPU, 100, 24))
		lines = append(lines, subtle.Render(util))
		lines = append(lines, subtle.Render(therm))
		lines = append(lines, "")
	}
	content := strings.Join(lines, "\n")
	return box.Width(m.width/2 - 4).Render(content)
}

func (m model) renderUsers() string {
	if len(m.curr.Procs) == 0 {
		return box.Width(m.width/2 - 4).Render(subtle.Render("no running GPU processes"))
	}
	agg := make(map[string]float64)
	for _, p := range m.curr.Procs {
		agg[p.User] += p.UsedMemMB
	}
	var users []types.UserAgg
	for u, v := range agg {
		users = append(users, types.UserAgg{User: u, MemUsedMB: v})
	}
	sort.Slice(users, func(i, j int) bool { return users[i].MemUsedMB > users[j].MemUsedMB })

	var lines []string
	lines = append(lines, label.Render("Per‑user GPU memory (MB)") )
	max := 1.0
	for _, u := range users { if u.MemUsedMB > max { max = u.MemUsedMB } }
	for _, u := range users {
		bar := drawBar(u.MemUsedMB, max, 30)
		lines = append(lines, fmt.Sprintf("%-12s %s %5.0f", u.User, bar, u.MemUsedMB))
	}
	content := strings.Join(lines, "\n")
	return box.Width(m.width/2 - 4).Render(content)
}

func (m model) renderProcs() string {
	var b strings.Builder
	b.WriteString(label.Render("Top GPU processes (by used MB)") + "\n")
	if len(m.curr.Procs) == 0 {
		b.WriteString(subtle.Render("none"))
		return box.Width(m.width - 4).Render(b.String())
	}
	procs := append([]types.GPUProcess(nil), m.curr.Procs...)
	sort.Slice(procs, func(i, j int) bool { return procs[i].UsedMemMB > procs[j].UsedMemMB })
	maxN := 10
	if len(procs) < maxN { maxN = len(procs) }
	for i := 0; i < maxN; i++ {
		p := procs[i]
		b.WriteString(fmt.Sprintf("%5d  %-12s  %-22s  %6.0f MB  %s\n", p.PID, p.User, trim(p.ProcessName, 22), p.UsedMemMB, shortUUID(p.GPUUUID)))
	}
	return box.Width(m.width - 4).Render(b.String())
}

func (m model) renderHelp() string {
	if !m.showHelp {
		return subtle.Render("a: auto | r: refresh | s: save | h: history | ←/→: prev/next snap | ↑/↓: day | t: today | ?: help | q: quit")
	}
	return box.Width(m.width - 4).Render(strings.Join([]string{
		"a — Toggle auto-recording of live samples (every 5s)",
		"r — Refresh once (live mode)",
		"s — Save the current snapshot immediately",
		"h — Toggle History mode",
		"←/→ — Previous/Next snapshot of the selected date",
		"↑/↓ — Move one day back/forward",
		"t — Jump back to today and live mode",
		"q — Quit",
	}, "\n"))
}

func drawBar(value, max float64, width int) string {
	if max <= 0 { max = 1 }
	ratio := value / max
	if ratio < 0 { ratio = 0 }
	if ratio > 1 { ratio = 1 }
	filled := int(float64(width) * ratio)
	if filled < 0 { filled = 0 }
	if filled > width { filled = width }
	return bar.Width(filled).Render(strings.Repeat(" ", filled)) + subtle.Width(width-filled).Render(strings.Repeat("·", width-filled))
}

func trim(s string, n int) string {
	r := []rune(s)
	if len(r) <= n { return s }
	return string(r[:n-1]) + "…"
}

func shortUUID(u string) string {
	if len(u) <= 8 { return u }
	return u[len(u)-8:]
}
