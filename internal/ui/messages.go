package ui

import tea "github.com/charmbracelet/bubbletea"

// PushMsg asks App to push a new screen onto the stack.
type PushMsg struct{ Model tea.Model }

// PopMsg asks App to pop the current screen and return to the previous one.
type PopMsg struct{}

// ReplaceMsg asks App to replace the entire stack with a new root screen.
// Used after the auth wizard completes to swap in the dashboard.
type ReplaceMsg struct{ Model tea.Model }
