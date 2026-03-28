package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"ai-terminal/internal/config"
	"ai-terminal/internal/core/agent"
	"ai-terminal/internal/core/session"
	"ai-terminal/internal/providers"
	"ai-terminal/internal/tools/registry"
	"ai-terminal/internal/ui"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	version   = "1.0.0"
	commit    = "dev"
	date      = "2024-01-01"
)

var rootCmd = &cobra.Command{
	Use:   "ai-terminal",
	Short: "Enterprise AI Terminal Agent",
	Long: `AI Terminal - A powerful AI assistant for your terminal.

Features:
- Multiple LLM Providers (Gemini, OpenAI, Anthropic, Ollama)
- 15+ Built-in Tools
- Interactive TUI with Bubble Tea
- Session persistence
- MCP support
- Plan mode for complex tasks
- Code analysis and explanation
- Web search and fetch
- File operations

Examples:
  ai-terminal                    # Start interactive mode
  ai-terminal ask "Hello"       # Quick question
  ai-terminal config             # Create config
  ai-terminal --provider gemini  # Use specific provider`,
	Version: version,
	RunE:    runInteractive,
}

var (
	configPath string
	provider   string
	model      string
	verbose    bool
	noColor    bool
	apiKey     string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Config file path")
	rootCmd.PersistentFlags().StringVarP(&provider, "provider", "p", "", "LLM provider (gemini/openai/anthropic/ollama)")
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "", "Model name")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for the provider")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")

	rootCmd.AddCommand(nonInteractiveCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(sessionCmd)
}

var nonInteractiveCmd = &cobra.Command{
	Use:   "ask [prompt]",
	Short: "Ask AI without entering interactive mode",
	Long: `Ask AI a question without entering interactive mode.

Examples:
  ai-terminal ask "What is Go programming language?"
  ai-terminal ask -p gemini "Explain quantum computing"
  ai-terminal ask --api-key YOUR_KEY "Hello"`,
	Args: cobra.MinimumNArgs(1),
	RunE:  runNonInteractive,
}

var completionCmd = &cobra.Command{
	Use:   "completion [shell]",
	Short: "Generate shell completion scripts",
	Long: `Generate completion scripts for your shell.

Supported shells: bash, zsh, fish, powershell

Examples:
  ai-terminal completion bash > /etc/bash_completion.d/ai-terminal
  ai-terminal completion zsh > ~/.zsh/completions/_ai-terminal
  ai-terminal completion fish | source`,
	Args: cobra.ExactArgs(1),
	RunE:  runCompletion,
}

var configCmd = &cobra.Command{
	Use:   "config [action]",
	Short: "Manage configuration",
	Long: `Manage configuration file.

Actions:
  create  - Create default config
  edit    - Edit config in editor
  show    - Show current config
  reset   - Reset to defaults

Examples:
  ai-terminal config create
  ai-terminal config show
  ai-terminal config reset`,
	Args: cobra.MaximumNArgs(1),
	RunE:  runConfig,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Show version, build info, and dependencies",
	RunE:  runVersion,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start as HTTP server",
	Long: `Start AI Terminal as an HTTP server for remote access.

Examples:
  ai-terminal serve --port 8080
  ai-terminal serve --host 0.0.0.0 --port 8080 --api-key mykey`,
	RunE: runServe,
}

