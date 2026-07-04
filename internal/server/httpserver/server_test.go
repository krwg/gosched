package httpserver_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krwg/gosched/internal/server/httpserver"
)

func TestHealth(t *testing.T) {
	srv := httpserver.New(httpserver.Options{Addr: ":0"})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestScheduleAPI(t *testing.T) {
	srv := httpserver.New(httpserver.Options{Addr: ":0"})
	body := []byte(`{
		"algorithm": "edf",
		"workers": 2,
		"tasks": [
			{"id":"a","name":"A","duration":10,"deadline":100,"priority":1,"arrival_time":0}
		]
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/schedule", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Metrics struct {
			Completed int `json:"completed"`
		} `json:"metrics"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Metrics.Completed != 1 {
		t.Fatalf("completed=%d", payload.Metrics.Completed)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	srv := httpserver.New(httpserver.Options{Addr: ":0"})
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
}
