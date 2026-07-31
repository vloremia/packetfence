package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLiveHealthAndCorrelationID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	res := httptest.NewRecorder()
	(&Router{}).Handler().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	if res.Header().Get("X-Correlation-ID") == "" {
		t.Fatal("missing correlation id")
	}
}

func TestAdminRoutesRequireAPIKey(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/properties", nil)
	res := httptest.NewRecorder()
	(&Router{AdminAPIKey: "expected"}).Handler().ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", res.Code)
	}
}
