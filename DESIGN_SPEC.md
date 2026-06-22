# ado-dash — TUI Design Specification

---

## Design Language

### Principles

- **Information density over decoration.** Every character on screen earns its place. No decorative borders around things that don't need them.
- **Vim-native feel.** Users coming from gh-dash, lazygit, k9s expect leader keys, `j`/`k` navigation, and modal-style keybindings.
- **Semantic color, not decorative color.** Colors convey state: green = good, red = bad, yellow = attention needed. Never used just for variety.
- **Chrome hides on demand.** Sidebar, header, and footer can all be hidden to give content full screen real estate (zen mode).

### Grid & Spacing

Terminal cells are the unit. All sizing is in character cells (columns × rows).

```
1 unit = 1 terminal cell

Padding inside panes:    1 unit horizontal, 0 vertical (tight by default)
Gap between panes:       1 unit (the border character itself)
Minimum content width:   60 columns
Comfortable width:       120–160 columns
```

---

## Color System (Catppuccin Mocha — Default)

All colors reference named slots. No hardcoded hex anywhere outside the theme file.

| Slot | Hex | Role |
|---|---|---|
| `Base` | `#1e1e2e` | Main terminal background |
| `Surface` | `#181825` | Sidebar background, inactive panels |
| `Overlay` | `#313244` | Selected row highlight, active tab background, modal backdrop |
| `Text` | `#cdd6f4` | Primary text — titles, content, labels |
| `Muted` | `#6c7086` | Secondary text — timestamps, repo names, metadata |
| `Subtle` | `#45475a` | Borders, dividers, inactive tab underlines |
| `Green` | `#a6e3a1` | Active/succeeded state, approved vote |
| `Yellow` | `#f9e2af` | Warning, draft, in-progress, waiting vote |
| `Red` | `#f38ba8` | Rejected, failed, error state |
| `Blue` | `#89b4fa` | Links, info, PR IDs, informational |
| `Purple` | `#cba6f7` | Accents — active tab underline, focused pane border |
| `Accent` | `#89dceb` | Cursor, active keybinding hint, resizing indicator |
| `Badge` | `#f38ba8` | Notification dot |

### Semantic Mapping

| UI element | Color slot |
|---|---|
| Terminal background | `Base` |
| Sidebar, inactive tab bg | `Surface` |
| Selected row bg | `Overlay` |
| All body text | `Text` |
| Timestamps, age, repo, branch | `Muted` |
| Borders, tab underlines (inactive) | `Subtle` |
| Active tab underline | `Purple` |
| Focused pane border | `Purple` |
| PR #ID column | `Blue` |
| State: Active, Succeeded | `Green` |
| State: Draft, In Progress, Waiting | `Yellow` |
| State: Rejected, Failed, Blocked | `Red` |
| Sidebar active item | `Text` bold + `Purple` left gutter mark |
| Footer key hints | `Muted` |
| Notification badge `●` | `Badge` |
| Zen mode exit hint | `Muted` |

---

## Typography

Terminal has no font control. Styling is through ANSI attributes only.

| Style | Usage |
|---|---|
| **Bold** | Screen titles, PR title in preview pane, active sidebar item, column headers |
| *Italic* | Not used (poor terminal support — skip entirely) |
| Dim / Muted | Secondary metadata, inactive items, timestamps |
| Underline | Active tab label only |
| Strikethrough | Not used |
| Reverse | Not used (use `Overlay` bg instead) |

---

## Layout System

### Terminal Width Breakpoints

| Width | Layout |
|---|---|
| ≥ 140 cols | Full: sidebar (expanded) + list + preview pane |
| 100–139 cols | Reduced: sidebar (expanded) + list only (preview auto-hidden) |
| < 100 cols | Compact: sidebar (collapsed, 3 chars) + list only |

These are **override** states — they do not mutate the user's saved sidebar/preview preference.

### Layout Proportions (at ≥ 140 cols, defaults)

