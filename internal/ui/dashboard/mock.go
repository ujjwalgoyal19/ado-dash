package dashboard

// Mock PR data matching the design spec exactly.
// Replaced with live ADO API data in a later slice.

type prFlag string

const (
	flagNone   prFlag = ""
	flagDraft  prFlag = "draft"
	flagVote   prFlag = "vote"
	flagFailed prFlag = "failed"
)

type buildStatus string

const (
	buildPass buildStatus = "pass"
	buildFail buildStatus = "fail"
	buildRun  buildStatus = "run"
)

type PR struct {
	ID         string
	State      string // "active" | "draft" | "merged" | "closed"
	Title      string
	Author     string
	Repo       string
	Age        string
	Flag       prFlag
	Branch     string
	Approved   int
	Waiting    int
	Reviewers  int
	Build      buildStatus
	WorkItem   bool
}

var mockPRs = []PR{
	{"PR1042", "active", "Fix OAuth token refresh race", "alice", "api-gateway", "2h", flagNone, "alice/fix-auth-bug", 1, 2, 3, buildPass, false},
	{"PR1031", "draft", "Add retry / backoff to client", "carol", "api-gateway", "1d", flagDraft, "carol/retry-backoff", 0, 1, 2, buildRun, true},
	{"PR1014", "active", "Warm cache on cold start", "frank", "api-gateway", "2d", flagNone, "frank/cache-warm", 2, 0, 2, buildPass, true},
	{"PR0984", "active", "Drop legacy auth shim", "mike", "api-gateway", "8d", flagNone, "mike/drop-auth-shim", 2, 1, 3, buildPass, true},
	{"PR1038", "active", "Refactor list virtualization", "bob", "web-portal", "5h", flagNone, "bob/list-virt", 2, 0, 2, buildPass, true},
	{"PR1009", "draft", "Spike: edge rendering", "grace", "web-portal", "3d", flagDraft, "grace/edge-spike", 0, 0, 1, buildRun, false},
	{"PR1027", "active", "Rewrite onboarding docs", "dave", "docs-site", "1d", flagVote, "dave/onboarding-docs", 0, 2, 2, buildPass, true},
	{"PR1019", "active", "Bump grpc to 1.62", "erin", "platform-core", "2d", flagFailed, "erin/grpc-162", 1, 1, 3, buildFail, true},
	{"PR1003", "active", "Fix flaky reviewer test", "heidi", "platform-core", "4d", flagNone, "heidi/flaky-test", 2, 0, 2, buildPass, true},
	{"PR0998", "active", "Add audit log fields", "ivan", "data-svc", "5d", flagNone, "ivan/audit-log", 1, 1, 2, buildPass, true},
	{"PR0991", "active", "Tidy CI matrix", "judy", "infra", "6d", flagNone, "judy/ci-matrix", 1, 0, 1, buildPass, true},
}

var Repos = []string{"api-gateway", "web-portal", "docs-site", "platform-core", "data-svc", "infra"}

func PRsForRepo(repo string) []PR {
	var out []PR
	for _, p := range mockPRs {
		if p.Repo == repo {
			out = append(out, p)
		}
	}
	return out
}
