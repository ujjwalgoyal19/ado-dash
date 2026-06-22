package config

// Config holds the user-authored configuration loaded from config.yml.
// This struct is populated by the YAML loader and never auto-rewritten
// after the first-run wizard creates it.
type Config struct {
	Profiles []Profile `yaml:"profiles"`
	Sections Sections  `yaml:"sections"`
	Defaults Defaults  `yaml:"defaults"`
	Keybindings map[string]string `yaml:"keybindings"`
}

// Profile represents one Azure DevOps account (org URL + auth method).
type Profile struct {
	Name            string `yaml:"name"`
	OrganizationURL string `yaml:"organizationUrl"`
	Auth            Auth   `yaml:"auth"`
}

// Auth holds the authentication configuration for a profile.
type Auth struct {
	Type     string `yaml:"type"` // azure-cli | pat | service-principal
	Env      string `yaml:"env"`  // env var name holding the PAT
	TenantID string `yaml:"tenant,omitempty"`
	ClientID string `yaml:"clientId,omitempty"`
	SecretEnv string `yaml:"secretEnv,omitempty"`
}

// Sections holds the configurable sub-tab definitions for each section.
type Sections struct {
	PullRequests []SectionDef `yaml:"pullRequests"`
	WorkItems    []SectionDef `yaml:"workItems"`
	Pipelines    []SectionDef `yaml:"pipelines"`
	Releases     []SectionDef `yaml:"releases"`
}

// SectionDef defines a single sub-tab within a section.
type SectionDef struct {
	Title            string      `yaml:"title"`
	Role             string      `yaml:"role,omitempty"`
	Status           string      `yaml:"status,omitempty"`
	Draft            bool        `yaml:"draft,omitempty"`
	Filter           string      `yaml:"filter,omitempty"`
	Branch           string      `yaml:"branch,omitempty"`
	Limit            int         `yaml:"limit,omitempty"`
	Repositories     []string    `yaml:"repositories,omitempty"`
	WIQL             string      `yaml:"wiql,omitempty"`
	IterationResolved bool       `yaml:"iterationResolved,omitempty"`
	Columns          []ColumnDef `yaml:"columns,omitempty"`
}

// ColumnDef specifies a column to display in the list view.
type ColumnDef struct {
	Field string `yaml:"field"`
	Width int    `yaml:"width"`
}

// Defaults holds display preferences editable from the TUI settings page.
type Defaults struct {
	RefreshSeconds int    `yaml:"refreshSeconds"`
	Limit          int    `yaml:"limit"`
	Theme          string `yaml:"theme"`
	PreviewPane    string `yaml:"previewPane"`
	OpenPRIn       string `yaml:"openPrIn"`
	Notifications  string `yaml:"notifications"`
}
