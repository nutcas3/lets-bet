package config

// Config is the centralized configuration for all services.
type Config struct {
	Service     ServiceConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	NATS        NATSConfig
	Tenant      TenantConfig
	JWT         JWTConfig
	Security    SecurityConfig
	MPesa       MPesaConfig
	Flutterwave FlutterwaveConfig
	Sportradar  SportradarConfig
	SmileID     SmileIDConfig
	Tax         TaxConfig
	Crash       CrashConfig
	Bet         BetConfig
	Logging     LoggingConfig
	Features    FeatureFlags
}

// ConfigLoader interface for loading configuration
type ConfigLoader interface {
	Load() (*Config, error)
	Validate(*Config) error
}
