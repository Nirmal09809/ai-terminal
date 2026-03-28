package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-terminal/internal/config"
	"ai-terminal/pkg/types"
)

type Provider interface {
	Name() string
	Generate(ctx context.Context, req types.ProviderRequest) (*types.ProviderResponse, error)
	Stream(ctx context.Context, req types.ProviderRequest, onChunk func(types.StreamResponse)) error
	GetTools() []types.ToolDefinition
}

type Factory struct {
	cfg *config.Config
}

func NewFactory(cfg *config.Config) *Factory {
	return &Factory{cfg: cfg}
}

func (f *Factory) Create(name string) (Provider, error) {
	switch name {
	case "gemini":
		return NewGeminiProvider(f.cfg.Provider.Gemini)
	case "openai":
		return NewOpenAIProvider(f.cfg.Provider.OpenAI)
	case "anthropic":
		return NewAnthropicProvider(f.cfg.Provider.Anthropic)
	case "ollama":
		return NewOllamaProvider(f.cfg.Provider.Ollama)
	default:
		return NewGeminiProvider(f.cfg.Provider.Gemini)
	}
}

func (f *Factory) GetAvailableProviders() []string {
	return []string{"gemini", "openai", "anthropic", "ollama"}
}

type GeminiProvider struct {
	apiKey      string
	model       string
	temperature float64
	maxTokens   int
}

func NewGeminiProvider(cfg config.GeminiConfig) (*GeminiProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("gemini api key is required")
	}
	if cfg.Model == "" {
		cfg.Model = "gemini-2.0-flash"
	}
	if cfg.Temperature == 0 {
		cfg.Temperature = 0.7
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 8192
	}
	return &GeminiProvider{
		apiKey:      cfg.APIKey,
		model:       cfg.Model,
		temperature: cfg.Temperature,
		maxTokens:   cfg.MaxTokens,
	}, nil
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) GetTools() []types.ToolDefinition {
	return []types.ToolDefinition{
		{Name: "shell", Description: "Execute shell commands in terminal", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"command": {Type: "string", Description: "The shell command to execute"}}, Required: []string{"command"}}},
		{Name: "read", Description: "Read contents of a file", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string", Description: "Path to the file to read"}}, Required: []string{"path"}}},
		{Name: "write", Description: "Write content to a file", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string", Description: "File path"}, "content": {Type: "string", Description: "Content to write"}}, Required: []string{"path", "content"}}},
		{Name: "edit", Description: "Edit a file with search and replace", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string", Description: "File path"}, "old_string": {Type: "string", Description: "Text to replace"}, "new_string": {Type: "string", Description: "Replacement text"}}, Required: []string{"path", "old_string"}}},
		{Name: "delete", Description: "Delete a file or directory", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string", Description: "Path to delete"}, "recursive": {Type: "boolean", Description: "Delete recursively"}}, Required: []string{"path"}}},
		{Name: "mkdir", Description: "Create a directory", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string", Description: "Directory path"}}, Required: []string{"path"}}},
		{Name: "glob", Description: "Find files matching a pattern", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"pattern": {Type: "string", Description: "Glob pattern (e.g. *.go)"}}, Required: []string{"pattern"}}},
		{Name: "grep", Description: "Search for text in files", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"pattern": {Type: "string", Description: "Text pattern to search"}, "path": {Type: "string", Description: "Directory to search"}}, Required: []string{"pattern"}}},
		{Name: "search", Description: "Search the web", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"query": {Type: "string", Description: "Search query"}}, Required: []string{"query"}}},
		{Name: "fetch", Description: "Fetch content from a URL", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"url": {Type: "string", Description: "URL to fetch"}}, Required: []string{"url"}}},
		{Name: "git_status", Description: "Show git status", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}},
		{Name: "git_log", Description: "Show git commit history", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"count": {Type: "string", Description: "Number of commits"}}}},
		{Name: "git_diff", Description: "Show git diff", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}},
		{Name: "docker_ps", Description: "List running containers", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}},
		{Name: "system", Description: "Get system information", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}},
		{Name: "memory", Description: "Get memory usage", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}},
		{Name: "disk", Description: "Get disk usage", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}},
		{Name: "process", Description: "List running processes", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}},
		{Name: "ping", Description: "Ping a host", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"host": {Type: "string", Description: "Host to ping"}}}},
		{Name: "todo", Description: "Manage todo list", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"action": {Type: "string", Description: "Action: add, list, done, delete"}, "task": {Type: "string", Description: "Task description"}}}},
		{Name: "note", Description: "Take notes", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"action": {Type: "string", Description: "Action: add, list, delete"}, "title": {Type: "string", Description: "Note title"}, "content": {Type: "string", Description: "Note content"}}}},
		{Name: "npm", Description: "Run npm commands", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"args": {Type: "string", Description: "npm arguments"}}}},
		{Name: "pip", Description: "Run pip commands", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"args": {Type: "string", Description: "pip arguments"}}}},
		{Name: "go", Description: "Run go commands", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"args": {Type: "string", Description: "go arguments"}}}},
	}
}

