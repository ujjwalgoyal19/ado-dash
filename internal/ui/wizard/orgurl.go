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
	cv := ui.NewCanvas(wizH, wizW, t)

	// Header box (rows 0-2)
	cv.FillBG(0, 0, 3, wizW, t.Surface)
	cv.Box(0, 0, 3, wizW, t.Subtle, true)
	title := " ado-dash"
	cv.Put(1, 1, "ado-dash", t.Text, ui.Bold(), ui.BG(t.Surface))
	cv.Put(1, 1+len([]rune(title))-1, "  ·  Sign in  (1/2)", t.Muted, ui.BG(t.Surface))

	// Prompt (row 5)
	cv.Put(5, 10, "Enter your Azure DevOps organization URL:", t.Text)

	// Input box (rows 7-9): col 10, inner width 38
	inInner := 38
	inL := 10
	bc := t.Purple
	if m.err != "" {
		bc = t.Red
	}
	cv.Box(7, inL, 3, inInner+2, bc, false)

	// Inline text input — write the rendered value into the canvas row by col.
	// The textinput renders itself; we extract the visible text manually to keep
	// the canvas background uniform.
	val := m.input.Value()
	if val == "" {
		// placeholder
		cv.Put(8, inL+1, " "+m.input.Placeholder, t.Muted)
	} else {
		cv.Put(8, inL+1, " "+val, t.Text)
	}
	// cursor block
	cursorCol := inL + 2 + len([]rune(val))
	if cursorCol < inL+inInner {
		cv.Put(8, cursorCol, "█", t.Accent)
	}

	// Error or blank rows (10-11)
	if m.err != "" {
		cv.Put(10, inL+2, "✗ "+m.err, t.Red)
	}

	// Examples (rows 12-13)
	cv.Put(12, inL, "e.g.  https://dev.azure.com/my-company", t.Muted)
	cv.Put(13, inL+6, "https://my-company.visualstudio.com", t.Muted)

	// Footer (row wizH-1)
	c := 1
	cv.Put(wizH-1, c, "enter", t.Accent)
	c += 5
	cv.Put(wizH-1, c, " confirm", t.Muted)
	c += 8
	cv.Put(wizH-1, c, "  ·  ", t.Subtle)
	c += 5
	cv.Put(wizH-1, c, "esc", t.Accent)
	c += 3
	cv.Put(wizH-1, c, " quit", t.Muted)

	dialog := cv.Render()

	if m.width > wizW || m.height > wizH {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog,
			lipgloss.WithWhitespaceBackground(lipgloss.Color(t.Base)))
	}
	return dialog
}
