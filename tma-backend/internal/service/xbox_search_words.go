package service

import "strings"

// xboxGameSearchWords — запросы к поиску Xbox.com TR для покрытия каталога.
// Каждый запрос отдаёт ~50–90 уникальных ID; вместе дают 10k+ игр.
var xboxGameSearchWords = []string{
	"action", "adventure", "age", "anime", "arcade", "assetto", "assassin", "avengers",
	"batman", "battle", "baldur", "baseball", "basketball", "bethesda", "black", "borderlands",
	"boxing", "builder", "bundle", "bus", "call", "cars", "city", "classic", "civilization",
	"collection", "commando", "company", "corsa", "craft", "creed", "cricket", "crew",
	"cyber", "dance", "dark", "dead", "deluxe", "destiny", "diablo", "dirt", "disney",
	"dlc", "doom", "dragon", "drift", "duty", "edition", "elden", "elder", "empire",
	"empires", "fable", "fallout", "fantasy", "farm", "fifa", "fighting", "final",
	"flight", "football", "forza", "formula", "fortnite", "fps", "golf", "gold",
	"gran", "gta", "gears", "halo", "harry", "heroes", "hockey", "horror", "indie",
	"kingdom", "lego", "legends", "lego", "lord", "madden", "magic", "manager", "marvel",
	"mario", "metal", "minecraft", "modern", "moto", "motogp", "multiplayer", "nba",
	"need", "nfl", "ninja", "nintendo", "online", "open", "ops", "overwatch", "pack",
	"pass", "pes", "pixel", "plane", "platform", "potter", "premium", "prince", "project",
	"puzzle", "racing", "rally", "rayman", "red", "resident", "retro", "rings", "rpg",
	"rugby", "samurai", "scrolls", "season", "shooter", "sim", "simulator", "skyrim",
	"sniper", "soccer", "sonic", "souls", "space", "special", "speed", "spider", "sport",
	"star", "starcraft", "story", "strategy", "survival", "tactical", "tennis", "tomb",
	"total", "train", "truck", "turismo", "ufc", "ultimate", "volley", "vr", "war",
	"warcraft", "warfare", "watch", "witcher", "wolf", "world", "wrc", "wwe", "zelda",
	"zombie", "ubisoft", "square", "capcom", "bandai", "namco", "sega", "konami",
	"activision", "blizzard", "rockstar", "naughty", "insomniac", "fromsoftware",
	"bioware", "dice", "remedy", "playground", "rare", "obsidian", "arkane",
	"2020", "2021", "2022", "2023", "2024", "2025", "2026",
}

func buildXboxSearchQueries() []string {
	seen := map[string]bool{}
	out := make([]string, 0, 900)
	add := func(q string) {
		q = trimLower(q)
		if q == "" || seen[q] {
			return
		}
		seen[q] = true
		out = append(out, q)
	}

	for _, q := range xboxSearchQueries {
		add(q)
	}
	for _, q := range xboxGameSearchWords {
		add(q)
	}
	for c := 'a'; c <= 'z'; c++ {
		for d := 'a'; d <= 'z'; d++ {
			add(string([]byte{byte(c), byte(d)}))
		}
	}

	return out
}

func trimLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
