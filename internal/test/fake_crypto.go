package test

import "sync"

// FakeEncryptor implements cryptoutil.Encryptor for tests.
type FakeEncryptor struct {
	Ciphertext   []byte
	EncryptedKey string
	Err          error

	mu       sync.Mutex
	received [][]byte
}

// Encrypt records the plaintext and returns configured outputs or echoes the plaintext.
func (f *FakeEncryptor) Encrypt(plain []byte) ([]byte, string, error) {
	f.mu.Lock()
	f.received = append(f.received, append([]byte(nil), plain...))
	f.mu.Unlock()

	if f.Err != nil {
		return nil, "", f.Err
	}
	if f.Ciphertext != nil {
		return append([]byte(nil), f.Ciphertext...), f.EncryptedKey, nil
	}
	return append([]byte(nil), plain...), f.EncryptedKey, nil
}

// Calls returns how many times Encrypt was invoked.
func (f *FakeEncryptor) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.received)
}

// LastPlain returns the last plaintext passed to Encrypt.
func (f *FakeEncryptor) LastPlain() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.received) == 0 {
		return nil
	}
	return append([]byte(nil), f.received[len(f.received)-1]...)
}

// Received returns copies of all plaintexts passed to Encrypt.
func (f *FakeEncryptor) Received() [][]byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([][]byte, len(f.received))
	for i, v := range f.received {
		out[i] = append([]byte(nil), v...)
	}
	return out
}

// FakeDecryptor implements cryptoutil.Decryptor for tests.
type FakeDecryptor struct {
	Plaintext []byte
	Err       error

	mu       sync.Mutex
	received [][]byte
	keys     []string
}

// Decrypt records the ciphertext and returns configured output or echoes the ciphertext.
func (f *FakeDecryptor) Decrypt(ciphertext []byte, encryptedKey string) ([]byte, error) {
	f.mu.Lock()
	f.received = append(f.received, append([]byte(nil), ciphertext...))
	f.keys = append(f.keys, encryptedKey)
	f.mu.Unlock()

	if f.Err != nil {
		return nil, f.Err
	}
	if f.Plaintext != nil {
		return append([]byte(nil), f.Plaintext...), nil
	}
	return append([]byte(nil), ciphertext...), nil
}

// Calls returns how many times Decrypt was invoked.
func (f *FakeDecryptor) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.received)
}

// LastCipher returns the last ciphertext passed to Decrypt.
func (f *FakeDecryptor) LastCipher() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.received) == 0 {
		return nil
	}
	return append([]byte(nil), f.received[len(f.received)-1]...)
}

// LastKey returns the last encrypted key passed to Decrypt.
func (f *FakeDecryptor) LastKey() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.keys) == 0 {
		return ""
	}
	return f.keys[len(f.keys)-1]
}

// Keys returns a copy of all encrypted keys passed to Decrypt.
func (f *FakeDecryptor) Keys() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.keys))
	copy(out, f.keys)
	return out
}

// Ciphers returns copies of all ciphertexts passed to Decrypt.
func (f *FakeDecryptor) Ciphers() [][]byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([][]byte, len(f.received))
	for i, v := range f.received {
		out[i] = append([]byte(nil), v...)
	}
	return out
}
