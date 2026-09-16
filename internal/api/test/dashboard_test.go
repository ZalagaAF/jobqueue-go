package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDashboard_ReturnsHTML(t *testing.T) {
	srv, _, _ := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()

	srv.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, esperaba 200", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("Content-Type = %q, esperaba que empezara con text/html", contentType)
	}
}