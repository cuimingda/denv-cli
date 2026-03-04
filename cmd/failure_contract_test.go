package cmd

import (
	"errors"
	"io"
	"io/fs"
	"os/exec"
	"strings"
	"testing"
)

func TestFailureScenario_ConfigMissingHomebrew(t *testing.T) {
	oldLookup := executableLookup
	executableLookup = func(name string) (string, error) {
		return "", exec.ErrNotFound
	}
	defer func() {
		executableLookup = oldLookup
	}()

	_, err := buildInstallOperations(false)
	if err == nil {
		t.Fatal("expected buildInstallOperations to fail without brew")
	}
	if !strings.Contains(err.Error(), "homebrew is not installed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFailureScenario_InvalidInput(t *testing.T) {
	cmd := NewListCmdWithService(testCommandService())
	cmd.SetArgs([]string{"--output", "bad-value"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected invalid argument to fail")
	}
}

func TestFailureScenario_PermissionDeniedDuringInstall(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		if name == "php" {
			return "/usr/local/bin/php", nil
		}
		return "", exec.ErrNotFound
	}
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		if name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "php" {
			return fs.ErrPermission
		}
		return nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	err := InstallPHPWithOutput(io.Discard, true)
	if err == nil {
		t.Fatal("expected permission error")
	}
	if !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("expected wrapped permission error, got: %v", err)
	}
}

func TestFailureScenario_FileNotFoundWhileResolvingVersion(t *testing.T) {
	oldRunner := commandRunner
	commandRunner = func(_ string, _ ...string) ([]byte, error) {
		return nil, exec.ErrNotFound
	}
	defer func() {
		commandRunner = oldRunner
	}()

	if _, err := ToolVersionWithPath("php", "/definitely/missing/php"); err == nil {
		t.Fatal("expected missing file to fail in ToolVersionWithPath")
	}
}
