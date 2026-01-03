package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/entro314-labs/cool-kit/internal/config"
)

type ConfigState int

const (
	ConfigStateList ConfigState = iota
	ConfigStateEdit
)

type ConfigModel struct {
	state   ConfigState
	cfg     *config.Config
	keys    []string // Flattened keys for selection
	values  map[string]interface{}
	cursor  int
	width   int
	height  int
	editKey string
	editVal string
}

func NewConfigModel() (ConfigModel, error) {
	// Initialize config if needed
	if err := config.Initialize(); err != nil {
		return ConfigModel{}, err
	}
	cfg := config.Get()

	// Flatten config for display (simplified for demo)
	// In a real app, you might want to recurse or list specific manageable keys
	// Here we manually curated a list of editable settings

	values := make(map[string]interface{})
	keys := []string{
		"provider",
		"environment",
		"local.app_port",
		"local.websocket_port",
		"azure.location",
		"azure.vm_size",
	}

	values["provider"] = cfg.Provider
	values["environment"] = cfg.Environment
	values["local.app_port"] = cfg.Local.AppPort
	values["local.websocket_port"] = cfg.Local.WebSocketPort
	values["azure.location"] = cfg.Azure.Location
	values["azure.vm_size"] = cfg.Azure.VMSize

	return ConfigModel{
		state:  ConfigStateList,
		cfg:    cfg,
		keys:   keys,
		values: values,
		cursor: 0,
	}, nil
}

func (m ConfigModel) Init() tea.Cmd {
	return nil
}

func (m ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			if m.state == ConfigStateEdit {
				m.state = ConfigStateList
				return m, nil
			}
			return m, tea.Quit
		case tea.KeyUp:
			if m.state == ConfigStateList && m.cursor > 0 {
				m.cursor--
			}
		case tea.KeyDown:
			if m.state == ConfigStateList && m.cursor < len(m.keys)-1 {
				m.cursor++
			}
		case tea.KeyEnter:
			if m.state == ConfigStateList {
				m.state = ConfigStateEdit
				m.editKey = m.keys[m.cursor]
				m.editVal = fmt.Sprintf("%v", m.values[m.editKey])

				// Launch huh form for simple editing?
				// To keep it simple in this model, we'll just return a command to run an input form
				// But tea.Model can't easily spawn a blocking huh form and return.
				// We'll use a simple callback approach or just implement a text input here.
				// For "Sexy", let's use the Runner pattern again or just use huh inside the update loop?
				// Creating a new tea program inside Update is tricky.
				// simpler: just quit with a specific exit code/action and let the runner handle it?
				// OR: integrate bubbles/textinput.

				// Let's stick to list view for now, and rely on `config set` for edits,
				// or implement a text input within this model later.
				// For this iteration, let's treat "Enter" as "I want to edit this", quit, and let the caller handle it?
				// Or better: use `RunConfigMenu` to return the key to edit.
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m ConfigModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("⚙️ Configuration Manager"))
	s.WriteString("\n\n")

	for i, key := range m.keys {
		val := m.values[key]
		cursor := " "
		if m.cursor == i {
			cursor = "👉"
			s.WriteString(selectedStyle.Render(fmt.Sprintf("%s %-20s : %v", cursor, key, val)))
		} else {
			s.WriteString(fmt.Sprintf("%s %-20s : %v", cursor, key, val))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(descriptionStyle.Render("↑/↓: Navigate | Enter: Edit | Esc: Exit"))
	return s.String()
}

// RunConfigMenu runs the config menu and returns the selected key to edit (or empty if exit)
func RunConfigMenu() (string, error) {
	model, err := NewConfigModel()
	if err != nil {
		return "", err
	}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	m := finalModel.(ConfigModel)
	if m.state == ConfigStateEdit {
		return m.editKey, nil
	}
	return "", nil
}
