package server

import (
	"fmt"
	"net"
	"sort"
	"sync"

	"github.com/aluoty/relay/internal/protocol"
	"github.com/aluoty/relay/internal/store"
)

type hub struct {
	mu       sync.Mutex
	store    *store.Store
	groups   map[string]struct{}
	sessions map[net.Conn]*session
	byGroup  map[string]map[net.Conn]*session
}

func newHub(st *store.Store, groups []string) *hub {
	h := &hub{
		store:    st,
		groups:   make(map[string]struct{}),
		sessions: make(map[net.Conn]*session),
		byGroup:  make(map[string]map[net.Conn]*session),
	}
	for _, g := range groups {
		g = normalizeGroup(g)
		if validGroup(g) {
			h.groups[g] = struct{}{}
		}
	}
	if len(h.groups) == 0 {
		h.groups[protocol.DefaultGroup] = struct{}{}
	}
	return h
}

func (h *hub) groupNames() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.sortedGroupsLocked()
}

func (h *hub) sortedGroupsLocked() []string {
	names := make([]string, 0, len(h.groups))
	for name := range h.groups {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (h *hub) hasName(name string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, s := range h.sessions {
		if s.name == name {
			return true
		}
	}
	return false
}

func (h *hub) register(s *session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[s.conn] = s
	h.ensureGroupLocked(s.group)
	h.byGroup[s.group][s.conn] = s
}

func (h *hub) unregister(conn net.Conn) (*session, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s, ok := h.sessions[conn]
	if !ok {
		return nil, false
	}
	delete(h.sessions, conn)
	if members, exists := h.byGroup[s.group]; exists {
		delete(members, conn)
	}
	return s, true
}

func (h *hub) ensureGroupLocked(name string) {
	if _, ok := h.byGroup[name]; !ok {
		h.byGroup[name] = make(map[net.Conn]*session)
	}
}

func (h *hub) groupExists(name string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	_, ok := h.groups[name]
	return ok
}

func (h *hub) createGroup(name string) error {
	name = normalizeGroup(name)
	if !validGroup(name) {
		return fmt.Errorf("invalid group name %q (use a-z, 0-9, _, -)", name)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.groups[name] = struct{}{}
	h.ensureGroupLocked(name)
	return nil
}

func (h *hub) switchGroup(s *session, group string) error {
	group = normalizeGroup(group)
	h.mu.Lock()
	if _, ok := h.groups[group]; !ok {
		h.mu.Unlock()
		return fmt.Errorf("group %q does not exist (use /groups)", group)
	}

	old := s.group
	if old == group {
		h.mu.Unlock()
		return nil
	}

	if members, ok := h.byGroup[old]; ok {
		delete(members, s.conn)
	}
	s.group = group
	h.ensureGroupLocked(group)
	h.byGroup[group][s.conn] = s
	h.mu.Unlock()
	return nil
}

func (h *hub) send(conn net.Conn, msg protocol.Message) {
	protocol.Write(conn, msg)
}

func (h *hub) sendToSession(s *session, msg protocol.Message) {
	h.send(s.conn, msg)
}

func (h *hub) broadcastGroup(group string, msg protocol.Message, skip net.Conn) {
	data, err := protocol.MarshalLine(msg)
	if err != nil {
		return
	}

	h.mu.Lock()
	members := h.byGroup[group]
	h.mu.Unlock()

	for conn := range members {
		if conn == skip {
			continue
		}
		conn.Write(data)
	}
}

func (h *hub) broadcastGroupAll(group string, msg protocol.Message) {
	h.broadcastGroup(group, msg, nil)
}

func (h *hub) broadcastGroupsAll() {
	msg := protocol.Message{Type: protocol.TypeGroups, Groups: h.groupNames()}
	data, err := protocol.MarshalLine(msg)
	if err != nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.sessions {
		conn.Write(data)
	}
}

func (h *hub) usersInGroup(group string) ([]string, map[string]string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	members := h.byGroup[group]
	names := make([]string, 0, len(members))
	profiles := make(map[string]string, len(members))
	for _, s := range members {
		names = append(names, s.name)
		if s.avatar != "" {
			profiles[s.name] = s.avatar
		}
	}
	sort.Strings(names)
	return names, profiles
}

func (h *hub) sendGroupSnapshot(s *session) {
	users, profiles := h.usersInGroup(s.group)
	h.sendToSession(s, protocol.Message{
		Type:     protocol.TypeUsers,
		Group:    s.group,
		Users:    users,
		Profiles: profiles,
	})
}

func (h *hub) broadcastGroupUsers(group string) {
	users, profiles := h.usersInGroup(group)
	h.broadcastGroupAll(group, protocol.Message{
		Type:     protocol.TypeUsers,
		Group:    group,
		Users:    users,
		Profiles: profiles,
	})
}

func (h *hub) replayHistory(s *session) {
	for _, msg := range h.store.History(s.group) {
		h.sendToSession(s, msg)
	}
}

func (h *hub) welcome(s *session) {
	h.sendToSession(s, protocol.Message{Type: protocol.TypeGroups, Groups: h.groupNames()})
	h.replayHistory(s)
	h.sendGroupSnapshot(s)
	h.sendToSession(s, protocol.Sys(fmt.Sprintf("connected as %s in #%s", s.name, s.group)))
	h.broadcastGroup(s.group, protocol.Message{Type: protocol.TypeJoin, From: s.name, Group: s.group, Avatar: s.avatar}, s.conn)
	h.broadcastGroupUsers(s.group)
}

func (h *hub) leave(s *session) {
	h.broadcastGroupAll(s.group, protocol.Message{Type: protocol.TypeLeave, From: s.name, Group: s.group})
	h.broadcastGroupUsers(s.group)
}

func (h *hub) enterGroup(s *session) {
	h.replayHistory(s)
	h.sendGroupSnapshot(s)
	h.sendToSession(s, protocol.SysGroup(s.group, fmt.Sprintf("switched to #%s", s.group)))
	h.broadcastGroup(s.group, protocol.Message{Type: protocol.TypeJoin, From: s.name, Group: s.group, Avatar: s.avatar}, s.conn)
	h.broadcastGroupUsers(s.group)
}

func (h *hub) handleChat(s *session, text string) error {
	out := protocol.Message{
		Type:   protocol.TypeChat,
		From:   s.name,
		Text:   text,
		Group:  s.group,
		Avatar: s.avatar,
	}
	if err := h.store.Append(out); err != nil {
		return err
	}
	h.broadcastGroupAll(s.group, out)
	return nil
}

func (h *hub) setAvatar(s *session, avatar string) {
	s.avatar = avatar
	h.broadcastGroupAll(s.group, protocol.Message{
		Type:   protocol.TypeAvatar,
		From:   s.name,
		Group:  s.group,
		Avatar: avatar,
	})
	h.broadcastGroupUsers(s.group)
}
