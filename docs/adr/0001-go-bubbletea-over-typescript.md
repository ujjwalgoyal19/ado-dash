# Go + Bubble Tea over TypeScript/Bun

We chose Go with Bubble Tea (and Lip Gloss, Bubbles, Glamour) over the TypeScript/Bun prototype that was already started. The prototype is archived in `_archive/` for reference.

The reason is ecosystem fit: gh-dash, lazygit, and k9s — the tools ado-dash is modelled after — are all Go + Bubble Tea. Bubble Tea owns the render loop, alternate screen, and resize events, eliminating manual ANSI cursor math. `go build` produces a fully static binary, and GoReleaser handles multi-arch builds and the Homebrew tap automatically. None of these advantages exist in the Bun ecosystem at the same maturity level.

**Considered Options**
- TypeScript/Bun with a terminal library (existing prototype)
- Go + Bubble Tea (chosen)

**Consequences**
- The existing TypeScript code is superseded. Contributors must know Go.
- The Charm ecosystem (Lip Gloss, Bubbles, Glamour) provides the full component library — no need to build list, viewport, textinput, or spinner components from scratch.