```
┌──────────────────────────────────────────────────────────────────────┐
│ HEADER                                                    (1 row)    │
├──────────┬───────────────────────────────────────────────────────────┤
│ SIDEBAR  │ TAB BAR                                        (1 row)    │
│ 12 cols  ├────────────────────────────────────┬──────────────────────┤
│          │ LIST PANE                          │ PREVIEW PANE         │
│          │ flex (fills remaining)             │ 35% of content area  │
│          │                                    │                      │
├──────────┴────────────────────────────────────┴──────────────────────┤
│ FOOTER                                                    (1 row)    │
└──────────────────────────────────────────────────────────────────────┘
```

- **Sidebar**: 12 cols expanded, 3 cols collapsed, 0 cols hidden
- **Preview pane**: default 35% of content area (content area = total width − sidebar)
- **List pane**: fills remaining space between sidebar and preview pane
- **Header + Footer**: always 1 row each (hidden in zen mode)
- **Tab bar**: 1 row, always visible (stays in zen mode)

---

## Component Library

### 1. Header Bar

```
╭─────────────────────────────────────────────────────────────────────╮
│  ado-dash  MyOrg / MyProject                            [⚙]  [👤]  │
╰─────────────────────────────────────────────────────────────────────╯
```

- Height: 3 rows (top border + content + bottom border) — or 1 row with no border in compact mode
- Left: `ado-dash` in **bold** `Text`, then `·` in `Muted`, then `OrgName / ProjectName` in `Muted`
- Right: `[⚙]` and `[👤]` icon buttons, `Muted`, right-aligned
- Background: `Surface`
- Border: `Subtle` (rounded corners `╭╮╰╯`)

**Variants:**
- Dashboard: shows org + project
- Auth wizard: shows `· Sign in`
- Settings: shows `· Settings`
- Account switcher: shows `· Switch Account`

---

### 2. Sidebar

**Expanded (default, 12 cols wide):**

```
┌──────────┐
│          │
│ ▶ PRs  3 │
│   Work   │
│   Pipel  │
│   Releas │
│          │
└──────────┘
```

- Background: `Surface`
- Border right: single `│` in `Subtle`
- Active item: `▶` in `Purple`, name in `Text` bold, count badge in `Muted`
- Inactive items: name in `Muted`, no arrow
- Count: right-aligned within the 12-col space, in `Muted`
- Notification badge: `●` after count, in `Badge`
- Vertical padding: 1 blank row top and bottom
- Item height: 1 row per section

**Collapsed (3 cols wide):**

```
┌───┐
│   │
│▶P │
│ W │
│ I │
│ R │
│   │
└───┘
```

- Active item: `▶` + 1-char abbreviation (P/W/I/R), `Purple`
- Inactive: space + abbreviation, `Muted`
- Notification badge: replaces the space before the letter → `●W` in `Badge`

**Hidden:** sidebar not rendered; content takes full width.

---

### 3. Tab Bar

```
 [ My PRs  3 ]  [ Review  5 ]  [ Draft  1 ]  [ Waiting  0 ]
```

- Height: 1 row
- Active tab: label + count in `Text` bold, underlined, surrounded by `[` `]` in `Purple`
- Inactive tabs: label + count in `Muted`, surrounded by `[` `]` in `Subtle`
- Count of 0: still shown (so layout is stable)
- Loading state: count replaced with `…` in `Muted`
- Notification on a tab: count followed by `●` in `Badge`
- Horizontal padding: 1 space before each `[`, 1 space after each `]`

---

### 4. List Rows

**Column layout (PR list — default columns):**

```
 #      STATE    TITLE                          AUTHOR    REPO         AGE
 PR42   active   Fix authentication bug         alice     api-repo     2h
 PR38   active   Refactor UI components         bob       web-repo     1d
 PR31   draft    Add unit tests                 carol     api-repo     3d
```

- Column header row: all caps, `Muted`, `─` separator line below headers
- Minimum column widths enforced; `TITLE` column flexes to fill remaining space
- `AGE` is always right-aligned in its column

**Row states:**

