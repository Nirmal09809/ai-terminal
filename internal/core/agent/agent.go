package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ai-terminal/internal/config"
	"ai-terminal/internal/core/session"
	"ai-terminal/internal/tools/registry"
	"ai-terminal/pkg/types"
	"github.com/sirupsen/logrus"
)

const (
	MaxToolCalls = 10
)

type Engine struct {
	cfg            EngineConfig
	provider       Provider
	toolRegistry   *registry.Registry
	sessionManager *session.Manager
}

type EngineConfig struct {
	Provider       Provider
	ToolRegistry   *registry.Registry
	SessionManager *session.Manager
	ContextConfig  config.ContextConfig
}

type Provider interface {
	Name() string
	Generate(ctx context.Context, req types.ProviderRequest) (*types.ProviderResponse, error)
	Stream(ctx context.Context, req types.ProviderRequest, onChunk func(types.StreamResponse)) error
	GetTools() []types.ToolDefinition
}

func NewEngine(cfg EngineConfig) *Engine {
	return &Engine{
		cfg:            cfg,
		provider:       cfg.Provider,
		toolRegistry:   cfg.ToolRegistry,
		sessionManager: cfg.SessionManager,
	}
}

func (e *Engine) Run(ctx context.Context, userInput string) (string, error) {
	logrus.Info("Starting agent execution")

	messages := []types.Message{
		{
			Role:      types.RoleSystem,
			Content:   e.getSystemPrompt(),
			Timestamp: time.Now(),
		},
		{
			Role:      types.RoleUser,
			Content:   userInput,
			Timestamp: time.Now(),
		},
	}

	toolDefs := e.toolRegistry.GetDefinitions()

	for iteration := 0; iteration < MaxToolCalls; iteration++ {
		logrus.Infof("Iteration %d: Sending request to LLM", iteration+1)

		req := types.ProviderRequest{
			Model:       "default",
			Messages:    messages,
			Temperature: 0.7,
			MaxTokens:   4096,
			Tools:       toolDefs,
			Stream:      false,
		}

		resp, err := e.provider.Generate(ctx, req)
		if err != nil {
			return "", fmt.Errorf("failed to generate: %w", err)
		}

		logrus.Infof("Received response, content length: %d, tool calls: %d", len(resp.Content), len(resp.ToolCalls))

		// Check if we have tool calls
		if len(resp.ToolCalls) > 0 {
			logrus.Infof("Processing %d tool calls", len(resp.ToolCalls))

			// Add assistant message with tool calls
			assistantMsg := types.Message{
				Role:      types.RoleAssistant,
				Content:   resp.Content,
				Timestamp: time.Now(),
			}
			messages = append(messages, assistantMsg)

			// Execute each tool call
			for _, toolCall := range resp.ToolCalls {
				logrus.Infof("Executing tool: %s", toolCall.Function.Name)

				// Parse arguments
				var params map[string]interface{}
				if len(toolCall.Function.Arguments) > 0 {
					json.Unmarshal(toolCall.Function.Arguments, &params)
				}

				// Execute tool
				result, err := e.toolRegistry.Execute(toolCall.Function.Name, params)
				if err != nil {
					logrus.Errorf("Tool execution failed: %v", err)
					result = map[string]interface{}{
						"error": err.Error(),
					}
				}

				// Convert result to JSON string
				resultJSON, _ := json.Marshal(result)

				// Add tool result message
				toolResultMsg := types.Message{
					Role:       types.RoleTool,
					Content:    string(resultJSON),
					ToolCallID: toolCall.ID,
					ToolName:   toolCall.Function.Name,
					Timestamp:  time.Now(),
				}
				messages = append(messages, toolResultMsg)
			}

			// Continue to next iteration to get LLM response to tool results
			continue
		}

		// No tool calls, this is the final response
		if resp.Content != "" {
			assistantMsg := types.Message{
				Role:      types.RoleAssistant,
				Content:   resp.Content,
				Timestamp: time.Now(),
			}
			messages = append(messages, assistantMsg)
			return resp.Content, nil
		}

		// No content and no tool calls - something went wrong
		break
	}

	// Return the last assistant message if we hit max iterations
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == types.RoleAssistant && messages[i].Content != "" {
			return messages[i].Content, nil
		}
	}

	return "", fmt.Errorf("no response generated after %d iterations", MaxToolCalls)
}

