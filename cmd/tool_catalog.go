package cmd

import "strings"

type managedTool struct {
	name                   string
	displayName            string
	formula                string
	versionArgs            []string
	installable            bool
	planCheckCommand       bool
	planCheckFormula       bool
	runCheckCommand        bool
	runCheckFormula        bool
	alreadyInstalledLabel  string
	installExtraOperations [][]string
}

var (
	toolsCatalog     []managedTool
	toolLookup       map[string]managedTool
	listedTools      []string
	installableTools  []string
	installableOrder  = []string{
		"php",
		"python3",
		"node",
		"go",
		"curl",
		"git",
		"ffmpeg",
		"tree",
		"gh",
	}
)

func init() {
	toolsCatalog = []managedTool{
		{
			name:                  "php",
			displayName:           "php",
			formula:               "php",
			versionArgs:           []string{"--version"},
			installable:           true,
			planCheckCommand:      true,
			runCheckCommand:       true,
			alreadyInstalledLabel:  "php is already installed",
			installExtraOperations: nil,
		},
		{
			name:                  "python3",
			displayName:           "python3",
			formula:               "python3",
			versionArgs:           []string{"--version"},
			installable:           true,
			planCheckCommand:      true,
			planCheckFormula:      true,
			runCheckCommand:       false,
			runCheckFormula:       true,
			alreadyInstalledLabel:  "python3 is already installed by homebrew",
			installExtraOperations: nil,
		},
		{
			name:                  "node",
			displayName:           "node",
			formula:               "node",
			versionArgs:           []string{"--version"},
			installable:           true,
			planCheckCommand:      true,
			runCheckCommand:       true,
			alreadyInstalledLabel:  "node is already installed",
			installExtraOperations: nil,
		},
		{
			name:                  "go",
			displayName:           "go",
			formula:               "go",
			versionArgs:           []string{"version"},
			installable:           true,
			planCheckCommand:      true,
			runCheckCommand:       true,
			alreadyInstalledLabel:  "go is already installed",
			installExtraOperations: nil,
		},
		{
			name:                  "npm",
			displayName:           "npm",
			formula:               "node",
			versionArgs:           []string{"--version"},
			installable:           false,
			alreadyInstalledLabel:  "",
			installExtraOperations: nil,
		},
		{
			name:                  "curl",
			displayName:           "curl",
			formula:               "curl",
			versionArgs:           []string{"--version"},
			installable:            true,
			planCheckCommand:      true,
			runCheckCommand:       true,
			alreadyInstalledLabel:  "curl is already installed",
			installExtraOperations: [][]string{{"brew", "link", "curl", "--force"}},
		},
		{
			name:                  "gh",
			displayName:           "gh",
			formula:               "gh",
			versionArgs:           []string{"--version"},
			installable:           true,
			planCheckCommand:      true,
			runCheckCommand:       true,
			alreadyInstalledLabel:  "gh is already installed",
			installExtraOperations: nil,
		},
		{
			name:                  "git",
			displayName:           "git",
			formula:               "git",
			versionArgs:           []string{"--version"},
			installable:           true,
			planCheckCommand:      true,
			runCheckCommand:       true,
			alreadyInstalledLabel:  "git is already installed",
			installExtraOperations: nil,
		},
		{
			name:                  "ffmpeg",
			displayName:           "ffmpeg",
			formula:               "ffmpeg",
			versionArgs:           []string{"-version"},
			installable:           true,
			planCheckCommand:      true,
			runCheckCommand:       true,
			alreadyInstalledLabel:  "ffmpeg is already installed",
			installExtraOperations: nil,
		},
		{
			name:                  "tree",
			displayName:           "tree",
			formula:               "tree",
			versionArgs:           []string{"--version"},
			installable:           true,
			planCheckCommand:      true,
			runCheckCommand:       true,
			alreadyInstalledLabel:  "tree is already installed",
			installExtraOperations: nil,
		},
	}

	toolLookup = make(map[string]managedTool, len(toolsCatalog))
	listedTools = make([]string, 0, len(toolsCatalog))
	installableTools = make([]string, 0, len(toolsCatalog))

	for _, item := range toolsCatalog {
		toolLookup[item.name] = item
		listedTools = append(listedTools, item.name)
	}

	installed := make(map[string]bool, len(installableOrder))
	for _, name := range installableOrder {
		item, ok := toolLookup[name]
		if !ok || !item.installable {
			continue
		}
		installed[name] = true
		installableTools = append(installableTools, name)
	}

	for _, item := range toolsCatalog {
		if !item.installable {
			continue
		}
		if installed[item.name] {
			continue
		}
		installableTools = append(installableTools, item.name)
	}
}

func managedToolFor(name string) (managedTool, bool) {
	item, ok := toolLookup[name]
	return item, ok
}

func installOperationSequence(name string) ([]string, bool) {
	tool, ok := managedToolFor(name)
	if !ok {
		return nil, false
	}

	ops := make([]string, 0, 1+len(tool.installExtraOperations))
	ops = append(ops, "brew install "+tool.formula)
	for _, op := range tool.installExtraOperations {
		ops = append(ops, strings.Join(op, " "))
	}
	return ops, true
}

func brewFormulaForTool(name string) (string, bool) {
	tool, ok := managedToolFor(name)
	if !ok {
		return "", false
	}
	return tool.formula, true
}

func versionArgsForTool(name string) ([]string, bool) {
	tool, ok := managedToolFor(name)
	if !ok || len(tool.versionArgs) == 0 {
		return nil, false
	}
	return tool.versionArgs, true
}

func toolDisplayName(name string) (string, bool) {
	tool, ok := managedToolFor(name)
	if !ok {
		return "", false
	}
	return tool.displayName, true
}

func managedToolIsInstallable(name string) bool {
	tool, ok := managedToolFor(name)
	return ok && tool.installable
}
