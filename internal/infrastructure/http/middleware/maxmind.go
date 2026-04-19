package middleware

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

// MaxMindProvider implements GeoProvider using the GeoLite2-Country database.
// The database file must be available locally (e.g. downloaded via CI or
// mounted in the container). The provider is safe for concurrent use.
type MaxMindProvider struct {
	mu   sync.RWMutex
	db   *geoip2.Reader
	path string
}

// NewMaxMindProvider opens the GeoLite2-Country database at the given path.
// If the file does not exist, it returns nil; the geolocation middleware will
// fall back to CDN headers only.
func NewMaxMindProvider(dbPath string) (*MaxMindProvider, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Missing DB is not a fatal error; we just won't resolve IPs.
			return nil, nil
		}
		return nil, err
	}
	return &MaxMindProvider{db: db, path: dbPath}, nil
}

// Country returns the ISO-3166 alpha-2 country code for the IP address.
// If the IP is private, reserved, or not found, an empty string is returned.
func (p *MaxMindProvider) Country(ctx context.Context, ip net.IP) (string, error) {
	if p == nil {
		return "", nil
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.db == nil {
		return "", nil
	}
	record, err := p.db.Country(ip)
	if err != nil {
		// Database errors are expected for private/internal IPs.
		return "", nil
	}
	return record.Country.IsoCode, nil
}

// Close releases the underlying database resources.
func (p *MaxMindProvider) Close() error {
	if p == nil || p.db == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	err := p.db.Close()
	p.db = nil
	return err
}

// Reload attempts to refresh the database file (useful for hot-reloading
// updates). If the new file fails to load, the previous db remains active.
func (p *MaxMindProvider) Reload() error {
	if p == nil {
		return nil
	}
	newDB, err := geoip2.Open(p.path)
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.db != nil {
		_ = p.db.Close()
	}
	p.db = newDB
	return nil
}

// DefaultDBPath returns the conventional path for the GeoLite2-Country mmdb
// file under a data directory. It matches the path used in the Dockerfile.
func DefaultDBPath() string {
	return filepath.Join("data", "GeoLite2-Country.mmdb")
}
