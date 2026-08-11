package service

import (
	"testing"
	"time"

	"tma-backend/internal/domain"
)

func TestIsNonGameStoreItem(t *testing.T) {
	cases := map[string]bool{
		"Game Demo":                  true,
		"Dynamic Theme - Spider-Man": true,
		"Slot Machine Casino":        false,
		"GGmuks Casino: Slots (PS4)": false,
		"Elden Ring":                 false,
	}
	for title, want := range cases {
		if got := IsNonGameStoreItem(title); got != want {
			t.Fatalf("%q: got %v want %v", title, got, want)
		}
	}
}

func TestPreorderWithoutPriceIsSellable(t *testing.T) {
	future := time.Now().UTC().AddDate(0, 2, 0)
	item := &domain.CatalogImport{
		Title:       "Dune: Awakening",
		GameSection: "preorder",
		ReleaseDate: &future,
	}
	if !IsSellableCatalogItem(item) {
		t.Fatal("preorder with future date and no price should be sellable")
	}
}

func TestIsFreePrice(t *testing.T) {
	free := 0.0
	paid := 499.0
	one := 1.0
	if !IsFreePrice(&free, "") {
		t.Fatal("expected free for 0")
	}
	if IsFreePrice(&paid, "") {
		t.Fatal("expected paid for 499")
	}
	if !IsFreePrice(&one, "") {
		t.Fatal("expected free for 1 rub fallback")
	}
	if !IsFreePrice(nil, "Бесплатно") {
		t.Fatal("expected free for display text")
	}
}
