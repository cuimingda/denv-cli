package denv

import "strings"

type PathPolicy interface {
	IsManagedByHomebrew(path string) bool
	HomebrewDefaultToolPath(name string) string
}

type defaultPathPolicy struct {
	homebrewPrefixes []string
	homebrewBin      string
}

func DefaultPathPolicy() PathPolicy {
	return &defaultPathPolicy{
		homebrewPrefixes: []string{"/opt/homebrew/"},
		homebrewBin:      "/opt/homebrew/bin",
	}
}

func (p *defaultPathPolicy) IsManagedByHomebrew(path string) bool {
	for _, prefix := range p.homebrewPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

func (p *defaultPathPolicy) HomebrewDefaultToolPath(name string) string {
	return p.homebrewBin + "/" + name
}