func (p *GeminiProvider) Generate(ctx context.Context, req types.ProviderRequest) (*types.ProviderResponse, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.model, p.apiKey)

	contents := make([]map[string]interface{}, 0)
	for _, msg := range req.Messages {
		role := "user"
		if msg.Role == types.RoleAssistant {
			role = "model"
		} else if msg.Role == types.RoleTool {
			continue
		}
		parts := []map[string]string{{"text": msg.Content}}
		contents = append(contents, map[string]interface{}{"role": role, "parts": parts})
	}

	tools := convertToGeminiTools(req.Tools)

	body := map[string]interface{}{
		"contents":          contents,
		"generationConfig": map[string]interface{}{
			"temperature":      p.temperature,
			"maxOutputTokens":  p.maxTokens,
			"tool":             tools,
		},
	}

	if len(tools) > 0 {
		body["tools"] = []map[string]interface{}{{"functionDeclarations": tools}}
	}

	jsonBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(data, &result)

	content := ""
	var toolCalls []types.ToolCall

	if candidates, ok := result["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if cand, ok := candidates[0].(map[string]interface{}); ok {
			if contentMap, ok := cand["content"].(map[string]interface{}); ok {
				if parts, ok := contentMap["parts"].([]interface{}); ok {
					for _, part := range parts {
						if p, ok := part.(map[string]interface{}); ok {
							if text, ok := p["text"].(string); ok {
								content += text
							}
							if fn, ok := p["functionCall"].(map[string]interface{}); ok {
								name, _ := fn["name"].(string)
								args, _ := fn["args"].(map[string]interface{})
								argsJSON, _ := json.Marshal(args)
								toolCalls = append(toolCalls, types.ToolCall{
									ID:   fmt.Sprintf("call_%d", len(toolCalls)),
									Type: "function",
									Function: types.ToolFunction{
										Name:      name,
										Arguments: argsJSON,
									},
								})
							}
						}
					}
				}
			}
		}
	}

	finishReason := "stop"
	if result["promptFeedback"] != nil {
		finishReason = "stop"
	}

	return &types.ProviderResponse{
		Content:     content,
		ToolCalls:   toolCalls,
		FinishReason: finishReason,
	}, nil
}

