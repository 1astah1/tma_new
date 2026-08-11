package service

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseWantedGamesCSV(t *testing.T) {
	csv := "№,Игра,Шутер,Раздел,XBOX (консоль),XBOX (PC),PS4,PS5\n" +
		"1,Call of Duty: Modern Warfare 4,Да,Предзаказ,Да,Да,,Да\n" +
		"2,Battlefield 6,Да,Каталог,Да,Да,,Да\n" +
		"3,Battlefield 6,Да,Каталог,Да,,,Да\n" +
		",,,,,,,\n"

	games, err := ParseWantedGamesCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(games) != 2 {
		t.Fatalf("ожидалось 2 игры (дубль схлопывается), получено %d", len(games))
	}
	if !games[0].PS5 || games[0].PS4 || !games[0].Xbox || !games[0].PC {
		t.Fatalf("платформы разобраны неверно: %+v", games[0])
	}
	if !games[0].WantsPS() || !games[0].WantsXbox() {
		t.Fatalf("флаги платформ: %+v", games[0])
	}
}

func TestWantedKeyStripsEditionsAndLocale(t *testing.T) {
	cases := [][2]string{
		{"Call of Duty: Black Ops 7 - Cross-Gen Paketi", "call of duty black ops 7"},
		{"Call of Duty®: Modern Warfare® III", "call of duty modern warfare 3"},
		{"Call of Duty: Modern Warfare - Dijital Standart Sürüm", "call of duty modern warfare"},
		{"Battlefield 2042 Standard Edition", "battlefield 2042"},
	}
	for _, c := range cases {
		if got := wantedKey(c[0]); got != c[1] {
			t.Errorf("wantedKey(%q) = %q, ожидалось %q", c[0], got, c[1])
		}
	}
}

func TestPickBestTitleMatchExact(t *testing.T) {
	candidates := []titleCandidate{
		{title: "Battlefield V", value: "bfv"},
		{title: "Battlefield 6 - Standart Sürüm", value: "bf6"},
		{title: "Battlefield 2042", value: "bf2042"},
	}
	got, confidence := pickBestTitleMatch("Battlefield 6", candidates)
	if got != "bf6" {
		t.Fatalf("ожидался bf6, получено %v", got)
	}
	if confidence != "exact" {
		t.Fatalf("ожидалось точное совпадение, получено %q", confidence)
	}
}

// Самый опасный случай: похожие названия из одной серии не должны
// подменять друг друга — покупатель получит не ту игру.
func TestPickBestTitleMatchRejectsWrongSequel(t *testing.T) {
	candidates := []titleCandidate{
		{title: "Call of Duty: Black Ops 6", value: "bo6"},
		{title: "Call of Duty: Black Ops Cold War", value: "cw"},
	}
	if got, _ := pickBestTitleMatch("Call of Duty: Black Ops 7", candidates); got != nil {
		t.Fatalf("не должно быть совпадения, получено %v", got)
	}
}

func TestPickBestTitleMatchRejectsUnrelated(t *testing.T) {
	candidates := []titleCandidate{
		{title: "FIFA 24", value: "fifa"},
		{title: "Gran Turismo 7", value: "gt7"},
	}
	if got, _ := pickBestTitleMatch("DOOM: The Dark Ages", candidates); got != nil {
		t.Fatalf("не должно быть совпадения, получено %v", got)
	}
}

func TestPSIndexLookupFindsEditionVariants(t *testing.T) {
	index := &psTitleIndex{byKey: map[string][]psStoreItem{}}
	index.add(psStoreItem{ID: "PS4", Name: "DOOM: The Dark Ages - Standart Sürüm"})
	index.add(psStoreItem{ID: "PS5", Name: "DOOM: The Dark Ages"})
	index.add(psStoreItem{ID: "OTHER", Name: "DOOM Eternal"})

	found := index.lookup("DOOM: The Dark Ages")
	if len(found) != 2 {
		t.Fatalf("ожидалось 2 SKU одной игры, получено %d", len(found))
	}
	for _, item := range found {
		if item.ID == "OTHER" {
			t.Fatal("в выдачу попала другая игра серии")
		}
	}
}

func TestCanonicalGameTitle(t *testing.T) {
	cases := [][2]string{
		{"Call of Duty®: Black Ops 7 - Cross-Gen Paketi", "Call of Duty®: Black Ops 7"},
		{"Call of Duty: Modern Warfare - Dijital Standart Sürüm", "Call of Duty: Modern Warfare"},
		{"Call of Duty: WWII (Windows)", "Call of Duty: WWII"},
		{"Battlefield 6", "Battlefield 6"},
	}
	for _, c := range cases {
		if got := CanonicalGameTitle(c[0]); got != c[1] {
			t.Errorf("CanonicalGameTitle(%q) = %q, ожидалось %q", c[0], got, c[1])
		}
	}
}

