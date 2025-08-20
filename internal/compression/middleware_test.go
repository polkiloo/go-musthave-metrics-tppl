package compression

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/test"
)

func init() { gin.SetMode(gin.TestMode) }

func TestCompressionMiddleware_Basics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fc := test.NewFakeCompressor("gzip")
	fc.CaptureWrites = true

	r := gin.New()
	r.Use(Middleware(fc))
	r.POST("/echo", func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		c.Header("Content-Type", "application/json")
		c.String(200, string(b))
	})

	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(`{"ok":true}`))
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("want gzip response, got %q", got)
	}

	if !bytes.Contains(fc.Written(), []byte(`{"ok":true}`)) {
		t.Fatalf("fake compressor did not capture body")
	}
}

func TestCompressionMiddleware_RequestDecompress(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fc := test.NewFakeCompressor("gzip")
	r := gin.New()
	r.Use(Middleware(fc))
	r.POST("/sum", func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		if string(b) != "hello" {
			t.Fatalf("want body 'hello', got %q", string(b))
		}
		c.String(200, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/sum", strings.NewReader("hello"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("unexpected status %d", w.Code)
	}
}

func TestMW_Compresses_JSON_WhenClientAccepts(t *testing.T) {
	r := gin.New()
	r.Use(Middleware(NewGzip(gzip.BestSpeed)))
	r.GET("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/json", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if ce := w.Header().Get("Content-Encoding"); ce != "gzip" {
		t.Fatalf("want CE=gzip, got %q", ce)
	}
	vary := strings.ToLower(w.Header().Get("Vary"))
	if !strings.Contains(vary, "accept-encoding") {
		t.Fatalf("want Vary contain Accept-Encoding, got %q", vary)
	}
	body := gunzipTB(t, w.Body.Bytes())
	if !strings.Contains(string(body), `"ok":true`) {
		t.Fatalf("unexpected body after gunzip: %s", body)
	}
	if cl := w.Header().Get("Content-Length"); cl != "" {
		t.Fatalf("unexpected Content-Length: %q", cl)
	}
}

func TestMW_Compresses_HTML_WhenClientAccepts(t *testing.T) {
	r := gin.New()
	r.Use(Middleware(NewGzip(gzip.BestSpeed)))
	r.GET("/html", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, "<h1>hi</h1>")
	})

	req := httptest.NewRequest(http.MethodGet, "/html", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("want gzip")
	}
	got := gunzipTB(t, w.Body.Bytes())
	if string(got) != "<h1>hi</h1>" {
		t.Fatalf("got %q", got)
	}
}

func TestMW_NoCompress_WithoutAcceptEncoding(t *testing.T) {
	r := gin.New()
	r.Use(Middleware(NewGzip(gzip.BestSpeed)))
	r.GET("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/json", nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if ce := w.Header().Get("Content-Encoding"); ce != "" {
		t.Fatalf("unexpected CE %q", ce)
	}
}

func TestMW_NoCompress_UnsupportedCT(t *testing.T) {
	r := gin.New()
	r.Use(Middleware(NewGzip(gzip.BestSpeed)))
	want := []byte{0, 1, 2, 3}
	r.GET("/png", func(c *gin.Context) {
		c.Header("Content-Type", "image/png")
		c.Writer.Write(want)
	})

	req := httptest.NewRequest(http.MethodGet, "/png", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if ce := w.Header().Get("Content-Encoding"); ce != "" {
		t.Fatalf("unexpected CE for image/png: %q", ce)
	}
	if !bytes.Equal(w.Body.Bytes(), want) {
		t.Fatalf("body changed")
	}
}

func TestMW_NoCompress_WhenCTNotSetBeforeWrite(t *testing.T) {
	r := gin.New()
	r.Use(Middleware(NewGzip(gzip.BestSpeed)))
	r.GET("/raw", func(c *gin.Context) {
		c.Writer.Write([]byte("raw-bytes"))
	})

	req := httptest.NewRequest(http.MethodGet, "/raw", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if ce := w.Header().Get("Content-Encoding"); ce != "" {
		t.Fatalf("unexpected CE on raw")
	}
}

func TestMW_NoCompress_HEAD_and_204(t *testing.T) {
	r := gin.New()
	r.Use(Middleware(NewGzip(gzip.BestSpeed)))

	r.HEAD("/head", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusOK)
	})
	r.GET("/nocontent", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Status(http.StatusNoContent)
	})

	// HEAD
	{
		req := httptest.NewRequest(http.MethodHead, "/head", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if ce := w.Header().Get("Content-Encoding"); ce != "" {
			t.Fatalf("HEAD: unexpected CE %q", ce)
		}
	}
	// 204
	{
		req := httptest.NewRequest(http.MethodGet, "/nocontent", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if ce := w.Header().Get("Content-Encoding"); ce != "" {
			t.Fatalf("204: unexpected CE %q", ce)
		}
	}
}

func TestMW_RespectsExistingCE(t *testing.T) {
	r := gin.New()
	r.Use(Middleware(NewGzip(gzip.BestSpeed)))
	r.GET("/prece", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Header("Content-Encoding", "deflate") // pre-set
		c.String(http.StatusOK, `{"x":1}`)
	})

	req := httptest.NewRequest(http.MethodGet, "/prece", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if ce := w.Header().Get("Content-Encoding"); ce != "deflate" {
		t.Fatalf("must preserve existing CE, got %q", ce)
	}
}

func TestMW_DecompressesJSONRequest(t *testing.T) {
	r := gin.New()
	r.Use(Middleware(NewGzip(gzip.BestSpeed)))
	r.POST("/echo", func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, string(b))
	})

	orig := `{"a":1}`
	body := gzipBytesTB(t, []byte(orig))

	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if w.Body.String() != orig {
		t.Fatalf("want %q, got %q", orig, w.Body.String())
	}
}

