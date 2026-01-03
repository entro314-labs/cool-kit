package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	keys    []string
	values  map[string]interface{}
	cursor  int
	width   int
	height  int
	editKey string
	editVal string
}

// Styles for config menu
var (
	configTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00D4FF")).
				Bold(true)

	configBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)

	configKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9D76FF")).
			Width(22)

	configValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#CCCCCC"))

	configSelectedKeyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#14F195")).
				Bold(true).
				Width(22)

	configSelectedValueStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#14F195")).
					Bold(true)

	configHelpBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#444")).
				Padding(1, 2).
				Foreground(lipgloss.Color("#888"))

	configFooterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#555"))

	configFooterKeyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)
)

func NewConfigModel() (ConfigModel, error) {
	if err := config.Initialize(); err != nil {
		return ConfigModel{}, err
	}
	cfg := config.Get()

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
		width:  100,
		height: 24,
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
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m ConfigModel) View() string {
	var s strings.Builder

	// Calculate layout
	totalWidth := m.width
	if totalWidth < 80 {
		totalWidth = 80
	}
	if totalWidth > 100 {
		totalWidth = 100
	}

	// Header - centered
	header := configTitleStyle.Render("⚙️  Configuration Manager")
	headerLine := lipgloss.NewStyle().Width(totalWidth).Align(lipgloss.Center).Render(header)
	s.WriteString(headerLine)
	s.WriteString("\n\n")

	// Two-column layout
	colGap := 4
	leftWidth := (totalWidth - colGap) * 2 / 3
	rightWidth := totalWidth - leftWidth - colGap

	// LEFT: Config keys/values
	var configContent strings.Builder
	configContent.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#9D76FF")).Render("📋 Settings"))
	configContent.WriteString("\n\n")

	for i, key := range m.keys {
		val := m.values[key]
		if m.cursor == i {
			configContent.WriteString("▸ ")
			configContent.WriteString(configSelectedKeyStyle.Render(key))
			configContent.WriteString(configSelectedValueStyle.Render(fmt.Sprintf("%v", val)))
		} else {
			configContent.WriteString("  ")
			configContent.WriteString(configKeyStyle.Render(key))
			configContent.WriteString(configValueStyle.Render(fmt.Sprintf("%v", val)))
		}
		configContent.WriteString("\n")
	}

	leftPanel := configBoxStyle.Width(leftWidth).Render(configContent.String())

	// RIGHT: Help/context panel
	var helpContent strings.Builder
	helpContent.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#888")).Render("💡 Quick Help"))
	helpContent.WriteString("\n\n")

	selectedKey := m.keys[m.cursor]
	switch selectedKey {
	case "provider":
		helpContent.WriteString("Default cloud provider\nfor deployments.\n\nOptions: azure, aws,\ngcp, local, hetzner")
	case "environment":
		helpContent.WriteString("Deployment environment.\n\nOptions: development,\nstaging, production")
	case "local.app_port":
		helpContent.WriteString("Port for local Coolify\nweb interface.\n\nDefault: 8000")
	case "local.websocket_port":
		helpContent.WriteString("WebSocket port for\nreal-time updates.\n\nDefault: 6001")
	case "azure.location":
		helpContent.WriteString("Azure region for\nresource deployment.\n\nExample: eastus, westus2")
	case "azure.vm_size":
		helpContent.WriteString("Azure VM size for\nCoolify host.\n\nRecommended: Standard_B2s")
	default:
		helpContent.WriteString("Select a setting to\nview more information.")
	}

	rightPanel := configHelpBoxStyle.Width(rightWidth).Render(helpContent.String())

	// Join panels
	gap := strings.Repeat(" ", colGap)
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, gap, rightPanel)
	s.WriteString(panels)
	s.WriteString("\n\n")

	// Footer - centered
	footerParts := []string{
		configFooterKeyStyle.Render("↑↓") + " navigate",
		configFooterKeyStyle.Render("Enter") + " edit",
		configFooterKeyStyle.Render("Esc") + " exit",
	}
	footer := configFooterStyle.Render(strings.Join(footerParts, "  │  "))
	footerLine := lipgloss.NewStyle().Width(totalWidth).Align(lipgloss.Center).Render(footer)
	s.WriteString(footerLine)

	return s.String()
}

// RunConfigMenu runs the config menu and returns the selected key to edit
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
