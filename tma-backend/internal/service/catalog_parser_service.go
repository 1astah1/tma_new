package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lib/pq"
	"golang.org/x/net/proxy"
	"tma-backend/internal/domain"
	"tma-backend/internal/repository"
)

// STORE-MSF75508-NEWRELEASE Sony удалила — контейнер отдаёт 404,
// новинки и так приходят из FULLGAMES.
var psStoreContainers = []string{
	"STORE-MSF75508-FULLGAMES",
	"STORE-MSF75508-COMINGSOON",
}

var xboxBrowseTemplates = []string{
	"https://www.xbox.com/en-US/games/browse?xr=shellnav&page=%d",
	"https://www.xbox.com/en-US/games/browse/coming-soon?xr=shellnav&page=%d",
	"https://www.xbox.com/en-US/games/browse/new?page=%d",
}

var xboxChannelPages = []string{
	"https://www.xbox.com/en-US/games/browse/DynamicChannel.NewGames",
	"https://www.xbox.com/en-US/games/browse/DynamicChannel.ComingSoon",
	"https://www.xbox.com/en-US/games/browse/DynamicChannel.TopPaid",
	"https://www.xbox.com/en-US/games/browse/DynamicChannel.TopFree",
	"https://www.xbox.com/en-US/games/browse/DynamicChannel.Shooter",
	"https://www.xbox.com/en-US/games/browse/DynamicChannel.RacingAndFlying",
	"https://www.xbox.com/en-US/games/browse/DynamicChannel.Simulation",
	"https://www.xbox.com/en-US/games/browse/DynamicChannel.Sports",
	"https://www.xbox.com/en-US/games/browse/DynamicChannel.FamilyKids",
	"https://www.xbox.com/en-US/games/browse/DynamicChannel.Classics",
	"https://www.xbox.com/en-US/games/browse/DynamicChannel.IndieGames",
	"https://www.xbox.com/en-US/games/browse/DynamicChannel.Deals",
}

var xboxSearchQueries = []string{
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
	"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"game", "pro", "the", "war", "nba", "fifa", "call", "assassin",
}

type CatalogParserService struct {
	repo         *repository.CatalogImportRepo
	settingsRepo *repository.SettingsRepo
	client       *http.Client

	mu             sync.Mutex
	status         CatalogParserStatus
	cancel         context.CancelFunc
	useFreshUpsert bool
	lightImport    bool

	proxyClientOnce sync.Once
	proxyClientVal  *http.Client
}

func (s *CatalogParserService) proxyClient(ctx context.Context) *http.Client {
	s.proxyClientOnce.Do(func() {
		proxyStr := os.Getenv("HTTP_PROXY")
		if proxyStr == "" {
			proxyStr = os.Getenv("HTTPS_PROXY")
		}
		if proxyStr == "" {
			proxyStr = os.Getenv("ALL_PROXY")
		}
		if s.settingsRepo != nil && proxyStr == "" {
			if setting, err := s.settingsRepo.Get(ctx, "ps_store_proxy"); err == nil {
				if v, ok := setting["value"].(string); ok && v != "" {
					proxyStr = v
				}
			}
		}
		if proxyStr != "" {
			proxyURL, err := url.Parse(proxyStr)
			if err != nil {
				slog.Warn("Неверный URL прокси", "proxy", proxyStr, "error", err)
				s.proxyClientVal = s.client
				return
			}
			if proxyURL.Scheme == "socks5" {
				dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
				if err != nil {
					slog.Warn("Не удалось создать SOCKS5 dialer", "proxy", proxyStr, "error", err)
					s.proxyClientVal = s.client
					return
				}
				cd, ok := dialer.(proxy.ContextDialer)
				if !ok {
					slog.Warn("SOCKS5 dialer не поддерживает ContextDialer")
					s.proxyClientVal = s.client
					return
				}
				s.proxyClientVal = &http.Client{
					Timeout: 25 * time.Second,
					Transport: &http.Transport{
						DialContext:       cd.DialContext,
						DisableKeepAlives: true,
					},
				}
				return
			}
			s.proxyClientVal = &http.Client{
				Timeout: 25 * time.Second,
				Transport: &http.Transport{
					Proxy: func(*http.Request) (*url.URL, error) { return proxyURL, nil },
				},
			}
			return
		}
		s.proxyClientVal = s.client
	})
	return s.proxyClientVal
}

