package cmd

import (
    "bytes"
    "fmt"
    "io"
    "os/exec"
    "strings"
    "testing"
)

func TestNewListCmdDefaultShowsToolsOnly(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(name string) (string, error) {
        if name == "php" || name == "go" || name == "python3" {
            return "/usr/bin/" + name, nil
        }
        return "", exec.ErrNotFound
    }
    commandRunner = func(name string, args ...string) ([]byte, error) {
        switch name {
        case "php":
            return []byte("PHP 8.3.4 (cli) (built: Jan  1 2025 00:00:00)"), nil
        case "python3":
            return []byte("Python 3.12.4"), nil
        case "go":
            return []byte("go version go1.23.4 darwin/arm64"), nil
        default:
            return nil, nil
        }
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    cmd := NewListCmd()
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("list command failed: %v", err)
    }

    got := strings.TrimSpace(out.String())
    want := "php\npython3\nnode\ngo\nnpm\ncurl\ngh\ngit\nffmpeg\ntree"
    if got != want {
        t.Fatalf("unexpected list output:\nwant:\n%q\ngot:\n%q", want, got)
    }
}

func TestNewListCmdWithVersionAndPath(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(name string) (string, error) {
        if name == "php" || name == "go" || name == "python3" {
            return "/usr/bin/" + name, nil
        }
        return "", exec.ErrNotFound
    }
    commandRunner = func(name string, args ...string) ([]byte, error) {
        switch name {
        case "php":
            return []byte("PHP 8.3.4 (cli) (built: Jan  1 2025 00:00:00)"), nil
        case "python3":
            return []byte("Python 3.12.4"), nil
        case "go":
            return []byte("go version go1.23.4 darwin/arm64"), nil
        default:
            return nil, nil
        }
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    cmd := NewListCmd()
    cmd.SetOut(&bytes.Buffer{})
    cmd.SetArgs([]string{"--version", "--path"})
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("list command failed: %v", err)
    }

    got := strings.TrimSpace(out.String())
    want := "php 8.3.4 (/usr/bin/php)\npython3 3.12.4 (/usr/bin/python3)\nnode not found\ngo 1.23.4 (/usr/bin/go)\nnpm not found\ncurl not found\ngh not found\ngit not found\nffmpeg not found\ntree not found"
    if got != want {
        t.Fatalf("unexpected list output:\nwant:\n%q\ngot:\n%q", want, got)
    }
}

