package dashboard

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ujjwalgoyal19/ado-dash/internal/ui"
)

// Model is the main dashboard Bubble Tea model.
type Model struct {
	repo   string
	prs    []PR
	selIdx int
	width  int
	height int
}

// New returns a dashboard for the given repository using mock data.
func New(repo string) Model {
	prs := PRsForRepo(repo)
	return Model{repo: repo, prs: prs}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.selIdx < len(m.prs)-1 {
				m.selIdx++
			}
		case "k", "up":
			if m.selIdx > 0 {
				m.selIdx--
			}
		case "r":
			// repo switcher — placeholder
		}
	}
	return m, nil
}

func (m Model) View() string {
	_ = ui.ActiveTheme // ensure theme is referenced
	return Render(m.repo, m.selIdx, m.prs)
}
