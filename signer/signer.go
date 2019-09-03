package signer

import (
	"context"
)

// Signer is a generic interface for a signer
type Signer interface {
	Sign(ctx context.Context, message []byte, key *Key) ([]byte, error)
}
