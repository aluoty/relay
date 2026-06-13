package server

import (
	"regexp"
	"strings"
)

var groupPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,31}$`)

func normalizeGroup(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	name = strings.TrimPrefix(name, "#")
	return name
}

func validGroup(name string) bool {
	return groupPattern.MatchString(name)
}
