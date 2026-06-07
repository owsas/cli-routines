package tui

import "github.com/charmbracelet/bubbles/key"

// Key bindings used throughout the TUI.
type keyMap struct {
	Up          key.Binding
	Down        key.Binding
	Enter       key.Binding
	Escape      key.Binding
	Quit        key.Binding
	Help        key.Binding

	// Dashboard-specific
	Edit        key.Binding
	Delete      key.Binding
	Run         key.Binding
	Toggle      key.Binding
	New         key.Binding
	Logs        key.Binding
	StartDaemon key.Binding
	StopDaemon  key.Binding
	Refresh     key.Binding

	// Editor-specific
	Save     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding

	// Runner / Logs
	Back key.Binding
}

// FullKeyMap returns the complete set of key bindings.
func FullKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit routine"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete routine"),
		),
		Run: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "run routine"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle enabled"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new routine"),
		),
		Logs: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "view logs"),
		),
		StartDaemon: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "start daemon"),
		),
		StopDaemon: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "stop daemon"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev field"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

// ShortHelp returns bindings to show in the footer.
func (km keyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		km.Up, km.Down, km.Enter, km.Escape, km.Quit,
	}
}

// DashboardHelp returns the bindings for the dashboard screen.
func DashboardHelp() []key.Binding {
	km := FullKeyMap()
	return []key.Binding{
		km.Up, km.Down, km.Enter,
		km.Edit, km.Delete, km.Run, km.Toggle, km.New,
		km.Logs, km.StartDaemon, km.StopDaemon,
		km.Refresh, km.Help, km.Quit,
	}
}

// EditorHelp returns the bindings for the editor screen.
func EditorHelp() []key.Binding {
	km := FullKeyMap()
	return []key.Binding{
		km.Tab, km.ShiftTab, km.Enter,
		km.Save, km.Escape,
	}
}
