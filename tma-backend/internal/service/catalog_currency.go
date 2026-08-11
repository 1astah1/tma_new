package service

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	psStoreRegion        = "TR"
	psStoreLocale        = "tr"
	psStoreRegionUA      = "UA"
	psStoreLocaleUA      = "uk"
	xboxStoreMarket      = "US"
	xboxStoreLocale      = "en-US"
	xboxDescriptionLangs = "ru-RU,en-US"
	defaultTRYToRUB      = 2.85
)

var (
	tryRubMu      sync.RWMutex
	tryRubRate    = defaultTRYToRUB
	tryRubFetched time.Time
)

func TRYToRUBRate() float64 {
	tryRubMu.RLock()
	defer tryRubMu.RUnlock()
	return tryRubRate
}

func ConvertTRYToRUB(tryAmount float64) float64 {
	if tryAmount <= 0 {
		return 0
	}
	rate := TRYToRUBRate()
	if manual := ManualTRYToRUBRate(); manual > 0 {
		rate = manual
	}
	rub := tryAmount * rate
	return math.Round(rub*100) / 100
}

// TurkeyNominalPrice computes the nominal-based Turkey price:
// round up to nearest nominal (250,500,750,1000,1250,1500, then +250 step), then * 2.2
func TurkeyNominalPrice(tryAmount float64) float64 {
	if tryAmount <= 0 {
		return 0
	}
	nominals := []float64{250, 500, 750, 1000, 1250, 1500}
	markup := TurkeyMarkup()
	for _, n := range nominals {
		if tryAmount <= n {
			return math.Round(n*markup*100) / 100
		}
	}
	// Above 1500, round up to next 250
	nominal := math.Ceil(tryAmount/250) * 250
	return math.Round(nominal*markup*100) / 100
}

// UkrainePrice computes the Ukraine price: UAH * 2.3
func UkrainePrice(uahAmount float64) float64 {
	if uahAmount <= 0 {
		return 0
	}
	return math.Round(uahAmount*UkraineMarkup()*100) / 100
}

// XboxUSAPrice computes the Xbox USA price: USD * 80
func XboxUSAPrice(usdAmount float64) float64 {
	if usdAmount <= 0 {
		return 0
	}
	return math.Round(usdAmount*XboxUsdMultiplier()*100) / 100
}

func StorePriceToRUB(amount float64, currency string) (float64, bool) {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "RUB", "RUR":
		return math.Round(amount*100) / 100, amount > 0
	case "TRY":
		rub := ConvertTRYToRUB(amount)
		return rub, rub > 0
	default:
		return 0, false
	}
}

func parseDisplayPriceTRY(display string) float64 {
	display = strings.ToLower(strings.TrimSpace(display))
	if display == "" || strings.Contains(display, "free") || strings.Contains(display, "ücretsiz") {
		return 0
	}
	display = strings.ReplaceAll(display, "tl", "")
	display = strings.ReplaceAll(display, "try", "")
	display = strings.ReplaceAll(display, "₺", "")
	display = strings.ReplaceAll(display, "\u00a0", "")
	display = strings.ReplaceAll(display, " ", "")
	// «3.199,00 TL»: точка — разделитель тысяч, запятая — дробная часть.
	// Старый API отдавал цены без тысяч, поэтому раньше это не всплывало.
	display = strings.ReplaceAll(display, ".", "")
	display = strings.ReplaceAll(display, ",", ".")
	match := regexp.MustCompile(`[\d.]+`).FindString(display)
	if match == "" {
		return 0
	}
	value, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0
	}
	return value
}

// parseDisplayPriceUAH parses a display price string in Ukrainian hryvnia (UAH/₴)
func parseDisplayPriceUAH(display string) float64 {
	display = strings.ToLower(strings.TrimSpace(display))
	if display == "" || strings.Contains(display, "free") || strings.Contains(display, "безкоштовно") {
		return 0
	}
	display = strings.ReplaceAll(display, "uah", "")
	display = strings.ReplaceAll(display, "₴", "")
	display = strings.ReplaceAll(display, "\u00a0", "")
	display = strings.ReplaceAll(display, " ", "")
	display = strings.ReplaceAll(display, ",", ".")
	match := regexp.MustCompile(`[\d.]+`).FindString(display)
	if match == "" {
		return 0
	}
	value, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0
	}
	return value
}

func RefreshTRYRUBRate(ctx context.Context, client *http.Client) error {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.cbr-xml-daily.ru/daily_json.js", nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var payload struct {
		Valute map[string]struct {
			Nominal float64 `json:"Nominal"`
			Value   float64 `json:"Value"`
		} `json:"Valute"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	try, ok := payload.Valute["TRY"]
	if !ok || try.Nominal <= 0 || try.Value <= 0 {
		return err
	}
	rate := try.Value / try.Nominal
	if rate <= 0 {
		return err
	}
	tryRubMu.Lock()
	tryRubRate = rate
	tryRubFetched = time.Now()
	tryRubMu.Unlock()
	return nil
}

func ensureTRYRUBRate(ctx context.Context, client *http.Client) {
	tryRubMu.RLock()
	stale := tryRubFetched.IsZero() || time.Since(tryRubFetched) > 12*time.Hour
	tryRubMu.RUnlock()
	if !stale {
		return
	}
	if err := RefreshTRYRUBRate(ctx, client); err != nil {
		// keep last/default rate
	}
}