func TestNewListCmdWithVersionOnly(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(name string) (string, error) {
        if name == "php" || name == "go" || name == "python3" {
            return "/usr/bin/" + name, nil
        }
        return "", exec.ErrNotFound
    }
    commandRunner = func(name string, args ...string) ([]byte, error) {
        switch name {
        case "php":
            return []byte("PHP 8.3.4 (cli) (built: Jan  1 2025 00:00:00)"), nil
        case "python3":
            return []byte("Python 3.12.4"), nil
        case "go":
            return []byte("go version go1.23.4 darwin/arm64"), nil
        default:
            return nil, nil
        }
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    cmd := NewListCmd()
    out := &bytes.Buffer{}
    cmd.SetOut(out)
    cmd.SetArgs([]string{"--version"})

    if err := cmd.Execute(); err != nil {
        t.Fatalf("list command failed: %v", err)
    }

    got := strings.TrimSpace(out.String())
    want := "php 8.3.4\npython3 3.12.4\nnode not found\ngo 1.23.4\nnpm not found\ncurl not found\ngh not found\ngit not found\nffmpeg not found\ntree not found"
    if got != want {
        t.Fatalf("unexpected list output:\nwant:\n%q\ngot:\n%q", want, got)
    }
}

func TestNewListCmdWithPathOnly(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(name string) (string, error) {
        if name == "php" || name == "go" || name == "python3" {
            return "/usr/bin/" + name, nil
        }
        return "", exec.ErrNotFound
    }
    commandRunner = func(name string, args ...string) ([]byte, error) {
        switch name {
        case "php":
            return []byte("PHP 8.3.4 (cli) (built: Jan  1 2025 00:00:00)"), nil
        case "python3":
            return []byte("Python 3.12.4"), nil
        case "go":
            return []byte("go version go1.23.4 darwin/arm64"), nil
        default:
            return nil, nil
        }
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    cmd := NewListCmd()
    out := &bytes.Buffer{}
    cmd.SetOut(out)
    cmd.SetArgs([]string{"--path"})

    if err := cmd.Execute(); err != nil {
        t.Fatalf("list command failed: %v", err)
    }

    got := strings.TrimSpace(out.String())
    want := "php /usr/bin/php\npython3 /usr/bin/python3\nnode not found\ngo /usr/bin/go\nnpm not found\ncurl not found\ngh not found\ngit not found\nffmpeg not found\ntree not found"
    if got != want {
        t.Fatalf("unexpected list output:\nwant:\n%q\ngot:\n%q", want, got)
    }
}

func TestNewListCmdShowsBrewInstalledToolWithInferredPath(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(string) (string, error) {
        return "", exec.ErrNotFound
    }
    commandRunner = func(name string, args ...string) ([]byte, error) {
        if name == "brew" {
            if len(args) == 3 && args[0] == "list" && args[1] == "--formula" && args[2] == "ffmpeg" {
                return []byte("ffmpeg\n"), nil
            }
            if len(args) == 2 && args[0] == "--prefix" && args[1] == "ffmpeg" {
                return []byte("/opt/homebrew/opt/ffmpeg\n"), nil
            }
        }
        if name == "/opt/homebrew/bin/ffmpeg" && len(args) > 0 && args[0] == "-version" {
            return []byte("ffmpeg version 8.0.1"), nil
        }
        if name == "/opt/homebrew/opt/ffmpeg/bin/ffmpeg" && len(args) > 0 && args[0] == "-version" {
            return []byte("ffmpeg version 8.0.1"), nil
        }
        return []byte(""), nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    cmd := NewListCmd()
    cmd.SetOut(&bytes.Buffer{})
    cmd.SetArgs([]string{"--version", "--path"})
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("list command failed: %v", err)
    }

    got := strings.TrimSpace(out.String())
    if !strings.Contains(got, "ffmpeg") {
        t.Fatalf("expected list output to include ffmpeg, got %q", got)
    }
    if !strings.Contains(got, "8.0.1") {
        t.Fatalf("expected ffmpeg version from brew-installed binary, got %q", got)
    }
}

func TestExtractVersion(t *testing.T) {
    got, err := extractVersion("go version go1.23.4 darwin/arm64")
    if err != nil {
        t.Fatalf("extract version failed: %v", err)
    }
    if got != "1.23.4" {
        t.Fatalf("expected version 1.23.4, got %q", got)
    }
}

func TestToolVersionUsesCommandForFfmpeg(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(name string) (string, error) {
        if name == "ffmpeg" {
            return "/opt/homebrew/bin/ffmpeg", nil
        }
        return "", exec.ErrNotFound
    }
    commandCalled := []string{}
    commandRunner = func(name string, args ...string) ([]byte, error) {
        commandCalled = append(commandCalled, name+" "+strings.Join(args, " "))
        if name == "ffmpeg" && len(args) > 0 && args[0] == "-version" {
            return []byte("ffmpeg version 8.0.1_1"), nil
        }
        return []byte(""), nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    got, err := ToolVersion("ffmpeg")
    if err != nil {
        t.Fatalf("ToolVersion(ffmpeg) failed: %v", err)
    }
    if got != "8.0.1" {
        t.Fatalf("expected version from ffmpeg command, got %q", got)
    }

    calledFfmpeg := false
    calledBrewInfoJSON := false
    for _, call := range commandCalled {
        if call == "ffmpeg -version" {
            calledFfmpeg = true
        }
        if call == "brew info --json=v2 ffmpeg" {
            calledBrewInfoJSON = true
        }
    }
    if !calledFfmpeg {
        t.Fatal("expected ffmpeg binary to be called for list version")
    }
    if calledBrewInfoJSON {
        t.Fatal("expected no brew json call for list version")
    }
}

func TestToolVersionForOutdatedUsesBrewForFfmpeg(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(name string) (string, error) {
        if name == "ffmpeg" || name == "brew" {
            return "/opt/homebrew/bin/" + name, nil
        }
        return "", exec.ErrNotFound
    }

    commandCalled := []string{}
    commandRunner = func(name string, args ...string) ([]byte, error) {
        commandCalled = append(commandCalled, name+" "+strings.Join(args, " "))
        if name == "brew" && len(args) >= 2 && args[0] == "info" && args[1] == "ffmpeg" {
            return []byte(`ffmpeg ✔: stable 8.0.1 (bottled), HEAD
https://ffmpeg.org/
Installed
/opt/homebrew/Cellar/ffmpeg/7.1.1_3 (287 files, 54.8MB)
/opt/homebrew/Cellar/ffmpeg/8.0_1 (285 files, 55.3MB) *`), nil
        }
        return []byte(""), nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    got, err := ToolVersionForOutdated("ffmpeg")
    if err != nil {
        t.Fatalf("ToolVersionForOutdated(ffmpeg) failed: %v", err)
    }
    if got != "8.0_1" {
        t.Fatalf("expected version from brew, got %q", got)
    }

    calledBrewInfo := false
    calledFfmpeg := false
    for _, call := range commandCalled {
        if call == "brew info ffmpeg" {
            calledBrewInfo = true
        }
        if strings.HasPrefix(call, "ffmpeg ") {
            calledFfmpeg = true
        }
    }
    if !calledBrewInfo {
        t.Fatal("expected to call brew info ffmpeg")
    }
    if calledFfmpeg {
        t.Fatal("expected not to call ffmpeg binary directly for outdated version")
    }
}

func TestIsInstallableTool(t *testing.T) {
    if !IsInstallableTool("php") {
        t.Fatal("expected php to be installable")
    }

    if !IsInstallableTool("curl") {
        t.Fatal("expected curl to be installable")
    }

    if !IsInstallableTool("git") {
        t.Fatal("expected git to be installable")
    }

    if IsInstallableTool("npm") {
        t.Fatal("expected npm to be not installable")
    }

    if IsInstallableTool("ruby") {
        t.Fatal("expected ruby to be unsupported")
    }
}

func TestRootHasListCommand(t *testing.T) {
	cmd := NewRootCmd()
	found := false
	installFound := false
    outdatedFound := false
    updateFound := false
    for _, sub := range cmd.Commands() {
        if sub.Name() == "list" {
            found = true
        }
        if sub.Name() == "install" {
            installFound = true
        }
        if sub.Name() == "outdated" {
            outdatedFound = true
        }
        if sub.Name() == "update" {
            updateFound = true
        }
    }
    if !found {
        t.Fatal("root command should include list subcommand")
    }
    if !installFound {
        t.Fatal("root command should include install subcommand")
    }
    if !outdatedFound {
        t.Fatal("root command should include outdated subcommand")
    }
	if !updateFound {
		t.Fatal("root command should include update subcommand")
	}
}

func TestRootHasVerboseFlag(t *testing.T) {
	cmd := NewRootCmd()
	if flag := cmd.PersistentFlags().Lookup("verbose"); flag == nil {
		t.Fatal("root command should expose persistent verbose flag")
	}
}

func TestRootVerboseRunsAndLogsInstallDryRun(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/" + name, nil
		}
		return "", exec.ErrNotFound
	}
	commandRunner = func(_ string, _ ...string) ([]byte, error) {
		return nil, nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--verbose", "install", "--dry-run"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("root verbose install dry-run failed: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "[verbose] planned") {
		t.Fatalf("expected verbose logs, got: %q", got)
	}
	if !strings.Contains(got, "[verbose] dry-run operation") {
		t.Fatalf("expected dry-run verbose logs, got: %q", got)
	}
	if !strings.Contains(got, "Would run: brew install php") {
		t.Fatalf("expected dry-run operations output, got: %q", got)
	}
}

func TestUpdateCommandUpdatesOnlyOutdatedTools(t *testing.T) {
	oldLookup := executableLookup
    oldRunner := commandRunner
    oldRunnerWithOutput := commandRunnerWithOutput
    executableLookup = func(name string) (string, error) {
        if name == "brew" || name == "php" || name == "node" {
            return "/usr/local/bin/" + name, nil
        }
        return "", exec.ErrNotFound
    }
    upgradeCalled := false
    npmUpdateCalled := false
    commandRunner = func(name string, args ...string) ([]byte, error) {
        if name == "php" && len(args) == 1 && args[0] == "--version" {
            return []byte("PHP 7.4.0"), nil
        }
        if name == "node" && len(args) == 1 && args[0] == "--version" {
            return []byte("v20.0.0"), nil
        }
        if name == "npm" && len(args) == 1 && args[0] == "--version" {
            return []byte("9.0.0"), nil
        }
        if name == "brew" && len(args) >= 3 && args[0] == "info" && args[1] == "--json=v2" && args[2] == "php" {
            return []byte(`{"formulae":[{"name":"php","versions":{"stable":"8.0.0"},"installed":[{"version":"7.4.0"}]}]}`), nil
        }
        if name == "brew" && len(args) >= 3 && args[0] == "info" && args[1] == "--json=v2" && args[2] == "node" {
            return []byte(`{"formulae":[{"name":"node","versions":{"stable":"20.0.0"},"installed":[{"version":"20.0.0"}]}]}`), nil
        }
        return []byte(""), nil
    }
    commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
        if name == "brew" && len(args) == 2 && args[0] == "upgrade" && args[1] == "php" {
            upgradeCalled = true
            return nil
        }
        if name == "npm" && len(args) == 3 && args[0] == "install" && args[1] == "-g" && args[2] == "npm@latest" {
            npmUpdateCalled = true
            return nil
        }
        return nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
        commandRunnerWithOutput = oldRunnerWithOutput
    }()

    cmd := NewUpdateCmd()
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("update command failed: %v", err)
    }

    if !upgradeCalled {
        t.Fatal("expected php to be upgraded because outdated")
    }
    if npmUpdateCalled {
        t.Fatal("expected npm to be skipped when not installed in this input")
    }
    if got := out.String(); strings.Contains(got, "no updates") {
        t.Fatalf("unexpected no updates output when updates exist: %q", got)
    }
}

