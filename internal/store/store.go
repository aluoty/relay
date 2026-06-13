package store

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"

	"github.com/aluoty/relay/internal/protocol"
)

type Store struct {
	path  string
	limit int
	mu    sync.RWMutex
	chat  []protocol.Message
}

func Open(path string, limit int) (*Store, error) {
	if limit <= 0 {
		limit = 500
	}
	s := &Store{path: path, limit: limit}
	if path == "" {
		return s, nil
	}
	return s, s.load()
}

func (s *Store) load() error {
	f, err := os.Open(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var msg protocol.Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}
		if msg.Type != protocol.TypeChat {
			continue
		}
		if msg.Group == "" {
			msg.Group = protocol.DefaultGroup
		}
		s.chat = append(s.chat, msg)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	s.trim()
	return nil
}

func (s *Store) History(group string) []protocol.Message {
	if group == "" {
		group = protocol.DefaultGroup
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]protocol.Message, 0)
	for _, msg := range s.chat {
		g := msg.Group
		if g == "" {
			g = protocol.DefaultGroup
		}
		if g == group {
			out = append(out, msg)
		}
	}
	return out
}

func (s *Store) Append(msg protocol.Message) error {
	if msg.Type != protocol.TypeChat {
		return nil
	}
	if msg.Group == "" {
		msg.Group = protocol.DefaultGroup
	}

	s.mu.Lock()
	s.chat = append(s.chat, msg)
	s.trim()
	s.mu.Unlock()

	if s.path == "" {
		return nil
	}

	data, err := protocol.MarshalLine(msg)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func (s *Store) trim() {
	if len(s.chat) <= s.limit {
		return
	}
	s.chat = s.chat[len(s.chat)-s.limit:]
}
