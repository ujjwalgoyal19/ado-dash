package wizard

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ujjwalgoyal19/ado-dash/internal/ui"
	"github.com/ujjwalgoyal19/ado-dash/internal/ui/dashboard"
)

type methodItem struct {
	method AuthMethod
	name   string
	desc   string
}

var methodItems = []methodItem{
	{AuthMethodAzureCLI, "Azure CLI", "Browser-based login — recommended"},
	{AuthMethodPAT, "Personal Access Token", "Env var reference"},
	{AuthMethodServicePrincipal, "Service Principal", "Client ID + secret (CI / scripts)"},
}

// Method is the second wizard step: choose an authentication method.
type Method struct {
	orgURL string
	cursor int
	width  int
	height int
}

func NewMethod(orgURL string) Method { return Method{orgURL: orgURL} }

func (m Method) Init() tea.Cmd { return nil }

func (m Method) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(methodItems)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter":
			return m, func() tea.Msg {
				// Launch dashboard with the first available repo.
				// Auth config persistence happens in a later slice.
				return ui.ReplaceMsg{Model: dashboard.New(dashboard.Repos[0])}
			}
		case "esc":
			return m, func() tea.Msg { return ui.PopMsg{} }
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Method) View() string {
	t := ui.ActiveTheme
	fg := func(col lipgloss.Color) lipgloss.Style { return lipgloss.NewStyle().Foreground(col) }
	fb := func(col lipgloss.Color) lipgloss.Style { return lipgloss.NewStyle().Foreground(col).Bold(true) }

	inner := wizW - 2
	title := " " + fb(t.Text).Render("ado-dash") + fg(t.Muted).Render("  ·  Sign in  (2/2)")

	rows := []string{
		fg(t.Subtle).Render("╭" + strings.Repeat("─", inner) + "╮"),
		fg(t.Subtle).Render("│") + title + strings.Repeat(" ", inner-lipgloss.Width(title)) + fg(t.Subtle).Render("│"),
		fg(t.Subtle).Render("╰" + strings.Repeat("─", inner) + "╯"),
		"",
		"",
		"  " + fg(t.Muted).Render("Organization: ") + fg(t.Text).Render(m.orgURL),
		"",
		"  " + fg(t.Green).Render("Azure CLI detected.") + " " + fg(t.Text).Render("Choose an authentication method:"),
		"",
	}

	const nameW = 22
	for i, it := range methodItems {
		sel := i == m.cursor
		br := fg(t.Subtle)
		if sel {
			br = fg(t.Purple)
		}
		nameRender := fb(t.Text)
		if !sel {
			nameRender = fg(t.Muted)
		}
		descRender := fg(t.Text)
		if !sel {
			descRender = fg(t.Muted)
		}

		prefix := "   "
		if sel {
			prefix = " " + fg(t.Purple).Render("▶") + " "
		}
		name := it.name + strings.Repeat(" ", max(0, nameW-len(it.name)))
		rows = append(rows,
			prefix+br.Render("[ ")+nameRender.Render(name)+br.Render(" ]")+" "+descRender.Render(it.desc),
		)
	}

	for len(rows) < wizH-1 {
		rows = append(rows, "")
	}

	dot := fg(t.Subtle).Render("  ·  ")
	footer := " " + fg(t.Accent).Render("j/k") + fg(t.Muted).Render(" move") +
		dot + fg(t.Accent).Render("enter") + fg(t.Muted).Render(" select") +
		dot + fg(t.Accent).Render("esc") + fg(t.Muted).Render(" back") +
		dot + fg(t.Accent).Render("q") + fg(t.Muted).Render(" quit")
	rows = append(rows, footer)

	dialog := strings.Join(rows, "\n")

	// Center the fixed-size dialog in the live terminal.
	if m.width > wizW || m.height > wizH {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceBackground(lipgloss.Color(ui.ActiveTheme.Base)))
	}
	return dialog
}