func TestUpdateCommandHasNoUpdates(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    oldRunnerWithOutput := commandRunnerWithOutput
    executableLookup = func(name string) (string, error) {
        if name == "brew" || name == "php" {
            return "/usr/local/bin/" + name, nil
        }
        return "", exec.ErrNotFound
    }
    commandCalls := 0
    commandRunner = func(name string, args ...string) ([]byte, error) {
        if name == "php" && len(args) == 1 && args[0] == "--version" {
            return []byte("7.4.0"), nil
        }
        if name == "brew" && len(args) >= 3 && args[0] == "info" && args[1] == "--json=v2" && args[2] == "php" {
            return []byte(`{"formulae":[{"name":"php","versions":{"stable":"7.4.0"},"installed":[{"version":"7.4.0"}]}]}`), nil
        }
        return []byte(""), nil
    }
    commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
        commandCalls++
        return nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
        commandRunnerWithOutput = oldRunnerWithOutput
    }()

    cmd := NewUpdateCmd()
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("update command failed: %v", err)
    }

    if commandCalls != 0 {
        t.Fatalf("expected no update actions, got %d", commandCalls)
    }
    if got := strings.TrimSpace(out.String()); got != "no updates available" {
        t.Fatalf("unexpected output, got %q", got)
    }
}

