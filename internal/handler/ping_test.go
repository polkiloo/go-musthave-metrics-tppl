package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pashagolub/pgxmock/v4"
)

func TestRegisterPing_OK(t *testing.T) {
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	pool.ExpectPing().WillReturnError(nil)

	gin.SetMode(gin.TestMode)
	h := &GinHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	h.Ping(c, pool)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if err := pool.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRegisterPing_Error(t *testing.T) {
	pool, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool: %v", err)
	}
	defer pool.Close()
	pool.ExpectPing().WillReturnError(errors.New("fail"))

	gin.SetMode(gin.TestMode)
	h := &GinHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	h.Ping(c, pool)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestRegisterPing_NoPool(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &GinHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	h.Ping(c, nil)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
