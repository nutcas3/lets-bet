package config

import "time"

// SportradarConfig represents Sportradar sports data provider configuration
type SportradarConfig struct {
	APIKey           string
	BaseURL          string
	Timeout          time.Duration
	RateLimit        int
	EnableSoccer     bool
	EnableBasketball bool
	EnableTennis     bool
	EnableCricket    bool
}

// SmileIDConfig represents SmileID identity verification configuration
type SmileIDConfig struct {
	PartnerID            string
	APIKey               string
	BaseURL              string
	Environment          string
	Timeout              time.Duration
	EnableKYC            bool
	EnableIDVerification bool
}
