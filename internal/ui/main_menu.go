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

// Styles for enhanced TUI
var (
	sectionTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true).
				MarginTop(1).
				MarginBottom(0)

	menuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#14F195")).
				Bold(true).
				PaddingLeft(0)

	dimDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			PaddingLeft(4)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF")).
			Bold(true).
			MarginBottom(1)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			MarginTop(1)
)

func NewMainMenuModel() MainMenuModel {
	sections := []MenuSection{
		{
			Title: "🏗️  Deploy Coolify (Pillar 1)",
			Choices: []MenuChoice{
				{Title: "Install Coolify", Description: "Install Coolify on any cloud provider or VM", Selection: SelectionInstall, Icon: "🚀"},
			},
		},
		{
			Title: "📦  Deploy Apps (Pillar 2)",
			Choices: []MenuChoice{
				{Title: "Deploy Application", Description: "Deploy app to Coolify with smart detection", Selection: SelectionDeploy, Icon: "🎯"},
				{Title: "Init Project", Description: "Initialize with templates (Dockerfile, etc.)", Selection: SelectionInit, Icon: "✨"},
				{Title: "Link Project", Description: "Link to existing Coolify application", Selection: SelectionLink, Icon: "🔗"},
			},
		},
		{
			Title: "📊  Monitor & Manage",
			Choices: []MenuChoice{
				{Title: "View Logs", Description: "Stream application logs", Selection: SelectionLogs, Icon: "📜"},
				{Title: "Health Status", Description: "Check deployment and service health", Selection: SelectionStatus, Icon: "💚"},
				{Title: "Manage Services", Description: "Start, stop, restart services", Selection: SelectionServices, Icon: "⚙️"},
				{Title: "Environment Vars", Description: "Manage environment variables", Selection: SelectionEnv, Icon: "🔐"},
			},
		},
		{
			Title: "🔧  Settings",
			Choices: []MenuChoice{
				{Title: "Manage Instances", Description: "Switch or add Coolify instances", Selection: SelectionInstances, Icon: "🌐"},
				{Title: "Configuration", Description: "View and edit CLI settings", Selection: SelectionConfig, Icon: "⚙️"},
			},
		},
		{
			Title: "🛠️  Tools",
			Choices: []MenuChoice{
				{Title: "Backup Instance", Description: "Backup Coolify data, volumes, and SSH keys", Selection: SelectionBackup, Icon: "💾"},
				{Title: "Generate Badge", Description: "Create deployment status badge for README", Selection: SelectionBadge, Icon: "🏷️"},
				{Title: "Generate CI Workflow", Description: "Create GitHub Actions deploy workflow", Selection: SelectionCI, Icon: "🔄"},
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

func (m MainMenuModel) View() string {
	var s strings.Builder

	// Header with ASCII art style
	s.WriteString(headerStyle.Render("╭───────────────────────────────────────╮"))
	s.WriteString("\n")
	s.WriteString(headerStyle.Render("│      🧊 Cool Kit - Coolify Toolkit    │"))
	s.WriteString("\n")
	s.WriteString(headerStyle.Render("╰───────────────────────────────────────╯"))
	s.WriteString("\n\n")

	choiceIdx := 0
	for _, section := range m.sections {
		// Section title
		if section.Title != "" {
			s.WriteString(sectionTitleStyle.Render(section.Title))
			s.WriteString("\n")
		}

		// Section items
		for _, choice := range section.Choices {
			cursor := "  "
			if m.cursor == choiceIdx {
				cursor = "▸ "
				line := fmt.Sprintf("%s%s %s", cursor, choice.Icon, choice.Title)
				s.WriteString(selectedItemStyle.Render(line))
			} else {
				line := fmt.Sprintf("%s%s %s", cursor, choice.Icon, choice.Title)
				s.WriteString(menuItemStyle.Render(line))
			}
			s.WriteString("\n")

			// Show description only for selected item
			if m.cursor == choiceIdx {
				s.WriteString(dimDescStyle.Render(choice.Description))
				s.WriteString("\n")
			}
			choiceIdx++
		}
	}

	s.WriteString("\n")
	s.WriteString(footerStyle.Render("↑/↓ navigate • Enter select • q quit"))
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
