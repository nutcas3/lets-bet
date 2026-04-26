package config

// Environment variables
const (
	// Service
	EnvServiceName        = "SERVICE_NAME"
	EnvServiceEnvironment = "SERVICE_ENVIRONMENT"
	EnvServicePort        = "SERVICE_PORT"

	// Database
	EnvDBHost            = "DB_HOST"
	EnvDBPort            = "DB_PORT"
	EnvDBName            = "DB_NAME"
	EnvDBUser            = "DB_USER"
	EnvDBPassword        = "DB_PASSWORD"
	EnvDBSSLMode         = "DB_SSL_MODE"
	EnvDBMaxOpenConns    = "DB_MAX_OPEN_CONNS"
	EnvDBMaxIdleConns    = "DB_MAX_IDLE_CONNS"
	EnvDBConnMaxLifetime = "DB_CONN_MAX_LIFETIME"

	// Redis
	EnvRedisHost         = "REDIS_HOST"
	EnvRedisPort         = "REDIS_PORT"
	EnvRedisPassword     = "REDIS_PASSWORD"
	EnvRedisDB           = "REDIS_DB"
	EnvRedisPoolSize     = "REDIS_POOL_SIZE"
	EnvRedisMinIdleConns = "REDIS_MIN_IDLE_CONNS"
	EnvRedisDialTimeout  = "REDIS_DIAL_TIMEOUT"
	EnvRedisReadTimeout  = "REDIS_READ_TIMEOUT"
	EnvRedisWriteTimeout = "REDIS_WRITE_TIMEOUT"

	// NATS
	EnvNATSURL           = "NATS_URL"
	EnvNATSMaxReconnects = "NATS_MAX_RECONNECTS"
	EnvNATSReconnectWait = "NATS_RECONNECT_WAIT"
	EnvNATSTimeout       = "NATS_TIMEOUT"
	EnvNATSPingInterval  = "NATS_PING_INTERVAL"
	EnvNATSMaxPingsOut   = "NATS_MAX_PINGS_OUT"

	// Tenant
	EnvTenantDefaultCountry  = "TENANT_DEFAULT_COUNTRY"
	EnvTenantDefaultCurrency = "TENANT_DEFAULT_CURRENCY"

	// JWT
	EnvJWTSecret         = "JWT_SECRET"
	EnvJWTExpirationTime = "JWT_EXPIRATION_TIME"
	EnvJWTIssuer         = "JWT_ISSUER"
	EnvJWTRefreshTime    = "JWT_REFRESH_TIME"

	// Security
	EnvSecurityCORSOrigins           = "SECURITY_CORS_ORIGINS"
	EnvSecurityCORSAllowedOrigins    = "SECURITY_CORS_ALLOWED_ORIGINS"
	EnvSecurityCORSAllowedMethods    = "SECURITY_CORS_ALLOWED_METHODS"
	EnvSecurityCORSAllowedHeaders    = "SECURITY_CORS_ALLOWED_HEADERS"
	EnvSecurityRateLimitPerMinute    = "SECURITY_RATE_LIMIT_PER_MINUTE"
	EnvSecurityRateLimitRequests     = "SECURITY_RATE_LIMIT_REQUESTS"
	EnvSecurityRateLimitWindow       = "SECURITY_RATE_LIMIT_WINDOW"
	EnvSecurityMaxBodySize           = "SECURITY_MAX_BODY_SIZE"
	EnvSecurityEnableHTTPS           = "SECURITY_ENABLE_HTTPS"
	EnvSecuritySessionTimeout        = "SECURITY_SESSION_TIMEOUT"
	EnvSecurityPasswordMinLength     = "SECURITY_PASSWORD_MIN_LENGTH"
	EnvSecurityPasswordRequireUpper  = "SECURITY_PASSWORD_REQUIRE_UPPER"
	EnvSecurityPasswordRequireLower  = "SECURITY_PASSWORD_REQUIRE_LOWER"
	EnvSecurityPasswordRequireNumber = "SECURITY_PASSWORD_REQUIRE_NUMBER"
	EnvSecurityPasswordRequireSymbol = "SECURITY_PASSWORD_REQUIRE_SYMBOL"

	// M-Pesa
	EnvMPesaConsumerKey        = "MPESA_CONSUMER_KEY"
	EnvMPesaConsumerSecret     = "MPESA_CONSUMER_SECRET"
	EnvMPesaShortCode          = "MPESA_SHORT_CODE"
	EnvMPesaPassKey            = "MPESA_PASS_KEY"
	EnvMPesaInitiatorName      = "MPESA_INITIATOR_NAME"
	EnvMPesaSecurityCredential = "MPESA_SECURITY_CREDENTIAL"
	EnvMPesaEnvironment        = "MPESA_ENVIRONMENT"
	EnvMPesaCallbackURL        = "MPESA_CALLBACK_URL"
	EnvMPesaTimeout            = "MPESA_TIMEOUT"

	// Flutterwave
	EnvFlutterwavePublicKey     = "FLUTTERWAVE_PUBLIC_KEY"
	EnvFlutterwaveSecretKey     = "FLUTTERWAVE_SECRET_KEY"
	EnvFlutterwaveEncryptionKey = "FLUTTERWAVE_ENCRYPTION_KEY"
	EnvFlutterwaveBaseURL       = "FLUTTERWAVE_BASE_URL"
	EnvFlutterwaveWebhookSecret = "FLUTTERWAVE_WEBHOOK_SECRET"
	EnvFlutterwaveTimeout       = "FLUTTERWAVE_TIMEOUT"

	// Sportradar
	EnvSportradarAPIKey           = "SPORTRADAR_API_KEY"
	EnvSportradarBaseURL          = "SPORTRADAR_BASE_URL"
	EnvSportradarTimeout          = "SPORTRADAR_TIMEOUT"
	EnvSportradarRateLimit        = "SPORTRADAR_RATE_LIMIT"
	EnvSportradarEnableSoccer     = "SPORTRADAR_ENABLE_SOCCER"
	EnvSportradarEnableBasketball = "SPORTRADAR_ENABLE_BASKETBALL"
	EnvSportradarEnableTennis     = "SPORTRADAR_ENABLE_TENNIS"
	EnvSportradarEnableCricket    = "SPORTRADAR_ENABLE_CRICKET"

	// SmileID
	EnvSmileIDPartnerID            = "SMILE_ID_PARTNER_ID"
	EnvSmileIDAPIKey               = "SMILE_ID_API_KEY"
	EnvSmileIDBaseURL              = "SMILE_ID_BASE_URL"
	EnvSmileIDEnvironment          = "SMILE_ID_ENVIRONMENT"
	EnvSmileIDTimeout              = "SMILE_ID_TIMEOUT"
	EnvSmileIDEnableKYC            = "SMILE_ID_ENABLE_KYC"
	EnvSmileIDEnableIDVerification = "SMILE_ID_ENABLE_ID_VERIFICATION"

	// Tax
	EnvTaxEnabled            = "TAX_ENABLED"
	EnvTaxDefaultRate        = "TAX_DEFAULT_RATE"
	EnvTaxWHTRate            = "TAX_WHT_RATE"
	EnvTaxGamingTaxRate      = "TAX_GAMING_TAX_RATE"
	EnvTaxTransactionTaxRate = "TAX_TRANSACTION_TAX_RATE"
	EnvTaxCurrency           = "TAX_CURRENCY"
	EnvTaxAutoCalculate      = "TAX_AUTO_CALCULATE"

	// Crash
	EnvCrashEnabled       = "CRASH_ENABLED"
	EnvCrashBaseURL       = "CRASH_BASE_URL"
	EnvCrashAPIKey        = "CRASH_API_KEY"
	EnvCrashSecretKey     = "CRASH_SECRET_KEY"
	EnvCrashTimeout       = "CRASH_TIMEOUT"
	EnvCrashMaxBetAmount  = "CRASH_MAX_BET_AMOUNT"
	EnvCrashMinBetAmount  = "CRASH_MIN_BET_AMOUNT"
	EnvCrashMaxMultiplier = "CRASH_MAX_MULTIPLIER"
	EnvCrashMinMultiplier = "CRASH_MIN_MULTIPLIER"
	EnvCrashHouseEdge     = "CRASH_HOUSE_EDGE"

	// Bet
	EnvBetMinBetAmount    = "BET_MIN_BET_AMOUNT"
	EnvBetMaxBetAmount    = "BET_MAX_BET_AMOUNT"
	EnvBetMaxParlaySize   = "BET_MAX_PARLAY_SIZE"
	EnvBetMaxOdds         = "BET_MAX_ODDS"
	EnvBetMinOdds         = "BET_MIN_ODDS"
	EnvBetStakeTimeout    = "BET_STAKE_TIMEOUT"
	EnvBetSettlementDelay = "BET_SETTLEMENT_DELAY"
	EnvBetCancelDelay     = "BET_CANCEL_DELAY"

	// Logging
	EnvLoggingLevel      = "LOGGING_LEVEL"
	EnvLoggingFormat     = "LOGGING_FORMAT"
	EnvLoggingOutput     = "LOGGING_OUTPUT"
	EnvLoggingFile       = "LOGGING_FILE"
	EnvLoggingMaxSize    = "LOGGING_MAX_SIZE"
	EnvLoggingMaxBackups = "LOGGING_MAX_BACKUPS"
	EnvLoggingMaxAge     = "LOGGING_MAX_AGE"
	EnvLoggingCompress   = "LOGGING_COMPRESS"

	// Feature Flags
	EnvFeatureEnableLiveBetting       = "FEATURE_ENABLE_LIVE_BETTING"
	EnvFeatureEnableVirtualSports     = "FEATURE_ENABLE_VIRTUAL_SPORTS"
	EnvFeatureEnableJackpot           = "FEATURE_ENABLE_JACKPOT"
	EnvFeatureEnablePromotions        = "FEATURE_ENABLE_PROMOTIONS"
	EnvFeatureEnableNotifications     = "FEATURE_ENABLE_NOTIFICATIONS"
	EnvFeatureEnableAnalytics         = "FEATURE_ENABLE_ANALYTICS"
	EnvFeatureEnableResponsibleGaming = "FEATURE_ENABLE_RESPONSIBLE_GAMING"
	EnvFeatureEnableMultiCurrency     = "FEATURE_ENABLE_MULTI_CURRENCY"
	EnvFeatureEnableMobileApp         = "FEATURE_ENABLE_MOBILE_APP"
	EnvFeatureEnableAPIV2             = "FEATURE_ENABLE_API_V2"
	EnvFeatureEnableBetaFeatures      = "FEATURE_ENABLE_BETA_FEATURES"
)
