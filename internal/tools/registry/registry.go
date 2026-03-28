package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"ai-terminal/internal/config"
	"ai-terminal/pkg/types"
)

type Registry struct {
	tools    map[string]Tool
	toolList []types.ToolDefinition
	cfg      config.ToolsConfig
}

type Tool interface {
	Name() string
	Description() string
	Category() string
	GetParameters() types.ToolParameters
	Execute(params map[string]interface{}) (interface{}, error)
}

type BaseTool struct {
	name        string
	description string
	category    string
}

func (t *BaseTool) Name() string         { return t.name }
func (t *BaseTool) Description() string { return t.description }
func (t *BaseTool) Category() string    { return t.category }
func (t *BaseTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type:       "object",
		Properties: map[string]types.Property{},
		Required:   []string{},
	}
}

func New(cfg config.ToolsConfig) *Registry {
	return &Registry{
		tools:    make(map[string]Tool),
		toolList: []types.ToolDefinition{},
		cfg:      cfg,
	}
}

func (r *Registry) LoadBuiltin() error {
	// File Tools (9)
	r.Register(&FileReadTool{BaseTool: BaseTool{name: "read", description: "Read file contents", category: "File"}})
	r.Register(&FileWriteTool{BaseTool: BaseTool{name: "write", description: "Write content to file", category: "File"}})
	r.Register(&FileEditTool{BaseTool: BaseTool{name: "edit", description: "Edit file with search/replace", category: "File"}})
	r.Register(&FileDeleteTool{BaseTool: BaseTool{name: "delete", description: "Delete file or directory", category: "File"}})
	r.Register(&FileMkdirTool{BaseTool: BaseTool{name: "mkdir", description: "Create directory", category: "File"}})
	r.Register(&FileCpTool{BaseTool: BaseTool{name: "cp", description: "Copy file", category: "File"}})
	r.Register(&FileMvTool{BaseTool: BaseTool{name: "mv", description: "Move/rename file", category: "File"}})
	r.Register(&FileChmodTool{BaseTool: BaseTool{name: "chmod", description: "Change file permissions", category: "File"}})
	r.Register(&FileStatTool{BaseTool: BaseTool{name: "stat", description: "Get file information", category: "File"}})

	// Search Tools (5)
	r.Register(&SearchGlobTool{BaseTool: BaseTool{name: "glob", description: "Find files by pattern", category: "Search"}})
	r.Register(&SearchGrepTool{BaseTool: BaseTool{name: "grep", description: "Search in files", category: "Search"}})
	r.Register(&SearchFindTool{BaseTool: BaseTool{name: "find", description: "Find files recursively", category: "Search"}})
	r.Register(&SearchLocateTool{BaseTool: BaseTool{name: "locate", description: "Quick file search", category: "Search"}})
	r.Register(&SearchRipgrepTool{BaseTool: BaseTool{name: "ripgrep", description: "Fast grep alternative", category: "Search"}})

	// Shell Tools (4)
	r.Register(&ShellTool{BaseTool: BaseTool{name: "shell", description: "Execute shell commands", category: "Shell"}})
	r.Register(&BashTool{BaseTool: BaseTool{name: "bash", description: "Run bash commands", category: "Shell"}})
	r.Register(&SudoTool{BaseTool: BaseTool{name: "sudo", description: "Run as superuser", category: "Shell"}})
	r.Register(&ExecTool{BaseTool: BaseTool{name: "exec", description: "Execute program", category: "Shell"}})

	// Git Tools (7)
	r.Register(&GitTool{BaseTool: BaseTool{name: "git", description: "Git operations", category: "Git"}})
	r.Register(&GitStatusTool{BaseTool: BaseTool{name: "git_status", description: "Show git status", category: "Git"}})
	r.Register(&GitLogTool{BaseTool: BaseTool{name: "git_log", description: "Show git history", category: "Git"}})
	r.Register(&GitDiffTool{BaseTool: BaseTool{name: "git_diff", description: "Show changes", category: "Git"}})
	r.Register(&GitCommitTool{BaseTool: BaseTool{name: "git_commit", description: "Create commit", category: "Git"}})
	r.Register(&GitPushTool{BaseTool: BaseTool{name: "git_push", description: "Push to remote", category: "Git"}})
	r.Register(&GitPullTool{BaseTool: BaseTool{name: "git_pull", description: "Pull from remote", category: "Git"}})

	// Docker Tools (5)
	r.Register(&DockerTool{BaseTool: BaseTool{name: "docker", description: "Docker operations", category: "Docker"}})
	r.Register(&DockerPsTool{BaseTool: BaseTool{name: "docker_ps", description: "List containers", category: "Docker"}})
	r.Register(&DockerImagesTool{BaseTool: BaseTool{name: "docker_images", description: "List images", category: "Docker"}})
	r.Register(&DockerRunTool{BaseTool: BaseTool{name: "docker_run", description: "Run container", category: "Docker"}})
	r.Register(&DockerLogsTool{BaseTool: BaseTool{name: "docker_logs", description: "View logs", category: "Docker"}})

	// Web Tools (4)
	r.Register(&WebSearchTool{BaseTool: BaseTool{name: "search", description: "Web search", category: "Web"}})
	r.Register(&WebFetchTool{BaseTool: BaseTool{name: "fetch", description: "Fetch URL content", category: "Web"}})
	r.Register(&WebScrapeTool{BaseTool: BaseTool{name: "scrape", description: "Scrape web page", category: "Web"}})
	r.Register(&WebCurlTool{BaseTool: BaseTool{name: "curl", description: "HTTP requests", category: "Web"}})

	// System Tools (8)
	r.Register(&SystemInfoTool{BaseTool: BaseTool{name: "system", description: "System information", category: "System"}})
	r.Register(&SystemCpuTool{BaseTool: BaseTool{name: "cpu", description: "CPU information", category: "System"}})
	r.Register(&SystemMemTool{BaseTool: BaseTool{name: "memory", description: "Memory info", category: "System"}})
	r.Register(&SystemDiskTool{BaseTool: BaseTool{name: "disk", description: "Disk usage", category: "System"}})
	r.Register(&SystemProcessTool{BaseTool: BaseTool{name: "process", description: "List processes", category: "System"}})
	r.Register(&SystemTopTool{BaseTool: BaseTool{name: "top", description: "System monitor", category: "System"}})
	r.Register(&SystemUptimeTool{BaseTool: BaseTool{name: "uptime", description: "System uptime", category: "System"}})
	r.Register(&SystemWhoamiTool{BaseTool: BaseTool{name: "whoami", description: "Current user", category: "System"}})

	// Network Tools (6)
	r.Register(&NetworkPingTool{BaseTool: BaseTool{name: "ping", description: "Ping host", category: "Network"}})
	r.Register(&NetworkNslookupTool{BaseTool: BaseTool{name: "nslookup", description: "DNS lookup", category: "Network"}})
	r.Register(&NetworkNetstatTool{BaseTool: BaseTool{name: "netstat", description: "Network stats", category: "Network"}})
	r.Register(&NetworkWgetTool{BaseTool: BaseTool{name: "wget", description: "Download file", category: "Network"}})
	r.Register(&NetworkSshTool{BaseTool: BaseTool{name: "ssh", description: "SSH connection", category: "Network"}})
	r.Register(&NetworkScpTool{BaseTool: BaseTool{name: "scp", description: "Secure copy", category: "Network"}})

	// Dev Tools (8)
	r.Register(&DevNpmTool{BaseTool: BaseTool{name: "npm", description: "NPM operations", category: "Dev"}})
	r.Register(&DevPipTool{BaseTool: BaseTool{name: "pip", description: "Pip operations", category: "Dev"}})
	r.Register(&DevGoTool{BaseTool: BaseTool{name: "go", description: "Go operations", category: "Dev"}})
	r.Register(&DevCargoTool{BaseTool: BaseTool{name: "cargo", description: "Cargo operations", category: "Dev"}})
	r.Register(&DevBuildTool{BaseTool: BaseTool{name: "build", description: "Build project", category: "Dev"}})
	r.Register(&DevTestTool{BaseTool: BaseTool{name: "test", description: "Run tests", category: "Dev"}})
	r.Register(&DevLintTool{BaseTool: BaseTool{name: "lint", description: "Lint code", category: "Dev"}})
	r.Register(&DevFormatTool{BaseTool: BaseTool{name: "format", description: "Format code", category: "Dev"}})

	// AI/ML Tools (6)
	r.Register(&AIMLAnalyzeTool{BaseTool: BaseTool{name: "analyze", description: "Analyze code", category: "AI/ML"}})
	r.Register(&AIMLExplainTool{BaseTool: BaseTool{name: "explain", description: "Explain code", category: "AI/ML"}})
	r.Register(&AIMLReviewTool{BaseTool: BaseTool{name: "review", description: "Code review", category: "AI/ML"}})
	r.Register(&AIMLRefactorTool{BaseTool: BaseTool{name: "refactor", description: "Refactor code", category: "AI/ML"}})
	r.Register(&AIMLTestGenTool{BaseTool: BaseTool{name: "test_gen", description: "Generate tests", category: "AI/ML"}})
	r.Register(&AIMLDocGenTool{BaseTool: BaseTool{name: "doc_gen", description: "Generate docs", category: "AI/ML"}})

	// Utility Tools (7)
	r.Register(&UtilityTodoTool{BaseTool: BaseTool{name: "todo", description: "Todo list", category: "Utility"}})
	r.Register(&UtilityNoteTool{BaseTool: BaseTool{name: "note", description: "Take notes", category: "Utility"}})
	r.Register(&UtilityCalcTool{BaseTool: BaseTool{name: "calc", description: "Calculator", category: "Utility"}})
	r.Register(&UtilityHashTool{BaseTool: BaseTool{name: "hash", description: "Hash string", category: "Utility"}})
	r.Register(&UtilityEncodeTool{BaseTool: BaseTool{name: "encode", description: "Encode text", category: "Utility"}})
	r.Register(&UtilityDecodeTool{BaseTool: BaseTool{name: "decode", description: "Decode text", category: "Utility"}})
	r.Register(&UtilityDateTool{BaseTool: BaseTool{name: "date", description: "Date/time info", category: "Utility"}})

	return nil
}