| State | Visual treatment |
|---|---|
| Normal (unselected) | `Text` on `Base` |
| Selected (cursor on) | `Text` bold on `Overlay` |
| Draft | All text `Muted` (dimmed); `[draft]` tag in `Yellow` after title |
| Needs attention | `#ID` column text in `Red` |
| Needs my vote | `STATE` column text in `Yellow` |
| Failed policy | `#ID` column text in `Red` |
| Loading | Row replaced with spinner row (see Spinner component) |

**Row height:** 1 row per item (no wrapping).

**Selected row indicator:** full-width `Overlay` background; no additional left gutter mark.

---

### 5. Preview Pane

```
┌────────────────────────────────┐
│  PR 42                         │
│  Fix authentication bug        │
│                                │
│  alice/fix-auth → main         │
│  active · 2 hours ago          │
│                                │
│  2 reviewers · 1 approved      │
│  ✓ Build passing               │
│                                │
│  ↳ enter to open detail        │
└────────────────────────────────┘
```

- Width: 35% of content area (default, resizable)
- Border left: single `│` in `Subtle`; right/top/bottom: none (flush with terminal edge)
- Top padding: 1 blank row
- PR number: `Blue`, bold, large (well, as large as terminal allows — bold)
- Title: `Text`, bold, wraps if needed (max 2 lines)
- Branch info: `Muted`
- Status summary: semantic color per state
- Policy status: `✓` in `Green`, `✗` in `Red`, `⏳` in `Yellow`
- Footer hint: `↳ key hint` in `Muted` at bottom of pane
- Empty state: `No item selected` centered in `Muted`

**Resize indicator:** when `<`/`>` is pressed, briefly show the current width percentage in the status bar in `Accent`: `  Preview: 40%  `

---

### 6. Footer / Status Bar

```
  g+p PRs · g+w Work · g+i Pipelines · g+r Releases · p project · q quit
```

- Height: 1 row
- Background: `Surface`
- All text: `Muted`
- Key labels: `Accent` (the key itself), `Muted` (the description)
- Contextual: changes per active screen/mode

**Status bar states:**
- Normal: key hints for current screen
- Clipboard fallback: `  URL: https://dev.azure.com/...  ` in `Text` (replaces hints briefly)
- Error: `  ✗ Failed to load: <reason>  r to retry  ` in `Red` + `Muted`
- Resize: `  Preview: 40%  ` in `Accent` (2s then reverts)

---

### 7. Spinner (Loading State)

Loading state replaces the list content area:

```
  ⠸ Loading pull requests…
```

- Icon: braille spinner character cycling through `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`
- Color: `Accent`
- Text: `Muted`
- Centered vertically in the pane
- Per-pane: each list pane spins independently; preview pane is blank while list loads

---

### 8. Error State

```
  ✗ Failed to load pull requests

  401 Unauthorized
  Your session may have expired.

  r  retry
  A  switch account
```

- `✗` in `Red`, message in `Text`
- Detail line in `Muted`
- Action hints in `Accent` (key) + `Muted` (description)
- Centered vertically in the pane

---

### 9. Notification Badge

- Character: `●`
- Color: `Badge` (`#f38ba8`)
- Positions: after section name in sidebar (expanded), replaces leading space in sidebar (collapsed), after count in tab bar
- Cleared when user navigates to the section

---

### 10. Confirmation Modal

```
┌─────────────────────────────────┐
│                                 │
│  Reject PR 42?                  │
│                                 │
│  This will vote –10 (Reject)    │
│  on "Fix authentication bug"    │
│                                 │
│  [ y ] confirm    [ n ] cancel  │
│                                 │
└─────────────────────────────────┘
```

- Width: 40 cols, centered on screen
- Background: `Overlay`
- Border: `Subtle` (rounded corners)
- Title: `Text` bold
- Body: `Muted`
- `y` in `Red` (destructive), `n` in `Muted`
- All other keypresses dismissed as cancel

---

### 11. Help Overlay

Full-screen overlay showing all keybindings for the current screen:

