package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"tma-backend/internal/config"
	"tma-backend/internal/logger"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.App.Environment)

	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	if err != nil {
		slog.Error("database connect failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewCatalogImportRepo(db)
	if err := repo.EnsureSchema(context.Background()); err != nil {
		slog.Error("ensure schema failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	parser := service.NewCatalogParserService(repo, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	slog.Info("starting catalog import (PS + Xbox)")
	if err := parser.RunAsync(ctx, false); err != nil {
		slog.Error("parser start failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	for {
		status := parser.Status()
		if !status.Running {
			break
		}
		slog.Info("parser running",
			slog.String("source", status.CurrentSource),
			slog.String("stage", status.CurrentStage),
			slog.Int("imported", status.Imported),
			slog.Float64("percent", status.Percent),
		)
		time.Sleep(5 * time.Second)
	}

	status := parser.Status()
	slog.Info("parser finished", slog.Int("imported", status.Imported), slog.Any("errors", status.Errors))

	if err := repo.BackfillImportMetadata(ctx); err != nil {
		slog.Warn("sql backfill failed", slog.String("error", err.Error()))
	}

	updated, err := parser.EnrichAllImports(ctx)
	if err != nil {
		slog.Error("api enrichment failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	withPublisher, _ := repo.CountWithPublisher(ctx)
	slog.Info("done", slog.Int("enriched", updated), slog.Int("with_publisher", withPublisher))
	slog.Info("for full rebuild use POST /api/v1/admin/catalog-imports/rebuild")
}
