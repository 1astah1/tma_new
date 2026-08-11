package service

import (
	"encoding/json"
	"fmt"
	"testing"
)

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

func xboxProductWithPrice(t *testing.T, listPrice float64, currency string) xboxProduct {
	t.Helper()
	raw := fmt.Sprintf(`{"DisplaySkuAvailabilities":[{"Availabilities":[{"OrderManagementData":{"Price":{"ListPrice":%v,"CurrencyCode":%q}}}]}]}`, listPrice, currency)
	var product xboxProduct
	if err := json.Unmarshal([]byte(raw), &product); err != nil {
		t.Fatalf("unmarshal xbox product: %v", err)
	}
	return product
}

// Xbox-каталог берётся только с рынка US: цена в USD пересчитывается множителем.
func TestXboxPriceUsesUSD(t *testing.T) {
	price, currency := xboxPrice(xboxProductWithPrice(t, 69.99, "USD"))
	if currency == nil || *currency != "USD" {
		t.Fatalf("expected USD currency, got %v", currency)
	}
	if want := XboxUSAPrice(69.99); price == nil || *price != want {
		t.Fatalf("expected %v RUB, got %v", want, price)
	}
}

func TestXboxPriceIgnoresNonUSD(t *testing.T) {
	price, currency := xboxProductPrice(t, "TRY")
	if price != nil || currency != nil {
		t.Fatalf("expected no price for non-USD market, got %v %v", price, currency)
	}
}

func xboxProductPrice(t *testing.T, currency string) (*float64, *string) {
	t.Helper()
	return xboxPrice(xboxProductWithPrice(t, 325, currency))
}

// Со страницы магазина цены приходят с разделителем тысяч, из старого API —
// без него. Разбирать надо оба вида.
func TestParseDisplayPriceTRYWithThousands(t *testing.T) {
	cases := map[string]float64{
		"849,00 TL":    849,
		"3.199,00 TL":  3199,
		"4.399,00 TL":  4399,
		"12.500,50 TL": 12500.5,
	}
	for raw, want := range cases {
		if got := parseDisplayPriceTRY(raw); got != want {
			t.Errorf("parseDisplayPriceTRY(%q) = %v, ожидалось %v", raw, got, want)
		}
	}
}

func TestParseDisplayPriceUAHWithSpaces(t *testing.T) {
	cases := map[string]float64{
		"UAH 194,00":   194,
		"UAH 3 399,00": 3399,
	}
	for raw, want := range cases {
		if got := parseDisplayPriceUAH(raw); got != want {
			t.Errorf("parseDisplayPriceUAH(%q) = %v, ожидалось %v", raw, got, want)
		}
	}
}
