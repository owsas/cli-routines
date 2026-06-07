package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	colPrimary   = lipgloss.Color("#7C3AED") // purple
	colSecondary = lipgloss.Color("#A78BFA") // lighter purple
	colSuccess   = lipgloss.Color("#10B981") // green
	colDanger    = lipgloss.Color("#EF4444") // red
	colWarning   = lipgloss.Color("#F59E0B") // amber
	colMuted     = lipgloss.Color("#6B7280") // gray
	colText      = lipgloss.Color("#E5E7EB") // light text
	colDim       = lipgloss.Color("#9CA3AF") // dim text
	colBg        = lipgloss.Color("#1F2937") // dark bg
	colBgAlt     = lipgloss.Color("#111827") // darker bg
	colBorder    = lipgloss.Color("#374151") // border

	// Layout
	appStyle = lipgloss.NewStyle().
		Padding(1, 2).
		Background(colBg)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colPrimary).
		MarginBottom(1)

	headerAccent = lipgloss.NewStyle().
		Bold(true).
		Foreground(colSecondary)
	
	subtitleStyle = lipgloss.NewStyle().
		Foreground(colDim).
		MarginBottom(1)

	// Daemon status
	statusRunning = lipgloss.NewStyle().
			Foreground(colSuccess).
			Bold(true)

	statusStopped = lipgloss.NewStyle().
			Foreground(colDanger).
			Bold(true)

	// Table styles
	tableHeader = lipgloss.NewStyle().
			Foreground(colSecondary).
			Bold(true).
			Padding(0, 2)

	tableCell = lipgloss.NewStyle().
			Foreground(colText).
			Padding(0, 2)

	tableSelected = lipgloss.NewStyle().
			Foreground(colText).
			Background(colPrimary).
			Padding(0, 2)

	// Buttons
	btnPrimary = lipgloss.NewStyle().
			Background(colPrimary).
			Foreground(colText).
			Padding(0, 2).
			Bold(true)

	btnDanger = lipgloss.NewStyle().
			Background(colDanger).
			Foreground(colText).
			Padding(0, 2).
			Bold(true)

	btnOutline = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorder).
			Foreground(colText).
			Padding(0, 1)

	// Form styles
	formLabel = lipgloss.NewStyle().
			Foreground(colSecondary).
			Width(18).
			Align(lipgloss.Right).
			Padding(0, 1)

	formInput = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorder).
			Padding(0, 1).
			Width(50)

	formInputFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colPrimary).
				Padding(0, 1).
				Width(50)

	formHelp = lipgloss.NewStyle().
			Foreground(colDim).
			Italic(true).
			MarginTop(1)

	// Error
	errorStyle = lipgloss.NewStyle().
			Foreground(colDanger).
			Bold(true).
			MarginTop(1)

	successStyle = lipgloss.NewStyle().
			Foreground(colSuccess).
			MarginTop(1)

	// Viewport / output
	outputStyle = lipgloss.NewStyle().
			Background(colBgAlt).
			Foreground(colText).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colBorder)

	// Footer
	footerStyle = lipgloss.NewStyle().
			Foreground(colDim).
			MarginTop(1).
			BorderTop(true).
			BorderForeground(colBorder).
			PaddingTop(1)

	// Help
	helpKey = lipgloss.NewStyle().
			Foreground(colSecondary).
			Bold(true)

	helpDesc = lipgloss.NewStyle().
			Foreground(colDim)

	// Checkmark / cross
	checkMark = lipgloss.NewStyle().Foreground(colSuccess).SetString("✓")
	crossMark = lipgloss.NewStyle().Foreground(colDanger).SetString("✗")

	// Radio / select
	radioSelected = lipgloss.NewStyle().
			Foreground(colPrimary).
			Bold(true).
			SetString("●")

	radioUnselected = lipgloss.NewStyle().
			Foreground(colMuted).
			SetString("○")
)
