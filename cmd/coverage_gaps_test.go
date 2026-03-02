package cmd

import (
	"fmt"
	"io"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestBuildInstallOperationsForToolRejectsUnsupportedTool(t *testing.T) {
	_, err := buildInstallOperationsForTool("ruby", false)
	if err == nil {
		t.Fatal("expected unsupported tool to return error")
	}
}

func TestBuildInstallOperationsFailsWithoutHomebrew(t *testing.T) {
	oldLookup := executableLookup
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "", exec.ErrNotFound
		}
		return "", exec.ErrNotFound
	}
	defer func() {
		executableLookup = oldLookup
	}()

	if _, err := buildInstallOperations(false); err == nil {
		t.Fatal("expected buildInstallOperations to fail when brew is missing")
	}
}

func TestBuildInstallOperationsSkipsInstalledTools(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "brew" || name == "php" || name == "node" || name == "go" || name == "curl" || name == "git" || name == "ffmpeg" || name == "tree" || name == "gh" || name == "python3" {
			return "/usr/local/bin/" + name, nil
		}
		return "", exec.ErrNotFound
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "brew" && len(args) >= 3 && args[0] == "list" && args[1] == "--formula" && args[2] == "python3" {
			return []byte("python3\n"), nil
		}
		return []byte(""), nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	ops, err := buildInstallOperations(false)
	if err != nil {
		t.Fatalf("buildInstallOperations failed: %v", err)
	}

	if len(ops) != 0 {
		t.Fatalf("expected no operations, got %v", ops)
	}
}

func TestBuildInstallOperationsIncludesCurlLinkOnlyWhenNeedInstall(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		if name == "curl" {
			return "", exec.ErrNotFound
		}
		return "/usr/local/bin/" + name, nil
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "brew" && len(args) == 3 && args[0] == "list" && args[1] == "--formula" && args[2] == "curl" {
			return []byte(""), nil
		}
		return []byte(""), nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	ops, err := buildInstallOperations(false)
	if err != nil {
		t.Fatalf("buildInstallOperations failed: %v", err)
	}

	expected := []string{"brew install curl", "brew link curl --force"}
	if !reflect.DeepEqual(ops, expected) {
		t.Fatalf("unexpected install operations, want %v got %v", expected, ops)
	}
}

func TestRunInstallOperationParsesAndForwardsArgs(t *testing.T) {
	oldRunnerWithOutput := commandRunnerWithOutput
	argsPassed := []string{}
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		argsPassed = append(argsPassed, name)
		argsPassed = append(argsPassed, args...)
		return nil
	}
	defer func() {
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	if err := runInstallOperation(io.Discard, "brew install ffmpeg"); err != nil {
		t.Fatalf("runInstallOperation failed: %v", err)
	}

	got := argsPassed
	want := []string{"brew", "install", "ffmpeg"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected forwarded args, want %v got %v", want, got)
	}
}

func TestRunInstallOperationEmptyInputIsNoop(t *testing.T) {
	if err := runInstallOperation(io.Discard, "   "); err != nil {
		t.Fatalf("expected no-op for empty op, got %v", err)
	}
}

func TestResolvedBrewBinaryPathFallsBackToOptBinPrefix(t *testing.T) {
	oldRunner := commandRunner
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "brew" && len(args) == 2 && args[0] == "--prefix" && args[1] == "ffmpeg" {
			return []byte("/opt/homebrew/opt/ffmpeg\n"), nil
		}
		return []byte(""), fmt.Errorf("unexpected command %s %v", name, args)
	}
	defer func() {
		commandRunner = oldRunner
	}()

	path, err := resolvedBrewBinaryPath("ffmpeg", "ffmpeg")
	if err != nil {
		t.Fatalf("resolvedBrewBinaryPath failed: %v", err)
	}
	if path != "/opt/homebrew/opt/ffmpeg/bin/ffmpeg" {
		t.Fatalf("unexpected resolved path %q", path)
	}
}

