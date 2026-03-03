package denv

import "fmt"

// ToolListItem describes one tool entry returned for list rendering.
type ToolListItem struct {
	Name          string
	DisplayName   string
	Installed     bool
	Version       string
	Path          string
	ManagedByBrew bool
}

// ListOptions controls which data fields are resolved for list output.
type ListOptions struct {
	ShowVersion bool
	ShowPath    bool
}

type OutdatedState string

const (
	OutdatedStateInvalidLatest  OutdatedState = "invalid_latest"
	OutdatedStateInvalidCurrent OutdatedState = "invalid_current"
	OutdatedStateNotInstalled   OutdatedState = "not_installed"
	OutdatedStateUpToDate       OutdatedState = "up_to_date"
	OutdatedStateOutdated       OutdatedState = "outdated"
)

// OutdatedItem describes one tool entry for outdated checks.
type OutdatedItem struct {
	Name        string
	DisplayName string
	Current     string
	Latest      string
	State       OutdatedState
	CheckError  string
}

type ToolCheckResult struct {
	Name        string
	DisplayName string
	Current     string
	Latest      string
	State       OutdatedState
	Installed   bool
	CheckError  error
}

type OutdatedCheck = ToolCheckResult

func newOutdatedCheck(name, displayName, current, latest string, state OutdatedState) OutdatedCheck {
	return OutdatedCheck{
		Name:        name,
		DisplayName: displayName,
		Current:     current,
		Latest:      latest,
		State:       state,
	}
}

func (c OutdatedCheck) toItem() OutdatedItem {
	checkErr := ""
	if c.CheckError != nil {
		checkErr = c.CheckError.Error()
	}

	return OutdatedItem{
		Name:        c.Name,
		DisplayName: c.DisplayName,
		Current:     c.Current,
		Latest:      c.Latest,
		State:       c.State,
		CheckError:  checkErr,
	}
}

// ListToolItems resolves listing information for all supported tools.
func listToolItems(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy, opts ListOptions) ([]ToolListItem, error) {
	supported := catalog.listedToolsCatalog()
	items := make([]ToolListItem, 0, len(supported))

	for _, name := range supported {
		lifecycle, ok := catalog.toolLifecycle(name)
		if !ok {
			return nil, fmt.Errorf("unsupported tool: %s", name)
		}
		item := ToolListItem{
			Name:          name,
			DisplayName:   lifecycle.DisplayName(name),
			ManagedByBrew: false,
			Installed:     false,
			Version:       "",
			Path:          "",
		}

		installed, commandPath, managedByBrew, err := lifecycle.IsInstalled(rt, catalog, pathPolicy)
		if err != nil {
			return nil, err
		}
		if installed {
			item.Installed = true
			item.Path = commandPath
			item.ManagedByBrew = managedByBrew
		}

		if opts.ShowVersion && installed {
			toolVersion, err := lifecycle.ResolveVersion(rt, catalog, commandPath, false)
			if err != nil {
				item.Installed = false
				item.Path = ""
				item.ManagedByBrew = false
			} else {
				item.Version = toolVersion
			}
		}

		items = append(items, item)
	}

	return items, nil
}

// OutdatedChecks evaluates outdated states for all supported tools as semantic values.
func outdatedChecks(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy) ([]OutdatedCheck, error) {
	supported := catalog.listedToolsCatalog()
	rows := make([]OutdatedCheck, 0, len(supported))

	for _, name := range supported {
		lifecycle, ok := catalog.toolLifecycle(name)
		if !ok {
			return nil, fmt.Errorf("unsupported tool: %s", name)
		}
		row := newOutdatedCheck(name, lifecycle.DisplayName(name), "", "", OutdatedStateNotInstalled)

		installed, _, _, err := lifecycle.IsInstalled(rt, catalog, pathPolicy)
		if err != nil {
			return nil, err
		}
		row.Installed = installed

		if !installed {
			row.Current = "<not installed>"
			latest, latestErr := lifecycle.ResolveLatestVersion(rt, catalog)
			if latestErr != nil {
				row.State = OutdatedStateInvalidLatest
				row.CheckError = latestErr
			} else {
				row.State = OutdatedStateNotInstalled
				row.Latest = latest
			}

			rows = append(rows, row)
			continue
		}

		current, currentErr := lifecycle.ResolveOutdatedCurrentVersion(rt, catalog)
		if currentErr != nil {
			row.State = OutdatedStateInvalidCurrent
			row.CheckError = currentErr
			rows = append(rows, row)
			continue
		}

		latest, latestErr := lifecycle.ResolveLatestVersion(rt, catalog)
		if latestErr != nil {
			row.State = OutdatedStateInvalidLatest
			row.CheckError = latestErr
			rows = append(rows, row)
			continue
		}

		row.Current = current
		row.Latest = latest
		if CompareVersions(current, latest) < 0 {
			row.State = OutdatedStateOutdated
		} else {
			row.State = OutdatedStateUpToDate
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// OutdatedItems returns outdated items for API compatibility.
func outdatedItems(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy) ([]OutdatedItem, error) {
	rows, err := outdatedChecks(rt, catalog, pathPolicy)
	if err != nil {
		return nil, err
	}
	checks := make([]OutdatedItem, 0, len(rows))
	for _, check := range rows {
		checks = append(checks, check.toItem())
	}
	return checks, nil
}

// OutdatedUpdatePlan extracts only tools that should be updated and rejects invalid version states.
func outdatedUpdatePlan(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy) ([]OutdatedItem, error) {
	rows, err := outdatedChecks(rt, catalog, pathPolicy)
	if err != nil {
		return nil, err
	}

	outdated := make([]OutdatedItem, 0, len(rows))
	for _, row := range rows {
		switch row.State {
		case OutdatedStateOutdated:
			outdated = append(outdated, row.toItem())
		case OutdatedStateInvalidCurrent, OutdatedStateInvalidLatest:
			if row.Current == "<not installed>" && row.State == OutdatedStateInvalidLatest {
				continue
			}
			return nil, NewOutdatedError(row.Name, row.State)
		}
	}

	return outdated, nil
}

// NewOutdatedError reports why outdated planning cannot continue.
func NewOutdatedError(toolName string, state OutdatedState) error {
	return &OutdatedError{ToolName: toolName, State: state}
}

type OutdatedError struct {
	ToolName string
	State    OutdatedState
}

func (e OutdatedError) Error() string {
	return e.ToolName + " has " + string(e.State)
}
