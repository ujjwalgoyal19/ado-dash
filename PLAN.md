# ado-dash — Implementation Plan

## Vision

A gh-dash-style terminal dashboard for Azure DevOps. Navigation modeled after vim/gh-dash conventions. Config works like VS Code: a YAML file for full control, a TUI settings page for common preferences, and smart defaults everywhere.

---

## Foundation Decision

**Stack: Go + Bubble Tea.** The TypeScript/Bun prototype is discarded. Rationale:

- gh-dash, lazygit, k9s, and the entire Charm ecosystem are Go — ado-dash fits directly into this lineage
- Bubble Tea owns the render loop, resize, and alternate screen — no manual ANSI cursor math
- `go build` produces a fully static binary with no runtime dependency
- GoReleaser handles multi-arch builds + Homebrew tap automatically
- Lip Gloss, Bubbles, and Glamour are purpose-built for exactly this kind of tool

### Core dependencies

| Package | Purpose |
|---|---|
| `github.com/charmbracelet/bubbletea` | Event loop, model/update/view |
| `github.com/charmbracelet/lipgloss` | Styling — colors, borders, layout |
| `github.com/charmbracelet/bubbles` | List, textinput, spinner, key, viewport |
| `github.com/charmbracelet/glamour` | Markdown rendering (PR descriptions, wiki) |
| `github.com/spf13/cobra` | CLI parsing (`--config`, `--doctor`, `--version`) |
| `gopkg.in/yaml.v3` | Config file parsing |
| `github.com/goreleaser/goreleaser` | Multi-arch builds + Homebrew tap |

---

## Architecture Overview

### Screen / State Machine

Navigation is a **screen stack**, not a linear chain. Any screen can push another on top and pop back to its caller. Settings and Account Switcher are pushed from Main Dashboard, not from Detail View.

```
  ┌─────────────────┐
  │   Auth Wizard   │  first run or auth-expired
  └────────┬────────┘
           │ auth resolved
  ┌────────▼────────┐
  │ Project Picker  │  first run, or `p` pushed from Main Dashboard
  └────────┬────────┘
           │ project selected
  ┌────────▼────────────────────────────────────┐
  │              Main Dashboard                  │◄──────────────┐
  │  sidebar + sub-tabs + list + preview pane   │               │
  └──┬──────────┬────────────┬──────────────────┘               │
     │ enter/d  │ , or ⚙    │ A or 👤                          │
     ▼          ▼            ▼                                   │
  Detail     Settings    Account                                 │
  View       Page        Switcher                                │
     │          │            │                                   │
     └──────────┴────────────┘  b/esc pops back to Main Dashboard┘
```

**All screens share these transient overlay states (not separate screens):**

| State | Trigger | Description |
|---|---|---|
| Loading / error / retry | data fetch | per-pane spinner + error + `r` to retry |
| Auth-expired | 401 on any request | re-auths silently; if fails, pushes Auth Wizard then returns |
| Help overlay | `?` | full-screen key reference, any key dismisses |
| Search / filter input | `/` | inline input field, keys captured until `esc` or `enter` |
| Confirmation prompt | destructive actions (reject, delete) | modal `y/n` before firing |
| Team picker | first `@CurrentIteration` use | inline picker, saves selection, never shown again |
| External process | `$EDITOR`, browser, clipboard, `az login` | raw-mode suspended, restored on exit |


---

## Screen Designs

### Screen 1 — Auth Wizard

Shown on first run or when all auth methods fail.

```
╭──────────────────────────────────────────────────────╮
│  ado-dash  ·  Sign in                                │
╰──────────────────────────────────────────────────────╯

  Azure CLI (az login) was not found on this machine.
  Choose an authentication method:

  ▶  [ Azure CLI    ]   Browser-based login — recommended
     [ Access Token ]   Paste a Personal Access Token
     [ Service Principal ]  Client ID + Secret (CI/scripts)

  j/k move · enter select · q quit
```

**Logic:**

1. On startup, if a saved auth method exists in `config.yml`, attempt a real ADO validation request (`GET /_apis/connectionData?api-version=7.1-preview.1`) — not just token minting. This proves the identity can reach the configured org.
2. If validation succeeds → proceed directly to Main Dashboard (or Project Picker if no saved project).
3. If validation fails (401, 403, network error) → show Auth Wizard with the failure reason shown.
4. If no saved auth → show Auth Wizard.

**Auth methods:**

- `azure-cli` — calls `az account get-access-token --resource <ADO_RESOURCE_ID>`, tokens cached with 5-min expiry buffer, auto-refreshed. Validates against ADO on first use.
- `pat` — wizard asks for the **env var name** that holds the token (not the token itself), e.g. `AZURE_DEVOPS_PAT`. The wizard reads the env var to verify it's set and non-empty, then makes a test ADO request to confirm it works before saving. Raw token values are rejected if entered directly. Config file must be `0600`; warn if not.
- `service-principal` — clientId + tenantId in config, client secret via env var reference. Uses client-credentials OAuth flow to get an ADO bearer token.

Auth selection and env var name are written to `~/.config/ado-dash/config.yml` by the wizard.

---

### Screen 2 — Project Picker

Shown on first run after auth, or when switching project (triggered by `p` from main dashboard).

```
╭──────────────────────────────────────────────────────╮
│  ado-dash  ·  Select project                         │
╰──────────────────────────────────────────────────────╯

  Organization: dev.azure.com/my-org

  > █                          (fuzzy search)

  ▶  MyProject          Git · Agile
     BackendServices    Git · Scrum
     MobileApp          Git · Agile
     InfraOps           Git · CMMI
     DataPlatform       Git · Scrum

  j/k move · enter select · / search · esc back
```

**Logic:**

- Calls `GET /_apis/projects?api-version=7.1` to list all projects.
- Fuzzy search filters inline as user types (no Enter needed).
- Selection is saved to `state.json` under `profiles[name].defaultProject` — not to `config.yml`.
- After selection → go to Main Dashboard.
- `p` from main dashboard re-opens this screen (non-destructive — can cancel with `esc`).

---

### Screen 3 — Main Dashboard

The persistent home screen. Left sidebar + content area.

