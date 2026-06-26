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
	"github.com/mdxz2048/podcast-hub/internal/store/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	mode := strings.TrimSpace(os.Getenv("RUNNER_MODE"))
	if mode == "" || mode == "disabled" {
		logger.Info("runner disabled", "mode", "disabled")
		return
	}
	if mode != "fixture_subprocess" {
		logger.Error("unsupported runner mode", "mode", mode)
		os.Exit(1)
	}
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	fixtureCommand := strings.TrimSpace(os.Getenv("RUNNER_FIXTURE_COMMAND"))
	if databaseURL == "" || fixtureCommand == "" {
		logger.Error("fixture runner requires DATABASE_URL and RUNNER_FIXTURE_COMMAND")
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

	jobService := jobs.NewService(postgres.NewJobStore(dbPool), nil)
	r := runner.New(jobService, runner.SubprocessExecutor{Command: fixtureCommand}, runner.Config{
		WorkspaceRoot: os.Getenv("RUNNER_WORKSPACE_ROOT"),
	})
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
