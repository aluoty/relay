package server

import (
	"net"

	"github.com/aluoty/relay/internal/protocol"
)

type session struct {
	conn   net.Conn
	name   string
	avatar string
	group  string
}

func newSession(conn net.Conn, name, group, avatar string) *session {
	if group == "" {
		group = protocol.DefaultGroup
	}
	return &session{conn: conn, name: name, group: group, avatar: avatar}
}
