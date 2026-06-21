package auth

import "context"

// Provider is the interface every auth implementation satisfies.
// Token returns a Bearer token for the given context.
// Validate proves the token can reach the configured ADO org by
// calling /_apis/connectionData — not just minting a token.
type Provider interface {
	Token(ctx context.Context) (string, error)
	Validate(ctx context.Context) error
}
