package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/aluoty/relay/internal/avatar"
	"github.com/aluoty/relay/internal/client"
	"github.com/aluoty/relay/internal/server"
	"github.com/aluoty/relay/internal/tlsconfig"
)

func Run(args []string) int {
	if len(args) == 0 {
		RootUsage(os.Stderr)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		return runHelp(args[1:])
	case "-V", "--version", "version":
		printVersion(os.Stdout)
		return 0
	case "server":
		return runServer(args[1:])
	case "connect":
		return runConnect(args[1:])
	case "avatars", "avatar":
		return runAvatars(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		RootUsage(os.Stderr)
		return 2
	}
}

func runHelp(args []string) int {
	topic := ""
	if len(args) > 0 {
		topic = args[0]
	}
	HelpUsage(os.Stderr, topic)
	if topic == "" {
		InChatCommands(os.Stderr)
		fmt.Fprintln(os.Stderr)
	}
	return 0
}

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "relay %s\n", Version)
}

func runServer(args []string) int {
	flags := newServerFlags()
	flags.Set.Usage = func() { ServerUsage(os.Stderr) }
	if _, err := flags.Set.Parse(args); err != nil {
		if err == ErrHelp {
			return 0
		}
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		return 2
	}

	if err := server.Run(server.Config{
		Listen:  *flags.Listen,
		Groups:  splitCSV(*flags.Groups),
		History: *flags.History,
		Limit:   *flags.Limit,
		TLS: tlsconfig.ServerConfig{
			CertFile: *flags.Cert,
			KeyFile:  *flags.Key,
		},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		return 1
	}
	return 0
}

func runConnect(args []string) int {
	flags := newConnectFlags()
	flags.Set.Usage = func() { ConnectUsage(os.Stderr) }
	if _, err := flags.Set.Parse(args); err != nil {
		if err == ErrHelp {
			return 0
		}
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		return 2
	}

	n := *flags.Name
	if n == "" {
		n = client.DefaultName()
	}

	if err := client.Run(client.Config{
		Addr:   *flags.Addr,
		Name:   n,
		Group:  *flags.Group,
		Avatar: *flags.Avatar,
		TLS: tlsconfig.ClientConfig{
			Enabled:  *flags.TLS,
			CAFile:   *flags.CA,
			Insecure: *flags.Insecure,
		},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		return 1
	}
	return 0
}

func runAvatars(args []string) int {
	showASCII := false
	showEmoji := false
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			AvatarsUsage(os.Stderr)
			return 0
		case "--ascii":
			showASCII = true
		case "--emoji":
			showEmoji = true
		default:
			fmt.Fprintf(os.Stderr, "avatars: unknown argument %q\n\n", arg)
			AvatarsUsage(os.Stderr)
			return 2
		}
	}
	if !showASCII && !showEmoji {
		showASCII = true
		showEmoji = true
	}

	if showASCII {
		printPresetSection(os.Stdout, "ASCII presets", avatar.ASCIIPresets, asciiPreview)
	}
	if showASCII && showEmoji {
		fmt.Fprintln(os.Stdout)
	}
	if showEmoji {
		printPresetSection(os.Stdout, "Emoji presets", avatar.EmojiPresets, emojiPreview)
	}
	return 0
}

func printPresetSection(w io.Writer, title string, presets map[string]string, preview func(string) string) {
	names := make([]string, 0, len(presets))
	for name := range presets {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Fprintf(w, "%s:\n", title)
	for _, name := range names {
		fmt.Fprintf(w, "  %-12s %s\n", name, preview(presets[name]))
	}
}

func asciiPreview(raw string) string {
	raw = strings.ReplaceAll(raw, "\n", `\n`)
	if len(raw) > 40 {
		return raw[:37] + "..."
	}
	return raw
}

func emojiPreview(alias string) string {
	return avatar.ParseText(alias)
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
