package signer

import (
	"context"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/ed25519"
)

type inMemorySigner struct {
	privateKey    ed25519.PrivateKey
	publicKeyHash string
}

// NewInMemorySigner creates a signer from a key stored plaintext in memory.
// It is not suitable for production use.
func NewInMemorySigner(privateKey ed25519.PrivateKey) Signer {
	publicKeyHash, err := blake2b.New(20, nil)
	if err != nil {
		panic(err.Error())
	}
	_, err = publicKeyHash.Write(privateKey.Public().(ed25519.PublicKey))
	if err != nil {
		panic(err.Error())
	}
	publicKeyHashBytes := publicKeyHash.Sum([]byte{})
	prefix, _ := hex.DecodeString(tzEd25519PublicKeyHash)
	publicKeyHashString := b58CheckEncode(prefix, publicKeyHashBytes)
	return &inMemorySigner{
		privateKey:    privateKey,
		publicKeyHash: publicKeyHashString,
	}
}

func (i *inMemorySigner) Sign(_ context.Context, message []byte, key *Key) ([]byte, error) {
	if key.PublicKeyHash != i.publicKeyHash {
		return nil, fmt.Errorf("unknown key %s, expected %s", key.PublicKeyHash, i.publicKeyHash)
	}
	return ed25519.Sign(i.privateKey, message), nil
}