```
╭─────────────────────────────────────────────────────────────────────╮
│  Keyboard shortcuts                                         esc close │
╰─────────────────────────────────────────────────────────────────────╯

  Navigation
  ──────────────────────────────────────────────────────────────────
  j / ↓           Move down
  k / ↑           Move up
  g + g           Jump to first item
  G               Jump to last item
  tab / ]         Next tab
  shift+tab / [   Previous tab

  Sections
  ──────────────────────────────────────────────────────────────────
  g + p           Pull Requests
  g + w           Work Items
  g + i           Pipelines
  g + r           Releases

  Actions
  ──────────────────────────────────────────────────────────────────
  enter / d       Open detail view
  o               Open in browser
  c               Copy URL
  r               Refresh current tab
  R               Refresh all tabs

  Layout
  ──────────────────────────────────────────────────────────────────
  \               Cycle sidebar (expanded → collapsed → hidden)
  < / >           Resize preview pane
  |               Reset preview pane width
  Z               Toggle zen mode

  any key         Close this overlay
```

- Background: `Base` full-screen
- Section headers: `Purple` bold + `─` rule line in `Subtle`
- Key column: `Accent` bold, fixed 18-col width left-aligned
- Description column: `Text`
- `esc close` in header: `Muted`, right-aligned

---

## Screen Designs

### Screen 1 — Auth Wizard, Step 1: Organization URL

```
╭──────────────────────────────────────────────────────────╮
│  ado-dash  ·  Sign in  (1/2)                             │
╰──────────────────────────────────────────────────────────╯




          Enter your Azure DevOps organization URL:

          ┌──────────────────────────────────────┐
          │ https://dev.azure.com/               │
          └──────────────────────────────────────┘

          e.g.  https://dev.azure.com/my-company
                https://my-company.visualstudio.com




  enter confirm · esc quit
```

- Input field: `Text` on `Overlay` background, `Purple` border when focused
- Placeholder text: `Muted`
- Example: `Muted`

**Validation error state:**

```
          ┌──────────────────────────────────────┐
          │ not-a-url                            │
          └──────────────────────────────────────┘
            ✗ Must be https://dev.azure.com/{org}
```

- Border becomes `Red`; error message in `Red` below the field

---

### Screen 2 — Auth Wizard, Step 2: Auth Method

```
╭──────────────────────────────────────────────────────────╮
│  ado-dash  ·  Sign in  (2/2)                             │
╰──────────────────────────────────────────────────────────╯

  Organization: dev.azure.com/my-company

  Azure CLI detected. Choose an authentication method:

  ▶  [ Azure CLI          ]   Browser-based login — recommended
     [ Personal Access Token ]   Env var reference
     [ Service Principal  ]   Client ID + secret (CI/scripts)




  j/k move · enter select · esc back · q quit
```

- `Azure CLI detected.` / `Azure CLI not found.` line: `Green` or `Yellow` accordingly
- Active item: `▶` in `Purple`, `[…]` border in `Purple`, description in `Text`
- Inactive items: `[…]` border in `Subtle`, description in `Muted`
- `esc back` returns to Step 1

**PAT sub-step (after selecting Access Token):**

```
  Enter the name of the environment variable holding your PAT:

  ┌──────────────────────────┐
  │ AZURE_DEVOPS_PAT         │
  └──────────────────────────┘

  The variable's value is your Personal Access Token.
  ado-dash never stores the token itself.

  enter confirm · esc back
```

---

### Screen 3 — Project Picker

```
╭──────────────────────────────────────────────────────────╮
│  ado-dash  ·  Select project                             │
╰──────────────────────────────────────────────────────────╯

  Organization: dev.azure.com/my-company

  ┌──────────────────────────────────────────────────────┐
  │ > █                                                  │
  └──────────────────────────────────────────────────────┘

  ▶  MyProject            Git · Agile
     BackendServices      Git · Scrum
     MobileApp            Git · Agile
     InfraOps             Git · CMMI
     DataPlatform         Git · Scrum




  j/k move · enter select · esc back
```

- Search input: `Accent` cursor, `Purple` border when focused
- Selected row: `Overlay` bg, `Text` bold
- Process type (Git · Agile): `Muted`
- Org URL: `Muted`