type CatalogParserStatus struct {
	Running          bool       `json:"running"`
	Full             bool       `json:"full"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	FinishedAt       *time.Time `json:"finished_at,omitempty"`
	Imported         int        `json:"imported"`
	Errors           []string   `json:"errors"`
	Sources          []string   `json:"sources"`
	CurrentSource    string     `json:"current_source"`
	CurrentStage     string     `json:"current_stage"`
	Processed        int        `json:"processed"`
	Total            int        `json:"total"`
	Percent          float64    `json:"percent"`
	EstimatedSeconds *int       `json:"estimated_seconds,omitempty"`
}

func NewCatalogParserService(repo *repository.CatalogImportRepo, settingsRepo *repository.SettingsRepo) *CatalogParserService {
	slog.Info("HTTP client created - прямой, без глобального прокси. Для PS Store используйте proxyClient()")
	return &CatalogParserService{
		repo:         repo,
		settingsRepo: settingsRepo,
		client: &http.Client{
			Timeout: 25 * time.Second,
		},
		status: CatalogParserStatus{
			Sources: []string{domain.CatalogSourcePSStore, domain.CatalogSourceXboxStore},
		},
	}
}

func (s *CatalogParserService) Status() CatalogParserStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

func (s *CatalogParserService) RunAsync(ctx context.Context, full bool) error {
	s.mu.Lock()
	if s.status.Running {
		s.mu.Unlock()
		return errors.New("catalog parser is already running")
	}
	now := time.Now()
	s.status = CatalogParserStatus{
		Running:       true,
		Full:          full,
		StartedAt:     &now,
		Sources:       []string{domain.CatalogSourcePSStore, domain.CatalogSourceXboxStore},
		CurrentSource: "PS Store",
		CurrentStage:  "Подготовка импорта",
	}
	s.mu.Unlock()

	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	s.cancel = cancel

	go s.run(runCtx, full)
	return nil
}

func (s *CatalogParserService) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	now := time.Now()
	s.cancel = nil
	s.status.Running = false
	s.status.FinishedAt = &now
	s.status.CurrentStage = "Остановлено"
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (s *CatalogParserService) RunPSImportSync(ctx context.Context, fresh bool) (int, error) {
	ensureTRYRUBRate(ctx, s.client)
	s.useFreshUpsert = fresh
	s.lightImport = false
	defer func() { s.useFreshUpsert = false; s.lightImport = false }()
	return s.importPSStore(ctx, true)
}

func (s *CatalogParserService) RunPSImportLight(ctx context.Context) (int, error) {
	s.lightImport = true
	defer func() { s.lightImport = false }()
	return s.importPSStore(ctx, false)
}

func (s *CatalogParserService) RunXboxImportSync(ctx context.Context, fresh bool) (int, error) {
	ensureTRYRUBRate(ctx, s.client)
	s.useFreshUpsert = fresh
	defer func() { s.useFreshUpsert = false }()
	return s.importXboxStore(ctx, true)
}

func (s *CatalogParserService) RunImportSync(ctx context.Context, fresh bool) (int, error) {
	ensureTRYRUBRate(ctx, s.client)
	s.useFreshUpsert = fresh
	defer func() { s.useFreshUpsert = false }()
	imported := 0
	psCount, err := s.importPSStore(ctx, true)
	if err != nil {
		return imported, err
	}
	imported += psCount
	xboxCount, err := s.importXboxStore(ctx, true)
	if err != nil {
		return imported, err
	}
	imported += xboxCount
	return imported, nil
}

func (s *CatalogParserService) upsertImport(ctx context.Context, item *domain.CatalogImport) error {
	if s.useFreshUpsert {
		return s.repo.UpsertFresh(ctx, item)
	}
	return s.repo.UpsertPending(ctx, item)
}

func (s *CatalogParserService) run(ctx context.Context, full bool) {
	ensureTRYRUBRate(ctx, s.client)
	imported := 0
	errs := []string{}

	s.lightImport = true
	psCount, err := s.importPSStore(ctx, full)
	s.lightImport = false
	if err != nil {
		errs = append(errs, "PS Store: "+err.Error())
		slog.Warn("PS Store import failed", slog.String("error", err.Error()))
	}
	imported += psCount

	xboxCount, err := s.importXboxStore(ctx, full)
	if err != nil {
		errs = append(errs, "Xbox Store: "+err.Error())
		slog.Warn("Xbox Store import failed", slog.String("error", err.Error()))
	}
	imported += xboxCount

	finished := time.Now()
	s.mu.Lock()
	s.status.Running = false
	s.cancel = nil
	s.status.FinishedAt = &finished
	s.status.Imported = imported
	s.status.Errors = errs
	if len(errs) > 0 {
		s.status.CurrentStage = "Завершено с ошибками"
	} else {
		s.status.CurrentStage = "Завершено"
	}
	s.status.Percent = 100
	s.status.EstimatedSeconds = nil
	s.mu.Unlock()
}

func (s *CatalogParserService) updateProgress(source, stage string, processed, total, imported int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status.CurrentSource = source
	s.status.CurrentStage = stage
	s.status.Processed = processed
	s.status.Total = total
	s.status.Imported = imported
	if total > 0 {
		s.status.Percent = float64(processed) / float64(total) * 100
		if s.status.Percent > 100 {
			s.status.Percent = 100
		}
	} else {
		s.status.Percent = 0
	}

	if s.status.StartedAt != nil && processed > 0 && total > processed {
		elapsed := time.Since(*s.status.StartedAt).Seconds()
		remaining := int((elapsed / float64(processed)) * float64(total-processed))
		s.status.EstimatedSeconds = &remaining
	} else {
		s.status.EstimatedSeconds = nil
	}
}

type psStoreResponse struct {
	Links        []psStoreItem `json:"links"`
	TotalResults int           `json:"total_results"`
}

type psStoreItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"long_desc"`
	ProviderName string `json:"provider_name"`
	ReleaseDate  string `json:"release_date"`
	Images       []struct {
		URL  string      `json:"url"`
		Type interface{} `json:"type"`
	} `json:"images"`
	DefaultSku *struct {
		DisplayPrice string `json:"display_price"`
		Price        int    `json:"price"`
	} `json:"default_sku"`
	Skus []struct {
		Entitlements []struct {
			PreOrderFlag bool `json:"preorder_flag"`
		} `json:"entitlements"`
	} `json:"skus"`
	PlayablePlatform []string `json:"playable_platform"`
	TopCategory      string   `json:"top_category"`
	GameContentTypes []struct {
		Key string `json:"key"`
	} `json:"gameContentTypesList"`
}

