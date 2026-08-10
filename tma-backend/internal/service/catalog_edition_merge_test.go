package service

import (
	"encoding/json"
	"testing"

	"tma-backend/internal/domain"
)

func TestMergeProductPricesUFCSiblings(t *testing.T) {
	ps := domain.Product{
		TitleKey: "ufc 6",
		Platform: domain.PlatformPS5,
		Type:     domain.ProductTypeGame,
		Prices:   json.RawMessage(`{"tr":8799,"ua":4330}`),
	}
	xbox := domain.Product{
		TitleKey: "ufc 6",
		Platform: domain.PlatformXbox,
		Type:     domain.ProductTypeGame,
		Price:    5599.2,
	}

	merged := MergeProductPrices([]domain.Product{ps, xbox})
	if merged == nil {
		t.Fatal("expected merged prices")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(merged, &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	catalog, ok := payload["edition_catalog"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected edition_catalog, got %#v", payload)
	}
	if _, ok := catalog["ps_tr"]; !ok {
		t.Fatalf("expected ps_tr in catalog: %#v", catalog)
	}
	if _, ok := catalog["ps_ua"]; !ok {
		t.Fatalf("expected ps_ua in catalog: %#v", catalog)
	}
}
