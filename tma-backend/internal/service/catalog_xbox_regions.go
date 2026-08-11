package service

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"tma-backend/internal/domain"
)

// Xbox торгует в разных странах, и турецкая витрина заметно дешевле
// американской — покупателю нужен этот выбор так же, как на PlayStation.
const xboxTurkeyMarket = "TR"

// fetchXboxMarketPrices возвращает цены в рублях по нужному рынку:
// ключ — идентификатор товара, значение — уже пересчитанная цена.
func (s *CatalogParserService) fetchXboxMarketPrices(ctx context.Context, ids []string, market string) map[string]float64 {
	if len(ids) == 0 {
		return nil
	}

	q := url.Values{}
	q.Set("bigIds", strings.Join(ids, ","))
	q.Set("market", market)
	q.Set("languages", "en-US")
	q.Set("MS-CV", "DGU1mcuYo0WMMp+F.1")

	var payload xboxCatalogResponse
	if err := s.getJSON(ctx, "https://displaycatalog.mp.microsoft.com/v7.0/products?"+q.Encode(), &payload); err != nil {
		return nil
	}

	result := make(map[string]float64, len(payload.Products))
	for _, product := range payload.Products {
		if rub, ok := xboxMarketPriceRUB(product, market); ok {
			result[product.ProductID] = rub
		}
	}
	return result
}

// xboxMarketPriceRUB пересчитывает цену рынка в рубли по тем же формулам,
// что и остальной каталог: лиры — через номинал карты, доллары — множителем.
func xboxMarketPriceRUB(product xboxProduct, market string) (float64, bool) {
	var best float64
	var currency string

	for _, sku := range product.DisplaySkuAvailabilities {
		for _, availability := range sku.Availabilities {
			if availability.OrderManagementData == nil {
				continue
			}
			p := availability.OrderManagementData.Price
			amount := p.ListPrice
			if amount <= 0 {
				amount = p.MSRP
			}
			if amount <= 0 {
				continue
			}
			if amount > best {
				best = amount
				currency = strings.ToUpper(strings.TrimSpace(p.CurrencyCode))
			}
		}
	}
	if best <= 0 {
		return 0, false
	}

	switch currency {
	case "TRY":
		return TurkeyNominalPrice(best), true
	case "USD":
		return XboxUSAPrice(best), true
	case "UAH":
		return UkrainePrice(best), true
	}
	return 0, false
}

// attachXboxTurkeyPrice дописывает в позицию цену турецкой витрины Xbox.
func (s *CatalogParserService) attachXboxTurkeyPrice(ctx context.Context, item *domain.CatalogImport) {
	if item == nil || item.ExternalID == "" {
		return
	}
	prices := s.fetchXboxMarketPrices(ctx, []string{item.ExternalID}, xboxTurkeyMarket)
	rub, ok := prices[item.ExternalID]
	if !ok || rub <= 0 {
		return
	}

	current := map[string]float64{}
	_ = json.Unmarshal(item.Prices, &current)
	current["xbox_tr"] = rub
	if encoded, err := json.Marshal(current); err == nil {
		item.Prices = encoded
	}

	// Показываем «от» по самой дешёвой витрине — покупатель выберет сам.
	if item.OriginalPriceRUB == nil || rub < *item.OriginalPriceRUB {
		item.OriginalPriceRUB = &rub
	}
}
