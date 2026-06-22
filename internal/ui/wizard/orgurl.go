package wizard

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ujjwalgoyal19/ado-dash/internal/ui"
)

// OrgURL is the first wizard step: enter the Azure DevOps org URL.
type OrgURL struct {
	input  textinput.Model
	err    string
	width  int
	height int
}

func NewOrgURL() OrgURL {
	ti := textinput.New()
	ti.Placeholder = "https://dev.azure.com/my-org"
	ti.CharLimit = 256
	ti.Width = 36
	ti.Prompt = ""
	ti.Focus()
	return OrgURL{input: ti}
}

func (m OrgURL) Init() tea.Cmd { return textinput.Blink }

func (m OrgURL) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			url := strings.TrimSpace(m.input.Value())
			if !isValidOrgURL(url) {
				m.err = "Must be https://dev.azure.com/{org}"
				return m, nil
			}
			next := NewMethod(url)
			return m, func() tea.Msg { return ui.PushMsg{Model: next} }
		case "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func isValidOrgURL(s string) bool {
	return strings.HasPrefix(s, "https://dev.azure.com/") ||
		strings.HasSuffix(s, ".visualstudio.com")
}

func (m OrgURL) View() string {
	t := ui.ActiveTheme
	fg := func(col lipgloss.Color) lipgloss.Style { return lipgloss.NewStyle().Foreground(col) }
	fb := func(col lipgloss.Color) lipgloss.Style { return lipgloss.NewStyle().Foreground(col).Bold(true) }

	inner := wizW - 2
	title := " " + fb(t.Text).Render("ado-dash") + fg(t.Muted).Render("  ·  Sign in  (1/2)")

	rows := []string{
		fg(t.Subtle).Render("╭" + strings.Repeat("─", inner) + "╮"),
		fg(t.Subtle).Render("│") + title + strings.Repeat(" ", inner-lipgloss.Width(title)) + fg(t.Subtle).Render("│"),
		fg(t.Subtle).Render("╰" + strings.Repeat("─", inner) + "╯"),
		"",
		"",
		strings.Repeat(" ", 10) + fg(t.Text).Render("Enter your Azure DevOps organization URL:"),
		"",
	}

	inInner := 38
	lpad := strings.Repeat(" ", 10)
	bc := fg(t.Purple)
	if m.err != "" {
		bc = fg(t.Red)
	}

	inputContent := " " + m.input.View()
	rightPad := strings.Repeat(" ", max(0, inInner-1-lipgloss.Width(inputContent)))
	rows = append(rows,
		lpad+bc.Render("┌"+strings.Repeat("─", inInner)+"┐"),
		lpad+bc.Render("│")+inputContent+rightPad+bc.Render("│"),
		lpad+bc.Render("└"+strings.Repeat("─", inInner)+"┘"),
	)

	if m.err != "" {
		rows = append(rows,
			lpad+"  "+fg(t.Red).Render("✗ "+m.err),
			"",
		)
	} else {
		rows = append(rows, "", "")
	}

	rows = append(rows,
		lpad+fg(t.Muted).Render("e.g.  https://dev.azure.com/my-company"),
		strings.Repeat(" ", 16)+fg(t.Muted).Render("https://my-company.visualstudio.com"),
	)

	for len(rows) < wizH-1 {
		rows = append(rows, "")
	}

	footer := " " + fg(t.Accent).Render("enter") + fg(t.Muted).Render(" confirm") +
		fg(t.Subtle).Render("  ·  ") +
		fg(t.Accent).Render("esc") + fg(t.Muted).Render(" quit")
	rows = append(rows, footer)

	dialog := strings.Join(rows, "\n")

	// Center the fixed-size dialog in the live terminal.
	if m.width > wizW || m.height > wizH {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceBackground(lipgloss.Color(t.Base)))
	}
	return dialog
}
