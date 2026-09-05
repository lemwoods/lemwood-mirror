package geoip

import "testing"

// 验证内嵌 ip2region 数据可解析，以及生产事件写入依赖的国家口径。
func TestLookup(t *testing.T) {
	cases := []struct {
		ip      string
		country string
		ok      bool
	}{
		{"223.155.116.201", "中国", true},
		{"120.230.26.13", "中国", true},
		{"8.8.8.8", "美国", true},
		{"192.168.1.1", "", false},
		{"not-an-ip", "", false},
	}
	for _, c := range cases {
		country, _, _, ok := Lookup(c.ip)
		if ok != c.ok {
			t.Errorf("Lookup(%q) ok=%v, want %v", c.ip, ok, c.ok)
			continue
		}
		if country != c.country {
			t.Errorf("Lookup(%q) country=%q, want %q", c.ip, country, c.country)
		}
	}
}
