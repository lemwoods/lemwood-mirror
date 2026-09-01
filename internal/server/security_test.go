package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"lemwood_mirror/internal/config"
)

// CORS 策略：公共 API 通配符放行（*）；Admin 路径不暴露跨域授权，
// 避免放大 CSRF/凭据攻击面。
func TestSecurityMiddlewareCORSPolicy(t *testing.T) {
	cfg := &config.Config{PowEnabled: false, AppealContact: "test-contact", AdminEnabled: false}
	_, handler, _ := setupDownloadHandlerState(t, cfg, 0, "hello")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/stats", nil)
	req.Header.Set("Origin", "https://example.test")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("public API Access-Control-Allow-Origin = %q, want *", got)
	}

	reqAdmin := httptest.NewRequest(http.MethodGet, "/api/v2/admin/config", nil)
	reqAdmin.Header.Set("Origin", "https://example.test")
	recAdmin := httptest.NewRecorder()
	handler.ServeHTTP(recAdmin, reqAdmin)
	if got := recAdmin.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("admin API Access-Control-Allow-Origin = %q, want empty", got)
	}
}
