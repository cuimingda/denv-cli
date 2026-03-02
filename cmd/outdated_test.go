package cmd

import (
    "bytes"
    "os/exec"
    "strings"
    "testing"
)

func TestCmpVersions(t *testing.T) {
    cases := []struct {
        current string
        latest  string
		want    int
	}{
		{"1.23.4", "1.23.5", -1},
		{"1.23.4", "1.23.4", 0},
		{"1.10.0", "1.2.4", 1},
		{"8.0_1", "8.0.1_4", -1},
	}

    for _, c := range cases {
        got := cmpVersions(c.current, c.latest)
        if got != c.want {
            t.Fatalf("cmpVersions(%q, %q)=%d, expected %d", c.current, c.latest, got, c.want)
        }
    }
}

func TestParseBrewStableVersionUsesRevision(t *testing.T) {
    payload := `{"formulae":[{"name":"ffmpeg","revision":4,"versions":{"stable":"8.0.1"}}]}`

    got, err := parseBrewStableVersion([]byte(payload))
    if err != nil {
        t.Fatalf("parseBrewStableVersion failed: %v", err)
    }
    if got != "8.0.1_4" {
        t.Fatalf("expected 8.0.1_4, got %q", got)
    }
}

func TestOutdatedShowsOutdatedTool(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(name string) (string, error) {
        if name == "tree" {
            return "/usr/local/bin/tree", nil
        }
        return "", exec.ErrNotFound
    }
    commandRunner = func(name string, args ...string) ([]byte, error) {
        if name == "tree" && len(args) == 1 && args[0] == "--version" {
            return []byte("tree version 2.1.3"), nil
        }
        if name == "brew" && len(args) >= 3 && args[0] == "info" {
            payload := `{"formulae":[{"name":"tree","versions":{"stable":"2.3.1"}}]}`
            return []byte(payload), nil
        }
        return []byte(""), nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    cmd := NewOutdatedCmd()
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("outdated command failed: %v", err)
    }

    got := strings.TrimSpace(out.String())
    lines := strings.Split(got, "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "tree ") {
            if line != "tree 2.1.3 < 2.3.1" {
                t.Fatalf("unexpected outdated line, got %q", line)
            }
            return
        }
    }

    t.Fatal("expected tree outdated output")
}

func TestOutdatedShowsUpToDateTool(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(name string) (string, error) {
        if name == "tree" {
            return "/usr/local/bin/tree", nil
        }
        return "", exec.ErrNotFound
    }
    commandRunner = func(name string, args ...string) ([]byte, error) {
        if name == "tree" && len(args) == 1 && args[0] == "--version" {
            return []byte("tree version 2.3.1"), nil
        }
        if name == "brew" && len(args) >= 3 && args[0] == "info" {
            payload := `{"formulae":[{"name":"tree","versions":{"stable":"2.3.1"}}]}`
            return []byte(payload), nil
        }
        return []byte(""), nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    cmd := NewOutdatedCmd()
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("outdated command failed: %v", err)
    }

    got := strings.TrimSpace(out.String())
    lines := strings.Split(got, "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "tree ") {
            if line != "tree 2.3.1" {
                t.Fatalf("unexpected up-to-date line, got %q", line)
            }
            return
        }
    }

    t.Fatal("expected tree up-to-date output")
}

func TestOutdatedHandlesMissingTool(t *testing.T) {
    oldLookup := executableLookup
    oldRunner := commandRunner
    executableLookup = func(string) (string, error) {
        return "", exec.ErrNotFound
    }
    commandRunner = func(name string, args ...string) ([]byte, error) {
        return []byte(""), nil
    }
    defer func() {
        executableLookup = oldLookup
        commandRunner = oldRunner
    }()

    cmd := NewOutdatedCmd()
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("outdated command failed: %v", err)
    }

    got := strings.TrimSpace(out.String())
    lines := strings.Split(got, "\n")
    if len(lines) != len(SupportedTools()) {
        t.Fatalf("expected %d lines, got %d", len(SupportedTools()), len(lines))
    }

    for _, name := range SupportedTools() {
        expected := name + " not found"
        for _, line := range lines {
            if strings.TrimSpace(line) == expected {
                goto found
            }
        }
        t.Fatalf("missing expected line %q", expected)
    found:
    }
}
