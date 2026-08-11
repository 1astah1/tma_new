package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/lib/pq"
	"tma-backend/internal/domain"
	"tma-backend/internal/repository"
)

type CatalogCurationService struct {
	imports  *repository.CatalogImportRepo
	products *repository.ProductRepo
	parser   *CatalogParserService
}

type CatalogCurationResult struct {
	KeysUpdated         int64 `json:"keys_updated"`
	ImportsRejected     int64 `json:"imports_rejected"`
	ProductsDeleted     int64 `json:"products_deleted"`
	ProductsHidden      int64 `json:"products_hidden"`
	ProductsSynced      int64 `json:"products_synced"`
	DescriptionsUpdated int   `json:"descriptions_updated"`
	Enriched            int   `json:"enriched"`
	Published           int64 `json:"published"`
	LinkedExisting      int64 `json:"linked_existing"`
	Activated           int64 `json:"activated"`
	Reclassified        bool  `json:"reclassified"`
}

func NewCatalogCurationService(imports *repository.CatalogImportRepo, products *repository.ProductRepo, parser *CatalogParserService) *CatalogCurationService {
	return &CatalogCurationService{imports: imports, products: products, parser: parser}
}

func (s *CatalogCurationService) RefreshTitleKeys(ctx context.Context) (int64, error) {
	updated := int64(0)
	rows, err := s.imports.ListImportKeyRows(ctx)
	if err != nil {
		return 0, err
	}
	for _, row := range rows {
		key := CatalogTitleKey(row.Title)
		family := PlatformFamilyFromImport(row.Source, row.Platforms)
		if key == "" {
			continue
		}
		if err := s.imports.UpdateImportKeys(ctx, row.ID, key, family); err != nil {
			return updated, err
		}
		updated++
	}

	gameRows, err := s.products.ListGameIDs(ctx)
	if err != nil {
		return updated, err
	}
	for _, row := range gameRows {
		key := CatalogTitleKey(row.Title)
		if key == "" {
			continue
		}
		if err := s.products.UpdateTitleKey(ctx, row.ID, key); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}

func (s *CatalogCurationService) Deduplicate(ctx context.Context) (*CatalogCurationResult, error) {
	result := &CatalogCurationResult{}
	if err := s.imports.BackfillImportMetadata(ctx); err != nil {
		return result, err
	}
	result.Reclassified = true

	keys, err := s.RefreshTitleKeys(ctx)
	if err != nil {
		return result, err
	}
	result.KeysUpdated = keys

	rejected, err := s.imports.RemoveDuplicateImports(ctx)
	if err != nil {
		return result, err
	}
	result.ImportsRejected = rejected

	deleted, hidden, err := s.products.RemoveDuplicateProducts(ctx)
	if err != nil {
		return result, err
	}
	result.ProductsDeleted = deleted
	result.ProductsHidden = hidden
	return result, nil
}

func (s *CatalogCurationService) AutoPublishFresh(ctx context.Context, limit int) (*CatalogCurationResult, error) {
	return s.publishSections(ctx, []string{"preorder", "new"}, limit, 300)
}

// PublishWantedImports публикует всё, что импортировано адресно по списку
// желаемых игр, включая раздел «каталог»: AutoPublishFresh намеренно берёт
// только предзаказы и новинки, а список хотим видеть целиком.
func (s *CatalogCurationService) PublishWantedImports(ctx context.Context, limit int) (*CatalogCurationResult, error) {
	result, err := s.publishSections(ctx, []string{"preorder", "new", "game"}, limit, 5000)
	if err != nil {
		return result, err
	}

	// Публикация создаёт только новые карточки. Существующие надо ещё и
	// обновить, иначе повторный прогон не исправит ни цену, ни раздел.
	synced, syncErr := s.products.SyncMetadataFromImports(ctx, MinPaidPriceRUB())
	if syncErr != nil {
		return result, syncErr
	}
	result.ProductsSynced += synced

	if cards, cardErr := s.products.SyncCardFromImports(ctx); cardErr != nil {
		return result, cardErr
	} else {
		result.ProductsSynced += cards
	}
	return result, nil
}

func (s *CatalogCurationService) publishSections(ctx context.Context, sections []string, limit, fallbackLimit int) (*CatalogCurationResult, error) {
	if limit <= 0 {
		limit = fallbackLimit
	}
	result, err := s.Deduplicate(ctx)
	if err != nil {
		return result, err
	}

	items, err := s.imports.ListPendingBySections(ctx, sections, limit)
	if err != nil {
		return result, err
	}

	for _, item := range items {
		if !IsSellableCatalogItem(&item) {
			_ = s.imports.MarkRejected(ctx, item.ID)
			result.ImportsRejected++
			continue
		}
		if item.TitleKey == "" {
			item.TitleKey = NormalizeGameTitle(item.Title)
			item.PlatformFamily = PlatformFamilyFromImport(item.Source, item.Platforms)
		}
		platform := domain.Platform(PickProductPlatform(item.Platforms, ""))

		if existing, err := s.products.FindByTitleKeyPlatform(ctx, item.TitleKey, platform); err == nil && existing != nil {
			if err := s.imports.MarkApproved(ctx, item.ID, existing.ID); err == nil {
				result.LinkedExisting++
			}
			continue
		}

		if approved, err := s.imports.FindApprovedByTitleKey(ctx, item.TitleKey, item.PlatformFamily); err == nil && approved != nil && approved.ProductID != nil {
			if err := s.imports.MarkApproved(ctx, item.ID, *approved.ProductID); err == nil {
				result.LinkedExisting++
			}
			continue
		}

		price := EffectiveCatalogPriceRUB(&item)
		if price <= 0 {
			// Карточка с нулём вместо цены обманывает покупателя — ждём,
			// пока в сторе появится цена, даже если это предзаказ.
			continue
		}
		if price < MinPaidPriceRUB() && !IsPreorderCatalogItem(&item) {
			continue
		}

		product := &domain.Product{
			Title:           item.Title,
			TitleKey:        item.TitleKey,
			Description:     item.Description,
			Platform:        platform,
			Type:            domain.ProductTypeGame,
			GameSection:     item.GameSection,
			ReleaseDate:     item.ReleaseDate,
			Price:           price,
			DiscountPercent: 0,
			Variants:        json.RawMessage("[]"),
			ImageURL:        item.ImageURL,
			DeliveryMethods: pq.StringArray{string(domain.DeliveryMethodActivation)},
			Prices:          item.Prices,
			Status:          domain.ProductStatusActive,
		}
		if err := s.products.Create(ctx, product); err != nil {
			if errors.Is(err, domain.ErrDuplicate) {
				continue
			}
			slog.Warn("auto publish failed", slog.String("title", item.Title), slog.String("error", err.Error()))
			continue
		}
		if err := s.imports.MarkApproved(ctx, item.ID, product.ID); err != nil {
			continue
		}
		result.Published++
	}

	activated, err := s.products.ActivateFreshGames(ctx, MinPaidPriceRUB())
	if err != nil {
		return result, err
	}
	result.Activated = activated

	dedupe, err := s.Deduplicate(ctx)
	if err == nil && dedupe != nil {
		result.ImportsRejected += dedupe.ImportsRejected
		result.ProductsDeleted += dedupe.ProductsDeleted
		result.ProductsHidden += dedupe.ProductsHidden
	}

	hidden, err := s.products.DeactivateUnsellableGames(ctx, MinPaidPriceRUB())
	if err == nil {
		result.ProductsHidden += hidden
	}
	return result, nil
}

func (s *CatalogCurationService) RejectUnsellableImports(ctx context.Context) (int64, error) {
	items, err := s.imports.ListPendingForQuality(ctx, 8000)
	if err != nil {
		return 0, err
	}
	rejected := int64(0)
	for _, item := range items {
		if IsSellableCatalogItem(&item) {
			continue
		}
		if err := s.imports.MarkRejected(ctx, item.ID); err != nil {
			continue
		}
		rejected++
	}
	return rejected, nil
}

func (s *CatalogCurationService) RejectAllUnsellableImports(ctx context.Context) (int64, error) {
	total := int64(0)
	for {
		n, err := s.RejectUnsellableImports(ctx)
		if err != nil {
			return total, err
		}
		if n == 0 {
			break
		}
		total += n
	}
	return total, nil
}

func (s *CatalogCurationService) RefreshCatalog(ctx context.Context) (*CatalogCurationResult, error) {
	result := &CatalogCurationResult{}

	if s.parser != nil {
		enriched, err := s.parser.EnrichAllImports(ctx)
		if err != nil {
			slog.Warn("catalog enrich failed", slog.String("error", err.Error()))
		}
		result.Enriched = enriched

		descriptions, err := s.RefreshRussianDescriptions(ctx, 300)
		if err != nil {
			slog.Warn("russian descriptions refresh failed", slog.String("error", err.Error()))
		}
		result.DescriptionsUpdated = descriptions
	}

	if err := s.imports.BackfillImportMetadata(ctx); err != nil {
		return result, err
	}
	result.Reclassified = true

	if _, err := s.imports.ClearInvalidReleaseDates(ctx); err != nil {
		return result, err
	}
	if _, err := s.products.ClearInvalidReleaseDates(ctx); err != nil {
		return result, err
	}

	rejected, err := s.RejectUnsellableImports(ctx)
	if err != nil {
		return result, err
	}
	result.ImportsRejected += rejected

	synced, err := s.products.SyncMetadataFromImports(ctx, MinPaidPriceRUB())
	if err != nil {
		return result, err
	}
	result.ProductsSynced = synced

	hidden, err := s.products.DeactivateUnsellableGames(ctx, MinPaidPriceRUB())
	if err != nil {
		return result, err
	}
	result.ProductsHidden += hidden

	dedupe, err := s.Deduplicate(ctx)
	if err != nil {
		return result, err
	}
	result.KeysUpdated = dedupe.KeysUpdated
	result.ImportsRejected += dedupe.ImportsRejected
	result.ProductsDeleted += dedupe.ProductsDeleted
	result.ProductsHidden += dedupe.ProductsHidden

	// Publish all sellable pending imports (not only preorder/new).
	for {
		items, err := s.imports.ListAllPending(ctx, 0, 500)
		if err != nil {
			return result, err
		}
		if len(items) == 0 {
			break
		}
		batchPublished := int64(0)
		for _, item := range items {
			if !IsSellableCatalogItem(&item) {
				_ = s.imports.MarkRejected(ctx, item.ID)
				result.ImportsRejected++
				continue
			}
			if item.TitleKey == "" {
				item.TitleKey = NormalizeGameTitle(item.Title)
				item.PlatformFamily = PlatformFamilyFromImport(item.Source, item.Platforms)
			}
			platform := domain.Platform(PickProductPlatform(item.Platforms, ""))
			if existing, err := s.products.FindByTitleKeyPlatform(ctx, item.TitleKey, platform); err == nil && existing != nil {
				if err := s.imports.MarkApproved(ctx, item.ID, existing.ID); err == nil {
					result.LinkedExisting++
				}
				continue
			}
			if approved, err := s.imports.FindApprovedByTitleKey(ctx, item.TitleKey, item.PlatformFamily); err == nil && approved != nil && approved.ProductID != nil {
				if err := s.imports.MarkApproved(ctx, item.ID, *approved.ProductID); err == nil {
					result.LinkedExisting++
				}
				continue
			}
			price := EffectiveCatalogPriceRUB(&item)
			if price < MinPaidPriceRUB() && !IsPreorderCatalogItem(&item) {
				continue
			}
			product := &domain.Product{
				Title:           item.Title,
				TitleKey:        item.TitleKey,
				Description:     item.Description,
				Platform:        platform,
				Type:            domain.ProductTypeGame,
				GameSection:     item.GameSection,
				ReleaseDate:     item.ReleaseDate,
				Price:           price,
				DiscountPercent: 0,
				Variants:        json.RawMessage("[]"),
				ImageURL:        item.ImageURL,
				DeliveryMethods: pq.StringArray{string(domain.DeliveryMethodActivation)},
				Prices:          item.Prices,
				Status:          domain.ProductStatusActive,
			}
			if err := s.products.Create(ctx, product); err != nil {
				if errors.Is(err, domain.ErrDuplicate) {
					continue
				}
				slog.Warn("catalog refresh publish failed", slog.String("title", item.Title), slog.String("error", err.Error()))
				continue
			}
			if err := s.imports.MarkApproved(ctx, item.ID, product.ID); err != nil {
				continue
			}
			batchPublished++
			result.Published++
		}
		if batchPublished == 0 {
			break
		}
	}

	activated, err := s.products.ActivateFreshGames(ctx, MinPaidPriceRUB())
	if err != nil {
		return result, err
	}
	result.Activated = activated

	return result, nil
}

func (s *CatalogCurationService) SyncCatalog(ctx context.Context, fullImport bool) (*CatalogCurationResult, error) {
	if s.parser != nil {
		if err := s.parser.RunAsync(ctx, fullImport); err != nil {
			slog.Warn("catalog parser already running", slog.String("error", err.Error()))
		}
		for {
			if !s.parser.Status().Running {
				break
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(2 * time.Second):
			}
		}
	}
	return s.RefreshCatalog(ctx)
}

func (s *CatalogCurationService) RefreshRussianDescriptions(ctx context.Context, batchSize int) (int, error) {
	if s.parser == nil {
		return 0, nil
	}
	updated := 0
	for offset := 0; ; offset += batchSize {
		rows, err := s.imports.ListLinkedForDescriptionRefresh(ctx, batchSize, offset)
		if err != nil {
			return updated, err
		}
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			desc := s.parser.RussianDescription(ctx, row.Source, row.ExternalID, row.Description)
			if desc == "" || desc == strings.TrimSpace(row.Description) {
				continue
			}
			if err := s.products.UpdateDescription(ctx, row.ProductID, desc); err != nil {
				continue
			}
			_ = s.imports.UpdateDescription(ctx, row.ImportID, desc)
			updated++
		}
		if len(rows) < batchSize {
			break
		}
	}
	return updated, nil
}