func (p *GeminiProvider) Stream(ctx context.Context, req types.ProviderRequest, onChunk func(types.StreamResponse)) error {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?key=%s&alt=sse", p.model, p.apiKey)

	contents := make([]map[string]interface{}, 0)
	for _, msg := range req.Messages {
		role := "user"
		if msg.Role == types.RoleAssistant {
			role = "model"
		}
		contents = append(contents, map[string]interface{}{
			"role":  role,
			"parts": []map[string]string{{"text": msg.Content}},
		})
	}

	body := map[string]interface{}{
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"temperature":     p.temperature,
			"maxOutputTokens": p.maxTokens,
		},
	}

	jsonBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				onChunk(types.StreamResponse{Done: true})
				break
			}
			var result map[string]interface{}
			json.Unmarshal([]byte(data), &result)
			if candidates, ok := result["candidates"].([]interface{}); ok && len(candidates) > 0 {
				if cand, ok := candidates[0].(map[string]interface{}); ok {
					if contentMap, ok := cand["content"].(map[string]interface{}); ok {
						if parts, ok := contentMap["parts"].([]interface{}); ok && len(parts) > 0 {
							if part, ok := parts[0].(map[string]interface{}); ok {
								if text, ok := part["text"].(string); ok {
									onChunk(types.StreamResponse{Content: text})
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func convertToGeminiTools(tools []types.ToolDefinition) []map[string]interface{} {
	if len(tools) == 0 {
		return nil
	}
	result := make([]map[string]interface{}, 0, len(tools))
	for _, t := range tools {
		result = append(result, map[string]interface{}{
			"name":        t.Name,
			"description": t.Description,
			"parameters": t.Parameters,
		})
	}
	return result
}

type OpenAIProvider struct {
	apiKey      string
	model       string
	baseURL     string
	temperature float64
	maxTokens   int
}

func NewOpenAIProvider(cfg config.OpenAIConfig) (*OpenAIProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openai api key is required")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4o"
	}
	return &OpenAIProvider{
		apiKey:      cfg.APIKey,
		model:       cfg.Model,
		baseURL:     cfg.BaseURL,
		temperature: cfg.Temperature,
		maxTokens:   cfg.MaxTokens,
	}, nil
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) GetTools() []types.ToolDefinition {
	return []types.ToolDefinition{
		{Name: "shell", Description: "Execute shell commands in terminal", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"command": {Type: "string", Description: "The shell command to execute"}}, Required: []string{"command"}}},
		{Name: "read", Description: "Read contents of a file", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string", Description: "Path to the file to read"}}, Required: []string{"path"}}},
		{Name: "write", Description: "Write content to a file", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string", Description: "File path"}, "content": {Type: "string", Description: "Content to write"}}, Required: []string{"path", "content"}}},
		{Name: "edit", Description: "Edit a file with search and replace", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string", Description: "File path"}, "old_string": {Type: "string", Description: "Text to replace"}, "new_string": {Type: "string", Description: "Replacement text"}}, Required: []string{"path", "old_string"}}},
		{Name: "delete", Description: "Delete a file or directory", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string", Description: "Path to delete"}, "recursive": {Type: "boolean", Description: "Delete recursively"}}, Required: []string{"path"}}},
		{Name: "mkdir", Description: "Create a directory", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string", Description: "Directory path"}}, Required: []string{"path"}}},
		{Name: "glob", Description: "Find files matching a pattern", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"pattern": {Type: "string", Description: "Glob pattern (e.g. *.go)"}}, Required: []string{"pattern"}}},
		{Name: "grep", Description: "Search for text in files", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"pattern": {Type: "string", Description: "Text pattern to search"}, "path": {Type: "string", Description: "Directory to search"}}, Required: []string{"pattern"}}},
		{Name: "search", Description: "Search the web", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"query": {Type: "string", Description: "Search query"}}, Required: []string{"query"}}},
		{Name: "fetch", Description: "Fetch content from a URL", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"url": {Type: "string", Description: "URL to fetch"}}, Required: []string{"url"}}},
	}
}

func (p *OpenAIProvider) Generate(ctx context.Context, req types.ProviderRequest) (*types.ProviderResponse, error) {
	url := p.baseURL + "/chat/completions"

	messages := make([]map[string]interface{}, 0)
	for _, msg := range req.Messages {
		messages = append(messages, map[string]interface{}{
			"role":    string(msg.Role),
			"content": msg.Content,
		})
	}

	body := map[string]interface{}{
		"model":       p.model,
		"messages":    messages,
		"temperature": p.temperature,
		"max_tokens":  p.maxTokens,
	}

	if len(req.Tools) > 0 {
		body["tools"] = convertToOpenAITools(req.Tools)
		body["tool_choice"] = "auto"
	}

	jsonBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(data, &result)

	content := ""
	var toolCalls []types.ToolCall

	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := choice["message"].(map[string]interface{}); ok {
				if c, ok := msg["content"].(string); ok {
					content = c
				}
				if tc, ok := msg["tool_calls"].([]interface{}); ok {
					for _, tcRaw := range tc {
						if tcMap, ok := tcRaw.(map[string]interface{}); ok {
							fn, _ := tcMap["function"].(map[string]interface{})
							name, _ := fn["name"].(string)
							argsStr, _ := fn["arguments"].(string)
							toolCalls = append(toolCalls, types.ToolCall{
								ID:   fmt.Sprintf("call_%d", len(toolCalls)),
								Type: "function",
								Function: types.ToolFunction{
									Name:      name,
									Arguments: json.RawMessage(argsStr),
								},
							})
						}
					}
				}
			}
			if fr, ok := choice["finish_reason"].(string); ok {
				if fr == "tool_calls" {
					// Continue with tool calls
				}
			}
		}
	}

	return &types.ProviderResponse{
		Content:     content,
		ToolCalls:   toolCalls,
		FinishReason: "stop",
	}, nil
}