func (s *CatalogParserService) importPSStore(ctx context.Context, _ bool) (int, error) {
	const pageSize = 30
	imported := 0
	seen := map[string]bool{}

	for _, container := range psStoreContainers {
		total := 0
		for start := 0; ; start += pageSize {
			if err := ctx.Err(); err != nil {
				return imported, err
			}
			s.updateProgress("PS Store", "Каталог "+container, start, total, imported)
			endpoint := fmt.Sprintf(
				"https://store.playstation.com/store/api/chihiro/00_09_000/container/%s/%s/19/%s?size=%d&start=%d",
				psStoreRegion, psStoreLocale, container, pageSize, start,
			)
			var payload psStoreResponse
			if err := s.getJSONWithClient(ctx, s.proxyClient(ctx), endpoint, &payload); err != nil {
				slog.Warn("PS Store API недоступен. Возможно, нужен VPN/прокси — установите HTTP_PROXY или укажите прокси в Настройках админки.",
					slog.String("container", container),
					slog.String("error", err.Error()),
				)
				break
			}
			if payload.TotalResults > 0 {
				total = payload.TotalResults
			}
			if len(payload.Links) == 0 {
				break
			}
			for _, game := range payload.Links {
				if err := ctx.Err(); err != nil {
					return imported, err
				}
				if game.ID == "" || strings.TrimSpace(game.Name) == "" || seen[game.ID] {
					continue
				}
				seen[game.ID] = true
				if !s.lightImport && (game.DefaultSku == nil || game.DefaultSku.Price <= 0) {
					if detailed, err := s.fetchPSStoreItem(ctx, game.ID); err == nil {
						game = detailed
					}
				}
				item := s.catalogImportFromPS(ctx, game)
				if item == nil {
					continue
				}
				if err := s.upsertImport(ctx, item); err != nil && !errors.Is(err, sql.ErrNoRows) {
					return imported, err
				}
				imported++
			}
			if payload.TotalResults > 0 && start+pageSize >= payload.TotalResults {
				break
			}
		}
	}
	return imported, nil
}

func (s *CatalogParserService) catalogImportFromPS(ctx context.Context, game psStoreItem) *domain.CatalogImport {
	title := strings.TrimSpace(game.Name)
	if title == "" || IsNonGameStoreItem(title) {
		return nil
	}
	if enGame, err := s.fetchPSStoreItemEN(ctx, game.ID); err == nil {
		if enTitle := strings.TrimSpace(enGame.Name); enTitle != "" && !ContainsTurkishLetters(enTitle) {
			title = enTitle
		}
	}
	descText := cleanDescription(game.Description)
	if !s.lightImport {
		if ru := s.fetchPSDescriptionRU(ctx, game.ID); ru != "" {
			descText = ru
		}
	}
	description := nullableString(descText)
	imageURL := firstPSImage(game.Images)
	platforms := psPlatforms(game.PlayablePlatform)
	priceTR, _ := psStorePrice(game)
	var priceUA *float64
	if !s.lightImport {
		priceUA = s.fetchPSStorePriceUA(ctx, game.ID)
	}

	// У предзаказов и части новинок старый API молчит про цену, хотя на сайте
	// магазина она есть. Дочитываем со страницы товара, иначе игра не попадёт
	// в каталог вовсе.
	pageRelease := time.Time{}
	if priceTR == nil && !s.lightImport {
		priceTR, pageRelease = s.psPageTurkeyPrice(ctx, game.ID)
	}
	if priceUA == nil && !s.lightImport {
		priceUA = s.psPageUkrainePrice(ctx, game.ID)
	}
	display := ""
	if game.DefaultSku != nil {
		display = game.DefaultSku.DisplayPrice
	}
	raw, _ := json.Marshal(game)
	releaseAt := parseReleaseDate(game.ReleaseDate)
	if releaseAt.IsZero() && !pageRelease.IsZero() {
		releaseAt = pageRelease
	}
	isPreorder := psIsPreorder(game, releaseAt)

	// Build prices JSON: Turkey nominal, Ukraine UAH
	prices := map[string]float64{}
	if priceTR != nil {
		prices["tr"] = *priceTR
	}
	if priceUA != nil {
		prices["ua"] = *priceUA
	}
	var bestPrice *float64
	for _, p := range prices {
		if bestPrice == nil || p < *bestPrice {
			bestPrice = &p
		}
	}
	pricesRaw, _ := json.Marshal(prices)

	sellable := false
	if bestPrice != nil && !IsFreePrice(bestPrice, display) && *bestPrice >= MinPaidPriceRUB() {
		sellable = true
	}
	if isPreorder || (!releaseAt.IsZero() && releaseAt.After(time.Now().UTC())) {
		sellable = true
	}
	if !sellable {
		return nil
	}
	section, year := classifyImportMetadata(releaseAt, isPreorder)
	titleKey := NormalizeGameTitle(strings.TrimSpace(game.Name))
	platformFamily := PlatformFamilyFromImport(domain.CatalogSourcePSStore, platforms)

	return &domain.CatalogImport{
		ExternalID:       game.ID,
		Source:           domain.CatalogSourcePSStore,
		Title:            title,
		TitleKey:         titleKey,
		PlatformFamily:   platformFamily,
		Description:      description,
		ImageURL:         imageURL,
		Platforms:        platforms,
		GameSection:      section,
		ReleaseYear:      year,
		ReleaseDate:      timePtr(releaseAt),
		Publisher:        strings.TrimSpace(game.ProviderName),
		OriginalPriceRUB: bestPrice,
		OriginalCurrency: nullableString("RUB"),
		Prices:           pricesRaw,
		Raw:              raw,
		Status:           domain.CatalogImportPending,
	}
}

