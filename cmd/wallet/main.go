package main

import (
	"context"
	"fmt"
	"os"

	"github.com/betting-platform/internal/infrastructure/config"
	"github.com/betting-platform/internal/infrastructure/database"
	wallethttp "github.com/betting-platform/internal/infrastructure/http"
	"github.com/betting-platform/internal/infrastructure/http/health"
	"github.com/betting-platform/internal/infrastructure/http/middleware"
	"github.com/betting-platform/internal/infrastructure/logging"
	"github.com/betting-platform/internal/infrastructure/metrics"
	"github.com/betting-platform/internal/infrastructure/ratelimit"
	"github.com/betting-platform/internal/infrastructure/server"
	"github.com/betting-platform/internal/infrastructure/tracing"
	"github.com/gorilla/mux"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	logger := logging.Setup(cfg.Logging.Level, cfg.Logging.Format)
	logger.Info("starting wallet", "env", cfg.Service.Environment)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize OpenTelemetry tracing
	tracerCfg := tracing.DefaultConfig("wallet")
	cleanup, err := tracing.InitTracer(ctx, tracerCfg)
	if err != nil {
		logger.Error("failed to initialize tracer", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	// Initialize database
	db, err := database.NewPostgresConnection(database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Services will be initialized when wallet handler is updated

	// Initialize MaxMind provider (optional)
	geoProvider, err := middleware.NewMaxMindProvider(middleware.DefaultDBPath())
	if err != nil {
		logger.Warn("failed to load maxmind database", "error", err)
	}

	r := mux.NewRouter()

	// Initialize metrics
	rec := metrics.New("wallet")
	rec.RegisterRoutes(r)

	// Apply middleware stack
	r.Use(tracing.HTTPMiddleware("wallet"))
	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(middleware.CORS(cfg.Security))
	r.Use(rec.Middleware)
	r.Use(middleware.Geolocation(middleware.GeoConfig{
		Provider: geoProvider,
		Allowed:  cfg.Tenant.AllowedCountries,
	}))

	// Use Redis-backed rate limiter to prevent memory leak DDoS vulnerability
	rateLimitConfig := &ratelimit.Config{
		IPRequestsPerWindow:     cfg.Security.RateLimitRequests,
		IPWindow:                cfg.Security.RateLimitWindow,
		UserRequestsPerWindow:   cfg.Security.RateLimitRequests * 2, // Allow more for authenticated users
		UserWindow:              cfg.Security.RateLimitWindow,
		GlobalRequestsPerWindow: cfg.Security.RateLimitRequests * 10, // Global limit
		GlobalWindow:            cfg.Security.RateLimitWindow,
		RedisAddr:               cfg.Redis.Addr(),
		RedisPassword:           cfg.Redis.Password,
		RedisDB:                 0,
		UserPrefix:              "rate_limit:wallet:user:",
		IPPrefix:                "rate_limit:wallet:ip:",
		GlobalPrefix:            "rate_limit:wallet:global:",
	}

	// Use existing Redis rate limiter directly
	redisRateLimiter, err := ratelimit.NewRedisLimiter(ctx, rateLimitConfig)
	if err != nil {
		logger.Error("failed to create Redis rate limiter", "error", err)
		os.Exit(1)
	}
	defer redisRateLimiter.Close()

	// Create proxy validator for X-Forwarded-For spoofing prevention
	proxyValidator, err := middleware.NewDefaultProxyValidator([]string{}) // Add trusted proxy IPs if needed
	if err != nil {
		logger.Error("failed to create proxy validator", "error", err)
		os.Exit(1)
	}

	// Apply rate limiting middleware
	r.Use(middleware.RateLimitMiddleware(redisRateLimiter, rateLimitConfig, proxyValidator))

	// Health checks
	h := health.NewHandler("wallet", "dev")
	h.Register(&health.PostgresChecker{DB: db})
	h.RegisterRoutes(r)

	// Wallet handlers
	wallethttp.NewWalletHandler().RegisterRoutes(r)

	addr := fmt.Sprintf(":%d", cfg.Service.Port)
	if err := server.RunHTTP(ctx, addr, r, logger); err != nil {
		logger.Error("server terminated with error", "error", err)
		os.Exit(1)
	}
}
