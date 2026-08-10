package service

import (
	"context"
	"strings"

	"tma-backend/internal/domain"
)

type ProductFamilyReader interface {
	ListActiveByTitleKey(ctx context.Context, titleKey string) ([]domain.Product, error)
}

func SelectBestFamilyProducts(variants []domain.Product) []domain.Product {
	bestByFamily := map[string]domain.Product{}
	for _, variant := range variants {
		family := PlatformFamilyFromProduct(variant.Platform)
		existing, ok := bestByFamily[family]
		if !ok || ProductPlatformRank(variant.Platform) > ProductPlatformRank(existing.Platform) {
			bestByFamily[family] = variant
		}
	}
	out := make([]domain.Product, 0, len(bestByFamily))
	for _, family := range []string{"ps", "xbox", "pc", "other"} {
		if variant, ok := bestByFamily[family]; ok {
			out = append(out, variant)
		}
	}
	return out
}

func productTitleKey(product domain.Product) string {
	key := strings.TrimSpace(product.TitleKey)
	if key == "" {
		key = NormalizeGameTitle(product.Title)
	}
	return key
}

// EnrichListProductPrices merges sibling regional/edition prices into list cards (home, catalog).
func EnrichListProductPrices(ctx context.Context, reader ProductFamilyReader, products []domain.Product) {
	if reader == nil || len(products) == 0 {
		return
	}

	cache := map[string][]domain.Product{}
	for i := range products {
		if products[i].Type != domain.ProductTypeGame {
			continue
		}
		titleKey := productTitleKey(products[i])
		if titleKey == "" {
			continue
		}

		family, ok := cache[titleKey]
		if !ok {
			siblings, err := reader.ListActiveByTitleKey(ctx, titleKey)
			if err != nil || len(siblings) == 0 {
				cache[titleKey] = nil
				continue
			}
			family = SelectBestFamilyProducts(siblings)
			cache[titleKey] = family
		}
		if len(family) == 0 {
			continue
		}
		if merged := MergeProductPrices(family); merged != nil {
			products[i].Prices = merged
		}
	}
}