func TestIsAuxiliaryStoreItem(t *testing.T) {
	aux := []string{
		"Call of Duty®: Vanguard - улучшение с набором 'Два поколения'",
		"Battlefield 2042 Season Pass",
		"FIFA 24 - 1050 FC Points",
		"DOOM Eternal: The Ancient Gods — Expansion",
	}
	for _, title := range aux {
		if !IsAuxiliaryStoreItem(title) {
			t.Errorf("%q должно быть отсеяно как не-игра", title)
		}
	}

	games := []string{
		"Call of Duty®: Black Ops 7 - Cross-Gen Paketi",
		"Battlefield 6",
		"DOOM: The Dark Ages",
	}
	for _, title := range games {
		if IsAuxiliaryStoreItem(title) {
			t.Errorf("%q — это игра, отсеивать нельзя", title)
		}
	}
}

func TestIsFullGamePSItem(t *testing.T) {
	full := psStoreItem{Name: "Battlefield 6", TopCategory: "downloadable_game"}
	full.GameContentTypes = append(full.GameContentTypes, struct {
		Key string `json:"key"`
	}{Key: "FULL_GAME"})
	if !isFullGamePSItem(full) {
		t.Fatal("полная игра должна проходить фильтр")
	}

	addon := psStoreItem{Name: "Battlefield 6 Season Pass", TopCategory: "add_on"}
	addon.GameContentTypes = append(addon.GameContentTypes, struct {
		Key string `json:"key"`
	}{Key: "ADD_ON"})
	if isFullGamePSItem(addon) {
		t.Fatal("дополнение не должно попадать в каталог")
	}
}

func TestStripStoreBoilerplate(t *testing.T) {
	raw := "Чтобы играть в эту игру на PS5, возможно, потребуется обновить системное ПО. " +
		"Хотя эта игра поддерживает PS5, некоторые функции PS4 недоступны. " +
		"Вернитесь на поля сражений будущего в новой части легендарной серии."
	got := StripStoreBoilerplate(raw)
	if strings.HasPrefix(got, "Чтобы играть") {
		t.Fatalf("преамбула не убрана: %q", got)
	}
	if !strings.HasPrefix(got, "Вернитесь на поля") {
		t.Fatalf("потерян полезный текст: %q", got)
	}
}

func TestStripStoreBoilerplateKeepsPlainText(t *testing.T) {
	raw := "Шутер от первого лица с кампанией и мультиплеером."
	if got := StripStoreBoilerplate(raw); got != raw {
		t.Fatalf("обычное описание изменено: %q", got)
	}
}

func TestStripStoreBoilerplateDropsSupportLinks(t *testing.T) {
	raw := "Более подробная информация приведена на странице PlayStation. com/bc. Демоны вторглись на марсианскую базу."
	got := StripStoreBoilerplate(raw)
	if !strings.HasPrefix(got, "Демоны") {
		t.Fatalf("служебное предложение не убрано: %q", got)
	}
}

// Регресс: раньше метрика делила на длину короткого названия, из-за чего
// игра идеально «совпадала» со своим сезонным пропуском.
func TestPickBestTitleMatchRejectsSeasonPass(t *testing.T) {
	candidates := []titleCandidate{
		{title: "Battlefield 6 Season Pass", value: "pass"},
		{title: "Battlefield 6 - Endurance Weapon Skin Set", value: "skins"},
	}
	if got, _ := pickBestTitleMatch("Battlefield 6", candidates); got != nil {
		t.Fatalf("к игре подобрано дополнение: %v", got)
	}
}

func TestPickBestTitleMatchStillFindsRealGame(t *testing.T) {
	candidates := []titleCandidate{
		{title: "Battlefield 6 Season Pass", value: "pass"},
		{title: "Battlefield™ 6 - Standart Sürüm", value: "game"},
	}
	got, _ := pickBestTitleMatch("Battlefield 6", candidates)
	if got != "game" {
		t.Fatalf("игра не найдена, получено %v", got)
	}
}

func TestWantedKeyIgnoresPlatformSuffix(t *testing.T) {
	if wantedKey("Battlefield 2042 для PS5") != wantedKey("Battlefield™ 2042") {
		t.Fatalf("платформа в названии мешает совпадению: %q vs %q",
			wantedKey("Battlefield 2042 для PS5"), wantedKey("Battlefield™ 2042"))
	}
}

