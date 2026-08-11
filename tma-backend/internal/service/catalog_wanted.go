package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"tma-backend/internal/domain"
)

// WantedGame — строка списка «что мы хотим продавать». Список задаёт только
// намерение: какие игры и под какие платформы. Цены, описания, картинки,
// даты релиза и раздел всегда берутся из магазинов, а не из списка —
// список стареет, магазин нет.
type WantedGame struct {
	Row   int
	Title string
	PS4   bool
	PS5   bool
	Xbox  bool
	PC    bool
}

func (w WantedGame) WantsPS() bool   { return w.PS4 || w.PS5 }
func (w WantedGame) WantsXbox() bool { return w.Xbox || w.PC }

type WantedMatch struct {
	Row        int    `json:"row"`
	Title      string `json:"title"`
	Source     string `json:"source"`
	ExternalID string `json:"external_id"`
	StoreTitle string `json:"store_title"`
	Section    string `json:"section"`
	Confidence string `json:"confidence"`
}

type WantedImportReport struct {
	Total        int            `json:"total"`
	Imported     int            `json:"imported"`
	MatchedPS    int            `json:"matched_ps"`
	MatchedXbox  int            `json:"matched_xbox"`
	NotFound     []string       `json:"not_found"`
	Matches      []WantedMatch  `json:"matches"`
	StartedAt    time.Time      `json:"started_at"`
	FinishedAt   time.Time      `json:"finished_at"`
	PSIndexSize  int            `json:"ps_index_size"`
	SectionStats map[string]int `json:"section_stats"`
}

var wantedHeaderAliases = map[string][]string{
	"title": {"игра", "название", "game", "title"},
	"ps4":   {"ps4"},
	"ps5":   {"ps5"},
	"xbox":  {"xbox (консоль)", "xbox", "xbox консоль"},
	"pc":    {"xbox (pc)", "pc", "windows"},
}

func matchHeader(header string) string {
	normalized := strings.ToLower(strings.TrimSpace(header))
	for field, aliases := range wantedHeaderAliases {
		for _, alias := range aliases {
			if normalized == alias {
				return field
			}
		}
	}
	return ""
}

func isYes(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "да", "yes", "+", "true", "1", "v":
		return true
	}
	return false
}

// ParseWantedGamesCSV читает выгрузку списка (CSV из Google Sheets).
func ParseWantedGamesCSV(r io.Reader) ([]WantedGame, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("не читается заголовок: %w", err)
	}

	columns := map[string]int{}
	for i, name := range header {
		if field := matchHeader(name); field != "" {
			if _, exists := columns[field]; !exists {
				columns[field] = i
			}
		}
	}
	if _, ok := columns["title"]; !ok {
		return nil, fmt.Errorf("в CSV нет колонки с названием игры")
	}

	cell := func(record []string, field string) string {
		idx, ok := columns[field]
		if !ok || idx >= len(record) {
			return ""
		}
		return record[idx]
	}

	games := make([]WantedGame, 0, 512)
	seen := map[string]bool{}
	row := 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		row++

		title := strings.TrimSpace(cell(record, "title"))
		if title == "" {
			continue
		}
		key := NormalizeGameTitle(title)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true

		games = append(games, WantedGame{
			Row:   row,
			Title: title,
			PS4:   isYes(cell(record, "ps4")),
			PS5:   isYes(cell(record, "ps5")),
			Xbox:  isYes(cell(record, "xbox")),
			PC:    isYes(cell(record, "pc")),
		})
	}
	return games, nil
}

// ─────────────── сопоставление названий ───────────────

