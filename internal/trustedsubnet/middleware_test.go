package trustedsubnet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/polkiloo/go-musthave-metrics-tppl/internal/server"
)

func TestShouldValidate(t *testing.T) {
	cases := []struct {
		name string
		req  *http.Request
		want bool
	}{
		{name: "nil", req: nil, want: false},
		{name: "non post", req: httptest.NewRequest(http.MethodGet, "/update", nil), want: false},
		{name: "update", req: httptest.NewRequest(http.MethodPost, "/update", nil), want: true},
		{name: "updates", req: httptest.NewRequest(http.MethodPost, "/updates", nil), want: true},
		{name: "other path", req: httptest.NewRequest(http.MethodPost, "/value", nil), want: false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldValidate(tt.req); got != tt.want {
				t.Fatalf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestNewMiddleware_InvalidCIDR(t *testing.T) {
	cfg := &server.AppConfig{TrustedSubnet: "bad"}
	if _, err := NewMiddleware(cfg, nil); err == nil {
		t.Fatalf("expected error for invalid cidr")
	}
}

func TestMiddleware_AllowsTrustedIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &server.AppConfig{TrustedSubnet: "10.0.0.0/24"}
	mw, err := NewMiddleware(cfg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := gin.New()
	r.Use(mw)
	r.POST("/update", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, "/update", nil)
	req.Header.Set("X-Real-IP", "10.0.0.5")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestMiddleware_ForbiddenWhenOutsideSubnet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &server.AppConfig{TrustedSubnet: "10.0.0.0/24"}
	mw, err := NewMiddleware(cfg, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := gin.New()
	r.Use(mw)
	r.POST("/update", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, "/update", nil)
	req.Header.Set("X-Real-IP", "192.168.0.1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func TestMiddleware_SkippedWhenConfigEmpty(t *testing.T) {
	mw, err := NewMiddleware(&server.AppConfig{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mw != nil {
		t.Fatalf("expected nil middleware when trusted subnet is empty")
	}
}
