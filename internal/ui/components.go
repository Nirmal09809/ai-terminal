package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	colorPrimary   = lipgloss.Color("86")    // Green
	colorSecondary = lipgloss.Color("205")    // Pink
	colorAccent    = lipgloss.Color("39")     // Blue
	colorWarning   = lipgloss.Color("226")    // Yellow
	colorError     = lipgloss.Color("196")    // Red
	colorSuccess   = lipgloss.Color("82")     // Bright Green
	colorMuted     = lipgloss.Color("245")     // Gray
	colorBg       = lipgloss.Color("236")     // Dark bg
	
	// Styles
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("63")).
			Padding(0, 2).
			Bold(true)
	
	panelStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)
	
	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("236")).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63"))
	
	userMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)
	
	assistantMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)
	
	systemMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Italic(true)
	
	errorMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
	
	toolStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Italic(true)
	
	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	
	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))
	
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Background(lipgloss.Color("236")).
			Padding(0, 1).
			Margin(0, 1)
	
	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	
	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2)
	
	statusOnline = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)
	
	statusOffline = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)

func GetWelcomeScreen() string {
	var sb strings.Builder
	
	sb.WriteString("\n")
	sb.WriteString(boxStyle.Render(fmt.Sprintf(`
%s

%s  AI Terminal - Your AI Assistant

%s
  • Multiple LLM Providers (Gemini, OpenAI, Claude, Ollama)
  • 50+ Built-in Tools  
  • Interactive TUI with Beautiful UI
  • Session Persistence
  • MCP Support
  • Plan Mode for Complex Tasks

%s
  Type your message and press %s to start!
  Press %s for help, %s to quit

`,
		logo(),
		headerStyle.Render(" Welcome to "),
		lipgloss.NewStyle().Foreground(colorAccent).Render("Features:"),
		lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render("Enter"),
		lipgloss.NewStyle().Foreground(colorPrimary).Render("?"),
		lipgloss.NewStyle().Foreground(colorPrimary).Render("Ctrl+C"),
	)))
	sb.WriteString("\n")
	
	return sb.String()
}

func logo() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("63")).
		Padding(1, 2).
		Render(fmt.Sprintf(`
   ██████╗ ████████╗
   ██╔══██╗╚══██╔══╝
   ██████╔╝   ██║   
   ██╔══██╗   ██║   
   ██║  ██║   ██║   
   ╚═╝  ╚═╝   ╚═╝   
   
   v1.0.0 - Enterprise AI Agent
`))
}

func GetHelpScreen() string {
	cols := []table.Column{
		{Title: "Key", Width: 12},
		{Title: "Action", Width: 25},
		{Title: "Description", Width: 40},
	}
	
	rows := []table.Row{
		{"Enter", "Send message", "Send your message to AI"},
		{"Ctrl+C", "Quit", "Exit the application"},
		{"Ctrl+L", "Clear", "Clear the screen"},
		{"?", "Help", "Show this help"},
		{"/ask", "Quick ask", "Quick question without TUI"},
		{"/session", "Sessions", "Manage sessions"},
		{"/config", "Config", "Edit configuration"},
		{"/provider", "Provider", "Switch LLM provider"},
		{"/tools", "Tools", "List all available tools"},
		{"/clear", "Clear", "Clear conversation"},
		{"/exit", "Exit", "Exit application"},
	}
	
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(false),
	)
	
	return panelStyle.Render(fmt.Sprintf(`
%s

%s

%s

%s
  Quick Commands:
  /ask <question>    - Ask without entering TUI
  /session list      - List all sessions
  /session resume   - Resume a session
  /config show      - Show config
  /provider gemini  - Switch to Gemini
  /provider openai  - Switch to OpenAI
  /tools            - List all 50+ tools
`,
		headerStyle.Render(" Keyboard Shortcuts "),
		t.View(),
		lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("Features:"),
		lipgloss.NewStyle().Foreground(colorMuted).Render("  • Multi-turn conversations"),
		lipgloss.NewStyle().Foreground(colorMuted).Render("  • File operations (read, write, edit, delete)"),
		lipgloss.NewStyle().Foreground(colorMuted).Render("  • Shell command execution"),
		lipgloss.NewStyle().Foreground(colorMuted).Render("  • Web search and fetch"),
		lipgloss.NewStyle().Foreground(colorMuted).Render("  • Code analysis and explanation"),
		lipgloss.NewStyle().Foreground(colorMuted).Render("  • Git and Docker operations"),
		lipgloss.NewStyle().Foreground(colorMuted).Render("  • System and network tools"),
		lipgloss.NewStyle().Foreground(colorMuted).Render("  • Todo and memory management"),
	))
}

