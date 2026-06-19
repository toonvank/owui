# owui

A fast terminal UI for [Open WebUI](https://github.com/open-webui/open-webui), written in Go.

One static binary — full-screen chat in your shell with streaming, markdown, RAG, and model switching. No browser tab, no Electron, no `npm install`.

> **Not a server.** owui connects to your existing Open WebUI instance via its API. It does not replace Open WebUI or Open CoreUI.

## Features

- **Interactive TUI** — full-screen chat with in-place streaming and markdown rendering
- **First-run setup** — `owui setup` wizard for server URL and API key
- **Model picker** — `/model` with provider tags, capabilities, and fuzzy filter
- **Server chats** — `/chats` to browse, resume, and pin conversations; auto-title after first exchange
- **Local sessions** — auto-saved history in `~/.config/owui/sessions/`; `/sessions` picker
- **RAG** — attach files (`/file upload`) and knowledge collections (`/knowledge`)
- **Filters & tools** — interactive toggles via `/filters` and `/tools`
- **Multi-profile** — named server profiles in config; `/profile` hot-switch
- **Slash commands** — export, copy, regen, search, fork, and more
- **Scriptable** — `owui chat "hello"`, stdin pipes, `--json`, `--resume`
- **Diagnostics** — `owui doctor` checks config and connectivity

## Requirements

- Go 1.22+ (to build from source)
- A running Open WebUI server you can reach over HTTPS/HTTP
- An API key from **Settings → Account** in Open WebUI (or email/password for `owui auth login`)

## Install

**Pre-built binaries** (recommended) — [GitHub Releases](https://github.com/toonvank/owui/releases/latest):

```bash
# Linux amd64 example (v0.6.0)
curl -fsSL https://github.com/toonvank/owui/releases/download/v0.6.0/owui-v0.6.0-linux-amd64 -o owui
chmod +x owui
sudo mv owui /usr/local/bin/
```

Also available: `linux-arm64`, `darwin-amd64`, `darwin-arm64`, `windows-amd64.exe`, plus `SHA256SUMS`.

**From source:**

```bash
git clone https://github.com/toonvank/owui.git
cd owui
make install   # installs to ~/.local/bin/owui
```

Ensure `~/.local/bin` (or `/usr/local/bin`) is on your `PATH`.

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

Resume a server chat:

```bash
owui --resume <chat-id-prefix>
owui chat --resume abc123 "continue where we left off"
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
| `↑` `↓` | Navigate pickers / autocomplete / input history |
| `j` `k` | Scroll viewport (when `vim_keys: true`) |
| `/model` | Switch model (interactive picker) |
| `/chats` | Browse and resume server chats (`p` to pin) |
| `/sessions` | Browse local sessions |
| `/knowledge` | Pick RAG knowledge collection |
| `/filters` `/tools` | Toggle integrations interactively |
| `/profile` | Switch config profile |
| `?` | Toggle help |
| `C` | Collapse blocks (empty input only) |
| `Ctrl+C` | Quit |

### One-shot chat

```bash
owui chat "Explain this error: exit code 137"
owui chat --model llama3.2 "Summarize in 3 bullets"
echo "log contents" | owui chat
owui chat --json "ping"    # machine-readable output
```

### Other commands

```bash
owui models list
owui models info <id>
owui chats list
owui chats pin <id>
owui files upload ./doc.pdf
owui pull llama3.2 --progress
owui config profile list
owui doctor
owui auth status
owui status
owui completion fish > ~/.config/fish/completions/owui.fish
```

## Configuration

Config is stored locally at:

```
~/.config/owui/config.yaml
```

Example:

```yaml
base_url: http://localhost:3000
api_key: sk-your-key-here
default_model: llama3.2
stream: true
theme: dark          # dark | light | none
vim_keys: false
```

Named profiles let you keep multiple servers in one config file:

```bash
owui config profile add work --url https://work.example.com --api-key sk-...
owui config profile use work
owui --profile work
```

Sessions are stored per profile under `~/.config/owui/sessions/<profile>/`.

### Environment variables

| Variable | Description |
|----------|-------------|
| `OWUI_BASE_URL` | Server URL (overrides config) |
| `OWUI_API_KEY` | API key (overrides config) |
| `OWUI_MODEL` | Default model (overrides config) |
| `OWUI_PROFILE` | Active config profile |
| `NO_COLOR` | Disable ANSI colors |

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