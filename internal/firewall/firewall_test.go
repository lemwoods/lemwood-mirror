package firewall

import (
	"sync/atomic"
	"testing"
	"time"

	"lemwood_mirror/internal/config"
	"lemwood_mirror/internal/db"
)

func setupFirewall(t *testing.T, settings Settings, whitelist []string, banFunc BanFunc) *Manager {
	t.Helper()
	Init(settings, whitelist, banFunc)
	t.Cleanup(Close)
	return defaultManager
}

func TestWhitelistMatching(t *testing.T) {
	setupFirewall(t, Settings{}, []string{"192.0.2.1", "10.0.0.0/8"}, nil)

	if !Whitelisted("192.0.2.1") {
		t.Fatal("exact whitelisted IP should match")
	}
	if !Whitelisted("10.1.2.3") {
		t.Fatal("IP inside whitelisted CIDR should match")
	}
	if Whitelisted("203.0.113.5") {
		t.Fatal("non-whitelisted IP should not match")
	}
	if Whitelisted("not-an-ip") {
		t.Fatal("invalid IP should not match")
	}
	if Whitelisted("") {
		t.Fatal("empty IP should not match")
	}
}

func TestAllowWithinLimitThenDeny(t *testing.T) {
	setupFirewall(t, Settings{Enabled: true, PerMinute: 3, BanThreshold: 100}, nil, nil)

	for i := 0; i < 3; i++ {
		if d := Allow("192.0.2.1"); !d.Allowed {
			t.Fatalf("request %d denied unexpectedly: %+v", i+1, d)
		}
	}
	d := Allow("192.0.2.1")
	if d.Allowed {
		t.Fatal("request over limit should be denied")
	}
	if d.Banned {
		t.Fatal("denial below ban threshold should not ban")
	}
	if d.RetryAfter <= 0 || d.RetryAfter > time.Minute {
		t.Fatalf("RetryAfter = %s, want in (0, 1m]", d.RetryAfter)
	}
}

func TestAllowBansAfterThreshold(t *testing.T) {
	var bannedIP atomic.Pointer[string]
	var bannedReason atomic.Pointer[string]
	setupFirewall(t, Settings{Enabled: true, PerMinute: 1, BanThreshold: 2}, nil, func(ip, reason string) error {
		bannedIP.Store(&ip)
		bannedReason.Store(&reason)
		return nil
	})

	m := defaultManager
	start := m.now()

	if d := Allow("192.0.2.1"); !d.Allowed {
		t.Fatalf("first request should pass: %+v", d)
	}
	if d := Allow("192.0.2.1"); d.Banned {
		t.Fatal("first violation should not ban yet")
	}

	// 推进到下一个窗口：违规计数保留，新窗口第一个请求仍放行，
	// 第二个请求超限记违规并达到阈值，应触发封禁
	m.mu.Lock()
	m.now = func() time.Time { return start.Add(61 * time.Second) }
	m.mu.Unlock()

	if d := Allow("192.0.2.1"); !d.Allowed {
		t.Fatalf("first request of a new window should pass: %+v", d)
	}
	d := Allow("192.0.2.1")
	if d.Allowed {
		t.Fatal("request over limit should be denied")
	}
	if !d.Banned {
		t.Fatal("second violation should trigger ban")
	}
	if got := bannedIP.Load(); got == nil || *got != "192.0.2.1" {
		t.Fatalf("banFunc ip = %v, want 192.0.2.1", got)
	}
	if bannedReason.Load() == nil || *bannedReason.Load() == "" {
		t.Fatal("banFunc reason should not be empty")
	}
}

func TestAllowStrikesExpire(t *testing.T) {
	setupFirewall(t, Settings{Enabled: true, PerMinute: 1, BanThreshold: 2}, nil, nil)

	m := defaultManager
	start := m.now()

	Allow("192.0.2.1")
	if d := Allow("192.0.2.1"); d.Banned {
		t.Fatal("first violation should not ban")
	}

	m.mu.Lock()
	m.now = func() time.Time { return start.Add(strikeTTL + time.Minute) }
	m.mu.Unlock()

	if d := Allow("192.0.2.1"); d.Banned {
		t.Fatal("violation strike should expire and not accumulate into a ban")
	}
}

func TestAllowWhitelistedBypassesLimit(t *testing.T) {
	setupFirewall(t, Settings{Enabled: true, PerMinute: 1, BanThreshold: 1}, []string{"192.0.2.1"}, nil)

	for i := 0; i < 5; i++ {
		if d := Allow("192.0.2.1"); !d.Allowed {
			t.Fatalf("whitelisted request %d should pass: %+v", i+1, d)
		}
	}
	// 非白名单 IP 仍受限：其第一次请求放行，第二次超限拒绝
	if d := Allow("203.0.113.9"); !d.Allowed {
		t.Fatalf("non-whitelisted first request should pass: %+v", d)
	}
	if d := Allow("203.0.113.9"); d.Allowed {
		t.Fatal("non-whitelisted second request should be denied")
	}
}

func TestAllowDisabledSettingsAlwaysAllow(t *testing.T) {
	setupFirewall(t, Settings{Enabled: false, PerMinute: 1}, nil, nil)

	for i := 0; i < 5; i++ {
		if d := Allow("192.0.2.1"); !d.Allowed {
			t.Fatalf("disabled limiter should always allow: %+v", d)
		}
	}
}

func TestUpdateSettingsHotReload(t *testing.T) {
	setupFirewall(t, Settings{Enabled: true, PerMinute: 1000, BanThreshold: 3}, nil, nil)

	Allow("192.0.2.1")
	UpdateSettings(Settings{Enabled: true, PerMinute: 1, BanThreshold: 3}, nil)

	if got := GetSettings().PerMinute; got != 1 {
		t.Fatalf("PerMinute after update = %d, want 1", got)
	}
	if d := Allow("192.0.2.1"); d.Allowed {
		t.Fatal("request should be denied after tightening the limit")
	}
}

func TestRefreshBlacklistLoadsCIDREntries(t *testing.T) {
	base := t.TempDir()
	if db.DB != nil {
		_ = db.DB.Close()
		db.DB = nil
	}
	if err := db.InitDB(base, &config.Config{}); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() {
		Close()
		if db.DB != nil {
			_ = db.DB.Close()
			db.DB = nil
		}
	})

	Init(Settings{}, nil, nil)

	if err := db.AddIPToBlacklistWithSource("203.0.113.0/24", "测试网段", "local", "manual"); err != nil {
		t.Fatalf("AddIPToBlacklistWithSource() error = %v", err)
	}
	if err := db.AddIPToBlacklistWithSource("198.51.100.7", "精确 IP", "local", "manual"); err != nil {
		t.Fatalf("AddIPToBlacklistWithSource() error = %v", err)
	}
	if err := RefreshBlacklist(); err != nil {
		t.Fatalf("RefreshBlacklist() error = %v", err)
	}

	if !MatchBlacklistCIDR("203.0.113.55") {
		t.Fatal("IP inside banned CIDR should match")
	}
	if MatchBlacklistCIDR("203.0.114.55") {
		t.Fatal("IP outside banned CIDR should not match")
	}
	// 精确 IP 由 DB 查询承载，不进网段集合
	if MatchBlacklistCIDR("198.51.100.7") {
		t.Fatal("exact IP entry should not be matched as CIDR")
	}
}
