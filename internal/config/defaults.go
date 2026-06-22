package config

// DefaultConfig returns a Config populated with sensible defaults.
// Applied before user YAML is merged so every field has a value.
func DefaultConfig() Config {
	return Config{
		Defaults: Defaults{
			RefreshSeconds: 120,
			Limit:          30,
			Theme:          "catppuccin-mocha",
			PreviewPane:    "right",
			OpenPRIn:       "detail",
			Notifications:  "off",
		},
	}
}

// DefaultState returns a zero-value State with UI defaults applied.
func DefaultState() State {
	return State{
		Profiles: make(map[string]ProfileState),
		UI: UIState{
			Sidebar:          "expanded",
			PreviewPaneWidth: 35,
		},
	}
}