func GetToolsList() string {
	tools := []struct {
		Name        string
		Category    string
		Description string
	}{
		// File Tools
		{"read", "File", "Read file contents"},
		{"write", "File", "Write content to file"},
		{"edit", "File", "Edit file (search/replace)"},
		{"delete", "File", "Delete file/directory"},
		{"mkdir", "File", "Create directory"},
		{"cp", "File", "Copy file"},
		{"mv", "File", "Move file"},
		{"chmod", "File", "Change permissions"},
		{"chown", "File", "Change owner"},
		
		// Search Tools
		{"glob", "Search", "Find files by pattern"},
		{"grep", "Search", "Search in files"},
		{"ripgrep", "Search", "Fast grep search"},
		{"find", "Search", "Find files"},
		{"locate", "Search", "Quick file search"},
		
		// Shell Tools
		{"shell", "Shell", "Execute shell commands"},
		{"bash", "Shell", "Run bash commands"},
		{"sudo", "Shell", "Run as superuser"},
		{"exec", "Shell", "Execute program"},
		
		// Git Tools
		{"git", "Git", "Git operations"},
		{"git_status", "Git", "Show git status"},
		{"git_log", "Git", "Show git log"},
		{"git_diff", "Git", "Show changes"},
		{"git_commit", "Git", "Create commit"},
		{"git_push", "Git", "Push to remote"},
		{"git_pull", "Git", "Pull from remote"},
		
		// Docker Tools
		{"docker", "Docker", "Docker operations"},
		{"docker_ps", "Docker", "List containers"},
		{"docker_images", "Docker", "List images"},
		{"docker_run", "Docker", "Run container"},
		{"docker_logs", "Docker", "View container logs"},
		
		// Web Tools
		{"search", "Web", "Web search"},
		{"fetch", "Web", "Fetch URL content"},
		{"scrape", "Web", "Scrape web page"},
		{"api", "Web", "Call API"},
		
		// System Tools
		{"system", "System", "System information"},
		{"cpu", "System", "CPU info"},
		{"memory", "System", "Memory info"},
		{"disk", "System", "Disk usage"},
		{"process", "System", "List processes"},
		{"kill", "System", "Kill process"},
		{"top", "System", "System monitor"},
		
		// Network Tools
		{"ping", "Network", "Ping host"},
		{"curl", "Network", "HTTP requests"},
		{"wget", "Network", "Download file"},
		{"nslookup", "Network", "DNS lookup"},
		{"netstat", "Network", "Network stats"},
		{"ssh", "Network", "SSH connection"},
		{"scp", "Network", "Secure copy"},
		
		// Development Tools
		{"npm", "Dev", "NPM operations"},
		{"pip", "Dev", "Pip operations"},
		{"go", "Dev", "Go operations"},
		{"cargo", "Dev", "Cargo operations"},
		{"pytest", "Dev", "Run tests"},
		{"build", "Dev", "Build project"},
		{"compile", "Dev", "Compile code"},
		
		// AI/ML Tools
		{"analyze", "AI/ML", "Analyze code"},
		{"explain", "AI/ML", "Explain code"},
		{"review", "AI/ML", "Code review"},
		{"refactor", "AI/ML", "Refactor code"},
		{"test_gen", "AI/ML", "Generate tests"},
		{"doc_gen", "AI/ML", "Generate docs"},
		
		// Utility Tools
		{"todo", "Utility", "Todo list"},
		{"note", "Utility", "Take notes"},
		{"reminder", "Utility", "Set reminder"},
		{"calc", "Utility", "Calculator"},
		{"convert", "Utility", "Convert units"},
		{"hash", "Utility", "Hash string"},
		{"encode", "Utility", "Encode text"},
		{"decode", "Utility", "Decode text"},
		
		// Database Tools
		{"sql", "Database", "Run SQL query"},
		{"mongo", "Database", "MongoDB query"},
		{"redis", "Database", "Redis commands"},
		
		// Cloud Tools
		{"aws", "Cloud", "AWS operations"},
		{"gcloud", "Cloud", "GCP operations"},
		{"kubectl", "Cloud", "Kubernetes"},
		{"terraform", "Cloud", "Terraform ops"},
		
		// MCP Tools
		{"mcp", "MCP", "MCP tools"},
		{"mcp_list", "MCP", "List MCP servers"},
		{"mcp_add", "MCP", "Add MCP server"},
	}
	
	var sb strings.Builder
	sb.WriteString(headerStyle.Render(" Available Tools (50+) "))
	sb.WriteString("\n\n")
	
	// Group by category
	byCategory := make(map[string][]string)
	for _, t := range tools {
		byCategory[t.Category] = append(byCategory[t.Category], fmt.Sprintf("%s: %s", 
			lipgloss.NewStyle().Foreground(colorPrimary).Render(t.Name),
			t.Description))
	}
	
	for category, items := range byCategory {
		sb.WriteString(lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render(fmt.Sprintf("%s:", category)))
		sb.WriteString("\n")
		for _, item := range items {
			sb.WriteString(fmt.Sprintf("  • %s\n", item))
		}
		sb.WriteString("\n")
	}
	
	return panelStyle.Render(sb.String())
}

func FormatMessage(role, content string, timestamp time.Time) string {
	var roleStr, contentStr string
	
	switch role {
	case "user":
		roleStr = userMsgStyle.Render("👤 You")
		contentStr = content
	case "assistant":
		roleStr = assistantMsgStyle.Render("🤖 AI")
		contentStr = content
	case "system":
		roleStr = systemMsgStyle.Render("⚙️ System")
		contentStr = content
	default:
		roleStr = toolStyle.Render("🔧 " + role)
		contentStr = content
	}
	
	return fmt.Sprintf("%s %s\n%s",
		roleStr,
		timeStyle.Render(timestamp.Format("15:04")),
		lipgloss.NewStyle().Foreground(colorMuted).Render(contentStr))
}

func GetStatusBar(provider, model string, connected bool) string {
	status := statusOnline.Render("✓ Connected")
	if !connected {
		status = statusOffline.Render("✗ Disconnected")
	}
	
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("63")).
		Padding(0, 2).
		Render(fmt.Sprintf(" %s | Provider: %s | Model: %s | %s ",
		status, provider, model, time.Now().Format("15:04")))
}

func GetLoadingAnimation() string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frame := frames[int(time.Now().Unix())%10]
	return spinnerStyle.Render(fmt.Sprintf(" %s Thinking...", frame))
}

func GetErrorMessage(err error) string {
	return errorMsgStyle.Render(fmt.Sprintf(" ❌ Error: %s", err.Error()))
}

func GetSuccessMessage(msg string) string {
	return lipgloss.NewStyle().
		Foreground(colorSuccess).
		Bold(true).
		Render(fmt.Sprintf(" ✓ %s", msg))
}
