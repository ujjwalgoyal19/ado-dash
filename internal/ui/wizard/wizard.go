// Package wizard implements the first-run authentication wizard.
// Step 1 (OrgURL) collects the Azure DevOps organization URL.
// Step 2 (Method) lets the user pick Azure CLI, PAT, or Service Principal.
package wizard

const (
	wizW = 62 // canvas width in columns
	wizH = 26 // canvas height in rows
)

// AuthMethod identifies the chosen authentication strategy.
type AuthMethod int

const (
	AuthMethodAzureCLI AuthMethod = iota
	AuthMethodPAT
	AuthMethodServicePrincipal
)

// SelectedMethodMsg is the terminal message emitted when the user completes
// the wizard. App catches it, writes config, and launches the dashboard.
type SelectedMethodMsg struct {
	OrgURL string
	Method AuthMethod
}
