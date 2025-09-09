package sign

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMiddleware_SignsAndVerifies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware(NewSignerSHA256(), SignKey("secret")))
	r.POST("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	body := []byte("hello")
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("HashSHA256", NewSignerSHA256().Sign(body, SignKey("secret")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", w.Code, http.StatusOK)
	}
	want := NewSignerSHA256().Sign([]byte("ok"), SignKey("secret"))
	if got := w.Header().Get("HashSHA256"); got != want {
		t.Fatalf("header mismatch: got %q want %q", got, want)
	}
}

func TestMiddleware_MissingHash(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware(NewSignerSHA256(), SignKey("secret")))
	r.POST("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("hello")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", w.Code, http.StatusOK)
	}
	want := NewSignerSHA256().Sign([]byte("ok"), SignKey("secret"))
	if got := w.Header().Get("HashSHA256"); got != want {
		t.Fatalf("header mismatch: got %q want %q", got, want)
	}
}

func TestMiddleware_BadHash(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware(NewSignerSHA256(), SignKey("secret")))
	r.POST("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("hello")))
	req.Header.Set("HashSHA256", "bad")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want %d", w.Code, http.StatusBadRequest)
	}
}

func TestMiddleware_DefaultStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware(NewSignerSHA256(), SignKey("secret")))
	r.POST("/test", func(c *gin.Context) {
		_, _ = c.Writer.Write([]byte("ok"))
	})

	body := []byte("hello")
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("HashSHA256", NewSignerSHA256().Sign(body, SignKey("secret")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", w.Code, http.StatusOK)
	}
	want := NewSignerSHA256().Sign([]byte("ok"), SignKey("secret"))
	if got := w.Header().Get("HashSHA256"); got != want {
		t.Fatalf("header mismatch: got %q want %q", got, want)
	}
}

func TestMiddleware_NoKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware(NewSignerSHA256(), ""))
	r.POST("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("body")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", w.Code, http.StatusOK)
	}
	if h := w.Header().Get("HashSHA256"); h != "" {
		t.Fatalf("unexpected header %q", h)
	}
}

type errReader struct{}

func (errReader) Read(p []byte) (int, error) { return 0, errors.New("read error") }
func (errReader) Close() error               { return nil }

func TestMiddleware_ReadError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware(NewSignerSHA256(), SignKey("secret")))
	r.POST("/test", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	req := httptest.NewRequest(http.MethodPost, "/test", errReader{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want %d", w.Code, http.StatusBadRequest)
	}
}

func TestSignWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	w := &signWriter{ResponseWriter: ctx.Writer}

	if w.Written() {
		t.Fatalf("expected not written")
	}
	if w.Status() != 0 {
		t.Fatalf("unexpected status %d", w.Status())
	}
	n, err := w.Write([]byte("hi"))
	if err != nil || n != 2 {
		t.Fatalf("write error: %v n=%d", err, n)
	}
	if !w.Written() {
		t.Fatalf("expected written")
	}
	if w.Size() != 2 {
		t.Fatalf("size=%d want %d", w.Size(), 2)
	}
	m, err := w.WriteString("!")
	if err != nil || m != 1 {
		t.Fatalf("writeString error: %v m=%d", err, m)
	}
	if w.Size() != 3 {
		t.Fatalf("size=%d want %d", w.Size(), 3)
	}
	w.WriteHeader(201)
	if w.Status() != 201 {
		t.Fatalf("status=%d want %d", w.Status(), 201)
	}
	w.WriteHeaderNow()
}
