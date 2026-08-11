package service

import (
	"regexp"
	"strings"

	"github.com/lib/pq"
	"tma-backend/internal/domain"
)

var (
	titleNoiseRe      = regexp.MustCompile(`[®™©]`)
	spaceRe           = regexp.MustCompile(`\s+`)
	editionRe         = regexp.MustCompile(`(?i)\s+[-–—]?\s*(standard|deluxe|ultimate|premium|gold|complete|goty|game of the year|game preview|cross[- ]gen|bundle|edition|digital deluxe|special edition|collectors collector's|vault edition|day one|pre-?order|набор|издание|версия|стандартное|делuxe|делюкс)(\s+edition|\s+версия|\s+издание)?.*$`)
	platformInTitleRe = regexp.MustCompile(`(?i)\s*[\(\[]?(playstation|ps4|ps5|xbox one|xbox series|xbox|pc|windows)[\)\]]?\s*$`)
)

func NormalizeGameTitle(title string) string {
	s := strings.TrimSpace(title)
	if s == "" {
		return ""
	}
	s = titleNoiseRe.ReplaceAllString(s, "")
	s = editionRe.ReplaceAllString(s, "")
	s = platformInTitleRe.ReplaceAllString(s, "")
	s = strings.ToLower(s)
	s = spaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func PlatformFamilyFromImport(source string, platforms pq.StringArray) string {
	for _, p := range platforms {
		switch strings.ToLower(p) {
		case "ps5", "ps4":
			return "ps"
		case "xbox":
			return "xbox"
		case "pc":
			return "pc"
		}
	}
	switch source {
	case domain.CatalogSourcePSStore:
		return "ps"
	case domain.CatalogSourceXboxStore:
		return "xbox"
	}
	return "other"
}

func PlatformFamilyFromProduct(platform domain.Platform) string {
	switch platform {
	case domain.PlatformPS4, domain.PlatformPS5:
		return "ps"
	case domain.PlatformXbox:
		return "xbox"
	case domain.PlatformPC:
		return "pc"
	default:
		return "other"
	}
}

func ProductPlatformRank(platform domain.Platform) int {
	switch platform {
	case domain.PlatformPS5:
		return 30
	case domain.PlatformXbox:
		return 25
	case domain.PlatformPS4:
		return 20
	case domain.PlatformPC:
		return 15
	default:
		return 0
	}
}

func PickProductPlatform(platforms pq.StringArray, requested string) string {
	if requested != "" {
		for _, platform := range platforms {
			if platform == requested {
				return requested
			}
		}
	}
	for _, preferred := range []string{"ps5", "ps4", "xbox", "pc"} {
		for _, platform := range platforms {
			if platform == preferred {
				return platform
			}
		}
	}
	return "xbox"
}
