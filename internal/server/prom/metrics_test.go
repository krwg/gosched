package prom_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krwg/gosched/internal/server/prom"
)

func TestHandlerAndRecord(t *testing.T) {
	prom.RecordRun("EDF", 5, 1, 12.5, 80.0, 3)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	prom.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	body := rec.Body.String()
	if len(body) < 50 {
		t.Fatal("expected metrics body")
	}
}
