// Package dashboard renders the main PR dashboard screen.
// Layout: header (3 rows) | sidebar (col 0-12) + content (col 13+) | footer (1 row).
// All dimensions are driven by the live terminal size passed into Render().
package dashboard

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ujjwalgoyal19/ado-dash/internal/ui"
)

const sbR = 12 // sidebar right border column (fixed width)

// cell holds a rune + its lipgloss color slots.
type cell struct {
	ch   rune
	fg   lipgloss.Color
	bg   lipgloss.Color
	bold bool
	dim  bool
}

// canvas is a character-grid renderer that mirrors the design's term.jsx approach.
type canvas struct {
	rows, cols int
	cells      [][]cell
	theme      ui.Theme
}

func newCanvas(rows, cols int, t ui.Theme) *canvas {
	cells := make([][]cell, rows)
	for r := range cells {
		cells[r] = make([]cell, cols)
		for c := range cells[r] {
			cells[r][c] = cell{ch: ' ', fg: t.Text, bg: t.Base}
		}
	}
	return &canvas{rows: rows, cols: cols, cells: cells, theme: t}
}

// put writes str at (r,c) with the given style options.
func (cv *canvas) put(r, c int, str string, fg lipgloss.Color, opts ...option) {
	o := applyOpts(opts)
	for i, ch := range str {
		cc := c + i
		if r < 0 || r >= cv.rows || cc < 0 || cc >= cv.cols {
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

func (cv *canvas) putR(r, endCol int, str string, fg lipgloss.Color, opts ...option) {
	cv.put(r, endCol-len([]rune(str))+1, str, fg, opts...)
}

func (cv *canvas) fillBG(r0, c0, h, w int, bg lipgloss.Color) {
	for r := r0; r < r0+h; r++ {
		for c := c0; c < c0+w; c++ {
			if r >= 0 && r < cv.rows && c >= 0 && c < cv.cols {
				cv.cells[r][c].bg = bg
			}
		}
	}
}

func (cv *canvas) hline(r, c0, c1 int, fg lipgloss.Color) {
	for c := c0; c <= c1; c++ {
		cv.put(r, c, "─", fg)
	}
}

func (cv *canvas) vline(r0, r1, c int, fg lipgloss.Color) {
	for r := r0; r <= r1; r++ {
		cv.put(r, c, "│", fg)
	}
}

func (cv *canvas) box(r0, c0, h, w int, fg lipgloss.Color, rounded bool) {
	r1, c1 := r0+h-1, c0+w-1
	tl, tr, bl, br := "┌", "┐", "└", "┘"
	if rounded {
		tl, tr, bl, br = "╭", "╮", "╰", "╯"
	}
	cv.hline(r0, c0+1, c1-1, fg)
	cv.hline(r1, c0+1, c1-1, fg)
	cv.vline(r0+1, r1-1, c0, fg)
	cv.vline(r0+1, r1-1, c1, fg)
	cv.put(r0, c0, tl, fg)
	cv.put(r0, c1, tr, fg)
	cv.put(r1, c0, bl, fg)
	cv.put(r1, c1, br, fg)
}

// render converts the canvas to a string with ANSI color codes.
func (cv *canvas) render() string {
	var sb strings.Builder
	for ri, row := range cv.cells {
		var i int
		for i < len(row) {
			// group consecutive cells with the same styling
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
			s := string(runes)
			st := lipgloss.NewStyle().Foreground(run[0].fg)
			if run[0].bg != cv.theme.Base {
				st = st.Background(run[0].bg)
			}
			if run[0].bold {
				st = st.Bold(true)
			}
			if run[0].dim {
				st = st.Faint(true)
			}
			sb.WriteString(st.Render(s))
		}
		if ri < cv.rows-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// ── option helpers ────────────────────────────────────────────────

type opts struct {
	bold bool
	dim  bool
	bg   lipgloss.Color
}

type option func(*opts)

func bold() option          { return func(o *opts) { o.bold = true } }
func dim() option           { return func(o *opts) { o.dim = true } }
func bg(c lipgloss.Color) option { return func(o *opts) { o.bg = c } }

func applyOpts(os []option) opts {
	var o opts
	for _, fn := range os {
		fn(&o)
	}
	return o
}

// ── Helpers ───────────────────────────────────────────────────────

func rpad(s string, w int) string {
	r := []rune(s)
	if len(r) >= w {
		return string(r[:w])
	}
	return s + strings.Repeat(" ", w-len(r))
}

func lpad(s string, w int) string {
	r := []rune(s)
	if len(r) >= w {
		return string(r[:w])
	}
	return strings.Repeat(" ", w-len(r)) + s
}

// ── Header ────────────────────────────────────────────────────────

func drawHeader(cv *canvas, repo string, width int) {
	t := cv.theme
	cv.fillBG(0, 0, 3, width, t.Surface)
	cv.box(0, 0, 3, width, t.Subtle, true)

	c := 2
	cv.put(1, c, "ado-dash", t.Text, bold(), bg(t.Surface))
	c += 8
	cv.put(1, c, "  my-company / MyProject", t.Muted, bg(t.Surface))
	c += 24
	cv.put(1, c, "  /  ", t.Subtle, bg(t.Surface))
	c += 5
	cv.put(1, c, repo, t.Blue, bg(t.Surface))
	cv.putR(1, width-3, "[⚙]  [👤] ", t.Muted, bg(t.Surface))
}

// ── Sidebar ───────────────────────────────────────────────────────

var sidebarItems = []struct {
	label  string
	count  string
	active bool
	badge  bool
}{
	{"PRs", "", true, true},
	{"Work", "12", false, false},
	{"Pipes", "3", false, false},
	{"Rel", "1", false, false},
}

func drawSidebar(cv *canvas, r0, rb int, prCount int) {
	t := cv.theme
	r := r0 + 1
	for _, it := range sidebarItems {
		cnt := it.count
		if it.active {
			cnt = fmt.Sprintf("%d", prCount)
		}
		if it.active {
			cv.put(r, 1, "▶", t.Purple)
			cv.put(r, 3, it.label, t.Text, bold())
			cv.putR(r, 9, cnt, t.Muted)
			if it.badge {
				cv.put(r, 11, "●", t.Badge)
			}
		} else {
			cv.put(r, 3, it.label, t.Muted)
			cv.putR(r, 9, cnt, t.Muted)
		}
		r++
	}
}

// ── Tab bar ───────────────────────────────────────────────────────

type tab struct {
	label  string
	count  string
	active bool
	dot    bool
}

func drawTabBar(cv *canvas, r, c0 int, tabs []tab) {
	t := cv.theme
	c := c0 + 1
	for _, tb := range tabs {
		br := t.Subtle
		if tb.active {
			br = t.Purple
		}
		cv.put(r, c, "[ ", br)
		c += 2
		if tb.active {
			cv.put(r, c, tb.label, t.Text, bold())
		} else {
			cv.put(r, c, tb.label, t.Muted)
		}
		c += len([]rune(tb.label))
		countStr := "  " + tb.count
		if tb.active {
			cv.put(r, c, countStr, t.Text, bold())
		} else {
			cv.put(r, c, countStr, t.Muted)
		}
		c += len([]rune(countStr))
		if tb.dot {
			cv.put(r, c, " ●", t.Badge)
			c += 2
		}
		cv.put(r, c, " ]", br)
		c += 4
	}
}

// ── PR list columns ───────────────────────────────────────────────

type listCols struct {
	idX, stateX, titleX, titleW, authX, authW, repoX, repoW, ageEnd int
}

func calcCols(c0, c1 int) listCols {
	left := c0 + 1
	ageEnd := c1 - 1
	ageX := ageEnd - 3
	repoW := 12
	repoX := ageX - 2 - repoW
	authW := 11
	authX := repoX - 1 - authW
	stateX := left + 8
	titleX := stateX + 10
	titleW := authX - 2 - titleX
	_ = left
	return listCols{stateX - 8, stateX, titleX, titleW, authX, authW, repoX, repoW, ageEnd}
}

func drawListHeader(cv *canvas, r, c0, c1 int) {
	t := cv.theme
	x := calcCols(c0, c1)
	cv.put(r, x.idX, "#", t.Muted)
	cv.put(r, x.stateX, "STATE", t.Muted)
	cv.put(r, x.titleX, "TITLE", t.Muted)
	cv.put(r, x.authX, "AUTHOR", t.Muted)
	cv.put(r, x.repoX, "REPO", t.Muted)
	cv.putR(r, x.ageEnd, "AGE", t.Muted)
	cv.hline(r+1, c0+1, c1-1, t.Subtle)
}

var stateColor = map[string]func(ui.Theme) lipgloss.Color{
	"active": func(t ui.Theme) lipgloss.Color { return t.Green },
	"draft":  func(t ui.Theme) lipgloss.Color { return t.Yellow },
	"merged": func(t ui.Theme) lipgloss.Color { return t.Purple },
	"closed": func(t ui.Theme) lipgloss.Color { return t.Muted },
}

func drawPRRow(cv *canvas, r, c0, c1 int, pr PR, selected bool) {
	t := cv.theme
	x := calcCols(c0, c1)
	if selected {
		cv.fillBG(r, c0+1, 1, c1-c0-1, t.Overlay)
		cv.put(r, c0, "▶", t.Purple)
	}

	idFG := t.Blue
	if pr.Flag == flagFailed {
		idFG = t.Red
	}
	cv.put(r, x.idX, rpad(pr.ID, 7), idFG, boolBold(selected))

	stateFG := t.Green
	if cf, ok := stateColor[pr.State]; ok {
		stateFG = cf(t)
	}
	if pr.Flag == flagVote {
		stateFG = t.Yellow
	}
	isDim := pr.Flag == flagDraft
	stOpts := []option{boolBold(selected)}
	if isDim {
		stOpts = append(stOpts, dim())
	}
	cv.put(r, x.stateX, rpad(pr.State, 9), stateFG, stOpts...)

	if isDim {
		area := x.titleW - 8
		cv.put(r, x.titleX, rpad(pr.Title, area), t.Muted, dim())
		cv.put(r, x.titleX+area+1, "[draft]", t.Yellow)
	} else {
		cv.put(r, x.titleX, rpad(pr.Title, x.titleW), t.Text, boolBold(selected))
	}

	cv.put(r, x.authX, rpad(pr.Author, x.authW), t.Muted, boolBold(selected))
	cv.put(r, x.repoX, rpad(pr.Repo, x.repoW), t.Muted, boolBold(selected))
	cv.putR(r, x.ageEnd, pr.Age, t.Muted, boolBold(selected))
}

func boolBold(b bool) option {
	if b {
		return bold()
	}
	return func(*opts) {}
}

// ── Preview pane ──────────────────────────────────────────────────

func drawPreview(cv *canvas, r0, c0, c1 int, pr *PR) {
	t := cv.theme
	L := c0 + 2
	if pr == nil {
		msg := "No item selected"
		cv.put(r0+12, c0+(c1-c0-len(msg))/2, msg, t.Muted)
		return
	}

	r := r0 + 1
	num := strings.TrimLeft(strings.TrimPrefix(pr.ID, "PR"), "0")
	cv.put(r, L, "PR "+num, t.Blue, bold())
	r++

	// wrap title to 2 lines
	W := c1 - L - 1
	words := strings.Fields(pr.Title)
	var lines []string
	cur := ""
	for _, w := range words {
		if cur == "" {
			cur = w
		} else if len(cur)+1+len(w) <= W {
			cur += " " + w
		} else {
			lines = append(lines, cur)
			cur = w
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	for _, l := range lines[:min(2, len(lines))] {
		cv.put(r, L, l, t.Text, bold())
		r++
	}
	r++

	cv.put(r, L, pr.Author, t.Blue)
	cv.put(r, L+len(pr.Author), " → main", t.Muted)
	r++

	sfg := t.Green
	if pr.Flag == flagVote {
		sfg = t.Yellow
	}
	if cf, ok := stateColor[pr.State]; ok {
		sfg = cf(t)
	}
	cv.put(r, L, pr.State, sfg)
	cv.put(r, L+len(pr.State), " · "+pr.Age+" ago", t.Muted)
	r += 2

	cv.put(r, L, fmt.Sprintf("%d reviewers · %d approved", pr.Reviewers, pr.Approved), t.Muted)
	r += 2

	if pr.Approved > 0 {
		cv.put(r, L, "✓", t.Green)
		cv.put(r, L+2, fmt.Sprintf("%d approved", pr.Approved), t.Text)
	} else {
		cv.put(r, L, "–", t.Muted)
		cv.put(r, L+2, "no approvals yet", t.Muted)
	}
	r++

	if pr.Waiting > 0 {
		cv.put(r, L, "⏳", t.Yellow)
		cv.put(r, L+3, fmt.Sprintf("%d waiting", pr.Waiting), t.Text)
		r++
	}

	bldIcon, bldFG, bldMsg := "✓", t.Green, "Build passing"
	switch pr.Build {
	case buildFail:
		bldIcon, bldFG, bldMsg = "✗", t.Red, "Build failing"
	case buildRun:
		bldIcon, bldFG, bldMsg = "⏳", t.Yellow, "Build running"
	}
	cv.put(r, L, bldIcon, bldFG)
	cv.put(r, L+2, bldMsg, t.Text)
	r++

	wiIcon, wiFG := "✓", t.Green
	if !pr.WorkItem {
		wiIcon, wiFG = "✗", t.Red
	}
	cv.put(r, L, wiIcon, wiFG)
	cv.put(r, L+2, "Work item linked", t.Text)
}

// ── Footer ────────────────────────────────────────────────────────

type footerSeg struct{ k, d string }

var dashFooter = []footerSeg{
	{"g+p", "PRs"}, {"g+w", "Work"}, {"r", "repo"},
	{"\\", "sidebar"}, {"Z", "zen"}, {"?", "help"}, {"q", "quit"},
}

func drawFooter(cv *canvas, r int, segs []footerSeg) {
	t := cv.theme
	cv.fillBG(r, 0, 1, cv.cols, t.Surface)
	c := 2
	for i, s := range segs {
		cv.put(r, c, s.k, t.Accent, bg(t.Surface))
		c += len([]rune(s.k))
		cv.put(r, c, " "+s.d, t.Muted, bg(t.Surface))
		c += len(s.d) + 1
		if i < len(segs)-1 {
			cv.put(r, c, "  ·  ", t.Subtle, bg(t.Surface))
			c += 5
		}
	}
}

// ── Build full dashboard ─────────────────────────────────────────

// Render builds the complete dashboard view string for the given repo and
// selected PR index.
// Render builds the complete dashboard view string sized to the live terminal.
// width and height come from tea.WindowSizeMsg; fall back to 140×40 if zero.
func Render(repo string, selIdx int, prs []PR, width, height int) string {
	if width < 40 {
		width = 140
	}
	if height < 10 {
		height = 40
	}

	t := ui.ActiveTheme
	cv := newCanvas(height, width, t)

	// Header (rows 0-2)
	drawHeader(cv, repo, width)

	// Outer content box: rows 3 .. height-3; footer at height-1.
	top, bot := 3, height-3
	cv.box(top, 0, bot-top+1, width, t.Subtle, false)
	cv.vline(top+1, bot-1, sbR, t.Subtle)
	cv.put(top, sbR, "┬", t.Subtle)
	cv.put(bot, sbR, "┴", t.Subtle)

	// Sidebar
	drawSidebar(cv, top, bot, len(prs))

	// Tab bar + divider
	tabs := []tab{
		{"My PRs", fmt.Sprintf("%d", len(prs)), true, false},
		{"Review", "5", false, false},
		{"Draft", "1", false, false},
		{"Waiting", "2", false, true},
	}
	drawTabBar(cv, top+1, sbR, tabs)

	// Preview divider: 65% of available content width.
	c0, c1 := sbR, width-1
	contentW := c1 - c0
	prevDiv := c0 + (contentW * 65 / 100)

	// Tab divider + preview column junctions
	cv.hline(top+2, c0+1, c1-1, t.Subtle)
	cv.put(top+2, c0, "├", t.Subtle)
	cv.put(top+2, c1, "┤", t.Subtle)
	cv.put(top+2, prevDiv, "┬", t.Subtle)
	cv.put(bot, prevDiv, "┴", t.Subtle)
	cv.vline(top+3, bot-1, prevDiv, t.Subtle)

	// PR list
	listTop := top + 2
	if len(prs) > 0 {
		drawListHeader(cv, listTop+1, c0, prevDiv)
		r := listTop + 3
		for i, pr := range prs {
			if r > bot-1 {
				break
			}
			drawPRRow(cv, r, c0, prevDiv, pr, i == selIdx)
			r++
		}
	}

	// Preview pane
	var sel *PR
	if selIdx >= 0 && selIdx < len(prs) {
		sel = &prs[selIdx]
	}
	drawPreview(cv, listTop, prevDiv, c1, sel)
	if sel != nil {
		cv.putR(bot-1, c1-2, "↳ d to open", t.Muted)
	}

	// Footer
	drawFooter(cv, height-1, dashFooter)

	return cv.render()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// LPad left-pads s to width w. Exported for use by the detail view.
func LPad(s string, w int) string { return lpad(s, w) }
