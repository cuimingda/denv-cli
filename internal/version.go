package denv

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var versionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\d+\.\d+\.\d+`),
	regexp.MustCompile(`\d+\.\d+`),
}

type VersionParser interface {
	Parse(raw string) (string, error)
}

type RegexVersionParser struct{}

func (RegexVersionParser) Parse(raw string) (string, error) {
	for _, pattern := range versionPatterns {
		if match := pattern.FindString(raw); match != "" {
			return match, nil
		}
	}
	return "", fmt.Errorf("version not found")
}

type VersionSource interface {
	CurrentVersion(rt Runtime, catalog *toolCatalog, name string, commandPath string) (string, error)
	LatestVersion(rt Runtime, catalog *toolCatalog, name string) (string, error)
}

type commandVersionSource struct{}

func (commandVersionSource) CurrentVersion(rt Runtime, catalog *toolCatalog, name string, commandPath string) (string, error) {
	return toolVersionFromCommandSource(rt, catalog, name, commandPath)
}

func (commandVersionSource) LatestVersion(rt Runtime, catalog *toolCatalog, name string) (string, error) {
	return toolVersionFromCommandSource(rt, catalog, name, "")
}

type brewOutdatedVersionSource struct{}

func (brewOutdatedVersionSource) CurrentVersion(rt Runtime, catalog *toolCatalog, name string, _ string) (string, error) {
	formula, ok := catalog.brewFormulaForTool(name)
	if !ok {
		return "", fmt.Errorf("unsupported tool: %s", name)
	}
	return toolVersionFromBrewList(rt, formula)
}

func (brewOutdatedVersionSource) LatestVersion(rt Runtime, catalog *toolCatalog, name string) (string, error) {
	formula, ok := catalog.brewFormulaForTool(name)
	if !ok {
		return "", fmt.Errorf("unsupported tool: %s", name)
	}
	return toolLatestVersionByBrew(rt, formula)
}

type npmLatestVersionSource struct{}

func (npmLatestVersionSource) CurrentVersion(rt Runtime, catalog *toolCatalog, name string, commandPath string) (string, error) {
	return commandVersionSource{}.CurrentVersion(rt, catalog, name, commandPath)
}

func (npmLatestVersionSource) LatestVersion(rt Runtime, catalog *toolCatalog, name string) (string, error) {
	return toolLatestVersionByNpm(rt)
}

func ToolVersion(rt Runtime, name string) (string, error) {
	return ToolVersionWithCatalog(rt, NewToolCatalog(), name)
}

func ToolVersionWithCatalog(rt Runtime, catalog *toolCatalog, name string) (string, error) {
	return extractVersionFromSource(rt, commandVersionSource{}, catalog, name, "")
}

func ToolVersionWithPath(rt Runtime, name, commandPath string) (string, error) {
	return ToolVersionWithPathWithCatalog(rt, NewToolCatalog(), name, commandPath)
}

func ToolVersionWithPathWithCatalog(rt Runtime, catalog *toolCatalog, name, commandPath string) (string, error) {
	version, err := ToolVersionWithCatalog(rt, catalog, name)
	if err == nil {
		return version, nil
	}

	if commandPath == "" || commandPath == name {
		return "", err
	}

	return ToolVersionFromPath(rt, commandPath, name)
}

func ToolVersionForOutdated(rt Runtime, name string) (string, error) {
	return ToolVersionForOutdatedWithCatalog(rt, NewToolCatalog(), name)
}

func ToolVersionForOutdatedWithCatalog(rt Runtime, catalog *toolCatalog, name string) (string, error) {
	if name == "npm" {
		return ToolVersion(rt, name)
	}

	_, ok := catalog.brewFormulaForTool(name)
	if !ok {
		return ToolVersion(rt, name)
	}

	return brewOutdatedVersionSource{}.CurrentVersion(rt, catalog, name, "")
}

func ToolVersionFromPath(rt Runtime, commandPath, name string) (string, error) {
	return toolVersionFromCommandSource(rt, NewToolCatalog(), name, commandPath)
}

func toolVersionFromBrewList(rt Runtime, formula string) (string, error) {
	rt = NormalizeRuntime(rt)
	output, err := rt.CommandRunner("brew", "info", formula)
	if err == nil {
		if version := extractBrewCurrentInstallVersion(string(output)); version != "" {
			return version, nil
		}
	}

	output, err = rt.CommandRunner("brew", "info", "--json=v2", formula)
	if err != nil {
		if formulaVersion, versionErr := ExtractVersion(string(output)); versionErr == nil {
			return formulaVersion, nil
		}
		return "", fmt.Errorf("brew info failed: %w", err)
	}

	if version := extractBrewCurrentInstallVersion(string(output)); version != "" {
		return version, nil
	}

	if version, parseErr := parseBrewStableVersionPayload(output); parseErr == nil {
		return version, nil
	}

	return "", fmt.Errorf("failed to parse brew version")
}

func extractBrewCurrentInstallVersion(output string) string {
	if version := extractBrewCurrentInstallVersionFromJSON([]byte(output)); version != "" {
		return version
	}

	var fallback string
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, "/opt/homebrew/Cellar/") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) == 0 {
			continue
		}
		if !strings.HasPrefix(fields[0], "/opt/homebrew/Cellar/") {
			continue
		}

		pathParts := strings.Split(fields[0], "/")
		if len(pathParts) == 0 {
			continue
		}

		candidate := pathParts[len(pathParts)-1]
		if candidate == "" {
			continue
		}
		fallback = candidate

		if len(fields) >= 2 && fields[len(fields)-1] == "*" {
			return candidate
		}
	}
	return fallback
}

func extractBrewCurrentInstallVersionFromJSON(output []byte) string {
	var payload struct {
		Formulae []struct {
			Installed []struct {
				Version string `json:"version"`
			} `json:"installed"`
		} `json:"formulae"`
	}

	if err := json.Unmarshal(output, &payload); err != nil {
		return ""
	}

	if len(payload.Formulae) == 0 || len(payload.Formulae[0].Installed) == 0 {
		return ""
	}

	for i := range payload.Formulae[0].Installed {
		v := payload.Formulae[0].Installed[i].Version
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func ToolLatestVersion(rt Runtime, name string) (string, error) {
	return ToolLatestVersionWithCatalog(rt, NewToolCatalog(), name)
}

func ToolLatestVersionWithCatalog(rt Runtime, catalog *toolCatalog, name string) (string, error) {
	if name == "npm" {
		return latestVersionFromSource(rt, npmLatestVersionSource{}, catalog, name)
	}
	return latestVersionFromSource(rt, brewOutdatedVersionSource{}, catalog, name)
}

func latestVersionFromSource(rt Runtime, source VersionSource, catalog *toolCatalog, name string) (string, error) {
	if source == nil {
		return "", fmt.Errorf("version source is nil")
	}

	_, ok := catalog.brewFormulaForTool(name)
	if !ok && name == "npm" {
		return source.LatestVersion(rt, catalog, name)
	}

	if !ok {
		return "", fmt.Errorf("unsupported tool: %s", name)
	}

	return source.LatestVersion(rt, catalog, name)
}

func toolLatestVersionByBrew(rt Runtime, formula string) (string, error) {
	rt = NormalizeRuntime(rt)
	output, err := rt.CommandRunner("brew", "info", "--json=v2", formula)
	if err != nil {
		return "", fmt.Errorf("brew info failed: %w", err)
	}
	return ParseBrewStableVersion(output)
}

func toolLatestVersionByNpm(rt Runtime) (string, error) {
	rt = NormalizeRuntime(rt)
	output, err := rt.CommandRunner("npm", "view", "npm", "version")
	if err != nil {
		return "", fmt.Errorf("npm latest version failed: %w", err)
	}

	return ExtractVersion(string(output))
}

func ParseBrewStableVersion(output []byte) (string, error) {
	return parseBrewStableVersionPayload(output)
}

func parseBrewStableVersionPayload(output []byte) (string, error) {
	var payload struct {
		Formulae []struct {
			Versions struct {
				Stable string `json:"stable"`
			} `json:"versions"`
			Revision int `json:"revision"`
		} `json:"formulae"`
	}

	if err := json.Unmarshal(output, &payload); err == nil {
		if len(payload.Formulae) > 0 {
			formula := payload.Formulae[0]
			if formula.Revision > 0 && formula.Versions.Stable != "" {
				return fmt.Sprintf("%s_%d", formula.Versions.Stable, formula.Revision), nil
			}
			if formula.Versions.Stable != "" {
				return formula.Versions.Stable, nil
			}
		}
	}

	var list []struct {
		Versions struct {
			Stable string `json:"stable"`
		} `json:"versions"`
	}

	if err := json.Unmarshal(output, &list); err == nil {
		if len(list) > 0 && list[0].Versions.Stable != "" {
			return list[0].Versions.Stable, nil
		}
	}

	return "", fmt.Errorf("failed to parse brew version")
}

func toolVersionFromCommandSource(rt Runtime, catalog *toolCatalog, name string, commandPath string) (string, error) {
	rt = NormalizeRuntime(rt)
	versionArgs, ok := catalog.versionArgsForTool(name)
	if !ok {
		return "", fmt.Errorf("unsupported tool: %s", name)
	}

	var output []byte
	var err error
	if commandPath != "" && commandPath != name {
		output, err = rt.CommandRunner(commandPath, versionArgs...)
	} else {
		output, err = rt.CommandRunner(name, versionArgs...)
	}
	if err != nil {
		return "", fmt.Errorf("get version failed: %w", err)
	}

	return ExtractVersion(string(output))
}

func extractVersionFromSource(rt Runtime, source VersionSource, catalog *toolCatalog, name string, commandPath string) (string, error) {
	return source.CurrentVersion(rt, catalog, name, commandPath)
}

func ExtractVersion(out string) (string, error) {
	return RegexVersionParser{}.Parse(out)
}