---

### Screen 4 — Main Dashboard (Full Layout)

```
╭─────────────────────────────────────────────────────────────────────────╮
│  ado-dash  my-company / MyProject                           [⚙]  [👤]  │
╰─────────────────────────────────────────────────────────────────────────╯
┌───────────┬─────────────────────────────────────────────────────────────┐
│           │ [ My PRs  3 ]  [ Review  5 ]  [ Draft  1 ]  [ Waiting  2 ] │
│ ▶ PRs   8 ├────────────────────────────────────────┬────────────────────┤
│   Work    │  #       STATE    TITLE          AUTHOR  AGE │               │
│   Pipel   │  ────────────────────────────────────────── │  PR 42         │
│   Releas  │▶ PR42    active   Fix auth bug   alice   2h  │  Fix auth bug  │
│           │  PR38    active   Refactor UI    bob     1d  │               │
│           │  PR31    draft    Add tests      carol   3d  │  alice → main  │
│           │  PR27    active   Update docs    dave    4d  │  active · 2h   │
│           │                                             │               │
│           │                                             │  ✓ 1 approved  │
│           │                                             │  ⏳ 2 waiting   │
│           │                                             │  ✓ Build pass  │
│           │                                             │               │
│           │                                             │  ↳ d to open   │
└───────────┴────────────────────────────────────────────┴────────────────┘
  g+p PRs · g+w Work · \ sidebar · Z zen · ? help · q quit
```

**Color annotations:**
- `PR42` ID column: `Blue`
- `active` state: `Green`
- `draft` state: `Yellow`
- `alice`, `bob` etc: `Muted`
- `2h`, `1d` age: `Muted`
- Preview pane title `PR 42`: `Blue` bold
- Preview pane subtitle: `Text` bold
- Preview pane metadata: `Muted`
- Preview pane `✓`: `Green`, `⏳`: `Yellow`
- Preview pane hint `↳ d to open`: `Muted`
- Tab bar active `[ My PRs  3 ]`: `Text` bold, underlined, `Purple` brackets
- Tab bar inactive: `Muted`, `Subtle` brackets

---

### Screen 5 — Main Dashboard (Zen Mode)

```
 [ My PRs  3 ]  [ Review  5 ]  [ Draft  1 ]  [ Waiting  2 ]
 ───────────────────────────────────────────────────────────
 #       STATE    TITLE                    AUTHOR  AGE
 ─────────────────────────────────────────────────────────────────────────
▶PR42    active   Fix authentication bug   alice   2h
 PR38    active   Refactor UI components   bob     1d
 PR31    draft    Add unit tests           carol   3d
 PR27    active   Update documentation     dave    4d




                                                  Z exit zen
```

- No header, no sidebar, no bordered footer
- Tab bar stays (needed for navigation)
- `Z exit zen` hint in bottom-right corner: `Muted`
- List pane fills full terminal width

---

### Screen 6 — Main Dashboard (Sidebar Collapsed)

```
╭───────────────────────────────────────────────────────────────────────╮
│  ado-dash  my-company / MyProject                         [⚙]  [👤]  │
╰───────────────────────────────────────────────────────────────────────╯
┌───┬───────────────────────────────────────────────────────────────────┐
│   │ [ My PRs  3 ]  [ Review  5 ]  [ Draft  1 ]                       │
│▶P ├──────────────────────────────────────────────────────────────────┤
│ W │  #       STATE    TITLE                       AUTHOR  AGE         │
│ I │  PR42    active   Fix authentication bug      alice   2h          │
│ R │▶ PR38    active   Refactor UI components      bob     1d          │
│   │  PR31    draft    Add unit tests              carol   3d          │
└───┴──────────────────────────────────────────────────────────────────┘
  g+p PRs · g+w Work · \ expand · q quit
```

- Preview pane hidden (under 139 cols threshold at this width)
- Sidebar shows 1-char abbreviations in `Purple` (active) / `Muted` (inactive)

---

### Screen 7 — PR Detail View

