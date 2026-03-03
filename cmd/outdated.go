package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/cuimingda/denv-cli/internal/denv"
	"github.com/spf13/cobra"
)

func NewOutdatedCmd() *cobra.Command {
	ctx := ensureCLIContext(NewCLIContext())
	return NewOutdatedCmdWithService(outdatedCommandService{
		supportedTools:  ctx.Discovery.SupportedTools,
		outdatedItems:   ctx.VersionResolver.OutdatedItems,
	})
}

func NewOutdatedCmdWithService(svc OutdatedCommandService) *cobra.Command {
	if svc == nil {
		ctx := ensureCLIContext(NewCLIContext())
		svc = outdatedCommandService{
			supportedTools:  ctx.Discovery.SupportedTools,
			outdatedItems:   ctx.VersionResolver.OutdatedItems,
		}
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
			doingf(cmd, "check outdated status for %d tools", len(svc.SupportedTools()))

			rows, err := svc.OutdatedItems()
			if err != nil {
				return err
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

type outdatedCommandService struct {
	supportedTools func() []string
	outdatedItems  func() ([]denv.OutdatedItem, error)
}

func (s outdatedCommandService) SupportedTools() []string {
	return s.supportedTools()
}

func (s outdatedCommandService) OutdatedItems() ([]denv.OutdatedItem, error) {
	return s.outdatedItems()
}

func renderOutdatedPlain(out io.Writer, rows []denv.OutdatedItem, useColor bool) error {
	for _, row := range rows {
		line := renderOutdatedLine(row, useColor)
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}

func renderOutdatedLine(row denv.OutdatedItem, useColor bool) string {
	switch row.State {
	case denv.OutdatedStateInvalidLatest:
		return row.DisplayName + " invalid latest version"
	case denv.OutdatedStateInvalidCurrent:
		return row.DisplayName + " invalid current version"
	case denv.OutdatedStateNotInstalled:
		return row.DisplayName + " <not installed> " + row.Latest
	case denv.OutdatedStateUpToDate:
		current := row.Current
		if useColor {
			current = colorize(colorGreen, current)
		}
		return row.DisplayName + " " + current
	case denv.OutdatedStateOutdated:
		current := row.Current
		if useColor {
			current = colorize(colorRed, current)
		}
		return fmt.Sprintf("%s %s < %s", row.DisplayName, current, row.Latest)
	default:
		return row.DisplayName
	}
}

func renderOutdatedJSON(out io.Writer, rows []denv.OutdatedItem) error {
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

func renderOutdatedTable(out io.Writer, rows []denv.OutdatedItem) error {
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
