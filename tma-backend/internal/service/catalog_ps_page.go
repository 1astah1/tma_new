package service

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"time"
)

// Старый chihiro-API по части игр не отдаёт SKU вообще: у предзаказов и части
// новинок цены там просто нет, хотя на сайте магазина она есть. Для таких
// позиций забираем цену и дату релиза со страницы товара.
var (
	psPageBasePriceRe = regexp.MustCompile(`"basePrice":"([^"]+)"`)
	psPageReleaseRe   = regexp.MustCompile(`"releaseDate":"([^"]+)"`)
)

type psPageInfo struct {
	DisplayPrice string
	ReleaseDate  time.Time
}

// fetchPSStorePage читает страницу товара. Первая цена на странице относится
// к самому товару — страница строится вокруг него, дальше идут издания и
// сопутствующие предложения.
func (s *CatalogParserService) fetchPSStorePage(ctx context.Context, locale, externalID string) (psPageInfo, bool) {
	if externalID == "" || locale == "" {
		return psPageInfo{}, false
	}

	endpoint := fmt.Sprintf("https://store.playstation.com/%s/product/%s", locale, url.PathEscape(externalID))
	body, err := s.getTextWithClient(ctx, s.proxyClient(ctx), endpoint)
	if err != nil {
		return psPageInfo{}, false
	}

	info := psPageInfo{}
	if match := psPageBasePriceRe.FindStringSubmatch(body); len(match) > 1 {
		info.DisplayPrice = match[1]
	}
	if match := psPageReleaseRe.FindStringSubmatch(body); len(match) > 1 {
		info.ReleaseDate = parseReleaseDate(match[1])
	}

	return info, info.DisplayPrice != "" || !info.ReleaseDate.IsZero()
}

// psPageTurkeyPrice — цена в рублях по турецкой витрине, минуя старый API.
func (s *CatalogParserService) psPageTurkeyPrice(ctx context.Context, externalID string) (*float64, time.Time) {
	info, ok := s.fetchPSStorePage(ctx, "tr-tr", externalID)
	if !ok {
		return nil, time.Time{}
	}
	amount := parseDisplayPriceTRY(info.DisplayPrice)
	if amount <= 0 {
		return nil, info.ReleaseDate
	}
	rub := TurkeyNominalPrice(amount)
	return &rub, info.ReleaseDate
}

// psPageUkrainePrice — то же для украинской витрины.
func (s *CatalogParserService) psPageUkrainePrice(ctx context.Context, externalID string) *float64 {
	info, ok := s.fetchPSStorePage(ctx, "uk-ua", externalID)
	if !ok {
		return nil
	}
	amount := parseDisplayPriceUAH(info.DisplayPrice)
	if amount <= 0 {
		return nil
	}
	rub := UkrainePrice(amount)
	return &rub
}
