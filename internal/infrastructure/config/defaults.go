package config

import "time"

// Default values
const (
	DefaultServiceName        = "betting-platform"
	DefaultServiceEnvironment = "development"
	DefaultServicePort        = 8080

	DefaultDBHost            = "localhost"
	DefaultDBPort            = 5432
	DefaultDBName            = "betting_platform"
	DefaultDBUser            = "postgres"
	DefaultDBSSLMode         = "disable"
	DefaultDBMaxOpenConns    = 25
	DefaultDBMaxIdleConns    = 5
	DefaultDBConnMaxLifetime = 5 * time.Minute

	DefaultRedisHost         = "localhost"
	DefaultRedisPort         = 6379
	DefaultRedisDB           = 0
	DefaultRedisPoolSize     = 10
	DefaultRedisMinIdleConns = 5
	DefaultRedisDialTimeout  = 5 * time.Second
	DefaultRedisReadTimeout  = 3 * time.Second
	DefaultRedisWriteTimeout = 3 * time.Second

	DefaultNATSURL           = "nats://localhost:4222"
	DefaultNATSMaxReconnects = 5
	DefaultNATSReconnectWait = 2 * time.Second
	DefaultNATSTimeout       = 5 * time.Second
	DefaultNATSPingInterval  = 2 * time.Minute
	DefaultNATSMaxPingsOut   = 3

	DefaultTenantDefaultCountry  = "KE"
	DefaultTenantDefaultCurrency = "KES"

	DefaultJWTExpirationTime = 24 * time.Hour
	DefaultJWTRefreshTime    = 168 * time.Hour
	DefaultJWTIssuer         = "betting-platform"

	DefaultSecurityRateLimitPerMinute    = 100
	DefaultSecurityMaxBodySize           = 10 * 1024 * 1024
	DefaultSecurityEnableHTTPS           = true
	DefaultSecuritySessionTimeout        = 24 * time.Hour
	DefaultSecurityPasswordMinLength     = 8
	DefaultSecurityPasswordRequireUpper  = true
	DefaultSecurityPasswordRequireLower  = true
	DefaultSecurityPasswordRequireNumber = true
	DefaultSecurityPasswordRequireSymbol = false

	DefaultMPesaEnvironment = "sandbox"
	DefaultMPesaTimeout     = 30 * time.Second

	DefaultFlutterwaveBaseURL = "https://api.flutterwave.com/v3"
	DefaultFlutterwaveTimeout = 30 * time.Second

	DefaultSportradarBaseURL          = "https://api.sportradar.com"
	DefaultSportradarTimeout          = 30 * time.Second
	DefaultSportradarRateLimit        = 100
	DefaultSportradarEnableSoccer     = true
	DefaultSportradarEnableBasketball = true
	DefaultSportradarEnableTennis     = true
	DefaultSportradarEnableCricket    = true

	DefaultSmileIDEnvironment          = "sandbox"
	DefaultSmileIDTimeout              = 30 * time.Second
	DefaultSmileIDEnableKYC            = true
	DefaultSmileIDEnableIDVerification = true

	DefaultTaxEnabled            = true
	DefaultTaxDefaultRate        = 0.16
	DefaultTaxWHTRate            = 0.15
	DefaultTaxGamingTaxRate      = 0.20
	DefaultTaxTransactionTaxRate = 0.00
	DefaultTaxCurrency           = "KES"
	DefaultTaxAutoCalculate      = true

	DefaultCrashEnabled       = false
	DefaultCrashTimeout       = 30 * time.Second
	DefaultCrashMaxBetAmount  = 10000.0
	DefaultCrashMinBetAmount  = 10.0
	DefaultCrashMaxMultiplier = 1000.0
	DefaultCrashMinMultiplier = 1.1
	DefaultCrashHouseEdge     = 0.05

	DefaultBetMinBetAmount    = 10.0
	DefaultBetMaxBetAmount    = 100000.0
	DefaultBetMaxParlaySize   = 10
	DefaultBetMaxOdds         = 1000.0
	DefaultBetMinOdds         = 1.1
	DefaultBetStakeTimeout    = 30 * time.Second
	DefaultBetSettlementDelay = 5 * time.Minute
	DefaultBetCancelDelay     = 1 * time.Minute

	DefaultLoggingLevel      = "info"
	DefaultLoggingFormat     = "json"
	DefaultLoggingOutput     = "stdout"
	DefaultLoggingFile       = "logs/app.log"
	DefaultLoggingMaxSize    = 100
	DefaultLoggingMaxBackups = 3
	DefaultLoggingMaxAge     = 28
	DefaultLoggingCompress   = true

	DefaultFeatureEnableLiveBetting       = true
	DefaultFeatureEnableVirtualSports     = true
	DefaultFeatureEnableJackpot           = true
	DefaultFeatureEnablePromotions        = true
	DefaultFeatureEnableNotifications     = true
	DefaultFeatureEnableAnalytics         = true
	DefaultFeatureEnableResponsibleGaming = true
	DefaultFeatureEnableMultiCurrency     = false
	DefaultFeatureEnableMobileApp         = false
	DefaultFeatureEnableAPIV2             = false
	DefaultFeatureEnableBetaFeatures      = false
)
