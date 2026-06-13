package client

import "strings"

// escapeTview makes user text safe for TextView dynamic color mode.
// Literal "[" must be written as "[[]".
func escapeTview(s string) string {
	return strings.ReplaceAll(s, "[", "[[]")
}