func TestUpdateCommandUpdatesBrewInstalledToolWithoutPath(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	oldRunnerWithOutput := commandRunnerWithOutput
	executableLookup = func(name string) (string, error) {
		if name == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", exec.ErrNotFound
	}
	upgradeCalled := false
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "brew" {
			if len(args) == 3 && args[0] == "list" && args[1] == "--formula" {
				if args[2] == "php" {
					return []byte("php\n"), nil
				}
				return []byte(""), nil
			}

			if len(args) >= 3 && args[0] == "info" && args[1] != "--json=v2" && args[2] == "php" {
				return []byte(`Installed
/opt/homebrew/Cellar/php/8.0.0 (123 files)`), nil
			}

			if len(args) >= 3 && args[0] == "info" && args[1] == "--json=v2" && args[2] == "php" {
				return []byte(`{"formulae":[{"name":"php","versions":{"stable":"8.1.0"},"installed":[{"version":"8.0.0"}]}]}`), nil
			}
		}
		return []byte(""), nil
	}
	commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
		if name == "brew" && len(args) == 2 && args[0] == "upgrade" && args[1] == "php" {
			upgradeCalled = true
			return nil
		}
		return nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
		commandRunnerWithOutput = oldRunnerWithOutput
	}()

	cmd := NewUpdateCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("update command failed: %v", err)
	}

	if !upgradeCalled {
		t.Fatal("expected brew upgrade php to be called")
	}

	if got := strings.TrimSpace(out.String()); got == "no updates available" {
		t.Fatalf("unexpected no updates output, got: %q", got)
	}
}

