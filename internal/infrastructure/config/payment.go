package config

import "time"

// MPesaConfig represents MPesa payment gateway configuration
type MPesaConfig struct {
	ConsumerKey        string
	ConsumerSecret     string
	ShortCode          string
	PassKey            string
	InitiatorName      string
	SecurityCredential string
	Environment        string
	CallbackURL        string
	Timeout            time.Duration
}

// FlutterwaveConfig represents Flutterwave payment gateway configuration
type FlutterwaveConfig struct {
	PublicKey     string
	SecretKey     string
	EncryptionKey string
	BaseURL       string
	WebhookSecret string
	Timeout       time.Duration
}
