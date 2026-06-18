# owui

Terminal client for [Open WebUI](https://github.com/open-webui/open-webui). Chat from your shell with a full-screen TUI — streaming, markdown, model switching, and chat resume.

> **Not a server.** owui connects to your existing Open WebUI instance via its API. It does not replace Open WebUI or Open CoreUI.

## Features

- **Interactive TUI** — full-screen chat with in-place streaming and markdown rendering
- **First-run setup** — `owui setup` wizard for server URL and API key
- **Model picker** — `/model` with live fuzzy filter and arrow-key navigation
- **Chat resume** — `/chats` or `/resume` to browse and restore server conversations
- **Local sessions** — auto-saved history in `~/.config/owui/sessions/`
- **Slash commands** — models, filters, streaming toggle, and more
- **Scriptable** — one-shot `owui chat "hello"` for pipelines and automation

## Requirements

- Go 1.22+ (to build from source)
- A running Open WebUI server you can reach over HTTPS/HTTP
- An API key from **Settings → Account** in Open WebUI (or email/password for `owui auth login`)

## Install

```bash
git clone https://github.com/toonvank/owui.git
cd owui
make install   # installs to ~/.local/bin/owui
```

Ensure `~/.local/bin` is on your `PATH`.

## Quick start

Run `owui` with no config — a setup wizard starts automatically:

```bash
owui
```

Or configure explicitly:

```bash
owui setup
owui          # start chatting
```

### Change server later

```bash
owui config set url https://your-server.example.com
owui setup    # full reconfigure (URL, auth, model)
```

Inside the TUI, `/server` shows the current endpoint and setup hints.

## Usage

### Interactive mode (default)

```bash
owui
```

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `↑` `↓` | Navigate pickers / autocomplete |
| `/model` | Switch model (interactive picker) |
| `/chats` | Browse and resume server chats |
| `?` | Toggle help |
| `C` | Collapse blocks (empty input only) |
| `Ctrl+C` | Quit |

### One-shot chat

```bash
owui chat "Explain this error: exit code 137"
owui chat --model llama3.2 "Summarize in 3 bullets"
```

### Other commands

```bash
owui models list
owui chats list
owui auth status
owui status
owui completion fish > ~/.config/fish/completions/owui.fish
```

## Configuration

Config is stored locally at:

```
~/.config/owui/config.yaml
```

Example (use your own values — **never commit this file**):

```yaml
base_url: http://localhost:3000
api_key: sk-your-key-here
default_model: llama3.2
stream: true
```

### Environment variables

| Variable | Description |
|----------|-------------|
| `OWUI_BASE_URL` | Server URL (overrides config) |
| `OWUI_API_KEY` | API key (overrides config) |
| `OWUI_MODEL` | Default model (overrides config) |

## Security

- **API keys and passwords stay on your machine** in `~/.config/owui/` — they are not part of this repository.
- Do not commit `config.yaml`, `.env`, or session files.
- Prefer API keys over password login for day-to-day use.

## Development

```bash
git clone https://github.com/toonvank/owui.git
cd owui
go test ./...
make build
./bin/owui
```

## Related projects

| Project | What it is |
|---------|------------|
| [Open WebUI](https://github.com/open-webui/open-webui) | The web UI and API server owui connects to |
| [Open CoreUI](https://github.com/xxnuo/open-coreui) | Community Rust lite fork of the server (separate project) |

## License

MIT