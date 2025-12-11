package cryptoutil

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

// stubReader allows controlling successful reads before returning an error.
type stubReader struct {
	successes int
	err       error
}

func (r *stubReader) Read(p []byte) (int, error) {
	if r.successes > 0 {
		for i := range p {
			p[i] = byte(r.successes)
		}
		r.successes--
		return len(p), nil
	}
	return 0, r.err
}

func TestNewEncryptorFromPublicKeyFile(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		enc, err := NewEncryptorFromPublicKeyFile("")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if enc != nil {
			t.Fatalf("expected nil encryptor, got %v", enc)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		if _, err := NewEncryptorFromPublicKeyFile("/tmp/missing.pub"); err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("invalid pem", func(t *testing.T) {
		path := writeTempFile(t, []byte("invalid"))
		if _, err := NewEncryptorFromPublicKeyFile(path); err == nil {
			t.Fatal("expected error for invalid pem")
		}
	})

	t.Run("success", func(t *testing.T) {
		_, _, pubPath := generateKeyPairFiles(t)
		enc, err := NewEncryptorFromPublicKeyFile(pubPath)
		if err != nil {
			t.Fatalf("expected encryptor, got %v", err)
		}
		if enc == nil {
			t.Fatal("encryptor should not be nil")
		}
	})
}

func TestNewDecryptorFromPrivateKeyFile(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		dec, err := NewDecryptorFromPrivateKeyFile("")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if dec != nil {
			t.Fatalf("expected nil decryptor, got %v", dec)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		if _, err := NewDecryptorFromPrivateKeyFile("/tmp/missing.key"); err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("invalid pem", func(t *testing.T) {
		path := writeTempFile(t, []byte("invalid"))
		if _, err := NewDecryptorFromPrivateKeyFile(path); err == nil {
			t.Fatal("expected error for invalid pem")
		}
	})

	t.Run("non rsa pkcs8", func(t *testing.T) {
		ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("generate ecdsa: %v", err)
		}
		raw, err := x509.MarshalPKCS8PrivateKey(ecdsaKey)
		if err != nil {
			t.Fatalf("marshal ecdsa: %v", err)
		}
		pemData := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: raw})
		path := writeTempFile(t, pemData)
		if _, err := NewDecryptorFromPrivateKeyFile(path); err == nil || !strings.Contains(err.Error(), "not an RSA private key") {
			t.Fatalf("expected non RSA error, got %v", err)
		}
	})

	t.Run("pkcs8 parse error", func(t *testing.T) {
		pemData := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("junk bytes")})
		path := writeTempFile(t, pemData)
		if _, err := NewDecryptorFromPrivateKeyFile(path); err == nil {
			t.Fatal("expected parse error")
		}
	})

	t.Run("pkcs1", func(t *testing.T) {
		_, privPath, _ := generateKeyPairFiles(t)
		dec, err := NewDecryptorFromPrivateKeyFile(privPath)
		if err != nil {
			t.Fatalf("expected decryptor, got %v", err)
		}
		if dec == nil {
			t.Fatal("decryptor should not be nil")
		}
	})

	t.Run("pkcs8", func(t *testing.T) {
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("generate key: %v", err)
		}
		raw, err := x509.MarshalPKCS8PrivateKey(priv)
		if err != nil {
			t.Fatalf("marshal pkcs8: %v", err)
		}
		pemData := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: raw})
		path := writeTempFile(t, pemData)

		dec, err := NewDecryptorFromPrivateKeyFile(path)
		if err != nil {
			t.Fatalf("expected decryptor, got %v", err)
		}
		if dec == nil {
			t.Fatal("decryptor should not be nil")
		}
	})
}