func TestToolVersionWithPathFallsBackToProvidedCommandPath(t *testing.T) {
	oldRunner := commandRunner
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "ffmpeg" {
			return nil, fmt.Errorf("command not available")
		}
		if name == "/opt/homebrew/bin/ffmpeg" && len(args) == 1 && args[0] == "-version" {
			return []byte("ffmpeg version 8.0.1_4"), nil
		}
		return []byte(""), nil
	}
	defer func() {
		commandRunner = oldRunner
	}()

	version, err := ToolVersionWithPath("ffmpeg", "/opt/homebrew/bin/ffmpeg")
	if err != nil {
		t.Fatalf("ToolVersionWithPath failed: %v", err)
	}
	if version != "8.0.1" {
		t.Fatalf("unexpected version %q", version)
	}
}

func TestBuildNodeInstallOperationsForceBypassesNodeAndNpm(t *testing.T) {
	oldLookup := executableLookup
	executableLookup = func(name string) (string, error) {
		if name == "brew" || name == "node" || name == "npm" {
			return "/usr/local/bin/" + name, nil
		}
		return "", exec.ErrNotFound
	}
	defer func() {
		executableLookup = oldLookup
	}()

	ops, err := buildNodeInstallOperations(true)
	if err != nil {
		t.Fatalf("buildNodeInstallOperations failed: %v", err)
	}
	if len(ops) != 1 || ops[0] != "brew install node" {
		t.Fatalf("expected node install op with force, got %v", ops)
	}
}

func TestOutdatedCommandUsesColorizedCurrentVersionWhenStale(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "tree" || name == "brew" {
			return "/usr/local/bin/" + name, nil
		}
		return "", exec.ErrNotFound
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "tree" && len(args) > 0 && args[0] == "--version" {
			return []byte("tree version 2.1.3"), nil
		}
		if name == "brew" && len(args) >= 3 && args[0] == "info" && args[1] == "--json=v2" && args[2] == "tree" {
			return []byte(`{"formulae":[{"name":"tree","versions":{"stable":"2.3.1"},"installed":[{"version":"2.1.3"}]}]}`), nil
		}
		return []byte(""), nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	cmd := NewOutdatedCmd()
	cmd.SetOut(io.Discard)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("outdated command failed: %v", err)
	}
}

func TestInstallNodeWithOutputUsesBrewInstallWhenMissing(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		if name == "brew" || name == "node" || name == "npm" || name == "php" || name == "go" || name == "curl" || name == "git" || name == "ffmpeg" || name == "tree" || name == "gh" || name == "python3" {
			if name == "brew" {
				return "/opt/homebrew/bin/brew", nil
			}
			return "", exec.ErrNotFound
		}
		return "", exec.ErrNotFound
	}
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		if name != "brew" {
			return nil
		}
		if len(args) != 2 || args[0] != "install" || args[1] != "node" {
			t.Fatalf("unexpected command: %v", append([]string{name}, args...))
		}
		return nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	if err := InstallNodeWithOutput(io.Discard, false); err != nil {
		t.Fatalf("InstallNodeWithOutput failed: %v", err)
	}
}

func TestInstallNodeWithOutputFailsWithoutHomebrew(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		return "", exec.ErrNotFound
	}
	commandRunnerWithOutputCalled := false
	commandRunnerWithOutput = func(_ io.Writer, _ string, _ ...string) error {
		commandRunnerWithOutputCalled = true
		return nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	if err := InstallNodeWithOutput(io.Discard, false); err == nil {
		t.Fatal("expected InstallNodeWithOutput to fail when brew is missing")
	}
	if commandRunnerWithOutputCalled {
		t.Fatal("did not expect install command when brew is missing")
	}
}

func TestInstallNodeWithOutputSkipsWhenInstalledWithoutForceFalse(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		if name == "brew" || name == "node" {
			return "/usr/local/bin/" + name, nil
		}
		return "", exec.ErrNotFound
	}
	commandRunnerWithOutputCalled := false
	commandRunnerWithOutput = func(_ io.Writer, _ string, _ ...string) error {
		commandRunnerWithOutputCalled = true
		return nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	if err := InstallNodeWithOutput(io.Discard, false); err == nil {
		t.Fatal("expected InstallNodeWithOutput to skip existing node install")
	}
	if commandRunnerWithOutputCalled {
		t.Fatal("did not expect install commands to be executed when node exists and force is false")
	}
}

