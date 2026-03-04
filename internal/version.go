// internal/version.go 实现工具版本能力：命令版本提取、brew/npm 最新版本来源、版本文本解析与抽取接口。
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

// VersionParser 定义版本解析能力。
type VersionParser interface {
	Parse(raw string) (string, error)
}

// RegexVersionParser 使用正则提取版本文本。
type RegexVersionParser struct{}

func (RegexVersionParser) Parse(raw string) (string, error) {
	for _, pattern := range versionPatterns {
		if match := pattern.FindString(raw); match != "" {
			return match, nil
		}
	}
	return "", fmt.Errorf("version not found")
}

// VersionSource 定义版本来源。
type VersionSource interface {
	CurrentVersion(rt Runtime, catalog *toolCatalog, name string, commandPath string) (string, error)
	LatestVersion(rt Runtime, catalog *toolCatalog, name string) (string, error)
}

// commandVersionSource 从命令行执行输出提取版本。
type commandVersionSource struct{}

func (commandVersionSource) CurrentVersion(rt Runtime, catalog *toolCatalog, name string, commandPath string) (string, error) {
	return toolVersionFromCommandSource(rt, catalog, name, commandPath)
}

func (commandVersionSource) LatestVersion(rt Runtime, catalog *toolCatalog, name string) (string, error) {
	return toolVersionFromCommandSource(rt, catalog, name, "")
}

// brewOutdatedVersionSource 使用 brew 进行过期检测场景的版本读取。
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

// npmLatestVersionSource 使用 npm 接口读取 npm 最新版本。
type npmLatestVersionSource struct{}

func (npmLatestVersionSource) CurrentVersion(rt Runtime, catalog *toolCatalog, name string, commandPath string) (string, error) {
	return commandVersionSource{}.CurrentVersion(rt, catalog, name, commandPath)
}

func (npmLatestVersionSource) LatestVersion(rt Runtime, catalog *toolCatalog, name string) (string, error) {
	return toolLatestVersionByNpm(rt)
}

// ToolVersion 读取指定工具的当前版本。
func ToolVersion(rt Runtime, name string) (string, error) {
	return ToolVersionWithCatalog(rt, NewToolCatalog(), name)
}

// ToolVersionWithCatalog 使用显式 catalog 读取版本。
func ToolVersionWithCatalog(rt Runtime, catalog *toolCatalog, name string) (string, error) {
	return extractVersionFromSource(rt, commandVersionSource{}, catalog, name, "")
}

// ToolVersionWithPath 指定路径优先的版本读取入口。
func ToolVersionWithPath(rt Runtime, name, commandPath string) (string, error) {
	return ToolVersionWithPathWithCatalog(rt, NewToolCatalog(), name, commandPath)
}

// ToolVersionWithPathWithCatalog 失败回退时使用显式 commandPath 兜底。
func ToolVersionWithPathWithCatalog(rt Runtime, catalog *toolCatalog, name, commandPath string) (string, error) {
	// 先尝试标准命令路径；若失败再尝试明确路径
	version, err := ToolVersionWithCatalog(rt, catalog, name)
	if err == nil {
		return version, nil
	}

	if commandPath == "" || commandPath == name {
		return "", err
	}

	return ToolVersionFromPath(rt, commandPath, name)
}

// ToolVersionForOutdated 返回兼容 outdated 场景的版本读取策略。
func ToolVersionForOutdated(rt Runtime, name string) (string, error) {
	return ToolVersionForOutdatedWithCatalog(rt, NewToolCatalog(), name)
}

// ToolVersionForOutdatedWithCatalog 针对 npm 与 brew 区分来源。
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

// ToolVersionFromPath 从显式命令路径提取版本。
func ToolVersionFromPath(rt Runtime, commandPath, name string) (string, error) {
	return toolVersionFromCommandSource(rt, NewToolCatalog(), name, commandPath)
}

// toolVersionFromBrewList 优先从 `brew info` 文本读取版本，再回退 JSON。
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

// extractBrewCurrentInstallVersion 从文本或 JSON 中提取已安装版本。
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

// extractBrewCurrentInstallVersionFromJSON 专注解析 JSON 安装元数据中的版本字段。
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

// ToolLatestVersion 获取工具最新版本。
func ToolLatestVersion(rt Runtime, name string) (string, error) {
	return ToolLatestVersionWithCatalog(rt, NewToolCatalog(), name)
}

// ToolLatestVersionWithCatalog 按工具分流到不同来源。
func ToolLatestVersionWithCatalog(rt Runtime, catalog *toolCatalog, name string) (string, error) {
	if name == "npm" {
		return latestVersionFromSource(rt, npmLatestVersionSource{}, catalog, name)
	}
	return latestVersionFromSource(rt, brewOutdatedVersionSource{}, catalog, name)
}

// latestVersionFromSource 统一校验来源并读取最新版本。
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

// toolLatestVersionByBrew 使用 brew 的 JSON 接口读取版本。
func toolLatestVersionByBrew(rt Runtime, formula string) (string, error) {
	rt = NormalizeRuntime(rt)
	output, err := rt.CommandRunner("brew", "info", "--json=v2", formula)
	if err != nil {
		return "", fmt.Errorf("brew info failed: %w", err)
	}
	return ParseBrewStableVersion(output)
}

// toolLatestVersionByNpm 使用 npm 官方命令查询 npm 包版本。
func toolLatestVersionByNpm(rt Runtime) (string, error) {
	rt = NormalizeRuntime(rt)
	output, err := rt.CommandRunner("npm", "view", "npm", "version")
	if err != nil {
		return "", fmt.Errorf("npm latest version failed: %w", err)
	}

	return ExtractVersion(string(output))
}

// ParseBrewStableVersion 解析 brew JSON 的稳定版本。
func ParseBrewStableVersion(output []byte) (string, error) {
	return parseBrewStableVersionPayload(output)
}

// parseBrewStableVersionPayload 支持两种 brew json 结构。
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

// toolVersionFromCommandSource 执行工具版本命令并提取版本字段。
func toolVersionFromCommandSource(rt Runtime, catalog *toolCatalog, name string, commandPath string) (string, error) {
	rt = NormalizeRuntime(rt)
	versionArgs, ok := catalog.versionArgsForTool(name)
	if !ok {
		return "", fmt.Errorf("unsupported tool: %s", name)
	}

	// 若 commandPath 与工具名不同，说明是显式路径，优先走此路径执行
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

// extractVersionFromSource 聚合获取当前版本逻辑。
func extractVersionFromSource(rt Runtime, source VersionSource, catalog *toolCatalog, name string, commandPath string) (string, error) {
	return source.CurrentVersion(rt, catalog, name, commandPath)
}

// ExtractVersion 对外部可复用的版本提取入口。
func ExtractVersion(out string) (string, error) {
	return RegexVersionParser{}.Parse(out)
}
