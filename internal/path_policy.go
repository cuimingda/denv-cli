// internal/path_policy.go 定义路径判定策略接口及默认实现，用于判断命令是否由 Homebrew 管理及推断默认路径。
package denv

import "strings"

type PathPolicy interface {
	IsManagedByHomebrew(path string) bool
	HomebrewDefaultToolPath(name string) string
}

type defaultPathPolicy struct {
	// homebrewPrefixes 保存判断「brew 管理」命令路径时会匹配的前缀集合
	homebrewPrefixes []string
	// homebrewBin 是默认的 brew bin 目录
	homebrewBin      string
}

// DefaultPathPolicy 使用当前项目预设的 Homebrew 目录策略。
func DefaultPathPolicy() PathPolicy {
	return &defaultPathPolicy{
		homebrewPrefixes: []string{"/opt/homebrew/"},
		homebrewBin:      "/opt/homebrew/bin",
	}
}

// IsManagedByHomebrew 通过路径前缀判断是否为 Homebrew 管理的可执行文件。
func (p *defaultPathPolicy) IsManagedByHomebrew(path string) bool {
	for _, prefix := range p.homebrewPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// HomebrewDefaultToolPath 生成工具在默认 Homebrew bin 目录中的路径。
func (p *defaultPathPolicy) HomebrewDefaultToolPath(name string) string {
	return p.homebrewBin + "/" + name
}
