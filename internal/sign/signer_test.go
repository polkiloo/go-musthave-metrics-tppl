package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestSignerSHA256_Sign(t *testing.T) {
	s := NewSignerSHA256()
	data := []byte("data")
	key := SignKey("key")
	got := s.Sign(data, key)
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	want := hex.EncodeToString(h.Sum(nil))
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if sig := s.Sign(data, ""); sig != "" {
		t.Fatalf("expected empty signature, got %q", sig)
	}
}

func TestSignerSHA256_Verify(t *testing.T) {
	s := NewSignerSHA256()
	data := []byte("hello")
	key := SignKey("secret")
	sig := s.Sign(data, key)
	if !s.Verify(data, key, sig) {
		t.Fatalf("expected verification success")
	}
	if s.Verify(data, key, "bad") {
		t.Fatalf("expected verification failure")
	}
	if !s.Verify(data, "", "whatever") {
		t.Fatalf("expected success with empty key")
	}
}

func TestProvideSigner(t *testing.T) {
	s, err := ProvideSigner("secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.(*SignerSHA256); !ok {
		t.Fatalf("unexpected signer type %T", s)
	}
}