```
╭──────────────────────────────────────────────────────────────────────────╮
│  api-repo  ·  PR 42                                                      │
│                                                                          │
│  Fix authentication bug in OAuth flow                                    │
╰──────────────────────────────────────────────────────────────────────────╯
 [ Overview ]  [ Reviewers ]  [ Checks ]  [ Files ]  [ Threads ]

╭──────────────────────────────────────────────────────────────────────────╮
│  State: Active      Draft: No      Auto-complete: Off                    │
│  alice/fix-auth-bug → main                                               │
│  Created 2 hours ago by Alice Smith                                      │
│                                                                          │
│  Description ──────────────────────────────────────────────────────────  │
│                                                                          │
│  This PR fixes the OAuth token refresh flow that was causing             │
│  intermittent 401s on long sessions. The root cause was a race           │
│  condition in the token cache — if two requests fired simultaneously     │
│  and both found an expired token, both would try to refresh and the      │
│  second refresh would overwrite the first with a slightly stale token.   │
│                                                                          │
│  Changes:                                                                │
│  - Added mutex around token refresh in `TokenCache.Get()`               │
│  - Added retry with exponential backoff on 401                           │
│  - Added test coverage for the concurrent refresh scenario               │
│                                                                          │
╰──────────────────────────────────────────────────────────────────────────╯
  [/] h/l tabs · j/k scroll · a approve · x reject · o browser · b back
```

**Color annotations:**
- `PR 42` in header: `Blue` bold
- Title in header: `Text` bold
- Active tab `[ Overview ]`: `Purple` underline, `Text` bold
- Inactive tabs: `Subtle` brackets, `Muted` text
- `State: Active` label: `Muted`, value: `Green`
- `alice/fix-auth-bug → main`: `Blue` → `Muted` arrow → `Text`
- `Created … by …`: `Muted`
- Description rendered as markdown (bold, bullets, code spans in `Accent`)
- Scrollbar indicator (if content overflows): single `│` on right edge in `Subtle`

---

### Screen 8 — PR Detail, Reviewers Tab

```
 [ Overview ]  [ Reviewers ]  [ Checks ]  [ Files ]  [ Threads ]

╭──────────────────────────────────────────────────────────────────────────╮
│  Reviewers (2 of 2 responded)                                            │
│  ─────────────────────────────────────────────────────────────────────   │
│                                                                          │
│  ✓  Alice Smith          Approved                    required            │
│  –  Bob Johnson          No vote                     optional            │
│                                                                          │
│  ─────────────────────────────────────────────────────────────────────   │
│  My vote:  No vote                                                       │
│                                                                          │
│  a approve  A approve+suggestions  x reject  w wait-for-author          │
╰──────────────────────────────────────────────────────────────────────────╯
```

**Vote icons and colors:**
- `✓` Approved: `Green`
- `~` Approved with suggestions: `Yellow`
- `✗` Rejected: `Red`
- `⏳` Wait for author: `Yellow`
- `–` No vote: `Muted`

**Badge colors:**
- `required`: `Red` dim
- `optional`: `Muted`

---

### Screen 9 — PR Detail, Checks Tab

```
 [ Overview ]  [ Reviewers ]  [ Checks ]  [ Files ]  [ Threads ]

╭──────────────────────────────────────────────────────────────────────────╮
│  Policy Evaluations                                                      │
│  ─────────────────────────────────────────────────────────────────────   │
│                                                                          │
│  ✓  Minimum 2 reviewers            Approved          blocking            │
│  ✓  CI Build                       Passed            blocking            │
│  ✗  Work item linked               Not linked        blocking            │
│  ✓  No active comments             Satisfied         blocking            │
│  ✓  Branch policies: main          Compliant         blocking            │
│                                                                          │
╰──────────────────────────────────────────────────────────────────────────╯
```

- `✓` in `Green`, `✗` in `Red`, `⏳` in `Yellow`
- Status text: same color as icon
- `blocking` badge: `Red` dim
- `non-blocking` badge: `Muted`

---

### Screen 10 — Settings Page

