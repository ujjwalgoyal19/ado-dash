# Classic Release API uses a separate base URL (vsrm.dev.azure.com)

All Classic Release API calls go to `https://vsrm.dev.azure.com/{org}`, not `https://dev.azure.com/{org}`. The ADO client maintains a second base URL specifically for release requests.

This is an undocumented-but-stable ADO infrastructure split. Release data is served from a different subdomain than the rest of the ADO APIs. There is no supported way to proxy release calls through the main `dev.azure.com` endpoint — they simply 404.

**Consequences**
- The ADO client (`internal/client/client.go`) exposes both `baseURL` and `releaseBaseURL` fields.
- Auth headers are the same for both endpoints — the same token works across subdomains.
- When adding release support in Phase 6, all release and approval endpoints must use `releaseBaseURL`.
