package main

import (
	"io"
	"io/fs"
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
		t.Fatal("did not expect binary `denv` to be generated")
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
	}
	for _, relPath := range fileList {
		srcPath := filepath.Join(srcRoot, relPath)
		dstPath := filepath.Join(dstRoot, relPath)
		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	if err := copyDir(srcRoot, dstRoot, "cmd"); err != nil {
		return err
	}
	if err := copyDir(srcRoot, dstRoot, "internal"); err != nil {
		return err
	}

	return nil
}

func copyDir(srcRoot, dstRoot, relDir string) error {
	return filepath.WalkDir(filepath.Join(srcRoot, relDir), func(srcPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, relErr := filepath.Rel(srcRoot, srcPath)
		if relErr != nil {
			return relErr
		}

		dstPath := filepath.Join(dstRoot, rel)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		return copyFile(srcPath, dstPath)
	})
}

func copyFile(srcPath, dstPath string) error {
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
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}
