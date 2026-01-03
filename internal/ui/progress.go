package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressModel represents an interactive deployment progress UI
type ProgressModel struct {
	provider    string
	steps       []DeploymentStep
	currentStep int
	progress    progress.Model
	spinner     spinner.Model
	logViewport viewport.Model
	logs        []LogEntry
	width       int
	height      int
	done        bool
	err         error
	startTime   time.Time
}

// DeploymentStep represents a single deployment step
type DeploymentStep struct {
	Name        string
	Description string
	Status      StepStatus
	StartTime   time.Time
	EndTime     time.Time
	Progress    float64
}

// StepStatus represents the status of a deployment step
type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepComplete
	StepFailed
	StepSkipped
)

// LogEntry represents a log message
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Message   string
	Step      int
}

// LogLevel represents log severity
type LogLevel int

const (
	LogInfo LogLevel = iota
	LogSuccess
	LogWarning
	LogError
	LogDebug
)

// Progress message types
type (
	StepStartMsg struct {
		StepIndex int
		StepName  string
	}

	StepProgressMsg struct {
		StepIndex int
		Progress  float64
		Message   string
	}

	StepCompleteMsg struct {
		StepIndex int
		Success   bool
		Message   string
	}

	LogMsg struct {
		Level   LogLevel
		Message string
	}

	DeploymentCompleteMsg struct {
		Success bool
		Message string
	}

	tickMsg time.Time
)

// Styles for enhanced visuals
var (
	progressTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#9D76FF"}).
				Padding(0, 2).
				MarginBottom(1)

	stepBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#9D76FF"}).
			Padding(1, 2)

	logBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444")).
			Padding(0, 1)

	successBannerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#14F195")).
				Background(lipgloss.Color("#1a2e1a")).
				Padding(1, 4).
				MarginTop(1).
				MarginBottom(1)

	errorBannerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF5F87")).
				Background(lipgloss.Color("#2e1a1a")).
				Padding(1, 4).
				MarginTop(1).
				MarginBottom(1)

	urlBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#14F195")).
			Foreground(lipgloss.Color("#14F195")).
			Padding(0, 2).
			MarginTop(1)

	summaryLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888")).
				Width(14)

	summaryValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fff")).
				Bold(true)

	progressFooterStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666")).
				Italic(true).
				MarginTop(1)
)

// NewProgressModel creates a new progress model
func NewProgressModel(provider string, steps []DeploymentStep) ProgressModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)

	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
		FPS:    time.Second / 12,
	}
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#9D76FF"))

	vp := viewport.New(80, 6)
	vp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#888"))

	return ProgressModel{
		provider:    provider,
		steps:       steps,
		currentStep: 0,
		progress:    p,
		spinner:     s,
		logViewport: vp,
		logs:        []LogEntry{},
		width:       80,
		height:      24,
		startTime:   time.Now(),
	}
}

// Init initializes the progress model
func (m ProgressModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.startNextStep(),
		m.tickCmd(),
	)
}

// tickCmd returns a command that ticks every second
func (m ProgressModel) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles updates
func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDown:
			var cmd tea.Cmd
			m.logViewport, cmd = m.logViewport.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 24
		m.logViewport.Width = msg.Width - 6
		stepBoxStyle = stepBoxStyle.Width(msg.Width - 4)
		logBoxStyle = logBoxStyle.Width(msg.Width - 4)

	case tickMsg:
		if !m.done {
			cmds = append(cmds, m.tickCmd())
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case StepStartMsg:
		if msg.StepIndex < len(m.steps) {
			m.steps[msg.StepIndex].Status = StepRunning
			m.steps[msg.StepIndex].StartTime = time.Now()
			m.currentStep = msg.StepIndex
			m.addLog(LogInfo, fmt.Sprintf("Starting: %s", msg.StepName))
		}

	case StepProgressMsg:
		if msg.StepIndex < len(m.steps) {
			m.steps[msg.StepIndex].Progress = msg.Progress
			if msg.Message != "" {
				m.addLog(LogDebug, msg.Message)
			}
			// Auto-complete step when progress reaches 100%
			if msg.Progress >= 1.0 && m.steps[msg.StepIndex].Status == StepRunning {
				m.steps[msg.StepIndex].Status = StepComplete
				m.steps[msg.StepIndex].EndTime = time.Now()
			}
		}

	case StepCompleteMsg:
		if msg.StepIndex < len(m.steps) {
			m.steps[msg.StepIndex].EndTime = time.Now()
			if msg.Success {
				m.steps[msg.StepIndex].Status = StepComplete
				m.steps[msg.StepIndex].Progress = 1.0
				m.addLog(LogSuccess, fmt.Sprintf("✓ %s", msg.Message))
			} else {
				m.steps[msg.StepIndex].Status = StepFailed
				m.addLog(LogError, fmt.Sprintf("✗ %s", msg.Message))
				m.err = fmt.Errorf("deployment step failed: %s", msg.Message)
			}

			if msg.Success && msg.StepIndex+1 < len(m.steps) {
				return m, m.startNextStep()
			}
		}

	case LogMsg:
		m.addLog(msg.Level, msg.Message)

	case DeploymentCompleteMsg:
		m.done = true
		if msg.Success {
			m.addLog(LogSuccess, "🎉 "+msg.Message)
		} else {
			m.addLog(LogError, "❌ "+msg.Message)
			m.err = fmt.Errorf("deployment failed: %s", msg.Message)
		}
		return m, tea.Quit
	}

	return m, tea.Batch(cmds...)
}