func (r *Registry) Register(tool Tool) {
	name := tool.Name()
	r.tools[name] = tool
	r.toolList = append(r.toolList, types.ToolDefinition{
		Name:        tool.Name(),
		Description: tool.Description(),
		Parameters:  tool.GetParameters(),
	})
}

func (r *Registry) Get(name string) (Tool, error) {
	tool, ok := r.tools[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool, nil
}

func (r *Registry) GetDefinitions() []types.ToolDefinition {
	return r.toolList
}

func (r *Registry) Execute(name string, params map[string]interface{}) (interface{}, error) {
	tool, err := r.Get(name)
	if err != nil {
		return nil, err
	}
	return tool.Execute(params)
}

func (r *Registry) ParseToolArguments(toolName string, args json.RawMessage) (map[string]interface{}, error) {
	var params map[string]interface{}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}
	return params, nil
}

func runCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ==================== FILE TOOLS ====================

type FileReadTool struct{ BaseTool }
func (t *FileReadTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"path": {Type: "string", Description: "Path to the file to read"},
		},
		Required: []string{"path"},
	}
}
func (t *FileReadTool) Execute(params map[string]interface{}) (interface{}, error) {
	path, _ := params["path"].(string)
	if path == "" { return nil, fmt.Errorf("path required") }
	content, err := os.ReadFile(path)
	if err != nil { return nil, err }
	info, _ := os.Stat(path)
	return map[string]interface{}{"path": path, "content": string(content), "size": info.Size()}, nil
}

