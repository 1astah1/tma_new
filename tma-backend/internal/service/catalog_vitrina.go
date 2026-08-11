package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"tma-backend/internal/domain"
	"tma-backend/internal/repository"
)

type VitrinaPublishResult struct {
	Published         int64 `json:"published"`
	LinkedExisting    int64 `json:"linked_existing"`
	Rejected          int64 `json:"rejected"`
	ScoresUpdated     int64 `json:"scores_updated"`
	PreordersRequeued int64 `json:"preorders_requeued"`
}

type VitrinaService struct {
	imports  *repository.CatalogImportRepo
	products *repository.ProductRepo
	settings *repository.SettingsRepo
}

func NewVitrinaService(imports *repository.CatalogImportRepo, products *repository.ProductRepo, settings *repository.SettingsRepo) *VitrinaService {
	return &VitrinaService{imports: imports, products: products, settings: settings}
}

type ScoreContext struct {
	Pinned           bool
	OrdersByTitleKey int
}

func (v *VitrinaService) ComputeScore(product domain.Product, ctx ScoreContext) float64 {
	score := 0.0
	if ctx.Pinned {
		score += 1000
	}
	score += CuratedPopularityBoost(product.TitleKey, product.Title)
	score += ShooterAudienceBoost(product.TitleKey, product.Title)
	score += OrderCountBoost(ctx.OrdersByTitleKey)
	if product.Price >= MinPaidPriceRUB() {
		score += 50
	}
	if product.ImageURL != nil && *product.ImageURL != "" {
		score += 20
	}
	if product.ReleaseDate != nil && !product.ReleaseDate.IsZero() {
		days := time.Since(product.ReleaseDate.UTC()).Hours() / 24
		if days >= 0 && days <= 180 {
			score += 35 - (days / 180 * 35)
		} else if days < 0 && days >= -365 {
			score += 30
		}
	}
	switch product.Platform {
	case domain.PlatformPS5, domain.PlatformXbox:
		score += 15
	case domain.PlatformPC:
		score += 10
	}
	return score
}

func (v *VitrinaService) loadPinnedIDs(ctx context.Context) map[uuid.UUID]bool {
	pinned := map[uuid.UUID]bool{}
	setting, err := v.settings.Get(ctx, "popular_product_ids")
	if err == nil {
		raw, _ := setting["value"].(string)
		var ids []string
		if json.Unmarshal([]byte(raw), &ids) == nil {
			for _, id := range ids {
				parsed, err := uuid.Parse(id)
				if err == nil {
					pinned[parsed] = true
				}
			}
		}
	}
	for _, id := range v.loadHomeCategoryPopularIDs(ctx) {
		pinned[id] = true
	}
	return pinned
}

func (v *VitrinaService) loadHomeCategoryPopularIDs(ctx context.Context) []uuid.UUID {
	setting, err := v.settings.Get(ctx, "home_categories")
	if err != nil {
		return nil
	}
	raw, _ := setting["value"].(string)
	var categories []struct {
		ID         string   `json:"id"`
		SectionKey string   `json:"section_key"`
		ProductIDs []string `json:"product_ids"`
	}
	if json.Unmarshal([]byte(raw), &categories) != nil {
		return nil
	}
	for _, cat := range categories {
		isPopular := cat.SectionKey == "popular" || cat.ID == "home-feed-popular" || cat.ID == "home-section-popular"
		if !isPopular || len(cat.ProductIDs) == 0 {
			continue
		}
		out := make([]uuid.UUID, 0, len(cat.ProductIDs))
		for _, id := range cat.ProductIDs {
			parsed, err := uuid.Parse(id)
			if err == nil {
				out = append(out, parsed)
			}
		}
		return out
	}
	return nil
}

func (v *VitrinaService) loadPinnedOrder(ctx context.Context) []uuid.UUID {
	order := make([]uuid.UUID, 0)
	seen := map[uuid.UUID]bool{}
	appendIDs := func(ids []uuid.UUID) {
		for _, id := range ids {
			if seen[id] {
				continue
			}
			seen[id] = true
			order = append(order, id)
		}
	}
	setting, err := v.settings.Get(ctx, "popular_product_ids")
	if err == nil {
		raw, _ := setting["value"].(string)
		var ids []string
		if json.Unmarshal([]byte(raw), &ids) == nil {
			parsed := make([]uuid.UUID, 0, len(ids))
			for _, id := range ids {
				if u, err := uuid.Parse(id); err == nil {
					parsed = append(parsed, u)
				}
			}
			appendIDs(parsed)
		}
	}
	appendIDs(v.loadHomeCategoryPopularIDs(ctx))
	return order
}