func TestHybridEncryptorEncryptErrors(t *testing.T) {
	t.Run("not configured", func(t *testing.T) {
		var enc *HybridEncryptor
		if _, _, err := enc.Encrypt([]byte("hi")); err == nil {
			t.Fatal("expected configuration error")
		}
	})

	t.Run("sym key error", func(t *testing.T) {
		reset := swapRandReader(&stubReader{successes: 0, err: errors.New("rand fail")})
		t.Cleanup(reset)

		enc := &HybridEncryptor{pub: &rsa.PublicKey{}}
		if _, _, err := enc.Encrypt([]byte("hi")); err == nil || !strings.Contains(err.Error(), "generate sym key") {
			t.Fatalf("expected sym key error, got %v", err)
		}
	})

	t.Run("nonce error", func(t *testing.T) {
		reset := swapRandReader(&stubReader{successes: 1, err: errors.New("nonce fail")})
		t.Cleanup(reset)

		enc := &HybridEncryptor{pub: &rsa.PublicKey{}}
		if _, _, err := enc.Encrypt([]byte("hi")); err == nil || !strings.Contains(err.Error(), "generate nonce") {
			t.Fatalf("expected nonce error, got %v", err)
		}
	})

	t.Run("cipher error", func(t *testing.T) {
		resetCipher := swapCipher(func([]byte) (cipher.Block, error) {
			return nil, errors.New("cipher fail")
		})
		resetGCM := swapGCM(cipher.NewGCM)
		resetRand := swapRandReader(&stubReader{successes: 2})
		t.Cleanup(resetCipher)
		t.Cleanup(resetGCM)
		t.Cleanup(resetRand)

		enc := &HybridEncryptor{pub: &rsa.PublicKey{}}
		if _, _, err := enc.Encrypt([]byte("hi")); err == nil || !strings.Contains(err.Error(), "cipher fail") {
			t.Fatalf("expected cipher error, got %v", err)
		}
	})

	t.Run("gcm error", func(t *testing.T) {
		resetCipher := swapCipher(aes.NewCipher)
		resetGCM := swapGCM(func(cipher.Block) (cipher.AEAD, error) { return nil, errors.New("gcm fail") })
		resetRand := swapRandReader(&stubReader{successes: 2})
		t.Cleanup(resetCipher)
		t.Cleanup(resetGCM)
		t.Cleanup(resetRand)

		enc := &HybridEncryptor{pub: &rsa.PublicKey{}}
		if _, _, err := enc.Encrypt([]byte("hi")); err == nil || !strings.Contains(err.Error(), "gcm fail") {
			t.Fatalf("expected gcm error, got %v", err)
		}
	})

	t.Run("encrypt sym key error", func(t *testing.T) {
		resetCipher := swapCipher(aes.NewCipher)
		resetGCM := swapGCM(cipher.NewGCM)
		resetRand := swapRandReader(&stubReader{successes: 3, err: errors.New("rand exhausted")})
		t.Cleanup(resetCipher)
		t.Cleanup(resetGCM)
		t.Cleanup(resetRand)

		enc := &HybridEncryptor{pub: &rsa.PublicKey{}}
		if _, _, err := enc.Encrypt([]byte("hi")); err == nil || !strings.Contains(err.Error(), "encrypt sym key") {
			t.Fatalf("expected rsa encrypt error, got %v", err)
		}
	})
}

