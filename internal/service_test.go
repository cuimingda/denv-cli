package denv

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"testing"
)

func TestNewInstallPlanServicePanicsWhenRegistryMissing(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected newInstallPlanService to panic when registry is nil")
		}
	}()
	_ = newInstallPlanService(Runtime{}, NewToolCatalog(), nil)
}

func TestServiceSupportedToolsAndInstallableOrder(t *testing.T) {
	svc := NewService(Runtime{})

	supported := svc.SupportedTools()
	if len(supported) == 0 {
		t.Fatalf("supported tools should not be empty")
	}

	if len(svc.InstallableTools()) == 0 {
		t.Fatalf("installable tools should not be empty")
	}
}

func TestServiceIsCommandAvailable(t *testing.T) {
	svc := NewService(Runtime{
		ExecutableLookup: func(name string) (string, error) {
			if name == "php" {
				return "/usr/bin/php", nil
			}
			return "", exec.ErrNotFound
		},
	})

	if !svc.IsCommandAvailable("php") {
		t.Fatalf("expected php available")
	}
}

func TestServiceToolInstallStateFromCommandPath(t *testing.T) {
	svc := NewService(Runtime{
		ExecutableLookup: func(name string) (string, error) {
			if name == "ffmpeg" {
				return "/usr/local/bin/ffmpeg", nil
			}
			return "", exec.ErrNotFound
		},
	})

	installed, commandPath, byBrew, err := svc.ToolInstallState("ffmpeg")
	if err != nil {
		t.Fatalf("tool install state failed: %v", err)
	}
	if !installed {
		t.Fatalf("expected tool to be installed")
	}
	if commandPath != "/usr/local/bin/ffmpeg" {
		t.Fatalf("expected command path /usr/local/bin/ffmpeg, got %q", commandPath)
	}
	if byBrew {
		t.Fatalf("expected non-homebrew command path")
	}
}

func TestServiceToolInstallStateByBrewFormula(t *testing.T) {
	svc := NewService(Runtime{
		ExecutableLookup: func(name string) (string, error) { return "", exec.ErrNotFound },
		CommandRunner: func(name string, args ...string) ([]byte, error) {
			switch {
			case name == "brew" && len(args) == 3 && args[0] == "list" && args[1] == "--formula" && args[2] == "ffmpeg":
				return []byte("ffmpeg\n"), nil
			case name == "brew" && len(args) == 2 && args[0] == "--prefix" && args[1] == "ffmpeg":
				return []byte("/opt/homebrew/opt/ffmpeg"), nil
			default:
				return nil, errors.New("unexpected command")
			}
		},
	})

	installed, commandPath, byBrew, err := svc.ToolInstallState("ffmpeg")
	if err != nil {
		t.Fatalf("tool install state failed: %v", err)
	}
	if !installed {
		t.Fatalf("expected tool to be installed by brew")
	}
	if !byBrew {
		t.Fatalf("expected installed by brew")
	}
	if commandPath != "/opt/homebrew/opt/ffmpeg/bin/ffmpeg" {
		t.Fatalf("unexpected command path: %q", commandPath)
	}
}

func TestServiceBuildInstallOperationsCanBypassWithExistingBinary(t *testing.T) {
	svc := NewService(Runtime{
		ExecutableLookup: func(name string) (string, error) {
			if name == "brew" || name == "php" {
				return "/usr/local/bin/" + name, nil
			}
			return "", exec.ErrNotFound
		},
		CommandRunner: func(string, ...string) ([]byte, error) {
			t.Fatalf("should not execute install checks when command exists")
			return nil, nil
		},
	})

	ops, err := svc.BuildInstallOperationsForTool("php", false)
	if err != nil {
		t.Fatalf("build install operations failed: %v", err)
	}
	if len(ops) != 0 {
		t.Fatalf("expected no operations for existing install, got %v", ops)
	}
}

func TestServiceOutdatedChecksReturnsStructuredItems(t *testing.T) {
	svc := NewService(Runtime{
		ExecutableLookup: func(name string) (string, error) {
			return "", exec.ErrNotFound
		},
		CommandRunner: func(name string, args ...string) ([]byte, error) {
			if name == "brew" && len(args) >= 3 && args[0] == "info" && args[1] == "--json=v2" {
				formula := args[2]
				payload := `{"formulae":[{"name":"` + formula + `","versions":{"stable":"9.9.9"}}]}`
				return []byte(payload), nil
			}
			if name == "npm" && len(args) == 3 && args[0] == "view" && args[1] == "npm" && args[2] == "version" {
				return []byte("11.0.0"), nil
			}
			return []byte(""), nil
		},
	})

	rows, err := svc.OutdatedChecks()
	if err != nil {
		t.Fatalf("outdated checks failed: %v", err)
	}
	if len(rows) != len(svc.SupportedTools()) {
		t.Fatalf("expected %d checks, got %d", len(svc.SupportedTools()), len(rows))
	}

	for _, row := range rows {
		if row.Name == "" {
			t.Fatalf("expected check name")
		}
		if row.State == "" {
			t.Fatalf("expected state for %q", row.Name)
		}
		if row.CheckError != nil {
			t.Fatalf("unexpected check error for %q: %s", row.Name, row.CheckError)
		}
	}
}

func TestServiceUpdateToolWithOutputRequiresInstall(t *testing.T) {
	svc := NewService(Runtime{
		ExecutableLookup: func(name string) (string, error) {
			if name == "brew" {
				return "/opt/homebrew/bin/brew", nil
			}
			return "", exec.ErrNotFound
		},
		CommandRunner: func(name string, args ...string) ([]byte, error) {
			if name == "brew" && len(args) >= 2 && args[0] == "list" && args[1] == "--formula" {
				return []byte(""), nil
			}
			return nil, nil
		},
		CommandRunnerWithOutput: func(_ io.Writer, _ string, _ ...string) error {
			t.Fatalf("should not run update for not installed tools")
			return nil
		},
	})

	if err := svc.UpdateToolWithOutput(bytes.NewBuffer(nil), "php"); err == nil {
		t.Fatalf("expected update to fail for uninstalled tool")
	}
}
