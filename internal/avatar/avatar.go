package avatar

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/enescakir/emoji"
	"github.com/mattn/go-runewidth"
)

const (
	MaxWidth = 16
	MaxLines = 3
	MaxRunes = 48
)

var ASCIIPresets = map[string]string{
	"cat":       "=^..^=",
	"bot":       "[o_o]",
	"fox":       "(^^)",
	"star":      "*",
	"wave":      "\\(^o^)/",
	"face":      "(:",
	"cool":      "( B)",
	"heart":     "<3",
	"ghost":     "(o_o)",
	"sword":     "|==>",
	"skull":     "x_x",
	"star_eyes": "*_*",
	"happy":     "^_^",
	"sad":       "T_T",
	"cry":       "T_T",
	"confused":  "@_@",
	"sleepy":    "-_-",
	"wink":      "^_~",
	"surprised": "O_O",
	"shocked":   "0_0",
	"awkward":   ">_<",
	"angry":     ">:(",
	"derp":      ":P",
	"smirk":     ":3",
	"blank":     "._.",
	"dead":      "x_x",
	"bear":      "(`.`)",
	"fish":      "<><",
	"bird":      ">v<",
	"bug":       "[\\_/]",
	"robot":     "[::]",
	"alien":     "(@_@)",
	"diamond":   "<><>",
	"music":     "(~_~)",
	"sparkle":   "*+*",
	"fire":      "(/\\)",
	"shrug":     "\\_( '-')_/",
}

var EmojiPresets = map[string]string{
	"smile":    ":smile:",
	"grin":     ":grin:",
	"joy":      ":joy:",
	"laugh":    ":laughing:",
	"wink_e":   ":wink:",
	"heart_e":  ":heart:",
	"fire_e":   ":fire:",
	"star_e":   ":star:",
	"thumbsup": ":+1:",
	"thumbsdn": ":-1:",
	"clap":     ":clap:",
	"wave_e":   ":wave:",
	"eyes":     ":eyes:",
	"think":    ":thinking:",
	"cool_e":   ":sunglasses:",
	"party":    ":tada:",
	"100":      ":100:",
	"rocket":   ":rocket:",
	"cat_e":    ":cat:",
	"dog":      ":dog:",
	"fox_e":    ":fox_face:",
	"pizza":    ":pizza:",
	"coffee":   ":coffee:",
	"moon":     ":crescent_moon:",
	"sun":      ":sunny:",
	"rainbow":  ":rainbow:",
	"ghost_e":  ":ghost:",
	"skull_e":  ":skull:",
	"robot_e":  ":robot:",
	"alien_e":  ":alien:",
}

func Resolve(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("avatar cannot be empty")
	}

	key := strings.ToLower(raw)
	if preset, ok := ASCIIPresets[key]; ok {
		return sanitizeASCII(preset)
	}
	if alias, ok := EmojiPresets[key]; ok {
		return sanitizeDisplay(emoji.Parse(alias))
	}
	if strings.Contains(raw, ":") {
		parsed := emoji.Parse(raw)
		if parsed != raw {
			return sanitizeDisplay(parsed)
		}
	}
	return sanitizeASCII(raw)
}

func ParseText(text string) string {
	return emoji.Parse(text)
}

func sanitizeASCII(raw string) (string, error) {
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

func sanitizeDisplay(raw string) (string, error) {
	raw = strings.ReplaceAll(raw, `\n`, "\n")
	lines := strings.Split(raw, "\n")
	if len(lines) > MaxLines {
		return "", fmt.Errorf("avatar exceeds %d lines", MaxLines)
	}

	out := make([]string, 0, len(lines))
	total := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if runewidth.StringWidth(line) > MaxWidth {
			return "", fmt.Errorf("avatar line exceeds %d columns", MaxWidth)
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

func Prefix(v string) string {
	if v == "" {
		return ""
	}
	line, _, _ := strings.Cut(v, "\n")
	return line
}

func FormatSpeaker(v, name string) string {
	if v == "" {
		return name
	}
	return fmt.Sprintf("%s %s", Prefix(v), name)
}

func PresetNames() []string {
	seen := make(map[string]struct{})
	names := make([]string, 0, len(ASCIIPresets)+len(EmojiPresets))
	for name := range ASCIIPresets {
		seen[name] = struct{}{}
		names = append(names, name)
	}
	for name := range EmojiPresets {
		if _, ok := seen[name]; ok {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func HelpPresets() string {
	return strings.Join(PresetNames(), ", ")
}
