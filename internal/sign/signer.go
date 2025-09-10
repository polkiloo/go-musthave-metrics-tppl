package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"go.uber.org/fx"
)

type SignKey string

type Signer interface {
	Sign(data []byte, key SignKey) string
	Verify(data []byte, key SignKey, sig string) bool
}

type SignerSHA256 struct{}

func (s *SignerSHA256) Sign(data []byte, key SignKey) string {
	if key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func (s *SignerSHA256) Verify(data []byte, key SignKey, sig string) bool {
	if key == "" {
		return true
	}
	expected := s.Sign(data, key)
	return hmac.Equal([]byte(expected), []byte(sig))
}

func NewSignerSHA256() *SignerSHA256 {
	return &SignerSHA256{}
}

var _ Signer = NewSignerSHA256()

func ProvideSigner(key SignKey) (Signer, error) {
	return NewSignerSHA256(), nil
}

var Module = fx.Module(
	"signer",
	fx.Provide(ProvideSigner),
)