```
╭──────────────────────────────────────────────────────────╮
│  ado-dash  ·  Settings                                   │
╰──────────────────────────────────────────────────────────╯

  Refresh Interval     [ 120s          ▼ ]   (30s | 60s | 120s | 300s)
  Preview Pane         [ Right         ▼ ]   (Right | Bottom | Off)
  Preview Pane Width   [ 35%           ▼ ]   (use < / > in dashboard)
  Sidebar              [ Expanded      ▼ ]   (Expanded | Collapsed | Hidden)
  Theme                [ Catppuccin Mocha ▼ ]
  Open PR in           [ Detail        ▼ ]   (Detail | Browser)
  Notifications        [ Off           ▼ ]   (Off | Terminal Bell | OS)
  Default Team         [ Platform Team ▼ ]   (re-opens team picker)

  ──────────────────────────────────────────────────────────
  Config   ~/.config/ado-dash/config.yml     [ open in $EDITOR ]

  j/k move · enter/space toggle · b/esc back
```

- Focused row: full-width `Overlay` background
- Label column: `Text`, fixed 22-col width
- Value `[…▼]`: `Accent` brackets + `Purple` ▼, value text in `Text`
- Hint column: `Muted`
- Separator `──`: `Subtle`
- `[ open in $EDITOR ]`: `Accent` brackets, `Text` label

---

### Screen 11 — Account Switcher

```
╭──────────────────────────────────────────────────────────╮
│  ado-dash  ·  Switch Account                             │
╰──────────────────────────────────────────────────────────╯

  ▶  Work                                                  
     dev.azure.com/my-company  ·  Alice Smith  ·  az-cli  
                                                           
     Personal                                             
     dev.azure.com/side-proj   ·  alice@gmail  ·  PAT     
                                                           
     + Add account                                        




  j/k move · enter switch · d delete · b/esc back
```

- Active item: `▶` in `Purple`, profile name in `Text` bold
- Inactive items: profile name in `Text`
- Metadata line (org, identity, auth): `Muted`, indented 5 chars
- `+ Add account`: `Accent`
- Currently active profile: `▶` instead of space

---

### Screen 12 — Inline Team Picker

```
╭──────────────────────────────────────────────────────────╮
│  Select your primary team                                │
│  Used for Current Sprint queries                         │
╰──────────────────────────────────────────────────────────╯

  ▶  Platform Team
     Mobile Team
     Backend Services
     DevOps Guild



  j/k move · enter select
  (saved to config — won't be asked again)
```

- Appears as an overlay on top of the work item list pane
- Width: 50 cols, centered in content area
- Title: `Text` bold; subtitle: `Muted`
- Active item: `▶` in `Purple`, `Text` bold
- Inactive: `Muted`
- Footer hint: `Muted`

---

## Interaction Patterns

### Loading → Content Transition

1. User switches tab or section
2. Immediately: spinner appears in list pane, preview pane shows blank
3. Data arrives: spinner disappears, list rows appear (no flash/redraw artifact — Bubble Tea handles this)
4. First row auto-selected; preview pane populates

### Vote Action Flow

1. User presses `a` on a PR in detail view
2. Status bar briefly shows: `  Voting… ` with `Accent` spinner
3. Success: Reviewers tab updates; status bar shows `  ✓ Voted: Approved  ` in `Green` for 2s then reverts
4. Failure: status bar shows `  ✗ Vote failed: <reason>  ` in `Red`

### Destructive Action Flow (e.g. Reject)

1. User presses `x`
2. Confirmation modal appears (see component above)
3. `y` → fires, same flow as above
4. `n` / any other key → modal dismissed, no action

### Resize Flow

1. User presses `>`
2. Preview pane width increases by 5%
3. Status bar immediately shows `  Preview: 40%  ` in `Accent`
4. After 2 seconds status bar reverts to key hints
5. New width written to `state.json` on each keypress

### Auth Expiry Mid-Session