// View renders the progress UI
func (m ProgressModel) View() string {
	if m.done {
		return m.renderComplete()
	}

	var b strings.Builder

	// Header
	header := progressTitleStyle.Width(m.width).Render(
		fmt.Sprintf("🚀 Deploying Coolify to %s", strings.ToUpper(m.provider)),
	)
	b.WriteString(header)
	b.WriteString("\n\n")

	// Overall progress
	completedSteps := 0
	for _, step := range m.steps {
		if step.Status == StepComplete {
			completedSteps++
		}
	}
	overallProgress := float64(completedSteps) / float64(len(m.steps))
	elapsed := time.Since(m.startTime).Round(time.Second)

	progressLine := fmt.Sprintf("Progress: %d/%d steps • Elapsed: %s", completedSteps, len(m.steps), elapsed)
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Render(progressLine))
	b.WriteString("\n")
	b.WriteString(m.progress.ViewAs(overallProgress))
	b.WriteString("\n\n")

	// Steps list
	var stepsContent strings.Builder
	for i, step := range m.steps {
		icon, style := m.getStepIconAndStyle(step.Status)

		// Step name with icon
		if step.Status == StepRunning {
			// Show spinner for running step
			stepLine := fmt.Sprintf("%s %s %s", m.spinner.View(), icon, step.Name)
			stepsContent.WriteString(style.Render(stepLine))

			// Show elapsed time for running step
			if !step.StartTime.IsZero() {
				stepElapsed := time.Since(step.StartTime).Round(time.Second)
				stepsContent.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render(
					fmt.Sprintf(" (%s)", stepElapsed),
				))
			}

			// Show progress bar for running step
			if step.Progress > 0 && step.Progress < 1 {
				stepsContent.WriteString("\n    ")
				miniProgress := progress.New(progress.WithDefaultGradient(), progress.WithWidth(30), progress.WithoutPercentage())
				stepsContent.WriteString(miniProgress.ViewAs(step.Progress))
			}
		} else {
			stepLine := fmt.Sprintf("  %s %s", icon, step.Name)
			stepsContent.WriteString(style.Render(stepLine))

			// Show duration for completed steps
			if step.Status == StepComplete && !step.EndTime.IsZero() {
				duration := step.EndTime.Sub(step.StartTime).Round(time.Millisecond * 100)
				stepsContent.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Render(
					fmt.Sprintf(" (%s)", duration),
				))
			}
		}

		if i < len(m.steps)-1 {
			stepsContent.WriteString("\n")
		}
	}

	b.WriteString(stepBoxStyle.Width(m.width - 4).Render(stepsContent.String()))
	b.WriteString("\n\n")

	// Log viewport
	var logsContent strings.Builder
	logsContent.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#888")).Render("📋 Activity Log"))
	logsContent.WriteString("\n\n")

	// Build log content
	for _, log := range m.logs {
		logsContent.WriteString(m.formatLog(log))
		logsContent.WriteString("\n")
	}

	m.logViewport.SetContent(logsContent.String())
	m.logViewport.GotoBottom()

	b.WriteString(logBoxStyle.Width(m.width - 4).Height(8).Render(m.logViewport.View()))
	b.WriteString("\n")

	// Footer
	b.WriteString(progressFooterStyle.Render("↑/↓ scroll logs • Ctrl+C cancel"))

	return b.String()
}

