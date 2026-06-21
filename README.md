# ado-dash

A terminal dashboard for Azure DevOps, inspired by [`gh-dash`](https://github.com/dlvhdr/gh-dash).

Built in Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea).

> **Status:** Active development. See [PLAN.md](./PLAN.md) for the full implementation plan and [GitHub Issues](https://github.com/ujjwalgoyal19/ado-dash/issues) for progress.

## Planned features

- Pull requests — configurable sections, multi-repo fan-out, vote actions, policy status
- Work items — WIQL-backed sections, current-sprint auto-detection
- Pipelines — Build API, stage breakdown, queue/cancel
- Releases — pending approval gates, one-key approve/reject
- Vim keybindings, zen mode, collapsible sidebar, resizable preview pane
- Themes — Catppuccin, Dracula, Tokyo Night, Nord, Gruvbox, and more
- Auth via Azure CLI, PAT (env var reference), or Service Principal
- Config via YAML + TUI settings page (no mandatory hand-editing)

## Install

Coming soon via Homebrew and npm. See [Issue #15](https://github.com/ujjwalgoyal19/ado-dash/issues/15).

```bash
# future
brew install ujjwalgoyal19/tap/ado-dash
```

## Build from source

Requires Go 1.23+.

```bash
git clone https://github.com/ujjwalgoyal19/ado-dash
cd ado-dash
go build ./cmd/ado-dash
./ado-dash --help
```

## Config

`ado-dash` looks for config in order:

1. `--config /path/to/config.yml`
2. `~/.config/ado-dash/config.yml`

The first-run wizard creates the config file. See [PLAN.md](./PLAN.md) for the full schema.

## Development

See [PLAN.md](./PLAN.md) for architecture, API details, and build phases.
See [DESIGN_SPEC.md](./DESIGN_SPEC.md) for the full TUI design specification.
