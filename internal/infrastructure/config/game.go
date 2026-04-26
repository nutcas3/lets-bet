package config

import "time"

// TaxConfig represents tax configuration
type TaxConfig struct {
	Enabled            bool
	DefaultRate        float64
	WHTRate            float64
	GamingTaxRate      float64
	TransactionTaxRate float64
	Currency           string
	AutoCalculate      bool
}

// CrashConfig represents crash game configuration
type CrashConfig struct {
	Enabled       bool
	BaseURL       string
	APIKey        string
	SecretKey     string
	Timeout       time.Duration
	MaxBetAmount  float64
	MinBetAmount  float64
	MaxMultiplier float64
	MinMultiplier float64
	HouseEdge     float64
}

// BetConfig represents betting configuration
type BetConfig struct {
	MinBetAmount    float64
	MaxBetAmount    float64
	MaxParlaySize   int
	MaxOdds         float64
	MinOdds         float64
	StakeTimeout    time.Duration
	SettlementDelay time.Duration
	CancelDelay     time.Duration
}
