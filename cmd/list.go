package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

const (
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

type listOutputMode string

const (
	listOutputPlain   listOutputMode = "plain"
	listOutputJSON    listOutputMode = "json"
	listOutputTable   listOutputMode = "table"
	listOutputNoColor listOutputMode = "no-color"
)

func NewListCmd() *cobra.Command {
	return NewListCmdWithService(NewCLIContext().Service)
}

func NewListCmdWithService(svc ToolService) *cobra.Command {
	if svc == nil {
		svc = NewCLIContext().Service
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List supported developer tools",
		RunE: func(cmd *cobra.Command, _ []string) error {
			showVersion, _ := cmd.Flags().GetBool("version")
			showPath, _ := cmd.Flags().GetBool("path")
			outputMode, _ := cmd.Flags().GetString("output")

			mode, err := parseListOutput(outputMode)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			colorOutput := useColorOutput(out) && mode != listOutputNoColor

			verbosef(cmd, "list started with version=%t path=%t output=%s", showVersion, showPath, mode)
			start := time.Now()
			items := make([]listItem, 0, len(svc.SupportedTools()))
			for _, name := range svc.SupportedTools() {
				verbosef(cmd, "resolving tool: %s", name)
				item := listItem{
					Name:          name,
					DisplayName:   svc.ToolDisplayName(name),
					ManagedByBrew: false,
					Installed:     false,
					Version:       "",
					Path:          "",
				}

				installed, path, _, stateErr := svc.ToolInstallState(name)
				if stateErr != nil {
					return stateErr
				}
				if installed {
					item.Installed = true
					item.Path = path
					item.ManagedByBrew = svc.IsManagedByHomebrew(path)
				}

				if showVersion && installed {
					if toolVersion, err := svc.ToolVersionWithPath(name, path); err == nil {
						item.Version = toolVersion
					} else {
						item.Installed = false
						item.ManagedByBrew = false
						item.Path = ""
					}
				}

				items = append(items, item)
			}
			verbosef(cmd, "list scan completed in %s", time.Since(start))

			return renderList(cmd.OutOrStdout(), mode, listRenderOptions{
				colorOutput: colorOutput,
				showVersion: showVersion,
				showPath:    showPath,
			}, items)
		},
	}

	cmd.Flags().Bool("version", false, "show versions for discovered tools")
	cmd.Flags().Bool("path", false, "show executable paths for discovered tools")
	cmd.Flags().String("output", string(listOutputPlain), "output format: plain|json|table|no-color")
	return cmd
}

type listRenderOptions struct {
	colorOutput bool
	showVersion bool
	showPath    bool
}

type listItem struct {
	Name          string
	DisplayName   string
	Installed     bool
	Version       string
	Path          string
	ManagedByBrew bool
}

func parseListOutput(raw string) (listOutputMode, error) {
	mode := listOutputMode(raw)
	switch mode {
	case listOutputPlain, listOutputJSON, listOutputTable, listOutputNoColor:
		return mode, nil
	}
	return "", fmt.Errorf("invalid output: %s", raw)
}

func renderList(out io.Writer, mode listOutputMode, opts listRenderOptions, items []listItem) error {
	switch mode {
	case listOutputJSON:
		payload := make([]map[string]any, 0, len(items))
		for _, item := range items {
			record := map[string]any{
				"name":            item.Name,
				"display_name":    item.DisplayName,
				"installed":       item.Installed,
				"path":            item.Path,
				"managed_by_brew": item.ManagedByBrew,
			}
			if opts.showVersion {
				record["version"] = item.Version
			}
			payload = append(payload, record)
		}

		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, string(raw))
		return err

	case listOutputTable:
		return renderListTable(out, opts, items)
	default:
		return renderListPlain(out, opts, items)
	}
}

func renderListPlain(out io.Writer, opts listRenderOptions, items []listItem) error {
	for _, item := range items {
		name := item.DisplayName
		if opts.colorOutput {
			if item.Installed {
				name = colorize(colorGreen, name)
			} else {
				name = colorize(colorRed, name)
			}
		}

		if !opts.showVersion && !opts.showPath {
			_, err := fmt.Fprintln(out, name)
			if err != nil {
				return err
			}
			continue
		}

		suffix := itemLineSuffix(item, opts)
		line := name
		if suffix != "" {
			line = fmt.Sprintf("%s %s", name, suffix)
		}
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}

func itemLineSuffix(item listItem, opts listRenderOptions) string {
	suffixParts := make([]string, 0, 2)
	if opts.showVersion {
		if item.Installed {
			suffixParts = append(suffixParts, item.Version)
		} else {
			suffixParts = append(suffixParts, "not found")
		}
	}

	if opts.showPath {
		if item.Installed && item.Path != "" {
			path := item.Path
			if opts.showVersion {
				path = fmt.Sprintf("(%s)", path)
			}
			if !item.ManagedByBrew && opts.colorOutput {
				// keep original behavior: path is highlighted only when installed but not managed by brew
				path = colorize(colorRed, path)
			}
			suffixParts = append(suffixParts, path)
		}

		if !item.Installed && !opts.showVersion {
			suffixParts = append(suffixParts, "not found")
		}
	}

	return strings.Join(suffixParts, " ")
}

func renderListTable(out io.Writer, opts listRenderOptions, items []listItem) error {
	header := "TOOL\t"
	if opts.showVersion {
		header += "VERSION\t"
	}
	if opts.showPath {
		header += "PATH\t"
	}
	header += "MANAGED_BY_BREW"

	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, header); err != nil {
		return err
	}

	for _, item := range items {
		row := []string{item.Name}
		if opts.showVersion {
			if item.Installed {
				row = append(row, item.Version)
			} else {
				row = append(row, "not found")
			}
		}
		if opts.showPath {
			if item.Installed {
				row = append(row, item.Path)
			} else {
				row = append(row, "")
			}
		}
		row = append(row, fmt.Sprintf("%t", item.ManagedByBrew))
		_, err := fmt.Fprintln(tw, strings.Join(row, "\t"))
		if err != nil {
			return err
		}
	}
	return tw.Flush()
}

func useColorOutput(out io.Writer) bool {
	_, ok := out.(*os.File)
	return ok
}

func colorize(color string, text string) string {
	return color + text + colorReset
}
