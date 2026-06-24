package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Canvas is a fixed-size character grid. Every cell is initialised with the
// theme's Base background so rendering always emits explicit colour codes —
// no cell is ever transparent. Both the dashboard and the wizard use this
// type so they share identical rendering behaviour.
type Canvas struct {
	Rows, Cols int
	cells      [][]canvasCell
	Theme      Theme
}

type canvasCell struct {
	ch   rune
	fg   lipgloss.Color
	bg   lipgloss.Color
	bold bool
	dim  bool
}

// NewCanvas returns a Canvas with every cell filled to the theme's Base colour.
func NewCanvas(rows, cols int, t Theme) *Canvas {
	cells := make([][]canvasCell, rows)
	for r := range cells {
		cells[r] = make([]canvasCell, cols)
		for c := range cells[r] {
			cells[r][c] = canvasCell{ch: ' ', fg: t.Text, bg: t.Base}
		}
	}
	return &Canvas{Rows: rows, Cols: cols, cells: cells, Theme: t}
}

// Put writes str at (r,c). Opts control bold/dim/background overrides.
func (cv *Canvas) Put(r, c int, str string, fg lipgloss.Color, opts ...CanvasOpt) {
	o := applyCanvasOpts(opts)
	for i, ch := range str {
		cc := c + i
		if r < 0 || r >= cv.Rows || cc < 0 || cc >= cv.Cols {
			continue
		}
		cell := &cv.cells[r][cc]
		cell.ch = ch
		cell.fg = fg
		cell.bold = o.bold
		cell.dim = o.dim
		if o.bg != "" {
			cell.bg = o.bg
		}
	}
}

// Noop returns a CanvasOpt that does nothing. Useful as a conditional fallback.
func Noop() CanvasOpt { return func(*CanvasOpts) {} }

// PutR right-aligns str so it ends at endCol.
func (cv *Canvas) PutR(r, endCol int, str string, fg lipgloss.Color, opts ...CanvasOpt) {
	cv.Put(r, endCol-len([]rune(str))+1, str, fg, opts...)
}

// FillBG sets the background colour for a rectangular region.
func (cv *Canvas) FillBG(r0, c0, h, w int, bg lipgloss.Color) {
	for r := r0; r < r0+h; r++ {
		for c := c0; c < c0+w; c++ {
			if r >= 0 && r < cv.Rows && c >= 0 && c < cv.Cols {
				cv.cells[r][c].bg = bg
			}
		}
	}
}

// HLine draws a horizontal line of '─' from c0 to c1 inclusive.
func (cv *Canvas) HLine(r, c0, c1 int, fg lipgloss.Color) {
	for c := c0; c <= c1; c++ {
		cv.Put(r, c, "─", fg)
	}
}

// VLine draws a vertical line of '│' from r0 to r1 inclusive.
func (cv *Canvas) VLine(r0, r1, c int, fg lipgloss.Color) {
	for r := r0; r <= r1; r++ {
		cv.Put(r, c, "│", fg)
	}
}

// Box draws a border box. rounded selects ╭╮╰╯ corners; otherwise ┌┐└┘.
func (cv *Canvas) Box(r0, c0, h, w int, fg lipgloss.Color, rounded bool) {
	r1, c1 := r0+h-1, c0+w-1
	tl, tr, bl, br := "┌", "┐", "└", "┘"
	if rounded {
		tl, tr, bl, br = "╭", "╮", "╰", "╯"
	}
	cv.HLine(r0, c0+1, c1-1, fg)
	cv.HLine(r1, c0+1, c1-1, fg)
	cv.VLine(r0+1, r1-1, c0, fg)
	cv.VLine(r0+1, r1-1, c1, fg)
	cv.Put(r0, c0, tl, fg)
	cv.Put(r0, c1, tr, fg)
	cv.Put(r1, c0, bl, fg)
	cv.Put(r1, c1, br, fg)
}

// Render converts the canvas to an ANSI string. Background is always emitted
// so that placing the output inside lipgloss.Place produces no transparent gaps.
func (cv *Canvas) Render() string {
	var sb strings.Builder
	for ri, row := range cv.cells {
		var i int
		for i < len(row) {
			start := i
			for i < len(row) &&
				row[i].fg == row[start].fg &&
				row[i].bg == row[start].bg &&
				row[i].bold == row[start].bold &&
				row[i].dim == row[start].dim {
				i++
			}
			run := row[start:i]
			var runes []rune
			for _, cl := range run {
				runes = append(runes, cl.ch)
			}
			st := lipgloss.NewStyle().
				Foreground(run[0].fg).
				Background(run[0].bg) // always explicit — no transparent cells
			if run[0].bold {
				st = st.Bold(true)
			}
			if run[0].dim {
				st = st.Faint(true)
			}
			sb.WriteString(st.Render(string(runes)))
		}
		if ri < cv.Rows-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// ── Option helpers ────────────────────────────────────────────────

// CanvasOpts holds resolved styling options for a cell.
type CanvasOpts struct {
	bold bool
	dim  bool
	bg   lipgloss.Color
}

// CanvasOpt is a functional option for Put/PutR.
type CanvasOpt func(*CanvasOpts)

// Bold makes the cell bold.
func Bold() CanvasOpt { return func(o *CanvasOpts) { o.bold = true } }

// Dim makes the cell faint.
func Dim() CanvasOpt { return func(o *CanvasOpts) { o.dim = true } }

// BG overrides the cell background colour.
func BG(c lipgloss.Color) CanvasOpt { return func(o *CanvasOpts) { o.bg = c } }

func applyCanvasOpts(os []CanvasOpt) CanvasOpts {
	var o CanvasOpts
	for _, fn := range os {
		fn(&o)
	}
	return o
}
