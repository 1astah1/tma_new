package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Позиции, которые нельзя продавать как саму игру: апгрейды между поколениями,
// дополнения, пропуски, валюта. По названию они почти неотличимы от игры и
// стоят дёшево — покупатель решит, что взял игру, и получит апгрейд.
var auxiliaryItemRe = regexp.MustCompile(`(?i)(улучшени|обновлени|апгрейд|upgrade|y[uü]kseltme|дополнени|add[- ]?on|expansion|dlc|season pass|battle pass|сезонный пропуск|пропуск|bo[eé]st|бустер|points|бонус|валют|currency|soundtrack|саундтрек|тема|theme|аватар|avatar|demo|демо|пробн|trial|beta|бета|weapon skin|skin set|gear set|cosmetic|космети|outfit|облик|скин|набор оружия|charm|эмблем|emblem|calling card|operator pack|starter pack|founder|\bset\b|бандл)`)

// Издание для двух поколений консолей — это сама игра, а не докупка.
var crossGenRe = regexp.MustCompile(`(?i)(cross[- ]?gen|перекр[её]стн\p{L}*\s+издани\p{L}*)`)

// IsAuxiliaryStoreItem — позиция магазина не является полноценной игрой.
func IsAuxiliaryStoreItem(title string) bool {
	if strings.TrimSpace(title) == "" {
		return true
	}
	// «Cross-Gen Bundle/Paketi» — это как раз издание самой игры, не дополнение.
	cleaned := crossGenRe.ReplaceAllString(title, " ")
	return auxiliaryItemRe.MatchString(cleaned)
}

// isFullGamePSItem доверяет типу контента из самого стора, а не названию.
func isFullGamePSItem(item psStoreItem) bool {
	for _, contentType := range item.GameContentTypes {
		switch strings.ToUpper(strings.TrimSpace(contentType.Key)) {
		case "FULL_GAME", "GAME_BUNDLE", "PREMIUM_EDITION":
			return true
		case "ADD_ON", "DLC", "CURRENCY", "SUBSCRIPTION", "DEMO":
			return false
		}
	}
	switch strings.ToLower(strings.TrimSpace(item.TopCategory)) {
	case "downloadable_game", "game_bundle", "bundle", "premium_edition":
		return true
	case "add_on", "game_related_content", "currency", "demo":
		return false
	}
	// Тип не указан — решаем по названию.
	return !IsAuxiliaryStoreItem(item.Name)
}

// ─────────────── человеческое название карточки ───────────────

var (
	wantedTailRe  = regexp.MustCompile(`(?i)(\s+[-–—]\s*|\s*[-–—]\s+)[^-–—]*\b(s[uü]r[uü]m[uü]?|paket(i|leri)?|edition|bundle|pack|издани\p{L}*|набор|версия|standart|standard|deluxe|ultimate|premium|gold|complete|cross[- ]?gen|dijital|digital)\b[^-–—]*$`)
	wantedTrailRe = regexp.MustCompile(`(?i)\s*[\(\[](windows|pc|ps4|ps5|ps|xbox)[^)\]]*[\)\]]\s*$`)

	// Хвост с перечислением платформ: «для PS5», «Xbox One ve Xbox Series X|S»,
	// «PS4 & PS5». Он не про то, какая это игра, а про то, где она продаётся —
	// в названии карточки только мешает и разводит платформы по разным ключам.
	wantedPlatformTailWordRe = regexp.MustCompile(`(?i)[\s,/&+|·-]+(для|for|на|ps4|ps5|ps|playstation|xbox|one|series|pc|windows|win|ve|and|и|x\|s|x|s)$`)
	wantedNoiseTitleRe       = regexp.MustCompile(`\s{2,}`)

	// Хвостовые слова издания без отделителя: «NBA 2K25 All-Star Edition».
	wantedTrailingWordRe = regexp.MustCompile(`(?i)\s+(edition|s[uü]r[uü]m[uü]*|paket(i|leri)?|bundle|издание|версия|all-star|standard|standart|deluxe|ultimate|premium|gold|complete|definitive|goty|anniversary|collection|koleksiyon[ual]*|eksiksiz|dijital|digital|cross[- ]?gen|konsol|console|консол\p{L}*)$`)
)

// stripPlatformTail срезает хвост из платформ и союзов, но не трогает
// название, целиком состоящее из них.
func stripPlatformTail(title string) string {
	for i := 0; i < 8; i++ {
		next := wantedPlatformTailWordRe.ReplaceAllString(title, "")
		if next == title || strings.TrimSpace(next) == "" {
			break
		}
		title = next
	}
	return title
}

// CanonicalGameTitle делает из магазинного названия одно человеческое:
// «Call of Duty: Black Ops 7 - Cross-Gen Paketi» → «Call of Duty: Black Ops 7».
// Карточка одна на все платформы, поэтому и название должно быть одно.
// «Издание WWE 2K26 …» — приставка издания перед названием игры.
var wantedLeadingEditionRe = regexp.MustCompile(`(?i)^(издани\p{L}*|версия|набор)\s+`)