func TestInstallCurlWithOutputIncludesLinkStep(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}

	commands := make([]string, 0)
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		commands = append(commands, fmt.Sprintf("%s %s", name, strings.Join(args, " ")))
		return nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	if err := InstallCurlWithOutput(io.Discard, false); err != nil {
		t.Fatalf("InstallCurlWithOutput failed: %v", err)
	}

	expected := []string{"brew install curl", "brew link curl --force"}
	if !reflect.DeepEqual(commands, expected) {
		t.Fatalf("unexpected curl install flow, want %v got %v", expected, commands)
	}
}

func TestInstallPython3WithOutputChecksHomebrewFormulaStatus(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	oldRunnerWithOutput := commandRunnerWithOutput
	var installCalled bool
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	installedFormula := false
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "brew" && len(args) == 3 && args[0] == "list" && args[1] == "--formula" && args[2] == "python3" {
			if installedFormula {
				return []byte("python3\n"), nil
			}
			return []byte(""), nil
		}
		return []byte(""), nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	installedFormula = true
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		installCalled = true
		if len(args) != 2 || args[0] != "install" || args[1] != "python3" {
			t.Fatalf("unexpected command: %v", append([]string{name}, args...))
		}
		return nil
	}
	if err := InstallPython3WithOutput(io.Discard, false); err == nil {
		t.Fatal("expected InstallPython3WithOutput to skip when brew formula already exists")
	}
	if installCalled {
		t.Fatal("did not expect install command to run when brew formula already exists")
	}

	installCalled = false
	installedFormula = false
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		installCalled = true
		if len(args) != 2 || args[0] != "install" || args[1] != "python3" {
			t.Fatalf("unexpected command: %v", append([]string{name}, args...))
		}
		return nil
	}

	if err := InstallPython3WithOutput(io.Discard, false); err != nil {
		t.Fatalf("expected InstallPython3WithOutput to install when formula is not present: %v", err)
	}
	if !installCalled {
		t.Fatal("expected InstallPython3WithOutput to execute brew install when formula is missing")
	}

	installedFormula = true
	commandRunnerWithOutput = func(_ io.Writer, _ string, _ ...string) error {
		installCalled = true
		return nil
	}

	if err := InstallPython3WithOutput(io.Discard, true); err != nil {
		t.Fatalf("force install should execute brew install python3: %v", err)
	}
	if !installCalled {
		t.Fatal("expected InstallPython3WithOutput(force) to execute brew install")
	}
}

func TestInstallToolRejectsUnsupportedTool(t *testing.T) {
	if err := InstallTool("ruby"); err == nil {
		t.Fatal("expected InstallTool to reject unsupported tools")
	}
}

func TestInstallPHPWithOutputRunsBrewInstallWhenMissing(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	called := false
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		called = true
		if name != "brew" || len(args) != 2 || args[0] != "install" || args[1] != "php" {
			t.Fatalf("unexpected install command: %v %v", name, args)
		}
		return nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	if err := InstallPHPWithOutput(io.Discard, false); err != nil {
		t.Fatalf("InstallPHPWithOutput failed: %v", err)
	}
	if !called {
		t.Fatal("expected brew install php to be executed")
	}
}

func TestInstallPHPWithOutputSkipsWhenInstalled(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		if name == "brew" || name == "php" {
			return "/usr/local/bin/" + name, nil
		}
		return "", exec.ErrNotFound
	}
	commandRunnerWithOutputCalled := false
	commandRunnerWithOutput = func(_ io.Writer, _ string, _ ...string) error {
		commandRunnerWithOutputCalled = true
		return nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	if err := InstallPHPWithOutput(io.Discard, false); err == nil {
		t.Fatal("expected InstallPHPWithOutput to skip when php exists")
	}
	if commandRunnerWithOutputCalled {
		t.Fatal("did not expect brew install command when php already exists")
	}
}