func TestPickBestTitleMatchWithPlatformSuffix(t *testing.T) {
	candidates := []titleCandidate{
		{title: "Battlefield™ 2042", value: "game"},
		{title: "Battlefield 2042 Year 1 Pass", value: "pass"},
	}
	got, _ := pickBestTitleMatch("Battlefield 2042 для PS5", candidates)
	if got != "game" {
		t.Fatalf("ожидалась игра, получено %v", got)
	}
}

// В турецком сторе стандартного издания часто нет — только «Eksiksiz Sürüm»
// (полное) или «Ragnarök Edition». Без отсева локальных слов игра не находится.
func TestWantedKeyStripsTurkishEditionWords(t *testing.T) {
	pairs := [][2]string{
		{"Assassin's Creed Valhalla", "Assassin's Creed Valhalla - Eksiksiz Sürüm"},
		{"Anno 1800 Console Edition", "Anno 1800 Konsol Sürümü"},
	}
	for _, pair := range pairs {
		if wantedKey(pair[0]) != wantedKey(pair[1]) {
			t.Errorf("не сошлись: %q vs %q", wantedKey(pair[0]), wantedKey(pair[1]))
		}
	}
}

// Дефис внутри слова — не отделитель издания: «All-Star Edition» нельзя
// резать по дефису, иначе в каталоге появляется «NBA 2K25 All».
func TestCanonicalGameTitleKeepsHyphenatedWords(t *testing.T) {
	cases := [][2]string{
		{"NBA 2K25 All-Star Edition", "NBA 2K25"},
		{"Spider-Man 2", "Spider-Man 2"},
		{"Battlefield 6 - Standart Sürüm", "Battlefield 6"},
	}
	for _, c := range cases {
		if got := CanonicalGameTitle(c[0]); got != c[1] {
			t.Errorf("CanonicalGameTitle(%q) = %q, ожидалось %q", c[0], got, c[1])
		}
	}
}

// Строки списка для одной игры отличаются хвостом с платформами. Если он
// попадает в ключ, PS и Xbox расходятся по разным карточкам.
func TestCanonicalTitleDropsPlatformTail(t *testing.T) {
	cases := [][2]string{
		{"Battlefield 2042 для PS5", "Battlefield 2042"},
		{"EA SPORTS FC 26 Standart Sürüm Xbox One ve Xbox Series X|S", "EA SPORTS FC 26"},
		{"Tom Clancy's Rainbow Six Extraction PS4 & PS5", "Tom Clancy's Rainbow Six Extraction"},
		{"Battlefield 6", "Battlefield 6"},
	}
	for _, c := range cases {
		if got := CanonicalGameTitle(c[0]); got != c[1] {
			t.Errorf("CanonicalGameTitle(%q) = %q, ожидалось %q", c[0], got, c[1])
		}
	}
}

func TestOnePlatformKeyForOneGame(t *testing.T) {
	rows := []string{
		"Battlefield 2042 для PS5",
		"Battlefield 2042",
		"Battlefield™ 2042 Standart Sürüm",
	}
	first := wantedKey(CanonicalGameTitle(rows[0]))
	for _, row := range rows[1:] {
		if got := wantedKey(CanonicalGameTitle(row)); got != first {
			t.Errorf("%q даёт ключ %q, а ожидался общий %q", row, got, first)
		}
	}
}

// Реальные расхождения из каталога: после нормализации платформы одной игры
// обязаны получить общий ключ.
func TestSplitCardsGetCommonKey(t *testing.T) {
	groups := [][]string{
		{"Insurgency: Sandstorm [PS4 & PS5]", "Insurgency: Sandstorm"},
		{"RESIDENT EVIL 3 for Xbox", "RESIDENT EVIL 3"},
		{"Anno 1800 Konsol", "Anno 1800 Console"},
		{"Издание WWE 2K26 King of Kings", "WWE 2K26 King of Kings"},
		{"DEATH STRANDING DIRECTOR’S CUT", "DEATH STRANDING DIRECTOR'S CUT"},
		{"Cult of the Lamb: Woolhaven", "Cult of the Lamb - Woolhaven"},
		{"Gungrave G.O.R.E", "Gungrave G.O.R.E."},
		{"DRAGON BALL Z: KAKAROT PS4 & PS5", "DRAGON BALL Z : KAKAROT"},
	}
	for _, group := range groups {
		want := wantedKey(CanonicalGameTitle(group[0]))
		for _, title := range group[1:] {
			if got := wantedKey(CanonicalGameTitle(title)); got != want {
				t.Errorf("%q → %q, а %q → %q — карточки не склеятся", group[0], want, title, got)
			}
		}
	}
}

