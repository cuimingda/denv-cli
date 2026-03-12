// internal/workflows.go 封装 list/outdated 两条主工作流所需的领域数据结构与计算入口。
package denv

import (
	"fmt"
	"io"
)

// ToolListItem 表示命令列表展示的一条工具记录，既包含元数据也包含当前环境状态。
type ToolListItem struct {
	Name          string
	DisplayName   string
	Installed     bool
	Version       string
	Path          string
	ManagedByBrew bool
}

// ListOptions 控制列表查询时要补充的字段，减少不必要的 I/O 与外部命令调用。
type ListOptions struct {
	ShowVersion bool
	ShowPath    bool
}

type OutdatedState string

// OutdatedState 用语义状态替代字符串拼接，避免后续状态判断散落在字符串判等中。
const (
	OutdatedStateInvalidLatest  OutdatedState = "invalid_latest"
	OutdatedStateInvalidCurrent OutdatedState = "invalid_current"
	OutdatedStateNotInstalled   OutdatedState = "not_installed"
	OutdatedStateUpToDate       OutdatedState = "up_to_date"
	OutdatedStateOutdated       OutdatedState = "outdated"
)

// OutdatedItem 是展示层友好的过期检查结果结构，状态已转为可渲染文本字段。
type OutdatedItem struct {
	Name        string
	DisplayName string
	Current     string
	Latest      string
	State       OutdatedState
	CheckError  string
}

// ToolCheckResult 是命令/API 层内部使用的过期检查结构体，保留原始错误便于上层决策。
type ToolCheckResult struct {
	Name        string
	DisplayName string
	Current     string
	Latest      string
	State       OutdatedState
	Installed   bool
	CheckError  error
}

// OutdatedCheck 是对 ToolCheckResult 的别名，强调该类型在过期检查流程中的语义角色。
type OutdatedCheck = ToolCheckResult

// newOutdatedCheck 创建一条统一初始化的检查记录，减少重复组装逻辑。
func newOutdatedCheck(name, displayName, current, latest string, state OutdatedState) OutdatedCheck {
	return OutdatedCheck{
		Name:        name,
		DisplayName: displayName,
		Current:     current,
		Latest:      latest,
		State:       state,
	}
}

// toItem 将内部校验结果转换为展示层可序列化的 OutdatedItem，并做 error 字符串脱敏。
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

// ListToolItems 计算列表页面需要的工具状态数据。
// 关键思路：
// 1) 先枚举可见工具；
// 2) 再逐个判断安装与路径；
// 3) 最后按开关决定是否解析版本，避免不必要命令执行。
func listToolItems(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy, opts ListOptions) ([]ToolListItem, error) {
	// 只遍历 catalog 对外暴露的工具清单，保持与展示列表一致。
	supported := catalog.listedToolsCatalog()
	// 预分配长度避免 append 扩容。
	items := make([]ToolListItem, 0, len(supported))

	for _, name := range supported {
		// 允许 catalog 懒加载后被替换，兜底保证返回 unsupported 时可明确报错。
		lifecycle, ok := catalog.toolLifecycle(name)
		if !ok {
			return nil, fmt.Errorf("unsupported tool: %s", name)
		}
		// 基础对象先按“未安装/未知”状态初始化，再按探测结果补齐。
		item := ToolListItem{
			Name:          name,
			DisplayName:   lifecycle.DisplayName(name),
			ManagedByBrew: false,
			Installed:     false,
			Version:       "",
			Path:          "",
		}

		// installed 为 true 则 path 和 managed 标识有意义，否则保留空值避免误导。
		installed, commandPath, managedByBrew, err := lifecycle.IsInstalled(rt, catalog, pathPolicy)
		if err != nil {
			return nil, err
		}
		if installed {
			item.Installed = true
			item.Path = commandPath
			item.ManagedByBrew = managedByBrew
		}

		// 只有安装存在时才解析版本，这一点可显著减少慢命令调用。
		if opts.ShowVersion && installed {
			// 尝试从生命周期解析版本，失败视为“当前不可用”，返回空版本保持 list 继续处理可读性。
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

// outdatedChecks 计算每个工具的过期语义状态（内部共享数据模型），后续可映射展示或更新计划。
func outdatedChecks(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy) ([]OutdatedCheck, error) {
	// 先拿工具列表，再逐项构建状态，失败则终止避免给出不完整的误报。
	supported := catalog.listedToolsCatalog()
	rows := make([]OutdatedCheck, 0, len(supported))

	for _, name := range supported {
		row, err := outdatedCheckWithOutput(rt, catalog, pathPolicy, io.Discard, name)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// OutdatedItems 用于对外 API 的兼容结果，返回字符串化错误字段。
func outdatedItems(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy) ([]OutdatedItem, error) {
	rows, err := outdatedChecks(rt, catalog, pathPolicy)
	if err != nil {
		return nil, err
	}
	// 统一转换，避免每个调用端重复做 error 到 string 的映射。
	checks := make([]OutdatedItem, 0, len(rows))
	for _, check := range rows {
		checks = append(checks, check.toItem())
	}
	return checks, nil
}

// OutdatedUpdatePlan 根据检测结果筛选可更新工具，并在检测到版本读取异常时直接失败，阻断盲目更新。
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
			// 未安装但 latest 获取失败允许跳过；其他无效状态应立即打断以避免错误更新。
			if row.Current == "<not installed>" && row.State == OutdatedStateInvalidLatest {
				continue
			}
			return nil, NewOutdatedError(row.Name, row.State)
		}
	}

	return outdated, nil
}

// NewOutdatedError 生成可定位工具与状态的更新阻断错误。
func NewOutdatedError(toolName string, state OutdatedState) error {
	return &OutdatedError{ToolName: toolName, State: state}
}

// OutdatedError 让调用方既能拿到工具名，也能拿到结构化状态便于上层展示与分支处理。
type OutdatedError struct {
	ToolName string
	State    OutdatedState
}

// Error 实现 error 接口，返回工具名和过期状态文本，便于日志与提示文案拼接。
func (e OutdatedError) Error() string {
	return e.ToolName + " has " + string(e.State)
}
