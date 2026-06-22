# Split user config and runtime state into two separate files

User-authored settings live in `config.yml`; runtime state lives in `state.json`. The TUI writes `state.json` freely (on every resize, navigation, or preference toggle) but never rewrites `config.yml` after the first-run wizard.

A single file would require either (a) the TUI rewriting YAML the user has hand-edited — destroying comments and formatting — or (b) blocking the TUI from persisting state across sessions. Splitting by ownership avoids both problems: the user controls their YAML; the TUI owns the JSON.

**Consequences**
- `config.yml`: profiles, sections, columns, keybindings, display defaults. Stable. User-controlled.
- `state.json`: active profile, default project per profile, team ID, sidebar state, preview pane width. Volatile. TUI-controlled. Document clearly that hand-editing it is unsupported.
