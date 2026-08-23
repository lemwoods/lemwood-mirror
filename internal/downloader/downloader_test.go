package downloader

import "testing"

func TestIsSafePathComponent(t *testing.T) {
	for _, name := range []string{"ok.zip", "1.2.3"} {
		if !isSafePathComponent(name) {
			t.Errorf("%q should be safe", name)
		}
	}
	for _, name := range []string{"", ".", "..", "../escape", `..\\escape`, "/tmp"} {
		if isSafePathComponent(name) {
			t.Errorf("%q should be rejected", name)
		}
	}
}