```
╭─────────────────────────────────────────────────────────────────────╮
│  ado-dash  MyOrg / MyProject                            [⚙]  [👤]  │
╰─────────────────────────────────────────────────────────────────────╯
┌─────────┬───────────────────────────────────────────────────────────┐
│         │  [ My PRs  3 ]  [ Review  5 ]  [ Draft  1 ]              │
│ ▶ PRs   │ ┌─────────────────────────────────────┬─────────────────┐ │
│  Work   │ │ #     STATE   TITLE           WHO AGE│                 │ │
│  Pipel  │ │▶PR42  active  Fix auth bug   alice 2h│  PR 42          │ │
│  Releas │ │ PR38  active  Refactor UI    bob   1d│  Fix auth bug   │ │
│         │ │ PR31  draft   Add tests      carol 3d│                 │ │
│         │ │                                      │  alice → main   │ │
│         │ │                                      │  active · 2h    │ │
│         │ │                                      │  ↳ v to review  │ │
│         │ └─────────────────────────────────────┴─────────────────┘ │
└─────────┴───────────────────────────────────────────────────────────┘
  g+p PRs · g+w Work · g+i Pipelines · g+r Releases · p project · q quit
```

**Sidebar sections (fixed, always present):**

- `PRs` — Pull Requests
- `Work` — Work Items
- `Pipel` — Pipelines & Builds
- `Releas` — Releases & Approvals

**Sub-tabs:** configurable per section in YAML (see Config section below). Defaults provided.

**Layout proportions (default):**

- Sidebar: fixed ~12 chars wide (collapsible to 3 chars)
- List pane: grows to fill space left by preview pane
- Preview pane: default 35% of content area width (resizable)

---

### Sidebar states

Three states, cycled with `\`:

```
Expanded (default)     Collapsed               Hidden
┌──────────┐           ┌───┐                   (no sidebar)
│ ▶ PRs    │           │▶P │
│   Work   │           │ W │
│   Pipel  │           │ I │
│   Releas │           │ R │
└──────────┘           └───┘
  12 chars              3 chars
```

- **Expanded**: full section names, item counts shown
- **Collapsed**: 1-char abbreviation per section (P/W/I/R), still shows active indicator
- **Hidden**: sidebar entirely removed; section switching via `g+*` keys only

State is persisted in config as `sidebar: expanded | collapsed | hidden`.

---

### Zen Mode

Toggle with `Z` (shift+z). Hides everything except the core content pane for the current screen.

**On Main Dashboard:**
```
(no header, no sidebar, no footer)
 [ My PRs  3 ]  [ Review  5 ]  [ Draft  1 ]
┌─────────────────────────────┬─────────────────┐
│ #     STATE  TITLE    WHO AGE│                 │
│▶PR42  active Fix...  alice 2h│  PR 42          │
│ PR38  active Refac.. bob   1d│  Fix auth bug   │
└─────────────────────────────┴─────────────────┘
Z exit zen
```

**On Detail View:**
```
(no header chrome, no footer)
 [ Overview ]  [ Reviewers ]  [ Checks ]  [ Files ]  [ Threads ]
╭──────────────────────────────────────────────────────────────╮
│  ...full tab content, more vertical space...                 │
╰──────────────────────────────────────────────────────────────╯
```

Sub-tabs bar stays in zen mode (needed for navigation). Only the outer chrome (header bar, sidebar, status footer) is hidden.

---

### Preview Pane Resize

The preview pane width is a percentage of the content area. Resize live with:

| Key        | Action                          |
| ---------- | ------------------------------- |
| `>`        | Grow preview pane by 5%         |
| `<`        | Shrink preview pane by 5%       |
| `\|`       | Reset preview pane to default   |

Bounds: minimum 20%, maximum 60%. Steps of 5%.

The current width is saved to `state.json` as `ui.previewPaneWidth: 35` so it persists across sessions.

---

### Screen 4 — Detail View (PR)

Triggered by `enter` or `d` on a PR. Full screen.

```
╭───────────────────────────────────────────────────────────────────╮
│  my-repo · PR 42                                                  │
│                                                                   │
│  Fix authentication bug in OAuth flow                             │
╰───────────────────────────────────────────────────────────────────╯
 [ Overview ]  [ Reviewers ]  [ Checks ]  [ Files ]  [ Threads ]

╭───────────────────────────────────────────────────────────────────╮
│  State: Active   Draft: No   Auto-complete: Off                   │
│  alice/fix-auth-bug → main                                        │
│  Created 2 hours ago by Alice Smith                               │
│                                                                   │
│  Description:                                                     │
│  This PR fixes the OAuth token refresh flow that was causing      │
│  intermittent 401s on long sessions. Adds retry logic with        │
│  exponential backoff...                (scrollable)               │
╰───────────────────────────────────────────────────────────────────╯

  [/]/h/l tabs · j/k scroll · b/esc back · o browser · a approve · q quit
```

**Tabs:**

1. **Overview** — title, description, state, draft, branches, created by, auto-complete settings
2. **Reviewers** — vote per reviewer (Approved ✓ / Waiting – / Rejected ✗), required/optional badge
3. **Checks** — policy evaluations (min reviewers, required build, comment-required, work-item-link, etc.)
4. **Files** — list of changed files per latest iteration (already built, moved here)
5. **Threads** — discussion threads + comments (already built, moved here)

**Actions from detail view:**

- `a` — vote approve
- `A` — approve with suggestions
- `x` — reject
- `w` — vote wait-for-author
- `o` — open in browser
- `v` — open browser review (files view)
- `c` — copy PR URL to clipboard

---

### Screen 5 — Detail View (Pipeline Run)

```
╭───────────────────────────────────────────────────────────────────╮
│  CI Pipeline · Run #1042  ·  feature/fix-auth                     │
╰───────────────────────────────────────────────────────────────────╯
 [ Summary ]  [ Stages ]  [ Logs ]  [ Artifacts ]

╭───────────────────────────────────────────────────────────────────╮
│  Status: Succeeded   Duration: 4m 32s   Triggered by: PR push     │
│  Branch: feature/fix-auth   Commit: a3f29b1                       │
│                                                                   │
│  Stages:                                                          │
│  ✓ Build          1m 12s                                          │
│  ✓ Test           2m 08s  (142 passed, 0 failed)                  │
│  ✓ Security Scan  1m 12s                                          │
╰───────────────────────────────────────────────────────────────────╯
  [/]/h/l tabs · j/k scroll · b back · n queue new run · q quit
