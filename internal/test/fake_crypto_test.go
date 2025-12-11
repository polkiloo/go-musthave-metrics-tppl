package test

import (
	"errors"
	"testing"
)

func TestFakeEncryptor_Defaults(t *testing.T) {
	f := &FakeEncryptor{}

	cipher, key, err := f.Encrypt([]byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(cipher) != "hello" {
		t.Fatalf("ciphertext mismatch: %q", cipher)
	}
	if key != "" {
		t.Fatalf("expected empty key, got %q", key)
	}
	if f.Calls() != 1 {
		t.Fatalf("expected 1 call, got %d", f.Calls())
	}
	if got := string(f.LastPlain()); got != "hello" {
		t.Fatalf("unexpected recorded plaintext: %q", got)
	}
}

func TestFakeEncryptor_ConfiguredOutput(t *testing.T) {
	f := &FakeEncryptor{Ciphertext: []byte("cipher"), EncryptedKey: "key"}

	cipher, key, err := f.Encrypt([]byte("ignored"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(cipher) != "cipher" || key != "key" {
		t.Fatalf("unexpected output: %q %q", cipher, key)
	}
}

func TestFakeEncryptor_Error(t *testing.T) {
	wantErr := errors.New("boom")
	f := &FakeEncryptor{Err: wantErr}

	if _, _, err := f.Encrypt([]byte("data")); err != wantErr {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}

func TestFakeDecryptor_Defaults(t *testing.T) {
	f := &FakeDecryptor{}

	plain, err := f.Decrypt([]byte("cipher"), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(plain) != "cipher" {
		t.Fatalf("plaintext mismatch: %q", plain)
	}
	if f.Calls() != 1 {
		t.Fatalf("expected 1 call, got %d", f.Calls())
	}
	if got := string(f.LastCipher()); got != "cipher" {
		t.Fatalf("unexpected recorded ciphertext: %q", got)
	}
	if f.LastKey() != "" {
		t.Fatalf("expected empty key, got %q", f.LastKey())
	}
}

func TestFakeDecryptor_ConfiguredOutput(t *testing.T) {
	f := &FakeDecryptor{Plaintext: []byte("plain")}

	plain, err := f.Decrypt([]byte("ignored"), "encKey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(plain) != "plain" {
		t.Fatalf("unexpected output: %q", plain)
	}
	if f.LastKey() != "encKey" {
		t.Fatalf("expected key to be recorded")
	}
}

func TestFakeDecryptor_Error(t *testing.T) {
	wantErr := errors.New("boom")
	f := &FakeDecryptor{Err: wantErr}

	if _, err := f.Decrypt([]byte("data"), "k"); err != wantErr {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}
