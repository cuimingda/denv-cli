package main

import (
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "runtime"
    "testing"
)

func TestGoBuildTargets(t *testing.T) {
    _, thisFile, _, ok := runtime.Caller(0)
    if !ok {
        t.Fatal("failed to resolve test file path")
    }

    projectRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
    tempDir := t.TempDir()

    if err := copyBuildFixture(projectRoot, tempDir); err != nil {
        t.Fatalf("copy fixture failed: %v", err)
    }

    cmdBuildDot := exec.Command("go", "build", ".")
    cmdBuildDot.Dir = tempDir
    if out, err := cmdBuildDot.CombinedOutput(); err == nil {
        t.Fatalf("expected `go build .` to fail in project root, got nil error; output: %s", out)
    }

    if _, err := os.Stat(filepath.Join(tempDir, "denv-cli")); err == nil {
        t.Fatal("did not expect binary `denv-cli` to be generated")
    }
    if _, err := os.Stat(filepath.Join(tempDir, "denv")); err == nil {
        t.Fatal("did not expect binary `denv` after `go build .`")
    }

    cmdBuildCmd := exec.Command("go", "build", "./cmd/denv")
    cmdBuildCmd.Dir = tempDir
    if out, err := cmdBuildCmd.CombinedOutput(); err != nil {
        t.Fatalf("expected `go build ./cmd/denv` success, got error: %v; output: %s", err, out)
    }

    if _, err := os.Stat(filepath.Join(tempDir, "denv")); err != nil {
        t.Fatalf("expected `denv` binary to be generated, but file missing: %v", err)
    }

    if _, err := os.Stat(filepath.Join(tempDir, "cmd", "denv", "denv")); err == nil {
        t.Fatal("`go build ./cmd/denv` should generate root output `denv`, not nested `cmd/denv/denv`")
    }
}

func copyBuildFixture(srcRoot, dstRoot string) error {
    fileList := []string{
        "go.mod",
        "go.sum",
        filepath.Join("cmd", "denv", "main.go"),
        filepath.Join("cmd", "root.go"),
        filepath.Join("cmd", "list.go"),
        filepath.Join("cmd", "tools.go"),
        filepath.Join("cmd", "install.go"),
    }

    for _, relPath := range fileList {
        srcPath := filepath.Join(srcRoot, relPath)
        dstPath := filepath.Join(dstRoot, relPath)

        if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
            return err
        }

        srcFile, err := os.Open(srcPath)
        if err != nil {
            return err
        }
        defer srcFile.Close()

        dstFile, err := os.Create(dstPath)
        if err != nil {
            return err
        }

        if _, err := io.Copy(dstFile, srcFile); err != nil {
            return err
        }

        if err := dstFile.Close(); err != nil {
            return err
        }
    }

    return nil
}
