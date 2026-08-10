package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	curation := service.NewCatalogCurationService(imports, products, parser)
	vitrina := service.NewVitrinaService(imports, products, repository.NewSettingsRepo(db))

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Minute)
	defer cancel()

	if os.Getenv("PUBLISH_ONLY") == "1" {
		pub, err := vitrina.PublishAll(ctx)
		if err != nil {
			log.Fatal("publish:", err)
		}
		synced, _ := products.SyncMetadataFromImports(ctx, service.MinPaidPriceRUB())
		_, _ = vitrina.UpdateAllScores(ctx)
		out := map[string]interface{}{
			"published":       pub.Published,
			"linked_existing": pub.LinkedExisting,
			"products_synced": synced,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
		fmt.Println("done")
		return
	}

	imported, err := parser.RunXboxImportSync(ctx, false)
	if err != nil {
		log.Fatal("import:", err)
	}
	enriched, _ := parser.EnrichXboxOnly(ctx)
	rejected, _ := curation.RejectUnsellableImports(ctx)
	pub, err := vitrina.PublishAll(ctx)
	if err != nil {
		log.Fatal("publish:", err)
	}
	synced, _ := products.SyncMetadataFromImports(ctx, service.MinPaidPriceRUB())
	_, _ = vitrina.UpdateAllScores(ctx)

	out := map[string]interface{}{
		"imported":         imported,
		"enriched":         enriched,
		"imports_rejected": rejected,
		"published":        pub.Published,
		"linked_existing":  pub.LinkedExisting,
		"products_synced":  synced,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
	fmt.Println("done")
}
