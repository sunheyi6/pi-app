package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"pi-desktop/backend/piagent"
)

type fakePackageRunner struct {
	result     piagent.PackageResult
	err        error
	operations []piagent.PackageOperation
}

func (f *fakePackageRunner) Run(_ context.Context, _ string, operation piagent.PackageOperation) (piagent.PackageResult, error) {
	f.operations = append(f.operations, operation)
	return f.result, f.err
}

func TestInstallPackageRestartsActiveSession(t *testing.T) {
	runner := &fakePackageRunner{result: piagent.PackageResult{Output: "installed"}}
	app := NewApp()
	app.curCwd = `C:\project`
	app.activeSessionPath = `C:\session.jsonl`
	app.packageManager = runner

	var gotCwd string
	var gotSession string
	app.restartAgent = func(cwd string, sessionPath string) error {
		gotCwd = cwd
		gotSession = sessionPath
		return nil
	}

	result, err := app.InstallPackage("npm:demo", "project")
	if err != nil {
		t.Fatalf("InstallPackage() error = %v", err)
	}
	if gotCwd != app.curCwd || gotSession != app.activeSessionPath {
		t.Fatalf("restart = (%q, %q), want (%q, %q)", gotCwd, gotSession, app.curCwd, app.activeSessionPath)
	}
	if len(runner.operations) != 1 || runner.operations[0].Scope != piagent.ScopeProject {
		t.Fatalf("operations = %#v", runner.operations)
	}
	if !strings.Contains(result, `"restarted":true`) {
		t.Fatalf("InstallPackage() = %s, want restarted status", result)
	}
}

func TestFailedPackageChangeDoesNotRestartAgent(t *testing.T) {
	runner := &fakePackageRunner{err: errors.New("install failed")}
	app := NewApp()
	app.curCwd = t.TempDir()
	app.packageManager = runner
	app.restartAgent = func(string, string) error {
		t.Fatal("restartAgent should not run after failed package operation")
		return nil
	}

	if _, err := app.RemovePackage("npm:demo", "global"); err == nil {
		t.Fatal("RemovePackage() error = nil, want failure")
	}
}

func TestRetryAgentStartupUsesLastRestartContext(t *testing.T) {
	runner := &fakePackageRunner{result: piagent.PackageResult{Output: "updated"}}
	app := NewApp()
	app.curCwd = `C:\project`
	app.activeSessionPath = `C:\session.jsonl`
	app.packageManager = runner

	attempts := 0
	app.restartAgent = func(cwd string, sessionPath string) error {
		attempts++
		if cwd != `C:\project` || sessionPath != `C:\session.jsonl` {
			t.Fatalf("restart context = (%q, %q)", cwd, sessionPath)
		}
		if attempts == 1 {
			return errors.New("temporary restart failure")
		}
		return nil
	}

	result, err := app.UpdateAllPackages()
	if err != nil {
		t.Fatalf("UpdateAllPackages() error = %v", err)
	}
	if !strings.Contains(result, `"restartError":"temporary restart failure"`) {
		t.Fatalf("UpdateAllPackages() = %s, want restart error", result)
	}

	if err := app.RetryAgentStartup(); err != nil {
		t.Fatalf("RetryAgentStartup() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("restart attempts = %d, want 2", attempts)
	}
}

func TestListPackagesFiltersRequestedScope(t *testing.T) {
	runner := &fakePackageRunner{result: piagent.PackageResult{
		Packages: []piagent.PackageInfo{
			{Source: "npm:project", Scope: piagent.ScopeProject},
			{Source: "npm:global", Scope: piagent.ScopeGlobal},
		},
	}}
	app := NewApp()
	app.curCwd = t.TempDir()
	app.packageManager = runner

	result, err := app.ListPackages("global")
	if err != nil {
		t.Fatalf("ListPackages() error = %v", err)
	}
	if strings.Contains(result, "npm:project") || !strings.Contains(result, "npm:global") {
		t.Fatalf("ListPackages() = %s, want only global package", result)
	}
}