func firstPSImage(images []struct {
	URL  string      `json:"url"`
	Type interface{} `json:"type"`
}) *string {
	for _, image := range images {
		imageType := strings.ToLower(fmt.Sprint(image.Type))
		if image.URL != "" && strings.Contains(imageType, "cover") {
			return nullableString(image.URL)
		}
	}
	for _, image := range images {
		if image.URL != "" {
			return nullableString(image.URL)
		}
	}
	return nil
}

func psPlatforms(values []string) pq.StringArray {
	platforms := map[string]bool{}
	for _, value := range values {
		normalized := strings.ToLower(value)
		if strings.Contains(normalized, "ps5") {
			platforms["ps5"] = true
		}
		if strings.Contains(normalized, "ps4") {
			platforms["ps4"] = true
		}
	}
	if len(platforms) == 0 {
		platforms["ps5"] = true
	}
	return mapKeys(platforms)
}

func psStorePrice(game psStoreItem) (*float64, *string) {
	if game.DefaultSku == nil {
		return nil, nil
	}
	var tryAmount float64
	if game.DefaultSku.Price > 0 {
		tryAmount = float64(game.DefaultSku.Price) / 100.0
	} else if parsed := parseDisplayPriceTRY(game.DefaultSku.DisplayPrice); parsed > 0 {
		tryAmount = parsed
	} else {
		return nil, nil
	}
	rub := TurkeyNominalPrice(tryAmount)
	return &rub, nullableString("TRY")
}

func psStorePriceUA(game psStoreItem) *float64 {
	if game.DefaultSku == nil {
		return nil
	}
	var uahAmount float64
	if game.DefaultSku.Price > 0 {
		uahAmount = float64(game.DefaultSku.Price) / 100.0
	} else if parsed := parseDisplayPriceUAH(game.DefaultSku.DisplayPrice); parsed > 0 {
		uahAmount = parsed
	} else {
		return nil
	}
	rub := UkrainePrice(uahAmount)
	return &rub
}

var xboxProductIDPattern = regexp.MustCompile(`9[A-Z0-9]{11}`)

func (s *CatalogParserService) collectXboxProductIDs(body string, seen map[string]bool, ids *[]string) int {
	added := 0
	for _, id := range xboxProductIDPattern.FindAllString(body, -1) {
		if seen[id] {
			continue
		}
		seen[id] = true
		if ids != nil {
			*ids = append(*ids, id)
		}
		added++
	}
	return added
}

func (s *CatalogParserService) collectXboxSearchIDs(ctx context.Context, queries []string, seen map[string]bool) int {
	type searchJob struct {
		query string
		page  int
	}
	jobs := make([]searchJob, 0, len(queries)*2)
	for _, query := range queries {
		jobs = append(jobs, searchJob{query: query, page: 1})
		jobs = append(jobs, searchJob{query: query, page: 2})
	}

	var mu sync.Mutex
	added := 0
	sem := make(chan struct{}, 6)
	var wg sync.WaitGroup

	for i, job := range jobs {
		if err := ctx.Err(); err != nil {
			break
		}
		if i%40 == 0 {
			s.updateProgress("Xbox Store", "Поиск игр на Xbox.com TR", i, len(jobs), added)
		}
		wg.Add(1)
		go func(j searchJob) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			endpoint := fmt.Sprintf(
				"https://www.xbox.com/en-US/search/results?q=%s",
				url.QueryEscape(j.query),
			)
			if j.page > 1 {
				endpoint += fmt.Sprintf("&page=%d", j.page)
			}
			body, err := s.getTextWithRetry(ctx, endpoint, 2)
			if err != nil {
				return
			}
			mu.Lock()
			added += s.collectXboxProductIDs(body, seen, nil)
			mu.Unlock()
		}(job)
	}
	wg.Wait()
	return added
}

