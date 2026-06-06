package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"cli-routines/core"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Screens ---

type screen int

const (
	screenDashboard screen = iota
	screenEditor
	screenRunner
	screenLogViewer
	screenHelp
)

// --- Editor field types ---

type fieldKind int

const (
	fieldText fieldKind = iota
	fieldTextArea
	fieldBool
	fieldSelect
)

// editorField describes one field in the editor form.
type editorField struct {
	label    string
	kind     fieldKind
	value    *string  // pointer to the model's field value
	checked  *bool    // for bool fields
	options  []string // for select fields
	selected *int     // for select fields

	// Which executor types this field applies to (empty = always)
	appliesTo []string

	// textinput model (nil for non-text fields)
	ti     *textinput.Model
	focused bool
}

// --- Model ---

type model struct {
	screen    screen
	keys      keyMap
	help      help.Model
	spinner   spinner.Model
	spinning  bool
	width     int
	height    int
	err       error
	statusMsg string

	// Config
	config    *core.Config
	configErr error

	// Daemon status (refreshed periodically)
	daemonRunning bool
	daemonPID     int
	daemonChecked bool

	// Dashboard table
	table      table.Model

	// Editor
	editIdx     int // index in config.Routines (-1 = new)
	editFields  []editorField
	editFocus   int
	editDirty   bool

	// Editor field values (pointed to by editorField.value)
	editName, editDesc, editSchedule, editFolder string
	editEnabled, editNotify                     bool
	editExecutorType                            string
	editCommand                                 string
	editPrompt                                  string
	editModel                                   string
	editDangerouslySkipPerm                     bool
	editPermSelected                            int
	editExecSelected                            int

	// Runner
	runName   string
	runOutput string
	runErr    error
	runDone   bool
	runVP     viewport.Model
	runVPReady bool

	// Log viewer
	logVP      viewport.Model
	logVPReady bool
	logContent string
	logErr     error

	// Async result channel
	resultCh chan tea.Msg
}

// --- Messages ---

type configLoadedMsg struct {
	cfg *core.Config
	err error
}

type daemonStatusMsg struct {
	running bool
	pid     int
}

type routineFinishedMsg struct {
	name   string
	output string
	err    error
}

type logLoadedMsg struct {
	content string
	err     error
}

type errMsg struct {
	err error
}

// NewModel creates the initial model.
func NewModel() (*model, error) {
	ti := textinput.New()
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(colPrimary)

	m := &model{
		screen:   screenDashboard,
		keys:     FullKeyMap(),
		help:     help.New(),
		spinner:  spinner.New(spinner.WithStyle(lipgloss.NewStyle().Foreground(colSecondary))),
		resultCh: make(chan tea.Msg, 10),
	}

	// Initialize table
	m.initTable()

	return m, nil
}

func (m *model) initTable() {
	columns := []table.Column{
		{Title: "#", Width: 3},
		{Title: "Name", Width: 22},
		{Title: "Schedule", Width: 14},
		{Title: "Executor", Width: 12},
		{Title: "Notify", Width: 7},
		{Title: "Enabled", Width: 8},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colBorder).
		BorderBottom(true).
		Bold(false).
		Foreground(colSecondary).
		Padding(0, 1)

	s.Selected = s.Selected.
		Foreground(colText).
		Background(colPrimary).
		Bold(false)

	s.Cell = s.Cell.
		Padding(0, 1)

	t.SetStyles(s)
	m.table = t
}

// --- Init ---

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.loadConfig(),
		m.checkDaemon(),
		m.spinner.Tick,
		m.listenForResult(),
	)
}

// listenForResult returns a command that waits for async results.
func (m *model) listenForResult() tea.Cmd {
	return func() tea.Msg {
		return <-m.resultCh
	}
}

func (m *model) loadConfig() tea.Cmd {
	return func() tea.Msg {
		cfg, err := core.LoadConfig()
		return configLoadedMsg{cfg: cfg, err: err}
	}
}

func (m *model) checkDaemon() tea.Cmd {
	return func() tea.Msg {
		running := core.IsDaemonRunning()
		pid := 0
		if running {
			p, err := core.ReadPid()
			if err == nil {
				pid = p
			}
		}
		return daemonStatusMsg{running: running, pid: pid}
	}
}

