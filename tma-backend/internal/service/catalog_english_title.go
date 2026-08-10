package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"tma-backend/internal/domain"
)

var turkishLetters = map[rune]bool{
	'ç': true, 'Ç': true, 'ğ': true, 'Ğ': true, 'ı': true, 'İ': true,
	'ö': true, 'Ö': true, 'ş': true, 'Ş': true, 'ü': true, 'Ü': true,
}

func ContainsTurkishLetters(s string) bool {
	for _, r := range s {
		if turkishLetters[r] {
			return true
		}
	}
	return false
}

type EnglishTitleResolver struct {
	client *http.Client
	cache  sync.Map
	lookup func(ctx context.Context, productID uuid.UUID) (externalID string, source string, ok bool)
}

func NewEnglishTitleResolver(lookup func(ctx context.Context, productID uuid.UUID) (string, string, bool)) *EnglishTitleResolver {
	return &EnglishTitleResolver{
		client: &http.Client{Timeout: 8 * time.Second},
		lookup: lookup,
	}
}

func (r *EnglishTitleResolver) Resolve(ctx context.Context, productID uuid.UUID, title string) string {
	if title == "" || !ContainsTurkishLetters(title) {
		return title
	}
	if cached, ok := r.cache.Load(productID); ok {
		if v := cached.(string); v != "" {
			return v
		}
	}
	externalID, source, ok := r.lookup(ctx, productID)
	if !ok || externalID == "" || source != domain.CatalogSourcePSStore {
		return title
	}
	enTitle, err := fetchPSEnglishTitle(ctx, r.client, externalID)
	if err != nil || enTitle == "" {
		return title
	}
	r.cache.Store(productID, enTitle)
	return enTitle
}

func fetchPSEnglishTitle(ctx context.Context, client *http.Client, externalID string) (string, error) {
	endpoint := fmt.Sprintf(
		"https://store.playstation.com/store/api/chihiro/00_09_000/container/US/en/19/%s/0",
		url.PathEscape(externalID),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ps store status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var payload struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	name := strings.TrimSpace(payload.Name)
	if name == "" || ContainsTurkishLetters(name) {
		return "", fmt.Errorf("empty or non-english title")
	}
	return name, nil
}
