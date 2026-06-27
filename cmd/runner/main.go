package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mdxz2048/podcast-hub/internal/jobs"
	"github.com/mdxz2048/podcast-hub/internal/runner"
	"github.com/mdxz2048/podcast-hub/internal/sources"
	"github.com/mdxz2048/podcast-hub/internal/store/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	mode := strings.TrimSpace(os.Getenv("RUNNER_MODE"))
	if mode == "" || mode == "disabled" {
		logger.Info("runner disabled", "mode", "disabled")
		return
	}
	if mode != "docker_trusted_admin" {
		logger.Error("unsupported runner mode", "mode", mode)
		os.Exit(1)
	}
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	image := strings.TrimSpace(os.Getenv("RUNNER_PYTHON_BASIC_IMAGE"))
	if databaseURL == "" || image == "" {
		logger.Error("docker trusted admin runner requires DATABASE_URL and RUNNER_PYTHON_BASIC_IMAGE")
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	dbPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		logger.Error("failed to connect database", "error", err.Error())
		os.Exit(1)
	}
	defer dbPool.Close()

	connectorStore := postgres.NewConnectorStore(dbPool)
	secretCipher, err := sources.NewSecretCipher(os.Getenv("SECRETS_MASTER_KEY"))
	if err != nil {
		logger.Error("failed to configure runner secret cipher", "error", err.Error())
		os.Exit(1)
	}
	sourceService := sources.NewService(postgres.NewSourceStore(dbPool), connectorStore, secretCipher)
	jobService := jobs.NewService(postgres.NewJobStore(dbPool), nil)
	executor := runner.DockerTrustedAdminExecutor{Config: runner.DockerTrustedAdminConfig{
		Image:                image,
		ConnectorPackagePath: getEnv("RUNNER_FIXTURE_PACKAGE_DIR", "internal/runner/testdata"),
	}}
	r := runner.New(jobService, executor, runner.Config{WorkspaceRoot: os.Getenv("RUNNER_WORKSPACE_ROOT"), SecretProvider: sourceService})
	if err := r.RunOnce(ctx); err != nil {
		if errors.Is(err, runner.ErrNoQueuedJob) {
			logger.Info("no queued import job")
			return
		}
		logger.Error("runner failed", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("runner finished one import job")
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
