package config

import "time"

// ServiceConfig represents service configuration
type ServiceConfig struct {
	Name        string
	Environment string
	Port        int
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NATSConfig represents NATS configuration
type NATSConfig struct {
	URL           string
	MaxReconnects int
	ReconnectWait time.Duration
	Timeout       time.Duration
	PingInterval  time.Duration
	MaxPingsOut   int
}

// TenantConfig represents tenant configuration
type TenantConfig struct {
	DefaultCountry      string
	DefaultCurrency     string
	SupportedCountries  []string
	SupportedCurrencies []string
	AllowedCountries    []string
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret         string
	ExpirationTime time.Duration
	Issuer         string
	RefreshTime    time.Duration
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	CORSOrigins           []string
	CORSAllowedOrigins    []string
	CORSAllowedMethods    []string
	CORSAllowedHeaders    []string
	RateLimitPerMinute    int
	RateLimitRequests     int
	RateLimitWindow       time.Duration
	MaxBodySize           int64
	EnableHTTPS           bool
	SessionTimeout        time.Duration
	PasswordMinLength     int
	PasswordRequireUpper  bool
	PasswordRequireLower  bool
	PasswordRequireNumber bool
	PasswordRequireSymbol bool
}
