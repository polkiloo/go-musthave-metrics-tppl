package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	// CryptoKeyHeader carries the encrypted symmetric key used for payload encryption.
	CryptoKeyHeader = "Crypto-Key"
	nonceSize       = 12
	keySize         = 32
)

var (
	randReader   io.Reader                               = rand.Reader
	newAESCipher func([]byte) (cipher.Block, error)      = aes.NewCipher
	newGCM       func(cipher.Block) (cipher.AEAD, error) = cipher.NewGCM
)

// Encryptor encrypts payloads for transport.
type Encryptor interface {
	Encrypt(plain []byte) (ciphertext []byte, encryptedKey string, err error)
}

// Decryptor decrypts payloads produced by Encryptor.
type Decryptor interface {
	Decrypt(ciphertext []byte, encryptedKey string) ([]byte, error)
}

// HybridEncryptor implements hybrid RSA + AES-GCM encryption.
type HybridEncryptor struct {
	pub *rsa.PublicKey
}

// HybridDecryptor decrypts payloads encrypted with HybridEncryptor.
type HybridDecryptor struct {
	priv *rsa.PrivateKey
}

// NewEncryptorFromPublicKeyFile reads a PEM-encoded RSA public key and constructs an Encryptor.
func NewEncryptorFromPublicKeyFile(path string) (Encryptor, error) {
	if path == "" {
		return nil, nil
	}
	pub, err := loadPublicKey(path)
	if err != nil {
		return nil, err
	}
	return &HybridEncryptor{pub: pub}, nil
}

// NewDecryptorFromPrivateKeyFile reads a PEM-encoded RSA private key and constructs a Decryptor.
func NewDecryptorFromPrivateKeyFile(path string) (Decryptor, error) {
	if path == "" {
		return nil, nil
	}
	priv, err := loadPrivateKey(path)
	if err != nil {
		return nil, err
	}
	return &HybridDecryptor{priv: priv}, nil
}

// Encrypt encrypts plain using AES-256-GCM with a random key, which is then encrypted with RSA.
func (e *HybridEncryptor) Encrypt(plain []byte) ([]byte, string, error) {
	if e == nil || e.pub == nil {
		return nil, "", errors.New("encryptor is not configured")
	}

	symKey := make([]byte, keySize)
	if _, err := io.ReadFull(randReader, symKey); err != nil {
		return nil, "", fmt.Errorf("generate sym key: %w", err)
	}
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(randReader, nonce); err != nil {
		return nil, "", fmt.Errorf("generate nonce: %w", err)
	}

	block, err := newAESCipher(symKey)
	if err != nil {
		return nil, "", fmt.Errorf("cipher: %w", err)
	}
	gcm, err := newGCM(block)
	if err != nil {
		return nil, "", fmt.Errorf("gcm: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plain, nil)

	encKey, err := rsa.EncryptOAEP(sha256.New(), randReader, e.pub, symKey, nil)
	if err != nil {
		return nil, "", fmt.Errorf("encrypt sym key: %w", err)
	}

	return ciphertext, base64.StdEncoding.EncodeToString(encKey), nil
}

// Decrypt decrypts ciphertext produced by HybridEncryptor.
func (d *HybridDecryptor) Decrypt(ciphertext []byte, encryptedKey string) ([]byte, error) {
	if d == nil || d.priv == nil {
		return nil, errors.New("decryptor is not configured")
	}

	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	symKey, err := decodeAndDecryptKey(encryptedKey, d.priv)
	if err != nil {
		return nil, err
	}

	block, err := newAESCipher(symKey)
	if err != nil {
		return nil, fmt.Errorf("cipher: %w", err)
	}
	gcm, err := newGCM(block)
	if err != nil {
		return nil, fmt.Errorf("gcm: %w", err)
	}

	nonce := ciphertext[:nonceSize]
	data := ciphertext[nonceSize:]

	plain, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt payload: %w", err)
	}
	return plain, nil
}

func decodeAndDecryptKey(encoded string, priv *rsa.PrivateKey) ([]byte, error) {
	if encoded == "" {
		return nil, errors.New("missing encrypted key")
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode key: %w", err)
	}
	key, err := rsa.DecryptOAEP(sha256.New(), randReader, priv, raw, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt key: %w", err)
	}
	if len(key) != keySize {
		return nil, errors.New("invalid symmetric key length")
	}
	return key, nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("no PEM data found")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return key, nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("no PEM data found")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	keyAny, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := keyAny.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	return key, nil
}