// renderComplete renders the completion screen
func (m ProgressModel) renderComplete() string {
	var b strings.Builder
	totalDuration := time.Since(m.startTime).Round(time.Second)

	if m.err != nil {
		// Error screen
		b.WriteString(errorBannerStyle.Width(m.width - 8).Render("❌  DEPLOYMENT FAILED"))
		b.WriteString("\n\n")

		// Use structured error formatting
		b.WriteString(FormatErrorBox("Deployment Error", m.err, m.width))
		b.WriteString("\n\n")
	} else {
		// Success screen
		b.WriteString(successBannerStyle.Width(m.width - 8).Render("🎉  DEPLOYMENT SUCCESSFUL!"))
		b.WriteString("\n")

		// Dashboard URL (extract from last log)
		dashboardURL := m.extractDashboardURL()
		if dashboardURL != "" {
			urlContent := fmt.Sprintf("🌐 Dashboard: %s", dashboardURL)
			b.WriteString(urlBoxStyle.Width(m.width - 4).Render(urlContent))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Summary box
	var summaryContent strings.Builder
	summaryContent.WriteString(lipgloss.NewStyle().Bold(true).MarginBottom(1).Render("📊 Deployment Summary"))
	summaryContent.WriteString("\n\n")

	summaryContent.WriteString(summaryLabelStyle.Render("Provider:"))
	summaryContent.WriteString(summaryValueStyle.Render(strings.ToUpper(m.provider)))
	summaryContent.WriteString("\n")

	summaryContent.WriteString(summaryLabelStyle.Render("Duration:"))
	summaryContent.WriteString(summaryValueStyle.Render(totalDuration.String()))
	summaryContent.WriteString("\n")

	summaryContent.WriteString(summaryLabelStyle.Render("Steps:"))
	summaryContent.WriteString(summaryValueStyle.Render(fmt.Sprintf("%d/%d completed", m.countCompleted(), len(m.steps))))
	summaryContent.WriteString("\n")

	summaryBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444")).
		Padding(1, 2).
		Width(m.width - 4).
		Render(summaryContent.String())

	b.WriteString(summaryBox)
	b.WriteString("\n\n")

	// Next steps (only on success)
	if m.err == nil {
		nextSteps := lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Render("Next steps:\n")
		nextSteps += lipgloss.NewStyle().Foreground(lipgloss.Color("#14F195")).Render("  → Open the dashboard and complete initial setup\n")
		nextSteps += lipgloss.NewStyle().Foreground(lipgloss.Color("#14F195")).Render("  → Configure SSL with a custom domain")
		b.WriteString(nextSteps)
	}

	return b.String()
}

// Helper methods
func (m ProgressModel) getStepIconAndStyle(status StepStatus) (string, lipgloss.Style) {
	switch status {
	case StepPending:
		return "○", lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
	case StepRunning:
		return "◐", lipgloss.NewStyle().Foreground(lipgloss.Color("#9D76FF")).Bold(true)
	case StepComplete:
		return "✓", lipgloss.NewStyle().Foreground(lipgloss.Color("#14F195"))
	case StepFailed:
		return "✗", lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87"))
	case StepSkipped:
		return "⊘", lipgloss.NewStyle().Foreground(lipgloss.Color("#666")).Italic(true)
	default:
		return "○", lipgloss.NewStyle()
	}
}

func (m *ProgressModel) addLog(level LogLevel, message string) {
	m.logs = append(m.logs, LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Step:      m.currentStep,
	})
}

func (m ProgressModel) formatLog(log LogEntry) string {
	timestamp := log.Timestamp.Format("15:04:05")

	var icon string
	var style lipgloss.Style

	switch log.Level {
	case LogInfo:
		icon = "ℹ"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("#3291FF"))
	case LogSuccess:
		icon = "✓"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("#14F195"))
	case LogWarning:
		icon = "⚠"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBCA04"))
	case LogError:
		icon = "✗"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87"))
	case LogDebug:
		icon = "•"
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
	}

	return style.Render(fmt.Sprintf("[%s] %s %s", timestamp, icon, log.Message))
}

func (m ProgressModel) countCompleted() int {
	count := 0
	for _, step := range m.steps {
		if step.Status == StepComplete {
			count++
		}
	}
	return count
}

func (m ProgressModel) startNextStep() tea.Cmd {
	nextIdx := m.currentStep
	if m.currentStep > 0 {
		nextIdx = m.currentStep + 1
	}

	if nextIdx < len(m.steps) {
		return func() tea.Msg {
			return StepStartMsg{
				StepIndex: nextIdx,
				StepName:  m.steps[nextIdx].Name,
			}
		}
	}

	return nil
}

func (m ProgressModel) extractDashboardURL() string {
	// Look for URL in last few logs
	for i := len(m.logs) - 1; i >= 0 && i >= len(m.logs)-5; i-- {
		log := m.logs[i]
		if strings.Contains(log.Message, "http://") || strings.Contains(log.Message, "https://") {
			// Extract URL
			parts := strings.Fields(log.Message)
			for _, part := range parts {
				if strings.HasPrefix(part, "http://") || strings.HasPrefix(part, "https://") {
					return part
				}
			}
		}
	}
	return ""
}
