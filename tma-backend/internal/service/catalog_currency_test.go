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
