package middleware

import (
	"net"
	"strings"
)

// ProxyValidator validates whether requests come from trusted proxies
type ProxyValidator struct {
	trustedProxyCIDRs []*net.IPNet
	cloudflareCIDRs   []*net.IPNet
}

// NewProxyValidator creates a new proxy validator with trusted proxy CIDRs
func NewProxyValidator(trustedProxies []string, cloudflareIPRanges []string) (*ProxyValidator, error) {
	validator := &ProxyValidator{
		trustedProxyCIDRs: make([]*net.IPNet, 0),
		cloudflareCIDRs:   make([]*net.IPNet, 0),
	}

	// Parse trusted proxy CIDRs
	for _, cidr := range trustedProxies {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		validator.trustedProxyCIDRs = append(validator.trustedProxyCIDRs, ipNet)
	}

	// Parse Cloudflare CIDRs
	for _, cidr := range cloudflareIPRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		validator.cloudflareCIDRs = append(validator.cloudflareCIDRs, ipNet)
	}

	return validator, nil
}

// IsTrustedProxy checks if the given IP is from a trusted proxy
func (pv *ProxyValidator) IsTrustedProxy(ipStr string) bool {
	// Remove port if present
	ipStr = strings.Split(ipStr, ":")[0]

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check trusted proxy CIDRs
	for _, cidr := range pv.trustedProxyCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}

	// Check Cloudflare CIDRs
	for _, cidr := range pv.cloudflareCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}

	return false
}

// GetClientIP safely extracts the real client IP from the request
// Only trusts X-Forwarded-For if the request came from a trusted proxy
func (pv *ProxyValidator) GetClientIP(remoteAddr string, xForwardedFor string, xRealIP string) string {
	// Remove port from remoteAddr
	remoteAddr = strings.Split(remoteAddr, ":")[0]

	// Only trust proxy headers if the request came from a trusted proxy
	if pv.IsTrustedProxy(remoteAddr) {
		// Trust X-Forwarded-For from trusted proxies
		if xForwardedFor != "" {
			// X-Forwarded-For can contain multiple IPs: client, proxy1, proxy2
			// The leftmost IP is the original client
			parts := strings.Split(xForwardedFor, ",")
			if len(parts) > 0 {
				clientIP := strings.TrimSpace(parts[0])
				if clientIP != "" {
					return clientIP
				}
			}
		}

		// Fall back to X-Real-IP
		if xRealIP != "" {
			return xRealIP
		}
	}

	// Default to RemoteAddr if not from trusted proxy or headers invalid
	return remoteAddr
}

// Default Cloudflare IPv4 ranges (as of 2026)
var defaultCloudflareIPv4Ranges = []string{
	"173.245.48.0/20",
	"103.21.244.0/22",
	"103.22.200.0/22",
	"103.31.4.0/22",
	"141.101.64.0/18",
	"108.162.192.0/18",
	"190.93.240.0/20",
	"188.114.96.0/20",
	"197.234.240.0/22",
	"198.41.128.0/17",
	"162.158.0.0/15",
	"104.16.0.0/13",
	"104.24.0.0/14",
	"172.64.0.0/13",
	"131.0.72.0/22",
	"2400:cb00::/32",
	"2606:4700::/32",
	"2803:f800::/32",
	"2405:b500::/32",
	"2405:8100::/32",
	"2a06:98c0::/29",
	"2c0f:f248::/32",
}

// Default Cloudflare IPv6 ranges (as of 2026)
var defaultCloudflareIPv6Ranges = []string{
	"2606:4700::/32",
	"2803:f800::/32",
	"2405:b500::/32",
	"2405:8100::/32",
	"2a06:98c0::/29",
	"2c0f:f248::/32",
	"2606:4700::/32",
	"2803:f800::/32",
	"2405:b500::/32",
	"2405:8100::/32",
	"2a06:98c0::/29",
	"2c0f:f248::/32",
}

// NewDefaultProxyValidator creates a proxy validator with default Cloudflare ranges
func NewDefaultProxyValidator(trustedProxies []string) (*ProxyValidator, error) {
	allCloudflareRanges := append(defaultCloudflareIPv4Ranges, defaultCloudflareIPv6Ranges...)
	return NewProxyValidator(trustedProxies, allCloudflareRanges)
}