func TestInstallCommandRejectsArguments(t *testing.T) {
    cmd := NewInstallCmd()
    cmd.SetArgs([]string{"php"})
    if err := cmd.Execute(); err == nil {
        t.Fatal("expected extra arguments to be rejected")
    }
}

func TestInstallNodeChecksExistingInstallations(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(name string) (string, error) {
        if name == "brew" || name == "node" {
            return "/usr/local/bin/" + name, nil
        }
        return "", exec.ErrNotFound
    }
    commandCalls := 0
    commandRunner = func(_ string, _ ...string) ([]byte, error) {
        commandCalls++
        return nil, nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    if err := InstallNode(); err == nil {
        t.Fatal("expected node install to fail when node is already installed")
    }
    if commandCalls != 0 {
        t.Fatalf("unexpected command invocations, got %d", commandCalls)
    }
}

func TestInstallCommandForceBypassesExistingInstallationCheck(t *testing.T) {
    oldLookup := executableLookup
    oldRunnerWithOutput := commandRunnerWithOutput
    executableLookup = func(name string) (string, error) {
        if name == "brew" || name == "node" || name == "npm" {
            return "/usr/local/bin/" + name, nil
        }
        return "", exec.ErrNotFound
    }
    operationCount := 0
    commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
        if name == "brew" && len(args) > 0 && args[0] == "install" {
            operationCount++
        }
        return nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunnerWithOutput = oldRunnerWithOutput
    }()

    cmd := NewInstallCmd()
    cmd.SetArgs([]string{"--force"})
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("install command failed with --force: %v", err)
    }

    if operationCount == 0 {
        t.Fatal("expected install operations to be executed")
    }
}