func TestInstallGoWithOutputRunsBrewInstallWhenMissing(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	called := false
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		called = true
		if name != "brew" || len(args) != 2 || args[0] != "install" || args[1] != "go" {
			t.Fatalf("unexpected install command: %v %v", name, args)
		}
		return nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	if err := InstallGoWithOutput(io.Discard, false); err != nil {
		t.Fatalf("InstallGoWithOutput failed: %v", err)
	}
	if !called {
		t.Fatal("expected brew install go to be executed")
	}
}

func TestInstallGitWithOutputRunsBrewInstallWhenMissing(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	called := false
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		called = true
		if name != "brew" || len(args) != 2 || args[0] != "install" || args[1] != "git" {
			t.Fatalf("unexpected install command: %v %v", name, args)
		}
		return nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	if err := InstallGitWithOutput(io.Discard, false); err != nil {
		t.Fatalf("InstallGitWithOutput failed: %v", err)
	}
	if !called {
		t.Fatal("expected brew install git to be executed")
	}
}

func TestInstallFFmpegWithOutputRunsBrewInstallWhenMissing(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	called := false
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		called = true
		if name != "brew" || len(args) != 2 || args[0] != "install" || args[1] != "ffmpeg" {
			t.Fatalf("unexpected install command: %v %v", name, args)
		}
		return nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	if err := InstallFFmpegWithOutput(io.Discard, false); err != nil {
		t.Fatalf("InstallFFmpegWithOutput failed: %v", err)
	}
	if !called {
		t.Fatal("expected brew install ffmpeg to be executed")
	}
}

func TestInstallTreeWithOutputRunsBrewInstallWhenMissing(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	called := false
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		called = true
		if name != "brew" || len(args) != 2 || args[0] != "install" || args[1] != "tree" {
			t.Fatalf("unexpected install command: %v %v", name, args)
		}
		return nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	if err := InstallTreeWithOutput(io.Discard, false); err != nil {
		t.Fatalf("InstallTreeWithOutput failed: %v", err)
	}
	if !called {
		t.Fatal("expected brew install tree to be executed")
	}
}

func TestInstallGHWithOutputRunsBrewInstallWhenMissing(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	called := false
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		called = true
		if name != "brew" || len(args) != 2 || args[0] != "install" || args[1] != "gh" {
			t.Fatalf("unexpected install command: %v %v", name, args)
		}
		return nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	if err := InstallGHWithOutput(io.Discard, false); err != nil {
		t.Fatalf("InstallGHWithOutput failed: %v", err)
	}
	if !called {
		t.Fatal("expected brew install gh to be executed")
	}
}

func TestInstallPHPExecutesBrewInstallWhenMissing(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	installed := false
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if !installed && name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "php" {
			return []byte(""), nil
		}
		return nil, nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	if err := InstallPHP(); err != nil {
		t.Fatalf("InstallPHP failed: %v", err)
	}
}

func TestInstallPHPSkipsWhenAlreadyInstalled(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "brew" || name == "php" {
			return "/usr/local/bin/" + name, nil
		}
		return "", exec.ErrNotFound
	}
	commandRunner = func(_ string, _ ...string) ([]byte, error) {
		t.Fatal("did not expect any install command when php already exists")
		return nil, nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	if err := InstallPHP(); err == nil {
		t.Fatal("expected InstallPHP to skip existing binary")
	}
}