func (p *OpenAIProvider) Stream(ctx context.Context, req types.ProviderRequest, onChunk func(types.StreamResponse)) error {
	url := p.baseURL + "/chat/completions"

	messages := make([]map[string]interface{}, 0)
	for _, msg := range req.Messages {
		messages = append(messages, map[string]interface{}{
			"role":    string(msg.Role),
			"content": msg.Content,
		})
	}

	body := map[string]interface{}{
		"model":       p.model,
		"messages":    messages,
		"temperature": p.temperature,
		"max_tokens":  p.maxTokens,
		"stream":      true,
	}

	if len(req.Tools) > 0 {
		body["tools"] = convertToOpenAITools(req.Tools)
	}

	jsonBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				onChunk(types.StreamResponse{Done: true})
				break
			}
			var chunk map[string]interface{}
			json.Unmarshal([]byte(data), &chunk)
			if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok {
							onChunk(types.StreamResponse{Content: content})
						}
					}
				}
			}
		}
	}
	return nil
}

func convertToOpenAITools(tools []types.ToolDefinition) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(tools))
	for _, t := range tools {
		result = append(result, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  t.Parameters,
			},
		})
	}
	return result
}

type AnthropicProvider struct {
	apiKey      string
	model       string
	temperature float64
	maxTokens   int
}

func NewAnthropicProvider(cfg config.AnthropicConfig) (*AnthropicProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("anthropic api key required")
	}
	if cfg.Model == "" {
		cfg.Model = "claude-3-5-sonnet-20241022"
	}
	return &AnthropicProvider{
		apiKey:      cfg.APIKey,
		model:       cfg.Model,
		temperature: cfg.Temperature,
		maxTokens:   cfg.MaxTokens,
	}, nil
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

func (p *AnthropicProvider) GetTools() []types.ToolDefinition {
	return []types.ToolDefinition{
		{Name: "shell", Description: "Execute shell commands in terminal", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"command": {Type: "string", Description: "The shell command to execute"}}, Required: []string{"command"}}},
		{Name: "read", Description: "Read contents of a file", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string", Description: "Path to the file to read"}}, Required: []string{"path"}}},
		{Name: "write", Description: "Write content to a file", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string", Description: "File path"}, "content": {Type: "string", Description: "Content to write"}}, Required: []string{"path", "content"}}},
	}
}

func (p *AnthropicProvider) Generate(ctx context.Context, req types.ProviderRequest) (*types.ProviderResponse, error) {
	url := "https://api.anthropic.com/v1/messages"

	system := ""
	messages := make([]map[string]interface{}, 0)
	for _, msg := range req.Messages {
		if msg.Role == types.RoleSystem {
			system = msg.Content
		} else {
			messages = append(messages, map[string]interface{}{
				"role":    string(msg.Role),
				"content": msg.Content,
			})
		}
	}

	body := map[string]interface{}{
		"model":       p.model,
		"messages":    messages,
		"max_tokens":  p.maxTokens,
		"temperature": p.temperature,
	}

	if system != "" {
		body["system"] = []string{system}
	}

	if len(req.Tools) > 0 {
		tools := make([]map[string]interface{}, 0, len(req.Tools))
		for _, t := range req.Tools {
			tools = append(tools, map[string]interface{}{
				"name":        t.Name,
				"description": t.Description,
				"input_schema": t.Parameters,
			})
		}
		body["tools"] = tools
	}

	jsonBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(data, &result)

	content := ""
	var toolCalls []types.ToolCall

	if contentSlice, ok := result["content"].([]interface{}); ok {
		for _, c := range contentSlice {
			if cMap, ok := c.(map[string]interface{}); ok {
				if cType, ok := cMap["type"].(string); ok {
					if cType == "text" {
						if text, ok := cMap["text"].(string); ok {
							content += text
						}
					} else if cType == "tool_use" {
						id, _ := cMap["id"].(string)
						name, _ := cMap["name"].(string)
						input, _ := cMap["input"].(map[string]interface{})
						inputJSON, _ := json.Marshal(input)
						toolCalls = append(toolCalls, types.ToolCall{
							ID:   id,
							Type: "function",
							Function: types.ToolFunction{
								Name:      name,
								Arguments: inputJSON,
							},
						})
					}
				}
			}
		}
	}

	return &types.ProviderResponse{
		Content:     content,
		ToolCalls:   toolCalls,
		FinishReason: "stop",
	}, nil
}

