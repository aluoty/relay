package client

import (
	"fmt"
	"os"

	"github.com/aluoty/relay/internal/ascii"
	"github.com/aluoty/relay/internal/protocol"
	"github.com/aluoty/relay/internal/tlsconfig"
)

type Config struct {
	Addr    string
	Name    string
	Group   string
	Avatar  string
	TLS     tlsconfig.ClientConfig
}

func Run(cfg Config) error {
	conn, err := cfg.TLS.Dial(cfg.Addr)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	avatar, err := ascii.Resolve(cfg.Avatar)
	if err != nil && cfg.Avatar != "" {
		return err
	}

	group := cfg.Group
	if group == "" {
		group = protocol.DefaultGroup
	}

	if err := protocol.Write(conn, protocol.Join(cfg.Name, group, avatar)); err != nil {
		return fmt.Errorf("join: %w", err)
	}

	ui := newUI(cfg.Addr, cfg.Name, group, avatar)
	return ui.run(conn)
}

func DefaultName() string {
	if user := os.Getenv("USER"); user != "" {
		return user
	}
	if user := os.Getenv("LOGNAME"); user != "" {
		return user
	}
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return "guest"
}
