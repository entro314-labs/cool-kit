package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/entro314-labs/cool-kit/internal/config"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2).
			MarginTop(1).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EE6FF8")).
			Bold(true)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A49FA5")).
				MarginLeft(2)

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

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#14F195")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)
)

// State represents the current state of the UI
type State int

const (
	StateSelectProvider State = iota
	StateConfigureProvider
	StateDeploying
	StateComplete
	StateError
)

// Model represents the main UI model
type Model struct {
	state       State
	Provider    string // Made public
	config      *config.Config
	choices     []Choice
	cursor      int
	selected    int
	loading     bool
	err         error
	progress    int
	totalSteps  int
	currentStep string
	logs        []string
	width       int
	height      int
}

// Choice represents a selectable option
type Choice struct {
	Title       string
	Description string
	Value       string
	Icon        string
}

// NewModel creates a new UI model
func NewModel() Model {
	providers := []Choice{
		{
			Title:       "Azure",
			Description: "Deploy to Microsoft Azure cloud",
			Value:       "azure",
			Icon:        "☁️",
		},
		{
			Title:       "Local",
			Description: "Set up local development environment",
			Value:       "local",
			Icon:        "🏠",
		},
		{
			Title:       "Production",
			Description: "Deploy to production environment",
			Value:       "production",
			Icon:        "🚀",
		},
	}

	return Model{
		state:      StateSelectProvider,
		choices:    providers,
		cursor:     0,
		totalSteps: 5,
		width:      80,
		height:     24,
		logs:       []string{},
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles updates
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleEnter()

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case tea.KeyLeft:
			if m.state == StateConfigureProvider {
				return m.handleBack()
			}

		case tea.KeyRight, tea.KeyTab:
			if m.state == StateConfigureProvider {
				return m.handleNext()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case ProgressMsg:
		m.progress = msg.Progress
		m.currentStep = msg.Step
		m.logs = append(m.logs, fmt.Sprintf("✅ %s", msg.Step))

	case ErrorMsg:
		m.err = msg.Err
		m.state = StateError
		m.logs = append(m.logs, fmt.Sprintf("❌ Error: %v", msg.Err))

	case CompleteMsg:
		m.state = StateComplete
		m.logs = append(m.logs, "🎉 Deployment completed successfully!")
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	switch m.state {
	case StateSelectProvider:
		return m.renderProviderSelection()
	case StateConfigureProvider:
		return m.renderConfiguration()
	case StateDeploying:
		return m.renderDeployment()
	case StateComplete:
		return m.renderComplete()
	case StateError:
		return m.renderError()
	default:
		return "Unknown state"
	}
}

// renderProviderSelection renders the provider selection screen
func (m Model) renderProviderSelection() string {
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("🚀 Coolify Community CLI"))
	content.WriteString("\n\n")
	content.WriteString("Choose your deployment provider:\n\n")

	// Choices
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = "👉"
			content.WriteString(selectedStyle.Render(fmt.Sprintf("%s %s %s", cursor, choice.Icon, choice.Title)))
			content.WriteString("\n")
			content.WriteString(descriptionStyle.Render(fmt.Sprintf("   %s", choice.Description)))
		} else {
			content.WriteString(fmt.Sprintf("%s %s %s", cursor, choice.Icon, choice.Title))
			content.WriteString("\n")
			content.WriteString(descriptionStyle.Render(fmt.Sprintf("   %s", choice.Description)))
		}
		content.WriteString("\n")
	}

	// Instructions
	content.WriteString("\n")
	content.WriteString(descriptionStyle.Render("↑/↓ - Navigate  |  Enter - Select  |  Esc - Quit"))

	return content.String()
}

