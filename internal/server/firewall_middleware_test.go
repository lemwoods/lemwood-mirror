package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"lemwood_mirror/internal/config"
	"lemwood_mirror/internal/db"
	"lemwood_mirror/internal/firewall"
)

// setupFirewallMiddlewareTest 复用下载处理器的测试环境（临时 SQLite + 文件 + 路由）。
func setupFirewallMiddlewareTest(t *testing.T) http.Handler {
	t.Helper()

	cfg := &config.Config{PowEnabled: false, AppealContact: "test-contact", AdminEnabled: false}
	_, handler, _ := setupDownloadHandlerState(t, cfg, 0, "hello")
	return handler
}

// doRequest 通过 RemoteAddr 模拟客户端来源（上游安全加固后，
// X-Forwarded-For 仅在 RemoteAddr 为回环/内网代理时被信任，单测直接用 RemoteAddr 更直接）。
func doRequest(handler http.Handler, target, clientIP string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if clientIP != "" {
		req.RemoteAddr = clientIP + ":12345"
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestSecurityMiddlewareBlocksCIDRBlacklist(t *testing.T) {
	handler := setupFirewallMiddlewareTest(t)

	if err := db.AddIPToBlacklistWithSource("203.0.113.0/24", "测试网段封禁", "local", "manual"); err != nil {
		t.Fatalf("AddIPToBlacklistWithSource() error = %v", err)
	}
	firewall.Init(firewall.Settings{}, nil, nil)
	t.Cleanup(firewall.Close)
	if err := firewall.RefreshBlacklist(); err != nil {
		t.Fatalf("RefreshBlacklist() error = %v", err)
	}

	rec := doRequest(handler, "/api/v2/stats", "203.0.113.99")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("IP inside banned CIDR: status = %d, want 403, body = %s", rec.Code, rec.Body.String())
	}

	rec = doRequest(handler, "/api/v2/stats", "203.0.114.99")
	if rec.Code == http.StatusForbidden || rec.Code == http.StatusTooManyRequests {
		t.Fatalf("IP outside banned CIDR should pass: status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestSecurityMiddlewareRateLimitReturns429(t *testing.T) {
	handler := setupFirewallMiddlewareTest(t)

	firewall.Init(firewall.Settings{Enabled: true, PerMinute: 2, BanThreshold: 1000}, nil, nil)
	t.Cleanup(firewall.Close)

	for i := 0; i < 2; i++ {
		rec := doRequest(handler, "/api/v2/stats", "198.51.100.7")
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("request %d should not be rate limited", i+1)
		}
	}

	rec := doRequest(handler, "/api/v2/stats", "198.51.100.7")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("third request status = %d, want 429, body = %s", rec.Code, rec.Body.String())
	}
	if retryAfter := rec.Header().Get("Retry-After"); retryAfter == "" {
		t.Fatal("429 response should include Retry-After header")
	}

	// 其他 IP 不受影响
	rec = doRequest(handler, "/api/v2/stats", "198.51.100.8")
	if rec.Code == http.StatusTooManyRequests {
		t.Fatal("other IP should not be rate limited")
	}
}

func TestSecurityMiddlewareWhitelistBypassesRateLimit(t *testing.T) {
	handler := setupFirewallMiddlewareTest(t)

	firewall.Init(firewall.Settings{Enabled: true, PerMinute: 1, BanThreshold: 1}, []string{"198.51.100.7"}, nil)
	t.Cleanup(firewall.Close)

	for i := 0; i < 5; i++ {
		rec := doRequest(handler, "/api/v2/stats", "198.51.100.7")
		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("whitelisted request %d should pass: status = %d", i+1, rec.Code)
		}
	}
}

func TestSecurityMiddlewareRateLimitAutoBans(t *testing.T) {
	handler := setupFirewallMiddlewareTest(t)

	banFunc := func(ip, reason string) error {
		return db.AddIPToBlacklistWithSource(ip, reason, "local", "rate_limit")
	}
	firewall.Init(firewall.Settings{Enabled: true, PerMinute: 1, BanThreshold: 1}, nil, banFunc)
	t.Cleanup(firewall.Close)

	// 第 1 个请求放行，第 2 个超限并达到封禁阈值（自动写黑名单）
	if rec := doRequest(handler, "/api/v2/stats", "198.51.100.7"); rec.Code == http.StatusTooManyRequests {
		t.Fatal("first request should pass")
	}
	rec := doRequest(handler, "/api/v2/stats", "198.51.100.7")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want 429", rec.Code)
	}

	if !db.IsIPBlacklisted("198.51.100.7") {
		t.Fatal("rate limit auto-ban should have written the blacklist")
	}

	// 封禁生效后，后续请求被黑名单检查直接拒绝
	rec = doRequest(handler, "/api/v2/stats", "198.51.100.7")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("banned IP status = %d, want 403, body = %s", rec.Code, rec.Body.String())
	}
}
