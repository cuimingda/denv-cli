package cmd

import (
    "bytes"
    "fmt"
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
    want := "php\npython3\nnode\ngo\nnpm"
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
    want := "php 8.3.4 (/usr/bin/php)\npython3 3.12.4 (/usr/bin/python3)\nnode not found\ngo 1.23.4 (/usr/bin/go)\nnpm not found"
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
    want := "php 8.3.4\npython3 3.12.4\nnode not found\ngo 1.23.4\nnpm not found"
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
    want := "php /usr/bin/php\npython3 /usr/bin/python3\nnode not found\ngo /usr/bin/go\nnpm not found"
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

func TestIsInstallableTool(t *testing.T) {
    if !IsInstallableTool("php") {
        t.Fatal("expected php to be installable")
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
    for _, sub := range cmd.Commands() {
        if sub.Name() == "list" {
            found = true
        }
        if sub.Name() == "install" {
            installFound = true
        }
    }
    if !found {
        t.Fatal("root command should include list subcommand")
    }
    if !installFound {
        t.Fatal("root command should include install subcommand")
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
