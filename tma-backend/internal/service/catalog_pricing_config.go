package service

import (
	"context"
	"encoding/json"
	"sync"

	"tma-backend/internal/repository"
)

const PricingSettingsKey = "pricing_formulas"

// PricingFormulas — настраиваемые коэффициенты цен каталога.
type PricingFormulas struct {
	TRYToRUBManual    float64 `json:"try_rub_manual"`
	TurkeyMarkup      float64 `json:"turkey_markup"`
	UkraineMarkup     float64 `json:"ukraine_markup"`
	XboxUsdMultiplier float64 `json:"xbox_usd_multiplier"`
	MinPriceRUB       float64 `json:"min_price_rub"`
}

func DefaultPricingFormulas() PricingFormulas {
	return PricingFormulas{
		TurkeyMarkup:      2.2,
		UkraineMarkup:     2.3,
		XboxUsdMultiplier: 80,
		MinPriceRUB:       149,
	}
}

var (
	pricingMu  sync.RWMutex
	pricingCfg = DefaultPricingFormulas()
)

func currentPricing() PricingFormulas {
	pricingMu.RLock()
	defer pricingMu.RUnlock()
	return pricingCfg
}

func MinPaidPriceRUB() float64 {
	cfg := currentPricing()
	if cfg.MinPriceRUB > 0 {
		return cfg.MinPriceRUB
	}
	return 149
}

func TurkeyMarkup() float64 {
	cfg := currentPricing()
	if cfg.TurkeyMarkup > 0 {
		return cfg.TurkeyMarkup
	}
	return 2.2
}

func UkraineMarkup() float64 {
	cfg := currentPricing()
	if cfg.UkraineMarkup > 0 {
		return cfg.UkraineMarkup
	}
	return 2.3
}

func XboxUsdMultiplier() float64 {
	cfg := currentPricing()
	if cfg.XboxUsdMultiplier > 0 {
		return cfg.XboxUsdMultiplier
	}
	return 80
}

func ManualTRYToRUBRate() float64 {
	return currentPricing().TRYToRUBManual
}

func ApplyPricingFormulas(raw PricingFormulas) {
	def := DefaultPricingFormulas()
	if raw.TurkeyMarkup > 0 {
		def.TurkeyMarkup = raw.TurkeyMarkup
	}
	if raw.UkraineMarkup > 0 {
		def.UkraineMarkup = raw.UkraineMarkup
	}
	if raw.XboxUsdMultiplier > 0 {
		def.XboxUsdMultiplier = raw.XboxUsdMultiplier
	}
	if raw.MinPriceRUB > 0 {
		def.MinPriceRUB = raw.MinPriceRUB
	}
	if raw.TRYToRUBManual > 0 {
		def.TRYToRUBManual = raw.TRYToRUBManual
	}
	pricingMu.Lock()
	pricingCfg = def
	pricingMu.Unlock()
}

func ParsePricingFormulasValue(value interface{}) (PricingFormulas, error) {
	def := DefaultPricingFormulas()
	if value == nil {
		return def, nil
	}
	var raw []byte
	switch v := value.(type) {
	case string:
		if v == "" {
			return def, nil
		}
		raw = []byte(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return def, err
		}
		raw = b
	}
	var parsed PricingFormulas
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return def, err
	}
	if parsed.TurkeyMarkup > 0 {
		def.TurkeyMarkup = parsed.TurkeyMarkup
	}
	if parsed.UkraineMarkup > 0 {
		def.UkraineMarkup = parsed.UkraineMarkup
	}
	if parsed.XboxUsdMultiplier > 0 {
		def.XboxUsdMultiplier = parsed.XboxUsdMultiplier
	}
	if parsed.MinPriceRUB > 0 {
		def.MinPriceRUB = parsed.MinPriceRUB
	}
	if parsed.TRYToRUBManual > 0 {
		def.TRYToRUBManual = parsed.TRYToRUBManual
	}
	return def, nil
}

func LoadPricingFormulasFromSettings(ctx context.Context, repo *repository.SettingsRepo) error {
	if repo == nil {
		return nil
	}
	setting, err := repo.Get(ctx, PricingSettingsKey)
	if err != nil {
		return nil
	}
	cfg, err := ParsePricingFormulasValue(setting["value"])
	if err != nil {
		return err
	}
	ApplyPricingFormulas(cfg)
	return nil
}

func GetPricingFormulas() PricingFormulas {
	return currentPricing()
}
