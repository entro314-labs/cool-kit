package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2).
			MarginTop(1).
			MarginBottom(1)

	// List styles
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	// Configuration styles
	inputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF75B7"))

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 3).
			MarginTop(1)

	activeButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Background(lipgloss.Color("#EE6FF8")).
				Padding(0, 3).
				MarginTop(1)
)

// State represents the current state of the UI
type State int

const (
	StateSelectProvider State = iota
	StateConfigureProvider
)

// Wrapper for list items
type providerItem struct {
	title string
	desc  string
	value string
	icon  string
}

func (i providerItem) Title() string       { return i.icon + " " + i.title }
func (i providerItem) Description() string { return i.desc }
func (i providerItem) FilterValue() string { return i.title }

// Model represents the main UI model
type Model struct {
	state         State
	Provider      string
	Configuration map[string]string // Collected configuration

	// Bubbles
	list       list.Model
	inputs     []textinput.Model
	focusIndex int

	width  int
	height int
}

// NewModel creates a new UI model
func NewModel() Model {
	items := []list.Item{
		providerItem{title: "Azure", desc: "Deploy to Microsoft Azure cloud", value: "azure", icon: "☁️"},
		providerItem{title: "Local", desc: "Set up local development environment", value: "local", icon: "🏠"},
		providerItem{title: "Production", desc: "Deploy to production environment", value: "production", icon: "🚀"},
		providerItem{title: "AWS", desc: "Deploy to Amazon Web Services", value: "aws", icon: "☁️"},
		providerItem{title: "GCP", desc: "Deploy to Google Cloud Platform", value: "gcp", icon: "☁️"},
		providerItem{title: "Hetzner", desc: "Deploy to Hetzner Cloud", value: "hetzner", icon: "☁️"},
		providerItem{title: "DigitalOcean", desc: "Deploy to DigitalOcean", value: "digitalocean", icon: "🌊"},
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Provider"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return Model{
		state:         StateSelectProvider,
		list:          l,
		width:         80,
		height:        24,
		Configuration: make(map[string]string),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4) // adjust for margins
	}

	// Handle state-specific updates
	switch m.state {
	case StateSelectProvider:
		// Check for selection
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "enter" {
				if i, ok := m.list.SelectedItem().(providerItem); ok {
					m.Provider = i.value
					m.setupInputs(m.Provider)
					m.state = StateConfigureProvider
				} else {
					// Should not happen if items are strictly typed, but good practice
					return m, tea.Batch(tea.Println("Error: Invalid selection type"))
				}
				return m, nil
			}
		}
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)

	case StateConfigureProvider:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				if m.focusIndex == len(m.inputs) {
					// Submit button selected
					m.collectConfiguration()
					return m, tea.Quit
				}
				// Next input
				m.focusIndex++
				if m.focusIndex > len(m.inputs) {
					m.focusIndex = 0
				}

			case "tab", "down":
				m.focusIndex++
				if m.focusIndex > len(m.inputs) {
					m.focusIndex = 0
				}
			case "shift+tab", "up":
				m.focusIndex--
				if m.focusIndex < 0 {
					m.focusIndex = len(m.inputs)
				}
			case "esc":
				m.state = StateSelectProvider
			}
		}

		// Update inputs
		for i := range m.inputs {
			if i == m.focusIndex {
				m.inputs[i].Focus()
				m.inputs[i].TextStyle = inputStyle
			} else {
				m.inputs[i].Blur()
				m.inputs[i].TextStyle = lipgloss.NewStyle()
			}
			m.inputs[i], cmd = m.inputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) setupInputs(provider string) {
	m.inputs = []textinput.Model{}
	m.focusIndex = 0

	addInput := func(placeholder, value string, echo bool) {
		t := textinput.New()
		t.Placeholder = placeholder
		t.SetValue(value)
		if !echo {
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}
		m.inputs = append(m.inputs, t)
	}

	switch provider {
	case "azure":
		addInput("Location", "East US", true)
		addInput("VM Size", "Standard_B2s", true)
		addInput("Disk Size (GB)", "30", true)
		addInput("SSH Key Path", "~/.ssh/id_rsa.pub", true)
	case "local":
		addInput("App Port", "8000", true)
		addInput("WebSocket Port", "6001", true)
		addInput("Data Directory", "~/.coolify", true)
	case "production":
		addInput("Domain", "", true)
		addInput("SSH Host", "", true)
		addInput("SSH User", "root", true)
		addInput("SSH Key Path", "~/.ssh/id_rsa", true)
	default:
		// Generic inputs
		addInput("Configuration 1", "", true)
		addInput("Configuration 2", "", true)
	}
}

func (m *Model) collectConfiguration() {
	// Map inputs back to known keys.
	// In a real app, strict mapping would be better, but index-based mapping works for this wizard flow.
	var keys []string

	switch m.Provider {
	case "azure":
		keys = []string{"location", "vm_size", "disk_size", "ssh_key"}
	case "local":
		keys = []string{"app_port", "ws_port", "data_dir"}
	case "production":
		keys = []string{"domain", "ssh_host", "ssh_user", "ssh_key"}
	default:
		keys = []string{"config1", "config2"}
	}

	for i, input := range m.inputs {
		if i < len(keys) {
			m.Configuration[keys[i]] = input.Value()
		}
	}
}

// View renders the UI
func (m Model) View() string {
	switch m.state {
	case StateSelectProvider:
		return docStyle.Render(m.list.View())
	case StateConfigureProvider:
		return m.renderConfiguration()
	default:
		return "Unknown state"
	}
}

// renderConfiguration renders the configuration screen
func (m Model) renderConfiguration() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("⚙️ Configure Deployment"))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "Provider: %s\n\n", strings.ToUpper(m.Provider))

	for i, input := range m.inputs {
		b.WriteString(input.View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n")

	// Render submit button
	btn := "[ Submit ]"
	if m.focusIndex == len(m.inputs) {
		btn = activeButtonStyle.Render(btn)
	} else {
		btn = buttonStyle.Render(btn)
	}
	b.WriteString(btn)

	return docStyle.Render(b.String())
}
