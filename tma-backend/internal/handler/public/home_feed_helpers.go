package public

import (
	"sort"
	"strings"
	"time"

	"tma-backend/internal/domain"
	"tma-backend/internal/service"
)

const homeFeedFetchMultiplier = 3
const homeFeedCandidateMultiplier = 8

func homeFeedFetchLimit(limit int) int {
	if limit <= 0 {
		limit = 12
	}
	fetch := limit * homeFeedFetchMultiplier
	if fetch < limit+12 {
		fetch = limit + 12
	}
	return fetch
}

func homeFeedCandidatePool(limit int) int {
	if limit <= 0 {
		limit = 12
	}
	pool := limit * homeFeedCandidateMultiplier
	if pool < 80 {
		pool = 80
	}
	if pool > 200 {
		pool = 200
	}
	return pool
}

func compareReleaseDate(a, b *time.Time, newestFirst bool) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil {
		return false
	}
	if b == nil {
		return true
	}
	if newestFirst {
		return a.After(*b)
	}
	return a.Before(*b)
}

func prioritizeFeedProducts(products []domain.Product, newestFirst bool) {
	sort.SliceStable(products, func(i, j int) bool {
		pi := service.HomeFeedPriorityScore(products[i].TitleKey, products[i].Title)
		pj := service.HomeFeedPriorityScore(products[j].TitleKey, products[j].Title)
		if pi != pj {
			return pi > pj
		}
		return compareReleaseDate(products[i].ReleaseDate, products[j].ReleaseDate, newestFirst)
	})
}

func productFeedKey(p domain.Product) string {
	if key := strings.TrimSpace(p.TitleKey); key != "" {
		return key
	}
	return service.NormalizeGameTitle(p.Title)
}

func productFeedAliases(p domain.Product) []string {
	keys := []string{
		productFeedKey(p),
		service.NormalizeGameTitle(p.Title),
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, key)
	}
	return out
}

func feedKeysSimilar(a, b string) bool {
	if a == b {
		return true
	}
	if len(a) < 5 || len(b) < 5 {
		return false
	}
	return strings.Contains(a, b) || strings.Contains(b, a)
}

func isFeedProductTaken(p domain.Product, seen map[string]bool) bool {
	for _, key := range productFeedAliases(p) {
		if seen[key] {
			return true
		}
		for existing := range seen {
			if feedKeysSimilar(existing, key) {
				return true
			}
		}
	}
	return false
}

func registerFeedProduct(p domain.Product, seen map[string]bool) {
	for _, key := range productFeedAliases(p) {
		seen[key] = true
	}
}

func selectFeedProducts(products []domain.Product, seen map[string]bool, limit int) []domain.Product {
	if limit <= 0 {
		limit = 12
	}
	out := make([]domain.Product, 0, limit)
	for _, p := range products {
		if p.Status != domain.ProductStatusActive {
			continue
		}
		if isFeedProductTaken(p, seen) {
			continue
		}
		out = append(out, p)
		registerFeedProduct(p, seen)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func assembleHomeFeedSections(preordersRaw, newRaw, popularRaw []domain.Product, limit int) (preorders, newReleases, popular []domain.Product) {
	if limit <= 0 {
		limit = 12
	}
	seen := map[string]bool{}
	preorders = selectFeedProducts(preordersRaw, seen, limit)
	newReleases = selectFeedProducts(newRaw, seen, limit)
	popular = selectFeedProducts(popularRaw, seen, limit)
	return preorders, newReleases, popular
}
