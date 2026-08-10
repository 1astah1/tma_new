package service

import (
	"regexp"
	"strings"
)

// Curated boosts mirror PlayStation / Xbox download and sales charts (2025–2026).
// Tier 1 — annual blockbusters; tier 2 — strong shooters & evergreen hits; tier 3 — solid catalog staples.
// Keys are matched against normalized title_key via strings.Contains.
type popularityTier struct {
	keys  []string
	score float64
}

var curatedPopularity = []popularityTier{
	{
		keys: []string{
			"call of duty", "battlefield", "grand theft auto", "fortnite",
			"ea sports fc", "fifa", "nba 2k", "madden nfl",
		},
		score: 700,
	},
	{
		keys: []string{
			"helldivers", "apex legends", "counter-strike", "destiny 2", "rainbow six",
			"arc raiders", "resident evil requiem", "red dead redemption", "minecraft",
			"overwatch", "pubg", "warzone", "the finals", "borderlands", "doom", "halo",
			"fallout", "007 first light", "ghost of yotei",
		},
		score: 500,
	},
	{
		keys: []string{
			"titanfall", "escape from tarkov", "sniper elite", "far cry", "cyberpunk",
			"assassin's creed", "ghost of tsushima", "the last of us", "uncharted",
			"god of war", "marvel's spider-man", "spider-man", "elden ring", "forza",
			"monster hunter", "dying light", "dead by daylight", "payday", "back 4 blood",
			"ready or not", "remnant", "deep rock galactic", "no man's sky", "rocket league",
			"warhammer 40,000", "warhammer 40000", "hell let loose", "squad", "insurgency",
			"killzone", "returnal", "ratchet", "gran turismo", "nba live",
		},
		score: 350,
	},
}

var shooterKeywords = []string{
	"shooter", "fps", "battle royale", "tactical shooter", "extraction shooter",
	"call of duty", "battlefield", "counter-strike", "apex legends", "valorant", "pubg",
	"fortnite", "overwatch", "rainbow six", "destiny", "halo", "doom", "borderlands",
	"helldivers", "arc raiders", "the finals", "titanfall", "warzone", "sniper",
	"hell let loose", "insurgency", "squad", "killzone", "left 4 dead", "back 4 blood",
}

func effectiveTitleKey(titleKey, title string) string {
	if titleKey != "" {
		return titleKey
	}
	return NormalizeGameTitle(title)
}

func CuratedPopularityBoost(titleKey, title string) float64 {
	key := effectiveTitleKey(titleKey, title)
	if key == "" {
		return 0
	}
	best := 0.0
	for _, tier := range curatedPopularity {
		for _, needle := range tier.keys {
			if strings.Contains(key, needle) && tier.score > best {
				best = tier.score
				break
			}
		}
	}
	return best
}

func ShooterAudienceBoost(titleKey, title string) float64 {
	key := effectiveTitleKey(titleKey, title)
	if key == "" {
		return 0
	}
	for _, kw := range shooterKeywords {
		if strings.Contains(key, kw) {
			return 80
		}
	}
	return 0
}

const (
	homeFeedPriorityCOD        = 10000
	homeFeedPriorityTopShooter = 8000
	homeFeedPriorityShooter    = 5000
)

// HomeFeedPriorityScore ranks products on the home feed (preorders / new releases).
// Call of Duty is always first among shooters; then other FPS blockbusters.
func HomeFeedPriorityScore(titleKey, title string) float64 {
	key := effectiveTitleKey(titleKey, title)
	if key == "" {
		return 0
	}
	if strings.Contains(key, "call of duty") {
		return homeFeedPriorityCOD
	}
	curated := CuratedPopularityBoost(titleKey, title)
	if curated >= 700 {
		return homeFeedPriorityTopShooter + curated
	}
	if curated >= 500 || ShooterAudienceBoost(titleKey, title) > 0 {
		return homeFeedPriorityShooter + curated
	}
	return curated
}

func OrderCountBoost(count int) float64 {
	if count <= 0 {
		return 0
	}
	boost := float64(count) * 30
	if boost > 250 {
		return 250
	}
	return boost
}

var coinPackPrefixRe = regexp.MustCompile(`^\d+[\s\d]*\s*(монет|monet|coins|points|кредит|vc|v-bucks|vbucks|v bucks|коин)`)
var multiQuantityPrefixRe = regexp.MustCompile(`^\d+\s+\d`)
var largeQuantityPrefixRe = regexp.MustCompile(`^\d{3,}\s+\S`)
var fortniteCurrencyRe = regexp.MustCompile(`^fortnite:\s*\d`)

// IsPopularFeedExcluded filters in-game currency packs and similar non-game SKUs from auto popular.
func IsPopularFeedExcluded(titleKey, title string) bool {
	key := effectiveTitleKey(titleKey, title)
	if key == "" {
		return false
	}
	lower := strings.ToLower(key)
	if strings.HasPrefix(lower, "007") {
		return false
	}
	if coinPackPrefixRe.MatchString(lower) {
		return true
	}
	if multiQuantityPrefixRe.MatchString(lower) {
		return true
	}
	if largeQuantityPrefixRe.MatchString(lower) && (strings.Contains(lower, "call of duty") || strings.Contains(lower, "fortnite")) {
		return true
	}
	if fortniteCurrencyRe.MatchString(lower) {
		return true
	}
	if strings.Contains(lower, "battlefield") {
		if strings.Contains(lower, " pro") || strings.Contains(lower, "search and destroy") || strings.Contains(lower, "advanced") {
			return true
		}
	}
	if strings.Contains(lower, "монет") || strings.Contains(lower, " coins") || strings.Contains(lower, "points") {
		return true
	}
	if strings.HasPrefix(lower, "пакет ") || strings.HasPrefix(lower, "набор ") || strings.Contains(lower, " bundle") {
		if strings.Contains(lower, "battlefield") || strings.Contains(lower, "call of duty") {
			return true
		}
	}
	return false
}
