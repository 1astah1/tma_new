package service

import "testing"

func TestConvertTRYToRUB(t *testing.T) {
	tryRubMu.Lock()
	old := tryRubRate
	tryRubRate = 2.5
	tryRubMu.Unlock()
	defer func() {
		tryRubMu.Lock()
		tryRubRate = old
		tryRubMu.Unlock()
	}()

	if got := ConvertTRYToRUB(325); got != 812.5 {
		t.Fatalf("expected 812.5 got %v", got)
	}
}

func TestParseDisplayPriceTRY(t *testing.T) {
	if got := parseDisplayPriceTRY("849,00 TL"); got != 849 {
		t.Fatalf("expected 849 got %v", got)
	}
}

func TestXboxPriceAcceptsTRY(t *testing.T) {
	product := xboxProduct{
		DisplaySkuAvailabilities: []struct {
			Sku struct {
				MarketProperties []struct {
					FirstAvailableDate string `json:"FirstAvailableDate"`
				} `json:"MarketProperties"`
				Properties struct {
					IsPreOrder bool `json:"IsPreOrder"`
				} `json:"Properties"`
			} `json:"Sku"`
			Availabilities []struct {
				OrderManagementData *struct {
					Price struct {
						MSRP         float64 `json:"MSRP"`
						ListPrice    float64 `json:"ListPrice"`
						CurrencyCode string  `json:"CurrencyCode"`
					} `json:"Price"`
				} `json:"OrderManagementData"`
			} `json:"Availabilities"`
		}{
			{
				Availabilities: []struct {
					OrderManagementData *struct {
						Price struct {
							MSRP         float64 `json:"MSRP"`
							ListPrice    float64 `json:"ListPrice"`
							CurrencyCode string  `json:"CurrencyCode"`
						} `json:"Price"`
					} `json:"OrderManagementData"`
				}{
					{
						OrderManagementData: &struct {
							Price struct {
								MSRP         float64 `json:"MSRP"`
								ListPrice    float64 `json:"ListPrice"`
								CurrencyCode string  `json:"CurrencyCode"`
							} `json:"Price"`
						}{
							Price: struct {
								MSRP         float64 `json:"MSRP"`
								ListPrice    float64 `json:"ListPrice"`
								CurrencyCode string  `json:"CurrencyCode"`
							}{ListPrice: 325, CurrencyCode: "TRY"},
						},
					},
				},
			},
		},
	}
	tryRubMu.Lock()
	old := tryRubRate
	tryRubRate = 2.0
	tryRubMu.Unlock()
	defer func() {
		tryRubMu.Lock()
		tryRubRate = old
		tryRubMu.Unlock()
	}()

	price, currency := xboxPrice(product)
	if currency == nil || *currency != "TRY" {
		t.Fatalf("expected TRY currency")
	}
	if price == nil || *price != 650 {
		t.Fatalf("expected 650 RUB got %v", price)
	}
}