var (
	wantedNoiseRe = regexp.MustCompile(`[®™©\x{2122}]`)
	wantedPunctRe = regexp.MustCompile(`[^\p{L}\p{N}]+`)
	wantedRomanRe = regexp.MustCompile(`^(i{1,3}|iv|v|vi{1,3}|ix|x)$`)
	// «Xbox Series X|S» целиком, в любом месте названия: иначе «X» превращается
	// в римскую десятку и ломает ключ.
	wantedSeriesRe = regexp.MustCompile(`(?i)(xbox\s+)?series\s+x\s*\|?\s*s?\b`)
	// Год в скобках — пометка составителя списка, в сторе его нет.
	wantedYearRe = regexp.MustCompile(`\((19|20)\d{2}\)`)
	// Слова издания отдельным токеном: регексп с \b кириллицу не ловит,
	// в RE2 граница слова определяется только по ASCII.
	wantedEditionTokenRe = regexp.MustCompile(`(?i)^(standard|standart|deluxe|ultimate|premium|gold|complete|goty|edition|s[uü]r[uü]m[uü]*|paket\p{L}*|dijital|digital|sayisal|sayısal|pakiet|eksiksiz|koleksiyon\p{L}*|konsol|console|консол\p{L}*|alt[ıi]n|[öo]zel|издани\p{L}*|версия|версии|набор|полн\p{L}*|золот\p{L}*|расширенн\p{L}*|коллекционн\p{L}*|юбилейн\p{L}*|подарочн\p{L}*|делюкс)$`)
	wantedEditionRe      = regexp.MustCompile(`(?i)\b(standard|standart|deluxe|ultimate|premium|gold|complete|goty|cross[- ]?gen|bundle|edition|s[uü]r[uü]m[uü]*|paket(i|leri)?|dijital|digital|sayisal|sayısal|pakiet|eksiksiz|koleksiyon[ual]*|konsol|console|консол\p{L}*|alt[ıi]n|[öo]zel|издани\p{L}*|версия|набор|полн\p{L}*|золот\p{L}*|расширенн\p{L}*|коллекционн\p{L}*|юбилейн\p{L}*|definitive|anniversary|collection)\b`)
)

var romanValues = map[string]string{
	"i": "1", "ii": "2", "iii": "3", "iv": "4", "v": "5",
	"vi": "6", "vii": "7", "viii": "8", "ix": "9", "x": "10",
}

// wantedKey — агрессивная нормализация: убирает издания, локальные слова
// («Sürüm», «Paketi»), пунктуацию и переводит римские цифры в арабские,
// чтобы «Modern Warfare III» и «Modern Warfare 3» сошлись.
func wantedKey(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = wantedNoiseRe.ReplaceAllString(s, " ")
	s = wantedSeriesRe.ReplaceAllString(s, " ")
	// Год в скобках — пометка составителя списка, в сторе его нет:
	// «Call of Duty 4: Modern Warfare (2007)».
	s = wantedYearRe.ReplaceAllString(s, " ")
	// Пунктуацию убираем до слов издания: в RE2 граница слова считается по
	// ASCII, поэтому в «Ultimate-издание» русское слово иначе не находится.
	s = wantedPunctRe.ReplaceAllString(s, " ")
	s = wantedEditionRe.ReplaceAllString(s, " ")

	fields := make([]string, 0, 8)
	for _, token := range strings.Fields(s) {
		if wantedStopTokens[token] || wantedEditionTokenRe.MatchString(token) {
			continue
		}
		// Римские цифры переводим потокенно: по всей строке нельзя, потому что
		// в RE2 граница слова считается по ASCII и «için» превращается в «1çin».
		if value, ok := romanValues[token]; ok && wantedRomanRe.MatchString(token) {
			fields = append(fields, value)
			continue
		}
		// Одиночные буквы — огрызки турецких окончаний («sürümü» → «ü»).
		if runes := []rune(token); len(runes) == 1 && !unicode.IsDigit(runes[0]) {
			continue
		}
		fields = append(fields, token)
	}
	return strings.Join(fields, " ")
}

// Токены, ничего не говорящие о том, какая это игра: платформы, служебные
// предлоги, слово «игра». Без их отсева «Battlefield 2042 для PS5» из списка
// не сходится с «Battlefield 2042» из стора.
var wantedStopTokens = map[string]bool{
	"ps4": true, "ps5": true, "ps": true, "playstation": true,
	"xbox": true, "series": true, "one": true,
	"pc": true, "windows": true, "win": true,
	"для": true, "на": true, "и": true, "the": true, "for": true,
	"ve": true, "and": true, "или": true, "için": true, "icin": true, "версии": true,
	"игра": true, "game": true, "версия": true,
}