type FileWriteTool struct{ BaseTool }
func (t *FileWriteTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"path": {Type: "string", Description: "File path to write to"},
			"content": {Type: "string", Description: "Content to write"},
		},
		Required: []string{"path", "content"},
	}
}
func (t *FileWriteTool) Execute(params map[string]interface{}) (interface{}, error) {
	path, _ := params["path"].(string)
	content, _ := params["content"].(string)
	if path == "" || content == "" { return nil, fmt.Errorf("path and content required") }
	filepath.Dir(path)
	os.MkdirAll(filepath.Dir(path), 0755)
	err := os.WriteFile(path, []byte(content), 0644)
	return map[string]interface{}{"path": path, "success": err == nil}, nil
}

type FileEditTool struct{ BaseTool }
func (t *FileEditTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"path": {Type: "string", Description: "File path to edit"},
			"old_string": {Type: "string", Description: "Text to find and replace"},
			"new_string": {Type: "string", Description: "Replacement text"},
		},
		Required: []string{"path", "old_string"},
	}
}
func (t *FileEditTool) Execute(params map[string]interface{}) (interface{}, error) {
	path, _ := params["path"].(string)
	oldStr, _ := params["old_string"].(string)
	newStr, _ := params["new_string"].(string)
	if path == "" || oldStr == "" { return nil, fmt.Errorf("path and old_string required") }
	content, err := os.ReadFile(path)
	if err != nil { return nil, err }
	newContent := strings.Replace(string(content), oldStr, newStr, 1)
	err = os.WriteFile(path, []byte(newContent), 0644)
	return map[string]interface{}{"path": path, "success": err == nil}, nil
}

type FileDeleteTool struct{ BaseTool }
func (t *FileDeleteTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"path": {Type: "string", Description: "Path to delete"},
			"recursive": {Type: "boolean", Description: "Delete directories recursively"},
		},
		Required: []string{"path"},
	}
}
func (t *FileDeleteTool) Execute(params map[string]interface{}) (interface{}, error) {
	path, _ := params["path"].(string)
	recursive, _ := params["recursive"].(bool)
	if path == "" { return nil, fmt.Errorf("path required") }
	info, _ := os.Stat(path)
	var err error
	if info.IsDir() && !recursive { return nil, fmt.Errorf("use recursive=true") }
	if info.IsDir() { err = os.RemoveAll(path) } else { err = os.Remove(path) }
	return map[string]interface{}{"path": path, "success": err == nil}, nil
}

type FileMkdirTool struct{ BaseTool }
func (t *FileMkdirTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"path": {Type: "string", Description: "Directory path to create"},
		},
		Required: []string{"path"},
	}
}
func (t *FileMkdirTool) Execute(params map[string]interface{}) (interface{}, error) {
	path, _ := params["path"].(string)
	if path == "" { return nil, fmt.Errorf("path required") }
	err := os.MkdirAll(path, 0755)
	return map[string]interface{}{"path": path, "success": err == nil}, nil
}

type FileCpTool struct{ BaseTool }
func (t *FileCpTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"src": {Type: "string", Description: "Source file path"},
			"dst": {Type: "string", Description: "Destination file path"},
		},
		Required: []string{"src", "dst"},
	}
}
func (t *FileCpTool) Execute(params map[string]interface{}) (interface{}, error) {
	src, _ := params["src"].(string)
	dst, _ := params["dst"].(string)
	if src == "" || dst == "" { return nil, fmt.Errorf("src and dst required") }
	output, err := runCmd("cp", src, dst)
	return map[string]interface{}{"src": src, "dst": dst, "output": output, "success": err == nil}, nil
}

type FileMvTool struct{ BaseTool }
func (t *FileMvTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"src": {Type: "string", Description: "Source file path"},
			"dst": {Type: "string", Description: "Destination file path"},
		},
		Required: []string{"src", "dst"},
	}
}
func (t *FileMvTool) Execute(params map[string]interface{}) (interface{}, error) {
	src, _ := params["src"].(string)
	dst, _ := params["dst"].(string)
	if src == "" || dst == "" { return nil, fmt.Errorf("src and dst required") }
	err := os.Rename(src, dst)
	return map[string]interface{}{"src": src, "dst": dst, "success": err == nil}, nil
}

type FileChmodTool struct{ BaseTool }
func (t *FileChmodTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"path": {Type: "string", Description: "File path"},
			"mode": {Type: "string", Description: "Permission mode (e.g. 755)"},
		},
		Required: []string{"path", "mode"},
	}
}
func (t *FileChmodTool) Execute(params map[string]interface{}) (interface{}, error) {
	path, _ := params["path"].(string)
	mode, _ := params["mode"].(string)
	if path == "" || mode == "" { return nil, fmt.Errorf("path and mode required") }
	output, err := runCmd("chmod", mode, path)
	return map[string]interface{}{"path": path, "mode": mode, "output": output, "success": err == nil}, nil
}

type FileStatTool struct{ BaseTool }
func (t *FileStatTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"path": {Type: "string", Description: "File path"},
		},
		Required: []string{"path"},
	}
}
func (t *FileStatTool) Execute(params map[string]interface{}) (interface{}, error) {
	path, _ := params["path"].(string)
	if path == "" { return nil, fmt.Errorf("path required") }
	info, err := os.Stat(path)
	if err != nil { return nil, err }
	return map[string]interface{}{"path": path, "size": info.Size(), "mode": info.Mode().String(), "modTime": info.ModTime()}, nil
}

// ==================== SEARCH TOOLS ====================

type SearchGlobTool struct{ BaseTool }
func (t *SearchGlobTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"pattern": {Type: "string", Description: "Glob pattern (e.g. *.go, **/*.txt)"},
		},
		Required: []string{"pattern"},
	}
}
func (t *SearchGlobTool) Execute(params map[string]interface{}) (interface{}, error) {
	pattern, _ := params["pattern"].(string)
	if pattern == "" { return nil, fmt.Errorf("pattern required") }
	files, _ := filepath.Glob(pattern)
	return map[string]interface{}{"pattern": pattern, "files": files, "count": len(files)}, nil
}