func (e *Engine) RunWithSession(ctx context.Context, sessionID, userInput string) (string, error) {
	sess, err := e.sessionManager.GetOrCreate(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	if len(sess.Messages) == 0 {
		sess.Messages = append(sess.Messages, types.Message{
			Role:      types.RoleSystem,
			Content:   e.getSystemPrompt(),
			Timestamp: time.Now(),
		})
	}

	sess.Messages = append(sess.Messages, types.Message{
		Role:      types.RoleUser,
		Content:   userInput,
		Timestamp: time.Now(),
	})

	toolDefs := e.toolRegistry.GetDefinitions()

	for iteration := 0; iteration < MaxToolCalls; iteration++ {
		req := types.ProviderRequest{
			Model:       "default",
			Messages:    sess.Messages,
			Temperature: 0.7,
			MaxTokens:   4096,
			Tools:       toolDefs,
			Stream:      false,
		}

		resp, err := e.provider.Generate(ctx, req)
		if err != nil {
			return "", fmt.Errorf("failed to generate: %w", err)
		}

		if len(resp.ToolCalls) > 0 {
			sess.Messages = append(sess.Messages, types.Message{
				Role:      types.RoleAssistant,
				Content:   resp.Content,
				Timestamp: time.Now(),
			})

			for _, toolCall := range resp.ToolCalls {
				var params map[string]interface{}
				if len(toolCall.Function.Arguments) > 0 {
					json.Unmarshal(toolCall.Function.Arguments, &params)
				}

				result, err := e.toolRegistry.Execute(toolCall.Function.Name, params)
				if err != nil {
					result = map[string]interface{}{"error": err.Error()}
				}

				resultJSON, _ := json.Marshal(result)

				sess.Messages = append(sess.Messages, types.Message{
					Role:       types.RoleTool,
					Content:    string(resultJSON),
					ToolCallID: toolCall.ID,
					ToolName:   toolCall.Function.Name,
					Timestamp:  time.Now(),
				})
			}
			continue
		}

		if resp.Content != "" {
			sess.Messages = append(sess.Messages, types.Message{
				Role:      types.RoleAssistant,
				Content:   resp.Content,
				Timestamp: time.Now(),
			})
			e.sessionManager.Save(sess)
			return resp.Content, nil
		}
		break
	}

	e.sessionManager.Save(sess)

	for i := len(sess.Messages) - 1; i >= 0; i-- {
		if sess.Messages[i].Role == types.RoleAssistant && sess.Messages[i].Content != "" {
			return sess.Messages[i].Content, nil
		}
	}

	return "", nil
}

func (e *Engine) Stream(ctx context.Context, userInput string, onChunk func(string)) error {
	messages := []types.Message{
		{
			Role:      types.RoleSystem,
			Content:   e.getSystemPrompt(),
			Timestamp: time.Now(),
		},
		{
			Role:      types.RoleUser,
			Content:   userInput,
			Timestamp: time.Now(),
		},
	}

	toolDefs := e.toolRegistry.GetDefinitions()

	for iteration := 0; iteration < MaxToolCalls; iteration++ {
		req := types.ProviderRequest{
			Model:       "default",
			Messages:    messages,
			Temperature: 0.7,
			MaxTokens:   4096,
			Tools:       toolDefs,
			Stream:      true,
		}

		var fullContent strings.Builder
		err := e.provider.Stream(ctx, req, func(chunk types.StreamResponse) {
			if chunk.Content != "" {
				fullContent.WriteString(chunk.Content)
				onChunk(chunk.Content)
			}
		})
		if err != nil {
			return fmt.Errorf("failed to stream: %w", err)
		}

		// Check for tool calls in the response - need to make a non-streaming request
		req.Stream = false
		resp, err := e.provider.Generate(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to check tool calls: %w", err)
		}

		if len(resp.ToolCalls) > 0 {
			messages = append(messages, types.Message{
				Role:      types.RoleAssistant,
				Content:   fullContent.String(),
				Timestamp: time.Now(),
			})

			for _, toolCall := range resp.ToolCalls {
				var params map[string]interface{}
				if len(toolCall.Function.Arguments) > 0 {
					json.Unmarshal(toolCall.Function.Arguments, &params)
				}

				result, err := e.toolRegistry.Execute(toolCall.Function.Name, params)
				if err != nil {
					result = map[string]interface{}{"error": err.Error()}
				}

				resultJSON, _ := json.Marshal(result)
				messages = append(messages, types.Message{
					Role:       types.RoleTool,
					Content:    string(resultJSON),
					ToolCallID: toolCall.ID,
					Timestamp:  time.Now(),
				})
			}
			onChunk("\n\n")
			continue
		}

		break
	}

	return nil
}

func (e *Engine) GetProvider() Provider {
	return e.provider
}

func (e *Engine) GetToolRegistry() *registry.Registry {
	return e.toolRegistry
}

func (e *Engine) getSystemPrompt() string {
	return `You are an AI assistant with access to various tools to help users with their tasks.

You have access to the following tools:
- shell: Execute shell commands in terminal
- read: Read contents of a file
- write: Write content to a file (creates new or overwrites)
- edit: Edit a file with search and replace
- delete: Delete a file or directory
- mkdir: Create a directory
- cp: Copy files
- mv: Move/rename files
- chmod: Change file permissions
- stat: Get file information
- glob: Find files matching a pattern
- grep: Search for text in files
- find: Find files recursively
- search: Search the web
- fetch: Fetch content from URLs
- git: Git operations
- git_status: Show git status
- git_log: Show git history
- git_diff: Show changes
- git_commit: Create commit
- git_push: Push to remote
- git_pull: Pull from remote
- docker: Docker operations
- docker_ps: List containers
- docker_images: List images
- docker_run: Run container
- docker_logs: View container logs
- system: Get system information
- cpu: CPU information
- memory: Memory usage
- disk: Disk usage
- process: List processes
- top: System monitor
- uptime: System uptime
- whoami: Current user
- ping: Ping a host
- nslookup: DNS lookup
- netstat: Network connections
- wget: Download files
- npm: NPM operations
- pip: Pip operations
- go: Go operations
- cargo: Cargo operations
- build: Build project
- test: Run tests
- todo: Manage todo list
- note: Take notes
- calc: Calculator
- hash: Hash string
- encode: Encode text (base64)
- decode: Decode text
- date: Date/time

Guidelines:
1. Use tools when needed to complete tasks
2. Always read files before editing
3. Use shell for running commands
4. Be careful with destructive operations (delete, rm -rf, etc.)
5. Ask user for clarification when needed
6. Provide clear explanations of what you're doing

When you need to execute a tool, use the tool calls format.`
}