// --- Update ---

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update table height
		m.table.SetHeight(max(5, msg.Height-14))

		// Update viewports
		if m.runVPReady {
			m.runVP.Width = min(msg.Width-6, 120)
			m.runVP.Height = msg.Height - 10
		}
		if m.logVPReady {
			m.logVP.Width = min(msg.Width-6, 120)
			m.logVP.Height = msg.Height - 10
		}

	case configLoadedMsg:
		if msg.err != nil {
			m.configErr = msg.err
			m.config = &core.Config{}
		} else {
			m.config = msg.cfg
			if m.config == nil {
				m.config = &core.Config{}
			}
		}
		m.refreshTable()

	case daemonStatusMsg:
		m.daemonRunning = msg.running
		m.daemonPID = msg.pid
		m.daemonChecked = true

	case routineFinishedMsg:
		m.runDone = true
		m.runOutput = msg.output
		m.runErr = msg.err
		if m.runVPReady {
			m.runVP.SetContent(m.formatOutput(msg.output, msg.err))
			m.runVP.GotoTop()
		}
		cmds = append(cmds, m.listenForResult())

	case logLoadedMsg:
		m.logContent = msg.content
		m.logErr = msg.err
		if m.logVPReady && len(msg.content) > 0 {
			m.logVP.SetContent(msg.content)
			m.logVP.GotoTop()
		}
		cmds = append(cmds, m.listenForResult())

	case spinner.TickMsg:
		if m.spinning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case errMsg:
		m.err = msg.err
	}

	// Handle input based on current screen
	switch m.screen {
	case screenDashboard:
		cmd := m.updateDashboard(msg)
		cmds = append(cmds, cmd)
	case screenEditor:
		cmd := m.updateEditor(msg)
		cmds = append(cmds, cmd)
	case screenRunner:
		cmd := m.updateRunner(msg)
		cmds = append(cmds, cmd)
	case screenLogViewer:
		cmd := m.updateLogViewer(msg)
		cmds = append(cmds, cmd)
	case screenHelp:
		m.updateHelp(msg)
	}

	return m, tea.Batch(cmds...)
}

// --- Dashboard Update ---

func (m *model) updateDashboard(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.screen = screenHelp

		case key.Matches(msg, m.keys.Enter):
			if len(m.config.Routines) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.config.Routines) {
					m.startEditing(idx)
				}
			}

		case key.Matches(msg, m.keys.Edit):
			if len(m.config.Routines) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.config.Routines) {
					m.startEditing(idx)
				}
			}

		case key.Matches(msg, m.keys.New):
			m.startEditing(-1)

		case key.Matches(msg, m.keys.Delete):
			if len(m.config.Routines) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.config.Routines) {
					m.deleteRoutine(idx)
				}
			}

		case key.Matches(msg, m.keys.Toggle):
			if len(m.config.Routines) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.config.Routines) {
					m.config.Routines[idx].Enabled = !m.config.Routines[idx].Enabled
					core.SaveConfig(m.config)
					m.refreshTable()
				}
			}

		case key.Matches(msg, m.keys.Run):
			if len(m.config.Routines) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.config.Routines) {
					m.startRoutine(idx)
				}
			}

		case key.Matches(msg, m.keys.Logs):
			m.startLogViewer()

		case key.Matches(msg, m.keys.StartDaemon):
			return m.startDaemonCmd()

		case key.Matches(msg, m.keys.StopDaemon):
			return m.stopDaemonCmd()

		case key.Matches(msg, m.keys.Refresh):
			return tea.Batch(m.loadConfig(), m.checkDaemon())
		}

		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return cmd
	}

	return nil
}

// --- Editor Update ---

