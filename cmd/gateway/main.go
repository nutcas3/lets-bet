package main

import (
	"context"
	"fmt"
	"os"

	"github.com/betting-platform/internal/infrastructure/config"
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
	cfg, err := config.LoadFromFile(".env")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	logger := logging.Setup(cfg.Logging.Level, cfg.Logging.Format)
	logger.Info("starting gateway", "env", cfg.Service.Environment)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize OpenTelemetry tracing
	tracerCfg := tracing.DefaultConfig("gateway")
	cleanup, err := tracing.InitTracer(ctx, tracerCfg)
	if err != nil {
		logger.Error("failed to initialize tracer", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	// Initialize MaxMind GeoIP provider (optional)
	geoProvider, err := middleware.NewMaxMindProvider(middleware.DefaultDBPath())
	if err != nil {
		logger.Warn("failed to load maxmind database", "error", err)
	}

	// Initialize Redis rate limiter
	rateLimiterConfig := ratelimit.DefaultConfig()
	rateLimiterConfig.RedisAddr = cfg.Redis.Addr()
	rateLimiterConfig.RedisPassword = cfg.Redis.Password
	rateLimiterConfig.RedisDB = cfg.Redis.DB

	rateLimiter, err := ratelimit.NewRedisLimiter(ctx, rateLimiterConfig)
	if err != nil {
		logger.Error("failed to initialize Redis rate limiter", "error", err)
		os.Exit(1)
	}
	defer rateLimiter.Close()

	r := mux.NewRouter()

	rec := metrics.New("gateway")
	rec.RegisterRoutes(r)

	// Apply middleware stack
	r.Use(tracing.HTTPMiddleware("gateway"))
	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(middleware.CORS(cfg.Security))
	r.Use(rec.Middleware)
	r.Use(middleware.Geolocation(middleware.GeoConfig{
		Provider: geoProvider,
		Allowed:  cfg.Tenant.AllowedCountries,
	}))

	// Apply Redis rate limiter middleware
	r.Use(rateLimiter.HTTPMiddleware())

	health.NewHandler("gateway", "dev").RegisterRoutes(r)

	// Gateway handlers will be added when implemented

	addr := fmt.Sprintf(":%d", cfg.Service.Port)
	if err := server.RunHTTP(ctx, addr, r, logger); err != nil {
		logger.Error("server terminated with error", "error", err)
		os.Exit(1)
	}
}
