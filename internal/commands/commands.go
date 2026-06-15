package commands

import (
	"strings"

	"github.com/aluoty/relay/internal/avatar"
)

type Kind int

const (
	KindNone Kind = iota
	KindHelp
	KindGroup
	KindGroups
	KindCreate
	KindAvatar
)

type Command struct {
	Kind Kind
	Arg  string
}

func Parse(input string) (Command, string) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "/") {
		return Command{}, avatar.ParseText(input)
	}

	fields := strings.Fields(input)
	if len(fields) == 0 {
		return Command{Kind: KindHelp}, ""
	}

	switch strings.ToLower(strings.TrimPrefix(fields[0], "/")) {
	case "help", "h", "?":
		return Command{Kind: KindHelp}, ""
	case "group", "g", "channel", "ch":
		if len(fields) < 2 {
			return Command{Kind: KindHelp}, ""
		}
		return Command{Kind: KindGroup, Arg: normalizeName(fields[1])}, ""
	case "groups", "gs", "channels":
		return Command{Kind: KindGroups}, ""
	case "create", "new":
		if len(fields) < 2 {
			return Command{Kind: KindHelp}, ""
		}
		return Command{Kind: KindCreate, Arg: normalizeName(fields[1])}, ""
	case "avatar", "char", "me":
		if len(fields) < 2 {
			return Command{Kind: KindHelp}, ""
		}
		return Command{Kind: KindAvatar, Arg: strings.Join(fields[1:], " ")}, ""
	default:
		return Command{Kind: KindHelp}, ""
	}
}

func normalizeName(name string) string {
	name = strings.TrimPrefix(strings.TrimSpace(name), "#")
	return strings.ToLower(name)
}

// HelpFormatted returns tview dynamic-color help text for the modal overlay.
func HelpFormatted() string {
	ascii := "cat, bot, star_eyes, happy, sad, wink, awkward, shrug, surprised, angry, bear, robot"
	emoji := "smile, party, wave_e, heart_e, thumbsup, cat_e, rocket, pizza, coffee"

	return strings.TrimSpace(`
[yellow]Commands[white]
  [green]/group[white] [gray]<name>[white]      switch channel       [gray](/g, /channel)[white]
  [green]/groups[white]                         list channels
  [green]/create[white] [gray]<name>[white]     create a channel
  [green]/avatar[white] [gray]<text>[white]      set avatar           [gray](/char, /me)[white]
  [green]/help[white]                            show this help

[yellow]Navigation[white]
  [green]Ctrl+G[white]     focus group list  [gray](toggle — press again to return)[white]
  [green]Esc[white]         return to message input
  [green]Tab[white]         cycle input → groups → users → input
  [green]Enter[white]       send message  [gray](or switch group when list focused)[white]
  [green]1-9[white]         quick group switch  [gray](when group list focused)[white]
  [green]Ctrl+C[white]      quit

[yellow]Chat emojis[white]
  Type GitHub-style aliases in messages:
    [gray]:wave:  :smile:  :+1:  :tada:  :heart:[white]

[yellow]Avatars[white]
  Presets:        [gray]relay avatars[white]  [gray](or relay avatars --ascii / --emoji)[white]
  ASCII preset:   [gray]/avatar star_eyes[white]   →  [gray]*_*[white]
  Emoji preset:   [gray]/avatar smile[white]       →  😄
  Emoji alias:    [gray]/avatar :cat:[white]
  Custom ASCII:   [gray]/avatar ^_^[white]

  ASCII presets:  [gray]` + ascii + `[white]
  Emoji presets:  [gray]` + emoji + `[white]
`)
}