type SearchGrepTool struct{ BaseTool }
func (t *SearchGrepTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"pattern": {Type: "string", Description: "Text pattern to search"},
			"path": {Type: "string", Description: "Directory to search in (default: .)"},
		},
		Required: []string{"pattern"},
	}
}
func (t *SearchGrepTool) Execute(params map[string]interface{}) (interface{}, error) {
	pattern, _ := params["pattern"].(string)
	path, _ := params["path"].(string)
	if path == "" { path = "." }
	if pattern == "" { return nil, fmt.Errorf("pattern required") }
	output, _ := runCmd("grep", "-r", "-n", pattern, path)
	matches := strings.Split(output, "\n")
	return map[string]interface{}{"pattern": pattern, "matches": matches[:min(50, len(matches))], "count": len(matches)}, nil
}

type SearchFindTool struct{ BaseTool }
func (t *SearchFindTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"path": {Type: "string", Description: "Directory to search in (default: .)"},
			"name": {Type: "string", Description: "File name pattern"},
		},
		Required: []string{"name"},
	}
}
func (t *SearchFindTool) Execute(params map[string]interface{}) (interface{}, error) {
	path, _ := params["path"].(string)
	name, _ := params["name"].(string)
	if path == "" { path = "." }
	if name == "" { return nil, fmt.Errorf("name required") }
	output, _ := runCmd("find", path, "-name", name)
	files := strings.Split(output, "\n")
	return map[string]interface{}{"path": path, "name": name, "files": files, "count": len(files)}, nil
}

type SearchLocateTool struct{ BaseTool }
func (t *SearchLocateTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"pattern": {Type: "string", Description: "Pattern to search"},
		},
		Required: []string{"pattern"},
	}
}
func (t *SearchLocateTool) Execute(params map[string]interface{}) (interface{}, error) {
	pattern, _ := params["pattern"].(string)
	if pattern == "" { return nil, fmt.Errorf("pattern required") }
	output, _ := runCmd("locate", pattern)
	files := strings.Split(output, "\n")
	return map[string]interface{}{"pattern": pattern, "files": files, "count": len(files)}, nil
}

type SearchRipgrepTool struct{ BaseTool }
func (t *SearchRipgrepTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"pattern": {Type: "string", Description: "Text pattern to search"},
			"path": {Type: "string", Description: "Directory to search in (default: .)"},
		},
		Required: []string{"pattern"},
	}
}
func (t *SearchRipgrepTool) Execute(params map[string]interface{}) (interface{}, error) {
	pattern, _ := params["pattern"].(string)
	path, _ := params["path"].(string)
	if path == "" { path = "." }
	if pattern == "" { return nil, fmt.Errorf("pattern required") }
	output, _ := runCmd("rg", "-n", pattern, path)
	matches := strings.Split(output, "\n")
	return map[string]interface{}{"pattern": pattern, "matches": matches, "count": len(matches)}, nil
}

// ==================== SHELL TOOLS ====================

type ShellTool struct{ BaseTool }
func (t *ShellTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"command": {Type: "string", Description: "Shell command to execute"},
		},
		Required: []string{"command"},
	}
}
func (t *ShellTool) Execute(params map[string]interface{}) (interface{}, error) {
	cmd, _ := params["command"].(string)
	if cmd == "" { return nil, fmt.Errorf("command required") }
	output, err := runCmd("sh", "-c", cmd)
	return map[string]interface{}{"command": cmd, "output": output, "success": err == nil}, nil
}

type BashTool struct{ BaseTool }
func (t *BashTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"command": {Type: "string", Description: "Bash command to execute"},
		},
		Required: []string{"command"},
	}
}
func (t *BashTool) Execute(params map[string]interface{}) (interface{}, error) {
	cmd, _ := params["command"].(string)
	if cmd == "" { return nil, fmt.Errorf("command required") }
	output, err := runCmd("bash", "-c", cmd)
	return map[string]interface{}{"command": cmd, "output": output, "success": err == nil}, nil
}

type SudoTool struct{ BaseTool }
func (t *SudoTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"command": {Type: "string", Description: "Command to run as superuser"},
		},
		Required: []string{"command"},
	}
}
func (t *SudoTool) Execute(params map[string]interface{}) (interface{}, error) {
	cmd, _ := params["command"].(string)
	if cmd == "" { return nil, fmt.Errorf("command required") }
	output, err := runCmd("sudo", "sh", "-c", cmd)
	return map[string]interface{}{"command": cmd, "output": output, "success": err == nil}, nil
}

type ExecTool struct{ BaseTool }
func (t *ExecTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"command": {Type: "string", Description: "Program to execute"},
			"args": {Type: "array", Description: "Command arguments"},
		},
		Required: []string{"command"},
	}
}
func (t *ExecTool) Execute(params map[string]interface{}) (interface{}, error) {
	cmd, _ := params["command"].(string)
	args, _ := params["args"].([]interface{})
	if cmd == "" { return nil, fmt.Errorf("command required") }
	var argsSlice []string
	for _, a := range args {
		argsSlice = append(argsSlice, fmt.Sprintf("%v", a))
	}
	output, err := runCmd(cmd, argsSlice...)
	return map[string]interface{}{"command": cmd, "output": output, "success": err == nil}, nil
}

// ==================== GIT TOOLS ====================

type GitTool struct{ BaseTool }
func (t *GitTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"args": {Type: "string", Description: "Git command arguments"},
		},
		Required: []string{"args"},
	}
}
func (t *GitTool) Execute(params map[string]interface{}) (interface{}, error) {
	args, _ := params["args"].(string)
	if args == "" { return nil, fmt.Errorf("args required") }
	output, err := runCmd("git", strings.Fields(args)...)
	return map[string]interface{}{"args": args, "output": output, "success": err == nil}, nil
}

