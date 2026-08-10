package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"tma-backend/internal/config"
	"tma-backend/internal/logger"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

func main() {
	message := "🎉 <b>У НАС ДР!</b>"
	if len(os.Args) > 1 {
		message = strings.Join(os.Args[1:], " ")
	}

	cfg := config.Load()
	logger.Init(cfg.App.Environment)

	if cfg.Telegram.BotToken == "" {
		slog.Error("BOT_TOKEN is not set")
		os.Exit(1)
	}

	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	if err != nil {
		slog.Error("database connect failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	userRepo := repository.NewUserRepo(db)
	notifSvc := service.NewNotificationService(cfg.Telegram.BotToken)
	notifSvc.SetTMAURL(cfg.App.TMAURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	slog.Info("starting broadcast", slog.String("message", message))

	result, err := notifSvc.BroadcastToAllUsers(ctx, userRepo, message, nil)
	if err != nil {
		slog.Error("broadcast failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("broadcast finished",
		slog.Int("total", result.Total),
		slog.Int("sent", result.Sent),
		slog.Int("failed", result.Failed),
		slog.Int("blocked", result.Blocked),
	)
}
