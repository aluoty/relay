package ascii

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	MaxWidth  = 16
	MaxLines  = 3
	MaxRunes  = 48
)

var Presets = map[string]string{
	"cat":   "=^..^=",
	"bot":   "[o_o]",
	"fox":   "(^^)",
	"star":  "*",
	"wave":  "\\(^o^)/",
	"face":  "(:",
	"cool":  "( B)",
	"heart": "<3",
	"ghost": "(o_o)",
	"sword": "|==>",
	"skull": "x_x",
}

func Resolve(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("avatar cannot be empty")
	}
	if preset, ok := Presets[strings.ToLower(raw)]; ok {
		return Sanitize(preset)
	}
	return Sanitize(raw)
}

func Sanitize(raw string) (string, error) {
	raw = strings.ReplaceAll(raw, `\n`, "\n")
	lines := strings.Split(raw, "\n")
	if len(lines) > MaxLines {
		return "", fmt.Errorf("avatar exceeds %d lines", MaxLines)
	}

	out := make([]string, 0, len(lines))
	total := 0
	for _, line := range lines {
		line = strings.Map(keepASCII, line)
		if utf8.RuneCountInString(line) > MaxWidth {
			return "", fmt.Errorf("avatar line exceeds %d characters", MaxWidth)
		}
		total += utf8.RuneCountInString(line)
		out = append(out, line)
	}
	if total > MaxRunes {
		return "", fmt.Errorf("avatar exceeds %d characters", MaxRunes)
	}
	return strings.Join(out, "\n"), nil
}

func keepASCII(r rune) rune {
	if r >= 0x20 && r <= 0x7E {
		return r
	}
	return -1
}

func Prefix(avatar string) string {
	if avatar == "" {
		return ""
	}
	line, _, _ := strings.Cut(avatar, "\n")
	return line
}

func FormatSpeaker(avatar, name string) string {
	if avatar == "" {
		return name
	}
	return fmt.Sprintf("%s %s", Prefix(avatar), name)
}

func PresetNames() []string {
	names := make([]string, 0, len(Presets))
	for name := range Presets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
