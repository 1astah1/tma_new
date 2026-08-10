package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"tma-backend/internal/repository"
)

var ErrCatalogRebuildRunning = errors.New("catalog rebuild is already running")

type CatalogRebuildStatus struct {
	Running    bool                  `json:"running"`
	Stage      string                `json:"stage"`
	StartedAt  *time.Time            `json:"started_at,omitempty"`
	FinishedAt *time.Time            `json:"finished_at,omitempty"`
	Result     *CatalogRebuildResult `json:"result,omitempty"`
	Error      string                `json:"error,omitempty"`
}

type CatalogRebuildResult struct {
	Reset           *repository.CatalogResetResult `json:"reset,omitempty"`
	Imported        int                            `json:"imported"`
	Enriched        int                            `json:"enriched"`
	ImportsRejected int64                          `json:"imports_rejected"`
	Published       int64                          `json:"published"`
	LinkedExisting  int64                          `json:"linked_existing"`
	ProductsSynced  int64                          `json:"products_synced"`
	ProductsHidden  int64                          `json:"products_hidden"`
	ScoresUpdated   int64                          `json:"scores_updated"`
	KeysUpdated     int64                          `json:"keys_updated"`
	ProductsDeleted int64                          `json:"products_deleted"`
}

type CatalogRebuildService struct {
	imports  *repository.CatalogImportRepo
	products *repository.ProductRepo
	settings *repository.SettingsRepo
	parser   *CatalogParserService
	curation *CatalogCurationService
	vitrina  *VitrinaService

	mu     sync.Mutex
	status CatalogRebuildStatus
}

func NewCatalogRebuildService(
	imports *repository.CatalogImportRepo,
	products *repository.ProductRepo,
	settings *repository.SettingsRepo,
	parser *CatalogParserService,
	curation *CatalogCurationService,
	vitrina *VitrinaService,
) *CatalogRebuildService {
	return &CatalogRebuildService{
		imports:  imports,
		products: products,
		settings: settings,
		parser:   parser,
		curation: curation,
		vitrina:  vitrina,
	}
}

func (s *CatalogRebuildService) Status() CatalogRebuildStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

func (s *CatalogRebuildService) setStage(stage string) {
	s.mu.Lock()
	s.status.Stage = stage
	s.mu.Unlock()
}

func (s *CatalogRebuildService) RunAsync(ctx context.Context) error {
	s.mu.Lock()
	if s.status.Running {
		s.mu.Unlock()
		return ErrCatalogRebuildRunning
	}
	now := time.Now()
	s.status = CatalogRebuildStatus{
		Running:   true,
		Stage:     "Сброс каталога",
		StartedAt: &now,
	}
	s.mu.Unlock()

	go func() {
		runCtx := context.WithoutCancel(ctx)
		result, err := s.run(runCtx)
		finished := time.Now()
		s.mu.Lock()
		s.status.Running = false
		s.status.FinishedAt = &finished
		if err != nil {
			s.status.Error = err.Error()
			s.status.Stage = "Ошибка"
		} else {
			s.status.Result = result
			s.status.Stage = "Готово"
		}
		s.mu.Unlock()
	}()
	return nil
}

func (s *CatalogRebuildService) RunSync(ctx context.Context) (*CatalogRebuildResult, error) {
	return s.run(ctx)
}

func (s *CatalogRebuildService) run(ctx context.Context) (*CatalogRebuildResult, error) {
	result := &CatalogRebuildResult{}

	s.parser.Stop()
	s.setStage("Сброс каталога")
	reset, err := s.imports.ResetGameCatalog(ctx)
	if err != nil {
		return result, err
	}
	result.Reset = reset
	_ = s.settings.Upsert(ctx, "popular_product_ids", []string{})

	s.setStage("Парсинг PS Store и Xbox Store TR (цены в ₽)")
	imported, err := s.parser.RunImportSync(ctx, true)
	if err != nil {
		return result, err
	}
	result.Imported = imported

	s.setStage("Обогащение цен и дат")
	enriched, err := s.parser.EnrichAllImports(ctx)
	if err != nil {
		slog.Warn("enrich all failed", slog.String("error", err.Error()))
	}
	result.Enriched = enriched

	s.setStage("Классификация и валидация")
	if err := s.imports.BackfillImportMetadata(ctx); err != nil {
		return result, err
	}
	if _, err := s.imports.ClearInvalidReleaseDates(ctx); err != nil {
		return result, err
	}
	rejected, err := s.curation.RejectAllUnsellableImports(ctx)
	if err != nil {
		return result, err
	}
	result.ImportsRejected = rejected

	s.setStage("Удаление дублей")
	dedupe, err := s.curation.Deduplicate(ctx)
	if err != nil {
		return result, err
	}
	result.KeysUpdated = dedupe.KeysUpdated
	result.ImportsRejected += dedupe.ImportsRejected
	result.ProductsDeleted = dedupe.ProductsDeleted
	result.ProductsHidden = dedupe.ProductsHidden

	s.setStage("Публикация витрины")
	pub, err := s.vitrina.PublishAll(ctx)
	if err != nil {
		return result, err
	}
	result.Published = pub.Published
	result.LinkedExisting = pub.LinkedExisting
	result.ImportsRejected += pub.Rejected

	s.setStage("Рейтинг популярного")
	scores, err := s.vitrina.UpdateAllScores(ctx)
	if err != nil {
		return result, err
	}
	result.ScoresUpdated = scores

	synced, err := s.products.SyncMetadataFromImports(ctx, MinPaidPriceRUB())
	if err != nil {
		return result, err
	}
	result.ProductsSynced = synced

	hidden, err := s.products.DeactivateUnsellableGames(ctx, MinPaidPriceRUB())
	if err != nil {
		return result, err
	}
	result.ProductsHidden = hidden

	slog.Info("catalog rebuild finished",
		slog.Int("imported", result.Imported),
		slog.Int("enriched", result.Enriched),
		slog.Int64("published", result.Published),
		slog.Int64("rejected", result.ImportsRejected),
	)
	return result, nil
}
