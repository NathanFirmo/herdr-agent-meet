# Agent Meet

A video call where your coding agents report to you. Every agent running in your Herdr session joins as a participant. When one writes something new, Agent Meet speaks it out loud and shows it as captions, like a teammate giving you a status update.

- Each agent gets a tile and a voice; Claude Code agents show the model they run on.
- See who is thinking or running a tool, and for how long, with a warning when an agent may be stuck.
- Chat with @mentions and images, delivered straight into an agent's session.
- Answer an agent's questions and approvals from a card in the call.
- Start new agents from the call with any shell command, rename them, or close them. New agents are told their replies are spoken, so they answer in short spoken prose.
- Reads Claude Code session transcripts, so whole replies are spoken even after they scroll off screen.
- Runs in Portuguese or English, or has agents reply in any other language, with `MEET_LANGUAGE`.

Source, configuration and docs: https://github.com/NathanFirmo/herdr-agent-meet

## Run

Requires a running Herdr session. The app needs Kokoro next to it for the voices and runs as your user, so it can open the Herdr socket.

```sh
docker network create agent-meet

docker run -d --name kokoro --network agent-meet --restart unless-stopped \
  ghcr.io/remsky/kokoro-fastapi-cpu:v0.2.4

mkdir -p /tmp/agent-meet
docker run -d --name agent-meet --network agent-meet --restart unless-stopped \
  --user "$(id -u):$(id -g)" \
  -p 127.0.0.1:8765:8765 \
  -e KOKORO_URL=http://kokoro:8880 \
  -e MEET_LANGUAGE=en-US \
  -e HOST_HOME="$HOME" \
  -e CLAUDE_TRANSCRIPT_DIRS="$HOME/.claude*" \
  -e UPLOAD_DIR=/tmp/agent-meet \
  -v "$HOME/.config/herdr:/run/herdr" \
  -v /tmp/agent-meet:/tmp/agent-meet \
  -v "$HOME:$HOME:ro" \
  nathanfirmo/herdr-agent-meet:latest
```

Open http://localhost:8765.

| Mount | Why |
|---|---|
| `~/.config/herdr` → `/run/herdr` | The Herdr socket |
| `/tmp/agent-meet` | Chat images and participant names, at the same path agents read them from |
| `$HOME` (read-only, same path) | Claude Code transcripts and folder completion |

`MEET_LANGUAGE` is a BCP 47 tag such as `pt-BR` (the default), `en-US` or `es-ES`; it sets the UI, voice dictation, summaries and voices. Set `SUMMARY_COMMAND` to summarize what agents say with any CLI, or `TTS=none` for captions only. All options are in the GitHub README.

Images are published for `linux/amd64` and `linux/arm64`, tagged by version, such as `0.2.0` and `0.2`, and `latest`.
