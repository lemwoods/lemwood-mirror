package netutil

import (
	"net"
	"net/http"
	"strings"
)

// ExtractClientIP returns the canonical client IP from common proxy headers.
func ExtractClientIP(r *http.Request) string {
	remote := normalizeIP(r.RemoteAddr)
	// Forwarding headers are trusted only from local/private reverse proxies.
	// Otherwise a caller could spoof the IP used for rate limits and audit logs.
	if ip := net.ParseIP(remote); ip != nil && (ip.IsLoopback() || ip.IsPrivate()) {
		candidates := []string{
			firstNonEmpty(strings.Split(r.Header.Get("X-Forwarded-For"), ",")),
			r.Header.Get("X-Real-IP"),
		}
		for _, candidate := range candidates {
			if ip := normalizeIP(candidate); ip != "" {
				return ip
			}
		}
	}
	if remote != "" {
		return remote
	}
	return ""
}

func firstNonEmpty(values []string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
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