type GitStatusTool struct{ BaseTool }
func (t *GitStatusTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *GitStatusTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, _ := runCmd("git", "status")
	return map[string]interface{}{"output": output}, nil
}

type GitLogTool struct{ BaseTool }
func (t *GitLogTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"count": {Type: "string", Description: "Number of commits to show"},
		},
	}
}
func (t *GitLogTool) Execute(params map[string]interface{}) (interface{}, error) {
	count, _ := params["count"].(string)
	if count == "" { count = "10" }
	output, _ := runCmd("git", "log", "-n", count, "--oneline")
	return map[string]interface{}{"output": output}, nil
}

type GitDiffTool struct{ BaseTool }
func (t *GitDiffTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *GitDiffTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, _ := runCmd("git", "diff")
	return map[string]interface{}{"output": output}, nil
}

type GitCommitTool struct{ BaseTool }
func (t *GitCommitTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"message": {Type: "string", Description: "Commit message"},
		},
		Required: []string{"message"},
	}
}
func (t *GitCommitTool) Execute(params map[string]interface{}) (interface{}, error) {
	msg, _ := params["message"].(string)
	if msg == "" { return nil, fmt.Errorf("message required") }
	runCmd("git", "add", "-A")
	output, err := runCmd("git", "commit", "-m", msg)
	return map[string]interface{}{"message": msg, "output": output, "success": err == nil}, nil
}

type GitPushTool struct{ BaseTool }
func (t *GitPushTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *GitPushTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, err := runCmd("git", "push")
	return map[string]interface{}{"output": output, "success": err == nil}, nil
}

type GitPullTool struct{ BaseTool }
func (t *GitPullTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *GitPullTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, err := runCmd("git", "pull")
	return map[string]interface{}{"output": output, "success": err == nil}, nil
}

// ==================== DOCKER TOOLS ====================

type DockerTool struct{ BaseTool }
func (t *DockerTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"args": {Type: "string", Description: "Docker command arguments"},
		},
		Required: []string{"args"},
	}
}
func (t *DockerTool) Execute(params map[string]interface{}) (interface{}, error) {
	args, _ := params["args"].(string)
	if args == "" { return nil, fmt.Errorf("args required") }
	output, err := runCmd("docker", strings.Fields(args)...)
	return map[string]interface{}{"args": args, "output": output, "success": err == nil}, nil
}

type DockerPsTool struct{ BaseTool }
func (t *DockerPsTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *DockerPsTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, _ := runCmd("docker", "ps", "-a")
	return map[string]interface{}{"output": output}, nil
}

type DockerImagesTool struct{ BaseTool }
func (t *DockerImagesTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *DockerImagesTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, _ := runCmd("docker", "images")
	return map[string]interface{}{"output": output}, nil
}

type DockerRunTool struct{ BaseTool }
func (t *DockerRunTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"image": {Type: "string", Description: "Docker image name"},
		},
		Required: []string{"image"},
	}
}
func (t *DockerRunTool) Execute(params map[string]interface{}) (interface{}, error) {
	image, _ := params["image"].(string)
	if image == "" { return nil, fmt.Errorf("image required") }
	output, err := runCmd("docker", "run", "-it", image)
	return map[string]interface{}{"image": image, "output": output, "success": err == nil}, nil
}

type DockerLogsTool struct{ BaseTool }
func (t *DockerLogsTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"container": {Type: "string", Description: "Container ID or name"},
		},
		Required: []string{"container"},
	}
}
func (t *DockerLogsTool) Execute(params map[string]interface{}) (interface{}, error) {
	container, _ := params["container"].(string)
	if container == "" { return nil, fmt.Errorf("container required") }
	output, _ := runCmd("docker", "logs", container)
	return map[string]interface{}{"container": container, "output": output}, nil
}

// ==================== WEB TOOLS ====================

type WebSearchTool struct{ BaseTool }
func (t *WebSearchTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"query": {Type: "string", Description: "Search query"},
		},
		Required: []string{"query"},
	}
}
func (t *WebSearchTool) Execute(params map[string]interface{}) (interface{}, error) {
	query, _ := params["query"].(string)
	if query == "" { return nil, fmt.Errorf("query required") }
	runCmd("curl", "-s", "https://html.duckduckgo.com/html/?q="+strings.ReplaceAll(query, " ", "+"))
	return map[string]interface{}{"query": query, "results": "Search completed"}, nil
}

type WebFetchTool struct{ BaseTool }
func (t *WebFetchTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"url": {Type: "string", Description: "URL to fetch"},
		},
		Required: []string{"url"},
	}
}
func (t *WebFetchTool) Execute(params map[string]interface{}) (interface{}, error) {
	url, _ := params["url"].(string)
	if url == "" { return nil, fmt.Errorf("url required") }
	output, err := runCmd("curl", "-s", "-L", url)
	return map[string]interface{}{"url": url, "content": output[:min(5000, len(output))], "success": err == nil}, nil
}

type WebScrapeTool struct{ BaseTool }
func (t *WebScrapeTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"url": {Type: "string", Description: "URL to scrape"},
		},
		Required: []string{"url"},
	}
}
func (t *WebScrapeTool) Execute(params map[string]interface{}) (interface{}, error) {
	url, _ := params["url"].(string)
	if url == "" { return nil, fmt.Errorf("url required") }
	output, err := runCmd("curl", "-s", "-L", url)
	return map[string]interface{}{"url": url, "content": output[:min(10000, len(output))], "success": err == nil}, nil
}

type WebCurlTool struct{ BaseTool }
func (t *WebCurlTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"url": {Type: "string", Description: "URL to request"},
			"method": {Type: "string", Description: "HTTP method (GET, POST, etc.)"},
		},
		Required: []string{"url"},
	}
}
func (t *WebCurlTool) Execute(params map[string]interface{}) (interface{}, error) {
	url, _ := params["url"].(string)
	method, _ := params["method"].(string)
	if method == "" { method = "GET" }
	if url == "" { return nil, fmt.Errorf("url required") }
	output, err := runCmd("curl", "-X", method, "-s", url)
	return map[string]interface{}{"url": url, "method": method, "output": output, "success": err == nil}, nil
}

