package piagent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPackageArgs(t *testing.T) {
	tests := []struct {
		name string
		op   PackageOperation
		want []string
	}{
		{"project install", PackageOperation{Action: "install", Source: "npm:demo", Scope: ScopeProject}, []string{"install", "npm:demo", "-l"}},
		{"global remove", PackageOperation{Action: "remove", Source: "npm:demo", Scope: ScopeGlobal}, []string{"remove", "npm:demo"}},
		{"single update", PackageOperation{Action: "update", Source: "npm:demo"}, []string{"update", "--extension", "npm:demo"}},
		{"all updates", PackageOperation{Action: "update"}, []string{"update", "--extensions"}},
		{"list", PackageOperation{Action: "list"}, []string{"list"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := packageArgs(tt.op)
			if err != nil {
				t.Fatalf("packageArgs() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("packageArgs() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestPackageArgsRejectsMissingSource(t *testing.T) {
	for _, action := range []string{"install", "remove"} {
		t.Run(action, func(t *testing.T) {
			_, err := packageArgs(PackageOperation{Action: action})
			if err == nil {
				t.Fatal("packageArgs() error = nil, want missing source error")
			}
		})
	}
}

func TestPackageManagerRejectsMissingProjectDirectory(t *testing.T) {
	manager := PackageManager{
		Executable: "pi",
		RunCommand: func(context.Context, string, []string, string) (string, string, error) {
			t.Fatal("RunCommand should not run for missing project directory")
			return "", "", nil
		},
	}

	_, err := manager.Run(context.Background(), filepath.Join(t.TempDir(), "missing"), PackageOperation{
		Action: "install",
		Source: "npm:demo",
		Scope:  ScopeProject,
	})
	if err == nil || !strings.Contains(err.Error(), "project directory") {
		t.Fatalf("Run() error = %v, want project directory error", err)
	}
}

func TestPackageManagerRejectsMissingLocalPath(t *testing.T) {
	manager := PackageManager{
		Executable: "pi",
		RunCommand: func(context.Context, string, []string, string) (string, string, error) {
			t.Fatal("RunCommand should not run for missing local package")
			return "", "", nil
		},
	}

	_, err := manager.Run(context.Background(), t.TempDir(), PackageOperation{
		Action: "install",
		Source: "./missing-package",
		Scope:  ScopeProject,
	})
	if err == nil || !strings.Contains(err.Error(), "local package path") {
		t.Fatalf("Run() error = %v, want local package path error", err)
	}
}

func TestPackageManagerReturnsStderr(t *testing.T) {
	manager := PackageManager{
		Executable: "pi",
		RunCommand: func(context.Context, string, []string, string) (string, string, error) {
			return "", "network unavailable", errors.New("exit status 1")
		},
	}

	_, err := manager.Run(context.Background(), t.TempDir(), PackageOperation{
		Action: "install",
		Source: "npm:demo",
	})
	if err == nil || !strings.Contains(err.Error(), "network unavailable") {
		t.Fatalf("Run() error = %v, want stderr text", err)
	}
}

func TestPackageManagerReportsTimeout(t *testing.T) {
	manager := PackageManager{
		Executable: "pi",
		Timeout:    time.Millisecond,
		RunCommand: func(ctx context.Context, _ string, _ []string, _ string) (string, string, error) {
			<-ctx.Done()
			return "", "", ctx.Err()
		},
	}

	_, err := manager.Run(context.Background(), t.TempDir(), PackageOperation{
		Action: "update",
	})
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("Run() error = %v, want timeout error", err)
	}
}

func TestPackageManagerParsesListOutput(t *testing.T) {
	manager := PackageManager{
		Executable: "pi",
		RunCommand: func(context.Context, string, []string, string) (string, string, error) {
			return "npm:pi-web-access\nproject: git:github.com/acme/tools\n", "", nil
		},
	}

	result, err := manager.Run(context.Background(), t.TempDir(), PackageOperation{Action: "list"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(result.Packages) != 2 {
		t.Fatalf("Run() packages = %#v, want 2 rows", result.Packages)
	}
	if result.Packages[0].Source != "npm:pi-web-access" {
		t.Fatalf("first source = %q", result.Packages[0].Source)
	}
	if result.Packages[1].Scope != ScopeProject || result.Packages[1].Source != "git:github.com/acme/tools" {
		t.Fatalf("second package = %#v", result.Packages[1])
	}
}

func TestPiExecutableCandidatesIncludeWindowsGlobalLocations(t *testing.T) {
	home := filepath.Join("C:", "Users", "demo")
	got := piExecutableCandidates(home)

	want := []string{
		filepath.Join(home, "AppData", "Local", "pnpm", "pi.CMD"),
		filepath.Join(home, "AppData", "Roaming", "npm", "pi.cmd"),
	}
	for _, candidate := range want {
		found := false
		for _, item := range got {
			if item == candidate {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("piExecutableCandidates() = %#v, missing %q", got, candidate)
		}
	}
}

func TestPackageManagerAcceptsExistingLocalPath(t *testing.T) {
	root := t.TempDir()
	localPackage := filepath.Join(root, "extension.ts")
	if err := os.WriteFile(localPackage, []byte("export default {}"), 0o600); err != nil {
		t.Fatal(err)
	}

	var gotArgs []string
	manager := PackageManager{
		Executable: "pi",
		RunCommand: func(_ context.Context, _ string, args []string, _ string) (string, string, error) {
			gotArgs = args
			return "ok", "", nil
		},
	}

	_, err := manager.Run(context.Background(), root, PackageOperation{
		Action: "install",
		Source: localPackage,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !reflect.DeepEqual(gotArgs, []string{"install", localPackage}) {
		t.Fatalf("RunCommand args = %#v", gotArgs)
	}
}
