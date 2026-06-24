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

func boolBold(b bool) ui.CanvasOpt {
	if b {
		return ui.Bold()
	}
	return ui.Noop()
}

// ── Header ────────────────────────────────────────────────────────

func drawHeader(cv *ui.Canvas, repo string, width int) {
	t := cv.Theme
	cv.FillBG(0, 0, 3, width, t.Surface)
	cv.Box(0, 0, 3, width, t.Subtle, true)

	c := 2
	cv.Put(1, c, "ado-dash", t.Text, ui.Bold(), ui.BG(t.Surface))
	c += 8
	cv.Put(1, c, "  my-company / MyProject", t.Muted, ui.BG(t.Surface))
	c += 24
	cv.Put(1, c, "  /  ", t.Subtle, ui.BG(t.Surface))
	c += 5
	cv.Put(1, c, repo, t.Blue, ui.BG(t.Surface))
	cv.PutR(1, width-3, "[⚙]  [👤] ", t.Muted, ui.BG(t.Surface))
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

func drawSidebar(cv *ui.Canvas, r0, rb int, prCount int) {
	t := cv.Theme
	r := r0 + 1
	for _, it := range sidebarItems {
		cnt := it.count
		if it.active {
			cnt = fmt.Sprintf("%d", prCount)
		}
		if it.active {
			cv.Put(r, 1, "▶", t.Purple)
			cv.Put(r, 3, it.label, t.Text, ui.Bold())
			cv.PutR(r, 9, cnt, t.Muted)
			if it.badge {
				cv.Put(r, 11, "●", t.Badge)
			}
		} else {
			cv.Put(r, 3, it.label, t.Muted)
			cv.PutR(r, 9, cnt, t.Muted)
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

func drawTabBar(cv *ui.Canvas, r, c0 int, tabs []tab) {
	t := cv.Theme
	c := c0 + 1
	for _, tb := range tabs {
		br := t.Subtle
		if tb.active {
			br = t.Purple
		}
		cv.Put(r, c, "[ ", br)
		c += 2
		if tb.active {
			cv.Put(r, c, tb.label, t.Text, ui.Bold())
		} else {
			cv.Put(r, c, tb.label, t.Muted)
		}
		c += len([]rune(tb.label))
		countStr := "  " + tb.count
		if tb.active {
			cv.Put(r, c, countStr, t.Text, ui.Bold())
		} else {
			cv.Put(r, c, countStr, t.Muted)
		}
		c += len([]rune(countStr))
		if tb.dot {
			cv.Put(r, c, " ●", t.Badge)
			c += 2
		}
		cv.Put(r, c, " ]", br)
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

func drawListHeader(cv *ui.Canvas, r, c0, c1 int) {
	t := cv.Theme
	x := calcCols(c0, c1)
	cv.Put(r, x.idX, "#", t.Muted)
	cv.Put(r, x.stateX, "STATE", t.Muted)
	cv.Put(r, x.titleX, "TITLE", t.Muted)
	cv.Put(r, x.authX, "AUTHOR", t.Muted)
	cv.Put(r, x.repoX, "REPO", t.Muted)
	cv.PutR(r, x.ageEnd, "AGE", t.Muted)
	cv.HLine(r+1, c0+1, c1-1, t.Subtle)
}

var stateColor = map[string]func(ui.Theme) lipgloss.Color{
	"active": func(t ui.Theme) lipgloss.Color { return t.Green },
	"draft":  func(t ui.Theme) lipgloss.Color { return t.Yellow },
	"merged": func(t ui.Theme) lipgloss.Color { return t.Purple },
	"closed": func(t ui.Theme) lipgloss.Color { return t.Muted },
}

func drawPRRow(cv *ui.Canvas, r, c0, c1 int, pr PR, selected bool) {
	t := cv.Theme
	x := calcCols(c0, c1)
	if selected {
		cv.FillBG(r, c0+1, 1, c1-c0-1, t.Overlay)
		cv.Put(r, c0, "▶", t.Purple)
	}

	idFG := t.Blue
	if pr.Flag == flagFailed {
		idFG = t.Red
	}
	cv.Put(r, x.idX, rpad(pr.ID, 7), idFG, boolBold(selected))

	stateFG := t.Green
	if cf, ok := stateColor[pr.State]; ok {
		stateFG = cf(t)
	}
	if pr.Flag == flagVote {
		stateFG = t.Yellow
	}
	isDim := pr.Flag == flagDraft
	stOpts := []ui.CanvasOpt{boolBold(selected)}
	if isDim {
		stOpts = append(stOpts, ui.Dim())
	}
	cv.Put(r, x.stateX, rpad(pr.State, 9), stateFG, stOpts...)

	if isDim {
		area := x.titleW - 8
		cv.Put(r, x.titleX, rpad(pr.Title, area), t.Muted, ui.Dim())
		cv.Put(r, x.titleX+area+1, "[draft]", t.Yellow)
	} else {
		cv.Put(r, x.titleX, rpad(pr.Title, x.titleW), t.Text, boolBold(selected))
	}

	cv.Put(r, x.authX, rpad(pr.Author, x.authW), t.Muted, boolBold(selected))
	cv.Put(r, x.repoX, rpad(pr.Repo, x.repoW), t.Muted, boolBold(selected))
	cv.PutR(r, x.ageEnd, pr.Age, t.Muted, boolBold(selected))
}

// ── Preview pane ──────────────────────────────────────────────────

func drawPreview(cv *ui.Canvas, r0, c0, c1 int, pr *PR) {
	t := cv.Theme
	L := c0 + 2
	if pr == nil {
		msg := "No item selected"
		cv.Put(r0+12, c0+(c1-c0-len(msg))/2, msg, t.Muted)
		return
	}

	r := r0 + 1
	num := strings.TrimLeft(strings.TrimPrefix(pr.ID, "PR"), "0")
	cv.Put(r, L, "PR "+num, t.Blue, ui.Bold())
	r++

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
		cv.Put(r, L, l, t.Text, ui.Bold())
		r++
	}
	r++

	cv.Put(r, L, pr.Author, t.Blue)
	cv.Put(r, L+len(pr.Author), " → main", t.Muted)
	r++

	sfg := t.Green
	if pr.Flag == flagVote {
		sfg = t.Yellow
	}
	if cf, ok := stateColor[pr.State]; ok {
		sfg = cf(t)
	}
	cv.Put(r, L, pr.State, sfg)
	cv.Put(r, L+len(pr.State), " · "+pr.Age+" ago", t.Muted)
	r += 2

	cv.Put(r, L, fmt.Sprintf("%d reviewers · %d approved", pr.Reviewers, pr.Approved), t.Muted)
	r += 2

	if pr.Approved > 0 {
		cv.Put(r, L, "✓", t.Green)
		cv.Put(r, L+2, fmt.Sprintf("%d approved", pr.Approved), t.Text)
	} else {
		cv.Put(r, L, "–", t.Muted)
		cv.Put(r, L+2, "no approvals yet", t.Muted)
	}
	r++

	if pr.Waiting > 0 {
		cv.Put(r, L, "⏳", t.Yellow)
		cv.Put(r, L+3, fmt.Sprintf("%d waiting", pr.Waiting), t.Text)
		r++
	}

	bldIcon, bldFG, bldMsg := "✓", t.Green, "Build passing"
	switch pr.Build {
	case buildFail:
		bldIcon, bldFG, bldMsg = "✗", t.Red, "Build failing"
	case buildRun:
		bldIcon, bldFG, bldMsg = "⏳", t.Yellow, "Build running"
	}
	cv.Put(r, L, bldIcon, bldFG)
	cv.Put(r, L+2, bldMsg, t.Text)
	r++

	wiIcon, wiFG := "✓", t.Green
	if !pr.WorkItem {
		wiIcon, wiFG = "✗", t.Red
	}
	cv.Put(r, L, wiIcon, wiFG)
	cv.Put(r, L+2, "Work item linked", t.Text)
}

// ── Footer ────────────────────────────────────────────────────────

type footerSeg struct{ k, d string }

var dashFooter = []footerSeg{
	{"g+p", "PRs"}, {"g+w", "Work"}, {"r", "repo"},
	{"\\", "sidebar"}, {"Z", "zen"}, {"?", "help"}, {"q", "quit"},
}

func drawFooter(cv *ui.Canvas, r int, segs []footerSeg) {
	t := cv.Theme
	cv.FillBG(r, 0, 1, cv.Cols, t.Surface)
	c := 2
	for i, s := range segs {
		cv.Put(r, c, s.k, t.Accent, ui.BG(t.Surface))
		c += len([]rune(s.k))
		cv.Put(r, c, " "+s.d, t.Muted, ui.BG(t.Surface))
		c += len(s.d) + 1
		if i < len(segs)-1 {
			cv.Put(r, c, "  ·  ", t.Subtle, ui.BG(t.Surface))
			c += 5
		}
	}
}

// ── Build full dashboard ─────────────────────────────────────────

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
	cv := ui.NewCanvas(height, width, t)

	drawHeader(cv, repo, width)

	top, bot := 3, height-3
	cv.Box(top, 0, bot-top+1, width, t.Subtle, false)
	cv.VLine(top+1, bot-1, sbR, t.Subtle)
	cv.Put(top, sbR, "┬", t.Subtle)
	cv.Put(bot, sbR, "┴", t.Subtle)

	drawSidebar(cv, top, bot, len(prs))

	tabs := []tab{
		{"My PRs", fmt.Sprintf("%d", len(prs)), true, false},
		{"Review", "5", false, false},
		{"Draft", "1", false, false},
		{"Waiting", "2", false, true},
	}
	drawTabBar(cv, top+1, sbR, tabs)

	c0, c1 := sbR, width-1
	prevDiv := c0 + ((c1-c0)*65/100)

	cv.HLine(top+2, c0+1, c1-1, t.Subtle)
	cv.Put(top+2, c0, "├", t.Subtle)
	cv.Put(top+2, c1, "┤", t.Subtle)
	cv.Put(top+2, prevDiv, "┬", t.Subtle)
	cv.Put(bot, prevDiv, "┴", t.Subtle)
	cv.VLine(top+3, bot-1, prevDiv, t.Subtle)

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

	var sel *PR
	if selIdx >= 0 && selIdx < len(prs) {
		sel = &prs[selIdx]
	}
	drawPreview(cv, listTop, prevDiv, c1, sel)
	if sel != nil {
		cv.PutR(bot-1, c1-2, "↳ d to open", t.Muted)
	}

	drawFooter(cv, height-1, dashFooter)

	return cv.Render()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// LPad left-pads s to width w. Exported for use by the detail view.
func LPad(s string, w int) string { return lpad(s, w) }
