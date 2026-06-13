package client

import (
	"sort"
	"strings"

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
	return strings.TrimSpace(formatStatus(s.addr, s.self, s.group))
}

func formatStatus(addr, self, group string) string {
	return " [gray]connected to " + addr + " as [green]" + self + "[white] in [#" + group + "]  [gray]| Ctrl+G groups | 1-9 switch"
}

func formatGroupLabel(name, current string) string {
	if name == current {
		return "[green]# " + name
	}
	return "# " + name
}

func formatUserLine(name, v, self string) string {
	label := avatar.FormatSpeaker(v, name)
	if name == self {
		return "[green]" + label + "[white]"
	}
	return label
}

func formatChatLine(from, v, text, self string) string {
	speaker := avatar.FormatSpeaker(v, from)
	text = avatar.ParseText(text)
	if from == self {
		return "[green]" + speaker + ":[white] " + text
	}
	return "[yellow]" + speaker + ":[white] " + text
}
