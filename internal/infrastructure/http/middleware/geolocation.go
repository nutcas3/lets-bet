package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/betting-platform/internal/infrastructure/logging"
)

// GeoProvider resolves a client IP to an ISO-3166 alpha-2 country code.
// Implementations can use MaxMind, ipinfo.io, IP2Location, etc.
type GeoProvider interface {
	Country(ctx context.Context, ip net.IP) (string, error)
}

// GeoConfig controls the fencing middleware.
type GeoConfig struct {
	// Provider resolves IPs -> country codes. If nil, the middleware trusts
	// upstream `CF-IPCountry` / `X-Country-Code` headers only.
	Provider GeoProvider
	// Allowed is the whitelist of ISO codes (e.g. ["KE"]). Empty = allow all.
	Allowed []string
	// HeaderFallback is the request header that carries a trusted country code
	// when a CDN or edge proxy resolves it for us. "" disables it.
	// Defaults to "CF-IPCountry".
	HeaderFallback string
	// ProxyValidator validates proxy IPs to prevent header spoofing
	ProxyValidator *ProxyValidator
}

type geoKey struct{}

// CountryFromContext returns the resolved country, if any, for the request.
func CountryFromContext(ctx context.Context) (string, bool) {
	c, ok := ctx.Value(geoKey{}).(string)
	return c, ok && c != ""
}

// Geolocation returns a middleware that annotates the request context with the
// detected country and rejects requests from disallowed regions with 451.
func Geolocation(cfg GeoConfig) func(http.Handler) http.Handler {
	if cfg.HeaderFallback == "" {
		cfg.HeaderFallback = "CF-IPCountry"
	}
	allowed := make(map[string]struct{}, len(cfg.Allowed))
	for _, c := range cfg.Allowed {
		allowed[strings.ToUpper(c)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := logging.FromContext(r.Context())

			country := strings.ToUpper(r.Header.Get(cfg.HeaderFallback))
			if country == "" && cfg.Provider != nil {
				if ip := parseIP(clientIP(r, cfg.ProxyValidator)); ip != nil {
					if c, err := cfg.Provider.Country(r.Context(), ip); err == nil {
						country = strings.ToUpper(c)
					} else {
						logger.Warn("geoip lookup failed", "error", err)
					}
				}
			}

			if len(allowed) > 0 && country != "" {
				if _, ok := allowed[country]; !ok {
					http.Error(w, "service unavailable in your country", http.StatusUnavailableForLegalReasons)
					return
				}
			}

			ctx := context.WithValue(r.Context(), geoKey{}, country)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func parseIP(s string) net.IP {
	if s == "" {
		return nil
	}
	// Strip port if present.
	if host, _, err := net.SplitHostPort(s); err == nil {
		s = host
	}
	return net.ParseIP(s)
}