func (s *CatalogParserService) importXboxStore(ctx context.Context, _ bool) (int, error) {
	ids := []string{}
	seen := map[string]bool{}

	queries := buildXboxSearchQueries()
	s.collectXboxSearchIDs(ctx, queries, seen)
	for id := range seen {
		ids = append(ids, id)
	}

	for i, endpoint := range xboxChannelPages {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		s.updateProgress("Xbox Store", "Каналы Xbox.com US", i, len(xboxChannelPages), len(ids))
		body, err := s.getText(ctx, endpoint)
		if err != nil {
			slog.Warn("xbox channel failed", slog.String("url", endpoint), slog.String("error", err.Error()))
			continue
		}
		s.collectXboxProductIDs(body, seen, &ids)
	}

	for _, template := range xboxBrowseTemplates {
		for page := 1; page <= 5; page++ {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
			s.updateProgress("Xbox Store", "Каталог Xbox.com US", page-1, 5, len(ids))
			endpoint := fmt.Sprintf(template, page)
			body, err := s.getText(ctx, endpoint)
			if err != nil {
				slog.Warn("xbox browse page failed", slog.String("url", endpoint), slog.String("error", err.Error()))
				break
			}
			if s.collectXboxProductIDs(body, seen, &ids) == 0 && page > 1 {
				break
			}
		}
	}

	slog.Info("xbox product ids collected", slog.Int("count", len(ids)))

	imported := 0
	for start := 0; start < len(ids); start += 20 {
		if err := ctx.Err(); err != nil {
			return imported, err
		}
		end := start + 20
		if end > len(ids) {
			end = len(ids)
		}
		s.updateProgress("Xbox Store", "Загружаем детали и цены", start, len(ids), imported)
		items, err := s.fetchXboxDetails(ctx, ids[start:end])
		if err != nil {
			slog.Warn("xbox details batch failed",
				slog.Int("start", start),
				slog.Int("end", end),
				slog.String("error", err.Error()),
			)
			time.Sleep(2 * time.Second)
			continue
		}
		for _, item := range items {
			if err := ctx.Err(); err != nil {
				return imported, err
			}
			if err := s.upsertImport(ctx, item); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return imported, err
			}
			imported++
		}
		s.updateProgress("Xbox Store", "Сохраняем игры в очередь", end, len(ids), imported)
	}

	return imported, nil
}

type xboxCatalogResponse struct {
	Products []xboxProduct `json:"Products"`
}

type xboxLocalizedProperty struct {
	Language         string `json:"Language"`
	ProductTitle     string `json:"ProductTitle"`
	PublisherName    string `json:"PublisherName"`
	ShortDescription string `json:"ShortDescription"`
	Images           []struct {
		Uri          string `json:"Uri"`
		ImagePurpose string `json:"ImagePurpose"`
	} `json:"Images"`
}

type xboxProduct struct {
	ProductID        string `json:"ProductId"`
	AllowedPlatforms []struct {
		PlatformName string `json:"PlatformName"`
	} `json:"AllowedPlatforms"`
	Properties struct {
		Attributes []struct {
			Name  string `json:"Name"`
			Value string `json:"Value"`
		} `json:"Attributes"`
	} `json:"Properties"`
	MarketProperties []struct {
		OriginalReleaseDate string `json:"OriginalReleaseDate"`
	} `json:"MarketProperties"`
	LocalizedProperties      []xboxLocalizedProperty `json:"LocalizedProperties"`
	DisplaySkuAvailabilities []struct {
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
	} `json:"DisplaySkuAvailabilities"`
}

func (s *CatalogParserService) fetchXboxDetails(ctx context.Context, ids []string) ([]*domain.CatalogImport, error) {
	products, err := s.fetchXboxProducts(ctx, ids)
	if err != nil {
		return nil, err
	}
	items := make([]*domain.CatalogImport, 0, len(products))
	for _, product := range products {
		item := catalogImportFromXbox(product)
		if item != nil {
			items = append(items, item)
		}
	}
	return items, nil
}

func catalogImportFromXbox(product xboxProduct) *domain.CatalogImport {
	if product.ProductID == "" || len(product.LocalizedProperties) == 0 {
		return nil
	}
	localized := pickXboxLocalized(product.LocalizedProperties)
	if strings.TrimSpace(localized.ProductTitle) == "" {
		return nil
	}

	price, currency := xboxPrice(product)
	raw, _ := json.Marshal(product)
	releaseAt, isPreorder := xboxReleaseMeta(product)
	section, year := classifyImportMetadata(releaseAt, isPreorder)
	platforms := xboxPlatformsFromProduct(product, raw)
	title := strings.TrimSpace(localized.ProductTitle)
	if IsNonGameStoreItem(title) {
		return nil
	}
	prices := map[string]float64{}
	if price != nil {
		prices["xbox"] = *price
	}
	pricesRaw, _ := json.Marshal(prices)
	item := &domain.CatalogImport{
		ExternalID:       product.ProductID,
		Source:           domain.CatalogSourceXboxStore,
		Title:            title,
		TitleKey:         NormalizeGameTitle(title),
		PlatformFamily:   PlatformFamilyFromImport(domain.CatalogSourceXboxStore, platforms),
		Description:      nullableString(localized.ShortDescription),
		ImageURL:         xboxImage(localized.Images),
		Platforms:        platforms,
		GameSection:      section,
		ReleaseYear:      year,
		ReleaseDate:      timePtr(releaseAt),
		Publisher:        strings.TrimSpace(localized.PublisherName),
		OriginalPriceRUB: price,
		OriginalCurrency: currency,
		Prices:           pricesRaw,
		Raw:              raw,
		Status:           domain.CatalogImportPending,
	}
	if !IsSellableCatalogItem(item) {
		return nil
	}
	return item
}

func xboxPlatformsFromProduct(product xboxProduct, raw []byte) pq.StringArray {
	platforms := map[string]bool{}
	for _, p := range product.AllowedPlatforms {
		name := strings.ToLower(p.PlatformName)
		if strings.Contains(name, "xbox") {
			platforms["xbox"] = true
		}
		if strings.Contains(name, "pc") || strings.Contains(name, "windows") {
			platforms["pc"] = true
		}
	}
	for _, attr := range product.Properties.Attributes {
		text := strings.ToLower(attr.Name + " " + attr.Value)
		if strings.Contains(text, "playanywhere") || strings.Contains(text, "xboxplayanywhere") {
			platforms["xbox"] = true
			platforms["pc"] = true
		}
	}
	rawText := strings.ToLower(string(raw))
	if strings.Contains(rawText, "xbox") {
		platforms["xbox"] = true
	}
	if strings.Contains(rawText, "windows.desktop") || strings.Contains(rawText, "win32") ||
		strings.Contains(rawText, "xboxplayanywhere") || strings.Contains(rawText, `"pc"`) {
		platforms["pc"] = true
	}
	if len(platforms) == 0 {
		platforms["xbox"] = true
	}
	return mapKeys(platforms)
}

