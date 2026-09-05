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

// 验证 region 值归一到 34 个一级行政单位（直辖市/自治区/省/特别行政区）的口径。
func TestNormalizeRegion(t *testing.T) {
	cases := map[string]string{
		// 省全称/简称统一
		"广东省": "广东", "甘肃省": "甘肃", "海南": "海南", "黑龙江": "黑龙江",
		// 直辖市
		"北京市": "北京", "上海": "上海", "天津市": "天津", "重庆": "重庆",
		// 自治区全/简称统一
		"广西壮族自治区": "广西", "广西": "广西", "新疆维吾尔自治区": "新疆", "新疆": "新疆",
		"内蒙古自治区": "内蒙古", "宁夏回族自治区": "宁夏", "西藏自治区": "西藏",
		// 特别行政区与台湾视同省级
		"香港": "香港", "香港特别行政区": "香港", "澳门": "澳门", "台湾省": "台湾", "台湾": "台湾",
		// 地级市映射回所属省
		"广州市": "广东", "深圳市": "广东", "南阳市": "河南", "杭州市": "浙江",
		"武汉": "湖北", "成都": "四川", "西安": "陕西", "乌鲁木齐": "新疆",
		// 归一不了的返回空（由聚合方计入「其他」）
		"0": "", "": "", "火星": "",
	}
	for in, want := range cases {
		if got := NormalizeRegion(in); got != want {
			t.Errorf("NormalizeRegion(%q)=%q, want %q", in, got, want)
		}
	}
}