func (m *model) updateEditor(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Escape):
			if m.editDirty {
				// TODO: confirm discard
			}
			m.screen = screenDashboard
			m.refreshTable()
			return nil

		case key.Matches(msg, m.keys.Save):
			return m.saveRoutine()

		case key.Matches(msg, m.keys.Tab):
			m.focusNextField(1)

		case key.Matches(msg, m.keys.ShiftTab):
			m.focusNextField(-1)

		case key.Matches(msg, m.keys.Enter):
			// For select fields, cycle through options
			f := &m.editFields[m.editFocus]
			if f.kind == fieldSelect {
				if f.selected != nil {
					*f.selected = (*f.selected + 1) % len(f.options)
				}
				if f.label == "Executor Type" {
					m.onExecutorTypeChanged()
				}
				m.editDirty = true
			} else if f.kind == fieldBool {
				if f.checked != nil {
					*f.checked = !*f.checked
				}
				m.editDirty = true
			}

		case key.Matches(msg, m.keys.Down):
			m.focusNextField(1)

		case key.Matches(msg, m.keys.Up):
			m.focusNextField(-1)
		}
	}

	// Update focused field
	for i := range m.editFields {
		if i == m.editFocus {
			m.editFields[i].focused = true
		} else {
			m.editFields[i].focused = false
		}

		f := &m.editFields[i]
		if f.ti != nil {
			var cmd tea.Cmd
			*f.ti, cmd = f.ti.Update(msg)
			if cmd != nil {
				// value changed
				if f.value != nil {
					*f.value = f.ti.Value()
					m.editDirty = true
				}
				return cmd
			}
		}
	}

	return nil
}

func (m *model) focusNextField(dir int) {
	total := len(m.editFields)
	if total == 0 {
		return
	}

	next := m.editFocus
	for i := 0; i < total; i++ {
		next = (next + dir + total) % total
		if !m.editFields[next].applies() {
			continue
		}
		// For text fields, blur the current and focus the next
		if m.editFields[m.editFocus].ti != nil {
			m.editFields[m.editFocus].ti.Blur()
		}
		m.editFocus = next
		if m.editFields[next].ti != nil {
			m.editFields[next].ti.Focus()
		}
		return
	}
}

func (f *editorField) applies() bool {
	if len(f.appliesTo) == 0 {
		return true
	}
	// The parent model tracks executor type; we check at render time
	return true
}

func (m *model) setEditorFieldsVisible() {
	for i := range m.editFields {
		f := &m.editFields[i]
		if len(f.appliesTo) == 0 {
			f.appliesTo = nil
		}
	}
}

// onExecutorTypeChanged updates the dynamic fields when executor type changes.
func (m *model) onExecutorTypeChanged() {
	m.rebuildDynamicFields()
}

func (m *model) rebuildDynamicFields() {
	// Keep the first 8 fields (name through executor type), rebuild the rest
	base := 8
	if base > len(m.editFields) {
		base = len(m.editFields)
	}

	static := m.editFields[:min(base, len(m.editFields))]
	dynamic := m.buildDynamicFields(m.editExecutorType)

	m.editFields = make([]editorField, 0, len(static)+len(dynamic))
	m.editFields = append(m.editFields, static...)
	m.editFields = append(m.editFields, dynamic...)

	// Adjust focus if out of bounds
	if m.editFocus >= len(m.editFields) {
		m.editFocus = len(m.editFields) - 1
	}
}

// --- Runner Update ---

func (m *model) updateRunner(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back), key.Matches(msg, m.keys.Escape):
			m.screen = screenDashboard
			m.runDone = false
			m.runErr = nil
			m.runOutput = ""
		}
	}

	if m.runVPReady {
		var cmd tea.Cmd
		m.runVP, cmd = m.runVP.Update(msg)
		return cmd
	}
	return nil
}

// --- Log Viewer Update ---

func (m *model) updateLogViewer(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back), key.Matches(msg, m.keys.Escape):
			m.screen = screenDashboard
		case key.Matches(msg, m.keys.Refresh):
			return m.loadLogs()
		}
	}

	if m.logVPReady {
		var cmd tea.Cmd
		m.logVP, cmd = m.logVP.Update(msg)
		return cmd
	}
	return nil
}

// --- Help Update ---

func (m *model) updateHelp(msg tea.Msg) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Escape), key.Matches(msg, m.keys.Quit), key.Matches(msg, m.keys.Help):
			m.screen = screenDashboard
		}
	}
}

// --- Actions ---

