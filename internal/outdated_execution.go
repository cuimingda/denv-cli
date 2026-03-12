// internal/outdated_execution.go 提供 outdated 单工具执行与 brew update 流式输出能力。
package denv

import (
	"fmt"
	"io"
)

// outdatedCheckWithOutput 计算单个工具的 outdated 状态，并将关键过程日志写入输出。
func outdatedCheckWithOutput(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy, out io.Writer, name string) (OutdatedCheck, error) {
	catalog = ensureToolCatalog(catalog)
	if pathPolicy == nil {
		pathPolicy = DefaultPathPolicy()
	}
	if out == nil {
		out = io.Discard
	}

	lifecycle, ok := catalog.toolLifecycle(name)
	if !ok {
		return OutdatedCheck{}, fmt.Errorf("unsupported tool: %s", name)
	}

	row := newOutdatedCheck(name, lifecycle.DisplayName(name), "", "", OutdatedStateNotInstalled)

	outdatedLogf(out, "resolve install state")
	installed, _, _, err := lifecycle.IsInstalled(rt, catalog, pathPolicy)
	if err != nil {
		outdatedLogf(out, "install state error: %v", err)
		return OutdatedCheck{}, err
	}
	row.Installed = installed
	outdatedLogf(out, "installed=%t", installed)

	if !installed {
		row.Current = "<not installed>"
		outdatedLogf(out, "resolve latest version")
		latest, latestErr := lifecycle.ResolveLatestVersion(rt, catalog)
		if latestErr != nil {
			row.State = OutdatedStateInvalidLatest
			row.CheckError = latestErr
			outdatedLogf(out, "latest error: %v", latestErr)
		} else {
			row.State = OutdatedStateNotInstalled
			row.Latest = latest
			outdatedLogf(out, "latest=%s", latest)
		}
		return row, nil
	}

	outdatedLogf(out, "resolve current version")
	current, currentErr := lifecycle.ResolveOutdatedCurrentVersion(rt, catalog)
	if currentErr != nil {
		row.State = OutdatedStateInvalidCurrent
		row.CheckError = currentErr
		outdatedLogf(out, "current error: %v", currentErr)
		return row, nil
	}
	outdatedLogf(out, "current=%s", current)

	outdatedLogf(out, "resolve latest version")
	latest, latestErr := lifecycle.ResolveLatestVersion(rt, catalog)
	if latestErr != nil {
		row.State = OutdatedStateInvalidLatest
		row.CheckError = latestErr
		outdatedLogf(out, "latest error: %v", latestErr)
		return row, nil
	}
	outdatedLogf(out, "latest=%s", latest)

	row.Current = current
	row.Latest = latest
	if CompareVersions(current, latest) < 0 {
		row.State = OutdatedStateOutdated
	} else {
		row.State = OutdatedStateUpToDate
	}
	outdatedLogf(out, "state=%s", row.State)

	return row, nil
}

// runBrewUpdate 先校验 brew 可用性，再将 brew update 的输出透传给调用方。
func runBrewUpdate(rt Runtime, out io.Writer) error {
	rt = NormalizeRuntime(rt)
	if out == nil {
		out = io.Discard
	}
	if err := rt.CommandRunnerWithOutput(out, "brew", "update"); err != nil {
		if !IsBrewInstalled(rt) {
			return fmt.Errorf("homebrew is not installed")
		}
		return fmt.Errorf("brew update failed: %w", err)
	}
	return nil
}

func outdatedLogf(out io.Writer, format string, args ...any) {
	if out == nil {
		return
	}
	_, _ = fmt.Fprintf(out, format+"\n", args...)
}
