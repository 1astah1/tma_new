package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"tma-backend/internal/config"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

func main() {
	cfg := config.Load()
	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	imports := repository.NewCatalogImportRepo(db)
	products := repository.NewProductRepo(db)
	parser := service.NewCatalogParserService(imports, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Hour)
	defer cancel()

	enriched, err := parser.EnrichAllImports(ctx)
	if err != nil {
		log.Fatal("enrich:", err)
	}
	synced, err := products.SyncMetadataFromImports(ctx, service.MinPaidPriceRUB())
	if err != nil {
		log.Fatal("sync:", err)
	}
	hidden, _ := products.DeactivateUnsellableGames(ctx, service.MinPaidPriceRUB())

	out := map[string]interface{}{
		"enriched":         enriched,
		"products_synced":  synced,
		"products_hidden":  hidden,
		"try_rub_rate":     service.TRYToRUBRate(),
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}