func xboxPrice(product xboxProduct) (*float64, *string) {
	// Only USD from US market — no TRY
	var bestUSD float64
	for _, sku := range product.DisplaySkuAvailabilities {
		for _, availability := range sku.Availabilities {
			if availability.OrderManagementData == nil {
				continue
			}
			p := availability.OrderManagementData.Price
			price := p.ListPrice
			if price <= 0 {
				price = p.MSRP
			}
			if price <= 0 {
				continue
			}
			if strings.ToUpper(strings.TrimSpace(p.CurrencyCode)) == "USD" && price > bestUSD {
				bestUSD = price
			}
		}
	}
	if bestUSD > 0 {
		rub := XboxUSAPrice(bestUSD)
		return &rub, nullableString("USD")
	}
	return nil, nil
}

func xboxImage(images []struct {
	Uri          string `json:"Uri"`
	ImagePurpose string `json:"ImagePurpose"`
}) *string {
	for _, image := range images {
		if image.Uri == "" {
			continue
		}
		uri := image.Uri
		if strings.HasPrefix(uri, "//") {
			uri = "https:" + uri
		}
		if strings.HasPrefix(uri, "http") && strings.Contains(strings.ToLower(image.ImagePurpose), "poster") {
			return &uri
		}
	}
	for _, image := range images {
		if image.Uri == "" {
			continue
		}
		uri := image.Uri
		if strings.HasPrefix(uri, "//") {
			uri = "https:" + uri
		}
		if strings.HasPrefix(uri, "http") {
			return &uri
		}
	}
	return nil
}

func (s *CatalogParserService) getJSON(ctx context.Context, endpoint string, target interface{}) error {
	return s.getJSONWithClient(ctx, s.client, endpoint, target)
}

func (s *CatalogParserService) getJSONWithClient(ctx context.Context, client *http.Client, endpoint string, target interface{}) error {
	body, err := s.getTextWithClient(ctx, client, endpoint)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(body), target)
}

func (s *CatalogParserService) getTextWithRetry(ctx context.Context, endpoint string, attempts int) (string, error) {
	if attempts <= 0 {
		attempts = 1
	}
	var last error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		body, err := s.getText(ctx, endpoint)
		if err == nil {
			return body, nil
		}
		last = err
		time.Sleep(time.Duration(i+1) * 700 * time.Millisecond)
	}
	return "", last
}

func (s *CatalogParserService) getText(ctx context.Context, endpoint string) (string, error) {
	return s.getTextWithClient(ctx, s.client, endpoint)
}

func (s *CatalogParserService) getTextWithClient(ctx context.Context, client *http.Client, endpoint string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json,text/html;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status %d for %s", resp.StatusCode, endpoint)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func nullableString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func mapKeys(values map[string]bool) pq.StringArray {
	out := make(pq.StringArray, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	return out
}

func isValidReleaseDate(value time.Time) bool {
	if value.IsZero() {
		return false
	}
	year := value.Year()
	return year >= 1990 && year <= 2035
}

func parseReleaseDate(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			parsed = parsed.UTC()
			if !isValidReleaseDate(parsed) {
				return time.Time{}
			}
			return parsed
		}
	}
	return time.Time{}
}

func psIsPreorder(game psStoreItem, releaseAt time.Time) bool {
	for _, sku := range game.Skus {
		for _, entitlement := range sku.Entitlements {
			if entitlement.PreOrderFlag {
				return true
			}
		}
	}
	return !releaseAt.IsZero() && releaseAt.After(time.Now().UTC())
}

func xboxReleaseMeta(product xboxProduct) (time.Time, bool) {
	isPreorder := false
	for _, sku := range product.DisplaySkuAvailabilities {
		if sku.Sku.Properties.IsPreOrder {
			isPreorder = true
		}
		for _, market := range sku.Sku.MarketProperties {
			if parsed := parseReleaseDate(market.FirstAvailableDate); !parsed.IsZero() {
				if parsed.After(time.Now().UTC()) {
					isPreorder = true
				}
				return parsed, isPreorder
			}
		}
	}
	if len(product.MarketProperties) > 0 {
		return parseReleaseDate(product.MarketProperties[0].OriginalReleaseDate), isPreorder
	}
	return time.Time{}, isPreorder
}

func classifyImportMetadata(releaseAt time.Time, isPreorder bool) (string, *int) {
	now := time.Now().UTC()
	if isPreorder || (!releaseAt.IsZero() && releaseAt.After(now)) {
		return "preorder", releaseYear(releaseAt)
	}
	if !releaseAt.IsZero() && !releaseAt.Before(now.AddDate(0, 0, -90)) && !releaseAt.After(now) {
		return "new", releaseYear(releaseAt)
	}
	return "game", releaseYear(releaseAt)
}

