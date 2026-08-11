// import-wanted проводит список желаемых игр (CSV из таблицы) через PS Store
// и Xbox Store, складывает их в catalog_imports и публикует карточки.
//
// Использование:
//
//	import-wanted -file seeds/wanted_games.csv [-limit 50] [-dry] [-report out.json]
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"tma-backend/internal/config"
	"tma-backend/internal/logger"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

func main() {
	file := flag.String("file", "seeds/wanted_games.csv", "CSV со списком игр")
	limit := flag.Int("limit", 0, "обработать только первые N строк (0 — все)")
	dry := flag.Bool("dry", false, "только разобрать список, без обращения к магазинам")
	reportPath := flag.String("report", "", "куда сохранить JSON-отчёт")
	publish := flag.Bool("publish", true, "публиковать найденное в каталог")
	fixTitles := flag.Bool("fix-titles", false, "только привести названия карточек к канону, без обхода магазинов")
	audit := flag.Bool("audit", false, "проверить, что карточки ведут на саму игру, а не на дополнение")
	auditApply := flag.Bool("audit-apply", false, "вместе с -audit: скрыть найденные неверные карточки")
	flag.Parse()

	cfg := config.Load()
	logger.Init(cfg.App.Environment)

	handle, err := os.Open(*file)
	if err != nil {
		slog.Error("не открывается список", slog.String("file", *file), slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer handle.Close()

	wanted, err := service.ParseWantedGamesCSV(handle)
	if err != nil {
		slog.Error("не разбирается список", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if *limit > 0 && *limit < len(wanted) {
		wanted = wanted[:*limit]
	}
	slog.Info("список загружен", slog.Int("games", len(wanted)))

	if *dry {
		for _, game := range wanted {
			fmt.Printf("%4d  %-70s ps4=%v ps5=%v xbox=%v pc=%v\n",
				game.Row, game.Title, game.PS4, game.PS5, game.Xbox, game.PC)
		}
		return
	}

	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	if err != nil {
		slog.Error("нет подключения к базе", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()

	importRepo := repository.NewCatalogImportRepo(db)
	if err := importRepo.EnsureSchema(ctx); err != nil {
		slog.Error("схема каталога", slog.String("error", err.Error()))
		os.Exit(1)
	}
	productRepo := repository.NewProductRepo(db)
	settingsRepo := repository.NewSettingsRepo(db)
	if err := service.LoadPricingFormulasFromSettings(ctx, settingsRepo); err != nil {
		slog.Warn("формулы цен не загрузились, работаем на дефолтных", slog.String("error", err.Error()))
	}

	parser := service.NewCatalogParserService(importRepo, settingsRepo)
	curation := service.NewCatalogCurationService(importRepo, productRepo, parser)

	if *audit || *auditApply {
		report, err := curation.AuditCardMatches(ctx, *auditApply)
		if err != nil {
			slog.Error("аудит не прошёл", slog.String("error", err.Error()))
			os.Exit(1)
		}
		slog.Info("аудит привязок",
			slog.Int("проверено", report.Checked),
			slog.Int("неверных", report.Bad),
			slog.Int("скрыто", report.Hidden),
		)
		for _, row := range report.Examples {
			fmt.Printf("  %-40s [%s] → «%s» — %s\n", row.Title, row.Platform, row.StoreTitle, row.Reason)
		}
		return
	}

	if *fixTitles {
		fixed, err := curation.RecanonicalizeProductTitles(ctx)
		if err != nil {
			slog.Error("названия не поправлены", slog.String("error", err.Error()))
			os.Exit(1)
		}
		slog.Info("названия карточек приведены к канону", slog.Int64("исправлено", fixed))
		return
	}

	report, err := parser.ImportWantedGames(ctx, wanted)
	if err != nil {
		slog.Error("импорт завершился с ошибкой", slog.String("error", err.Error()))
	}

	slog.Info("импорт закончен",
		slog.Int("всего", report.Total),
		slog.Int("позиций_импортировано", report.Imported),
		slog.Int("ps", report.MatchedPS),
		slog.Int("xbox", report.MatchedXbox),
		slog.Int("не_найдено", len(report.NotFound)),
		slog.Int("индекс_ps", report.PSIndexSize),
	)

	if *publish {
		if fixed, fixErr := curation.RecanonicalizeProductTitles(ctx); fixErr == nil && fixed > 0 {
			slog.Info("названия карточек приведены к канону", slog.Int64("исправлено", fixed))
		}
		result, err := curation.PublishWantedImports(ctx, 5000)
		if err != nil {
			slog.Error("публикация не удалась", slog.String("error", err.Error()))
		} else {
			slog.Info("опубликовано",
				slog.Int64("новых_карточек", result.Published),
				slog.Int64("привязано_к_существующим", result.LinkedExisting),
				slog.Int64("отклонено", result.ImportsRejected),
			)
		}
	}

	if *reportPath != "" {
		data, _ := json.MarshalIndent(report, "", "  ")
		if err := os.WriteFile(*reportPath, data, 0o644); err != nil {
			slog.Error("отчёт не сохранён", slog.String("error", err.Error()))
		} else {
			slog.Info("отчёт сохранён", slog.String("file", *reportPath))
		}
	}
}
