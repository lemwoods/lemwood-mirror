package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityMiddlewareAllowsPublicCORS(t *testing.T) {
	t.Skip("requires initialized database and blacklist state")
	h := SecurityMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	r := httptest.NewRequest(http.MethodGet, "https://example.test/", nil)
	r.Header.Set("Origin", "https://evil.test")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want *", got)
	}
}
