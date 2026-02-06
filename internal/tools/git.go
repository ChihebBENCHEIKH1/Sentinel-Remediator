package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// GitTool provides Git operations for the agent
type GitTool struct {
	workDir     string
	githubToken string
	logger      *zap.Logger
}

// NewGitTool creates a new Git tool
func NewGitTool(workDir, githubToken string, logger *zap.Logger) *GitTool {
	return &GitTool{
		workDir:     workDir,
		githubToken: githubToken,
		logger:      logger,
	}
}

func (g *GitTool) Name() string {
	return "git"
}

func (g *GitTool) Description() string {
	return `Git operations tool for cloning repositories, creating branches, committing changes, and pushing to remote.
Available actions:
- clone: Clone a repository to the work directory
- branch: Create and switch to a new branch
- add: Stage files for commit
- commit: Commit staged changes
- push: Push changes to remote
- status: Get current git status
- diff: Get diff of current changes`
}

func (g *GitTool) Parameters() []Parameter {
	return []Parameter{
		{Name: "action", Type: "string", Description: "Git action: clone, branch, add, commit, push, status, diff", Required: true},
		{Name: "repo_url", Type: "string", Description: "Repository URL (for clone action)", Required: false},
		{Name: "branch_name", Type: "string", Description: "Branch name (for branch action)", Required: false},
		{Name: "files", Type: "array", Description: "Files to add (for add action, defaults to '.')", Required: false},
		{Name: "message", Type: "string", Description: "Commit message (for commit action)", Required: false},
		{Name: "repo_path", Type: "string", Description: "Path to repository within work directory", Required: false},
	}
}

func (g *GitTool) Execute(ctx context.Context, args map[string]any) (Result, error) {
	action, ok := args["action"].(string)
	if !ok {
		return NewErrorResult("action is required"), nil
	}

	// Determine the repository path
	repoPath := g.workDir
	if rp, ok := args["repo_path"].(string); ok && rp != "" {
		repoPath = filepath.Join(g.workDir, rp)
	}

	switch action {
	case "clone":
		return g.clone(ctx, args, repoPath)
	case "branch":
		return g.branch(ctx, args, repoPath)
	case "add":
		return g.add(ctx, args, repoPath)
	case "commit":
		return g.commit(ctx, args, repoPath)
	case "push":
		return g.push(ctx, repoPath)
	case "status":
		return g.status(ctx, repoPath)
	case "diff":
		return g.diff(ctx, repoPath)
	default:
		return NewErrorResult(fmt.Sprintf("unknown git action: %s", action)), nil
	}
}

func (g *GitTool) clone(ctx context.Context, args map[string]any, repoPath string) (Result, error) {
	repoURL, ok := args["repo_url"].(string)
	if !ok || repoURL == "" {
		return NewErrorResult("repo_url is required for clone action"), nil
	}

	// Inject token for authentication if HTTPS URL
	if g.githubToken != "" && strings.Contains(repoURL, "https://github.com") {
		repoURL = strings.Replace(repoURL, "https://", fmt.Sprintf("https://%s@", g.githubToken), 1)
	}

	// Clean the repo path if it exists
	if _, err := os.Stat(repoPath); err == nil {
		os.RemoveAll(repoPath)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", repoURL, repoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return NewErrorResult(fmt.Sprintf("git clone failed: %s\n%s", err, string(output))), nil
	}

	return NewSuccessResult(fmt.Sprintf("Repository cloned to %s\n%s", repoPath, string(output))), nil
}

func (g *GitTool) branch(ctx context.Context, args map[string]any, repoPath string) (Result, error) {
	branchName, ok := args["branch_name"].(string)
	if !ok || branchName == "" {
		return NewErrorResult("branch_name is required for branch action"), nil
	}

	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "checkout", "-b", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return NewErrorResult(fmt.Sprintf("git branch failed: %s\n%s", err, string(output))), nil
	}

	return NewSuccessResult(fmt.Sprintf("Created and switched to branch: %s\n%s", branchName, string(output))), nil
}

func (g *GitTool) add(ctx context.Context, args map[string]any, repoPath string) (Result, error) {
	files := "."
	if f, ok := args["files"].([]interface{}); ok && len(f) > 0 {
		fileStrs := make([]string, len(f))
		for i, file := range f {
			fileStrs[i] = fmt.Sprintf("%v", file)
		}
		files = strings.Join(fileStrs, " ")
	}

	cmdArgs := append([]string{"-C", repoPath, "add"}, strings.Split(files, " ")...)
	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return NewErrorResult(fmt.Sprintf("git add failed: %s\n%s", err, string(output))), nil
	}

	return NewSuccessResult(fmt.Sprintf("Staged files: %s\n%s", files, string(output))), nil
}

func (g *GitTool) commit(ctx context.Context, args map[string]any, repoPath string) (Result, error) {
	message, ok := args["message"].(string)
	if !ok || message == "" {
		return NewErrorResult("message is required for commit action"), nil
	}

	// Configure git user if not set
	exec.CommandContext(ctx, "git", "-C", repoPath, "config", "user.email", "sentinel@automated.fix").Run()
	exec.CommandContext(ctx, "git", "-C", repoPath, "config", "user.name", "Sentinel Remediator").Run()

	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "commit", "-m", message)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return NewErrorResult(fmt.Sprintf("git commit failed: %s\n%s", err, string(output))), nil
	}

	return NewSuccessResult(fmt.Sprintf("Committed: %s\n%s", message, string(output))), nil
}

func (g *GitTool) push(ctx context.Context, repoPath string) (Result, error) {
	// Get current branch name
	branchCmd := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to get branch name: %s", err)), nil
	}
	branch := strings.TrimSpace(string(branchOutput))

	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "push", "-u", "origin", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return NewErrorResult(fmt.Sprintf("git push failed: %s\n%s", err, string(output))), nil
	}

	return NewSuccessResult(fmt.Sprintf("Pushed to origin/%s\n%s", branch, string(output))), nil
}

func (g *GitTool) status(ctx context.Context, repoPath string) (Result, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return NewErrorResult(fmt.Sprintf("git status failed: %s\n%s", err, string(output))), nil
	}

	if len(output) == 0 {
		return NewSuccessResult("Working directory clean"), nil
	}
	return NewSuccessResult(string(output)), nil
}

func (g *GitTool) diff(ctx context.Context, repoPath string) (Result, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "diff")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return NewErrorResult(fmt.Sprintf("git diff failed: %s\n%s", err, string(output))), nil
	}

	if len(output) == 0 {
		return NewSuccessResult("No changes"), nil
	}
	return NewSuccessResult(string(output)), nil
}
