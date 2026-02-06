package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// FilesystemTool provides filesystem operations for the agent
type FilesystemTool struct {
	workDir string
	logger  *zap.Logger
}

// NewFilesystemTool creates a new filesystem tool
func NewFilesystemTool(workDir string, logger *zap.Logger) *FilesystemTool {
	return &FilesystemTool{
		workDir: workDir,
		logger:  logger,
	}
}

func (f *FilesystemTool) Name() string {
	return "filesystem"
}

func (f *FilesystemTool) Description() string {
	return `Filesystem operations tool for reading, writing, and modifying files.
Available actions:
- read: Read the contents of a file
- write: Write content to a file (creates or overwrites)
- patch: Search and replace text in a file
- list: List contents of a directory
- exists: Check if a file or directory exists
- delete: Delete a file`
}

func (f *FilesystemTool) Parameters() []Parameter {
	return []Parameter{
		{Name: "action", Type: "string", Description: "Filesystem action: read, write, patch, list, exists, delete", Required: true},
		{Name: "path", Type: "string", Description: "File or directory path (relative to work directory)", Required: true},
		{Name: "content", Type: "string", Description: "Content to write (for write action)", Required: false},
		{Name: "search", Type: "string", Description: "Text to search for (for patch action)", Required: false},
		{Name: "replace", Type: "string", Description: "Text to replace with (for patch action)", Required: false},
	}
}

func (f *FilesystemTool) Execute(ctx context.Context, args map[string]any) (Result, error) {
	action, ok := args["action"].(string)
	if !ok {
		return NewErrorResult("action is required"), nil
	}

	pathArg, ok := args["path"].(string)
	if !ok || pathArg == "" {
		return NewErrorResult("path is required"), nil
	}

	// Resolve the full path (ensure it's within work directory for security)
	fullPath := f.resolvePath(pathArg)
	if !strings.HasPrefix(fullPath, f.workDir) {
		return NewErrorResult("path must be within work directory"), nil
	}

	switch action {
	case "read":
		return f.read(fullPath)
	case "write":
		return f.write(fullPath, args)
	case "patch":
		return f.patch(fullPath, args)
	case "list":
		return f.list(fullPath)
	case "exists":
		return f.exists(fullPath)
	case "delete":
		return f.delete(fullPath)
	default:
		return NewErrorResult(fmt.Sprintf("unknown filesystem action: %s", action)), nil
	}
}

func (f *FilesystemTool) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(f.workDir, path))
}

func (f *FilesystemTool) read(path string) (Result, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to read file: %s", err)), nil
	}
	return NewSuccessResult(string(content)), nil
}

func (f *FilesystemTool) write(path string, args map[string]any) (Result, error) {
	content, ok := args["content"].(string)
	if !ok {
		return NewErrorResult("content is required for write action"), nil
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return NewErrorResult(fmt.Sprintf("failed to create directory: %s", err)), nil
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return NewErrorResult(fmt.Sprintf("failed to write file: %s", err)), nil
	}

	return NewSuccessResult(fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path)), nil
}

func (f *FilesystemTool) patch(path string, args map[string]any) (Result, error) {
	search, ok := args["search"].(string)
	if !ok || search == "" {
		return NewErrorResult("search is required for patch action"), nil
	}

	replace, ok := args["replace"].(string)
	if !ok {
		return NewErrorResult("replace is required for patch action"), nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to read file: %s", err)), nil
	}

	originalContent := string(content)
	if !strings.Contains(originalContent, search) {
		return NewErrorResult(fmt.Sprintf("search text not found in file: %s", path)), nil
	}

	newContent := strings.Replace(originalContent, search, replace, -1)
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return NewErrorResult(fmt.Sprintf("failed to write file: %s", err)), nil
	}

	count := strings.Count(originalContent, search)
	return NewSuccessResult(fmt.Sprintf("Replaced %d occurrence(s) in %s", count, path)), nil
}

func (f *FilesystemTool) list(path string) (Result, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to list directory: %s", err)), nil
	}

	var sb strings.Builder
	for _, entry := range entries {
		typeChar := "-"
		if entry.IsDir() {
			typeChar = "d"
		}
		info, _ := entry.Info()
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		sb.WriteString(fmt.Sprintf("%s %8d %s\n", typeChar, size, entry.Name()))
	}

	return NewSuccessResult(sb.String()), nil
}

func (f *FilesystemTool) exists(path string) (Result, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return Result{
			Success: true,
			Output:  "false",
			Data:    false,
		}, nil
	}
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to check path: %s", err)), nil
	}

	pathType := "file"
	if info.IsDir() {
		pathType = "directory"
	}
	return Result{
		Success: true,
		Output:  fmt.Sprintf("true (%s)", pathType),
		Data:    true,
	}, nil
}

func (f *FilesystemTool) delete(path string) (Result, error) {
	err := os.Remove(path)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to delete file: %s", err)), nil
	}
	return NewSuccessResult(fmt.Sprintf("Deleted: %s", path)), nil
}
