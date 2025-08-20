package compression

import (
	"compress/gzip"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

func TestRespWriter_StartsOnFirstWrite_AndOnlyWhenCTAllowed(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	base := c.Writer
	base.Header().Set("Content-Type", "application/json")
	rw := newRespWriter(base, NewGzip(gzip.BestSpeed), []string{"application/json", "text/html"})

	rw.WriteHeaderNow()
	if got := base.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("WriteHeaderNow must not set CE, got %q", got)
	}

	if _, err := rw.Write([]byte(`{"x":1}`)); err != nil {
		t.Fatalf("write: %v", err)
	}
	if got := base.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("want CE=gzip after first write, got %q", got)
	}
	rw.Close()
}

// если Content-Type не установлен — компрессия не включается
func TestRespWriter_NoCompress_WhenCTEmpty(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	base := c.Writer

	rw := newRespWriter(base, NewGzip(gzip.BestSpeed), []string{"application/json", "text/html"})

	if _, err := rw.Write([]byte("raw")); err != nil {
		t.Fatalf("write: %v", err)
	}
	if ce := base.Header().Get("Content-Encoding"); ce != "" {
		t.Fatalf("unexpected CE %q", ce)
	}
	rw.Close()
}

// если CE уже установлен раньше — respwriter его уважает и не переопределяет
func TestRespWriter_RespectsExistingCE(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	base := c.Writer

	base.Header().Set("Content-Type", "application/json")
	base.Header().Set("Content-Encoding", "deflate")

	rw := newRespWriter(base, NewGzip(gzip.BestSpeed), []string{"application/json", "text/html"})

	if _, err := rw.Write([]byte(`{"x":1}`)); err != nil {
		t.Fatalf("write: %v", err)
	}
	if ce := base.Header().Get("Content-Encoding"); ce != "deflate" {
		t.Fatalf("must keep existing CE, got %q", ce)
	}
	rw.Close()
}