func (m *model) startEditing(idx int) {
	m.screen = screenEditor
	m.editIdx = idx
	m.editDirty = false
	m.editFocus = 0

	if idx >= 0 && idx < len(m.config.Routines) {
		r := m.config.Routines[idx]
		m.editName = r.Name
		m.editDesc = r.Description
		m.editSchedule = r.Schedule
		m.editFolder = r.Folder
		m.editEnabled = r.Enabled
		m.editNotify = r.Notify
		m.editExecutorType = r.ExecutorType()

		// Parse executor-specific fields
		m.editCommand = ""
		m.editPrompt = ""
		m.editModel = ""
		m.editDangerouslySkipPerm = false

		switch m.editExecutorType {
		case "shell":
			var e struct {
				Command string `json:"command"`
			}
			if err := json.Unmarshal(r.ExecutorRaw, &e); err == nil {
				m.editCommand = e.Command
			}
		case "opencode":
			var e struct {
				Prompt string `json:"prompt"`
				Model  string `json:"model"`
			}
			if err := json.Unmarshal(r.ExecutorRaw, &e); err == nil {
				m.editPrompt = e.Prompt
				m.editModel = e.Model
			}
		case "claude":
			var e struct {
				Prompt                     string `json:"prompt"`
				Model                      string `json:"model"`
				PermissionMode             string `json:"permissionMode"`
				DangerouslySkipPermissions bool   `json:"dangerouslySkipPermissions"`
			}
			if err := json.Unmarshal(r.ExecutorRaw, &e); err == nil {
				m.editPrompt = e.Prompt
				m.editModel = e.Model
				m.editPermSelected = permModeToIndex(e.PermissionMode)
				m.editDangerouslySkipPerm = e.DangerouslySkipPermissions
			}
		}
	} else {
		// New routine defaults
		m.editName = ""
		m.editDesc = ""
		m.editSchedule = "0 9 * * 1-5"
		m.editFolder = os.Getenv("HOME")
		m.editEnabled = false
		m.editNotify = true
		m.editExecutorType = "shell"
		m.editCommand = ""
		m.editPrompt = ""
		m.editModel = ""
		m.editDangerouslySkipPerm = false
	}

	m.editExecSelected = 0
	switch m.editExecutorType {
	case "shell":
		m.editExecSelected = 0
	case "opencode":
		m.editExecSelected = 1
	case "claude":
		m.editExecSelected = 2
	}
	m.buildEditorFields()
}

func (m *model) buildEditorFields() {
	execTypes := []string{"shell", "opencode", "claude"}

	fields := []editorField{
		{label: "Name", kind: fieldText, value: &m.editName},
		{label: "Description", kind: fieldText, value: &m.editDesc},
		{label: "Schedule", kind: fieldText, value: &m.editSchedule},
		{label: "Folder", kind: fieldText, value: &m.editFolder},
		{label: "Enabled", kind: fieldBool, checked: &m.editEnabled},
		{label: "Notify", kind: fieldBool, checked: &m.editNotify},
		{label: "", kind: fieldText, value: nil}, // separator
		{label: "Executor Type", kind: fieldSelect, options: execTypes, selected: &m.editExecSelected},
	}

	// Build text inputs for text fields
	for i := range fields {
		f := &fields[i]
		if f.kind == fieldText && f.value != nil {
			ti := textinput.New()
			ti.Prompt = ""
			ti.Placeholder = f.label
			ti.SetValue(*f.value)
			ti.CharLimit = 0
			f.ti = &ti
		}
	}

	// Add dynamic fields based on executor type
	dynamic := m.buildDynamicFields(m.editExecutorType)
	fields = append(fields, dynamic...)

	m.editFields = fields

	// Focus first field
	m.editFocus = 0
	if len(fields) > 0 && fields[0].ti != nil {
		fields[0].ti.Focus()
	}
}

