package cli

import (
	"fmt"
	"io"
	"strings"
)

var Version = "0.1.0"

type ServerFlags struct {
	Set     *FlagSet
	Listen  *string
	Groups  *string
	History *string
	Limit   *int
	Cert    *string
	Key     *string
}

type ConnectFlags struct {
	Set      *FlagSet
	Addr     *string
	Name     *string
	Group    *string
	Avatar   *string
	TLS      *bool
	CA       *string
	Insecure *bool
}

func newServerFlags() *ServerFlags {
	fs := NewFlagSet("server")
	return &ServerFlags{
		Set:     fs,
		Listen:  fs.String("listen", ":9000", "address to listen on"),
		Groups:  fs.String("groups", "general,random", "comma-separated default groups"),
		History: fs.String("history", "relay.jsonl", "path to chat history file (empty to disable)"),
		Limit:   fs.Int("history-limit", 500, "max messages kept in memory and replayed to new clients"),
		Cert:    fs.String("tls-cert", "", "TLS certificate file"),
		Key:     fs.String("tls-key", "", "TLS private key file"),
	}
}

func newConnectFlags() *ConnectFlags {
	fs := NewFlagSet("connect")
	return &ConnectFlags{
		Set:      fs,
		Addr:     fs.String("addr", "localhost:9000", "relay server address"),
		Name:     fs.String("name", "", "display name (defaults to $USER)"),
		Group:    fs.String("group", "general", "initial group/channel"),
		Avatar:   fs.String("avatar", "", "ASCII avatar or preset (cat, bot, fox, star, wave...)"),
		TLS:      fs.Bool("tls", false, "use TLS when connecting"),
		CA:       fs.String("tls-ca", "", "custom CA bundle for TLS"),
		Insecure: fs.Bool("insecure", false, "skip TLS certificate verification"),
	}
}

func RootUsage(w io.Writer) {
	fmt.Fprintf(w, "relay %s - lightweight terminal chat relay\n\n", Version)
	fmt.Fprintf(w, "usage:\n")
	fmt.Fprintf(w, "  relay [--help] [--version]\n")
	fmt.Fprintf(w, "  relay server [--help] [flags]\n")
	fmt.Fprintf(w, "  relay connect [--help] [flags]\n")
	fmt.Fprintf(w, "  relay avatars [--ascii | --emoji]\n")
	fmt.Fprintf(w, "  relay version\n")
	fmt.Fprintf(w, "  relay help [command]\n\n")
	fmt.Fprintf(w, "Long flags use a double dash: --listen, --name, --help.\n")
	fmt.Fprintf(w, "Run `relay help <command>` for command-specific flags.\n")
}

func ServerUsage(w io.Writer) {
	fmt.Fprintf(w, "usage: relay server [--help] [flags]\n\n")
	fmt.Fprintf(w, "Start the relay chat server.\n\n")
	fmt.Fprintf(w, "flags:\n")
	for _, line := range newServerFlags().Set.FlagLines() {
		fmt.Fprintln(w, line)
	}
	fmt.Fprintf(w, "\nexamples:\n")
	fmt.Fprintf(w, "  relay server\n")
	fmt.Fprintf(w, "  relay server --listen :9000\n")
	fmt.Fprintf(w, "  relay server --groups general,random,dev\n")
	fmt.Fprintf(w, "  relay server --history=\"\" --tls-cert cert.pem --tls-key key.pem\n")
	fmt.Fprintln(w)
}

func ConnectUsage(w io.Writer) {
	fmt.Fprintf(w, "usage: relay connect [--help] [flags]\n\n")
	fmt.Fprintf(w, "Connect to a relay server with the terminal client.\n\n")
	fmt.Fprintf(w, "flags:\n")
	for _, line := range newConnectFlags().Set.FlagLines() {
		fmt.Fprintln(w, line)
	}
	fmt.Fprintf(w, "\nexamples:\n")
	fmt.Fprintf(w, "  relay connect --name alice --group general --avatar cat\n")
	fmt.Fprintf(w, "  relay connect --addr chat.example.com:9000 --tls\n")
	fmt.Fprintf(w, "  relay connect --tls --insecure\n")
	fmt.Fprintln(w)
}

func AvatarsUsage(w io.Writer) {
	fmt.Fprintf(w, "usage: relay avatars [--ascii | --emoji]\n\n")
	fmt.Fprintf(w, "List built-in avatar presets with previews.\n\n")
	fmt.Fprintf(w, "flags:\n")
	fmt.Fprintf(w, "  --ascii   list ASCII presets only\n")
	fmt.Fprintf(w, "  --emoji   list emoji presets only\n")
	fmt.Fprintf(w, "\nexamples:\n")
	fmt.Fprintf(w, "  relay avatars\n")
	fmt.Fprintf(w, "  relay avatars --ascii\n")
	fmt.Fprintln(w)
}

func HelpUsage(w io.Writer, topic string) {
	topic = strings.ToLower(strings.TrimSpace(topic))
	switch topic {
	case "", "help":
		RootUsage(w)
	case "server":
		ServerUsage(w)
	case "connect", "client":
		ConnectUsage(w)
	case "avatars", "avatar":
		AvatarsUsage(w)
	case "version":
		fmt.Fprintf(w, "usage: relay version\n\nPrint relay version.\n\n")
		fmt.Fprintf(w, "aliases: relay --version, relay -V\n\n")
	default:
		fmt.Fprintf(w, "unknown help topic %q\n\n", topic)
		RootUsage(w)
	}
}

func InChatCommands(w io.Writer) {
	fmt.Fprintf(w, "in-chat commands:\n")
	fmt.Fprintf(w, "  /group /groups /create /avatar /help\n")
}
