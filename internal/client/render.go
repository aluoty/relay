package client

import (
	"sort"
	"strings"

	"github.com/aluoty/relay/internal/ascii"
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

func newState(addr, self, group, avatar string) *state {
	return &state{
		addr:    addr,
		self:    self,
		group:   group,
		avatar:  avatar,
		groups:  []string{group},
		avatars: map[string]string{self: avatar},
	}
}

func (s *state) setSelfAvatar(avatar string) {
	s.avatar = avatar
	if avatar == "" {
		delete(s.avatars, s.self)
		return
	}
	s.avatars[s.self] = avatar
}

func (s *state) setUserAvatar(name, avatar string) {
	if avatar == "" {
		delete(s.avatars, name)
		return
	}
	s.avatars[name] = avatar
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
	for name, avatar := range profiles {
		s.setUserAvatar(name, avatar)
	}
}

func (s *state) statusText() string {
	return strings.TrimSpace(formatStatus(s.addr, s.self, s.group))
}

func formatStatus(addr, self, group string) string {
	return " [gray]connected to " + addr + " as [green]" + self + "[white] in [#" + group + "]"
}

func formatGroupLabel(name, current string) string {
	if name == current {
		return "[green]# " + name
	}
	return "# " + name
}

func formatUserLine(name, avatar, self string) string {
	label := ascii.FormatSpeaker(avatar, name)
	if name == self {
		return "[green]" + label + "[white]"
	}
	return label
}

func formatChatLine(from, avatar, text, self string) string {
	speaker := ascii.FormatSpeaker(avatar, from)
	if from == self {
		return "[green]" + speaker + ":[white] " + text
	}
	return "[yellow]" + speaker + ":[white] " + text
}