// И наоборот: разные игры серии не должны схлопнуться в одну карточку.
func TestDifferentGamesKeepDifferentKeys(t *testing.T) {
	pairs := [][2]string{
		{"FINAL FANTASY VI", "FINAL FANTASY I-VI"},
		{"Battlefield 6", "Battlefield 2042"},
		{"WWE 2K26 King of Kings", "WWE 2K26 Monday Night War"},
	}
	for _, pair := range pairs {
		a := wantedKey(CanonicalGameTitle(pair[0]))
		b := wantedKey(CanonicalGameTitle(pair[1]))
		if a == b {
			t.Errorf("%q и %q схлопнулись в один ключ %q", pair[0], pair[1], a)
		}
	}
}

func TestCanonicalTitleDropsConsoleWord(t *testing.T) {
	for _, raw := range []string{"Anno 1800 Konsol", "Anno 1800 Console"} {
		if got := CanonicalGameTitle(raw); got != "Anno 1800" {
			t.Errorf("CanonicalGameTitle(%q) = %q, ожидалось «Anno 1800»", raw, got)
		}
	}
}

// Ключ должен быть один и тот же независимо от того, кто его считает.
func TestCatalogTitleKeyIsSingleSourceOfTruth(t *testing.T) {
	for _, title := range []string{"Battlefield 2042 для PS5", "Battlefield 2042", "RESIDENT EVIL 3 for Xbox"} {
		if CatalogTitleKey(title) != wantedKey(CanonicalGameTitle(title)) {
			t.Errorf("расхождение ключа для %q", title)
		}
	}
	if CatalogTitleKey("Battlefield 2042 для PS5") != CatalogTitleKey("Battlefield 2042") {
		t.Fatal("платформы одной игры получили разные ключи")
	}
}

func TestEditionNameFromTitle(t *testing.T) {
	cases := [][3]string{
		{"HELLDIVERS™ 2 Super Citizen Sürümü", "HELLDIVERS 2", "Super Citizen Edition"},
		{"Marathon Deluxe Edition", "Marathon", "Deluxe Edition"},
		{"Battlefield 2042", "Battlefield 2042", "Standard Edition"},
	}
	for _, c := range cases {
		if got := editionNameFromTitle(c[0], c[1]); got != c[2] {
			t.Errorf("editionNameFromTitle(%q, %q) = %q, ожидалось %q", c[0], c[1], got, c[2])
		}
	}
}

