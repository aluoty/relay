package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/aluoty/relay/internal/avatar"
	"github.com/aluoty/relay/internal/protocol"
	"github.com/aluoty/relay/internal/store"
	"github.com/aluoty/relay/internal/tlsconfig"
)

type Config struct {
	Listen  string
	Groups  []string
	TLS     tlsconfig.ServerConfig
	History string
	Limit   int
}

func Run(cfg Config) error {
	st, err := store.Open(cfg.History, cfg.Limit)
	if err != nil {
		return fmt.Errorf("open history: %w", err)
	}

	ln, err := cfg.TLS.Listen(cfg.Listen)
	if err != nil {
		return err
	}
	defer ln.Close()

	logStartup(cfg)

	h := newHub(st, cfg.Groups)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go serveClient(h, conn)
	}
}

func logStartup(cfg Config) {
	if cfg.TLS.Enabled() {
		log.Printf("relay listening on %s (tls)", cfg.Listen)
	} else {
		log.Printf("relay listening on %s", cfg.Listen)
	}
	if cfg.History != "" {
		log.Printf("history: %s (limit %d)", cfg.History, cfg.Limit)
	}
	log.Printf("groups: %s", strings.Join(normalizeGroupList(cfg.Groups), ", "))
}

func normalizeGroupList(groups []string) []string {
	h := newHub(nil, groups)
	return h.groupNames()
}

func serveClient(h *hub, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	join, err := protocol.Read(reader)
	if err != nil || join.Type != protocol.TypeJoin || join.From == "" {
		protocol.Write(conn, protocol.Sys("expected join message"))
		return
	}

	name := strings.TrimSpace(join.From)
	group := normalizeGroup(join.Group)
	if group == "" {
		group = protocol.DefaultGroup
	}

	avatar, err := avatar.Resolve(join.Avatar)
	if err != nil && join.Avatar != "" {
		protocol.Write(conn, protocol.Sys(err.Error()))
		return
	}

	if !validGroup(group) {
		protocol.Write(conn, protocol.Sys(fmt.Sprintf("invalid group %q", group)))
		return
	}

	if h.hasName(name) {
		protocol.Write(conn, protocol.Sys(fmt.Sprintf("name %q is already in use", name)))
		return
	}

	if !h.groupExists(group) {
		protocol.Write(conn, protocol.Sys(fmt.Sprintf("group %q does not exist (try /groups)", group)))
		return
	}

	s := newSession(conn, name, group, avatar)
	h.register(s)
	h.welcome(s)

	for {
		msg, err := protocol.Read(reader)
		if err != nil {
			break
		}
		if err := dispatch(h, s, msg); err != nil {
			h.sendToSession(s, protocol.Sys(err.Error()))
		}
	}

	if _, ok := h.unregister(conn); ok {
		h.leave(s)
	}
}

func dispatch(h *hub, s *session, msg protocol.Message) error {
	switch msg.Type {
	case protocol.TypeChat:
		text := strings.TrimSpace(msg.Text)
		if text == "" {
			return nil
		}
		return h.handleChat(s, text)

	case protocol.TypeSwitch:
		group := normalizeGroup(msg.Group)
		old := s.group
		if err := h.switchGroup(s, group); err != nil {
			return err
		}
		if old != s.group {
			h.leaveSnapshot(old, s)
			h.sendToSession(s, protocol.Message{Type: protocol.TypeSwitch, Group: s.group})
			h.enterGroup(s)
		}
		return nil

	case protocol.TypeCreate:
		name := normalizeGroup(msg.Text)
		if err := h.createGroup(name); err != nil {
			return err
		}
		h.broadcastGroupsAll()
		h.sendToSession(s, protocol.Sys(fmt.Sprintf("created group #%s", name)))
		return nil

	case protocol.TypeAvatar:
		avatar, err := avatar.Resolve(msg.Avatar)
		if err != nil {
			return err
		}
		h.setAvatar(s, avatar)
		return nil
	}
	return nil
}

func (h *hub) leaveSnapshot(group string, s *session) {
	h.broadcastGroupAll(group, protocol.Message{Type: protocol.TypeLeave, From: s.name, Group: group})
	h.broadcastGroupUsers(group)
}