func CanonicalGameTitle(raw string) string {
	title := wantedLeadingEditionRe.ReplaceAllString(strings.TrimSpace(raw), "")
	if title == "" {
		return ""
	}
	for i := 0; i < 5; i++ {
		next := wantedTailRe.ReplaceAllString(title, "")
		next = wantedTrailRe.ReplaceAllString(next, "")
		next = wantedTrailingWordRe.ReplaceAllString(next, "")
		next = stripPlatformTail(next)
		next = strings.TrimSpace(strings.Trim(strings.TrimSpace(next), "-–—:,"))
		if next == "" || next == title {
			break
		}
		title = next
	}
	title = wantedNoiseTitleRe.ReplaceAllString(title, " ")
	return strings.TrimSpace(title)
}

// ─────────────── описания ───────────────

// Юридическая и техническая преамбула сторов: про обновление прошивки,
// подписки и онлайн-функции. К самой игре отношения не имеет, а в карточке
// съедает первый экран и выглядит как описание.
var storeBoilerplateRe = regexp.MustCompile(`(?i)^(чтобы играть в эту игру|для игры в эту игру|хотя эта игра|эта игра поддерживает|онлайн-функци|для игры по сети|загрузка этого продукта|использование этого продукта|требуется подписка|более подробн\p{L}* информаци|дополнительн\p{L}* информаци|подробнее (см|на)|this game (may )?require|to play this game|online features require|software subscription|downloading this product|(for )?more information|bu oyunu|bu [uü]r[uü]n)`)

// isLeadingNoiseSentence — предложение в начале описания, которое ничего не
// говорит об игре: юридическая преамбула, обрывок ссылки, огрызок в пару слов.
func isLeadingNoiseSentence(sentence string) bool {
	if storeBoilerplateRe.MatchString(sentence) {
		return true
	}
	lower := strings.ToLower(sentence)
	if strings.Contains(lower, "http") || strings.Contains(lower, "www.") || strings.Contains(lower, ".com") {
		return true
	}
	return len([]rune(sentence)) < 20
}

// StripStoreBoilerplate выбрасывает служебные предложения из начала описания.
func StripStoreBoilerplate(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}

	sentences := regexp.MustCompile(`(?U)[^.!?]+[.!?]+\s*`).FindAllString(trimmed, -1)
	if len(sentences) == 0 {
		return trimmed
	}

	kept := make([]string, 0, len(sentences))
	for _, sentence := range sentences {
		candidate := strings.TrimSpace(sentence)
		if candidate == "" {
			continue
		}
		if len(kept) == 0 && isLeadingNoiseSentence(candidate) {
			continue
		}
		kept = append(kept, candidate)
	}

	result := strings.TrimSpace(strings.Join(kept, " "))
	if result == "" {
		return trimmed
	}
	return result
}

// psDescription собирает описание по цепочке локалей: русское из RU-стора,
// затем русское из украинского, затем английское. Пустая карточка хуже,
// чем карточка с английским текстом.
func (s *CatalogParserService) psDescription(ctx context.Context, externalID, fallback string) string {
	if strings.TrimSpace(externalID) == "" {
		return cleanDescription(fallback)
	}

	locales := [][2]string{
		{psDescriptionRegion, psDescriptionLocale}, // RU/ru
		{psStoreRegionUA, "ru"},                    // UA/ru
		{"US", "en"},                               // US/en
	}
	for _, locale := range locales {
		endpoint := fmt.Sprintf(
			"https://store.playstation.com/store/api/chihiro/00_09_000/container/%s/%s/19/%s/0",
			locale[0], locale[1], url.PathEscape(externalID),
		)
		var game psStoreItem
		if err := s.getJSONWithClient(ctx, s.proxyClient(ctx), endpoint, &game); err != nil {
			continue
		}
		if text := StripStoreBoilerplate(cleanDescription(game.Description)); text != "" {
			return text
		}
	}
	return StripStoreBoilerplate(cleanDescription(fallback))
}

// CatalogTitleKey — единственный источник правды для ключа склейки карточек.
// Раньше их было два: адресный импорт считал ключ по каноническому названию,
// а дедупликация пересчитывала его через NormalizeGameTitle — и уже склеенные
// платформы снова разъезжались по разным карточкам.
func CatalogTitleKey(title string) string {
	if key := wantedKey(CanonicalGameTitle(title)); key != "" {
		return key
	}
	return NormalizeGameTitle(title)
}

// ─────────────── издания вместо базовой позиции ───────────────

