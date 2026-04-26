package config

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string
	Format     string
	Output     string
	File       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

// FeatureFlags represents feature flags
type FeatureFlags struct {
	EnableLiveBetting       bool
	EnableVirtualSports     bool
	EnableJackpot           bool
	EnablePromotions        bool
	EnableNotifications     bool
	EnableAnalytics         bool
	EnableResponsibleGaming bool
	EnableMultiCurrency     bool
	EnableMobileApp         bool
	EnableAPIV2             bool
	EnableBetaFeatures      bool
}
