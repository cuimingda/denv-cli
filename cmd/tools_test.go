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

func TestExtractVersion(t *testing.T) {
    got, err := extractVersion("go version go1.23.4 darwin/arm64")
    if err != nil {
        t.Fatalf("extract version failed: %v", err)
    }
    if got != "1.23.4" {
        t.Fatalf("expected version 1.23.4, got %q", got)
    }
}

func TestToolVersionUsesBrewForFfmpeg(t *testing.T) {
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
        if name == "brew" && len(args) >= 2 && args[0] == "info" && args[1] == "ffmpeg" {
            return []byte(`ffmpeg ✔: stable 8.0.1 (bottled), HEAD
https://ffmpeg.org/
Installed
/opt/homebrew/Cellar/ffmpeg/7.1.1_3 (287 files, 54.8MB)
/opt/homebrew/Cellar/ffmpeg/8.0_1 (285 files, 55.3MB) *`), nil
        }
        if name == "brew" && len(args) >= 3 && args[0] == "info" && args[1] == "--json=v2" && args[2] == "ffmpeg" {
            return []byte(`{"formulae":[{"name":"ffmpeg","versions":{"stable":"8.0.1"}}]}`), nil
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
    if got != "8.0_1" {
        t.Fatalf("expected version from brew, got %q", got)
    }

    calledBrewInfo := false
    calledBrewInfoJSON := false
    for _, call := range commandCalled {
        if call == "brew info ffmpeg" {
            calledBrewInfo = true
        }
        if call == "brew info --json=v2 ffmpeg" {
            calledBrewInfoJSON = true
        }
        if strings.HasPrefix(call, "ffmpeg ") {
            t.Fatalf("expected not to call ffmpeg binary directly for version, got %q", call)
        }
    }
    if !calledBrewInfo {
        t.Fatal("expected to call brew info ffmpeg")
    }
    if calledBrewInfoJSON {
        t.Fatal("unexpected fallback to JSON parser while installed path exists")
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
}

func TestInstallCommandRejectsUnsupportedTool(t *testing.T) {
    cmd := NewInstallCmd()
    cmd.SetArgs([]string{"rust"})
    if err := cmd.Execute(); err == nil {
        t.Fatal("expected unsupported tool to fail")
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
    commandInstalled := false
    commandRunnerWithOutput = func(_ io.Writer, name string, args ...string) error {
        if name == "brew" && len(args) > 0 && args[0] == "install" && len(args) == 2 && args[1] == "node" {
            commandInstalled = true
            return nil
        }
        return nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunnerWithOutput = oldRunnerWithOutput
    }()

    cmd := NewInstallCmd()
    cmd.SetArgs([]string{"--force", "node"})
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("install command failed with --force: %v", err)
    }

    if !commandInstalled {
        t.Fatal("expected brew install node command to be executed")
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
    oldRunner := commandRunner
    oldRunnerWithOutput := commandRunnerWithOutput
    executableLookup = func(name string) (string, error) {
        if name == "brew" {
            return "/opt/homebrew/bin/brew", nil
        }
        if name == "node" || name == "npm" {
            return "", exec.ErrNotFound
        }
        return "", exec.ErrNotFound
    }
    commandRunner = func(name string, args ...string) ([]byte, error) {
        if name == "brew" && len(args) > 0 && args[0] == "install" && len(args) == 2 && args[1] == "node" {
            return []byte("Homebrew output: success\n"), nil
        }
        return nil, nil
    }
    commandRunnerWithOutput = func(out io.Writer, name string, args ...string) error {
        if name == "brew" && len(args) > 0 && args[0] == "install" && len(args) == 2 && args[1] == "node" {
            _, _ = out.Write([]byte("Homebrew output: success\n"))
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
    cmd.SetArgs([]string{"node"})
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
    }()

    cmd := NewInstallCmd()
    cmd.SetArgs([]string{"curl"})
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