var shellCmd = &cobra.Command{
	Use:   "shell [command]",
	Short: "Execute a shell command through AI",
	Long: `Execute a shell command using AI assistance.

Examples:
  ai-terminal shell "list all files modified today"
  ai-terminal shell "find python files larger than 1MB"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runShell,
}

var sessionCmd = &cobra.Command{
	Use:   "session [action]",
	Short: "Manage sessions",
	Long: `Manage conversation sessions.

Actions:
  list    - List all sessions
  resume  - Resume a session
  delete - Delete a session

Examples:
  ai-terminal session list
  ai-terminal session resume session-id
  ai-terminal session delete session-id`,
	Args: cobra.MaximumNArgs(2),
	RunE: runSession,
}

func runInteractive(cmd *cobra.Command, args []string) error {
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if provider != "" {
		cfg.Provider.Default = provider
	}
	if model != "" {
		cfg.Provider.Model = model
	}
	if apiKey != "" {
		switch cfg.Provider.Default {
		case "gemini":
			cfg.Provider.Gemini.APIKey = apiKey
		case "openai":
			cfg.Provider.OpenAI.APIKey = apiKey
		case "anthropic":
			cfg.Provider.Anthropic.APIKey = apiKey
		}
	}

	providerFactory := providers.NewFactory(cfg)
	providerClient, err := providerFactory.Create(cfg.Provider.Default)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	toolRegistry := registry.New(cfg.Tools)
	if err := toolRegistry.LoadBuiltin(); err != nil {
		return fmt.Errorf("failed to load tools: %w", err)
	}

	sessionManager := session.NewManager()

	agentEngine := agent.NewEngine(agent.EngineConfig{
		Provider:       providerClient,
		ToolRegistry:   toolRegistry,
		SessionManager: sessionManager,
		ContextConfig:  cfg.Context,
	})

	appModel := ui.NewApp(ui.AppConfig{
		Agent:    agentEngine,
		Config:   cfg,
		Provider: providerClient,
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		os.Exit(0)
	}()

	p := tea.NewProgram(appModel)
	if err := p.Start(); err != nil {
		return fmt.Errorf("failed to start app: %w", err)
	}

	return nil
}

func runNonInteractive(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if provider != "" {
		cfg.Provider.Default = provider
	}
	if model != "" {
		cfg.Provider.Model = model
	}
	if apiKey != "" {
		switch cfg.Provider.Default {
		case "gemini":
			cfg.Provider.Gemini.APIKey = apiKey
		case "openai":
			cfg.Provider.OpenAI.APIKey = apiKey
		case "anthropic":
			cfg.Provider.Anthropic.APIKey = apiKey
		}
	}

	providerFactory := providers.NewFactory(cfg)
	providerClient, err := providerFactory.Create(cfg.Provider.Default)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	toolRegistry := registry.New(cfg.Tools)
	if err := toolRegistry.LoadBuiltin(); err != nil {
		return fmt.Errorf("failed to load tools: %w", err)
	}

	sessionManager := session.NewManager()

	agentEngine := agent.NewEngine(agent.EngineConfig{
		Provider:       providerClient,
		ToolRegistry:   toolRegistry,
		SessionManager: sessionManager,
		ContextConfig:  cfg.Context,
	})

	prompt := args[0]
	response, err := agentEngine.Run(context.Background(), prompt)
	if err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	fmt.Println(response)
	return nil
}

func runCompletion(cmd *cobra.Command, args []string) error {
	shell := args[0]
	switch shell {
	case "bash":
		rootCmd.GenBashCompletion(os.Stdout)
	case "zsh":
		rootCmd.GenZshCompletion(os.Stdout)
	case "fish":
		rootCmd.GenFishCompletion(os.Stdout, true)
	case "powershell":
		rootCmd.GenPowerShellCompletion(os.Stdout)
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
	}
	return nil
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfgPath := config.GetConfigPath()
	action := "show"
	if len(args) > 0 {
		action = args[0]
	}

	switch action {
	case "create":
		defaultConfig := config.Default()
		if err := config.Save(cfgPath, defaultConfig); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
		fmt.Printf("Created default config at: %s\n", cfgPath)

	case "show":
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		fmt.Printf("Config at: %s\n", cfgPath)
		fmt.Printf("Provider: %s\n", cfg.Provider.Default)
		fmt.Printf("Model: %s\n", cfg.Provider.Model)

	case "reset":
		os.Remove(cfgPath)
		defaultConfig := config.Default()
		if err := config.Save(cfgPath, defaultConfig); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
		fmt.Printf("Reset config to defaults at: %s\n", cfgPath)

	case "edit":
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		execCmd := exec.Command(editor, cfgPath)
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		return execCmd.Run()

	default:
		return fmt.Errorf("unknown action: %s (use: create, show, reset, edit)", action)
	}
	return nil
}

func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Printf("ai-terminal version %s\n", version)
	fmt.Printf("Build: %s\n", commit)
	fmt.Printf("Date: %s\n", date)
	fmt.Printf("\nFeatures:\n")
	fmt.Printf("  - Multiple LLM Providers (Gemini, OpenAI, Anthropic, Ollama)\n")
	fmt.Printf("  - 15+ Built-in Tools\n")
	fmt.Printf("  - Interactive TUI with Bubble Tea\n")
	fmt.Printf("  - Session Persistence\n")
	fmt.Printf("  - MCP Support\n")
	fmt.Printf("  - Plan Mode\n")
	fmt.Printf("  - Web Search & Fetch\n")
	fmt.Printf("  - Code Analysis\n")
	return nil
}

func runServe(cmd *cobra.Command, args []string) error {
	fmt.Println("HTTP server mode - Coming soon!")
	return nil
}

func runShell(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	if apiKey != "" {
		switch cfg.Provider.Default {
		case "gemini":
			cfg.Provider.Gemini.APIKey = apiKey
		case "openai":
			cfg.Provider.OpenAI.APIKey = apiKey
		case "anthropic":
			cfg.Provider.Anthropic.APIKey = apiKey
		}
	}

	providerFactory := providers.NewFactory(cfg)
	providerClient, _ := providerFactory.Create(cfg.Provider.Default)
	toolRegistry := registry.New(cfg.Tools)
	toolRegistry.LoadBuiltin()
	sessionManager := session.NewManager()

	agentEngine := agent.NewEngine(agent.EngineConfig{
		Provider:       providerClient,
		ToolRegistry:   toolRegistry,
		SessionManager: sessionManager,
		ContextConfig:  cfg.Context,
	})

	response, err := agentEngine.Run(context.Background(), "Execute shell command: "+args[0])
	if err != nil {
		return err
	}

	fmt.Println(response)
	return nil
}

func runSession(cmd *cobra.Command, args []string) error {
	manager := session.NewManager()

	if len(args) == 0 {
		sessions, err := manager.List()
		if err != nil {
			return err
		}
		fmt.Println("Sessions:")
		for _, s := range sessions {
			fmt.Printf("  %s - %s (%s)\n", s.ID, s.Title, s.UpdatedAt.Format("2006-01-02 15:04"))
		}
		return nil
	}

	action := args[0]
	if len(args) < 2 {
		return fmt.Errorf("session ID required")
	}

	switch action {
	case "delete":
		return manager.Delete(args[1])
	case "resume":
		sess, err := manager.Get(args[1])
		if err != nil {
			return err
		}
		fmt.Printf("Resuming session: %s\n", sess.Title)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
