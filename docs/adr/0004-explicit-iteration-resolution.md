# Resolve current iteration path explicitly via API instead of using @CurrentIteration

WIQL sections that need the current sprint substitute a literal Iteration Path string (e.g. `"MyProject\\Sprint 12"`) resolved at query time, rather than using the `@CurrentIteration` or `@CurrentIteration('[Project]\Team')` WIQL macros.

Microsoft's own documentation notes inconsistent behaviour of `@CurrentIteration` across REST clients — the team-qualified form works in the browser but behaves unreliably when submitted via the REST API. The explicit path is unambiguous and works across all ADO versions and process templates.

**How it works**
At startup (or on first use of an iteration-bound Sub-tab), ado-dash calls:
```
GET /{org}/{project}/{team}/_apis/work/teamsettings/iterations?$timeframe=current&api-version=7.1
```
and caches `response.value[0].path` per `(teamId, date)`. This value is substituted into WIQL before the query is sent. The cache is invalidated at sprint boundaries (when the date crosses into a new iteration).

**Consequences**
- Requires a Team to be resolved first. ado-dash auto-detects the team on first use and saves it to the State File.
- An extra API call is needed before WIQL execution for iteration-bound tabs. The cache makes subsequent loads free within a sprint.