```

---

### Screen 6 — Settings Page

Accessible via `⚙` icon or `,` key. Shows preferences that can be changed from the TUI. More advanced config (column customization, keybinding overrides) must be done in the YAML file.

```
╭──────────────────────────────────────────────────────╮
│  ado-dash  ·  Settings                               │
╰──────────────────────────────────────────────────────╯

  Refresh Interval     [ 120s     ▼ ]
  Preview Pane         [ Right    ▼ ]   (Right | Bottom | Off)
  Preview Pane Width   [ 35%      ▼ ]   (use < / > in dashboard to resize)
  Sidebar              [ Expanded ▼ ]   (Expanded | Collapsed | Hidden)
  Theme                [ Catppuccin Mocha ▼ ]   (see config for all options)
  Open PR in           [ Detail   ▼ ]   (Detail | Browser)
  Notifications        [ Off      ▼ ]   (Off | Terminal Bell | OS)
  Config file location   ~/.config/ado-dash/config.yml  [ open in $EDITOR ]

  j/k move · enter/space toggle · b/esc back
```

---

### Screen 7 — Account Switcher

Accessible via `👤` icon or `A` key. Lists named profiles (each = org URL + auth method).

```
╭──────────────────────────────────────────────────────╮
│  ado-dash  ·  Switch Account                         │
╰──────────────────────────────────────────────────────╯

  ▶  Work (dev.azure.com/my-company)     Alice Smith  · az-cli
     Personal (dev.azure.com/side-proj)  alice@gmail  · PAT
     + Add account

  j/k move · enter switch · d delete · b/esc back
```

Multiple accounts are stored as named profiles in config:

```yaml
profiles:
  - name: Work
    organizationUrl: https://dev.azure.com/my-company
    auth:
      type: azure-cli

  - name: Personal
    organizationUrl: https://dev.azure.com/side-proj
    auth:
      type: pat
      env: ADO_PERSONAL_PAT

