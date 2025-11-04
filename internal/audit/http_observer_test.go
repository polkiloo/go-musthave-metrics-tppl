package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPObserver_Notify_SendsRequest(t *testing.T) {
	var received bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("unexpected content type: %s", ct)
		}
		received = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	observer, err := NewHTTPObserver(srv.URL, srv.Client())
	if err != nil {
		t.Fatalf("new observer: %v", err)
	}

	if err := observer.Notify(context.Background(), Event{}); err != nil {
		t.Fatalf("notify: %v", err)
	}

	if !received {
		t.Fatalf("expected request to be received")
	}
}

func TestNewHTTPObserver_InvalidURL(t *testing.T) {
	if _, err := NewHTTPObserver(":://bad", nil); err == nil {
		t.Fatalf("expected error for invalid url")
	}
}
