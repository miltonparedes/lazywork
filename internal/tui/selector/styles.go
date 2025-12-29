package selector

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/miltonparedes/lazywork/internal/tui"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(tui.ColorPrimary)

	selectedStyle = lipgloss.NewStyle().
			Foreground(tui.ColorSelected).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(tui.ColorSelected)

	normalStyle = lipgloss.NewStyle().
			Foreground(tui.ColorNormal)

	dimStyle = lipgloss.NewStyle().
			Foreground(tui.ColorDim)

	currentMarker = lipgloss.NewStyle().
			Foreground(tui.ColorAccent).
			SetString(" ← current")

	helpStyle = lipgloss.NewStyle().
			Foreground(tui.ColorDim).
			Italic(true)
)
