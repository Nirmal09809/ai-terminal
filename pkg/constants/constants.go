package constants

const (
	Version   = "1.0.0"
	Commit    = "dev"
	Date      = "2024-01-01"
	AppName   = "ai-terminal"
	AppDesc   = "Enterprise AI Terminal Agent"
)

const (
	DefaultMaxTokens    = 8192
	DefaultTemperature  = 0.7
	DefaultTimeout      = 30
	DefaultMaxRetries   = 3
	DefaultMaxHistory   = 100
	DefaultSlidingWindow = 10
)

const (
	ToolShell    = "shell"
	ToolRead     = "read"
	ToolWrite    = "write"
	ToolEdit     = "edit"
	ToolDelete   = "delete"
	ToolGlob     = "glob"
	ToolGrep     = "grep"
	ToolSearch   = "search"
	ToolFetch    = "fetch"
	ToolAsk      = "ask"
	ToolConfirm  = "confirm"
	ToolTodo     = "todo"
	ToolMemory   = "memory"
	ToolAnalyze  = "analyze"
	ToolExplain  = "explain"
	ToolReview   = "review"
	ToolMCP      = "mcp"
	ToolPlan     = "plan"
	ToolExitPlan = "exit_plan"
)

const (
	ProviderGemini   = "gemini"
	ProviderOpenAI   = "openai"
	ProviderAnthropic = "anthropic"
	ProviderOllama   = "ollama"
	ProviderVertex   = "vertex"
)
