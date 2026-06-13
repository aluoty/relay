# relay

Lightweight terminal chat relay built with Go and [tview](https://github.com/rivo/tview).

One binary runs as a TCP relay server or as a terminal client. Chat is organized into groups (like Discord channels), users get ASCII avatars, messages persist as JSONL, and TLS is optional.

Best used when placed in binary paths, to check:
```bash
echo $PATH
```
Later after compiling, place the binary into one of those paths, so that 
```bash
relay
```
can be used, otherwise be in the relay/ directory and use
```bash
./relay
```

## Requirements

- Go 1.21+
- A terminal with UTF-8 support

## Install

```bash
go install github.com/aluoty/relay/cmd/relay@latest
```

Or build from source:

```bash
go build -o relay ./cmd/relay
```

## Quick start

**Terminal 1 — start the server**

```bash
relay server
```

**Terminal 2 & 3 — connect clients**

```bash
relay connect -name alice -group general -avatar cat
relay connect -name bob -group random -avatar bot
```

Type a message and press Enter. Press `Tab` to move between the group list, chat, and user list. Press `Ctrl+C` to quit.

## Commands

```text
relay server [flags]
relay connect [flags]
```

### `relay server`

| Flag | Default | Description |
|------|---------|-------------|
| `-listen` | `:9000` | Address to listen on |
| `-groups` | `general,random` | Default channels (comma-separated) |
| `-history` | `relay.jsonl` | Chat history file (JSONL). Set to empty to disable |
| `-history-limit` | `500` | Max messages stored and replayed per group |
| `-tls-cert` | | TLS certificate file |
| `-tls-key` | | TLS private key file |

Examples:

```bash
relay server -listen :9000
relay server -groups general,random,dev
relay server -history ""                      # no persistence
```

### `relay connect`

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `localhost:9000` | Server address |
| `-name` | `$USER` | Display name |
| `-group` | `general` | Initial channel |
| `-avatar` | | ASCII avatar text or preset name |
| `-tls` | `false` | Use TLS |
| `-tls-ca` | | Custom CA bundle (PEM) |
| `-insecure` | `false` | Skip TLS verification (development only) |

Examples:

```bash
relay connect -name alice -group general -avatar cat
relay connect -name bob -group random -avatar "=^..^="
relay connect -addr chat.example.com:9000 -tls
```

## In-chat commands

| Command | Description |
|---------|-------------|
| `/group <name>` | Switch channel (alias: `/g`, `/channel`) |
| `/groups` | List available channels |
| `/create <name>` | Create a new channel |
| `/avatar <text>` | Set avatar — ASCII preset, emoji alias, or custom text |
| `/help` | Show command help |

### Switching groups

| Action | How |
|--------|-----|
| Command | `/group random` or `/g random` |
| Focus list | `Ctrl+G` — press again to return to chat |
| Sidebar | `Enter` on a group to switch (returns to chat) |
| Quick keys | `1`–`9` while the group list is focused |
| Leave sidebar | `Esc` — always returns to message input |

### Avatars (ASCII + emoji)

Set a custom ASCII avatar:

```text
/avatar *_*
/avatar ^_^
/avatar >:)
```

Use an ASCII preset:

```text
/avatar cat
/avatar star_eyes
/avatar awkward
```

Use an emoji preset (via [enescakir/emoji](https://github.com/enescakir/emoji)):

```text
/avatar smile
/avatar party
/avatar :wave:
/avatar :cat:
```

ASCII presets include: `cat`, `bot`, `star_eyes`, `happy`, `sad`, `wink`, `awkward`, `shrug`, `surprised`, `angry`, `bear`, `robot`, and more.

Emoji presets include: `smile`, `grin`, `party`, `thumbsup`, `wave_e`, `heart_e`, `cat_e`, `rocket`, and more — or any `:alias:` supported by the emoji library.

Avatars appear next to your name in chat and in the user list (up to 3 lines, 16 columns wide).

Multi-line ASCII avatars use `\n`:

```text
/avatar /\\_/\\\n( o.o )
```

### Chat emojis

Type `:alias:` shorthand in messages — rendered with [enescakir/emoji](https://github.com/enescakir/emoji):

```text
hello :wave: good job :+1: :tada:
```

## UI layout

```text
┌ Groups ──┐ ┌ Relay Chat ──────────────┐ ┌ Users ───┐
│ # general│ │ =^..^= alice: hello       │ │ =^..^= alice
│ # random │ │ [o_o] bob: hi             │ │ [o_o] bob
└──────────┘ └────────────────────────────┘ └──────────┘
               connected to localhost:9000 as alice in #general  | Ctrl+G groups | Esc chat
               Message: _
```

## TLS

Generate a self-signed certificate for local development:

```bash
openssl req -x509 -newkey rsa:2048 \
  -keyout key.pem -out cert.pem -days 365 -nodes \
  -subj "/CN=localhost"
```

Run the server with TLS:

```bash
relay server -tls-cert cert.pem -tls-key key.pem
```

Connect with TLS:

```bash
relay connect -tls -insecure          # self-signed cert
relay connect -tls -tls-ca ca.pem     # custom CA
```

## Project layout

```text
cmd/relay/          CLI entrypoint
internal/
  avatar/           ASCII + emoji avatars and text parsing
  client/           tview terminal client
  commands/         slash-command parsing
  protocol/         wire format and message helpers
  server/           TCP relay hub (groups, sessions)
  store/            JSONL chat history (per group)
  tlsconfig/        optional TLS helpers
```

## Protocol

Clients send one JSON object per line over TCP:

| Type | Direction | Purpose |
|------|-----------|---------|
| `join` | client → server | Connect with name, group, optional avatar |
| `msg` | both | Chat message in a group |
| `switch` | client → server | Change active group |
| `create` | client → server | Create a new group |
| `avatar` | client → server | Update ASCII avatar |
| `users` | server → client | Online users in a group (with avatars) |
| `groups` | server → client | Available groups |
| `leave` | server → client | User left a group |
| `sys` | server → client | System notice |

Example session:

```json
{"t":"join","f":"alice","g":"general","a":"cat"}
{"t":"groups","gs":["general","random"]}
{"t":"msg","f":"alice","g":"general","x":"hello","a":"cat"}
{"t":"switch","g":"random"}
{"t":"users","g":"random","u":["alice"],"p":{"alice":"=^..^="}}
```

## Persistence

When `-history` is set (default: `relay.jsonl`), chat messages are appended with their group id and replayed when a client joins or switches to that group.

Join/leave events, user lists, and group lists are not persisted.

## License

See [LICENSE](LICENSE)