func TestInstallCurlChecksExistingInstallation(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(name string) (string, error) {
        if name == "brew" || name == "curl" {
            return "/usr/local/bin/" + name, nil
        }
        return "", exec.ErrNotFound
    }
    commandCalls := 0
    commandRunner = func(_ string, _ ...string) ([]byte, error) {
        commandCalls++
        return nil, nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    if err := InstallCurl(); err == nil {
        t.Fatal("expected curl install to fail when curl is already installed")
    }
    if commandCalls != 0 {
        t.Fatalf("unexpected command invocations, got %d", commandCalls)
    }
}

func TestInstallGitChecksExistingInstallation(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(name string) (string, error) {
        if name == "brew" || name == "git" {
            return "/usr/local/bin/" + name, nil
        }
        return "", exec.ErrNotFound
    }
    commandCalls := 0
    commandRunner = func(_ string, _ ...string) ([]byte, error) {
        commandCalls++
        return nil, nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    if err := InstallGit(); err == nil {
        t.Fatal("expected git install to fail when git is already installed")
    }
    if commandCalls != 0 {
        t.Fatalf("unexpected command invocations, got %d", commandCalls)
    }
}

func TestInstallPython3OnlyHonorsHomebrewInstall(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(name string) (string, error) {
        if name == "brew" {
            return "/opt/homebrew/bin/brew", nil
        }
        if name == "python3" {
            return "/usr/bin/python3", nil
        }
        return "", exec.ErrNotFound
    }

    invokedInstall := false
    commandRunner = func(name string, args ...string) ([]byte, error) {
        if name == "brew" && len(args) > 0 && args[0] == "list" {
            return []byte(""), fmt.Errorf("exit status 1")
        }
        if name == "brew" && len(args) > 0 && args[0] == "install" && len(args) == 2 && args[1] == "python3" {
            invokedInstall = true
            return []byte(""), nil
        }
        return nil, nil
    }

    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    if err := InstallPython3(); err != nil {
        t.Fatalf("expected python3 install to run with brew formula missing, got: %v", err)
    }
    if !invokedInstall {
        t.Fatal("expected brew install python3 command to be invoked")
    }
}

func TestInstallCommandShowsHomebrewOutput(t *testing.T) {
    oldLookup := executableLookup
    oldRunnerWithOutput := commandRunnerWithOutput
    executableLookup = func(name string) (string, error) {
        if name == "brew" {
            return "/opt/homebrew/bin/brew", nil
        }
        if name == "node" || name == "npm" || name == "php" || name == "go" || name == "curl" || name == "git" || name == "ffmpeg" || name == "tree" || name == "gh" || name == "python3" {
            return "", exec.ErrNotFound
        }
        return "", exec.ErrNotFound
    }
    commandRunnerWithOutput = func(out io.Writer, name string, args ...string) error {
        if name == "brew" && len(args) > 0 && args[0] == "install" {
            _, _ = out.Write([]byte("Homebrew output: success\n"))
            return nil
        }
        return nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunnerWithOutput = oldRunnerWithOutput
    }()

    cmd := NewInstallCmd()
    cmd.SetArgs([]string{})
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("install command failed: %v", err)
    }

    got := out.String()
    if got == "" {
        t.Fatal("expected install command output to include brew output, got empty output")
    }
	if !strings.Contains(got, "Homebrew output: success") {
        t.Fatalf("expected brew output to be shown, got: %q", got)
    }
}

func TestInstallCommandCurlLinksAfterInstall(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    oldRunnerWithOutput := commandRunnerWithOutput
    executableLookup = func(name string) (string, error) {
        if name == "brew" {
            return "/opt/homebrew/bin/brew", nil
        }
        if name == "curl" {
            return "", exec.ErrNotFound
        }
        return "", exec.ErrNotFound
    }
    commandRunner = func(_ string, _ ...string) ([]byte, error) {
        return nil, nil
    }

    installCalled := false
    linkCalled := false
    commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
        if name != "brew" {
            return nil
        }
        if len(args) == 2 && args[0] == "install" && args[1] == "curl" {
            installCalled = true
        }
        if len(args) == 3 && args[0] == "link" && args[1] == "curl" && args[2] == "--force" {
            linkCalled = true
        }
        return nil
    }

    defer func() {
        executableLookup = oldLookup
        commandRunnerWithOutput = oldRunnerWithOutput
        commandRunner = oldRunner
    }()

    cmd := NewInstallCmd()
    cmd.SetArgs([]string{})
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("install command failed: %v", err)
    }

    if !installCalled {
        t.Fatal("expected brew install curl command to be invoked")
    }
    if !linkCalled {
        t.Fatal("expected brew link curl --force command to be invoked")
    }
}

func TestInstallCommandDryRunShowsOperationsOnly(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    oldRunnerWithOutput := commandRunnerWithOutput
    executableLookup = func(name string) (string, error) {
        if name == "brew" {
            return "/opt/homebrew/bin/brew", nil
        }
        if name == "node" || name == "npm" || name == "php" || name == "go" || name == "curl" || name == "git" || name == "ffmpeg" || name == "tree" || name == "gh" || name == "python3" {
            return "", exec.ErrNotFound
        }
        return "", exec.ErrNotFound
    }
    commandRunner = func(_ string, _ ...string) ([]byte, error) {
        return []byte(""), nil
    }
    commandRunnerWithOutputCalled := false
    commandRunnerWithOutput = func(_ io.Writer, _ string, _ ...string) error {
        commandRunnerWithOutputCalled = true
        return nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
        commandRunnerWithOutput = oldRunnerWithOutput
    }()

    cmd := NewInstallCmd()
    cmd.SetArgs([]string{"--dry-run"})
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("install command failed: %v", err)
    }

    if commandRunnerWithOutputCalled {
        t.Fatal("expected dry-run to not execute install operations")
    }

	got := strings.TrimSpace(out.String())
	expected := strings.Join([]string{
		"Would run: brew install php",
		"Would run: brew install python3",
		"Would run: brew install node",
		"Would run: brew install go",
		"Would run: brew install curl",
		"Would run: brew link curl --force",
		"Would run: brew install git",
		"Would run: brew install ffmpeg",
		"Would run: brew install tree",
		"Would run: brew install gh",
	}, "\n")
	if got != expected {
		t.Fatalf("unexpected dry-run output, expected %q got %q", expected, got)
	}
}

