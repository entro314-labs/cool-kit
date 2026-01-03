package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MainMenuState represents the state of the main menu
type MainMenuState int

const (
	StateMenu MainMenuState = iota
	StateExit
)

type MainMenuSelection string

const (
	// Pillar 1: Deploy Coolify
	SelectionInstall MainMenuSelection = "install"

	// Pillar 2: Deploy TO Coolify
	SelectionDeploy MainMenuSelection = "deploy"
	SelectionInit   MainMenuSelection = "init"
	SelectionLink   MainMenuSelection = "link"
	SelectionLogs   MainMenuSelection = "logs"
	SelectionStatus MainMenuSelection = "status"

	// Management
	SelectionInstances MainMenuSelection = "instances"
	SelectionServices  MainMenuSelection = "services"
	SelectionEnv       MainMenuSelection = "env"
	SelectionConfig    MainMenuSelection = "config"

	// Tools
	SelectionBackup MainMenuSelection = "backup"
	SelectionBadge  MainMenuSelection = "badge"
	SelectionCI     MainMenuSelection = "ci"

	// Other
	SelectionHelp MainMenuSelection = "help"
	SelectionExit MainMenuSelection = "exit"
	SelectionNone MainMenuSelection = ""
)

// MenuSection groups related menu choices
type MenuSection struct {
	Title   string
	Choices []MenuChoice
}

type MainMenuModel struct {
	sections      []MenuSection
	cursor        int
	selected      MainMenuSelection
	width         int
	height        int
	flatChoices   []MenuChoice // flattened for navigation
	sectionBounds []int        // indices where sections start
}

type MenuChoice struct {
	Title       string
	Description string
	Selection   MainMenuSelection
	Icon        string
}

// Enhanced styles for modern widescreen TUI
var (
	// Logo/Brand colors
	brandPrimary   = lipgloss.Color("#7D56F4") // Purple
	brandSecondary = lipgloss.Color("#00D4FF") // Cyan
	brandAccent    = lipgloss.Color("#14F195") // Green

	// Header styles
	logoStyle = lipgloss.NewStyle().
			Foreground(brandSecondary).
			Bold(true)

	taglineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)

	// Section panel styles
	sectionBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(0, 1)

	sectionBoxActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(brandPrimary).
				Padding(0, 1)

	sectionTitleStyle = lipgloss.NewStyle().
				Foreground(brandPrimary).
				Bold(true).
				MarginBottom(1)

	// Menu item styles
	menuItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(brandAccent).
				Bold(true)

	// Description panel
	descBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(brandAccent).
			Padding(0, 2).
			Foreground(lipgloss.Color("#CCCCCC"))

	descTitleStyle = lipgloss.NewStyle().
			Foreground(brandAccent).
			Bold(true)

	descTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	// Footer styles
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	footerKeyStyle = lipgloss.NewStyle().
			Foreground(brandPrimary).
			Bold(true)

	footerSepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333"))
)

func NewMainMenuModel() MainMenuModel {
	sections := []MenuSection{
		{
			Title: "🏗️  DEPLOY COOLIFY",
			Choices: []MenuChoice{
				{Title: "Install Coolify", Description: "Install Coolify on any cloud provider or VM", Selection: SelectionInstall, Icon: "🚀"},
			},
		},
		{
			Title: "📦  DEPLOY APPS",
			Choices: []MenuChoice{
				{Title: "Deploy Application", Description: "Deploy app to Coolify with smart detection", Selection: SelectionDeploy, Icon: "🎯"},
				{Title: "Init Project", Description: "Initialize with templates (Dockerfile, etc.)", Selection: SelectionInit, Icon: "✨"},
				{Title: "Link Project", Description: "Link to existing Coolify application", Selection: SelectionLink, Icon: "🔗"},
			},
		},
		{
			Title: "📊  MONITOR",
			Choices: []MenuChoice{
				{Title: "View Logs", Description: "Stream application logs in real-time", Selection: SelectionLogs, Icon: "📜"},
				{Title: "Health Status", Description: "Check deployment and service health", Selection: SelectionStatus, Icon: "💚"},
				{Title: "Manage Services", Description: "Start, stop, restart services", Selection: SelectionServices, Icon: "⚙️"},
				{Title: "Environment Vars", Description: "Manage environment variables", Selection: SelectionEnv, Icon: "🔐"},
			},
		},
		{
			Title: "🔧  SETTINGS",
			Choices: []MenuChoice{
				{Title: "Manage Instances", Description: "Switch or add Coolify instances", Selection: SelectionInstances, Icon: "🌐"},
				{Title: "Configuration", Description: "View and edit CLI settings", Selection: SelectionConfig, Icon: "⚙️"},
			},
		},
		{
			Title: "🛠️  TOOLS",
			Choices: []MenuChoice{
				{Title: "Backup Instance", Description: "Backup Coolify data and volumes", Selection: SelectionBackup, Icon: "💾"},
				{Title: "Generate Badge", Description: "Create deployment status badge", Selection: SelectionBadge, Icon: "🏷️"},
				{Title: "Generate CI", Description: "Create GitHub Actions workflow", Selection: SelectionCI, Icon: "🔄"},
			},
		},
		{
			Title: "",
			Choices: []MenuChoice{
				{Title: "Help", Description: "Show help and documentation", Selection: SelectionHelp, Icon: "❓"},
				{Title: "Exit", Description: "Exit Cool Kit", Selection: SelectionExit, Icon: "👋"},
			},
		},
	}

	// Flatten choices for navigation
	var flat []MenuChoice
	var bounds []int
	for _, section := range sections {
		bounds = append(bounds, len(flat))
		flat = append(flat, section.Choices...)
	}

	return MainMenuModel{
		sections:      sections,
		cursor:        0,
		flatChoices:   flat,
		sectionBounds: bounds,
		width:         100,
		height:        24,
	}
}

