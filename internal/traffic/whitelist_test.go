package traffic

import (
	"testing"

	"lemwood_mirror/internal/db"
	"lemwood_mirror/internal/firewall"
)

// 白名单 IP 豁免流量预检与自动封禁（手动封禁不受影响，由 SecurityMiddleware 执行）。
func TestWhitelistedIPBypassesTrafficLimit(t *testing.T) {
	tracker := setupTrackerTest(t, 1)
	ip := "127.0.0.1"

	firewall.Init(firewall.Settings{}, []string{ip, "10.0.0.0/8"}, nil)
	t.Cleanup(firewall.Close)

	if !firewall.Whitelisted(ip) {
		t.Fatal("test setup: IP should be whitelisted")
	}

	// 预填超过限额的当日流量
	if err := db.RecordDownloadEvent(db.DownloadEvent{ClientIP: ip, BytesServed: 10 * testGB}); err != nil {
		t.Fatalf("RecordDownloadEvent() error = %v", err)
	}

	allowed, _, _, reason := tracker.ReserveTraffic(ip, 128)
	if !allowed {
		t.Fatalf("whitelisted IP should bypass reservation limit: %s", reason)
	}

	banned, reason, _ := tracker.CheckAndBan(ip)
	if banned {
		t.Fatalf("whitelisted IP should not be auto-banned: %s", reason)
	}
}

// 未命中白名单时行为不变：超限仍拒绝并封禁。
func TestNonWhitelistedIPStillLimited(t *testing.T) {
	tracker := setupTrackerTest(t, 1)

	firewall.Init(firewall.Settings{}, []string{"198.51.100.1"}, nil)
	t.Cleanup(firewall.Close)

	ip := "127.0.0.1"
	if firewall.Whitelisted(ip) {
		t.Fatal("test setup: IP should not be whitelisted")
	}

	if err := db.RecordDownloadEvent(db.DownloadEvent{ClientIP: ip, BytesServed: 2 * testGB}); err != nil {
		t.Fatalf("RecordDownloadEvent() error = %v", err)
	}

	allowed, _, _, _ := tracker.ReserveTraffic(ip, 128)
	if allowed {
		t.Fatal("non-whitelisted IP over limit should be rejected")
	}

	banned, _, _ := tracker.CheckAndBan(ip)
	if !banned {
		t.Fatal("non-whitelisted IP over limit should be auto-banned")
	}
}