func wantedTokens(title string) map[string]bool {
	tokens := map[string]bool{}
	for _, token := range strings.Fields(wantedKey(title)) {
		tokens[token] = true
	}
	return tokens
}

// titleSimilarity — доля общих токенов относительно БОЛЬШЕГО из названий.
// Делить на меньшее нельзя: тогда «Battlefield 6» совпадает с
// «Battlefield 6 Season Pass» на 100%, и карточка игры получает цену DLC.
func titleSimilarity(a, b string) float64 {
	ta, tb := wantedTokens(a), wantedTokens(b)
	if len(ta) == 0 || len(tb) == 0 {
		return 0
	}
	common := 0
	for token := range ta {
		if tb[token] {
			common++
		}
	}
	longer := len(ta)
	if len(tb) > longer {
		longer = len(tb)
	}
	return float64(common) / float64(longer)
}

type titleCandidate struct {
	title string
	value interface{}
}

// pickBestTitleMatch выбирает кандидата консервативно: сначала точное
// совпадение ключей, затем схожесть выше порога. Порог намеренно высокий —
// лучше не завести игру, чем показать покупателю не ту.
func pickBestTitleMatch(wanted string, candidates []titleCandidate) (interface{}, string) {
	target := wantedKey(wanted)
	if target == "" {
		return nil, ""
	}

	for _, candidate := range candidates {
		if wantedKey(candidate.title) == target {
			return candidate.value, "exact"
		}
	}

	bestScore := 0.0
	var best interface{}
	for _, candidate := range candidates {
		score := titleSimilarity(wanted, candidate.title)
		if score > bestScore {
			bestScore = score
			best = candidate.value
		}
	}
	if bestScore >= 0.85 {
		return best, "fuzzy"
	}
	return nil, ""
}

// ─────────────── индекс PS Store ───────────────

type psTitleIndex struct {
	byKey map[string][]psStoreItem
	size  int
}

func (idx *psTitleIndex) add(item psStoreItem) {
	key := wantedKey(item.Name)
	if key == "" {
		return
	}
	idx.byKey[key] = append(idx.byKey[key], item)
	idx.size++
}

// editions возвращает позиции той же игры с суффиксом издания:
// «HELLDIVERS 2» → «HELLDIVERS 2 Super Citizen Sürümü».
func (idx *psTitleIndex) editions(baseKey string) []psStoreItem {
	if baseKey == "" {
		return nil
	}
	prefix := baseKey + " "
	var result []psStoreItem
	for key, items := range idx.byKey {
		if strings.HasPrefix(key, prefix) {
			result = append(result, items...)
		}
	}
	return result
}

func (idx *psTitleIndex) lookup(title string) []psStoreItem {
	key := wantedKey(title)
	if key == "" {
		return nil
	}
	if items, ok := idx.byKey[key]; ok {
		return items
	}

	// Точного ключа нет — ищем по схожести среди ключей с общим первым словом,
	// чтобы не перебирать все 20k позиций на каждую игру.
	target := strings.Fields(key)
	if len(target) == 0 {
		return nil
	}
	var result []psStoreItem
	for candidateKey, items := range idx.byKey {
		if !strings.HasPrefix(candidateKey, target[0]) {
			continue
		}
		if titleSimilarity(key, candidateKey) >= 0.85 {
			result = append(result, items...)
		}
	}
	return result
}

// psIndexCacheTTL — сколько живёт слепок каталога на диске. Полный обход это
// ~50 запросов по 500 позиций; если гонять его на каждый прогон, PS Store
// отвечает 403 и блокирует IP на несколько часов.
const psIndexCacheTTL = 12 * time.Hour

func psIndexCachePath() string {
	return filepath.Join(os.TempDir(), "ps-store-index.json")
}