func TestMW_NoDecompress_UnsupportedRequestCT(t *testing.T) {
	r := gin.New()
	r.Use(Middleware(NewGzip(gzip.BestSpeed)))
	r.POST("/img", func(c *gin.Context) {
		b := make([]byte, 2)
		_, _ = c.Request.Body.Read(b)
		if b[0] != 0x1f || b[1] != 0x8b {
			t.Fatalf("request was unexpectedly decompressed")
		}
		c.String(http.StatusOK, "ok")
	})

	body := gzipBytesTB(t, []byte("PNG bytes"))
	req := httptest.NewRequest(http.MethodPost, "/img", bytes.NewReader(body))
	req.Header.Set("Content-Type", "image/png")
	req.Header.Set("Content-Encoding", "gzip")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
}

func TestMW_NewReaderError_Returns400(t *testing.T) {
	fc := test.NewFakeCompressor("gzip")
	fc.ErrNewReader = io.ErrUnexpectedEOF

	r := gin.New()
	r.Use(Middleware(fc))
	r.POST("/x", func(c *gin.Context) {
		c.String(http.StatusOK, "should not be called")
	})

	req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader(gzipBytesTB(t, []byte("abc"))))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400 on reader error, got %d", w.Code)
	}
}

func TestMW_NewWriterError_FallbackToPlain(t *testing.T) {
	fc := test.NewFakeCompressor("gzip")
	fc.ErrNewWriter = io.ErrClosedPipe

	r := gin.New()
	r.Use(Middleware(fc))
	r.GET("/x", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, `{"ok":true}`)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if ce := w.Header().Get("Content-Encoding"); ce != "" {
		t.Fatalf("must fallback to plain, got CE=%q", ce)
	}
	if !strings.Contains(w.Body.String(), `{"ok":true}`) {
		t.Fatalf("plain body expected")
	}
}

func TestMW_NoDecompress_WhenCEIsDifferent(t *testing.T) {
	r := gin.New()
	r.Use(Middleware(NewGzip(gzip.BestSpeed)))
	r.POST("/x", func(c *gin.Context) {
		b := make([]byte, 2)
		_, _ = c.Request.Body.Read(b)
		if b[0] != 0x1f || b[1] != 0x8b {
			t.Fatalf("body should remain gzipped because CE!=gzip")
		}
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader(gzipBytesTB(t, []byte("data"))))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "deflate") // not our compressor

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
}

func gunzipTB(t testing.TB, p []byte) []byte {
	t.Helper()
	gr, err := gzip.NewReader(bytes.NewReader(p))
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	defer gr.Close()
	out, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("gunzip read: %v", err)
	}
	return out
}

func gzipBytesTB(t testing.TB, p []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		t.Fatalf("gzip.NewWriterLevel: %v", err)
	}
	if _, err := zw.Write(p); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("gzip Close: %v", err)
	}
	return buf.Bytes()
}
