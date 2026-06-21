# Store env var name in config, not the PAT value itself

For PAT auth, `config.yml` stores the *name* of the environment variable holding the token (e.g. `"AZURE_DEVOPS_PAT"`), never the token value. The wizard rejects direct token input. ado-dash reads the referenced env var at runtime.

Config files are frequently committed to version control by accident, included in backup archives, or inspected by third-party tools. Storing a live PAT there would expose a credential with the blast radius of the entire Azure DevOps organization. The env var indirection means a leaked config file leaks only a variable name — which is harmless without the environment it runs in.

**Consequences**
- The config file must be `0600`; ado-dash warns on startup if it is not.
- Users who set `env: AZURE_DEVOPS_PAT` in config must export that variable in their shell before running ado-dash. The auth wizard explains this.
- Service Principal client secrets follow the same pattern (env var reference, not inline value).
