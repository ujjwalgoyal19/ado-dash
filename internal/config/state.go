package config

// State holds runtime state auto-written by the TUI.
// It lives in ~/.config/ado-dash/state.json and is never hand-edited.
type State struct {
	ActiveProfile string                    `json:"activeProfile"`
	Profiles      map[string]ProfileState   `json:"profiles"`
	UI            UIState                   `json:"ui"`
}

// ProfileState holds per-profile runtime selections.
type ProfileState struct {
	DefaultProject  string `json:"defaultProject"`
	DefaultTeamID   string `json:"defaultTeamId"`
	DefaultTeamName string `json:"defaultTeamName"`
}

// UIState holds layout preferences persisted across sessions.
type UIState struct {
	Sidebar         string `json:"sidebar"`          // expanded | collapsed | hidden
	PreviewPaneWidth int   `json:"previewPaneWidth"` // percentage 20–60
}