// Если продаём не стандартную версию, покупатель обязан видеть какую.
func TestEditionCatalogPricesNamesTheEdition(t *testing.T) {
	out := editionCatalogPrices([]byte(`{"tr":5000,"ua":4000}`), "Super Citizen Edition")

	var parsed struct {
		TR      float64 `json:"tr"`
		Catalog map[string][]struct {
			Name  string  `json:"name"`
			Price float64 `json:"price"`
		} `json:"edition_catalog"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("не разбирается: %v", err)
	}
	if parsed.TR != 5000 {
		t.Fatalf("потеряна цена tr: %v", parsed.TR)
	}
	if len(parsed.Catalog["ps_tr"]) != 1 || parsed.Catalog["ps_tr"][0].Name != "Super Citizen Edition" {
		t.Fatalf("издание не подписано: %+v", parsed.Catalog)
	}
	if len(parsed.Catalog["ps_ua"]) != 1 || parsed.Catalog["ps_ua"][0].Price != 4000 {
		t.Fatalf("украинская цена потеряна: %+v", parsed.Catalog)
	}
}

func TestPSIndexFindsEditions(t *testing.T) {
	index := &psTitleIndex{byKey: map[string][]psStoreItem{}}
	index.add(psStoreItem{ID: "base", Name: "HELLDIVERS™ 2"})
	index.add(psStoreItem{ID: "sc", Name: "HELLDIVERS™ 2 Super Citizen Sürümü"})
	index.add(psStoreItem{ID: "other", Name: "HELLDIVERS™ Dive Harder Edition"})

	found := index.editions(wantedKey("HELLDIVERS 2"))
	if len(found) != 1 || found[0].ID != "sc" {
		t.Fatalf("ожидалось только издание Super Citizen, получено %+v", found)
	}
}

func TestMatchProblem(t *testing.T) {
	ok := [][2]string{
		{"Kingdom Come: Deliverance", "Kingdom Come: Deliverance - Royal Edition"},
		{"DRAGON BALL Z : KAKAROT", "DRAGON BALL Z : KAKAROT - MASTER EDITION"},
		{"Battlefield 2042", "Battlefield™ 2042"},
	}
	for _, c := range ok {
		if got := matchProblem(c[0], c[1]); got != "" {
			t.Errorf("издание той же игры помечено как проблема: %q → %q (%s)", c[0], c[1], got)
		}
	}

	bad := [][2]string{
		{"NBA 2K26", "NBA 2K26 Pro Pass: Season 9 (ПК с Windows)"},
		{"UFC 4", "UFC® 4 — стартовый комплект"},
		{"Crash Bandicoot 4: Это вопрос времени", "Crash Bandicoot®"},
		{"Red Dead Redemption 2", "Red Dead Redemption"},
		{"EA SPORTS PGA TOUR", "EA SPORTS™ PGA TOUR™ — 5750 ОЧКОВ PGA TOUR"},
	}
	for _, c := range bad {
		if matchProblem(c[0], c[1]) == "" {
			t.Errorf("подмена товара не поймана: %q → %q", c[0], c[1])
		}
	}
}

// Набор скинов, который сам заявлен в списке, — не подмена: в названии
// карточки честно написано, что это набор.
func TestMatchProblemAllowsHonestDLCCards(t *testing.T) {
	if got := matchProblem(
		"Insurgency: Sandstorm - Endurance Weapon Skin Set",
		"Insurgency: Sandstorm - Endurance Weapon Skin Set"); got != "" {
		t.Fatalf("честный набор помечен как проблема: %s", got)
	}
}

func TestMatchProblemIgnoresTurkishPreposition(t *testing.T) {
	if got := matchProblem("Xbox Series X|S için The Quarry", "The Quarry для Xbox Series X|S"); got != "" {
		t.Fatalf("та же игра помечена как другая: %s", got)
	}
}

func TestMatchProblemIgnoresYearInBrackets(t *testing.T) {
	if got := matchProblem("Call of Duty 4: Modern Warfare (2007)", "Call of Duty 4: Modern Warfare"); got != "" {
		t.Fatalf("год в скобках сломал сверку: %s", got)
	}
}

// Осознанный размен: слово «комплект» больше не считается признаком
// дополнения, потому что кросс-ген издания игр называются так же. Из-за этого
// наборы машин попадают под подозрение — и это лучше, чем прятать саму игру.
func TestMatchProblemFlagsTranslatedCarPack(t *testing.T) {
	got := matchProblem(
		"Need for Speed Unbound - Porsche 959 S ‘87 Set",
		"Need for Speed™ Unbound — комплект Porsche 959 S 87")
	if got == "" {
		t.Fatal("ожидалась пометка: названия набора не совпадают пословно")
	}
}

func TestMatchProblemHandlesRussianEditionSuffix(t *testing.T) {
	if got := matchProblem("Back 4 Blood: Ultimate-издание", "Back 4 Blood: Ultimate Edition PS4 & PS5"); got != "" {
		t.Fatalf("та же игра помечена как другая: %s", got)
	}
}

// Кросс-ген издание — это сама игра, его нельзя принимать за дополнение.
func TestCrossGenBundleIsNotDLC(t *testing.T) {
	for _, title := range []string{
		"Call of Duty®: Black Ops 7 - Набор Cross-Gen",
		"Комплект перекрёстного издания STAR WARS™ Jedi",
	} {
		if IsAuxiliaryStoreItem(title) {
			t.Errorf("%q — это игра, а не дополнение", title)
		}
	}
}

// «Cult of the Lamb - Sinful Pack» на Xbox стоит 279 ₽ против 3217 ₽ за
// издание на PS — это докупка, а не игра.
func TestMatchProblemCatchesPackSuffix(t *testing.T) {
	for _, pair := range [][2]string{
		{"Cult of the Lamb: Sinful", "Cult of the Lamb - Sinful Pack"},
		{"Cult of the Lamb: Unholy", "Cult of the Lamb - Unholy Pack Bundle"},
	} {
		if matchProblem(pair[0], pair[1]) == "" {
			t.Errorf("докупка не поймана: %q → %q", pair[0], pair[1])
		}
	}
}

// Кросс-ген издание при этом остаётся игрой.
func TestCrossGenStillPasses(t *testing.T) {
	if got := matchProblem("Call of Duty: Black Ops 7", "Call of Duty®: Black Ops 7 - Набор Cross-Gen"); got != "" {
		t.Fatalf("кросс-ген издание помечено как докупка: %s", got)
	}
}