func (m *model) buildDynamicFields(execType string) []editorField {
	var fields []editorField

	switch execType {
	case "shell":
		fields = []editorField{
			{label: "Command", kind: fieldTextArea, value: &m.editCommand},
		}
	case "opencode":
		fields = []editorField{
			{label: "Prompt", kind: fieldTextArea, value: &m.editPrompt},
			{label: "Model", kind: fieldText, value: &m.editModel},
		}
	case "claude":
		pModes := []string{"", "default", "acceptEdits", "plan", "bypassPermissions"}
		fields = []editorField{
			{label: "Prompt", kind: fieldTextArea, value: &m.editPrompt},
			{label: "Model", kind: fieldText, value: &m.editModel},
			{label: "Permission Mode", kind: fieldSelect, options: pModes, selected: &m.editPermSelected},
			{label: "Skip Permissions", kind: fieldBool, checked: &m.editDangerouslySkipPerm},
		}
	default:
		return nil
	}

	// Initialize text inputs for text fields
	for i := range fields {
		f := &fields[i]
		if f.kind == fieldText && f.value != nil {
			ti := textinput.New()
			ti.Prompt = ""
			ti.Placeholder = f.label
			ti.SetValue(*f.value)
			ti.CharLimit = 0
			f.ti = &ti
		}
	}

	return fields
}

func (m *model) saveRoutine() tea.Cmd {
	// Update values from text inputs
	m.syncFieldValues()

	// Get executor type string
	execType := "shell"
	if m.editExecSelected >= 0 && m.editExecSelected < 3 {
		switch m.editExecSelected {
		case 0:
			execType = "shell"
		case 1:
			execType = "opencode"
		case 2:
			execType = "claude"
		}
	}

	// Build executor JSON
	var executorJSON json.RawMessage
	switch execType {
	case "shell":
		data, _ := json.Marshal(map[string]string{
			"type":    "shell",
			"command": m.editCommand,
		})
		executorJSON = data
	case "opencode":
		data, _ := json.Marshal(map[string]string{
			"type":   "opencode",
			"prompt": m.editPrompt,
			"model":  m.editModel,
		})
		executorJSON = data
	case "claude":
		permMode := ""
		if m.editPermSelected >= 0 && m.editPermSelected < len(permModeOptions) {
			permMode = permModeOptions[m.editPermSelected]
		}
		data, _ := json.Marshal(map[string]interface{}{
			"type":                      "claude",
			"prompt":                    m.editPrompt,
			"model":                     m.editModel,
			"permissionMode":            permMode,
			"dangerouslySkipPermissions": m.editDangerouslySkipPerm,
		})
		executorJSON = data
	}

	r := core.Routine{
		Name:        m.editName,
		Description: m.editDesc,
		Schedule:    m.editSchedule,
		Folder:      m.editFolder,
		ExecutorRaw: executorJSON,
		Enabled:     m.editEnabled,
		Notify:      m.editNotify,
	}

	if m.editIdx >= 0 && m.editIdx < len(m.config.Routines) {
		m.config.Routines[m.editIdx] = r
	} else {
		m.config.Routines = append(m.config.Routines, r)
	}

	m.editDirty = false
	return m.saveConfigAndRefresh()
}

func (m *model) syncFieldValues() {
	for i := range m.editFields {
		f := &m.editFields[i]
		if f.ti != nil && f.value != nil {
			*f.value = f.ti.Value()
		}
	}
}

func (m *model) deleteRoutine(idx int) {
	if idx < 0 || idx >= len(m.config.Routines) {
		return
	}
	m.config.Routines = append(m.config.Routines[:idx], m.config.Routines[idx+1:]...)
	// Save synchronously since this is called from a key handler
	core.SaveConfig(m.config)
	cfg, _ := core.LoadConfig()
	if cfg != nil {
		m.config = cfg
	}
	m.refreshTable()
}

func (m *model) saveConfigAndRefresh() tea.Cmd {
	return func() tea.Msg {
		err := core.SaveConfig(m.config)
		if err != nil {
			return errMsg{err: err}
		}
		// Reload config to refresh state
		cfg, err := core.LoadConfig()
		return configLoadedMsg{cfg: cfg, err: err}
	}
}

func (m *model) startRoutine(idx int) {
	if idx < 0 || idx >= len(m.config.Routines) {
		return
	}

	r := m.config.Routines[idx]
	m.screen = screenRunner
	m.runName = r.Name
	m.runOutput = ""
	m.runErr = nil
	m.runDone = false

	// Initialize viewport
	if !m.runVPReady {
		m.runVP = viewport.New(min(m.width-6, 120), m.height-10)
		m.runVPReady = true
	}
	m.runVP.SetContent("Starting routine...\n")

	// Run in background and send result through channel
	go func() {
		output, err := core.Execute(r)
		m.resultCh <- routineFinishedMsg{name: r.Name, output: output, err: err}
	}()
}

