package main

import (
	"context"
	"log/slog"
	stdhttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/mdxz2048/podcast-hub/config"
	"github.com/mdxz2048/podcast-hub/internal/auth"
	"github.com/mdxz2048/podcast-hub/internal/connectors"
	"github.com/mdxz2048/podcast-hub/internal/content"
	httpserver "github.com/mdxz2048/podcast-hub/internal/http"
	"github.com/mdxz2048/podcast-hub/internal/intake"
	"github.com/mdxz2048/podcast-hub/internal/jobs"
	"github.com/mdxz2048/podcast-hub/internal/mail"
	"github.com/mdxz2048/podcast-hub/internal/media"
	"github.com/mdxz2048/podcast-hub/internal/publication"
	"github.com/mdxz2048/podcast-hub/internal/ratelimit"
	"github.com/mdxz2048/podcast-hub/internal/security"
	"github.com/mdxz2048/podcast-hub/internal/sources"
	"github.com/mdxz2048/podcast-hub/internal/store/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err.Error())
		os.Exit(1)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect database", "error", err.Error())
		os.Exit(1)
	}
	defer dbPool.Close()
	if err := postgres.ApplyMigrations(ctx, dbPool, "migrations"); err != nil {
		logger.Error("failed to apply migrations", "error", err.Error())
		os.Exit(1)
	}

	redisClient, limiter := buildLimiter(ctx, cfg, logger)
	if redisClient != nil {
		defer redisClient.Close()
	}

	turnstileVerifier, err := buildTurnstileVerifier(cfg)
	if err != nil {
		logger.Error("failed to configure turnstile verifier", "error", err.Error())
		os.Exit(1)
	}
	mailerImpl := mail.NewSMTPMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPFrom)
	store := postgres.NewAuthStore(dbPool)
	connectorStore := postgres.NewConnectorStore(dbPool)
	connectorPackageStore := connectors.NewLocalPackageStore(cfg.ConnectorPackageLocalDir)
	connectorService := connectors.NewService(connectorStore, connectorPackageStore)
	secretCipher, err := sources.NewSecretCipher(cfg.SecretsMasterKey)
	if err != nil {
		logger.Error("failed to configure secret encryption", "error", err.Error())
		os.Exit(1)
	}
	sourceService := sources.NewService(postgres.NewSourceStore(dbPool), connectorStore, secretCipher)
	jobService := jobs.NewService(postgres.NewJobStore(dbPool), sourceService)
	mediaStore := media.NewLocalStore(cfg.StagingStoreDir, cfg.MediaStoreDir)
	contentStore := postgres.NewContentStore(dbPool, mediaStore)
	contentService := content.NewService(contentStore)
	publicationService := publication.NewService(postgres.NewPublicationStore(dbPool), store, mediaStore, cfg.SessionPepper)
	intakeService := intake.NewService(jobService, contentStore, intake.FileArtifactReader{Root: cfg.ImportArtifactStoreDir})
	authService := auth.NewService(store, mailerImpl, turnstileVerifier, limiter, auth.Options{
		SessionPepper:    cfg.SessionPepper,
		AuthCodePepper:   cfg.AuthCodePepper,
		SessionTTL:       cfg.SessionTTL,
		CodeTTL:          10 * time.Minute,
		ResetProofTTL:    10 * time.Minute,
		CodeMaxAttempts:  5,
		ResetMaxAttempts: 5,
	})
	server := httpserver.NewServer(cfg, authService, turnstileVerifier, httpserver.HealthDependencies{
		DB:       dbPool,
		Redis:    redisClient,
		SMTPHost: cfg.SMTPHost,
		SMTPPort: cfg.SMTPPort,
	}, connectorService, sourceService, jobService, contentService, intakeService, publicationService)
	httpSrv := &stdhttp.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           server.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
	}()

	logger.Info("starting api server", "addr", cfg.HTTPAddr, "env", cfg.AppEnv)
	if err := httpSrv.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
		logger.Error("api server stopped with error", "error", err.Error())
		os.Exit(1)
	}
}

func buildTurnstileVerifier(cfg config.Config) (security.TurnstileVerifier, error) {
	if cfg.TurnstileMode == "mock" {
		if cfg.IsProduction() {
			return nil, auth.ErrTurnstileFailed
		}
		return security.MockTurnstileVerifier{}, nil
	}
	return security.CloudflareTurnstileVerifier{
		SecretKey: cfg.TurnstileSecretKey,
	}, nil
}

func buildLimiter(ctx context.Context, cfg config.Config, logger *slog.Logger) (*redis.Client, auth.RateLimiter) {
	if cfg.RedisURL == "" {
		logger.Warn("REDIS_URL is empty, using in-memory limiter")
		return nil, ratelimit.NewMemoryLimiter()
	}
	options, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		options = &redis.Options{Addr: cfg.RedisURL}
	}
	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		if cfg.IsProduction() {
			logger.Error("failed to connect redis in production", "error", err.Error())
			os.Exit(1)
		}
		logger.Warn("failed to connect redis, using in-memory limiter", "error", err.Error())
		return client, ratelimit.NewMemoryLimiter()
	}
	return client, ratelimit.NewRedisLimiter(client)
}