func TestHybridEncryptorEncryptDecryptSuccess(t *testing.T) {
	priv, _, pubPath := generateKeyPairFiles(t)
	enc, err := NewEncryptorFromPublicKeyFile(pubPath)
	if err != nil {
		t.Fatalf("encryptor: %v", err)
	}
	dec := &HybridDecryptor{priv: priv}

	ciphertext, key, err := enc.Encrypt([]byte("payload"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	plain, err := dec.Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(plain) != "payload" {
		t.Fatalf("unexpected plain: %s", plain)
	}
}

func TestHybridDecryptorDecryptErrors(t *testing.T) {
	priv, _, pubPath := generateKeyPairFiles(t)
	dec := &HybridDecryptor{priv: priv}

	t.Run("not configured", func(t *testing.T) {
		var empty *HybridDecryptor
		if _, err := empty.Decrypt([]byte("data"), ""); err == nil {
			t.Fatal("expected configuration error")
		}
	})

	t.Run("ciphertext too short", func(t *testing.T) {
		if _, err := dec.Decrypt([]byte{1, 2}, "key"); err == nil || !strings.Contains(err.Error(), "ciphertext too short") {
			t.Fatalf("expected short ciphertext error, got %v", err)
		}
	})

	t.Run("missing encrypted key", func(t *testing.T) {
		if _, err := dec.Decrypt(bytes.Repeat([]byte{1}, nonceSize), ""); err == nil || !strings.Contains(err.Error(), "missing encrypted key") {
			t.Fatalf("expected missing key error, got %v", err)
		}
	})

	t.Run("invalid base64", func(t *testing.T) {
		if _, err := dec.Decrypt(bytes.Repeat([]byte{1}, nonceSize+1), "***"); err == nil || !strings.Contains(err.Error(), "decode key") {
			t.Fatalf("expected decode error, got %v", err)
		}
	})

	t.Run("decrypt key error", func(t *testing.T) {
		encoded := base64.StdEncoding.EncodeToString([]byte("not rsa"))
		if _, err := dec.Decrypt(bytes.Repeat([]byte{1}, nonceSize+1), encoded); err == nil || !strings.Contains(err.Error(), "decrypt key") {
			t.Fatalf("expected decrypt key error, got %v", err)
		}
	})

	t.Run("invalid symmetric length", func(t *testing.T) {
		rsaKey, err := loadPublicKey(pubPath)
		if err != nil {
			t.Fatalf("load public key: %v", err)
		}
		wrapped, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaKey, []byte{1}, nil)
		if err != nil {
			t.Fatalf("encrypt oaep: %v", err)
		}
		encoded := base64.StdEncoding.EncodeToString(wrapped)
		if _, err := dec.Decrypt(bytes.Repeat([]byte{1}, nonceSize+1), encoded); err == nil || !strings.Contains(err.Error(), "invalid symmetric key length") {
			t.Fatalf("expected length error, got %v", err)
		}
	})

	t.Run("cipher error", func(t *testing.T) {
		resetCipher := swapCipher(func([]byte) (cipher.Block, error) { return nil, errors.New("cipher boom") })
		resetGCM := swapGCM(cipher.NewGCM)
		t.Cleanup(resetCipher)
		t.Cleanup(resetGCM)

		keyData := bytes.Repeat([]byte{2}, keySize)
		rsaKey, err := loadPublicKey(pubPath)
		if err != nil {
			t.Fatalf("load public key: %v", err)
		}
		wrapped, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaKey, keyData, nil)
		if err != nil {
			t.Fatalf("encrypt oaep: %v", err)
		}
		encoded := base64.StdEncoding.EncodeToString(wrapped)

		if _, err := dec.Decrypt(bytes.Repeat([]byte{1}, nonceSize+1), encoded); err == nil || !strings.Contains(err.Error(), "cipher boom") {
			t.Fatalf("expected cipher error, got %v", err)
		}
	})

	t.Run("gcm error", func(t *testing.T) {
		resetCipher := swapCipher(aes.NewCipher)
		resetGCM := swapGCM(func(cipher.Block) (cipher.AEAD, error) { return nil, errors.New("gcm boom") })
		t.Cleanup(resetCipher)
		t.Cleanup(resetGCM)

		keyData := bytes.Repeat([]byte{2}, keySize)
		rsaKey, err := loadPublicKey(pubPath)
		if err != nil {
			t.Fatalf("load public key: %v", err)
		}
		wrapped, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaKey, keyData, nil)
		if err != nil {
			t.Fatalf("encrypt oaep: %v", err)
		}
		encoded := base64.StdEncoding.EncodeToString(wrapped)

		if _, err := dec.Decrypt(bytes.Repeat([]byte{1}, nonceSize+1), encoded); err == nil || !strings.Contains(err.Error(), "gcm boom") {
			t.Fatalf("expected gcm error, got %v", err)
		}
	})

	t.Run("decrypt payload error", func(t *testing.T) {
		resetCipher := swapCipher(aes.NewCipher)
		resetGCM := swapGCM(cipher.NewGCM)
		t.Cleanup(resetCipher)
		t.Cleanup(resetGCM)

		enc, err := NewEncryptorFromPublicKeyFile(pubPath)
		if err != nil {
			t.Fatalf("encryptor: %v", err)
		}
		ciphertext, key, err := enc.Encrypt([]byte("payload"))
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}
		ciphertext[len(ciphertext)-1] ^= 0xFF
		if _, err := dec.Decrypt(ciphertext, key); err == nil || !strings.Contains(err.Error(), "decrypt payload") {
			t.Fatalf("expected payload error, got %v", err)
		}
	})
}

