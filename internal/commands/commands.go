package commands

import (
	"strings"
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
		return Command{}, input
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

func HelpText() string {
	return strings.TrimSpace(`
Commands:
  /group <name>     switch channel (e.g. /group random)
  /groups           list channels
  /create <name>    create a new channel
  /avatar <text>    set ASCII avatar (or preset: cat, bot, fox, star, wave...)
  /help             show this help
`)
}
