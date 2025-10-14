package tui

import lg "github.com/charmbracelet/lipgloss"

var (
	accent     = lg.AdaptiveColor{Light: "#0F62FE", Dark: "#64B5F6"}
	danger     = lg.AdaptiveColor{Light: "#DA1E28", Dark: "#EF5350"}
	muted      = lg.AdaptiveColor{Light: "#7D7D7D", Dark: "#9E9E9E"}
	ok         = lg.AdaptiveColor{Light: "#198038", Dark: "#66BB6A"}

	headStyle  = lg.NewStyle().Bold(true).Foreground(accent).Padding(0,1)
	subtle     = lg.NewStyle().Foreground(muted)
	box        = lg.NewStyle().Border(lg.RoundedBorder()).BorderForeground(lg.Color("240")).Padding(1,2).Margin(0,1)
	bar        = lg.NewStyle().Background(accent).Foreground(lg.Color("0")).Bold(true)
	label      = lg.NewStyle().Bold(true)
	value      = lg.NewStyle().Faint(false)
	errStyle   = lg.NewStyle().Foreground(danger).Bold(true)
	okStyle    = lg.NewStyle().Foreground(ok).Bold(true)
)
