package types

import (
	"encoding/json"
	"time"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

type Message struct {
	Role       Role       `json:"role" yaml:"role"`
	Content    string     `json:"content" yaml:"content"`
	Name       string     `json:"name,omitempty" yaml:"name,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty" yaml:"tool_call_id,omitempty"`
	ToolName   string     `json:"tool_name,omitempty" yaml:"tool_name,omitempty"`
	Timestamp  time.Time  `json:"timestamp" yaml:"timestamp"`
	Metadata   Metadata   `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type Metadata struct {
	Model      string            `json:"model,omitempty" yaml:"model,omitempty"`
	Tokens     int               `json:"tokens,omitempty" yaml:"tokens,omitempty"`
	ToolCalls  []json.RawMessage `json:"tool_calls,omitempty" yaml:"tool_calls,omitempty"`
	FinishReason string          `json:"finish_reason,omitempty" yaml:"finish_reason,omitempty"`
}

type ToolCall struct {
	ID       string          `json:"id" yaml:"id"`
	Type     string          `json:"type" yaml:"type"`
	Function ToolFunction    `json:"function" yaml:"function"`
}

type ToolFunction struct {
	Name      string          `json:"name" yaml:"name"`
	Arguments json.RawMessage `json:"arguments" yaml:"arguments"`
}

type ToolResult struct {
	ToolCallID string      `json:"tool_call_id" yaml:"tool_call_id"`
	ToolName   string      `json:"tool_name" yaml:"tool_name"`
	Result    interface{}  `json:"result" yaml:"result"`
	Error     string       `json:"error,omitempty" yaml:"error,omitempty"`
	Success   bool         `json:"success" yaml:"success"`
	Duration  time.Duration `json:"duration" yaml:"duration"`
}

type ToolDefinition struct {
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description" yaml:"description"`
	Parameters  ToolParameters `json:"parameters" yaml:"parameters"`
}

type ToolParameters struct {
	Type       string                `json:"type" yaml:"type"`
	Properties map[string]Property   `json:"properties" yaml:"properties"`
	Required   []string              `json:"required" yaml:"required"`
}

type Property struct {
	Type        string      `json:"type" yaml:"type"`
	Description string      `json:"description" yaml:"description"`
	Default     interface{} `json:"default,omitempty" yaml:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty" yaml:"enum,omitempty"`
}

type Session struct {
	ID        string    `json:"id" yaml:"id"`
	Title     string    `json:"title" yaml:"title"`
	Messages  []Message `json:"messages" yaml:"messages"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
	Metadata  SessionMetadata `json:"metadata" yaml:"metadata"`
}

type SessionMetadata struct {
	Provider string `json:"provider" yaml:"provider"`
	Model    string `json:"model" yaml:"model"`
}

type ProviderRequest struct {
	Model       string     `json:"model" yaml:"model"`
	Messages    []Message  `json:"messages" yaml:"messages"`
	Temperature float64    `json:"temperature" yaml:"temperature"`
	MaxTokens   int        `json:"max_tokens" yaml:"max_tokens"`
	Tools       []ToolDefinition `json:"tools,omitempty" yaml:"tools,omitempty"`
	Stream      bool       `json:"stream" yaml:"stream"`
}

type ProviderResponse struct {
	Content     string      `json:"content" yaml:"content"`
	ToolCalls   []ToolCall  `json:"tool_calls,omitempty" yaml:"tool_calls,omitempty"`
	FinishReason string     `json:"finish_reason" yaml:"finish_reason"`
	Tokens      int         `json:"tokens" yaml:"tokens"`
}

type StreamResponse struct {
	Content  string `json:"content" yaml:"content"`
	Done     bool   `json:"done" yaml:"done"`
}
