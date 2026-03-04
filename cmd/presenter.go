// cmd/presenter.go 负责 list/outdated 的输出呈现层，统一 JSON/plain/table 等格式化结果。
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/cuimingda/denv-cli/internal"
)

type ToolPresenter interface {
	Render(out io.Writer) error
}

// NewListPresenter 创建列表展示实现。
func NewListPresenter(mode listOutputMode, opts listRenderOptions, items []denv.ToolListItem) ToolPresenter {
	return listPresenter{
		mode:  mode,
		opts:  opts,
		items: items,
	}
}

// NewOutdatedPresenter 创建过期检查展示实现。
func NewOutdatedPresenter(mode listOutputMode, rows []denv.ToolCheckResult, useColor bool) ToolPresenter {
	return outdatedPresenter{
		mode:     mode,
		rows:     rows,
		useColor: useColor,
	}
}

type listPresenter struct {
	mode  listOutputMode
	opts  listRenderOptions
	items []denv.ToolListItem
}

// Render 按模式分发到不同输出策略。
func (p listPresenter) Render(out io.Writer) error {
	switch p.mode {
	case listOutputJSON:
		return p.renderJSON(out)
	case listOutputTable:
		return p.renderTable(out)
	default:
		return p.renderPlain(out)
	}
}

// renderJSON 输出 JSON 结构。
func (p listPresenter) renderJSON(out io.Writer) error {
	payload := make([]map[string]any, 0, len(p.items))
	for _, item := range p.items {
		record := map[string]any{
			"name":            item.Name,
			"display_name":    item.DisplayName,
			"installed":       item.Installed,
			"path":            item.Path,
			"managed_by_brew": item.ManagedByBrew,
		}
		if p.opts.showVersion {
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
}

// renderPlain 输出纯文本（或行内附加字段）。
func (p listPresenter) renderPlain(out io.Writer) error {
	for _, item := range p.items {
		name := item.DisplayName
		if p.opts.colorOutput {
			if item.Installed {
				name = colorize(colorGreen, name)
			} else {
				name = colorize(colorRed, name)
			}
		}

		if !p.opts.showVersion && !p.opts.showPath {
			_, err := fmt.Fprintln(out, name)
			if err != nil {
				return err
			}
			continue
		}

		suffix := p.itemLineSuffix(item)
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

// itemLineSuffix 负责拼装单行后缀（版本/路径）。
func (p listPresenter) itemLineSuffix(item denv.ToolListItem) string {
	suffixParts := make([]string, 0, 2)
	if p.opts.showVersion {
		if item.Installed {
			suffixParts = append(suffixParts, item.Version)
		} else {
			suffixParts = append(suffixParts, "not found")
		}
	}

	if p.opts.showPath {
		if item.Installed && item.Path != "" {
			path := item.Path
			if p.opts.showVersion {
				path = fmt.Sprintf("(%s)", path)
			}
			if !item.ManagedByBrew && p.opts.colorOutput {
				// preserve previous behavior: only path for non-brew installed tools is highlighted
				path = colorize(colorRed, path)
			}
			suffixParts = append(suffixParts, path)
		}

		if !item.Installed && !p.opts.showVersion {
			suffixParts = append(suffixParts, "not found")
		}
	}

	return strings.Join(suffixParts, " ")
}

// renderTable 输出表格文本。
func (p listPresenter) renderTable(out io.Writer) error {
	header := "TOOL\t"
	if p.opts.showVersion {
		header += "VERSION\t"
	}
	if p.opts.showPath {
		header += "PATH\t"
	}
	header += "MANAGED_BY_BREW"

	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, header); err != nil {
		return err
	}

	for _, item := range p.items {
		row := []string{item.Name}
		if p.opts.showVersion {
			if item.Installed {
				row = append(row, item.Version)
			} else {
				row = append(row, "not found")
			}
		}
		if p.opts.showPath {
			if item.Installed {
				row = append(row, item.Path)
			} else {
				row = append(row, "")
			}
		}
		row = append(row, fmt.Sprintf("%t", item.ManagedByBrew))
		if _, err := fmt.Fprintln(tw, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

type outdatedPresenter struct {
	mode     listOutputMode
	rows     []denv.ToolCheckResult
	useColor bool
}

// Render 按过期状态输出 JSON/表格/文本。
func (p outdatedPresenter) Render(out io.Writer) error {
	switch p.mode {
	case listOutputJSON:
		return p.renderJSON(out)
	case listOutputTable:
		return p.renderTable(out)
	default:
		return p.renderPlain(out)
	}
}

// renderPlain 输出过期工具状态（带可读提示）。
func (p outdatedPresenter) renderPlain(out io.Writer) error {
	for _, row := range p.rows {
		line := renderOutdatedLine(row, p.useColor)
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}

// renderJSON 输出过期检查 JSON 结果。
func (p outdatedPresenter) renderJSON(out io.Writer) error {
	payload := make([]map[string]any, 0, len(p.rows))
	for _, row := range p.rows {
		record := map[string]any{
			"name":         row.Name,
			"display_name": row.DisplayName,
			"state":        row.State,
		}
		if row.CheckError != nil {
			record["check_error"] = row.CheckError.Error()
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

// renderTable 输出过期状态表格。
func (p outdatedPresenter) renderTable(out io.Writer) error {
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NAME\tCURRENT\tLATEST\tSTATE"); err != nil {
		return err
	}

	for _, row := range p.rows {
		current := row.Current
		if current == "" {
			current = "n/a"
		}
		latest := row.Latest
		if latest == "" {
			latest = "n/a"
		}
		state := string(row.State)
		if row.CheckError != nil {
			state = state + " (" + row.CheckError.Error() + ")"
		}
		if _, err := fmt.Fprintln(tw, strings.Join([]string{row.Name, current, latest, state}, "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// renderOutdatedLine 按状态拼装人类可读字符串。
func renderOutdatedLine(row denv.ToolCheckResult, useColor bool) string {
	errorSuffix := ""
	if row.CheckError != nil {
		errorSuffix = " (" + row.CheckError.Error() + ")"
	}

	switch row.State {
	case denv.OutdatedStateInvalidLatest:
		return row.DisplayName + " invalid latest version" + errorSuffix
	case denv.OutdatedStateInvalidCurrent:
		return row.DisplayName + " invalid current version" + errorSuffix
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