func (m *model) startDaemonCmd() tea.Cmd {
	return func() tea.Msg {
		// Check if already running
		if core.IsDaemonRunning() {
			pid, _ := core.ReadPid()
			return errMsg{err: fmt.Errorf("daemon already running (PID %d)", pid)}
		}

		cfg, err := core.LoadConfig()
		if err != nil {
			return errMsg{err: fmt.Errorf("cannot load config: %w", err)}
		}

		enabled := 0
		for _, r := range cfg.Routines {
			if r.Enabled {
				enabled++
			}
		}
		if enabled == 0 {
			return errMsg{err: fmt.Errorf("no enabled routines found")}
		}

		// Use the daemon command via subprocess
		// This re-executes the routines binary with start --foreground
		// But from the TUI we can't easily do this. Let's use the scheduler directly.
		return m.runDaemonInProcess(cfg)
	}
}

func (m *model) runDaemonInProcess(cfg *core.Config) tea.Cmd {
	scheduler := core.NewScheduler()
	if err := scheduler.Start(cfg); err != nil {
		return func() tea.Msg {
			return errMsg{err: fmt.Errorf("cannot start scheduler: %w", err)}
		}
	}

	core.WritePid(os.Getpid())
	core.AppendLog(fmt.Sprintf("Daemon started from TUI with PID %d", os.Getpid()))

	return func() tea.Msg {
		return daemonStatusMsg{running: true, pid: os.Getpid()}
	}
}

func (m *model) stopDaemonCmd() tea.Cmd {
	return func() tea.Msg {
		if !core.IsDaemonRunning() {
			return daemonStatusMsg{running: false, pid: 0}
		}
		err := core.KillDaemon()
		if err != nil {
			return errMsg{err: err}
		}
		return daemonStatusMsg{running: false, pid: 0}
	}
}

func (m *model) startLogViewer() {
	m.screen = screenLogViewer
	if !m.logVPReady {
		m.logVP = viewport.New(min(m.width-6, 120), m.height-10)
		m.logVPReady = true
	}
	m.logContent = ""
	m.logErr = nil
	m.logVP.SetContent("Loading logs...")

	// Load logs asynchronously
	go func() {
		path, err := core.LogPath()
		if err != nil {
			m.resultCh <- logLoadedMsg{err: err}
			return
		}
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				m.resultCh <- logLoadedMsg{content: "No log file found. Run a routine first."}
			} else {
				m.resultCh <- logLoadedMsg{err: err}
			}
			return
		}
		content := string(data)
		if content == "" {
			content = "No log entries yet."
		}
		m.resultCh <- logLoadedMsg{content: content}
	}()
}

func (m *model) loadLogs() tea.Cmd {
	return func() tea.Msg {
		path, err := core.LogPath()
		if err != nil {
			return logLoadedMsg{err: err}
		}
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return logLoadedMsg{content: "No log file found. Run a routine first."}
			}
			return logLoadedMsg{err: err}
		}
		content := string(data)
		if content == "" {
			content = "No log entries yet."
		}
		return logLoadedMsg{content: content}
	}
}

// --- Table Refresh ---

func (m *model) refreshTable() {
	rows := make([]table.Row, 0, len(m.config.Routines))
	for i, r := range m.config.Routines {
		enabled := "✗"
		if r.Enabled {
			enabled = "✓"
		}
		notify := "✗"
		if r.Notify {
			notify = "✓"
		}
		execType := r.ExecutorType()

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", i+1),
			truncate(r.Name, 20),
			r.Schedule,
			execType,
			notify,
			enabled,
		})
	}
	m.table.SetRows(rows)
}

func (m *model) formatOutput(output string, err error) string {
	var b strings.Builder
	if err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v\n\n", err)))
	}
	if output != "" {
		b.WriteString(output)
	}
	if b.Len() == 0 {
		b.WriteString("No output.")
	}
	return b.String()
}

