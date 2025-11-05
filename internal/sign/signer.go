package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"go.uber.org/fx"
)

// SignKey represents a shared secret used for request signing.
type SignKey string

// Signer defines behaviour for signing and verifying payload digests.
type Signer interface {
	Sign(data []byte, key SignKey) string
	Verify(data []byte, key SignKey, sig string) bool
}

// SignerSHA256 signs payloads using HMAC SHA-256.
type SignerSHA256 struct{}

// Sign calculates a hex-encoded HMAC SHA-256 signature for data using the provided key.
func (s *SignerSHA256) Sign(data []byte, key SignKey) string {
	if key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// Verify checks whether the provided signature matches the HMAC SHA-256 digest of the data.
func (s *SignerSHA256) Verify(data []byte, key SignKey, sig string) bool {
	if key == "" {
		return true
	}
	expected := s.Sign(data, key)
	return hmac.Equal([]byte(expected), []byte(sig))
}

// NewSignerSHA256 constructs a SignerSHA256 instance.
func NewSignerSHA256() *SignerSHA256 {
	return &SignerSHA256{}
}

// ProvideSigner is an fx constructor that returns the default signer implementation.
func ProvideSigner(key SignKey) (Signer, error) {
	return NewSignerSHA256(), nil
}

// Module describes the fx module for providing request signing facilities.
var Module = fx.Module(
	"signer",
	fx.Provide(ProvideSigner),
)

var _ Signer = NewSignerSHA256()
