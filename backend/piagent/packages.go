package piagent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type PackageScope string

const (
	ScopeGlobal  PackageScope = "global"
	ScopeProject PackageScope = "project"
)

type PackageOperation struct {
	Action string       `json:"action"`
	Source string       `json:"source,omitempty"`
	Scope  PackageScope `json:"scope,omitempty"`
}

type PackageInfo struct {
	Source string       `json:"source"`
	Scope  PackageScope `json:"scope"`
	Type   string       `json:"type"`
	Raw    string       `json:"raw,omitempty"`
}

type PackageResult struct {
	Output   string        `json:"output,omitempty"`
	Packages []PackageInfo `json:"packages,omitempty"`
}

type CommandRunner func(ctx context.Context, executable string, args []string, cwd string) (stdout string, stderr string, err error)

type PackageManager struct {
	Executable string
	Timeout    time.Duration
	RunCommand CommandRunner
}

func NewPackageManager() (*PackageManager, error) {
	executable, err := resolvePiExecutable()
	if err != nil {
		return nil, err
	}
	return &PackageManager{Executable: executable}, nil
}

func (m PackageManager) Run(ctx context.Context, cwd string, operation PackageOperation) (PackageResult, error) {
	args, err := packageArgs(operation)
	if err != nil {
		return PackageResult{}, err
	}
	if operation.Scope == ScopeProject {
		if info, statErr := os.Stat(cwd); statErr != nil || !info.IsDir() {
			return PackageResult{}, fmt.Errorf("project directory is not available: %s", cwd)
		}
	}
	if operation.Action == "install" && isLocalPackageSource(operation.Source) {
		localPath := operation.Source
		if !filepath.IsAbs(localPath) {
			localPath = filepath.Join(cwd, localPath)
		}
		if _, statErr := os.Stat(localPath); statErr != nil {
			return PackageResult{}, fmt.Errorf("local package path does not exist: %s", operation.Source)
		}
	}

	timeout := m.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	runner := m.RunCommand
	if runner == nil {
		runner = runPackageCommand
	}
	stdout, stderr, runErr := runner(runCtx, m.Executable, args, cwd)
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return PackageResult{}, fmt.Errorf("pi package command timed out after %s", timeout)
	}
	if runErr != nil {
		message := strings.TrimSpace(stderr)
		if message == "" {
			message = runErr.Error()
		}
		return PackageResult{}, fmt.Errorf("pi package command failed: %s", message)
	}

	result := PackageResult{Output: strings.TrimSpace(stdout)}
	if operation.Action == "list" {
		result.Packages = parsePackageList(stdout)
	}
	return result, nil
}

func packageArgs(operation PackageOperation) ([]string, error) {
	switch operation.Action {
	case "install", "remove":
		if strings.TrimSpace(operation.Source) == "" {
			return nil, fmt.Errorf("%s source is required", operation.Action)
		}
		args := []string{operation.Action, operation.Source}
		if operation.Scope == ScopeProject {
			args = append(args, "-l")
		}
		return args, nil
	case "update":
		if strings.TrimSpace(operation.Source) == "" {
			return []string{"update", "--extensions"}, nil
		}
		return []string{"update", "--extension", operation.Source}, nil
	case "list":
		return []string{"list"}, nil
	default:
		return nil, fmt.Errorf("unsupported package action: %s", operation.Action)
	}
}

func runPackageCommand(ctx context.Context, executable string, args []string, cwd string) (string, string, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = cwd
	cmd.Env = withPnpmPath(os.Environ())
	var stdout strings.Builder
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func isLocalPackageSource(source string) bool {
	return filepath.IsAbs(source) || strings.HasPrefix(source, ".")
}

func parsePackageList(output string) []PackageInfo {
	var packages []PackageInfo
	for _, rawLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		scope := ScopeGlobal
		source := line
		for _, prefix := range []struct {
			label string
			scope PackageScope
		}{
			{"project:", ScopeProject},
			{"global:", ScopeGlobal},
		} {
			if strings.HasPrefix(strings.ToLower(source), prefix.label) {
				scope = prefix.scope
				source = strings.TrimSpace(source[len(prefix.label):])
				break
			}
		}
		packages = append(packages, PackageInfo{
			Source: source,
			Scope:  scope,
			Type:   packageSourceType(source),
			Raw:    line,
		})
	}
	return packages
}

func packageSourceType(source string) string {
	switch {
	case strings.HasPrefix(source, "npm:"):
		return "npm"
	case strings.HasPrefix(source, "git:"),
		strings.HasPrefix(source, "https://"),
		strings.HasPrefix(source, "http://"),
		strings.HasPrefix(source, "ssh://"),
		strings.HasPrefix(source, "git://"):
		return "git"
	default:
		return "local"
	}
}
