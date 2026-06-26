package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mdxz2048/podcast-hub/config"
	"github.com/mdxz2048/podcast-hub/internal/admin"
	"github.com/mdxz2048/podcast-hub/internal/store/postgres"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/admin seed --email admin@example.invalid [--promote]")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "seed":
		if err := runSeed(os.Args[2:]); err != nil {
			slog.Error("admin seed failed", "error", err.Error())
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runSeed(args []string) error {
	fs := flag.NewFlagSet("seed", flag.ContinueOnError)
	email := fs.String("email", "", "admin email")
	promote := fs.Bool("promote", false, "allow promoting existing user to admin in development")
	password := fs.String("password", "", "initial admin password (do not use in production shells)")
	passwordEnv := fs.String("password-env", "ADMIN_SEED_PASSWORD", "read password from environment variable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	seedPassword, err := resolveSeedPassword(cfg, *password, *passwordEnv)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer dbPool.Close()
	store := postgres.NewAuthStore(dbPool)
	service := admin.NewService(store)
	result, user, err := service.SeedAdmin(ctx, admin.SeedInput{
		Email:        strings.TrimSpace(*email),
		Password:     seedPassword,
		AllowPromote: *promote,
		AppEnv:       cfg.AppEnv,
		IPSummary:    "cli",
		UserAgent:    "podcast-hub-admin-cli",
	})
	if err != nil {
		return err
	}
	fmt.Printf("admin seed result=%s email=%s role=%s status=%s\n", result, user.Email, user.Role, user.Status)
	return nil
}

func resolveSeedPassword(cfg config.Config, cliValue string, envKey string) (string, error) {
	if strings.TrimSpace(cliValue) != "" {
		return strings.TrimSpace(cliValue), nil
	}
	if envKey != "" {
		if envValue := strings.TrimSpace(os.Getenv(envKey)); envValue != "" {
			return envValue, nil
		}
	}
	if !strings.EqualFold(cfg.AppEnv, "development") {
		return "", errors.New("password is required outside development; set --password or --password-env")
	}
	if !term.IsTerminal(int(syscall.Stdin)) {
		return "", errors.New("development mode requires interactive terminal or --password-env")
	}
	fmt.Fprint(os.Stderr, "Enter initial admin password: ")
	raw, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}
	password := strings.TrimSpace(string(raw))
	if password == "" {
		return "", admin.ErrPasswordRequired
	}
	return password, nil
}