func TestInstallPHPWithOutputFailsWhenBrewInstallFails(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	commandRunnerWithOutput = func(_ io.Writer, _ string, _ ...string) error {
		return fmt.Errorf("install failed")
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	if err := InstallPHPWithOutput(io.Discard, false); err == nil {
		t.Fatal("expected InstallPHPWithOutput to propagate install failure")
	}
}

func TestInstallGoExecutesBrewInstallWhenMissing(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "go" {
			return []byte(""), nil
		}
		return nil, nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	if err := InstallGo(); err != nil {
		t.Fatalf("InstallGo failed: %v", err)
	}
}

func TestInstallGitExecutesBrewInstallWhenMissing(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "git" {
			return []byte(""), nil
		}
		return nil, nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	if err := InstallGit(); err != nil {
		t.Fatalf("InstallGit failed: %v", err)
	}
}

func TestInstallFFmpegExecutesBrewInstallWhenMissing(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "ffmpeg" {
			return []byte(""), nil
		}
		return nil, nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	if err := InstallFFmpeg(); err != nil {
		t.Fatalf("InstallFFmpeg failed: %v", err)
	}
}

func TestInstallTreeExecutesBrewInstallWhenMissing(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "tree" {
			return []byte(""), nil
		}
		return nil, nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	if err := InstallTree(); err != nil {
		t.Fatalf("InstallTree failed: %v", err)
	}
}

func TestInstallGHExecutesBrewInstallWhenMissing(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "gh" {
			return []byte(""), nil
		}
		return nil, nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	if err := InstallGH(); err != nil {
		t.Fatalf("InstallGH failed: %v", err)
	}
}

func TestInstallGHSkipsWhenInstalled(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "brew" || name == "gh" {
			return "/usr/local/bin/" + name, nil
		}
		return "", exec.ErrNotFound
	}
	commandRunner = func(_ string, _ ...string) ([]byte, error) {
		t.Fatal("did not expect any install command when gh already exists")
		return nil, nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	if err := InstallGH(); err == nil {
		t.Fatal("expected InstallGH to skip existing binary")
	}
}

func TestInstallToolRoutesToInstallFunction(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	commandCalled := false
	commandRunner = func(name string, args ...string) ([]byte, error) {
		commandCalled = true
		if name == "brew" && len(args) == 2 && args[0] == "install" && args[1] == "node" {
			return []byte(""), nil
		}
		return nil, nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	if err := InstallTool("node"); err != nil {
		t.Fatalf("InstallTool(node) failed: %v", err)
	}
	if !commandCalled {
		t.Fatal("expected node install command through InstallTool")
	}
}

func TestColorizeWrapsWithAnsiPrefixAndReset(t *testing.T) {
	got := colorize(colorRed, "node")
	if got != "\033[31mnode\033[0m" {
		t.Fatalf("unexpected colorized output %q", got)
	}
}

func TestInstallToolRoutesEachSupportedTool(t *testing.T) {
	testCases := []struct {
		tool string
	}{
		{tool: "php"},
		{tool: "python3"},
		{tool: "node"},
		{tool: "go"},
		{tool: "curl"},
		{tool: "git"},
		{tool: "ffmpeg"},
		{tool: "tree"},
		{tool: "gh"},
	}

	for _, tc := range testCases {
		t.Run(tc.tool, func(t *testing.T) {
			oldLookup := executableLookup
			oldRunner := commandRunner
			executableLookup = func(name string) (string, error) {
				if name == "brew" {
					return "/opt/homebrew/bin/brew", nil
				}
				return "", exec.ErrNotFound
			}
			commandRunner = func(name string, args ...string) ([]byte, error) {
				if name == "brew" && len(args) >= 1 {
					cmd := name + " " + strings.Join(args, " ")
					if tc.tool == "node" {
						if cmd == "brew install node" {
							return []byte(""), nil
						}
						return nil, nil
					}
					if tc.tool == "php" && cmd == "brew install php" {
						return []byte(""), nil
					}
					if tc.tool == "python3" && args[0] == "list" && len(args) >= 3 && args[1] == "--formula" && args[2] == "python3" {
						return []byte(""), nil
					}
					if tc.tool == "python3" && args[0] == "install" && args[1] == "python3" {
						return []byte(""), nil
					}
					if tc.tool == "go" && cmd == "brew install go" {
						return []byte(""), nil
					}
					if tc.tool == "curl" && cmd == "brew install curl" {
						return []byte(""), nil
					}
					if tc.tool == "git" && cmd == "brew install git" {
						return []byte(""), nil
					}
					if tc.tool == "ffmpeg" && cmd == "brew install ffmpeg" {
						return []byte(""), nil
					}
					if tc.tool == "tree" && cmd == "brew install tree" {
						return []byte(""), nil
					}
					if tc.tool == "gh" && cmd == "brew install gh" {
						return []byte(""), nil
					}
				}
				return nil, nil
			}
			defer func() {
				executableLookup = oldLookup
				commandRunner = oldRunner
			}()

			if err := InstallTool(tc.tool); err != nil {
				t.Fatalf("InstallTool(%s) failed: %v", tc.tool, err)
			}
		})
	}
}

func TestUpdateToolWithOutputUpdatesBrewInstalledTool(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	oldRunner := commandRunner

	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		if name == "php" {
			return "/usr/local/bin/php", nil
		}
		return "", exec.ErrNotFound
	}

	commands := make([]string, 0)
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		commands = append(commands, strings.TrimSpace(name+" "+strings.Join(args, " ")))
		if name == "brew" && len(args) == 2 && args[0] == "upgrade" && args[1] == "php" {
			return nil
		}
		return nil
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
		commandRunner = oldRunner
	}()

	if err := UpdateToolWithOutput(io.Discard, "php"); err != nil {
		t.Fatalf("UpdateToolWithOutput failed: %v", err)
	}
	if got := strings.Join(commands, ","); got != "brew upgrade php" {
		t.Fatalf("expected update command, got %q", got)
	}
}

