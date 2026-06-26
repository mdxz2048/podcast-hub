package http

import (
	stdhttp "net/http"
	"strings"
)

func clientIP(r *stdhttp.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func userAgentSummary(r *stdhttp.Request) string {
	ua := strings.TrimSpace(r.UserAgent())
	if len(ua) > 200 {
		return ua[:200]
	}
	return ua
}