1. Any API call returns 401
2. `ado-dash` attempts silent token refresh (azure-cli) or re-reads PAT env var
3. If refresh succeeds: the failed request is retried transparently
4. If refresh fails: a small banner appears at top of current screen:
   ```
   ┌─────────────────────────────────────────────┐
   │  ⚠  Session expired — re-authenticating…   │
   └─────────────────────────────────────────────┘
   ```
   Banner in `Yellow` on `Surface`. Auth wizard pushes on top if re-auth fails.

---

## Density & Column Widths

### PR List Default Column Widths

| Column | Width | Alignment | Color |
|---|---|---|---|
| `#` (ID) | 7 | Left | `Blue` |
| `STATE` | 10 | Left | Semantic |
| `TITLE` | Flex (min 20) | Left | `Text` |
| `AUTHOR` | 14 | Left | `Muted` |
| `REPO` | 14 | Left | `Muted` |
| `AGE` | 5 | Right | `Muted` |

### Work Item List Default Column Widths

| Column | Width | Alignment | Color |
|---|---|---|---|
| `#` (ID) | 7 | Left | `Blue` |
| `TYPE` | 8 | Left | `Muted` |
| `STATE` | 12 | Left | Semantic |
| `TITLE` | Flex (min 20) | Left | `Text` |
| `ASSIGNED` | 14 | Left | `Muted` |
| `AGE` | 5 | Right | `Muted` |

### Pipeline List Default Column Widths

| Column | Width | Alignment | Color |
|---|---|---|---|
| `NAME` | 24 | Left | `Text` |
| `STATUS` | 12 | Left | Semantic |
| `BRANCH` | 20 | Left | `Muted` |
| `COMMIT` | 9 | Left | `Blue` |
| `TRIGGER` | 14 | Left | `Muted` |
| `DURATION` | 8 | Right | `Muted` |
| `AGE` | 5 | Right | `Muted` |

---

## State String Mappings

### PR State → Color + Label

| ADO value | Display | Color |
|---|---|---|
| `active` | `active` | `Green` |
| `completed` | `merged` | `Purple` |
| `abandoned` | `closed` | `Muted` |

### PR Vote → Icon + Label + Color

| Vote | Icon | Label | Color |
|---|---|---|---|
| `10` | `✓` | `Approved` | `Green` |
| `5` | `~` | `Suggestions` | `Yellow` |
| `0` | `–` | `No vote` | `Muted` |
| `-5` | `⏳` | `Waiting` | `Yellow` |
| `-10` | `✗` | `Rejected` | `Red` |

### Build Status → Color + Label

| statusFilter | resultFilter | Display | Color |
|---|---|---|---|
| `inProgress` | — | `running` | `Yellow` |
| `notStarted` | — | `queued` | `Muted` |
| `cancelling` | — | `cancelling` | `Yellow` |
| `completed` | `succeeded` | `succeeded` | `Green` |
| `completed` | `partiallySucceeded` | `partial` | `Yellow` |
| `completed` | `failed` | `failed` | `Red` |
| `completed` | `canceled` | `canceled` | `Muted` |

---

## Border & Line Characters

| Element | Characters | Color |
|---|---|---|
| Panel borders | `─ │ ╭ ╮ ╰ ╯` | `Subtle` |
| Inner dividers | `─` | `Subtle` |
| Active/focus border | `─ │ ╭ ╮ ╰ ╯` | `Purple` |
| Section separator in detail view | `──────` (full width) | `Subtle` |
| Tab bracket (active) | `[` `]` | `Purple` |
| Tab bracket (inactive) | `[` `]` | `Subtle` |
| Sidebar active gutter | `▶` | `Purple` |
| List item cursor | `▶` | `Purple` |
| Settings value bracket | `[` `▼` `]` | `Accent` / `Purple` |
| Spinner | `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` | `Accent` |
| Check mark (pass) | `✓` | `Green` |
| Cross (fail) | `✗` | `Red` |
| Dash (no vote / neutral) | `–` | `Muted` |
| Waiting | `⏳` | `Yellow` |
| Badge | `●` | `Badge` |
| Arrow (branch) | `→` | `Muted` |