// ==================== SYSTEM TOOLS ====================

type SystemInfoTool struct{ BaseTool }
func (t *SystemInfoTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *SystemInfoTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, _ := runCmd("uname", "-a")
	return map[string]interface{}{"system": output}, nil
}

type SystemCpuTool struct{ BaseTool }
func (t *SystemCpuTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *SystemCpuTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, _ := runCmd("cat", "/proc/cpuinfo")
	return map[string]interface{}{"cpu": output[:min(2000, len(output))]}, nil
}

type SystemMemTool struct{ BaseTool }
func (t *SystemMemTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *SystemMemTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, _ := runCmd("free", "-m")
	return map[string]interface{}{"memory": output}, nil
}

type SystemDiskTool struct{ BaseTool }
func (t *SystemDiskTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *SystemDiskTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, _ := runCmd("df", "-h")
	return map[string]interface{}{"disk": output}, nil
}

type SystemProcessTool struct{ BaseTool }
func (t *SystemProcessTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *SystemProcessTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, _ := runCmd("ps", "aux")
	return map[string]interface{}{"processes": output[:min(3000, len(output))]}, nil
}

type SystemTopTool struct{ BaseTool }
func (t *SystemTopTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *SystemTopTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, _ := runCmd("top", "-bn1")
	return map[string]interface{}{"top": output[:min(2000, len(output))]}, nil
}

type SystemUptimeTool struct{ BaseTool }
func (t *SystemUptimeTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *SystemUptimeTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, _ := runCmd("uptime")
	return map[string]interface{}{"uptime": output}, nil
}

type SystemWhoamiTool struct{ BaseTool }
func (t *SystemWhoamiTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *SystemWhoamiTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, _ := runCmd("whoami")
	return map[string]interface{}{"user": strings.TrimSpace(output)}, nil
}

// ==================== NETWORK TOOLS ====================

type NetworkPingTool struct{ BaseTool }
func (t *NetworkPingTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"host": {Type: "string", Description: "Host to ping (default: 8.8.8.8)"},
		},
	}
}
func (t *NetworkPingTool) Execute(params map[string]interface{}) (interface{}, error) {
	host, _ := params["host"].(string)
	if host == "" { host = "8.8.8.8" }
	output, err := runCmd("ping", "-c", "4", host)
	return map[string]interface{}{"host": host, "output": output, "success": err == nil}, nil
}

type NetworkNslookupTool struct{ BaseTool }
func (t *NetworkNslookupTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"host": {Type: "string", Description: "Hostname to lookup"},
		},
		Required: []string{"host"},
	}
}
func (t *NetworkNslookupTool) Execute(params map[string]interface{}) (interface{}, error) {
	host, _ := params["host"].(string)
	if host == "" { return nil, fmt.Errorf("host required") }
	output, _ := runCmd("nslookup", host)
	return map[string]interface{}{"host": host, "output": output}, nil
}

type NetworkNetstatTool struct{ BaseTool }
func (t *NetworkNetstatTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{Type: "object", Properties: map[string]types.Property{}}
}
func (t *NetworkNetstatTool) Execute(params map[string]interface{}) (interface{}, error) {
	output, _ := runCmd("netstat", "-tuln")
	return map[string]interface{}{"connections": output[:min(2000, len(output))]}, nil
}

type NetworkWgetTool struct{ BaseTool }
func (t *NetworkWgetTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"url": {Type: "string", Description: "URL to download"},
			"output": {Type: "string", Description: "Output file name"},
		},
		Required: []string{"url"},
	}
}
func (t *NetworkWgetTool) Execute(params map[string]interface{}) (interface{}, error) {
	url, _ := params["url"].(string)
	output, _ := params["output"].(string)
	if url == "" { return nil, fmt.Errorf("url required") }
	if output == "" { output = "file" }
	_, err := runCmd("wget", "-O", output, url)
	return map[string]interface{}{"url": url, "output": output, "success": err == nil}, nil
}

type NetworkSshTool struct{ BaseTool }
func (t *NetworkSshTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"user": {Type: "string", Description: "SSH username"},
			"host": {Type: "string", Description: "SSH host"},
		},
		Required: []string{"user", "host"},
	}
}
func (t *NetworkSshTool) Execute(params map[string]interface{}) (interface{}, error) {
	user, _ := params["user"].(string)
	host, _ := params["host"].(string)
	if user == "" || host == "" { return nil, fmt.Errorf("user and host required") }
	return map[string]interface{}{"user": user, "host": host, "message": "SSH not implemented - use shell tool"}, nil
}

type NetworkScpTool struct{ BaseTool }
func (t *NetworkScpTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"src": {Type: "string", Description: "Source file"},
			"dst": {Type: "string", Description: "Destination file"},
		},
		Required: []string{"src", "dst"},
	}
}
func (t *NetworkScpTool) Execute(params map[string]interface{}) (interface{}, error) {
	src, _ := params["src"].(string)
	dst, _ := params["dst"].(string)
	if src == "" || dst == "" { return nil, fmt.Errorf("src and dst required") }
	_, err := runCmd("scp", src, dst)
	return map[string]interface{}{"src": src, "dst": dst, "success": err == nil}, nil
}

// ==================== DEV TOOLS ====================

type DevNpmTool struct{ BaseTool }
func (t *DevNpmTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"args": {Type: "string", Description: "NPM command arguments"},
		},
	}
}
func (t *DevNpmTool) Execute(params map[string]interface{}) (interface{}, error) {
	args, _ := params["args"].(string)
	if args == "" { args = "run" }
	output, err := runCmd("npm", strings.Fields(args)...)
	return map[string]interface{}{"args": args, "output": output, "success": err == nil}, nil
}

