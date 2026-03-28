package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-terminal/internal/core/agent"
	"ai-terminal/internal/config"
	"ai-terminal/internal/providers"
	"ai-terminal/pkg/types"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	config       *config.Config
	agent        *agent.Engine
	provider     providers.Provider

	messages     []Message
	input        textarea.Model
	spinner       spinner.Model
	loading      bool
	showHelp     bool
	showTools    bool
	streamBuffer string
	width        int
	height       int
	err          error
}

type Message struct {
	Role    string
	Content string
	Time    time.Time
}

type AppConfig struct {
	Agent    *agent.Engine
	Config   *config.Config
	Provider providers.Provider
}

func NewApp(cfg AppConfig) Model {
	input := textarea.New()
	input.Placeholder = "Type your message here..."
	input.SetWidth(70)
	input.SetHeight(3)
	input.Focus()
	input.ShowLineNumbers = false

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		config:    cfg.Config,
		agent:     cfg.Agent,
		provider:  cfg.Provider,
		messages:  []Message{},
		input:     input,
		spinner:   sp,
		loading:   false,
		showHelp:  false,
		showTools: false,
	}
}

func (m Model) Init() tea.Cmd {
	m.messages = append(m.messages, Message{
		Role:    "system",
		Content: GetWelcomeScreen(),
		Time:    time.Now(),
	})
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.SetWidth(msg.Width - 4)

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Enter):
			if m.loading {
				return m, nil
			}
			text := m.input.Value()
			if text == "" {
				return m, nil
			}

			// Handle commands
			if strings.HasPrefix(text, "/") {
				return m, m.handleCommand(text)
			}

			m.messages = append(m.messages, Message{
				Role:    "user",
				Content: text,
				Time:    time.Now(),
			})
			m.input.Reset()
			m.loading = true

			return m, m.sendMessage(text)

		case key.Matches(msg, keys.Help):
			m.showHelp = !m.showHelp
			m.showTools = false

		case key.Matches(msg, keys.Tools):
			m.showTools = !m.showTools
			m.showHelp = false

		case key.Matches(msg, keys.Clear):
			m.messages = []Message{}
			m.showHelp = false
			m.showTools = false

		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Esc):
			m.showHelp = false
			m.showTools = false
			if m.loading {
				m.loading = false
			}
		}

	case types.StreamResponse:
		m.streamBuffer += msg.Content

	case error:
		m.err = msg
		m.loading = false

	default:
		if m.loading {
			m.spinner, _ = m.spinner.Update(msg)
		}
	}

	m.input, _ = m.input.Update(msg)
	return m, nil
}

func (m Model) handleCommand(text string) tea.Cmd {
	cmd := strings.TrimPrefix(text, "/")
	parts := strings.Fields(cmd)
	
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "help", "?":
		m.showHelp = !m.showHelp
		m.showTools = false
	case "tools":
		m.showTools = !m.showTools
		m.showHelp = false
	case "clear":
		m.messages = []Message{}
		m.showHelp = false
		m.showTools = false
	case "exit", "quit":
		return tea.Quit
	case "version":
		m.messages = append(m.messages, Message{
			Role:    "system",
			Content: "AI Terminal v1.0.0 - Enterprise AI Agent",
			Time:    time.Now(),
		})
	case "providers":
		m.messages = append(m.messages, Message{
			Role:    "system",
			Content: GetProvidersList(),
			Time:    time.Now(),
		})
	default:
		if len(parts) > 1 {
			m.messages = append(m.messages, Message{
				Role:    "system",
				Content: fmt.Sprintf("Unknown command: /%s", parts[0]),
				Time:    time.Now(),
			})
		}
	}
	
	return nil
}

func (m Model) View() string {
	var sb strings.Builder

	// Status bar
	sb.WriteString(GetStatusBar(m.provider.Name(), m.config.Provider.Model, true))
	sb.WriteString("\n\n")

	// Show help or tools if requested
	if m.showHelp {
		sb.WriteString(GetHelpScreen())
		sb.WriteString("\n\n")
	}

	if m.showTools {
		sb.WriteString(GetToolsList())
		sb.WriteString("\n\n")
	}

	// Messages
	for _, msg := range m.messages {
		sb.WriteString(FormatMessage(msg.Role, msg.Content, msg.Time))
		sb.WriteString("\n")
	}

	// Loading
	if m.loading {
		sb.WriteString(GetLoadingAnimation())
		sb.WriteString("\n")
	}

	// Input
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Render("➤ "))
	sb.WriteString(m.input.View())

	// Footer
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Render(" Press ? for help  |  /tools for 50+ tools  |  Ctrl+C to quit "))

	return sb.String()
}

func (m Model) sendMessage(text string) tea.Cmd {
	return func() tea.Msg {
		err := m.agent.Stream(context.Background(), text, func(chunk string) {
			m.streamBuffer += chunk
		})

		if err != nil {
			return err
		}

		if m.streamBuffer != "" {
			m.messages = append(m.messages, Message{
				Role:    "assistant",
				Content: m.streamBuffer,
				Time:    time.Now(),
			})
			m.streamBuffer = ""
		}

		m.loading = false
		return nil
	}
}

func GetProvidersList() string {
	var sb strings.Builder
	sb.WriteString(headerStyle.Render(" Available Providers "))
	sb.WriteString("\n\n")
	
	providers := []struct {
		Name    string
		Status  string
		Model   string
	}{
		{"Gemini", "✓ Available", "gemini-2.0-flash"},
		{"OpenAI", "✓ Available", "gpt-4o"},
		{"Anthropic", "✓ Available", "claude-3-5-sonnet"},
		{"Ollama", "✓ Available", "llama3 (local)"},
	}
	
	for _, p := range providers {
		sb.WriteString(fmt.Sprintf(" %s: %s (%s)\n", 
			lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(p.Name),
			p.Status,
			p.Model))
	}
	
	return panelStyle.Render(sb.String())
}

type KeyMap struct {
	Enter  key.Binding
	Help   key.Binding
	Tools  key.Binding
	Clear  key.Binding
	Quit   key.Binding
	Esc    key.Binding
}

var (
	keys = KeyMap{
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "send"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Tools: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "tools"),
		),
		Clear: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "clear"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "q"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close"),
		),
	}
)
