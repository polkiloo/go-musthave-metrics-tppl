package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestHandle_InvalidMethod(t *testing.T) {
	h := NewUpdateHandler()
	req := httptest.NewRequest(http.MethodGet, "/update/gauge/x/1.0", nil)
	rec := httptest.NewRecorder()
	h.Handle(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if body := rec.Body.String(); !strings.Contains(body, "method not allowed") {
		t.Errorf("body = %q; want substring %q", body, "method not allowed")
	}
}

func TestHandle_InvalidContentType(t *testing.T) {
	h := NewUpdateHandler()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/x/1.0", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Handle(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}
	if body := rec.Body.String(); !strings.Contains(body, "invalid Content-Type") {
		t.Errorf("body = %q; want substring %q", body, "invalid Content-Type")
	}
}

func TestHandle_BadPath(t *testing.T) {
	h := NewUpdateHandler()
	cases := []string{
		"/update/",
		"/update/gauge/x",
	}
	for _, path := range cases {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("Content-Type", "text/plain")
		rec := httptest.NewRecorder()
		h.Handle(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("path %q -> status = %d; want %d", path, rec.Code, http.StatusNotFound)
		}
	}
}

func TestHandle_UnknownType(t *testing.T) {
	h := NewUpdateHandler()
	req := httptest.NewRequest(http.MethodPost, "/update/unknown/x/1", nil)
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	h.Handle(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), "unknown metric type") {
		t.Errorf("body = %q; want substring %q", rec.Body.String(), "unknown metric type")
	}
}

func TestHandle_InvalidGaugeValue(t *testing.T) {
	h := NewUpdateHandler()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/x/not-float", nil)
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()
	h.Handle(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), "invalid gauge value") {
		t.Errorf("body = %q; want substring %q", rec.Body.String(), "invalid gauge value")
	}
}

func TestHandle_SuccessGaugeAndCounter(t *testing.T) {
	h := NewUpdateHandler()
	reqG := httptest.NewRequest(http.MethodPost, "/update/gauge/temp/12.34", nil)
	reqG.Header.Set("Content-Type", "text/plain")
	recG := httptest.NewRecorder()
	h.Handle(recG, reqG)
	if recG.Code != http.StatusOK {
		t.Errorf("gauge: status = %d; want %d", recG.Code, http.StatusOK)
	}

	reqC := httptest.NewRequest(http.MethodPost, "/update/counter/hits/5", nil)
	reqC.Header.Set("Content-Type", "text/plain")
	recC := httptest.NewRecorder()
	h.Handle(recC, reqC)
	if recC.Code != http.StatusOK {
		t.Errorf("counter: status = %d; want %d", recC.Code, http.StatusOK)
	}
}

func TestHandle_ConcurrentRequests(t *testing.T) {
	h := NewUpdateHandler()
	var wg sync.WaitGroup
	n := 500
	wg.Add(n * 2)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			path := fmt.Sprintf("/update/gauge/g/%d.0", i)
			req := httptest.NewRequest(http.MethodPost, path, nil)
			req.Header.Set("Content-Type", "text/plain")
			rec := httptest.NewRecorder()
			h.Handle(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("gauge concurrent: status = %d; want %d", rec.Code, http.StatusOK)
			}
		}(i)
	}

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodPost, "/update/counter/c/1", nil)
			req.Header.Set("Content-Type", "text/plain")
			rec := httptest.NewRecorder()
			h.Handle(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("counter concurrent: status = %d; want %d", rec.Code, http.StatusOK)
			}
		}()
	}

	wg.Wait()
}