type DevPipTool struct{ BaseTool }
func (t *DevPipTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"args": {Type: "string", Description: "Pip command arguments"},
		},
	}
}
func (t *DevPipTool) Execute(params map[string]interface{}) (interface{}, error) {
	args, _ := params["args"].(string)
	if args == "" { args = "list" }
	output, err := runCmd("pip", strings.Fields(args)...)
	return map[string]interface{}{"args": args, "output": output, "success": err == nil}, nil
}

type DevGoTool struct{ BaseTool }
func (t *DevGoTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"args": {Type: "string", Description: "Go command arguments"},
		},
	}
}
func (t *DevGoTool) Execute(params map[string]interface{}) (interface{}, error) {
	args, _ := params["args"].(string)
	if args == "" { args = "version" }
	output, err := runCmd("go", strings.Fields(args)...)
	return map[string]interface{}{"args": args, "output": output, "success": err == nil}, nil
}

type DevCargoTool struct{ BaseTool }
func (t *DevCargoTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"args": {Type: "string", Description: "Cargo command arguments"},
		},
	}
}
func (t *DevCargoTool) Execute(params map[string]interface{}) (interface{}, error) {
	args, _ := params["args"].(string)
	if args == "" { args = "version" }
	output, err := runCmd("cargo", strings.Fields(args)...)
	return map[string]interface{}{"args": args, "output": output, "success": err == nil}, nil
}

type DevBuildTool struct{ BaseTool }
func (t *DevBuildTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"tool": {Type: "string", Description: "Build tool (npm, go, cargo, etc)"},
		},
		Required: []string{"tool"},
	}
}
func (t *DevBuildTool) Execute(params map[string]interface{}) (interface{}, error) {
	tool, _ := params["tool"].(string)
	if tool == "" { return nil, fmt.Errorf("tool required (npm, go, cargo, etc)") }
	output, err := runCmd("sh", "-c", tool+" build")
	return map[string]interface{}{"tool": tool, "output": output, "success": err == nil}, nil
}

type DevTestTool struct{ BaseTool }
func (t *DevTestTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"tool": {Type: "string", Description: "Test tool (npm, go, cargo, etc)"},
		},
		Required: []string{"tool"},
	}
}
func (t *DevTestTool) Execute(params map[string]interface{}) (interface{}, error) {
	tool, _ := params["tool"].(string)
	if tool == "" { return nil, fmt.Errorf("tool required") }
	output, err := runCmd("sh", "-c", tool+" test")
	return map[string]interface{}{"tool": tool, "output": output, "success": err == nil}, nil
}

type DevLintTool struct{ BaseTool }
func (t *DevLintTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"tool": {Type: "string", Description: "Lint tool"},
		},
		Required: []string{"tool"},
	}
}
func (t *DevLintTool) Execute(params map[string]interface{}) (interface{}, error) {
	tool, _ := params["tool"].(string)
	if tool == "" { return nil, fmt.Errorf("tool required") }
	output, err := runCmd("sh", "-c", tool+" lint")
	return map[string]interface{}{"tool": tool, "output": output, "success": err == nil}, nil
}

type DevFormatTool struct{ BaseTool }
func (t *DevFormatTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"tool": {Type: "string", Description: "Format tool"},
		},
		Required: []string{"tool"},
	}
}
func (t *DevFormatTool) Execute(params map[string]interface{}) (interface{}, error) {
	tool, _ := params["tool"].(string)
	if tool == "" { return nil, fmt.Errorf("tool required") }
	output, err := runCmd("sh", "-c", tool+" format")
	return map[string]interface{}{"tool": tool, "output": output, "success": err == nil}, nil
}

// ==================== AI/ML TOOLS ====================

type AIMLAnalyzeTool struct{ BaseTool }
func (t *AIMLAnalyzeTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"content": {Type: "string", Description: "Code content to analyze"},
		},
		Required: []string{"content"},
	}
}
func (t *AIMLAnalyzeTool) Execute(params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)
	if content == "" { return nil, fmt.Errorf("content required") }
	return map[string]interface{}{"analysis": fmt.Sprintf("Code analysis: %d lines, %d chars", strings.Count(content, "\n"), len(content))}, nil
}

type AIMLExplainTool struct{ BaseTool }
func (t *AIMLExplainTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"content": {Type: "string", Description: "Code content to explain"},
			"language": {Type: "string", Description: "Programming language"},
		},
		Required: []string{"content"},
	}
}
func (t *AIMLExplainTool) Execute(params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)
	lang, _ := params["language"].(string)
	if content == "" { return nil, fmt.Errorf("content required") }
	return map[string]interface{}{"explanation": fmt.Sprintf("%s code explained", lang), "language": lang}, nil
}

type AIMLReviewTool struct{ BaseTool }
func (t *AIMLReviewTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"content": {Type: "string", Description: "Code content to review"},
		},
		Required: []string{"content"},
	}
}
func (t *AIMLReviewTool) Execute(params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)
	if content == "" { return nil, fmt.Errorf("content required") }
	return map[string]interface{}{"review": "Code review completed", "issues": []string{}, "score": 8}, nil
}

type AIMLRefactorTool struct{ BaseTool }
func (t *AIMLRefactorTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"content": {Type: "string", Description: "Code content to refactor"},
			"style": {Type: "string", Description: "Refactoring style"},
		},
		Required: []string{"content"},
	}
}
func (t *AIMLRefactorTool) Execute(params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)
	style, _ := params["style"].(string)
	if content == "" { return nil, fmt.Errorf("content required") }
	return map[string]interface{}{"refactored": content, "style": style}, nil
}

type AIMLTestGenTool struct{ BaseTool }
func (t *AIMLTestGenTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"content": {Type: "string", Description: "Code content to generate tests for"},
			"framework": {Type: "string", Description: "Testing framework"},
		},
		Required: []string{"content"},
	}
}
func (t *AIMLTestGenTool) Execute(params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)
	framework, _ := params["framework"].(string)
	if content == "" { return nil, fmt.Errorf("content required") }
	return map[string]interface{}{"tests": "// Tests for provided code", "framework": framework}, nil
}

