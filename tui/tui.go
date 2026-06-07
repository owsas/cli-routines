package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the Bubble Tea TUI program.
func Run() error {
	m, err := NewModel()
	if err != nil {
		return fmt.Errorf("cannot initialize TUI: %w", err)
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}