func (m MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.selected = SelectionExit
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.flatChoices)-1 {
				m.cursor++
			}
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			m.cursor = len(m.flatChoices) - 1
		case "tab", "right", "l":
			// Jump to next section
			for i, bound := range m.sectionBounds {
				if bound > m.cursor && i < len(m.sectionBounds) {
					m.cursor = bound
					break
				}
			}
		case "shift+tab", "left", "h":
			// Jump to previous section
			for i := len(m.sectionBounds) - 1; i >= 0; i-- {
				if m.sectionBounds[i] < m.cursor {
					m.cursor = m.sectionBounds[i]
					break
				}
			}
		case "enter", " ":
			m.selected = m.flatChoices[m.cursor].Selection
			return m, tea.Quit
		case "1":
			m.selected = SelectionInstall
			return m, tea.Quit
		case "2":
			m.selected = SelectionDeploy
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// getSectionForCursor returns which section index the cursor is in
func (m MainMenuModel) getSectionForCursor() int {
	for i := len(m.sectionBounds) - 1; i >= 0; i-- {
		if m.cursor >= m.sectionBounds[i] {
			return i
		}
	}
	return 0
}

// renderSection renders a single section as a box
func (m MainMenuModel) renderSection(sectionIdx int, width int) string {
	section := m.sections[sectionIdx]
	isActiveSection := m.getSectionForCursor() == sectionIdx

	var content strings.Builder

	// Section title
	if section.Title != "" {
		content.WriteString(sectionTitleStyle.Render(section.Title))
		content.WriteString("\n")
	}

	// Items
	startIdx := m.sectionBounds[sectionIdx]
	for i, choice := range section.Choices {
		choiceIdx := startIdx + i
		isSelected := m.cursor == choiceIdx

		line := fmt.Sprintf("%s %s", choice.Icon, choice.Title)
		if isSelected {
			line = "▸ " + line
			content.WriteString(selectedItemStyle.Render(line))
		} else {
			line = "  " + line
			content.WriteString(menuItemStyle.Render(line))
		}
		content.WriteString("\n")
	}

	// Use active or inactive style
	boxStyle := sectionBoxStyle.Width(width)
	if isActiveSection {
		boxStyle = sectionBoxActiveStyle.Width(width)
	}

	return boxStyle.Render(content.String())
}

func (m MainMenuModel) View() string {
	var s strings.Builder

	// Calculate layout dimensions
	totalWidth := m.width
	if totalWidth < 80 {
		totalWidth = 80
	}
	if totalWidth > 110 {
		totalWidth = 110
	}

	// Header - spans full width, horizontally centered
	header := logoStyle.Render("🧊 COOL KIT") + "  " + taglineStyle.Render("The Complete Coolify Toolkit")
	headerLine := lipgloss.NewStyle().Width(totalWidth).Align(lipgloss.Center).Render(header)
	s.WriteString(headerLine)
	s.WriteString("\n\n")

	// Calculate column widths for 3-column layout (narrower columns)
	colGap := 2
	colWidth := (totalWidth - colGap*2) / 4 // Narrower: divide by 4 instead of 3
	if colWidth < 22 {
		colWidth = 22
	}
	if colWidth > 28 {
		colWidth = 28
	}

	// Left column: Deploy Coolify + Deploy Apps
	leftCol := m.renderSection(0, colWidth)
	leftCol += "\n"
	leftCol += m.renderSection(1, colWidth)

	// Middle column: Monitor + Settings
	midCol := m.renderSection(2, colWidth)
	midCol += "\n"
	midCol += m.renderSection(3, colWidth)

	// Right column: Tools + Help/Exit
	rightCol := m.renderSection(4, colWidth)
	rightCol += "\n"
	rightCol += m.renderSection(5, colWidth)

	// Join columns horizontally
	gap := strings.Repeat(" ", colGap)
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, gap, midCol, gap, rightCol)
	s.WriteString(lipgloss.NewStyle().Width(totalWidth).Align(lipgloss.Center).Render(columns))
	s.WriteString("\n\n")

	// Description box - full width at bottom
	currentChoice := m.flatChoices[m.cursor]
	descContent := descTitleStyle.Render(currentChoice.Icon+" "+currentChoice.Title) + "  " +
		descTextStyle.Render(currentChoice.Description)
	descWidth := colWidth*3 + colGap*2
	descBox := descBoxStyle.Width(descWidth).Render(descContent)
	s.WriteString(lipgloss.NewStyle().Width(totalWidth).Align(lipgloss.Center).Render(descBox))
	s.WriteString("\n\n")

	// Footer with keyboard hints - centered
	footerParts := []string{
		footerKeyStyle.Render("↑↓") + " navigate",
		footerKeyStyle.Render("←→") + " sections",
		footerKeyStyle.Render("Enter") + " select",
		footerKeyStyle.Render("q") + " quit",
	}
	footer := footerStyle.Render(strings.Join(footerParts, footerSepStyle.Render(" │ ")))
	footerLine := lipgloss.NewStyle().Width(totalWidth).Align(lipgloss.Center).Render(footer)
	s.WriteString(footerLine)

	return s.String()
}

// RunMainMenu runs the main menu and returns the selected action
func RunMainMenu() (MainMenuSelection, error) {
	p := tea.NewProgram(NewMainMenuModel(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return SelectionNone, err
	}
	if model, ok := m.(MainMenuModel); ok {
		return model.selected, nil
	}
	return SelectionNone, nil
}