func (v *VitrinaService) UpdateAllScores(ctx context.Context) (int64, error) {
	pinned := v.loadPinnedIDs(ctx)
	orderCounts, _ := v.products.OrderCountsByTitleKey(ctx)
	rows, err := v.products.ListActiveGamesForScoring(ctx)
	if err != nil {
		return 0, err
	}
	updated := int64(0)
	for _, row := range rows {
		titleKey := effectiveTitleKey(row.TitleKey, row.Title)
		score := v.ComputeScore(row, ScoreContext{
			Pinned:           pinned[row.ID],
			OrdersByTitleKey: orderCounts[titleKey],
		})
		if err := v.products.UpdateVitrinaScore(ctx, row.ID, score); err != nil {
			continue
		}
		updated++
	}
	return updated, nil
}

func (v *VitrinaService) PublishAll(ctx context.Context) (*VitrinaPublishResult, error) {
	result := &VitrinaPublishResult{}
	batch := 500
	for {
		items, err := v.imports.ListAllPending(ctx, 0, batch)
		if err != nil {
			return result, err
		}
		if len(items) == 0 {
			break
		}
		for _, item := range items {
			if !IsSellableCatalogItem(&item) {
				_ = v.imports.MarkRejected(ctx, item.ID)
				result.Rejected++
				continue
			}
			if item.TitleKey == "" {
				item.TitleKey = NormalizeGameTitle(item.Title)
				item.PlatformFamily = PlatformFamilyFromImport(item.Source, item.Platforms)
			}
			platform := domain.Platform(PickProductPlatform(item.Platforms, ""))
			if existing, err := v.products.FindByTitleKeyPlatform(ctx, item.TitleKey, platform); err == nil && existing != nil {
				if err := v.imports.MarkApproved(ctx, item.ID, existing.ID); err == nil {
					result.LinkedExisting++
				}
				continue
			}
			if approved, err := v.imports.FindApprovedByTitleKey(ctx, item.TitleKey, item.PlatformFamily); err == nil && approved != nil && approved.ProductID != nil {
				if err := v.imports.MarkApproved(ctx, item.ID, *approved.ProductID); err == nil {
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
			if err := v.products.Create(ctx, product); err != nil {
				if errors.Is(err, domain.ErrDuplicate) {
					continue
				}
				slog.Warn("vitrina publish failed", slog.String("title", item.Title), slog.String("error", err.Error()))
				continue
			}
			if err := v.imports.MarkApproved(ctx, item.ID, product.ID); err != nil {
				continue
			}
			result.Published++
		}
	}
	return result, nil
}

func (v *VitrinaService) RepublishPending(ctx context.Context) (*VitrinaPublishResult, error) {
	requeued, err := v.imports.RequeueRejectedPreorders(ctx)
	if err != nil {
		return nil, err
	}
	pub, err := v.PublishAll(ctx)
	if err != nil {
		return pub, err
	}
	pub.PreordersRequeued = requeued
	scores, err := v.UpdateAllScores(ctx)
	if err != nil {
		return pub, err
	}
	pub.ScoresUpdated = scores
	return pub, nil
}

func (v *VitrinaService) ListHybridPopular(ctx context.Context, limit int) ([]domain.Product, error) {
	if limit <= 0 {
		limit = 12
	}
	pinnedOrder := v.loadPinnedOrder(ctx)
	fetchLimit := limit * 3
	if fetchLimit < limit+12 {
		fetchLimit = limit + 12
	}
	products, err := v.products.ListHybridPopular(ctx, pinnedOrder, fetchLimit, MinPaidPriceRUB())
	if err != nil {
		return nil, err
	}
	return filterPopularProducts(products, limit), nil
}

func filterPopularProducts(products []domain.Product, limit int) []domain.Product {
	if len(products) == 0 {
		return products
	}
	out := make([]domain.Product, 0, limit)
	seen := map[string]bool{}
	for _, p := range products {
		if IsPopularFeedExcluded(p.TitleKey, p.Title) {
			continue
		}
		key := p.TitleKey
		if key == "" {
			key = NormalizeGameTitle(p.Title)
		}
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, p)
		if len(out) >= limit {
			break
		}
	}
	return out
}