func TestInstallCommandDryRunSkipsPython3AndInstallsFFmpegAndTree(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    oldRunnerWithOutput := commandRunnerWithOutput
    executableLookup = func(name string) (string, error) {
        switch name {
        case "brew":
            return "/opt/homebrew/bin/brew", nil
        case "python3":
            return "/usr/bin/python3", nil
        case "ffmpeg", "tree":
            return "", exec.ErrNotFound
        default:
            return "/usr/bin/" + name, nil
        }
    }
    commandRunner = func(_ string, _ ...string) ([]byte, error) {
        return []byte(""), nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
        commandRunnerWithOutput = oldRunnerWithOutput
    }()

    cmd := NewInstallCmd()
    cmd.SetArgs([]string{"--dry-run"})
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("install command failed: %v", err)
    }

    got := strings.TrimSpace(out.String())
    expected := strings.Join([]string{
        "Would run: brew install ffmpeg",
        "Would run: brew install tree",
    }, "\n")
    if got != expected {
        t.Fatalf("unexpected dry-run output, expected %q got %q", expected, got)
    }
}

func TestInstallCommandInstallsAllTools(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    oldRunnerWithOutput := commandRunnerWithOutput
    executableLookup = func(name string) (string, error) {
        if name == "brew" {
            return "/opt/homebrew/bin/brew", nil
        }
        return "", exec.ErrNotFound
    }

    installed := map[string]int{}
    linked := false
    commandRunner = func(_ string, _ ...string) ([]byte, error) {
        return nil, nil
    }
    commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
        if name != "brew" {
            return nil
        }
        if len(args) == 2 && args[0] == "install" {
            installed[args[1]]++
            return nil
        }
        if len(args) == 3 && args[0] == "link" && args[1] == "curl" && args[2] == "--force" {
            linked = true
            return nil
        }
        return nil
    }

    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
        commandRunnerWithOutput = oldRunnerWithOutput
    }()

    cmd := NewInstallCmd()
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("install command failed: %v", err)
    }

    for _, name := range InstallableTools() {
        if installed[name] != 1 {
            t.Fatalf("expected install command for %q to run exactly once, got %d; all installs: %v", name, installed[name], installed)
        }
    }

    for name, count := range installed {
        if count != 1 {
            t.Fatalf("unexpected install count %q: %d", name, count)
        }
    }

    if !linked {
        t.Fatal("expected brew link curl --force to be invoked")
    }
}

func TestInstallCommandInstallTreeUsesBrewTreeFormula(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    oldRunnerWithOutput := commandRunnerWithOutput
    executableLookup = func(name string) (string, error) {
        if name == "brew" {
            return "/opt/homebrew/bin/brew", nil
        }
        return "", exec.ErrNotFound
    }
    commandRunner = func(_ string, _ ...string) ([]byte, error) {
        return nil, nil
    }
    installs := []string{}
    commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
        if name == "brew" && len(args) == 2 && args[0] == "install" {
            installs = append(installs, args[1])
        }
        if name == "brew" && len(args) == 3 && args[0] == "link" && args[1] == "curl" && args[2] == "--force" {
            installs = append(installs, "curl-link-force")
        }
        return nil
    }

    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
        commandRunnerWithOutput = oldRunnerWithOutput
    }()

    cmd := NewInstallCmd()
    cmd.SetOut(&bytes.Buffer{})

    if err := cmd.Execute(); err != nil {
        t.Fatalf("install command failed: %v", err)
    }

    for _, name := range InstallableTools() {
        if name == "curl" {
            continue
        }
        expected := 0
        for _, install := range installs {
            if install == name {
                expected++
            }
        }
        if expected != 1 {
            t.Fatalf("expected install %q once in install sequence, got %d, installs: %v", name, expected, installs)
        }
    }

    treeCount := 0
    pythonCount := 0
    for _, install := range installs {
        if install == "tree" {
            treeCount++
        }
        if install == "python3" {
            pythonCount++
        }
    }
    if treeCount != 1 {
        t.Fatalf("expected tree to be installed once, got %d, installs: %v", treeCount, installs)
    }
    if pythonCount != 1 {
        t.Fatalf("expected python3 to be installed once, got %d, installs: %v", pythonCount, installs)
    }
}
