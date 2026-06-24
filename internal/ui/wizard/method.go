package wizard

import (
	"fmt"

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
	cv := ui.NewCanvas(wizH, wizW, t)

	// Header box (rows 0-2)
	cv.FillBG(0, 0, 3, wizW, t.Surface)
	cv.Box(0, 0, 3, wizW, t.Subtle, true)
	cv.Put(1, 1, "ado-dash", t.Text, ui.Bold(), ui.BG(t.Surface))
	cv.Put(1, 9, "  ·  Sign in  (2/2)", t.Muted, ui.BG(t.Surface))

	// Org URL confirmed (row 5)
	cv.Put(5, 2, "Organization: ", t.Muted)
	cv.Put(5, 16, m.orgURL, t.Text)

	// CLI detected + prompt (row 7)
	cv.Put(7, 2, "Azure CLI detected.", t.Green)
	cv.Put(7, 22, " Choose an authentication method:", t.Text)

	// Method list (rows 9-11)
	const nameW = 22
	for i, it := range methodItems {
		r := 9 + i
		sel := i == m.cursor
		br := t.Subtle
		if sel {
			br = t.Purple
		}
		if sel {
			cv.Put(r, 1, "▶", t.Purple)
		}
		cv.Put(r, 3, "[ ", br)
		name := it.name + fmt.Sprintf("%*s", nameW-len([]rune(it.name)), "")
		if sel {
			cv.Put(r, 5, name, t.Text, ui.Bold())
		} else {
			cv.Put(r, 5, name, t.Muted)
		}
		cv.Put(r, 5+nameW, " ]", br)
		if sel {
			cv.Put(r, 5+nameW+3, it.desc, t.Text)
		} else {
			cv.Put(r, 5+nameW+3, it.desc, t.Muted)
		}
	}

	// Footer (row wizH-1)
	c := 1
	for i, seg := range []struct{ k, d string }{
		{"j/k", "move"}, {"enter", "select"}, {"esc", "back"}, {"q", "quit"},
	} {
		cv.Put(wizH-1, c, seg.k, t.Accent)
		c += len([]rune(seg.k))
		cv.Put(wizH-1, c, " "+seg.d, t.Muted)
		c += len(seg.d) + 1
		if i < 3 {
			cv.Put(wizH-1, c, "  ·  ", t.Subtle)
			c += 5
		}
	}

	dialog := cv.Render()

	if m.width > wizW || m.height > wizH {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceBackground(lipgloss.Color(t.Base)))
	}
	return dialog
}