activeProfile: Work
```

---

## Keybindings

### Global (any screen)

| Key            | Action        |
| -------------- | ------------- |
| `q` / `ctrl+c` | Quit          |
| `?`            | Help overlay  |
| `Z`            | Toggle zen mode (hide chrome) |

### Main Dashboard only

| Key | Action                                       |
| --- | -------------------------------------------- |
| `\` | Cycle sidebar: expanded → collapsed → hidden |
| `g` → `g` | Jump to first item in list (vim `gg`) |
| `G`        | Jump to last item in list (vim `G`)   |

### Main Dashboard — Navigation

| Key               | Action                                  |
| ----------------- | --------------------------------------- |
| `g` → `p`         | Jump to Pull Requests section           |
| `g` → `w`         | Jump to Work Items section              |
| `g` → `i`         | Jump to Pipelines section               |
| `g` → `r`         | Jump to Releases section                |
| `tab` / `]`       | Next sub-tab                            |
| `shift+tab` / `[` | Previous sub-tab                        |
| `j` / `↓`         | Move down in list                       |
| `k` / `↑`         | Move up in list                         |
| `enter` / `d`     | Open detail view                        |
| `o`               | Open selected item in browser           |
| `r`               | Refresh current sub-tab                 |
| `R`               | Refresh all sub-tabs in current section |
| `p`               | Open Project Picker                     |
| `A`               | Open Account Switcher                   |
| `,`               | Open Settings                           |
| `c`               | Copy URL of selected item to clipboard  |
| `>`               | Grow preview pane by 5%                 |
| `<`               | Shrink preview pane by 5%               |
| `\|`              | Reset preview pane to default width     |

**Leader key logic:** pressing `g` sets `awaitingLeader = true` for 500ms. If a second key arrives within that window, the pair is processed as a command (`g+p`, `g+w`, `g+i`, `g+r`, `g+g`). If the window expires, `g` resets silently. When an input field is focused (search, filter), all keys bypass the leader logic. `G` (capital) is a direct binding (no leader needed) for jump-to-last, preserving the vim `gg`/`G` idiom.

### PR Detail View

| Key         | Action                               |
| ----------- | ------------------------------------ |
| `[` / `]`   | Switch tabs                          |
| `h` / `l`   | Switch tabs                          |
| `j` / `k`   | Scroll content / move file selection |
| `a`         | Vote: Approve                        |
| `A`         | Vote: Approve with suggestions       |
| `x`         | Vote: Reject                         |
| `w`         | Vote: Wait for author                |
| `o`         | Open PR in browser                   |
| `v`         | Open browser review (files view)     |
| `c`         | Copy PR URL                          |
| `b` / `esc` | Back to list                         |

### Configurable keybindings (advanced)

Any action can be rebound in `config.yml`:

```yaml
keybindings:
  copyUrl: "y" # override default 'c'
  approve: "ctrl+a"
  openBrowser: "o"
```

---

## Configuration

Config model: **YAML file + TUI settings page** (like VS Code).

- **Config file** (`~/.config/ado-dash/config.yml` or `.ado-dash.yml` in CWD): full control over profiles, sections, columns, keybindings, display options.
- **TUI Settings page**: subset of preferences that are safe to toggle in the UI — refresh interval, theme, preview pane position, open behavior, notifications.
- **Defaults**: every setting has a sensible default; config file is optional.
- **First-run wizard**: writes auth choice to `config.yml`. Sets `activeProfile` in `state.json`. User never needs to hand-edit YAML to get started.

### Full config schema (target)

The config file is split into two parts — written to different locations to avoid churning hand-edited YAML:

- **`~/.config/ado-dash/config.yml`** — user-authored: profiles, sections, columns, keybindings, display preferences. Never auto-written except by the first-run wizard.
- **`~/.config/ado-dash/state.json`** — runtime state only: last selected project per profile, sidebar width, preview pane width, sidebar collapse state. Written by TUI on every change. Never needs to be hand-edited.

```yaml
# ~/.config/ado-dash/config.yml  (user-authored — never auto-rewritten after first run)

# profiles: org URL + auth method only. No runtime state here.
profiles:
  - name: Work
    organizationUrl: https://dev.azure.com/my-company
    auth:
      type: azure-cli          # azure-cli | pat | service-principal
      # tenant: optional-tenant-id

  - name: Personal
    organizationUrl: https://dev.azure.com/side-proj
    auth:
      type: pat
      env: ADO_PERSONAL_PAT    # env var name — never paste the token inline

# activeProfile, defaultProject, defaultTeamId → state.json (see below)

# Per-section sub-tab definitions (like gh-dash sections)
sections:
  pullRequests:
    - title: My PRs
      role: author
      status: active
      # repositories: [repo-name-1, repo-name-2]
      # NOTE: ADO PR API only accepts one searchCriteria.repositoryId per request.
      # Multiple repos → ado-dash resolves names to IDs at startup, fans out one
      # request per repo, then merges + sorts results by updatedAt and dedupes.
      # Pagination runs independently per repo and stops when the section limit is met.
      columns:
        - { field: id,        width: 8  }
        - { field: state,     width: 10 }
        - { field: title,     width: 40 }
        - { field: author,    width: 18 }
        - { field: repo,      width: 14 }
        - { field: updatedAt, width: 10 }

    - title: Review Requested
      role: reviewer
      status: active
      # repositories: [api-repo]   # fan-out applies here too

    - title: Draft
      role: author
      status: active
      draft: true

  workItems:
    - title: Assigned to Me
      wiql: >-
        SELECT [System.Id], [System.Title], [System.State],
               [System.WorkItemType], [System.AssignedTo], [System.ChangedDate]
        FROM WorkItems
        WHERE [System.AssignedTo] = @Me
          AND [System.State] <> 'Closed'
        ORDER BY [System.ChangedDate] DESC

    - title: Current Sprint
      # iteration path is resolved at runtime via teams API and substituted
      # into the WIQL — see "Team Auto-Detection" section
      iterationResolved: true
      wiql: >-
        SELECT [System.Id], [System.Title], [System.State], [System.AssignedTo]
        FROM WorkItems
        WHERE [System.TeamProject] = @project
          AND [System.IterationPath] = '{currentIterationPath}'
          AND [System.State] NOT IN ('Closed', 'Resolved', 'Done')
        ORDER BY [Microsoft.VSTS.Common.Priority] ASC

  pipelines:
    - title: Recent Runs
      status: all # all | inProgress | succeeded | failed | canceled
      branch: main

    - title: Failing
      status: failed

  releases:
    - title: Pending Approval
      status: pending

# Display preferences (also editable from TUI Settings page)
defaults:
  refreshSeconds: 120
  limit: 30
  theme: catppuccin-mocha   # see Themes section for all options
  previewPane: right    # right | bottom | off
  openPrIn: detail      # detail | browser
  notifications: off    # off | bell | os
```

```json
{
  "_comment": "~/.config/ado-dash/state.json — auto-written by TUI, do not hand-edit",
  "activeProfile": "Work",
  "profiles": {
    "Work": {
      "defaultProject": "MyProject",
      "defaultTeamId": "abc-123-def",
      "defaultTeamName": "Platform Team"
    }
  },
  "ui": {
    "sidebar": "expanded",
    "previewPaneWidth": 35
  }
}
```

```yaml
# ~/.config/ado-dash/config.yml (continued)

# Keybinding overrides (optional — action names are stable across versions)
keybindings:
  copyUrl: "y"          # default: c
  approve: "ctrl+a"     # default: a
  openBrowser: "o"      # default: o (no-op override example)
```

---

## File Structure (target)

```
cmd/
  ado-dash/
    main.go                    # Cobra root command, wires everything

internal/
  auth/
    provider.go                # AuthProvider interface
    azurecli.go                # az account get-access-token
    pat.go                     # env-var PAT
    serviceprincipal.go        # client credentials OAuth

  config/
    config.go                  # YAML loader + validator (gopkg.in/yaml.v3)
    state.go                   # state.json reader/writer (encoding/json)
    defaults.go                # built-in default sections + queries

  client/
    client.go                  # ADO REST client, auth header injection
    pullrequests.go
    workitems.go
    pipelines.go               # Build API (dev.azure.com)
    releases.go                # Release API (vsrm.dev.azure.com)
    projects.go
    teams.go

  ui/
    app.go                     # Root Bubble Tea model — screen stack
    keys.go                    # Shared key bindings (bubbles/key)
    styles.go                  # Lip Gloss style definitions

    screens/
      authwizard.go
      projectpicker.go
      dashboard.go             # Sidebar + sub-tabs + list + preview
      detail.go                # PR / Work Item / Pipeline detail
      settings.go
      accountswitcher.go

    components/
      sidebar.go
      tabbar.go
      listview.go              # Configurable column table (wraps bubbles/list)
      previewpane.go           # bubbles/viewport
      fuzzysearch.go           # bubbles/textinput

    detail_tabs/
      pr/
        overview.go
        reviewers.go
        checks.go
        files.go
        threads.go
      workitem/
        overview.go
        relations.go
        history.go
      pipeline/
        summary.go
        stages.go
        logs.go
        artifacts.go

.goreleaser.yaml               # Multi-arch builds + Homebrew tap
go.mod
go.sum
```

---

## Build Phases

### Phase 0 — Go Scaffold

Archive the TypeScript prototype for reference. Start fresh in Go.

- `go mod init github.com/ujjwal-goyal/ado-dash`
- Add all core dependencies (`bubbletea`, `lipgloss`, `bubbles`, `glamour`, `cobra`, `yaml.v3`)
- `cmd/ado-dash/main.go` — Cobra root with `--config`, `--doctor`, `--version` flags
- `internal/config` — YAML loader + validator, `state.json` reader/writer, default sections
- `internal/auth` — `AuthProvider` interface + azure-cli + PAT + service-principal impls
- `internal/client` — ADO REST client: `GET /_apis/connectionData` to validate, `listProjects`
- `internal/ui/app.go` — root Bubble Tea model with screen stack, `WindowSizeMsg` handling
- Auth Wizard screen (silent azure-cli detection → validate against ADO → fallback to wizard)
- Project Picker screen (fuzzy filter via `bubbles/textinput`)
- `.goreleaser.yaml` — builds for `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64` + Homebrew tap

### Phase 1 — Main Dashboard Shell

- `MainDashboard` model: sidebar + tab bar + list pane + preview pane
- Lip Gloss layout: sidebar fixed width, preview pane flex width (stored in `state.json`)
- Sidebar collapse (`\`): expanded / collapsed (icon only) / hidden — `WindowSizeMsg` triggers auto-collapse at thresholds
- Zen mode (`Z`): hides sidebar + header + footer
- Preview pane resize (`<`/`>`/`|`) — stored in `state.json`
- Leader key (`g+p`, `g+w`, `g+i`, `g+r`, `g+g`) via `bubbles/key` sequence detection
- `G` / `g+g` for last/first item navigation
- PR section wired up with hardcoded "My PRs" + "Review Requested" sub-tabs (YAML config for sub-tabs in Phase 2)
- `bubbles/spinner` for loading states per pane
- Clipboard: `exec.Command("pbcopy"/"wl-copy"/"xclip"/"clip")` detected at startup

### Phase 2 — PR List Polish

- YAML-configurable sub-tabs for all sections
- Configurable columns (`sections.pullRequests[].columns`)
- Additional fields: `repo`, `draft`, `voteStatus` (from reviewer list in PR response — **treat as API assumption, validate during Phase 2 testing**)
- Row styling via Lip Gloss: `draft` → dimmed, `needs-vote` → cyan
- Client-side filters: `draft: true`, `filter: noVote` — paginate-until-limit using `$top=100` + `$skip` (ADO Git PR API does not use continuation tokens); cap at 10 pages
- Multi-repo fan-out: resolve repo names → IDs at startup; fetch the first page (100 items) from every repo in parallel; merge all results sorted by `updatedAt` desc; take top N; then continue pagination only for repos whose oldest returned item is newer than the Nth merged item (those repos might still have items that belong in the top N); stop when top N is stable or all repos are exhausted. This avoids starving later repos if the first repo has many recent PRs.
- `filter: needsAttention` deferred to Phase 3 (policy API required)

### Phase 3 — PR Detail Pane

- 5-tab detail view (Overview, Reviewers, Checks, Files, Threads) using `bubbles/viewport` for scrollable content
- `glamour` for PR description markdown rendering in Overview tab
- API additions: expand reviewers in PR fetch; policy evaluations:
  ```
  GET /{project}/_apis/policy/evaluations
    ?artifactId=vstfs:///CodeReview/CodeReviewId/{projectId}/{pullRequestId}
    &api-version=7.1-preview.1
  ```
- Row color coding: `failing-checks` → Lip Gloss red ID column (now policy data is available)
- `filter: needsAttention` — client-side post-Phase-3
- Actions: vote (approve/reject/wait-for-author/suggestions), copy URL

### Phase 4 — Pipeline Section

**Source of truth:** Use the **Build API** (`/_apis/build/builds`) as the primary source — it covers both YAML pipelines and classic builds, has richer filter params, and is more stable. The Pipelines Runs API (`/_apis/pipelines/{id}/runs`) is used only for YAML-specific stage/job drill-down in the detail view.

**Status model:** Build API splits running state and outcome into two fields:
- `statusFilter`: `inProgress` | `completed` | `notStarted` | `cancelling`
- `resultFilter`: `succeeded` | `failed` | `canceled` | `partiallySucceeded` (only meaningful when `statusFilter=completed`)

**Pagination:** Build API uses `continuationToken` (response header `x-ms-continuationtoken`), not `$skip`.

**Pipeline sub-tab config → server params:**
```
In Progress  → statusFilter=inProgress
Failed (24h) → statusFilter=completed&resultFilter=failed  + client-side date filter
My Branches  → requestedFor={meId}  (server param)
Main Branch  → branchName=refs/heads/main&statusFilter=completed
Recent       → statusFilter=completed  (no result filter)
```

- Add `listBuilds(params)` to ADO client with `continuationToken` pagination
- Pipeline detail view (Summary, Stages via timeline, Logs, Artifacts)
- Actions: queue build (`POST /_apis/build/builds`), cancel (`PATCH` with `status: "cancelling"`)

### Phase 5 — Work Item Polish

- Work item detail view (Overview, Relations, History)
- Actions: update state, add comment, link PR

### Phase 6 — Releases & Approvals Section

**Important:** Classic Release APIs use a different base URL — `https://vsrm.dev.azure.com/{org}` — not `https://dev.azure.com/{org}`. The ADO client must support a second base URL for release requests. Pagination uses `continuationToken` (response header `x-ms-continuationtoken`), not `$skip`.

- Add `releaseBaseUrl` to the client (`https://vsrm.dev.azure.com/{org}`)
- `listReleases`: `GET {releaseBaseUrl}/{project}/_apis/release/releases?api-version=7.1`
- `listReleaseDefinitions`: same base, `/_apis/release/definitions`
- `listPendingApprovals`: `GET {releaseBaseUrl}/{project}/_apis/release/approvals?assignedToFilter={meId}&statusFilter=pending&api-version=7.1`
- Paginate with `continuationToken` header on all release list calls
- Approval action: `PATCH {releaseBaseUrl}/{project}/_apis/release/approvals/{approvalId}?api-version=7.1` with `{ status: "approved"|"rejected", comments: "..." }`

### Phase 7 — Settings Page + Account Switcher

- Settings TUI screen (edits `defaults` in `config.yml` and `ui` in `state.json`)
- Account Switcher screen (profile list + add/remove/switch profile UI)
- Team picker screen (re-runnable from Settings)

### Phase 8 — Watch / Notification Mode

Separate feature with its own polling loop, diff logic, and delivery — not bundled into Phase 7.

- Background poller (configurable interval, default 60s, separate from display refresh)
- Diffs previous fetch results against current: new review requests, vote changes on my PRs, policy status changes, pipeline completions on my branches
- Persists last-known state to `state.json` for diff across restarts
- Delivery: `bell` (`\a`), `os` (macOS `osascript` / Linux `notify-send`), `off` (default)
- In-TUI `●` badge on sidebar section when new events arrive; clears on visit

### Phase 9 — Distribution

- GoReleaser: builds for `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`
- Homebrew tap via GoReleaser `brews:` config — `brew install ujjwal-goyal/tap/ado-dash`
- npm shim (thin wrapper that `execa`s the platform binary) for `npx ado-dash` support
- GitHub Actions release workflow triggered on `v*` tags

---

## Themes

### Architecture

All Lip Gloss styles reference a `Theme` struct — no hardcoded hex values anywhere in component code. Swapping themes is one struct assignment at startup.

```go
// internal/ui/theme.go

type Theme struct {
    // Backgrounds
    Base    lipgloss.Color  // main background
    Surface lipgloss.Color  // panel / sidebar background
    Overlay lipgloss.Color  // selected row, active tab, modal

    // Text
    Text    lipgloss.Color  // primary text
    Muted   lipgloss.Color  // secondary / dimmed text
    Subtle  lipgloss.Color  // borders, dividers

    // Semantic colors
    Green   lipgloss.Color  // active, approved, success
    Yellow  lipgloss.Color  // warning, waiting, draft
    Red     lipgloss.Color  // rejected, failed, error
    Blue    lipgloss.Color  // links, info
    Purple  lipgloss.Color  // highlights, special

    // UI chrome
    Accent  lipgloss.Color  // focused border, cursor
    Badge   lipgloss.Color  // notification dot
}

var ActiveTheme Theme
```

### Bundled themes

| Config value | Name | Base style |
|---|---|---|
| `catppuccin-mocha` | Catppuccin Mocha | Dark, warm pastel — **default** |
| `catppuccin-macchiato` | Catppuccin Macchiato | Dark, slightly lighter than Mocha |
| `catppuccin-frappe` | Catppuccin Frappé | Dark, cooler/greyer |
| `catppuccin-latte` | Catppuccin Latte | Light, warm |
| `dracula` | Dracula | Dark purple classic |
| `tokyo-night` | Tokyo Night | Dark blue, Neovim-popular |
| `tokyo-night-storm` | Tokyo Night Storm | Slightly lighter than Tokyo Night |
| `nord` | Nord | Dark, arctic cool blue |
| `gruvbox-dark` | Gruvbox Dark | Dark, warm retro |
| `gruvbox-light` | Gruvbox Light | Light, warm retro |
| `one-dark` | One Dark | Dark, Atom/VS Code style |
| `solarized-dark` | Solarized Dark | Dark, classic terminal |
| `solarized-light` | Solarized Light | Light, classic terminal |

### Color palettes

```go
// Catppuccin Mocha
CatppuccinMocha = Theme{
    Base:    "#1e1e2e",  Surface: "#181825",  Overlay: "#313244",
    Text:    "#cdd6f4",  Muted:   "#6c7086",  Subtle:  "#45475a",
    Green:   "#a6e3a1",  Yellow:  "#f9e2af",  Red:     "#f38ba8",
    Blue:    "#89b4fa",  Purple:  "#cba6f7",
    Accent:  "#89dceb",  Badge:   "#f38ba8",
}

// Catppuccin Latte (light)
CatppuccinLatte = Theme{
    Base:    "#eff1f5",  Surface: "#e6e9ef",  Overlay: "#ccd0da",
    Text:    "#4c4f69",  Muted:   "#8c8fa1",  Subtle:  "#bcc0cc",
    Green:   "#40a02b",  Yellow:  "#df8e1d",  Red:     "#d20f39",
    Blue:    "#1e66f5",  Purple:  "#8839ef",
    Accent:  "#04a5e5",  Badge:   "#d20f39",
}

// Tokyo Night
TokyoNight = Theme{
    Base:    "#1a1b26",  Surface: "#16161e",  Overlay: "#2f3549",
    Text:    "#c0caf5",  Muted:   "#565f89",  Subtle:  "#3b4261",
    Green:   "#9ece6a",  Yellow:  "#e0af68",  Red:     "#f7768e",
    Blue:    "#7aa2f7",  Purple:  "#bb9af7",
    Accent:  "#7dcfff",  Badge:   "#f7768e",
}

// Dracula
Dracula = Theme{
    Base:    "#282a36",  Surface: "#21222c",  Overlay: "#44475a",
    Text:    "#f8f8f2",  Muted:   "#6272a4",  Subtle:  "#44475a",
    Green:   "#50fa7b",  Yellow:  "#f1fa8c",  Red:     "#ff5555",
    Blue:    "#8be9fd",  Purple:  "#bd93f9",
    Accent:  "#ff79c6",  Badge:   "#ff5555",
}

// Nord
Nord = Theme{
    Base:    "#2e3440",  Surface: "#272c36",  Overlay: "#3b4252",
    Text:    "#eceff4",  Muted:   "#7b88a1",  Subtle:  "#434c5e",
    Green:   "#a3be8c",  Yellow:  "#ebcb8b",  Red:     "#bf616a",
    Blue:    "#81a1c1",  Purple:  "#b48ead",
    Accent:  "#88c0d0",  Badge:   "#bf616a",
}

// Gruvbox Dark
GruvboxDark = Theme{
    Base:    "#282828",  Surface: "#1d2021",  Overlay: "#3c3836",
    Text:    "#ebdbb2",  Muted:   "#928374",  Subtle:  "#504945",
    Green:   "#b8bb26",  Yellow:  "#fabd2f",  Red:     "#fb4934",
    Blue:    "#83a598",  Purple:  "#d3869b",
    Accent:  "#8ec07c",  Badge:   "#fb4934",
}
```

(One Dark, Solarized Dark/Light, remaining Catppuccin flavors, Gruvbox Light, Tokyo Night Storm follow the same pattern — defined in `internal/ui/themes/` as one file per theme family.)

### Phase placement

- **Phase 1**: establish `Theme` struct, wire into all Lip Gloss styles via `ActiveTheme`, ship Catppuccin Mocha + Latte + Dracula
- **Phase 7** (Settings): add theme picker to the Settings screen, allow live switching without restart
- Remaining themes added incrementally — each is just a `Theme{}` literal, no code changes

### Custom themes (future)

Users can define their own theme in `config.yml`:

```yaml
defaults:
  theme: my-theme

themes:
  my-theme:
    base:    "#0d1117"
    surface: "#161b22"
    overlay: "#21262d"
    text:    "#e6edf3"
    muted:   "#7d8590"
    subtle:  "#30363d"
    green:   "#3fb950"
    yellow:  "#d29922"
    red:     "#f85149"
    blue:    "#58a6ff"
    purple:  "#a371f7"
    accent:  "#79c0ff"
    badge:   "#f85149"
```

---

## Clipboard Implementation

`c` copies the URL of the selected item. Implemented in Go via `os/exec`. Detection order at startup (cached in `app.clipboardCmd`):

```go
// internal/ui/clipboard.go
func detectClipboard() (cmd string, args []string) {
    switch runtime.GOOS {
    case "darwin":
        return "pbcopy", nil
    case "windows":
        return "clip", nil
    default: // Linux/BSD
        if os.Getenv("WAYLAND_DISPLAY") != "" {
            if _, err := exec.LookPath("wl-copy"); err == nil {
                return "wl-copy", nil
            }
        }
        if _, err := exec.LookPath("xclip"); err == nil {
            return "xclip", []string{"-selection", "clipboard"}
        }
        if _, err := exec.LookPath("xsel"); err == nil {
            return "xsel", []string{"--clipboard", "--input"}
        }
    }
    return "", nil // no clipboard tool found
}
```

If no tool found: display the URL in the status bar as `  URL: <url>` so the user can read it. Never silently fail.

---

## Resize & Layout Handling

Bubble Tea owns the render loop, alternate screen, and cursor. Manual ANSI cursor math is not needed. Resize handling works as follows:

- Bubble Tea delivers a `tea.WindowSizeMsg{Width, Height}` whenever the terminal is resized. The root `App` model stores the current dimensions and propagates them to child models.
- Auto-collapse is a layout decision inside the `View()` method — if `msg.Width < threshold`, render the collapsed sidebar variant. No debounce needed; Bubble Tea coalesces rapid resize events naturally.
- Lip Gloss handles line truncation, padding, and width constraints. `lipgloss.Width()` returns the correct display width (ANSI-stripped, Unicode-aware) — use it for all column width calculations.

### Auto-collapse thresholds

| Terminal width | Layout effect |
|---|---|
| ≥ 140 cols | Sidebar + preview pane both visible at saved widths |
| 100–139 cols | Preview pane auto-hidden (list fills full content width) |
| < 100 cols | Preview pane hidden + sidebar auto-collapsed to 3 chars |

Auto-collapse is an *override*, not a state change. The `previewPaneWidth` and `sidebar` values in `state.json` are the user's *preferred* state; they restore when the terminal widens back above the threshold.

---

## Default Sections & Queries

### Pull Requests defaults

```yaml
sections:
  pullRequests:
    - title: My PRs
      role: author
      status: active

    - title: Review Requested
      role: reviewer
      status: active

    - title: Draft
      role: author
      status: active
      draft: true              # client-side filter on isDraft field

    - title: Waiting on Me
      role: reviewer
      status: active
      # client-side filter: I am a reviewer with vote = 0 (no vote yet)
      filter: noVote

    - title: Completed
      role: author
      status: completed
      limit: 15

    - title: Needs Attention
      role: author
      status: active
      # client-side filter: my PRs where any reviewer voted reject/wait-for-author,
      # OR a required policy is failing — surfaces PRs that need my action
      filter: needsAttention

    - title: All Active
      status: active
      # no role filter — full project view, useful for team leads
```

`draft: true` is a client-side filter (ADO doesn't support it as a server param in 7.1 — fetch active PRs, filter `pr.isDraft === true`).

`filter: noVote` is also client-side: fetch review-requested PRs, filter where my reviewer entry has `vote === 0`.

### Work Items defaults

```yaml
sections:
  workItems:
    - title: Assigned to Me
      wiql: >-
        SELECT [System.Id], [System.Title], [System.State],
               [System.WorkItemType], [System.AssignedTo], [System.ChangedDate]
        FROM WorkItems
        WHERE [System.TeamProject] = @project
          AND [System.AssignedTo] = @Me
          AND [System.State] NOT IN ('Closed', 'Resolved', 'Done')
        ORDER BY [System.ChangedDate] DESC

    - title: Current Sprint
      # {currentIterationPath} is resolved at runtime via teams API (see Team Auto-Detection)
      iterationResolved: true
      wiql: >-
        SELECT [System.Id], [System.Title], [System.State],
               [System.WorkItemType], [System.AssignedTo]
        FROM WorkItems
        WHERE [System.TeamProject] = @project
          AND [System.IterationPath] = '{currentIterationPath}'
          AND [System.AssignedTo] = @Me
          AND [System.State] NOT IN ('Closed', 'Resolved', 'Done')
        ORDER BY [Microsoft.VSTS.Common.Priority] ASC

    - title: My Team's Sprint
      iterationResolved: true
      wiql: >-
        SELECT [System.Id], [System.Title], [System.State],
               [System.WorkItemType], [System.AssignedTo]
        FROM WorkItems
        WHERE [System.TeamProject] = @project
          AND [System.IterationPath] = '{currentIterationPath}'
          AND [System.State] NOT IN ('Closed', 'Resolved', 'Done')
        ORDER BY [System.AssignedTo] ASC,
                 [Microsoft.VSTS.Common.Priority] ASC

    - title: High Priority Bugs
      wiql: >-
        SELECT [System.Id], [System.Title], [System.State],
               [System.AssignedTo], [Microsoft.VSTS.Common.Priority]
        FROM WorkItems
        WHERE [System.TeamProject] = @project
          AND [System.WorkItemType] = 'Bug'
          AND [Microsoft.VSTS.Common.Priority] <= 2
          AND [System.State] NOT IN ('Closed', 'Resolved', 'Done')
        ORDER BY [Microsoft.VSTS.Common.Priority] ASC,
                 [System.ChangedDate] DESC

    - title: Blocked
      # Tag filter is portable across all process templates.
      # State = 'Blocked' only exists in CMMI; tolerate failure silently on other processes.
      # ado-dash runs this query and falls back to tag-only if the state clause errors.
      wiql: >-
        SELECT [System.Id], [System.Title], [System.State],
               [System.AssignedTo], [System.Tags]
        FROM WorkItems
        WHERE [System.TeamProject] = @project
          AND [System.AssignedTo] = @Me
          AND [System.State] NOT IN ('Closed', 'Resolved', 'Done')
          AND (
            [System.Tags] CONTAINS 'blocked'
            OR [System.State] = 'Blocked'
          )
        ORDER BY [System.ChangedDate] DESC

    - title: Stale (mine, 14+ days)
      wiql: >-
        SELECT [System.Id], [System.Title], [System.State],
               [System.ChangedDate], [System.WorkItemType]
        FROM WorkItems
        WHERE [System.TeamProject] = @project
          AND [System.AssignedTo] = @Me
          AND [System.State] NOT IN ('Closed', 'Resolved', 'Done')
          AND [System.ChangedDate] <= @Today - 14
        ORDER BY [System.ChangedDate] ASC

    - title: Created by Me
      wiql: >-
        SELECT [System.Id], [System.Title], [System.State],
               [System.CreatedDate], [System.WorkItemType]
        FROM WorkItems
        WHERE [System.TeamProject] = @project
          AND [System.CreatedBy] = @Me
          AND [System.CreatedDate] >= @Today - 30
        ORDER BY [System.CreatedDate] DESC

    - title: Unassigned (active)
      iterationResolved: true
      wiql: >-
        SELECT [System.Id], [System.Title], [System.State],
               [System.WorkItemType], [System.CreatedDate]
        FROM WorkItems
        WHERE [System.TeamProject] = @project
          AND [System.AssignedTo] = ''
          AND [System.State] NOT IN ('Closed', 'Done', 'Resolved')
          AND [System.IterationPath] = '{currentIterationPath}'
        ORDER BY [Microsoft.VSTS.Common.Priority] ASC
```

### Pipelines defaults

```yaml
sections:
  pipelines:
    - title: In Progress
      status: inProgress     # running right now

    - title: Failed (24h)
      status: failed
      # client-side filter: completedDate >= now - 24h

    - title: My Branches
      # client-side filter: builds on branches pushed by @Me
      # server param: requestedFor = me.id

    - title: Main Branch
      branch: main
      status: all
      limit: 20

    - title: Recent
      status: all
      limit: 30
```

### Releases defaults

```yaml
sections:
  releases:
    - title: Pending Approval
      status: pending         # gates waiting for my approval

    - title: In Progress
      status: inProgress

    - title: Recent
      status: all
      limit: 20
```

---

## Watch / Notification Mode (Phase 7)

Poll interval: configurable, default 60s (separate from the display refresh interval).

**What to watch:**
- New PRs assigned to me for review
- PR vote changes on my PRs (someone approved/rejected)
- PR policy status changes (build passed/failed on my PR)
- Pipeline run completed (success or failure) on my branches

**Delivery channels:**
- `bell` — writes `\a` (BEL character) to the terminal, audible bell
- `os` — macOS: `osascript -e 'display notification ...'`; Linux: `notify-send`
- `off` (default)

**In-TUI indicator:** when a new notification arrives, show a `●` badge next to the relevant sidebar section. Clear on visit.

---

## Client-Side Filters (draft, noVote)

For sections that need client-side filtering (e.g. `draft: true`, `filter: noVote`), the fetch strategy is:

- **Paginate until `limit` filtered results are collected**, using `$top` + `$skip`. Fetch pages of a fixed page size (e.g. 100) and accumulate filtered results until we have `limit` items or the server returns fewer rows than the page size (meaning no more pages).
- This is correct regardless of how many PRs exist in the project and avoids pulling the entire PR list on every refresh.
- A hard cap of `maxFetchPages: 10` (configurable) prevents runaway fetching on very large projects.

```
page loop:
  fetch($top=100, $skip=offset)
  filter results → append to collected
  if collected.length >= limit → done
  if response.length < 100    → done (last page)
  offset += 100
  if pages >= maxFetchPages   → done (cap hit, show what we have)
```

---

## Team Auto-Detection (for `@CurrentIteration`)

WIQL's `@CurrentIteration` macro requires a team context: `@CurrentIteration('[ProjectName]\TeamName')`. Rather than making the user configure this manually, ado-dash auto-detects it:

**First time a work-item section using `@CurrentIteration` is loaded:**

1. Call `GET /_apis/projects/{projectId}/teams?$mine=true&api-version=7.1` — returns teams the authenticated user belongs to within the selected project. (Note: the org-level `/_apis/teams` endpoint does not support `$mine=true` in 7.1; use the project-scoped route.)
2. If exactly one team → use it automatically, save to config as `defaultTeamId` + `defaultTeamName`.
3. If multiple teams → show an inline team picker:

```
╭──────────────────────────────────────────────────────╮
│  Select your primary team                            │
│  (used for @CurrentIteration queries)                │
╰──────────────────────────────────────────────────────╯

  ▶  Platform Team
     Mobile Team
     Backend Services
     DevOps Guild

  j/k move · enter select · (saved to state.json)
```

4. Selection is saved to `state.json` under `profiles[name].defaultTeamId` + `defaultTeamName`.
5. Subsequent loads use the saved value — no picker shown again.
6. User can change their team from the Settings page at any time.

**Resolving the iteration path at query time:**

Do not use the bare `@CurrentIteration` macro or the team-qualified form `@CurrentIteration('[Project]\Team')` — both have inconsistent behavior across REST clients per Microsoft docs. Instead, resolve the current iteration explicitly before running the WIQL:

```
GET https://dev.azure.com/{org}/{project}/{team}/_apis/work/teamsettings/iterations?$timeframe=current&api-version=7.1
→ response.value[0].path  e.g. "MyProject\\Sprint 12"
```

Substitute the resolved path into the WIQL before sending:

```sql
WHERE [System.IterationPath] = 'MyProject\Sprint 12'
```

Cache the resolved path per `(teamId, date)` — it only changes at sprint boundaries.

**State fields (profile-scoped — all in `state.json`, never in `config.yml`):**

```json
{
  "activeProfile": "Work",
  "profiles": {
    "Work": {
      "defaultProject": "MyProject",
      "defaultTeamId": "abc-123-def",
      "defaultTeamName": "Platform Team"
    }
  }
}
```

Use `defaultTeamId` (not team name) in the iterations URL to survive team renames:
```
GET https://dev.azure.com/{org}/{project}/_apis/work/teamsettings/iterations
  ?$timeframe=current&api-version=7.1
  (with teamId=defaultTeamId in the URL path, URL-encoded)
```

**Settings page entry:**

```
Default Team   [ Platform Team ▼ ]   (re-runs team picker)
```
