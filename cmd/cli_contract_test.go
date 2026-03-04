package cmd

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runCLIBinary(t *testing.T, args ...string) (stdout string, stderr string, exitCode int) {
	t.Helper()
	workdir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("resolve cwd failed: %v", err)
	}
	projectRoot := filepath.Dir(workdir)

	cmd := exec.Command("go", append([]string{"run", "./cmd/denv"}, args...)...)
	cmd.Dir = projectRoot
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), "", 0
	}

	exitCode = 1
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	}
	return "", string(out), exitCode
}

func TestContract_RootHelpHasEntrypointsAndExitZero(t *testing.T) {
	out, _, code := runCLIBinary(t, "--help")
	if code != 0 {
		t.Fatalf("expect exit code 0, got %d", code)
	}
	if !strings.Contains(out, "list") || !strings.Contains(out, "install") || !strings.Contains(out, "outdated") || !strings.Contains(out, "update") {
		t.Fatalf("unexpected root help output: %q", out)
	}
}

func TestContract_InvalidListOutputArgReturnsNonZeroExitAndUsage(t *testing.T) {
	_, stderr, code := runCLIBinary(t, "list", "--output", "invalid")
	if code == 0 {
		t.Fatalf("expected non-zero exit code")
	}
	if !strings.Contains(stderr, "invalid output") {
		t.Fatalf("expected invalid output error, got: %q", stderr)
	}
}

func TestContract_ListCommandArgBehaviorPreserved(t *testing.T) {
	cmd := NewListCmdWithService(NewCLIContextWithRuntime(testRuntime()).CatalogContext)
	cmd.SetArgs([]string{"--version", "--path"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list with known args should pass: %v", err)
	}
}

func TestContract_OutdatedCommandArgBehaviorRejectsBadMode(t *testing.T) {
	cmd := NewOutdatedCmd()
	if err := cmd.ParseFlags([]string{"--output", "invalid"}); err != nil {
		if strings.Contains(err.Error(), "invalid") {
			return
		}
		t.Fatalf("unexpected parse error: %v", err)
	}
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected execute error for invalid output mode")
	}
}

func TestContract_CLIExitCodeInProcessForUnknownCommand(t *testing.T) {
	root := NewRootCmdWithContext(NewCLIContextWithRuntime(testRuntime()))
	root.SetArgs([]string{"does-not-exist"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected unknown command error")
	}
}
