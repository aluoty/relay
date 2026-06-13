# relay

Lightweight terminal chat relay built with Go and [tview](https://github.com/rivo/tview).

One binary runs as a TCP relay server or as a terminal client. Messages are broadcast to all connected clients, online users are shown in a sidebar, history is persisted as JSONL, and TLS is optional.

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
relay connect -name alice
relay connect -name bob
```

Type a message and press Enter. Press `Ctrl+C` to quit the client.

## Commands

```text
relay server [flags]
relay connect [flags]
```

### `relay server`

| Flag | Default | Description |
|------|---------|-------------|
| `-listen` | `:9000` | Address to listen on |
| `-history` | `relay.jsonl` | Chat history file (JSONL). Set to empty to disable |
| `-history-limit` | `500` | Max messages stored and replayed to new clients |
| `-tls-cert` | | TLS certificate file |
| `-tls-key` | | TLS private key file |

Examples:

```bash
relay server -listen :9000
relay server -history ""                      # no persistence
relay server -history /var/lib/relay/chat.jsonl -history-limit 1000
```

### `relay connect`

| Flag | Default | Description |
|------|---------|-------------|
| `-addr` | `localhost:9000` | Server address |
| `-name` | `$USER` | Display name in chat |
| `-tls` | `false` | Use TLS |
| `-tls-ca` | | Custom CA bundle (PEM) |
| `-insecure` | `false` | Skip TLS verification (development only) |

Examples:

```bash
relay connect -addr localhost:9000 -name alice
relay connect -addr chat.example.com:9000 -tls
relay connect -addr localhost:9000 -tls -insecure
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
  client/           tview terminal client
  protocol/         wire format (newline-delimited JSON)
  server/           TCP relay hub
  store/            JSONL chat history
  tlsconfig/        optional TLS helpers
```

## Protocol

Clients send one JSON object per line over TCP:

| Type | Purpose |
|------|---------|
| `join` | First message after connect; sets nickname |
| `msg` | Chat message |
| `leave` | User disconnected (server → clients) |
| `users` | Current online user list (server → clients) |
| `sys` | System notice (server → clients) |

Example:

```json
{"t":"join","f":"alice"}
{"t":"msg","f":"alice","x":"hello"}
{"t":"users","u":["alice","bob"]}
```

## Persistence

When `-history` is set (default: `relay.jsonl`), chat messages are appended to a JSONL file and replayed to newly connected clients (up to `-history-limit`).

Join/leave events and the user list are not persisted — only chat messages.

## License

MIT
