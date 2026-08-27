package netutil

import (
	"net"
	"net/http"
	"strings"
)

// ExtractClientIP returns the canonical client IP from common proxy headers.
//
// Forwarding headers are trusted only from local/private reverse proxies,
// otherwise a caller could spoof the IP used for rate limits and audit logs.
// Within X-Forwarded-For we walk from right to left and take the first entry
// that is NOT a trusted (loopback/private) proxy address: proxies append the
// peer address they saw, so the client-supplied entries live on the left and
// must be ignored to prevent bypassing rate limits / bans via forged headers.
func ExtractClientIP(r *http.Request) string {
	remote := normalizeIP(r.RemoteAddr)
	if ip := net.ParseIP(remote); ip != nil && (ip.IsLoopback() || ip.IsPrivate()) {
		remote = trustedClientIPFromForwardHeaders(r, remote)
	}
	return remote
}

// trustedClientIPFromForwardHeaders resolves the real client IP from forwarding
// headers behind a trusted proxy chain.
func trustedClientIPFromForwardHeaders(r *http.Request, remoteAddr string) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		if xri := normalizeIP(r.Header.Get("X-Real-IP")); xri != "" {
			return xri
		}
		return remoteAddr
	}

	parts := strings.Split(xff, ",")
	// 从右往左找第一个非可信代理地址（即真正的客户端来源）
	for i := len(parts) - 1; i >= 0; i-- {
		candidate := normalizeIP(parts[i])
		if candidate == "" {
			continue
		}
		if ip := net.ParseIP(candidate); ip != nil && (ip.IsLoopback() || ip.IsPrivate()) {
			continue // 可信代理节点写入的条目，继续向左找客户端
		}
		return candidate
	}

	// 整条链都是内网/回环地址时无真实公网来源可取，
	// 回退到最近一跳代理的 RemoteAddr，绝不采用客户端可注入的最左值。
	return remoteAddr
}

func normalizeIP(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}

	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")

	return value
}