func TestDecodeAndDecryptKeyBase64Error(t *testing.T) {
	priv, _, _ := generateKeyPairFiles(t)
	if _, err := decodeAndDecryptKey("***", priv); err == nil || !strings.Contains(err.Error(), "decode key") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestLoadPublicKeyErrors(t *testing.T) {
	t.Run("parse error", func(t *testing.T) {
		path := writeTempFile(t, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: []byte("junk")}))
		if _, err := loadPublicKey(path); err == nil {
			t.Fatal("expected parse error")
		}
	})

	t.Run("non rsa", func(t *testing.T) {
		ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("generate ecdsa: %v", err)
		}
		raw, err := x509.MarshalPKIXPublicKey(&ecdsaKey.PublicKey)
		if err != nil {
			t.Fatalf("marshal ecdsa: %v", err)
		}
		path := writeTempFile(t, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: raw}))
		if _, err := loadPublicKey(path); err == nil || !strings.Contains(err.Error(), "not an RSA public key") {
			t.Fatalf("expected non RSA error, got %v", err)
		}
	})
}

func TestMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("no crypto header", func(t *testing.T) {
		router := gin.New()
		router.POST("/", Middleware(nil), func(c *gin.Context) {
			data, _ := io.ReadAll(c.Request.Body)
			if string(data) != "plain" {
				t.Fatalf("expected plain body, got %s", data)
			}
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("plain"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status %d", rec.Code)
		}
	})

	t.Run("nil decryptor", func(t *testing.T) {
		router := gin.New()
		router.POST("/", Middleware(nil))

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("data"))
		req.Header.Set(CryptoKeyHeader, "key")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("read body error", func(t *testing.T) {
		router := gin.New()
		router.POST("/", Middleware(fakeDecryptor{plain: []byte("ignored")}))

		req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(errReader{}))
		req.Header.Set(CryptoKeyHeader, "key")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("decrypt error", func(t *testing.T) {
		router := gin.New()
		dec := &test.FakeDecryptor{Err: errors.New("boom")}
		router.POST("/", Middleware(dec))

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("data"))
		req.Header.Set(CryptoKeyHeader, "key")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("body too large", func(t *testing.T) {
		router := gin.New()
		called := false
		router.POST("/", Middleware(&test.FakeDecryptor{Plaintext: []byte("ignored")}))

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(bytes.Repeat([]byte("a"), maxEncryptedBodySize+1)))
		req.Header.Set(CryptoKeyHeader, "key")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("expected 413, got %d", rec.Code)
		}
		if called {
			t.Fatal("decryptor should not be called for oversized body")
		}
	})

	t.Run("success", func(t *testing.T) {
		router := gin.New()
		router.POST("/", Middleware(fakeDecryptor{plain: []byte("plain")}), func(c *gin.Context) {
			data, _ := io.ReadAll(c.Request.Body)
			if string(data) != "plain" {
				t.Fatalf("expected decrypted body, got %s", data)
			}
			if c.GetHeader(CryptoKeyHeader) != "" {
				t.Fatal("crypto header should be removed")
			}
			c.Status(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("enc"))
		req.Header.Set(CryptoKeyHeader, "key")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})
}

type fakeDecryptor struct {
	plain []byte
	err   error
}

func (f fakeDecryptor) Decrypt(ciphertext []byte, encryptedKey string) ([]byte, error) {
	return f.plain, f.err
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read error") }
func (errReader) Close() error             { return nil }

func generateKeyPairFiles(t *testing.T) (*rsa.PrivateKey, string, string) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})
	privPath := writeTempFile(t, privPEM)

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("marshal public: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	pubPath := writeTempFile(t, pubPEM)

	return priv, privPath, pubPath
}

func writeTempFile(t *testing.T, data []byte) string {
	t.Helper()

	f, err := os.CreateTemp("", "crypto-*.pem")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if _, err := f.Write(data); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp: %v", err)
	}
	return f.Name()
}

func swapRandReader(r io.Reader) func() {
	old := randReader
	randReader = r
	return func() { randReader = old }
}

func swapCipher(f func([]byte) (cipher.Block, error)) func() {
	old := newAESCipher
	newAESCipher = f
	return func() { newAESCipher = old }
}

func swapGCM(f func(cipher.Block) (cipher.AEAD, error)) func() {
	old := newGCM
	newGCM = f
	return func() { newGCM = old }
}
