package config

import (
	"bytes"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNormalizeConfigFirewallDefaults(t *testing.T) {
	cfg := &Config{StoragePath: "download"}
	if err := NormalizeConfig(cfg); err != nil {
		t.Fatalf("NormalizeConfig() error = %v", err)
	}
	if cfg.RateLimitPerMinute != 300 {
		t.Fatalf("RateLimitPerMinute = %d, want 300", cfg.RateLimitPerMinute)
	}
	if cfg.RateLimitBanThreshold != 3 {
		t.Fatalf("RateLimitBanThreshold = %d, want 3", cfg.RateLimitBanThreshold)
	}
}

func TestNormalizeConfigRejectsInvalidWhitelistEntry(t *testing.T) {
	cfg := &Config{
		StoragePath:       "download",
		FirewallWhitelist: []string{"192.168.1.1", "not-a-valid-entry"},
	}
	if err := NormalizeConfig(cfg); err == nil {
		t.Fatal("NormalizeConfig() should reject invalid whitelist entries")
	}

	cfg = &Config{
		StoragePath:       "download",
		FirewallWhitelist: []string{"192.168.1.1", "10.0.0.0/8"},
	}
	if err := NormalizeConfig(cfg); err != nil {
		t.Fatalf("NormalizeConfig() error = %v", err)
	}
}

// renderYAML 是后台保存与「配置项缺失自动补充」的唯一出口，
// 模板渲染结果必须能被 yaml.Unmarshal 无损回读（否则每次启动都会重写 config.yaml）。
func TestRenderYAMLFirewallRoundTrip(t *testing.T) {
	original := DefaultConfig()
	// 提供管理员凭据，避免 NormalizeConfig 因空凭据自动禁用后台（admin_enabled 翻转属预期行为，与本测试无关）
	original.AdminUser = "admin"
	original.AdminPassword = "test-password"
	original.FirewallWhitelist = []string{"192.168.1.1", "10.0.0.0/8"}
	original.RateLimitPerMinute = 120

	first, err := original.renderYAML()
	if err != nil {
		t.Fatalf("renderYAML() error = %v", err)
	}

	var parsed Config
	if err := yaml.Unmarshal(first, &parsed); err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v", err)
	}
	if err := NormalizeConfig(&parsed); err != nil {
		t.Fatalf("NormalizeConfig() error = %v", err)
	}

	second, err := parsed.renderYAML()
	if err != nil {
		t.Fatalf("renderYAML() error = %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Fatalf("renderYAML round trip is not stable:\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}

	if len(parsed.FirewallWhitelist) != 2 || parsed.FirewallWhitelist[0] != "192.168.1.1" {
		t.Fatalf("FirewallWhitelist round trip = %v", parsed.FirewallWhitelist)
	}
	if !parsed.RateLimitEnabled || parsed.RateLimitPerMinute != 120 {
		t.Fatalf("rate limit round trip: enabled=%v perMinute=%d", parsed.RateLimitEnabled, parsed.RateLimitPerMinute)
	}
}
