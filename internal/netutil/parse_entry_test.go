package netutil

import "testing"

func TestParseEntryAndValidEntry(t *testing.T) {
	tests := []struct {
		name     string
		entry    string
		wantIP   bool
		wantCIDR bool
		wantOK   bool
	}{
		{name: "exact ipv4", entry: "1.2.3.4", wantIP: true, wantOK: true},
		{name: "exact ipv6", entry: "::1", wantIP: true, wantOK: true},
		{name: "ipv4 cidr", entry: "10.0.0.0/8", wantCIDR: true, wantOK: true},
		{name: "ipv6 cidr", entry: "2001:db8::/32", wantCIDR: true, wantOK: true},
		{name: "host bits tolerated by ParseCIDR", entry: "10.1.2.3/8", wantCIDR: true, wantOK: true},
		{name: "garbage", entry: "not-an-ip", wantOK: false},
		{name: "empty", entry: "  ", wantOK: false},
		{name: "broken cidr", entry: "10.0.0.0/", wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ip, cidr, ok := ParseEntry(tc.entry)
			if ok != tc.wantOK {
				t.Fatalf("ParseEntry(%q) ok = %v, want %v", tc.entry, ok, tc.wantOK)
			}
			if (ip != nil) != tc.wantIP {
				t.Fatalf("ParseEntry(%q) exact ip presence = %v, want %v", tc.entry, ip != nil, tc.wantIP)
			}
			if (cidr != nil) != tc.wantCIDR {
				t.Fatalf("ParseEntry(%q) cidr presence = %v, want %v", tc.entry, cidr != nil, tc.wantCIDR)
			}
			if ValidEntry(tc.entry) != tc.wantOK {
				t.Fatalf("ValidEntry(%q) = %v, want %v", tc.entry, ValidEntry(tc.entry), tc.wantOK)
			}
		})
	}
}