func releaseYear(releaseAt time.Time) *int {
	if releaseAt.IsZero() {
		return nil
	}
	year := releaseAt.Year()
	return &year
}

func timePtr(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	v := value.UTC()
	return &v
}

func (s *CatalogParserService) fetchPSGame(ctx context.Context, externalID string) (*domain.CatalogImport, error) {
	game, err := s.fetchPSStoreItem(ctx, externalID)
	if err != nil {
		return nil, err
	}
	return s.catalogImportFromPS(ctx, game), nil
}

func (s *CatalogParserService) fetchPSStoreItem(ctx context.Context, externalID string) (psStoreItem, error) {
	endpoint := fmt.Sprintf(
		"https://store.playstation.com/store/api/chihiro/00_09_000/container/%s/%s/19/%s/0",
		psStoreRegion, psStoreLocale, url.PathEscape(externalID),
	)
	proxyCli := s.proxyClient(ctx)
	var game psStoreItem
	if err := s.getJSONWithClient(ctx, proxyCli, endpoint, &game); err != nil {
		return game, err
	}
	if game.ID == "" {
		game.ID = externalID
	}
	if strings.TrimSpace(game.Name) == "" {
		return game, fmt.Errorf("empty ps game payload for %s", externalID)
	}
	return game, nil
}

func (s *CatalogParserService) fetchPSStoreItemEN(ctx context.Context, externalID string) (psStoreItem, error) {
	endpoint := fmt.Sprintf(
		"https://store.playstation.com/store/api/chihiro/00_09_000/container/US/en/19/%s/0",
		url.PathEscape(externalID),
	)
	proxyCli := s.proxyClient(ctx)
	var game psStoreItem
	if err := s.getJSONWithClient(ctx, proxyCli, endpoint, &game); err != nil {
		return game, err
	}
	if game.ID == "" {
		game.ID = externalID
	}
	return game, nil
}

func (s *CatalogParserService) fetchPSStorePriceUA(ctx context.Context, externalID string) *float64 {
	proxyCli := s.proxyClient(ctx)
	endpoint := fmt.Sprintf(
		"https://store.playstation.com/store/api/chihiro/00_09_000/container/%s/%s/19/%s/0",
		psStoreRegionUA, psStoreLocaleUA, url.PathEscape(externalID),
	)
	var game psStoreItem
	if err := s.getJSONWithClient(ctx, proxyCli, endpoint, &game); err != nil {
		return nil
	}
	if game.DefaultSku == nil || game.DefaultSku.Price <= 0 {
		if detailed, err := s.fetchPSStoreItemUA(ctx, externalID); err == nil {
			game = detailed
		}
	}
	return psStorePriceUA(game)
}

func (s *CatalogParserService) fetchPSStoreItemUA(ctx context.Context, externalID string) (psStoreItem, error) {
	proxyCli := s.proxyClient(ctx)
	endpoint := fmt.Sprintf(
		"https://store.playstation.com/store/api/chihiro/00_09_000/container/%s/%s/19/%s/0",
		psStoreRegionUA, psStoreLocaleUA, url.PathEscape(externalID),
	)
	var game psStoreItem
	if err := s.getJSONWithClient(ctx, proxyCli, endpoint, &game); err != nil {
		return game, err
	}
	if game.ID == "" {
		game.ID = externalID
	}
	if strings.TrimSpace(game.Name) == "" {
		return game, fmt.Errorf("empty ps game payload for %s", externalID)
	}
	return game, nil
}

func psImportMetadataFromPS(game psStoreItem) *domain.CatalogImport {
	title := strings.TrimSpace(game.Name)
	if title == "" {
		return nil
	}
	raw, _ := json.Marshal(game)
	releaseAt := parseReleaseDate(game.ReleaseDate)
	isPreorder := psIsPreorder(game, releaseAt)
	section, year := classifyImportMetadata(releaseAt, isPreorder)
	priceTR, _ := psStorePrice(game)

	prices := map[string]float64{}
	if priceTR != nil {
		prices["tr"] = *priceTR
	}
	pricesRaw, _ := json.Marshal(prices)

	return &domain.CatalogImport{
		ExternalID:       game.ID,
		Source:           domain.CatalogSourcePSStore,
		Title:            title,
		TitleKey:         NormalizeGameTitle(title),
		PlatformFamily:   PlatformFamilyFromImport(domain.CatalogSourcePSStore, psPlatforms(game.PlayablePlatform)),
		GameSection:      section,
		ReleaseYear:      year,
		ReleaseDate:      timePtr(releaseAt),
		Publisher:        strings.TrimSpace(game.ProviderName),
		OriginalPriceRUB: priceTR,
		OriginalCurrency: strPtr("TRY"),
		Raw:              raw,
		Prices:           pricesRaw,
	}
}

// EnrichImportMetadata refreshes prices/dates from store APIs for relevant imports.
func (s *CatalogParserService) EnrichImportMetadata(ctx context.Context) (int, error) {
	return s.EnrichAllImports(ctx)
}

func (s *CatalogParserService) EnrichAllImports(ctx context.Context) (int, error) {
	ensureTRYRUBRate(ctx, s.client)
	total := 0
	psUpdated, err := s.enrichPSMetadata(ctx)
	if err != nil {
		return total, err
	}
	total += psUpdated
	xboxUpdated, err := s.enrichXboxMetadata(ctx)
	if err != nil {
		return total, err
	}
	total += xboxUpdated
	for pass := 0; pass < 5; pass++ {
		n, err := s.EnrichFocusedImports(ctx, 5000)
		if err != nil {
			return total, err
		}
		if n == 0 {
			break
		}
		total += n
	}
	return total, nil
}

