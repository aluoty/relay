package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"

	"github.com/aluoty/relay/internal/protocol"
	"github.com/aluoty/relay/internal/store"
	"github.com/aluoty/relay/internal/tlsconfig"
)

type Config struct {
	Listen  string
	TLS     tlsconfig.ServerConfig
	History string
	Limit   int
}

type room struct {
	mu      sync.Mutex
	clients map[net.Conn]string
	store   *store.Store
}

func newRoom(st *store.Store) *room {
	return &room{clients: make(map[net.Conn]string), store: st}
}

func (r *room) hasName(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, n := range r.clients {
		if n == name {
			return true
		}
	}
	return false
}

func (r *room) add(conn net.Conn, name string) {
	r.mu.Lock()
	r.clients[conn] = name
	r.mu.Unlock()
}

func (r *room) remove(conn net.Conn) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name, ok := r.clients[conn]
	delete(r.clients, conn)
	return name, ok
}

func (r *room) names() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	names := make([]string, 0, len(r.clients))
	for _, name := range r.clients {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *room) send(conn net.Conn, msg protocol.Message) {
	protocol.Write(conn, msg)
}

func (r *room) broadcast(msg protocol.Message, skip net.Conn) {
	data, err := protocol.MarshalLine(msg)
	if err != nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	for conn := range r.clients {
		if conn == skip {
			continue
		}
		conn.Write(data)
	}
}

func (r *room) broadcastAll(msg protocol.Message) {
	r.broadcast(msg, nil)
}

func (r *room) sendUsers(conn net.Conn) {
	r.send(conn, protocol.Message{Type: protocol.TypeUsers, Users: r.names()})
}

func (r *room) broadcastUsers() {
	r.broadcastAll(protocol.Message{Type: protocol.TypeUsers, Users: r.names()})
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

	if cfg.TLS.Enabled() {
		log.Printf("relay listening on %s (tls)", cfg.Listen)
	} else {
		log.Printf("relay listening on %s", cfg.Listen)
	}
	if cfg.History != "" {
		log.Printf("history: %s (limit %d)", cfg.History, cfg.Limit)
	}

	room := newRoom(st)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handleClient(room, conn)
	}
}

func handleClient(room *room, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	join, err := protocol.Read(reader)
	if err != nil || join.Type != protocol.TypeJoin || join.From == "" {
		protocol.Write(conn, protocol.Message{Type: protocol.TypeSys, Text: "expected join message"})
		return
	}

	name := join.From
	if room.hasName(name) {
		protocol.Write(conn, protocol.Message{Type: protocol.TypeSys, Text: fmt.Sprintf("name %q is already in use", name)})
		return
	}

	room.add(conn, name)

	for _, msg := range room.store.History() {
		room.send(conn, msg)
	}

	room.sendUsers(conn)
	room.send(conn, protocol.Message{Type: protocol.TypeSys, Text: fmt.Sprintf("connected as %s", name)})
	room.broadcast(protocol.Message{Type: protocol.TypeJoin, From: name}, conn)
	room.broadcastUsers()

	for {
		msg, err := protocol.Read(reader)
		if err != nil {
			break
		}
		if msg.Type != protocol.TypeChat || msg.Text == "" {
			continue
		}
		out := protocol.Message{Type: protocol.TypeChat, From: name, Text: msg.Text}
		if err := room.store.Append(out); err != nil {
			log.Printf("history append: %v", err)
		}
		room.broadcastAll(out)
	}

	if left, ok := room.remove(conn); ok {
		room.broadcastAll(protocol.Message{Type: protocol.TypeLeave, From: left})
		room.broadcastUsers()
	}
}
