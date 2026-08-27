// Package version provides unified SemVer-like version comparison
// used by both the server (launcher index) and selfupdate subsystems.
package version

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// IsStable reports whether a version string represents a stable release
// (i.e. not alpha, beta, rc, snapshot, pre, preview, or dev).
func IsStable(v string) bool {
	vLower := strings.ToLower(v)
	keywords := []string{"alpha", "beta", "rc", "snapshot", "pre", "dev", "preview"}
	for _, k := range keywords {
		if strings.Contains(vLower, k) {
			return false
		}
	}
	// 额外检查：如果包含横杠，通常也是非稳定版（如 1.2.3-v1）
	// 但有些启动器可能使用横杠作为正常版本号的一部分，所以以关键词优先
	return true
}

// IsParseable reports whether a version string looks like a semantic version
// (starts with a digit, optionally after a "v" prefix). Non-parseable tags
// like "alpha2" or "dev" are excluded from selfupdate version comparison.
func IsParseable(v string) bool {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if v == "" {
		return false
	}
	return unicode.IsDigit(rune(v[0]))
}

// SplitPreRelease splits "1.2.3-beta.1" into core "1.2.3" and suffix "beta.1".
func SplitPreRelease(v string) (string, string) {
	if idx := strings.Index(v, "-"); idx >= 0 {
		return v[:idx], v[idx+1:]
	}
	return v, ""
}

func parseFirstInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// Compare returns 1 if v1 > v2, -1 if v1 < v2, 0 if equal.
// Versions are compared segment by segment with numeric priority;
// a SemVer-style pre-release suffix (after "-") is ranked lower than the
// same version without a suffix.
func Compare(v1, v2 string) int {
	if v1 == v2 {
		return 0
	}

	v1Core, v1Pre := SplitPreRelease(strings.TrimPrefix(v1, "v"))
	v2Core, v2Pre := SplitPreRelease(strings.TrimPrefix(v2, "v"))

	parts1 := strings.Split(v1Core, ".")
	parts2 := strings.Split(v2Core, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 string
		if i < len(parts1) {
			p1 = parts1[i]
		}
		if i < len(parts2) {
			p2 = parts2[i]
		}

		if p1 == p2 {
			continue
		}

		n1, err1 := parseFirstInt(p1)
		n2, err2 := parseFirstInt(p2)

		if err1 == nil && err2 == nil {
			if n1 > n2 {
				return 1
			}
			if n1 < n2 {
				return -1
			}
			// 数字部分相同，按字符串字典序（如 2.0.0_beta-1 vs 2.0.0_beta-2）
			if p1 > p2 {
				return 1
			}
			if p1 < p2 {
				return -1
			}
		} else {
			if p1 > p2 {
				return 1
			}
			if p1 < p2 {
				return -1
			}
		}
	}

	// SemVer: 无 pre-release 的版本高于带 pre-release 的相同核心版本
	if cmp := comparePrerelease(v1Pre, v2Pre); cmp != 0 {
		return cmp
	}
	return 0
}

// comparePrerelease 按 SemVer §11 比较两个 pre-release 标识符串。
// 标识符按 "." 切分：纯数字标识符按数值比较且低于字母数字标识符；
// 字母数字标识符按 ASCII 字典序；标识符集合更长的版本更高。
func comparePrerelease(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" {
		return 1 // 无 pre-release 高于有 pre-release
	}
	if b == "" {
		return -1
	}

	idsA := strings.Split(a, ".")
	idsB := strings.Split(b, ".")
	for i := 0; i < len(idsA) && i < len(idsB); i++ {
		x, y := idsA[i], idsB[i]
		xNum, xErr := strconv.Atoi(x)
		yNum, yErr := strconv.Atoi(y)
		switch {
		case xErr == nil && yErr == nil:
			if xNum != yNum {
				if xNum > yNum {
					return 1
				}
				return -1
			}
		case xErr == nil: // 数字标识符低于字母数字标识符
			return -1
		case yErr == nil:
			return 1
		default:
			if x != y {
				if x > y {
					return 1
				}
				return -1
			}
		}
	}
	switch {
	case len(idsA) < len(idsB):
		return -1
	case len(idsA) > len(idsB):
		return 1
	}
	return 0
}