func loadPSIndexCache() (*psTitleIndex, bool) {
	path := psIndexCachePath()
	info, err := os.Stat(path)
	if err != nil || time.Since(info.ModTime()) > psIndexCacheTTL {
		return nil, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var items []psStoreItem
	if err := json.Unmarshal(data, &items); err != nil || len(items) == 0 {
		return nil, false
	}
	index := &psTitleIndex{byKey: map[string][]psStoreItem{}}
	for _, item := range items {
		index.add(item)
	}
	return index, true
}

func savePSIndexCache(items []psStoreItem) {
	data, err := json.Marshal(items)
	if err != nil {
		return
	}
	_ = os.WriteFile(psIndexCachePath(), data, 0o644)
}

// buildPSTitleIndex выкачивает турецкий каталог целиком: адресный поиск по
// одному тайтлу магазин не отдаёт, а полный список стоит ~50 запросов.
func (s *CatalogParserService) buildPSTitleIndex(ctx context.Context) (*psTitleIndex, error) {
	if cached, ok := loadPSIndexCache(); ok {
		slog.Info("индекс PS Store взят из кэша", slog.Int("items", cached.size))
		return cached, nil
	}

	index := &psTitleIndex{byKey: map[string][]psStoreItem{}}
	const pageSize = 500
	seen := map[string]bool{}
	collected := make([]psStoreItem, 0, 22000)

	for _, container := range psStoreContainers {
		start := 0
		for page := 0; page < 80; page++ {
			if err := ctx.Err(); err != nil {
				return index, err
			}
			endpoint := fmt.Sprintf(
				"https://store.playstation.com/store/api/chihiro/00_09_000/container/%s/%s/19/%s?size=%d&start=%d",
				psStoreRegion, psStoreLocale, container, pageSize, start,
			)
			// Стор регулярно отдаёт 5xx на отдельных страницах. Раньше это
			// обрывало весь контейнер и индекс молча получался неполным —
			// от прогона к прогону находились разные игры.
			var payload psStoreResponse
			var lastErr error
			for attempt := 0; attempt < 3; attempt++ {
				lastErr = s.getJSONWithClient(ctx, s.proxyClient(ctx), endpoint, &payload)
				if lastErr == nil {
					break
				}
				time.Sleep(time.Duration(attempt+1) * time.Second)
			}
			if lastErr != nil {
				slog.Warn("ps index page failed",
					slog.String("container", container),
					slog.Int("start", start),
					slog.String("error", lastErr.Error()))
				if strings.Contains(lastErr.Error(), "403") {
					// Стор закрылся от нашего IP. Дальше долбиться бессмысленно
					// и вредно — только продлевать блокировку.
					return index, fmt.Errorf("PS Store отвечает 403: IP временно заблокирован, нужен прокси или пауза")
				}
				if start == 0 {
					break // контейнера просто нет
				}
				start += pageSize
				continue
			}
			if len(payload.Links) == 0 {
				break
			}
			for _, item := range payload.Links {
				if item.ID == "" || seen[item.ID] {
					continue
				}
				seen[item.ID] = true
				index.add(item)
				collected = append(collected, item)
			}
			s.updateProgress("PS Store", "Индекс каталога TR", index.size, payload.TotalResults, 0)
			start += pageSize
			if payload.TotalResults > 0 && start >= payload.TotalResults {
				break
			}
			// Не частим: полный обход и так занимает минуту, а 403 стоит часов.
			select {
			case <-ctx.Done():
				return index, ctx.Err()
			case <-time.After(400 * time.Millisecond):
			}
		}
	}
	if len(collected) > 0 {
		savePSIndexCache(collected)
	}
	return index, nil
}

// ─────────────── поиск на Xbox ───────────────

func (s *CatalogParserService) searchXboxIDs(ctx context.Context, title string) []string {
	endpoint := fmt.Sprintf("https://www.xbox.com/en-US/search/results?q=%s", url.QueryEscape(title))
	body, err := s.getTextWithRetry(ctx, endpoint, 2)
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	ids := []string{}
	for _, id := range xboxProductIDPattern.FindAllString(body, -1) {
		if seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
		if len(ids) >= 24 {
			break
		}
	}
	return ids
}

func (s *CatalogParserService) resolveXboxWanted(ctx context.Context, want WantedGame) (*domain.CatalogImport, bool) {
	ids := s.searchXboxIDs(ctx, want.Title)
	if len(ids) == 0 {
		return nil, false
	}

	products, err := s.fetchXboxProducts(ctx, ids)
	if err != nil || len(products) == 0 {
		return nil, false
	}

	candidates := make([]titleCandidate, 0, len(products))
	for i := range products {
		localized := pickXboxLocalized(products[i].LocalizedProperties)
		title := strings.TrimSpace(localized.ProductTitle)
		if title == "" || IsAuxiliaryStoreItem(title) {
			continue
		}
		candidates = append(candidates, titleCandidate{title: title, value: products[i]})
	}

	match, _ := pickBestTitleMatch(want.Title, candidates)
	if match == nil {
		return nil, false
	}
	product, ok := match.(xboxProduct)
	if !ok {
		return nil, false
	}
	item := catalogImportFromXbox(product)
	if item == nil {
		return nil, false
	}
	if IsAuxiliaryStoreItem(item.Title) {
		return nil, false
	}
	item.Title = CanonicalGameTitle(want.Title)
	item.TitleKey = CatalogTitleKey(want.Title)
	if item.Description != nil {
		if text := StripStoreBoilerplate(*item.Description); text != "" {
			item.Description = &text
		}
	}
	return item, true
}

// ─────────────── основной проход ───────────────

// ImportWantedGames проводит список через магазины и складывает результат в
// catalog_imports. Цены считаются теми же формулами, что и в обычном парсере,
// раздел (предзаказ/новинка/каталог) — по актуальной дате релиза из магазина.
func (s *CatalogParserService) ImportWantedGames(ctx context.Context, wanted []WantedGame) (*WantedImportReport, error) {
	report := &WantedImportReport{
		Total:        len(wanted),
		StartedAt:    time.Now().UTC(),
		SectionStats: map[string]int{},
		NotFound:     []string{},
		Matches:      []WantedMatch{},
	}

	ensureTRYRUBRate(ctx, s.client)

	index, err := s.buildPSTitleIndex(ctx)
	if err != nil && index.size == 0 {
		return report, err
	}
	report.PSIndexSize = index.size
	slog.Info("ps index built", slog.Int("items", index.size))

	var mu sync.Mutex
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup

	for i, want := range wanted {
		if err := ctx.Err(); err != nil {
			break
		}
		if i%25 == 0 {
			s.updateProgress("Список", "Сверка со сторами", i, len(wanted), report.Imported)
		}

		wg.Add(1)
		go func(want WantedGame) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			found := false

			if want.WantsPS() {
				for _, item := range s.resolvePSWanted(ctx, want, index) {
					mu.Lock()
					report.MatchedPS++
					report.Imported++
					report.SectionStats[item.GameSection]++
					report.Matches = append(report.Matches, WantedMatch{
						Row: want.Row, Title: want.Title, Source: item.Source,
						ExternalID: item.ExternalID, StoreTitle: item.Title,
						Section: item.GameSection, Confidence: "ps",
					})
					mu.Unlock()
					if err := s.repo.UpsertWanted(ctx, item); err != nil {
						slog.Warn("wanted upsert failed", slog.String("title", item.Title), slog.String("error", err.Error()))
					}
					found = true
				}
			}

			if want.WantsXbox() {
				if item, ok := s.resolveXboxWanted(ctx, want); ok {
					mu.Lock()
					report.MatchedXbox++
					report.Imported++
					report.SectionStats[item.GameSection]++
					report.Matches = append(report.Matches, WantedMatch{
						Row: want.Row, Title: want.Title, Source: item.Source,
						ExternalID: item.ExternalID, StoreTitle: item.Title,
						Section: item.GameSection, Confidence: "xbox",
					})
					mu.Unlock()
					if err := s.repo.UpsertWanted(ctx, item); err != nil {
						slog.Warn("wanted upsert failed", slog.String("title", item.Title), slog.String("error", err.Error()))
					}
					found = true
				}
			}

			if !found {
				mu.Lock()
				report.NotFound = append(report.NotFound, want.Title)
				mu.Unlock()
			}
		}(want)
	}
	wg.Wait()

	if filled, err := s.repo.BackfillDescriptionsByTitleKey(ctx); err != nil {
		slog.Warn("описания не сведены между платформами", slog.String("error", err.Error()))
	} else if filled > 0 {
		slog.Info("описания сведены между платформами", slog.Int64("rows", filled))
	}

	sort.Strings(report.NotFound)
	sort.Slice(report.Matches, func(i, j int) bool { return report.Matches[i].Row < report.Matches[j].Row })
	report.FinishedAt = time.Now().UTC()
	return report, nil
}

