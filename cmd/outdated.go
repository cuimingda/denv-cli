package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

func NewOutdatedCmd() *cobra.Command {
	return NewOutdatedCmdWithService(NewCLIContext().Service)
}

func NewOutdatedCmdWithService(svc CommandService) *cobra.Command {
	if svc == nil {
		svc = NewCLIContext().Service
	}

	cmd := &cobra.Command{
		Use:   "outdated",
		Short: "Show outdated status for supported developer tools",
		RunE: func(cmd *cobra.Command, _ []string) error {
			outputMode, _ := cmd.Flags().GetString("output")
			mode, err := parseListOutput(outputMode)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			colorOutput := useColorOutput(out) && mode != listOutputNoColor
			start := time.Now()
			supported := svc.SupportedTools()
			doingf(cmd, "check outdated status for %d tools", len(supported))

			rows := make([]outdatedItem, 0, len(supported))
			for _, name := range supported {
				doingf(cmd, "checking %s", name)
				row := outdatedItem{
					Name:        name,
					DisplayName: svc.ToolDisplayName(name),
				}

				installed, _, _, err := svc.ToolInstallState(name)
				if err != nil {
					return err
				}

				if !installed {
					row.Current = "<not installed>"
					latest, latestErr := svc.ToolLatestVersion(name)
					if latestErr != nil {
						row.State = "invalid_latest"
					} else {
						row.State = "not_installed"
						row.Latest = latest
					}
					rows = append(rows, row)
					continue
				}

				current, currentErr := svc.ToolVersionForOutdated(name)
				if currentErr != nil {
					row.State = "invalid_current"
					rows = append(rows, row)
					continue
				}

				latest, latestErr := svc.ToolLatestVersion(name)
				if latestErr != nil {
					row.State = "invalid_latest"
					rows = append(rows, row)
					continue
				}

				row.Current = current
				row.Latest = latest
				if svc.CompareVersions(current, latest) < 0 {
					row.State = "outdated"
				} else {
					row.State = "up_to_date"
				}

				rows = append(rows, row)
			}
			doingf(cmd, "outdated check completed in %s", time.Since(start))

			switch mode {
			case listOutputJSON:
				return renderOutdatedJSON(out, rows)
			case listOutputTable:
				return renderOutdatedTable(out, rows)
			case listOutputNoColor, listOutputPlain:
				return renderOutdatedPlain(out, rows, colorOutput)
			default:
				return renderOutdatedPlain(out, rows, colorOutput)
			}
		},
	}

	cmd.Flags().String("output", string(listOutputPlain), "output format: plain|json|table|no-color")
	return cmd
}

type outdatedItem struct {
	Name        string
	DisplayName string
	Current     string
	Latest      string
	State       string
}

func renderOutdatedPlain(out io.Writer, rows []outdatedItem, useColor bool) error {
	for _, row := range rows {
		line := renderOutdatedLine(row, useColor)
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}

func renderOutdatedLine(row outdatedItem, useColor bool) string {
	switch row.State {
	case "invalid_latest":
		return row.DisplayName + " invalid latest version"
	case "invalid_current":
		return row.DisplayName + " invalid current version"
	case "not_installed":
		return row.DisplayName + " <not installed> " + row.Latest
	case "up_to_date":
		current := row.Current
		if useColor {
			current = colorize(colorGreen, current)
		}
		return row.DisplayName + " " + current
	case "outdated":
		current := row.Current
		if useColor {
			current = colorize(colorRed, current)
		}
		return fmt.Sprintf("%s %s < %s", row.DisplayName, current, row.Latest)
	default:
		return row.DisplayName
	}
}

func renderOutdatedJSON(out io.Writer, rows []outdatedItem) error {
	payload := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		record := map[string]any{
			"name":         row.Name,
			"display_name": row.DisplayName,
			"state":        row.State,
		}
		if row.Current != "" {
			record["current"] = row.Current
		}
		if row.Latest != "" {
			record["latest"] = row.Latest
		}
		payload = append(payload, record)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, string(raw))
	return err
}

func renderOutdatedTable(out io.Writer, rows []outdatedItem) error {
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tCURRENT\tLATEST\tSTATE"); err != nil {
		return err
	}

	for _, row := range rows {
		current := row.Current
		if current == "" {
			current = "n/a"
		}
		latest := row.Latest
		if latest == "" {
			latest = "n/a"
		}
		if _, err := fmt.Fprintln(tw, strings.Join([]string{row.Name, current, latest, row.State}, "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func isValidOutputMode(mode listOutputMode) bool {
	switch mode {
	case listOutputPlain, listOutputJSON, listOutputTable, listOutputNoColor:
		return true
	default:
		return false
	}
}