func TestUpdateToolWithOutputUpdatesNpmWithoutBrew(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	oldRunner := commandRunner

	executableLookup = func(name string) (string, error) {
		if name == "npm" {
			return "/usr/local/bin/npm", nil
		}
		return "", exec.ErrNotFound
	}
	commands := make([]string, 0)
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		commands = append(commands, strings.TrimSpace(name+" "+strings.Join(args, " ")))
		return nil
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
		commandRunner = oldRunner
	}()

	if err := UpdateToolWithOutput(io.Discard, "npm"); err != nil {
		t.Fatalf("UpdateToolWithOutput(npm) failed: %v", err)
	}
	if got := strings.Join(commands, ","); got != "npm install -g npm@latest" {
		t.Fatalf("expected npm update command, got %q", got)
	}
}

func TestUpdateToolWithOutputFailsWhenToolNotInstalled(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		return "", exec.ErrNotFound
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	if err := UpdateToolWithOutput(io.Discard, "php"); err == nil {
		t.Fatal("expected UpdateToolWithOutput to fail when tool is not installed")
	}
}

func TestUpdateToolWithOutputFailsWhenBrewMissingForBrewTool(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	oldRunner := commandRunner

	executableLookup = func(name string) (string, error) {
		if name == "php" {
			return "/usr/local/bin/php", nil
		}
		return "", exec.ErrNotFound
	}
	upgradeCalled := false
	commandRunnerWithOutput = func(_ io.Writer, _ string, _ ...string) error {
		upgradeCalled = true
		return nil
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
		commandRunner = oldRunner
	}()

	if err := UpdateToolWithOutput(io.Discard, "php"); err == nil {
		t.Fatal("expected UpdateToolWithOutput to fail when brew missing")
	}
	if upgradeCalled {
		t.Fatal("did not expect upgrade command without brew")
	}
}

func TestUpdateToolWithOutputFailsWhenUpgradeFails(t *testing.T) {
	oldLookup := executableLookup
	oldRunnerWithOutput := commandRunnerWithOutput
	oldRunner := commandRunner

	executableLookup = func(name string) (string, error) {
		if name == "brew" || name == "php" {
			return "/opt/homebrew/bin/" + name, nil
		}
		return "", exec.ErrNotFound
	}
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		if name == "brew" && len(args) == 2 && args[0] == "upgrade" && args[1] == "php" {
			return fmt.Errorf("upgrade failed")
		}
		return nil
	}
	commandRunner = func(name string, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunnerWithOutput = oldRunnerWithOutput
		commandRunner = oldRunner
	}()

	if err := UpdateToolWithOutput(io.Discard, "php"); err == nil {
		t.Fatal("expected UpdateToolWithOutput to propagate upgrade errors")
	}
}
