package httpx

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

func ClientIPMiddleware(trusted []netip.Prefix, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		remote, err := netip.ParseAddr(strings.Trim(host, "[]"))
		if err != nil {
			Fail(w, r, 400, "invalid_remote_address", "invalid client address", nil)
			return
		}
		client := remote
		if trustedAddr(remote, trusted) {
			parts := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
			for i := len(parts) - 1; i >= 0; i-- {
				parsed, e := netip.ParseAddr(strings.TrimSpace(parts[i]))
				if e != nil {
					continue
				}
				if trustedAddr(parsed, trusted) {
					continue
				}
				client = parsed
				break
			}
		}
		next.ServeHTTP(w, r.WithContext(WithClientIP(r.Context(), client.String())))
	})
}
func trustedAddr(a netip.Addr, prefixes []netip.Prefix) bool {
	for _, p := range prefixes {
		if p.Contains(a) {
			return true
		}
	}
	return false
}