// editionNameFromTitle достаёт человеческое имя издания из магазинного
// названия: «HELLDIVERS™ 2 Super Citizen Sürümü» → «Super Citizen Edition».
func editionNameFromTitle(storeTitle, canonicalTitle string) string {
	clean := func(value string) []string {
		v := wantedPunctOnlyRe.ReplaceAllString(value, " ")
		v = strings.ReplaceAll(v, ":", " ")
		return strings.Fields(v)
	}

	storeTokens := clean(storeTitle)
	baseTokens := clean(canonicalTitle)

	// Отрезаем от названия издания ту часть, что совпадает с названием игры.
	i := 0
	for i < len(storeTokens) && i < len(baseTokens) &&
		strings.EqualFold(storeTokens[i], baseTokens[i]) {
		i++
	}

	rest := make([]string, 0, len(storeTokens))
	for _, token := range storeTokens[i:] {
		if wantedEditionWordRe.MatchString(token) {
			continue
		}
		if len([]rune(token)) < 2 {
			continue
		}
		rest = append(rest, token)
	}

	if len(rest) == 0 {
		return "Standard Edition"
	}
	return strings.Join(rest, " ") + " Edition"
}

var (
	wantedPunctOnlyRe   = regexp.MustCompile(`[®™©]`)
	wantedEditionWordRe = regexp.MustCompile(`(?i)\b(s[uü]r[uü]m[uü]*|edition|издание|версия|paket(i|leri)?)\b`)
)

// editionCatalogPrices превращает обычные цены позиции в каталог изданий:
// покупатель должен видеть, что берёт не стандартную версию.
func editionCatalogPrices(raw json.RawMessage, editionName string) json.RawMessage {
	var prices map[string]float64
	if err := json.Unmarshal(raw, &prices); err != nil || len(prices) == 0 {
		return raw
	}

	catalog := map[string][]map[string]interface{}{}
	if tr, ok := prices["tr"]; ok && tr > 0 {
		catalog["ps_tr"] = []map[string]interface{}{{"id": "edition", "name": editionName, "price": tr}}
	}
	if ua, ok := prices["ua"]; ok && ua > 0 {
		catalog["ps_ua"] = []map[string]interface{}{{"id": "edition", "name": editionName, "price": ua}}
	}
	if len(catalog) == 0 {
		return raw
	}

	out := map[string]interface{}{"edition_catalog": catalog}
	for key, value := range prices {
		out[key] = value
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		return raw
	}
	return encoded
}

// matchProblem объясняет, почему карточка не соответствует позиции в сторе.
// Пустая строка — всё в порядке.
//
// Издание той же игры («Kingdom Come: Deliverance» → «… Royal Edition»)
// нарушением не считается: базовая версия часто вообще без цены, и издание —
// единственный способ её продать. А вот дополнение, валюта или другая игра
// серии — это уже подмена товара.
func matchProblem(cardTitle, storeTitle string) string {
	if strings.TrimSpace(storeTitle) == "" {
		return ""
	}
	// Если сама карточка — это заявленный в списке набор или скины, а в сторе
	// лежит ровно он же, подмены нет: покупатель видит в названии, что берёт.
	cardIsAuxiliary := IsAuxiliaryStoreItem(cardTitle)
	storeIsAuxiliary := IsAuxiliaryStoreItem(storeTitle)

	if storeIsAuxiliary && !cardIsAuxiliary {
		return "в сторе это дополнение, а не игра"
	}
	// Обе стороны — заявленный в списке набор. Названия наборов переводят
	// («Set» ↔ «комплект»), поэтому сверяем не пословно, а на общую основу.
	if cardIsAuxiliary && storeIsAuxiliary {
		if titleSimilarity(cardTitle, storeTitle) < 0.4 {
			return "в сторе другой набор: " + storeTitle
		}
		return ""
	}

	card := wantedTokens(cardTitle)
	store := wantedTokens(storeTitle)
	if len(card) == 0 || len(store) == 0 {
		return ""
	}

	// Все смысловые слова карточки должны быть в названии позиции.
	// Иначе это другая игра серии: «Crash Bandicoot 4» против «Crash Bandicoot».
	for token := range card {
		if !store[token] {
			return "в сторе другая игра: " + storeTitle
		}
	}

	// То, что осталось сверх названия игры, должно быть именем издания,
	// а не пропуском, валютой или набором косметики.
	for token := range store {
		if card[token] {
			continue
		}
		if dlcMarkerTokens[token] {
			return "в сторе дополнение к игре: " + storeTitle
		}
	}
	return ""
}

// Слова, которые в довеске к названию игры означают не издание, а докупку.
var dlcMarkerTokens = map[string]bool{
	"pass": true, "пропуск": true, "абонемент": true, "сезонный": true, "season": true,
	"очков": true, "очки": true, "points": true, "монет": true, "coins": true,
	"комплект": true, "костюмов": true, "костюм": true, "костюмы": true,
	"скинов": true, "скины": true, "outfit": true, "skins": true, "skin": true,
	"дополнение": true, "дополнения": true, "expansion": true, "dlc": true,
	"улучшение": true, "upgrade": true, "boost": true, "бустер": true,
	"валюта": true, "currency": true, "credits": true, "кредитов": true,
	"appearance": true, "внешний": true, "vip": true, "battle": true,
}
