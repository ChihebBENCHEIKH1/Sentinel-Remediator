package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

// DockerTool provides Docker build and verification operations
type DockerTool struct {
	workDir string
	logger  *zap.Logger
	timeout time.Duration
}

// NewDockerTool creates a new Docker tool
func NewDockerTool(workDir string, logger *zap.Logger) *DockerTool {
	return &DockerTool{
		workDir: workDir,
		logger:  logger,
		timeout: 5 * time.Minute, // Default build timeout
	}
}

func (d *DockerTool) Name() string {
	return "docker"
}

func (d *DockerTool) Description() string {
	return `Docker operations tool for building images and running verification tests.
Available actions:
- build: Build a Docker image from a Dockerfile
- run: Run a command in a container to verify the image works
- inspect: Inspect a built image for security issues
- cleanup: Remove built images to free space

Use this tool to verify that Dockerfile changes don't break the build.`
}

func (d *DockerTool) Parameters() []Parameter {
	return []Parameter{
		{Name: "action", Type: "string", Description: "Docker action: build, run, inspect, cleanup", Required: true},
		{Name: "dockerfile", Type: "string", Description: "Path to Dockerfile (for build action)", Required: false},
		{Name: "context", Type: "string", Description: "Build context directory (defaults to Dockerfile directory)", Required: false},
		{Name: "tag", Type: "string", Description: "Image tag (for build/run actions)", Required: false},
		{Name: "command", Type: "string", Description: "Command to run (for run action)", Required: false},
	}
}

func (d *DockerTool) Execute(ctx context.Context, args map[string]any) (Result, error) {
	action, ok := args["action"].(string)
	if !ok {
		return NewErrorResult("action is required"), nil
	}

	switch action {
	case "build":
		return d.build(ctx, args)
	case "run":
		return d.run(ctx, args)
	case "inspect":
		return d.inspect(ctx, args)
	case "cleanup":
		return d.cleanup(ctx, args)
	default:
		return NewErrorResult(fmt.Sprintf("unknown docker action: %s", action)), nil
	}
}

func (d *DockerTool) build(ctx context.Context, args map[string]any) (Result, error) {
	dockerfile := "Dockerfile"
	if df, ok := args["dockerfile"].(string); ok && df != "" {
		dockerfile = df
	}

	buildContext := "."
	if c, ok := args["context"].(string); ok && c != "" {
		buildContext = c
	}

	tag := "sentinel-test:latest"
	if t, ok := args["tag"].(string); ok && t != "" {
		tag = t
	}

	// Add timeout to context
	buildCtx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	cmd := exec.CommandContext(buildCtx, "docker", "build",
		"-f", dockerfile,
		"-t", tag,
		"--no-cache",
		buildContext,
	)
	cmd.Dir = d.workDir

	d.logger.Info("Building Docker image",
		zap.String("dockerfile", dockerfile),
		zap.String("tag", tag),
		zap.String("context", buildContext),
	)

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		// Extract useful error information from the output
		errorInfo := extractBuildError(outputStr)
		return Result{
			Success: false,
			Output:  outputStr,
			Error:   fmt.Sprintf("Docker build failed: %s\n\nError details:\n%s", err, errorInfo),
		}, nil
	}

	return Result{
		Success: true,
		Output:  fmt.Sprintf("Successfully built image: %s\n\n%s", tag, outputStr),
		Data:    map[string]string{"tag": tag},
	}, nil
}

func (d *DockerTool) run(ctx context.Context, args map[string]any) (Result, error) {
	tag := "sentinel-test:latest"
	if t, ok := args["tag"].(string); ok && t != "" {
		tag = t
	}

	command := "echo 'Container started successfully'"
	if c, ok := args["command"].(string); ok && c != "" {
		command = c
	}

	// Run with a timeout
	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "docker", "run", "--rm", tag, "sh", "-c", command)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return Result{
			Success: false,
			Output:  string(output),
			Error:   fmt.Sprintf("Docker run failed: %s", err),
		}, nil
	}

	return NewSuccessResult(fmt.Sprintf("Container test passed:\n%s", string(output))), nil
}

func (d *DockerTool) inspect(ctx context.Context, args map[string]any) (Result, error) {
	tag := "sentinel-test:latest"
	if t, ok := args["tag"].(string); ok && t != "" {
		tag = t
	}

	cmd := exec.CommandContext(ctx, "docker", "inspect",
		"--format", "{{json .Config}}",
		tag,
	)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return NewErrorResult(fmt.Sprintf("Docker inspect failed: %s", err)), nil
	}

	return NewSuccessResult(fmt.Sprintf("Image configuration:\n%s", string(output))), nil
}

func (d *DockerTool) cleanup(ctx context.Context, args map[string]any) (Result, error) {
	tag := "sentinel-test:latest"
	if t, ok := args["tag"].(string); ok && t != "" {
		tag = t
	}

	cmd := exec.CommandContext(ctx, "docker", "rmi", "-f", tag)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Image might not exist, which is fine
		d.logger.Debug("Docker cleanup warning", zap.String("output", string(output)))
	}

	return NewSuccessResult(fmt.Sprintf("Cleaned up image: %s", tag)), nil
}

// extractBuildError extracts the most relevant error information from build output
func extractBuildError(output string) string {
	lines := strings.Split(output, "\n")
	var errorLines []string
	capture := false

	for _, line := range lines {
		// Look for error indicators
		if strings.Contains(line, "error:") ||
			strings.Contains(line, "ERROR:") ||
			strings.Contains(line, "failed to") ||
			strings.Contains(line, "Failed to") ||
			strings.Contains(line, "Cannot") ||
			strings.Contains(line, "cannot") {
			capture = true
		}

		if capture {
			errorLines = append(errorLines, line)
		}

		// Stop after collecting enough context
		if len(errorLines) > 20 {
			break
		}
	}

	if len(errorLines) == 0 {
		// Return last 10 lines if no specific error found
		start := len(lines) - 10
		if start < 0 {
			start = 0
		}
		return strings.Join(lines[start:], "\n")
	}

	return strings.Join(errorLines, "\n")
}
