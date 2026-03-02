package cmd

import (
    "bytes"
    "os/exec"
    "strings"
    "testing"
)

func TestNewListCmdShowsVersionsAndMissingTools(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(name string) (string, error) {
        if name == "php" || name == "go" {
            return "/usr/bin/" + name, nil
        }
        return "", exec.ErrNotFound
    }
    commandRunner = func(name string, args ...string) ([]byte, error) {
        switch name {
        case "php":
            return []byte("PHP 8.3.4 (cli) (built: Jan  1 2025 00:00:00)"), nil
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
    want := "php 8.3.4\npython not found\nnode not found\nGo 1.23.4"
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

func TestIsSupportedTool(t *testing.T) {
    if !IsSupportedTool("php") {
        t.Fatal("expected php to be supported")
    }

    if IsSupportedTool("ruby") {
        t.Fatal("expected ruby to be unsupported")
    }
}

func TestRootHasListCommand(t *testing.T) {
    cmd := NewRootCmd()
    found := false
    for _, sub := range cmd.Commands() {
        if sub.Name() == "list" {
            found = true
            break
        }
    }
    if !found {
        t.Fatal("root command should include list subcommand")
    }
}
