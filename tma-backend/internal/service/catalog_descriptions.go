package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"tma-backend/internal/domain"
)

const (
	psDescriptionRegion = "RU"
	psDescriptionLocale = "ru"
)

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)
var htmlEntityRe = regexp.MustCompile(`&[a-zA-Z]+;|&#\d+;`)

func cleanDescription(raw string) string {
	s := htmlTagRe.ReplaceAllString(raw, "")
	s = htmlEntityRe.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.TrimSpace(s)
	// Collapse multiple spaces
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}

func (s *CatalogParserService) fetchPSDescriptionRU(ctx context.Context, externalID string) string {
	if strings.TrimSpace(externalID) == "" {
		return ""
	}
	endpoint := fmt.Sprintf(
		"https://store.playstation.com/store/api/chihiro/00_09_000/container/%s/%s/19/%s/0",
		psDescriptionRegion, psDescriptionLocale, url.PathEscape(externalID),
	)
	var game psStoreItem
	if err := s.getJSONWithClient(ctx, s.proxyClient(ctx), endpoint, &game); err != nil {
		return ""
	}
	return cleanDescription(game.Description)
}

func pickXboxLocalized(props []xboxLocalizedProperty) xboxLocalizedProperty {
	if len(props) == 0 {
		return xboxLocalizedProperty{}
	}
	for _, p := range props {
		lang := strings.ToLower(strings.TrimSpace(p.Language))
		if strings.HasPrefix(lang, "ru") && strings.TrimSpace(p.ShortDescription) != "" {
			return p
		}
	}
	for _, p := range props {
		if strings.TrimSpace(p.ShortDescription) != "" {
			return p
		}
	}
	return props[0]
}

func (s *CatalogParserService) RussianDescription(ctx context.Context, source string, externalID, current string) string {
	current = strings.TrimSpace(current)
	switch source {
	case domain.CatalogSourcePSStore:
		if ru := s.fetchPSDescriptionRU(ctx, externalID); ru != "" {
			return ru
		}
	case domain.CatalogSourceXboxStore:
		products, err := s.fetchXboxProducts(ctx, []string{externalID})
		if err == nil && len(products) > 0 {
			localized := pickXboxLocalized(products[0].LocalizedProperties)
			if ru := strings.TrimSpace(localized.ShortDescription); ru != "" {
				return ru
			}
		}
	}
	return current
}