// resolvePSWanted возвращает импорты по всем подходящим SKU: у одной игры на
// PS часто отдельные позиции под PS4 и PS5 — карточка потом склеится по title_key.
func (s *CatalogParserService) resolvePSWanted(ctx context.Context, want WantedGame, index *psTitleIndex) []*domain.CatalogImport {
	items := index.lookup(want.Title)
	if len(items) == 0 {
		return nil
	}

	candidates := make([]titleCandidate, 0, len(items))
	for i := range items {
		candidates = append(candidates, titleCandidate{title: items[i].Name, value: items[i]})
	}
	match, _ := pickBestTitleMatch(want.Title, candidates)
	if match == nil {
		return nil
	}
	chosen, ok := match.(psStoreItem)
	if !ok {
		return nil
	}

	// Берём и остальные SKU с тем же нормализованным названием — это и есть
	// PS4/PS5-версии одной игры.
	chosenKey := wantedKey(chosen.Name)
	canonicalTitle := CanonicalGameTitle(want.Title)
	canonicalKey := CatalogTitleKey(want.Title)

	result := []*domain.CatalogImport{}
	seen := map[string]bool{}
	for _, item := range items {
		if wantedKey(item.Name) != chosenKey || seen[item.ID] {
			continue
		}
		if !isFullGamePSItem(item) {
			continue
		}
		seen[item.ID] = true
		imp := s.catalogImportFromPS(ctx, item)
		if imp == nil {
			continue
		}
		imp.Title = canonicalTitle
		imp.TitleKey = canonicalKey
		description := ""
		if imp.Description != nil {
			description = StripStoreBoilerplate(*imp.Description)
		}
		if description == "" {
			description = s.psDescription(ctx, item.ID, item.Description)
		}
		if description != "" {
			imp.Description = &description
		} else {
			imp.Description = nil
		}
		result = append(result, imp)
	}

	if len(result) == 0 {
		// У базовой позиции часто нет цены (Sony не отдаёт SKU), а у издания
		// той же игры цена есть. Раньше игра из-за этого просто выпадала из
		// каталога — теперь берём самое дешёвое издание с ценой и честно
		// подписываем его названием издания.
		result = s.resolvePSEditionFallback(ctx, index, chosenKey, canonicalTitle, canonicalKey)
	}
	return result
}

// resolvePSEditionFallback подбирает самое дешёвое издание игры, у которого
// есть цена, и оформляет его отдельной позицией каталога изданий.
func (s *CatalogParserService) resolvePSEditionFallback(
	ctx context.Context,
	index *psTitleIndex,
	baseKey, canonicalTitle, canonicalKey string,
) []*domain.CatalogImport {
	var best *domain.CatalogImport
	var bestItem psStoreItem
	var bestPrice float64

	for _, item := range index.editions(baseKey) {
		if !isFullGamePSItem(item) || IsAuxiliaryStoreItem(item.Name) {
			continue
		}
		imp := s.catalogImportFromPS(ctx, item)
		if imp == nil || imp.OriginalPriceRUB == nil || *imp.OriginalPriceRUB <= 0 {
			continue
		}
		if best == nil || *imp.OriginalPriceRUB < bestPrice {
			best, bestItem, bestPrice = imp, item, *imp.OriginalPriceRUB
		}
	}
	if best == nil {
		return nil
	}

	best.Title = canonicalTitle
	best.TitleKey = canonicalKey
	best.Prices = editionCatalogPrices(best.Prices, editionNameFromTitle(bestItem.Name, canonicalTitle))
	return []*domain.CatalogImport{best}
}