// --- View ---

func (m *model) View() string {
	switch m.screen {
	case screenDashboard:
		return m.viewDashboard()
	case screenEditor:
		return m.viewEditor()
	case screenRunner:
		return m.viewRunner()
	case screenLogViewer:
		return m.viewLogViewer()
	case screenHelp:
		return m.viewHelp()
	default:
		return "unknown screen"
	}
}

func (m *model) viewDashboard() string {
	var b strings.Builder

	// Header
	b.WriteString(headerStyle.Render("⚡ cli-routines"))
	b.WriteString("\n")

	// Daemon status
	daemonStatus := statusStopped.Render("● STOPPED")
	if m.daemonRunning {
		daemonStatus = statusRunning.Render(fmt.Sprintf("● RUNNING (PID %d)", m.daemonPID))
	}
	b.WriteString(fmt.Sprintf("  Daemon: %s\n\n", daemonStatus))

	// Error display
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v\n", m.err)))
		m.err = nil
	}
	if m.configErr != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Config error: %v\n", m.configErr)))
	}

	// Routine count
	total := len(m.config.Routines)
	enabled := 0
	for _, r := range m.config.Routines {
		if r.Enabled {
			enabled++
		}
	}
	b.WriteString(fmt.Sprintf("  %d routine(s), %d enabled\n\n", total, enabled))

	// Table
	if total == 0 {
		b.WriteString("  No routines configured. Press ")
		b.WriteString(helpKey.Render("n"))
		b.WriteString(" to create one.\n\n")
	} else {
		b.WriteString(m.table.View())
		b.WriteString("\n")
	}

	// Help footer
	footerStyle := lipgloss.NewStyle().
		Foreground(colDim).
		MarginTop(1).
		BorderTop(true).
		BorderForeground(colBorder).
		PaddingTop(1)

	b.WriteString(footerStyle.Render(
		helpKey.Render("↑/↓") + " " + helpDesc.Render("navigate") + "  " +
			helpKey.Render("Enter/e") + " " + helpDesc.Render("edit") + "  " +
			helpKey.Render("n") + " " + helpDesc.Render("new") + "  " +
			helpKey.Render("d") + " " + helpDesc.Render("delete") + "  " +
			helpKey.Render("Space") + " " + helpDesc.Render("toggle") + "  " +
			helpKey.Render("r") + " " + helpDesc.Render("run") + "  " +
			helpKey.Render("l") + " " + helpDesc.Render("logs") + "  " +
			helpKey.Render("S/T") + " " + helpDesc.Render("start/stop daemon") + "  " +
			helpKey.Render("q") + " " + helpDesc.Render("quit"),
	))

	return appStyle.Render(b.String())
}

func (m *model) viewEditor() string {
	var b strings.Builder

	// Header
	title := "New Routine"
	if m.editIdx >= 0 {
		title = fmt.Sprintf("Editing: %s", m.editName)
	}
	b.WriteString(headerStyle.Render(fmt.Sprintf("✏️  %s", title)))
	b.WriteString("\n")

	// Form fields
	for _, f := range m.editFields {
		// Skip separators
		if f.label == "" && f.kind == fieldText && f.value == nil {
			b.WriteString("\n")
			continue
		}

		label := formLabel.Render(f.label + ":")

		switch f.kind {
		case fieldText:
			if f.ti != nil {
				if f.focused {
					f.ti.Prompt = "▸ "
					b.WriteString(label + formInputFocused.Render(f.ti.View()) + "\n")
				} else {
					f.ti.Prompt = "  "
					b.WriteString(label + formInput.Render(f.ti.View()) + "\n")
				}
			}

		case fieldTextArea:
			val := ""
			if f.value != nil {
				val = *f.value
			}
			style := formInput
			if f.focused {
				style = formInputFocused
			}
			// Show first few lines of the text area
			displayVal := val
			if len(displayVal) > 120 {
				displayVal = displayVal[:117] + "..."
			}
			displayVal = strings.ReplaceAll(displayVal, "\n", "↵ ")
			b.WriteString(label + style.Render(displayVal) + "\n")

		case fieldBool:
			status := "○"
			if f.checked != nil && *f.checked {
				status = "●"
			}
			style := lipgloss.NewStyle().Foreground(colMuted)
			if f.focused {
				style = lipgloss.NewStyle().Foreground(colPrimary).Bold(true)
			}
			b.WriteString(label + style.Render(status+" "+labelValue(f.label)) + "\n")

		case fieldSelect:
			var displayOpts []string
			for j, opt := range f.options {
				optDisplay := opt
				if optDisplay == "" {
					optDisplay = "(default)"
				}
				if f.selected != nil && j == *f.selected {
					displayOpts = append(displayOpts, radioSelected.Render()+" "+optDisplay)
				} else {
					displayOpts = append(displayOpts, radioUnselected.Render()+" "+optDisplay)
				}
			}
			style := lipgloss.NewStyle().Foreground(colMuted)
			if f.focused {
				style = lipgloss.NewStyle().Foreground(colPrimary)
			}
			b.WriteString(label + style.Render(strings.Join(displayOpts, "  ")) + "\n")
		}
	}

	// Hint
	b.WriteString("\n" + formHelp.Render(
		"Tab/↑↓ navigate  Enter toggle  Ctrl+S save  Esc cancel",
	))

	b.WriteString("\n")

	// Status bar
	if m.statusMsg != "" {
		b.WriteString(successStyle.Render(m.statusMsg))
	}
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	return appStyle.Render(b.String())
}

