package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

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

	products := repository.NewProductRepo(db)
	ctx := context.Background()

	synced, err := products.SyncMetadataFromImports(ctx, service.MinPaidPriceRUB())
	if err != nil {
		log.Fatal(err)
	}
	hidden, _ := products.DeactivateUnsellableGames(ctx, service.MinPaidPriceRUB())

	out := map[string]interface{}{
		"products_synced": synced,
		"products_hidden": hidden,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
}
