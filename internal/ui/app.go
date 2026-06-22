package ui

import tea "github.com/charmbracelet/bubbletea"

// App is the root Bubble Tea model. It owns a screen stack and the
// current terminal dimensions. Child models are pushed onto the stack
// and popped when the user navigates back.
type App struct {
	width  int
	height int
	stack  []tea.Model
}

// New returns an App with the given initial screen pushed.
func New(initial tea.Model) *App {
	return &App{stack: []tea.Model{initial}}
}

func (a *App) Init() tea.Cmd {
	if len(a.stack) == 0 {
		return nil
	}
	return a.stack[len(a.stack)-1].Init()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(a.stack) == 0 {
		return a, tea.Quit
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	case PushMsg:
		a.stack = append(a.stack, msg.Model)
		return a, tea.Batch(msg.Model.Init(), a.sizeCmd())
	case PopMsg:
		if len(a.stack) > 1 {
			a.stack = a.stack[:len(a.stack)-1]
		}
		return a, nil
	case ReplaceMsg:
		a.stack = []tea.Model{msg.Model}
		return a, tea.Batch(msg.Model.Init(), a.sizeCmd())
	}
	updated, cmd := a.stack[len(a.stack)-1].Update(msg)
	a.stack[len(a.stack)-1] = updated
	return a, cmd
}

// sizeCmd re-emits the stored terminal size as a WindowSizeMsg so that a
// newly pushed or replaced screen receives dimensions immediately.
func (a *App) sizeCmd() tea.Cmd {
	if a.width == 0 && a.height == 0 {
		return nil
	}
	w, h := a.width, a.height
	return func() tea.Msg { return tea.WindowSizeMsg{Width: w, Height: h} }
}

func (a *App) View() string {
	if len(a.stack) == 0 {
		return ""
	}
	return a.stack[len(a.stack)-1].View()
}
