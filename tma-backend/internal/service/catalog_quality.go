package service

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"tma-backend/internal/domain"
)

// nonGameTitlePatterns — только явный не-игровой контент из стора (демо, темы, подписки).
// Не фильтруем по жанру/названию: платная игра остаётся, даже если в названии «casino», «slot» и т.д.
var nonGameTitlePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(demo|trial|beta|preview|playable teaser|game hub)\b`),
	regexp.MustCompile(`(?i)\b(dynamic theme|avatar|wallpaper|soundtrack|ost|music pack)\b`),
	regexp.MustCompile(`(?i)\b(ps plus|playstation plus|xbox game pass|subscription only)\b`),
}

func parseDisplayPriceRUB(display string) float64 {
	display = strings.ToLower(strings.TrimSpace(display))
	if display == "" || strings.Contains(display, "бесплат") || strings.Contains(display, "free") {
		return 0
	}
	display = strings.ReplaceAll(display, "руб.", "")
	display = strings.ReplaceAll(display, "rub", "")
	display = strings.ReplaceAll(display, "₽", "")
	display = strings.ReplaceAll(display, "\u00a0", "")
	display = strings.ReplaceAll(display, " ", "")
	display = strings.ReplaceAll(display, ",", ".")
	match := regexp.MustCompile(`[\d.]+`).FindString(display)
	if match == "" {
		return 0
	}
	value, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0
	}
	return value
}

func IsFreePrice(price *float64, displayPrice string) bool {
	dp := strings.ToLower(strings.TrimSpace(displayPrice))
	if strings.Contains(dp, "free") || strings.Contains(dp, "бесплат") {
		return true
	}
	if price == nil {
		return parseDisplayPriceRUB(displayPrice) <= 0
	}
	if *price <= 0 || *price <= 1.01 {
		return true
	}
	return false
}

func IsNonGameStoreItem(title string) bool {
	title = strings.TrimSpace(title)
	if title == "" {
		return true
	}
	for _, pattern := range nonGameTitlePatterns {
		if pattern.MatchString(title) {
			return true
		}
	}
	return false
}

func EffectiveCatalogPriceRUB(item *domain.CatalogImport) float64 {
	if item == nil {
		return 0
	}
	if item.OriginalPriceRUB != nil && *item.OriginalPriceRUB >= MinPaidPriceRUB() {
		return *item.OriginalPriceRUB
	}
	if parsed := parseDisplayPriceRUB(extractDisplayPriceFromRaw(item)); parsed >= MinPaidPriceRUB() {
		return parsed
	}
	if parsedTRY := parseDisplayPriceTRY(extractDisplayPriceFromRaw(item)); parsedTRY > 0 {
		if rub := ConvertTRYToRUB(parsedTRY); rub >= MinPaidPriceRUB() {
			return rub
		}
	}
	if item.OriginalPriceRUB != nil && *item.OriginalPriceRUB > 0 {
		return *item.OriginalPriceRUB
	}
	return 0
}

func hasValidFutureRelease(item *domain.CatalogImport) bool {
	if item == nil || item.ReleaseDate == nil {
		return false
	}
	releaseAt := item.ReleaseDate.UTC()
	return releaseAt.Year() >= 1990 && releaseAt.Year() <= 2035 && releaseAt.After(time.Now().UTC())
}

func IsPreorderCatalogItem(item *domain.CatalogImport) bool {
	if item == nil {
		return false
	}
	if item.GameSection == "preorder" {
		return true
	}
	return hasValidFutureRelease(item)
}

func isExplicitlyFreeListing(display string, price *float64) bool {
	dp := strings.ToLower(strings.TrimSpace(display))
	if strings.Contains(dp, "free") || strings.Contains(dp, "бесплат") {
		return true
	}
	if price != nil && *price > 0 && *price <= 1.01 {
		return true
	}
	return false
}

func IsSellableCatalogItem(item *domain.CatalogImport) bool {
	if item == nil {
		return false
	}
	if IsNonGameStoreItem(item.Title) {
		return false
	}
	display := extractDisplayPriceFromRaw(item)
	if IsPreorderCatalogItem(item) && hasValidFutureRelease(item) {
		if isExplicitlyFreeListing(display, item.OriginalPriceRUB) {
			return false
		}
		return true
	}
	price := EffectiveCatalogPriceRUB(item)
	if price < MinPaidPriceRUB() {
		return false
	}
	ptr := &price
	return !IsFreePrice(ptr, display)
}

func extractDisplayPriceFromRaw(item *domain.CatalogImport) string {
	if item == nil || len(item.Raw) == 0 {
		return ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(item.Raw, &payload); err != nil {
		return ""
	}
	if sku, ok := payload["default_sku"].(map[string]interface{}); ok {
		if v, ok := sku["display_price"].(string); ok {
			return v
		}
	}
	return ""
}