func (s *CatalogParserService) EnrichFocusedImports(ctx context.Context, limit int) (int, error) {
	psIDs, xboxIDs, err := s.repo.ListExternalIDsForEnrichment(ctx, limit)
	if err != nil {
		return 0, err
	}
	updated := 0
	psUpdated, err := s.enrichPSIDs(ctx, psIDs)
	if err != nil {
		return updated, err
	}
	updated += psUpdated
	xboxUpdated, err := s.enrichXboxIDs(ctx, xboxIDs)
	if err != nil {
		return updated, err
	}
	updated += xboxUpdated
	return updated, nil
}

func (s *CatalogParserService) enrichPSMetadata(ctx context.Context) (int, error) {
	ids, err := s.repo.ListExternalIDsBySource(ctx, domain.CatalogSourcePSStore, 100000)
	if err != nil {
		return 0, err
	}
	return s.enrichPSIDs(ctx, ids)
}

func (s *CatalogParserService) enrichPSIDs(ctx context.Context, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	var updated atomic.Int32
	var failed atomic.Int32
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup

	for _, externalID := range ids {
		if err := ctx.Err(); err != nil {
			break
		}
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			game, err := s.fetchPSStoreItem(ctx, id)
			if err != nil {
				failed.Add(1)
				return
			}
			item := psImportMetadataFromPS(game)
			if item == nil {
				failed.Add(1)
				return
			}
			ok, err := s.repo.UpdateImportMetadata(ctx, item)
			if err == nil && ok {
				updated.Add(1)
			}
		}(externalID)
	}
	wg.Wait()

	slog.Info("PS metadata enrichment finished",
		slog.Int("updated", int(updated.Load())),
		slog.Int("failed", int(failed.Load())),
		slog.Int("total", len(ids)),
	)
	return int(updated.Load()), nil
}

func (s *CatalogParserService) EnrichXboxOnly(ctx context.Context) (int, error) {
	ensureTRYRUBRate(ctx, s.client)
	return s.enrichXboxMetadata(ctx)
}

func (s *CatalogParserService) enrichXboxMetadata(ctx context.Context) (int, error) {
	ids, err := s.repo.ListExternalIDsBySource(ctx, domain.CatalogSourceXboxStore, 100000)
	if err != nil {
		return 0, err
	}
	return s.enrichXboxIDs(ctx, ids)
}

func (s *CatalogParserService) enrichXboxIDs(ctx context.Context, ids []string) (int, error) {
	updated := 0
	for start := 0; start < len(ids); start += 20 {
		if err := ctx.Err(); err != nil {
			return updated, err
		}
		end := start + 20
		if end > len(ids) {
			end = len(ids)
		}
		products, err := s.fetchXboxProducts(ctx, ids[start:end])
		if err != nil {
			return updated, err
		}
		for _, product := range products {
			item := xboxImportMetadataFromXbox(product)
			if item == nil {
				continue
			}
			ok, err := s.repo.UpdateImportMetadata(ctx, item)
			if err != nil {
				continue
			}
			if ok {
				updated++
			}
		}
	}
	return updated, nil
}

func (s *CatalogParserService) fetchXboxProducts(ctx context.Context, ids []string) ([]xboxProduct, error) {
	q := url.Values{}
	q.Set("bigIds", strings.Join(ids, ","))
	q.Set("market", xboxStoreMarket)
	q.Set("languages", xboxDescriptionLangs)
	q.Set("MS-CV", "DGU1mcuYo0WMMp+F.1")
	endpoint := "https://displaycatalog.mp.microsoft.com/v7.0/products?" + q.Encode()

	var payload xboxCatalogResponse
	if err := s.getJSON(ctx, endpoint, &payload); err != nil {
		return nil, err
	}
	return payload.Products, nil
}

func xboxImportMetadataFromXbox(product xboxProduct) *domain.CatalogImport {
	if product.ProductID == "" || len(product.LocalizedProperties) == 0 {
		return nil
	}
	localized := pickXboxLocalized(product.LocalizedProperties)
	title := strings.TrimSpace(localized.ProductTitle)
	if title == "" {
		return nil
	}
	price, currency := xboxPrice(product)
	raw, _ := json.Marshal(product)
	releaseAt, isPreorder := xboxReleaseMeta(product)
	section, year := classifyImportMetadata(releaseAt, isPreorder)
	prices := map[string]float64{}
	if price != nil {
		prices["xbox"] = *price
	}
	pricesRaw, _ := json.Marshal(prices)
	return &domain.CatalogImport{
		ExternalID:       product.ProductID,
		Source:           domain.CatalogSourceXboxStore,
		Title:            title,
		TitleKey:         NormalizeGameTitle(title),
		PlatformFamily:   PlatformFamilyFromImport(domain.CatalogSourceXboxStore, xboxPlatformsFromProduct(product, raw)),
		GameSection:      section,
		ReleaseYear:      year,
		ReleaseDate:      timePtr(releaseAt),
		Publisher:        strings.TrimSpace(localized.PublisherName),
		OriginalPriceRUB: price,
		OriginalCurrency: currency,
		Prices:           pricesRaw,
		Raw:              raw,
	}
}