func (m *model) viewRunner() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("▶ Running: %s", m.runName)))
	b.WriteString("\n")

	if !m.runDone {
		b.WriteString(m.spinner.View() + " Running...\n")
	}

	if m.runVPReady {
		b.WriteString(outputStyle.Render(m.runVP.View()))
		b.WriteString("\n")
	}

	// Footer
	b.WriteString(footerStyle.Render(
		helpKey.Render("Esc") + " " + helpDesc.Render("back to dashboard"),
	))

	return appStyle.Render(b.String())
}

func (m *model) viewLogViewer() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("📋 Routine Logs"))
	b.WriteString("\n")

	if m.logErr != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error loading logs: %v\n", m.logErr)))
	} else if len(m.logContent) == 0 {
		b.WriteString(subtitleStyle.Render("Loading..."))
	} else if m.logVPReady {
		b.WriteString(outputStyle.Render(m.logVP.View()))
		b.WriteString("\n")
	}

	b.WriteString(footerStyle.Render(
		helpKey.Render("Esc") + " " + helpDesc.Render("back") + "  " +
			helpKey.Render("Ctrl+R") + " " + helpDesc.Render("refresh"),
	))

	return appStyle.Render(b.String())
}

func (m *model) viewHelp() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("⌨️  Keyboard Shortcuts"))
	b.WriteString("\n")

	items := []struct {
		key, desc string
	}{
		{"↑/k, ↓/j", "Navigate"},
		{"Enter", "Select / edit"},
		{"Space", "Toggle enabled"},
		{"e", "Edit selected routine"},
		{"n", "New routine"},
		{"d", "Delete routine"},
		{"r", "Run routine"},
		{"l", "View logs"},
		{"S", "Start daemon"},
		{"T", "Stop daemon"},
		{"Tab, Shift+Tab", "Navigate form fields"},
		{"Ctrl+S", "Save changes"},
		{"Ctrl+R", "Refresh"},
		{"Esc", "Back"},
		{"? / q", "Help / Quit"},
	}

	for _, item := range items {
		b.WriteString(fmt.Sprintf("  %-20s %s\n",
			helpKey.Render(item.key),
			helpDesc.Render(item.desc),
		))
	}

	b.WriteString("\n" + footerStyle.Render(
		helpKey.Render("Esc")+" "+helpDesc.Render("back to dashboard"),
	))

	return appStyle.Render(b.String())
}

// --- Helpers ---

func labelValue(s string) string {
	if len(s) > 40 {
		return s[:37] + "..."
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s
}

var permModeOptions = []string{"", "default", "acceptEdits", "plan", "bypassPermissions"}

func permModeToIndex(mode string) int {
	for i, m := range permModeOptions {
		if m == mode {
			return i
		}
	}
	return 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
