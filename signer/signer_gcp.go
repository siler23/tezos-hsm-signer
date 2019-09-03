package signer

import (
	"context"
	"crypto/elliptic"
	"fmt"

	cloudkms "cloud.google.com/go/kms/apiv1"
	"github.com/btcsuite/btcd/btcec"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
)

type googleCloudKMSSigner struct {
	kmsClient *cloudkms.KeyManagementClient
}

// NewGoogleCloudKMSSigner creates a signer backed by Google Cloud KMS
func NewGoogleCloudKMSSigner(kmsClient *cloudkms.KeyManagementClient) Signer {
	return &googleCloudKMSSigner{
		kmsClient: kmsClient,
	}
}

func (g *googleCloudKMSSigner) Sign(ctx context.Context, message []byte, key *Key) ([]byte, error) {
	req := &kmspb.AsymmetricSignRequest{
		Name: key.Name,
		Digest: &kmspb.Digest{
			// It's actually Blake2b.Sum256, not SHA256, but google doesn't know the difference
			Digest: &kmspb.Digest_Sha256{
				Sha256: message,
			},
		},
	}
	response, err := g.kmsClient.AsymmetricSign(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("asymmetric sign request failed: %+v", err)
	}
	signature, err := btcec.ParseDERSignature(response.Signature, elliptic.P256())
	if err != nil {
		return nil, fmt.Errorf("failed to parse ASN.1 encoded ECDSA signature")
	}
	sigBytes := append(signature.R.Bytes(), signature.S.Bytes()...)
	if len(sigBytes) != 64 {
		return nil, fmt.Errorf("unexpected signature length: %d bytes, expected %d bytes", len(sigBytes), 64)
	}
	return sigBytes, nil
}
