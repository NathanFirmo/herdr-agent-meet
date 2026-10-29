# Agent Meet

A video call where your coding agents report to you. Every agent running in your
Herdr session joins as a participant. When one writes something
new, Agent Meet summarizes it, speaks it out loud, and shows it as captions,
like a teammate giving you a status update.

- Each agent gets a tile and a voice; a blue border shows who is talking, and
  Claude Code agents show the model they run on.
- A spinner marks agents at work. A raised hand and a toast mark an agent
  waiting for your approval.
- Captions, a chat panel and a participant list.
- Chat with @mentions and images, delivered straight into an agent's session.
- Answer an agent's questions and approvals from a card in the call.
- Start a new agent from the call: it opens a Herdr workspace and runs any
  shell command, so aliases and flags work. Rename or close agents from the
  participant list.

## How it works

```
Herdr socket ──► Go backend ──► summarizer ──► TTS ──► SSE ──► React call UI
```

1. The backend polls Herdr through its socket API (`agent.list`, `agent.read`)
   and reads each agent's visible screen.
2. A screen parser extracts what the agent wrote. Text already on screen when
   the backend starts is skipped, so joining does not replay history.
3. New paragraphs go to the summarizer, then to the TTS engine.
4. The browser gets participants and speeches over Server-Sent Events and plays
   them one at a time.

### Ports and adapters

The call logic in `internal/meet` depends only on the interfaces in
`internal/meet/ports.go`. Adapters live in `internal/adapters` and are picked
by environment variables in `cmd/agent-meet/main.go`.

| Port | Adapters | Configured by |
|---|---|---|
| `AgentSource` | `herdr` (socket API) | `HERDR_SOCKET_PATH` |
| `ScreenParser` (one per agent kind) | `ClaudeCode` | registered in `main.go` |
| `Summarizer` | `command`, `none` | `SUMMARIZER`, `SUMMARY_COMMAND` |
| `Speaker` | `kokoro`, `none` | `TTS`, `KOKORO_URL`, `VOICES`, `SPEECH_SPEED` |

To add another summarizer or TTS engine, implement the interface in a new
adapter and add a case to the matching factory in `main.go`. To support another
agent kind, add a `ScreenParser` and register it under the kind Herdr reports
(`herdr agent list`).

## Requirements

- A running Herdr session
- Docker with Compose
- [Task](https://taskfile.dev)
- Go 1.26 and Node 24, only to run on the host

## Run with Docker

```sh
cp .env.example .env
task up
```

Open http://localhost:8080 and click **Participar agora**. Browsers only allow
audio after an interaction, so the call starts from that screen.

`task up` builds the app image and starts Kokoro-FastAPI next to it. The app
container runs as your user and mounts the Herdr config directory to reach
its socket.

```sh
task logs   # follow what the agents say
task down
```

## Run on the host

Runs Kokoro in Docker, the Go API on the host, and Vite with hot reload on
http://localhost:5173:

```sh
task dev
```

## Configuration

Set these in `.env`. Empty values use the defaults.

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Port published by Docker |
| `HERDR_DIR` | `~/.config/herdr` | Directory with `herdr.sock`, mounted in Docker |
| `HERDR_SOCKET_PATH` | `~/.config/herdr/herdr.sock` | Socket path when running on the host |
| `TTS` | `kokoro` | `kokoro`, or `none` for captions only |
| `KOKORO_URL` | `http://localhost:8880` | Kokoro-FastAPI URL (set by Compose in Docker) |
| `VOICES` | `pm_alex,pf_dora,pm_santa` | Kokoro voices, assigned to agents in turn |
| `SPEECH_SPEED` | `1.1` | Speech speed |
| `SUMMARIZER` | `command` when `SUMMARY_COMMAND` is set, else `none` | `command` or `none` (speaks the raw text) |
| `SUMMARY_COMMAND` | | Shell command that reads the prompt on stdin and prints the summary |
| `POLL_INTERVAL` | `3s` | How often Herdr is read |
| `HOST_HOME` | your home | Home used for `~` in folder completion; Compose sets it from `$HOME` |
| `UPLOAD_DIR` | `/tmp/agent-meet` | Where chat images are saved; mounted at the same path in Docker |

### Summary command

Any CLI works, as long as it reads the prompt on stdin and prints one sentence.
The command runs with `sh -c`, which does not load shell aliases: call the
binary directly, or wrap an alias with `zsh -ic '...'` when running on the host.

Example with Claude Code and Claude Haiku:

```sh
SUMMARY_COMMAND=claude -p --model haiku --setting-sources '' --strict-mcp-config --tools '' --no-session-persistence --disable-slash-commands
```

In Docker the command runs inside the app container, so the CLI must be
available there. Compose mounts `CLAUDE_BIN` at `/usr/local/bin/claude` and
`CLAUDE_CONFIG_DIR` at `/claude-config`; both fall back to `/dev/null` when
unset. To use Claude Code from the host, set in `.env`:

```sh
CLAUDE_BIN=/path/to/claude          # readlink -f "$(command -v claude)"
CLAUDE_CONFIG_DIR=/path/to/.claude  # account to bill
SUMMARY_COMMAND=CLAUDE_CONFIG_DIR=/claude-config claude -p --model haiku --setting-sources '' --strict-mcp-config --tools '' --no-session-persistence --disable-slash-commands
```

## Development

```sh
task check   # gofmt, go vet, oxlint and tsc
task build   # web bundle and Go binary in bin/
```

| Path | Contents |
|---|---|
| `cmd/agent-meet` | Entry point and adapter wiring |
| `internal/meet` | Call state, ports, and the polling loop |
| `internal/adapters` | Herdr, screen parsers, summarizers, and TTS engines |
| `internal/server` | HTTP API (SSE events and audio) and static files |
| `web` | React + TypeScript + Vite call UI |

## License

MIT
