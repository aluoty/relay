package client

import (
	"sort"

	"github.com/aluoty/relay/internal/avatar"
)

type state struct {
	addr    string
	self    string
	group   string
	avatar  string
	groups  []string
	users   []string
	avatars map[string]string
}

func newState(addr, self, group, avatarText string) *state {
	return &state{
		addr:    addr,
		self:    self,
		group:   group,
		avatar:  avatarText,
		groups:  []string{group},
		avatars: map[string]string{self: avatarText},
	}
}

func (s *state) setSelfAvatar(v string) {
	s.avatar = v
	if v == "" {
		delete(s.avatars, s.self)
		return
	}
	s.avatars[s.self] = v
}

func (s *state) setUserAvatar(name, v string) {
	if v == "" {
		delete(s.avatars, name)
		return
	}
	s.avatars[name] = v
}

func (s *state) avatarFor(name string) string {
	return s.avatars[name]
}

func (s *state) setGroups(groups []string) {
	s.groups = append([]string(nil), groups...)
	sort.Strings(s.groups)
}

func (s *state) setUsers(users []string, profiles map[string]string) {
	s.users = append([]string(nil), users...)
	sort.Strings(s.users)
	for name, v := range profiles {
		s.setUserAvatar(name, v)
	}
}

func (s *state) statusText() string {
	return formatStatus(s.addr, s.self, s.group)
}

func formatStatus(addr, self, group string) string {
	return " [gray]connected to " + escapeTview(addr) + " as [green]" + escapeTview(self) +
		"[white] in [#" + escapeTview(group) + "]  [gray]| Ctrl+G groups | Esc chat"
}

func formatGroupsStatus() string {
	return " [yellow]groups[white]  [gray]Enter switch · 1-9 quick · Esc back to chat · Ctrl+G toggle"
}

func formatGroupLabel(name, current string) string {
	prefix := "  # "
	if name == current {
		prefix = "● # "
	}
	return prefix + name
}

func formatUserLine(name, v, self string) string {
	speaker := escapeTview(avatar.FormatSpeaker(v, name))
	if name == self {
		return "[green]" + speaker + "[white]"
	}
	return speaker
}

func formatChatLine(from, v, text, self string) string {
	speaker := escapeTview(avatar.FormatSpeaker(v, from))
	text = escapeTview(avatar.ParseText(text))
	if from == self {
		return "[green]" + speaker + ":[white] " + text
	}
	return "[yellow]" + speaker + ":[white] " + text
}

func formatSystemLine(text string) string {
	return "[gray]* " + escapeTview(text)
}
