package service

import (
	"testing"
	"time"

	"tma-backend/internal/domain"
)

func TestCuratedPopularityBoost(t *testing.T) {
	cases := map[string]float64{
		"call of duty black ops 7": 700,
		"battlefield 6":            700,
		"grand theft auto v":       700,
		"fortnite":                 700,
		"helldivers 2":             500,
		"arc raiders":              500,
		"elden ring":               350,
		"totally unknown indie":    0,
	}
	for title, want := range cases {
		got := CuratedPopularityBoost(NormalizeGameTitle(title), title)
		if got != want {
			t.Fatalf("%q: got %.0f want %.0f", title, got, want)
		}
	}
}

func TestShooterAudienceBoost(t *testing.T) {
	if got := ShooterAudienceBoost("helldivers 2", "Helldivers 2"); got != 80 {
		t.Fatalf("helldivers: got %.0f want 80", got)
	}
	if got := ShooterAudienceBoost("elden ring", "Elden Ring"); got != 0 {
		t.Fatalf("elden ring should not get shooter boost, got %.0f", got)
	}
}

func TestOrderCountBoost(t *testing.T) {
	if got := OrderCountBoost(0); got != 0 {
		t.Fatalf("zero orders: got %.0f", got)
	}
	if got := OrderCountBoost(3); got != 90 {
		t.Fatalf("3 orders: got %.0f want 90", got)
	}
	if got := OrderCountBoost(20); got != 250 {
		t.Fatalf("20 orders should cap at 250, got %.0f", got)
	}
}

func TestIsPopularFeedExcluded(t *testing.T) {
	cases := map[string]bool{
		"9 500 монет call of duty: modern warfare": true,
		"500 монет call of duty: modern warfare":   true,
		"9 500 coins call of duty: modern warfare": true,
		"call of duty: black ops 7":                false,
		"call of duty: modern warfare 4":           false,
		"пакет 3 battlefield pro - battlefield 6":    true,
		"fortnite: 2400 v-bucks":                   true,
		"007 first light":                          false,
	}
	for key, want := range cases {
		if got := IsPopularFeedExcluded(key, key); got != want {
			t.Fatalf("%q: got %v want %v", key, got, want)
		}
	}
}

func TestHomeFeedPriorityScore(t *testing.T) {
	cod := HomeFeedPriorityScore("call of duty: black ops 7", "Call of Duty: Black Ops 7")
	bf := HomeFeedPriorityScore("battlefield 6", "Battlefield 6")
	helldivers := HomeFeedPriorityScore("helldivers 2", "Helldivers 2")
	indie := HomeFeedPriorityScore("cornfield", "Cornfield")
	if cod <= bf || cod <= helldivers {
		t.Fatalf("COD should rank above other shooters: cod=%.0f bf=%.0f helldivers=%.0f", cod, bf, helldivers)
	}
	if bf <= indie || helldivers <= indie {
		t.Fatalf("shooters should rank above indie: bf=%.0f helldivers=%.0f indie=%.0f", bf, helldivers, indie)
	}
}

func TestComputeScorePrefersHitOverObscureNewRelease(t *testing.T) {
	v := &VitrinaService{}
	recent := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	old := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	obscure := popularityTestProduct("Unknown Roguelike X", &recent, 499, "ps5", true)
	hit := popularityTestProduct("Call of Duty: Black Ops 7", &old, 499, "ps5", true)

	obscureScore := v.ComputeScore(obscure, ScoreContext{})
	hitScore := v.ComputeScore(hit, ScoreContext{})
	if hitScore <= obscureScore {
		t.Fatalf("hit should outrank obscure new release: hit=%.0f obscure=%.0f", hitScore, obscureScore)
	}
}

func popularityTestProduct(title string, release *time.Time, price float64, platform string, withImage bool) domain.Product {
	img := "https://example.com/cover.jpg"
	p := domain.Product{
		Title:       title,
		TitleKey:    NormalizeGameTitle(title),
		ReleaseDate: release,
		Price:       price,
		Platform:    domain.Platform(platform),
	}
	if withImage {
		p.ImageURL = &img
	}
	return p
}
