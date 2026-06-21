# Use Build API as primary pipeline data source

Pipeline list views use the Build API (`/_apis/build/builds`) as their primary source, not the Pipelines Runs API (`/_apis/pipelines/{id}/runs`).

The Build API covers both YAML pipelines and Classic builds, exposes richer server-side filter parameters (`statusFilter`, `resultFilter`, `requestedFor`, `branchName`), and uses a stable pagination model (`x-ms-continuationtoken` header). The Pipelines Runs API is YAML-only and has a narrower filter surface.

The Pipelines Runs API is used only in the Pipeline Detail view for YAML-specific stage and job drill-down, where its structured stage data is needed.

**Consequences**
- The Build API's status model has two orthogonal fields: `statusFilter` and `resultFilter`. Sub-tab config maps to these; see PLAN.md Phase 4 for the mapping table.
- Pagination for pipelines uses continuation tokens, not `$skip` — unlike the Git PR API. The ADO client must handle both styles.
