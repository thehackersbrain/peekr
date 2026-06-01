# peekr

A focused network monitor for your terminal — established connections, listening ports,
and per-interface throughput. Built with Go + Bubble Tea. No process manager, no system
info — use htop/btop and fastfetch for those.

## Features

- **Established connections** — local/remote addr, PID, process name, user
- **Listening ports** — same columns, same search
- **Both (split view)** — established on top, listening below, tab to switch pane
- **Network stats** — per-interface IP, total sent/recv, live TX/RX rates, packet counts
- **15 built-in themes** — all compiled in, no runtime files needed
- **Live theme picker** — color swatches + fake UI preview as you navigate
- **Search/filter** on connections (`s`)
- **Hide inactive interfaces** (`h`)
- **vim keys** everywhere (`j`/`k`)

## Themes

| Name              | Vibe           |
| ----------------- | -------------- |
| tokyo-night ★     | default        |
| nord              | arctic blue    |
| dracula           | classic purple |
| gruvbox           | warm retro     |
| catppuccin-frappe | soft pastel    |
| everforest-dark   | forest green   |
| everforest-light  | light forest   |
| kanagawa          | japanese ink   |
| moonlight         | deep blue      |
| monokai-pro       | vivid contrast |
| nightfox          | dark ocean     |
| oxocarbon         | IBM carbon     |
| zenbones-dark     | minimal dark   |
| zenbones-light    | minimal light  |
| cthulhain         | eldritch teal  |

## Install

```bash
go install -v github.com/thehackersbrain/peekr@latest
```

## Build

```sh
git clone <this repo>
cd peekr
go mod tidy
go build -o peekr ./cmd/peekr

# Install (optional)
sudo mv peekr /usr/local/bin/
```

Requires Go 1.22+. On Linux, `net_connections` needs root or `CAP_NET_ADMIN` to see all
connections. Run with `sudo peekr` or grant the capability:

```sh
sudo setcap cap_net_admin+eip /usr/local/bin/peekr
```

## Usage

```sh
peekr                   # tokyo-night theme, 3s refresh
peekr -t nord           # nord theme
peekr -t dracula -d 1   # dracula, 1 second refresh
peekr -v                # version
```

## Keys

| Key            | Action                          |
| -------------- | ------------------------------- |
| `↑/↓` or `j/k` | Navigate                        |
| `enter`        | Select / confirm                |
| `s`            | Search/filter connections       |
| `tab`          | Switch pane (Both view)         |
| `h`            | Toggle hide inactive interfaces |
| `t`            | Open theme picker               |
| `backspace/←`  | Go back                         |
| `q` / `Esc`    | Quit / go back                  |

## Project layout

```
peekr/
  cmd/peekr/main.go             entrypoint, flags
  internal/
    theme/theme.go              all 15 themes as Go structs + lipgloss Styles
    collector/collector.go      gopsutil wrappers for connections + iface stats
    ui/
      root.go                   root Bubble Tea model, screen routing, tick loop
      menu.go                   main menu with live connection counts
      connections.go            established / listening / both split view
      netstats.go               per-interface stats table
      themepicker.go            theme list + live swatch preview
```
