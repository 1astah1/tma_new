package public

import (
	"strings"
	"testing"

	"tma-backend/internal/domain"
)

func TestFeedKeysSimilar(t *testing.T) {
	if !feedKeysSimilar("ufc 6", "ultimate ufc 6 bundle") {
		t.Fatal("expected similar ufc keys")
	}
	if feedKeysSimilar("cat", "catalog") {
		t.Fatal("short keys should not fuzzy-match")
	}
}

func TestPrioritizeFeedProductsPutsCODFirst(t *testing.T) {
	items := []domain.Product{
		{Title: "Cornfield", TitleKey: "cornfield", Status: domain.ProductStatusActive},
		{Title: "Call of Duty: Black Ops 7", TitleKey: "call of duty: black ops 7", Status: domain.ProductStatusActive},
		{Title: "Helldivers 2", TitleKey: "helldivers 2", Status: domain.ProductStatusActive},
	}
	prioritizeFeedProducts(items, true)
	if !strings.Contains(items[0].TitleKey, "call of duty") {
		t.Fatalf("expected COD first, got %q", items[0].TitleKey)
	}
}

func TestSelectFeedProductsDedupesAcrossSections(t *testing.T) {
	seen := map[string]bool{}
	first := selectFeedProducts([]domain.Product{
		{Title: "UFC 6", TitleKey: "ufc 6", Status: domain.ProductStatusActive},
	}, seen, 12)
	second := selectFeedProducts([]domain.Product{
		{Title: "Ultimate UFC 6 Edition", TitleKey: "ultimate ufc 6 edition", Status: domain.ProductStatusActive},
	}, seen, 12)
	if len(first) != 1 {
		t.Fatalf("expected 1 preorder, got %d", len(first))
	}
	if len(second) != 0 {
		t.Fatalf("expected duplicate ufc game to be skipped in next section, got %d", len(second))
	}
}