// renderConfiguration renders the configuration screen
func (m Model) renderConfiguration() string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("⚙️ Configure Deployment"))
	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("Provider: %s\n\n", strings.ToUpper(m.Provider)))

	// Configuration form would go here
	content.WriteString("Configuration options for ")
	content.WriteString(m.Provider)
	content.WriteString(" deployment:\n\n")

	// Provider-specific configuration options
	switch m.Provider {
	case "azure":
		content.WriteString("  • Location: East US\n")
		content.WriteString("  • VM Size: Standard_B2s\n")
		content.WriteString("  • Disk Size: 30 GB\n")
		content.WriteString("  • SSH Key: ~/.ssh/id_rsa.pub\n")
	case "local":
		content.WriteString("  • App Port: 8000\n")
		content.WriteString("  • WebSocket Port: 6001\n")
		content.WriteString("  • Data Directory: ~/.coolify\n")
	case "production":
		content.WriteString("  • Domain: (required)\n")
		content.WriteString("  • SSH Host: (required)\n")
		content.WriteString("  • SSH User: root\n")
		content.WriteString("  • SSH Key: ~/.ssh/id_rsa\n")
	case "aws":
		content.WriteString("  • Region: us-east-1\n")
		content.WriteString("  • Instance Type: t3.medium\n")
		content.WriteString("  • Disk Size: 30 GB\n")
	case "gcp":
		content.WriteString("  • Project: (required)\n")
		content.WriteString("  • Zone: us-central1-a\n")
		content.WriteString("  • Machine Type: e2-medium\n")
	case "hetzner":
		content.WriteString("  • Location: nbg1\n")
		content.WriteString("  • Server Type: cx21\n")
		content.WriteString("  • SSH Key: ~/.ssh/id_rsa.pub\n")
	case "digitalocean":
		content.WriteString("  • Region: nyc1\n")
		content.WriteString("  • Size: s-2vcpu-2gb\n")
		content.WriteString("  • SSH Key: ~/.ssh/id_rsa.pub\n")
	default:
		content.WriteString("  • Default settings will be used\n")
	}

	// Navigation
	content.WriteString("\n")
	content.WriteString(buttonStyle.Render("← Back"))
	content.WriteString("   ")
	content.WriteString(activeButtonStyle.Render("Next →"))

	return content.String()
}

// renderDeployment renders the deployment progress screen
func (m Model) renderDeployment() string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("🚀 Deploying Coolify"))
	content.WriteString("\n\n")

	// Progress bar
	progressWidth := 40
	filled := int(float64(m.progress) / 100.0 * float64(progressWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", progressWidth-filled)
	content.WriteString(fmt.Sprintf("Progress: [%s] %d%%\n", bar, m.progress))

	// Current step
	if m.currentStep != "" {
		content.WriteString(fmt.Sprintf("\nCurrent: %s", m.currentStep))
	}

	// Logs
	if len(m.logs) > 0 {
		content.WriteString("\n\n")
		content.WriteString("📋 Deployment Log:\n")
		for _, log := range m.logs {
			content.WriteString(fmt.Sprintf("  %s\n", log))
		}
	}

	return content.String()
}

// renderComplete renders the completion screen
func (m Model) renderComplete() string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("🎉 Deployment Complete!"))
	content.WriteString("\n\n")
	content.WriteString(statusStyle.Render("✅ Coolify has been successfully deployed!"))
	content.WriteString("\n\n")

	// Show access information
	content.WriteString("📋 Access Information:\n")
	content.WriteString("  • URL: http://your-coolify-instance.com\n")
	content.WriteString("  • Admin: admin@coolify.local\n")
	content.WriteString("  • Password: admin123\n")

	content.WriteString("\n")
	content.WriteString(buttonStyle.Render("Press Enter to exit"))

	return content.String()
}

// renderError renders the error screen
func (m Model) renderError() string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("❌ Deployment Failed"))
	content.WriteString("\n\n")
	content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	content.WriteString("\n\n")

	// Show logs
	if len(m.logs) > 0 {
		content.WriteString("📋 Deployment Log:\n")
		for _, log := range m.logs {
			content.WriteString(fmt.Sprintf("  %s\n", log))
		}
	}

	content.WriteString("\n")
	content.WriteString(buttonStyle.Render("Press Enter to exit"))

	return content.String()
}

// Message types
type ProgressMsg struct {
	Progress int
	Step     string
}

type ErrorMsg struct {
	Err error
}

type CompleteMsg struct{}

// Helper methods
func (m Model) handleEnter() (Model, tea.Cmd) {
	switch m.state {
	case StateSelectProvider:
		m.Provider = m.choices[m.cursor].Value
		m.state = StateConfigureProvider
		return m, nil
	case StateComplete, StateError:
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m Model) handleBack() (Model, tea.Cmd) {
	if m.state == StateConfigureProvider {
		m.state = StateSelectProvider
	}
	return m, nil
}

func (m Model) handleNext() (Model, tea.Cmd) {
	if m.state == StateConfigureProvider {
		m.state = StateDeploying
		return m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
			// Simulate deployment progress
			return ProgressMsg{Progress: 25, Step: "Initializing deployment..."}
		})
	}
	return m, nil
}
