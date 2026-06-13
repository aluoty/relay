package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
)

type room struct {
	mu      sync.Mutex
	clients map[net.Conn]string
}

func newRoom() *room {
	return &room{clients: make(map[net.Conn]string)}
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

func (r *room) broadcast(msg Message, skip net.Conn) {
	data, err := jsonMarshalLine(msg)
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

func runServer(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	log.Printf("relay listening on %s", addr)

	room := newRoom()
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
	join, err := readMessage(reader)
	if err != nil || join.Type != msgJoin || join.From == "" {
		writeMessage(conn, Message{Type: msgSys, Text: "expected join message"})
		return
	}

	name := join.From
	room.add(conn, name)
	writeMessage(conn, Message{Type: msgSys, Text: fmt.Sprintf("connected as %s", name)})
	room.broadcast(Message{Type: msgJoin, From: name}, conn)

	for {
		msg, err := readMessage(reader)
		if err != nil {
			break
		}
		if msg.Type != msgChat || msg.Text == "" {
			continue
		}
		out := Message{Type: msgChat, From: name, Text: msg.Text}
		room.broadcast(out, nil)
	}

	if left, ok := room.remove(conn); ok {
		room.broadcast(Message{Type: msgLeave, From: left}, nil)
	}
}

func jsonMarshalLine(msg Message) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
