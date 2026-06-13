package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aluoty/relay/internal/client"
	"github.com/aluoty/relay/internal/server"
	"github.com/aluoty/relay/internal/tlsconfig"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "server":
		os.Exit(runServer(os.Args[2:]))
	case "connect":
		os.Exit(runConnect(os.Args[2:]))
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func runServer(args []string) int {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	listen := fs.String("listen", ":9000", "address to listen on")
	groups := fs.String("groups", "general,random", "comma-separated default groups")
	history := fs.String("history", "relay.jsonl", "path to chat history file (empty to disable)")
	limit := fs.Int("history-limit", 500, "max messages kept in memory and replayed to new clients")
	cert := fs.String("tls-cert", "", "TLS certificate file")
	key := fs.String("tls-key", "", "TLS private key file")
	_ = fs.Parse(args)

	if err := server.Run(server.Config{
		Listen:  *listen,
		Groups:  splitCSV(*groups),
		History: *history,
		Limit:   *limit,
		TLS: tlsconfig.ServerConfig{
			CertFile: *cert,
			KeyFile:  *key,
		},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		return 1
	}
	return 0
}

func runConnect(args []string) int {
	fs := flag.NewFlagSet("connect", flag.ExitOnError)
	addr := fs.String("addr", "localhost:9000", "relay server address")
	name := fs.String("name", "", "display name (defaults to $USER)")
	group := fs.String("group", "general", "initial group/channel")
	avatar := fs.String("avatar", "", "ASCII avatar or preset (cat, bot, fox, star, wave...)")
	useTLS := fs.Bool("tls", false, "use TLS when connecting")
	ca := fs.String("tls-ca", "", "custom CA bundle for TLS")
	insecure := fs.Bool("insecure", false, "skip TLS certificate verification")
	_ = fs.Parse(args)

	n := *name
	if n == "" {
		n = client.DefaultName()
	}

	if err := client.Run(client.Config{
		Addr:   *addr,
		Name:   n,
		Group:  *group,
		Avatar: *avatar,
		TLS: tlsconfig.ClientConfig{
			Enabled:  *useTLS,
			CAFile:   *ca,
			Insecure: *insecure,
		},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		return 1
	}
	return 0
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func usage() {
	fmt.Fprintf(os.Stderr, "relay - lightweight terminal chat relay\n\n")
	fmt.Fprintf(os.Stderr, "usage:\n")
	fmt.Fprintf(os.Stderr, "  relay server [flags]\n")
	fmt.Fprintf(os.Stderr, "  relay connect [flags]\n\n")
	fmt.Fprintf(os.Stderr, "server flags:\n")
	fmt.Fprintf(os.Stderr, "  -listen         listen address (default :9000)\n")
	fmt.Fprintf(os.Stderr, "  -groups         default channels (default general,random)\n")
	fmt.Fprintf(os.Stderr, "  -history        JSONL history file (default relay.jsonl, empty disables)\n")
	fmt.Fprintf(os.Stderr, "  -history-limit  messages to retain/replay (default 500)\n")
	fmt.Fprintf(os.Stderr, "  -tls-cert       TLS certificate file\n")
	fmt.Fprintf(os.Stderr, "  -tls-key        TLS private key file\n\n")
	fmt.Fprintf(os.Stderr, "connect flags:\n")
	fmt.Fprintf(os.Stderr, "  -addr           server address (default localhost:9000)\n")
	fmt.Fprintf(os.Stderr, "  -name           display name (default $USER)\n")
	fmt.Fprintf(os.Stderr, "  -group          initial channel (default general)\n")
	fmt.Fprintf(os.Stderr, "  -avatar         ASCII avatar or preset name\n")
	fmt.Fprintf(os.Stderr, "  -tls            enable TLS\n")
	fmt.Fprintf(os.Stderr, "  -tls-ca         custom CA bundle\n")
	fmt.Fprintf(os.Stderr, "  -insecure       skip TLS verification (dev only)\n\n")
	fmt.Fprintf(os.Stderr, "in-chat commands:\n")
	fmt.Fprintf(os.Stderr, "  /group /groups /create /avatar /help\n")
}