func (p *AnthropicProvider) Stream(ctx context.Context, req types.ProviderRequest, onChunk func(types.StreamResponse)) error {
	resp, err := p.Generate(ctx, req)
	if err != nil {
		return err
	}
	onChunk(types.StreamResponse{Content: resp.Content, Done: true})
	return nil
}

type OllamaProvider struct {
	url     string
	model   string
	timeout int
	client  *http.Client
}

func NewOllamaProvider(cfg config.OllamaConfig) (*OllamaProvider, error) {
	if cfg.URL == "" {
		cfg.URL = "http://localhost:11434"
	}
	if cfg.Model == "" {
		cfg.Model = "llama3"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 120
	}
	return &OllamaProvider{
		url:     cfg.URL,
		model:   cfg.Model,
		timeout: cfg.Timeout,
		client:  &http.Client{Timeout: time.Duration(cfg.Timeout) * time.Second},
	}, nil
}

func (p *OllamaProvider) Name() string { return "ollama" }

func (p *OllamaProvider) GetTools() []types.ToolDefinition {
	return []types.ToolDefinition{
		{Name: "shell", Description: "Execute shell commands", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"command": {Type: "string", Description: "Command"}}, Required: []string{"command"}}},
		{Name: "read", Description: "Read file", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string", Description: "Path"}}, Required: []string{"path"}}},
		{Name: "write", Description: "Write file", Parameters: types.ToolParameters{Type: "object", Properties: map[string]types.Property{"path": {Type: "string"}, "content": {Type: "string"}}, Required: []string{"path", "content"}}},
	}
}

func (p *OllamaProvider) Generate(ctx context.Context, req types.ProviderRequest) (*types.ProviderResponse, error) {
	messages := make([]map[string]interface{}, 0)
	for _, msg := range req.Messages {
		role := string(msg.Role)
		if role == "system" {
			role = "system"
		}
		messages = append(messages, map[string]interface{}{
			"role":    role,
			"content": msg.Content,
		})
	}

	body := map[string]interface{}{
		"model":   p.model,
		"messages": messages,
		"stream":  false,
	}

	jsonBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.url+"/api/chat", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(data, &result)

	content := ""
	if msg, ok := result["message"].(map[string]interface{}); ok {
		if c, ok := msg["content"].(string); ok {
			content = c
		}
	}

	return &types.ProviderResponse{
		Content:     content,
		FinishReason: "stop",
	}, nil
}

func (p *OllamaProvider) Stream(ctx context.Context, req types.ProviderRequest, onChunk func(types.StreamResponse)) error {
	messages := make([]map[string]interface{}, 0)
	for _, msg := range req.Messages {
		messages = append(messages, map[string]interface{}{
			"role":    string(msg.Role),
			"content": msg.Content,
		})
	}

	body := map[string]interface{}{
		"model":   p.model,
		"messages": messages,
		"stream":  true,
	}

	jsonBody, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.url+"/api/chat", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		var chunk map[string]interface{}
		json.Unmarshal([]byte(line), &chunk)
		if msg, ok := chunk["message"].(map[string]interface{}); ok {
			if content, ok := msg["content"].(string); ok {
				onChunk(types.StreamResponse{Content: content})
			}
		}
		if done, ok := chunk["done"].(bool); ok && done {
			onChunk(types.StreamResponse{Done: true})
			break
		}
	}
	return nil
}