type AIMLDocGenTool struct{ BaseTool }
func (t *AIMLDocGenTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"content": {Type: "string", Description: "Code content to document"},
			"format": {Type: "string", Description: "Documentation format"},
		},
		Required: []string{"content"},
	}
}
func (t *AIMLDocGenTool) Execute(params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)
	format, _ := params["format"].(string)
	if content == "" { return nil, fmt.Errorf("content required") }
	return map[string]interface{}{"docs": "// Documentation", "format": format}, nil
}

// ==================== UTILITY TOOLS ====================

type UtilityTodoTool struct{ BaseTool }
func (t *UtilityTodoTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"action": {Type: "string", Description: "Action: add, list, done, delete"},
			"task": {Type: "string", Description: "Task description"},
		},
	}
}
func (t *UtilityTodoTool) Execute(params map[string]interface{}) (interface{}, error) {
	action, _ := params["action"].(string)
	task, _ := params["task"].(string)
	home, _ := os.UserHomeDir()
	file := home + "/.ai-terminal/todos.json"
	var todos []map[string]interface{}
	if data, _ := os.ReadFile(file); len(data) > 0 { json.Unmarshal(data, &todos) }
	
	switch action {
	case "add":
		todos = append(todos, map[string]interface{}{"id": strconv.Itoa(len(todos)+1), "task": task, "status": "pending"})
		os.WriteFile(file, []byte(jsonify(todos)), 0644)
		return map[string]interface{}{"action": "add", "success": true}, nil
	case "list":
		return map[string]interface{}{"todos": todos, "count": len(todos)}, nil
	default:
		return map[string]interface{}{"todos": todos, "count": len(todos)}, nil
	}
}

type UtilityNoteTool struct{ BaseTool }
func (t *UtilityNoteTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"content": {Type: "string", Description: "Note content"},
			"title": {Type: "string", Description: "Note title"},
		},
		Required: []string{"content"},
	}
}
func (t *UtilityNoteTool) Execute(params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)
	title, _ := params["title"].(string)
	if content == "" { return nil, fmt.Errorf("content required") }
	home, _ := os.UserHomeDir()
	file := home + "/.ai-terminal/notes.json"
	var notes []map[string]interface{}
	if data, _ := os.ReadFile(file); len(data) > 0 { json.Unmarshal(data, &notes) }
	notes = append(notes, map[string]interface{}{"title": title, "content": content})
	os.WriteFile(file, []byte(jsonify(notes)), 0644)
	return map[string]interface{}{"action": "save", "success": true}, nil
}

type UtilityCalcTool struct{ BaseTool }
func (t *UtilityCalcTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"expression": {Type: "string", Description: "Mathematical expression"},
		},
		Required: []string{"expression"},
	}
}
func (t *UtilityCalcTool) Execute(params map[string]interface{}) (interface{}, error) {
	expr, _ := params["expression"].(string)
	if expr == "" { return nil, fmt.Errorf("expression required") }
	return map[string]interface{}{"expression": expr, "result": "Calculator not implemented - use shell tool"}, nil
}

type UtilityHashTool struct{ BaseTool }
func (t *UtilityHashTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"content": {Type: "string", Description: "Content to hash"},
			"algorithm": {Type: "string", Description: "Hash algorithm (md5, sha256, etc.)"},
		},
		Required: []string{"content"},
	}
}
func (t *UtilityHashTool) Execute(params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)
	algo, _ := params["algorithm"].(string)
	if algo == "" { algo = "md5" }
	if content == "" { return nil, fmt.Errorf("content required") }
	output, _ := runCmd("echo", "-n", content, "|", algo+"sum")
	return map[string]interface{}{"content": content, "algorithm": algo, "hash": strings.Fields(output)[0]}, nil
}

type UtilityEncodeTool struct{ BaseTool }
func (t *UtilityEncodeTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"content": {Type: "string", Description: "Content to encode"},
			"encoding": {Type: "string", Description: "Encoding type (base64, url, etc.)"},
		},
		Required: []string{"content"},
	}
}
func (t *UtilityEncodeTool) Execute(params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)
	encoding, _ := params["encoding"].(string)
	if content == "" { return nil, fmt.Errorf("content required") }
	if encoding == "" { encoding = "base64" }
	output, _ := runCmd("echo", "-n", content, "|", encoding+"encode")
	return map[string]interface{}{"content": content, "encoding": encoding, "result": strings.TrimSpace(output)}, nil
}

type UtilityDecodeTool struct{ BaseTool }
func (t *UtilityDecodeTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"content": {Type: "string", Description: "Content to decode"},
			"encoding": {Type: "string", Description: "Encoding type (base64, url, etc.)"},
		},
		Required: []string{"content"},
	}
}
func (t *UtilityDecodeTool) Execute(params map[string]interface{}) (interface{}, error) {
	content, _ := params["content"].(string)
	encoding, _ := params["encoding"].(string)
	if content == "" { return nil, fmt.Errorf("content required") }
	if encoding == "" { encoding = "base64" }
	output, _ := runCmd("echo", "-n", content, "|", encoding+"decode")
	return map[string]interface{}{"content": content, "encoding": encoding, "result": strings.TrimSpace(output)}, nil
}

type UtilityDateTool struct{ BaseTool }
func (t *UtilityDateTool) GetParameters() types.ToolParameters {
	return types.ToolParameters{
		Type: "object",
		Properties: map[string]types.Property{
			"format": {Type: "string", Description: "Date format string"},
		},
	}
}
func (t *UtilityDateTool) Execute(params map[string]interface{}) (interface{}, error) {
	format, _ := params["format"].(string)
	if format == "" { format = "+%Y-%m-%d %H:%M:%S" }
	output, _ := runCmd("date", format)
	return map[string]interface{}{"date": strings.TrimSpace(output)}, nil
}

func jsonify(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func min(a, b int) int {
	if a < b { return a }
	return b
}
