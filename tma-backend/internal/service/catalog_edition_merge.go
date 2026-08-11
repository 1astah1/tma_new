package service

import (
	"encoding/json"
	"strings"

	"tma-backend/internal/domain"
)

type editionOption struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	DiscountLabel string  `json:"discount_label,omitempty"`
}

type editionCatalogMap map[string][]editionOption

type parsedProductPrices struct {
	TR             *float64          `json:"tr"`
	UA             *float64          `json:"ua"`
	Xbox           *float64          `json:"xbox"`
	XboxTR         *float64          `json:"xbox_tr"`
	US             *float64          `json:"us"`
	EditionCatalog editionCatalogMap `json:"edition_catalog"`
}

func parseProductPricesJSON(raw json.RawMessage) parsedProductPrices {
	var prices parsedProductPrices
	if len(raw) == 0 {
		return prices
	}
	_ = json.Unmarshal(raw, &prices)
	return prices
}

func standardEdition(price float64) editionOption {
	return editionOption{
		ID:    "standard",
		Name:  "Standard Edition",
		Price: price,
	}
}

func mergeEditionLists(dst editionCatalogMap, key string, items []editionOption) {
	if len(items) == 0 {
		return
	}
	if len(dst[key]) == 0 {
		dst[key] = items
		return
	}
	seen := map[string]editionOption{}
	for _, item := range dst[key] {
		seen[item.ID] = item
	}
	for _, item := range items {
		if existing, ok := seen[item.ID]; !ok || item.Price < existing.Price {
			seen[item.ID] = item
		}
	}
	out := make([]editionOption, 0, len(seen))
	for _, item := range seen {
		out = append(out, item)
	}
	dst[key] = out
}

func appendSimplePrice(catalog editionCatalogMap, key string, price float64) {
	if price < MinPaidPriceRUB() {
		return
	}
	mergeEditionLists(catalog, key, []editionOption{standardEdition(price)})
}

func appendProductListPrice(catalog editionCatalogMap, product domain.Product) {
	if product.Price < MinPaidPriceRUB() {
		return
	}
	family := PlatformFamilyFromProduct(product.Platform)
	switch family {
	case "ps":
		if len(catalog["ps_tr"]) == 0 && len(catalog["ps_ua"]) == 0 {
			appendSimplePrice(catalog, "ps_tr", product.Price)
		}
	case "xbox", "pc":
		if len(catalog["xbox"]) == 0 {
			appendSimplePrice(catalog, "xbox", product.Price)
		}
	}
}

func catalogFromProductPrices(prices parsedProductPrices, platform domain.Platform) editionCatalogMap {
	catalog := editionCatalogMap{}
	if len(prices.EditionCatalog) > 0 {
		for key, items := range prices.EditionCatalog {
			mergeEditionLists(catalog, key, items)
		}
	}

	family := PlatformFamilyFromProduct(platform)
	switch family {
	case "ps":
		if prices.TR != nil {
			appendSimplePrice(catalog, "ps_tr", *prices.TR)
		}
		if prices.UA != nil {
			appendSimplePrice(catalog, "ps_ua", *prices.UA)
		}
	case "xbox", "pc":
		price := 0.0
		if prices.Xbox != nil {
			price = *prices.Xbox
		} else if prices.US != nil {
			price = *prices.US
		}
		if price > 0 {
			appendSimplePrice(catalog, "xbox", price)
		}
		if prices.XboxTR != nil && *prices.XboxTR > 0 {
			appendSimplePrice(catalog, "xbox_tr", *prices.XboxTR)
		}
	}
	return catalog
}

func knownEditionCatalog(titleKey string) editionCatalogMap {
	switch strings.ToLower(strings.TrimSpace(titleKey)) {
	case "call of duty: modern warfare 4":
		return editionCatalogMap{
			"ps_tr": {
				{ID: "standard", Name: "Standard Edition", Price: 6300},
				{ID: "vault", Name: "Vault Edition", Price: 7800, DiscountLabel: "−10%"},
			},
			"ps_ua": {
				{ID: "standard", Name: "Standard Edition", Price: 7300},
				{ID: "vault", Name: "Vault Edition", Price: 8900, DiscountLabel: "−10%"},
			},
			"xbox": {
				{ID: "standard", Name: "Standard Edition", Price: 5000},
				{ID: "vault", Name: "Vault Edition", Price: 6700, DiscountLabel: "−10%"},
			},
		}
	}
	return nil
}

// MergeProductPrices builds a unified edition_catalog across PS/Xbox siblings.
func MergeProductPrices(products []domain.Product) json.RawMessage {
	if len(products) == 0 {
		return nil
	}

	merged := editionCatalogMap{}
	var tr, ua, xbox, xboxTR, us *float64
	titleKey := ""

	for _, product := range products {
		if titleKey == "" {
			titleKey = strings.TrimSpace(product.TitleKey)
			if titleKey == "" {
				titleKey = NormalizeGameTitle(product.Title)
			}
		}
		prices := parseProductPricesJSON(product.Prices)
		part := catalogFromProductPrices(prices, product.Platform)
		for key, items := range part {
			mergeEditionLists(merged, key, items)
		}
		appendProductListPrice(merged, product)
		if prices.TR != nil {
			v := *prices.TR
			tr = &v
		}
		if prices.UA != nil {
			v := *prices.UA
			ua = &v
		}
		if prices.XboxTR != nil {
			v := *prices.XboxTR
			xboxTR = &v
		}
		if prices.Xbox != nil {
			v := *prices.Xbox
			xbox = &v
		} else if prices.US != nil {
			v := *prices.US
			us = &v
		}
	}

	if known := knownEditionCatalog(titleKey); len(known) > 0 {
		for key, items := range known {
			mergeEditionLists(merged, key, items)
		}
	}

	if len(merged) == 0 {
		return nil
	}

	out := map[string]interface{}{
		"edition_catalog": merged,
	}
	if tr != nil {
		out["tr"] = *tr
	}
	if ua != nil {
		out["ua"] = *ua
	}
	if xbox != nil {
		out["xbox"] = *xbox
	}
	if xboxTR != nil {
		out["xbox_tr"] = *xboxTR
	}
	if us != nil {
		out["us"] = *us
	}

	data, err := json.Marshal(out)
	if err != nil {
		return nil
	}
	return data
}

func HasMultiPlatformCatalog(prices json.RawMessage) bool {
	parsed := parseProductPricesJSON(prices)
	if len(parsed.EditionCatalog) == 0 {
		return false
	}
	families := map[string]bool{}
	for key := range parsed.EditionCatalog {
		switch key {
		case "ps_tr", "ps_ua":
			families["ps"] = true
		case "xbox":
			families["xbox"] = true
		case "pc":
			families["pc"] = true
		}
	}
	return len(families) > 1
}
