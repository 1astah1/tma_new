package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

// WantedListPath — путь к списку желаемых игр внутри контейнера.
const WantedListPath = "seeds/wanted_games.csv"

type wantedRunState struct {
	mu     sync.RWMutex
	report *WantedImportReport
	err    string
}

var wantedRun wantedRunState

// LastWantedImportReport — итог последнего прогона импорта по списку.
func LastWantedImportReport() (*WantedImportReport, string) {
	wantedRun.mu.RLock()
	defer wantedRun.mu.RUnlock()
	return wantedRun.report, wantedRun.err
}

// RunWantedImportAsync запускает импорт по списку в фоне: полный проход
// занимает десятки минут, HTTP-запрос столько не живёт.
func (s *CatalogParserService) RunWantedImportAsync(ctx context.Context, path string, publish func(context.Context) error) error {
	if path == "" {
		path = WantedListPath
	}

	handle, err := os.Open(path)
	if err != nil {
		return err
	}
	wanted, err := ParseWantedGamesCSV(handle)
	handle.Close()
	if err != nil {
		return err
	}
	if len(wanted) == 0 {
		return errors.New("список пуст")
	}

	s.mu.Lock()
	if s.status.Running {
		s.mu.Unlock()
		return errors.New("парсер уже работает")
	}
	now := time.Now()
	s.status = CatalogParserStatus{
		Running:       true,
		StartedAt:     &now,
		Sources:       []string{"wanted_list"},
		CurrentSource: "Список",
		CurrentStage:  "Подготовка",
	}
	s.mu.Unlock()

	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	s.cancel = cancel

	go func() {
		defer cancel()
		report, err := s.ImportWantedGames(runCtx, wanted)

		wantedRun.mu.Lock()
		wantedRun.report = report
		wantedRun.err = ""
		if err != nil {
			wantedRun.err = err.Error()
		}
		wantedRun.mu.Unlock()

		if err == nil && publish != nil {
			if publishErr := publish(runCtx); publishErr != nil {
				slog.Warn("публикация после импорта по списку не удалась", slog.String("error", publishErr.Error()))
			}
		}

		finished := time.Now()
		s.mu.Lock()
		s.status.Running = false
		s.status.FinishedAt = &finished
		s.status.CurrentStage = "Готово"
		if report != nil {
			s.status.Imported = report.Imported
		}
		if err != nil {
			s.status.Errors = append(s.status.Errors, err.Error())
		}
		s.mu.Unlock()
	}()

	return nil
}

// RecanonicalizeProductTitles приводит названия уже созданных карточек к тому
// же виду, что и у новых. Нужно после правок CanonicalGameTitle: карточки,
// заведённые прошлыми прогонами, иначе так и остаются с «Cross-Gen Bundle»
// и турецкими хвостами в названии.
func (s *CatalogCurationService) RecanonicalizeProductTitles(ctx context.Context) (int64, error) {
	products, err := s.products.ListGamesForTitleFix(ctx)
	if err != nil {
		return 0, err
	}

	var fixed int64
	for _, product := range products {
		canonical := CanonicalGameTitle(product.Title)
		if canonical == "" {
			continue
		}
		key := CatalogTitleKey(canonical)
		// Ключ пересчитываем всегда: название могло остаться прежним, а
		// разъехавшийся ключ как раз и разводит платформы по разным карточкам.
		if canonical == product.Title && key == product.TitleKey {
			continue
		}
		if err := s.products.UpdateTitle(ctx, product.ID, canonical, key); err != nil {
			slog.Warn("не удалось поправить название",
				slog.String("title", product.Title), slog.String("error", err.Error()))
			continue
		}
		fixed++
	}
	return fixed, nil
}

type MatchAuditRow struct {
	Title      string `json:"title"`
	Platform   string `json:"platform"`
	StoreTitle string `json:"store_title"`
	Reason     string `json:"reason"`
}

type MatchAuditReport struct {
	Checked  int             `json:"checked"`
	Bad      int             `json:"bad"`
	Hidden   int             `json:"hidden"`
	Examples []MatchAuditRow `json:"examples"`
}

// AuditCardMatches проверяет, что за карточкой в сторе лежит именно эта игра,
// а не дополнение к ней. Первые прогоны импорта отработали со слабым правилом
// сравнения названий, и часть карточек до сих пор ведёт на DLC — цена там
// в разы ниже, покупатель получит не то, что ждёт.
func (s *CatalogCurationService) AuditCardMatches(ctx context.Context, apply bool) (*MatchAuditReport, error) {
	rows, err := s.products.ListMatchesForAudit(ctx)
	if err != nil {
		return nil, err
	}

	report := &MatchAuditReport{Checked: len(rows)}
	for _, row := range rows {
		if strings.TrimSpace(row.StoreTitle) == "" {
			continue
		}

		reason := matchProblem(row.Title, row.StoreTitle)
		if reason == "" {
			continue
		}

		report.Bad++
		if len(report.Examples) < 40 {
			report.Examples = append(report.Examples, MatchAuditRow{
				Title: row.Title, Platform: row.Platform, StoreTitle: row.StoreTitle, Reason: reason,
			})
		}
		if apply {
			if err := s.products.DeactivateByID(ctx, row.ID); err != nil {
				slog.Warn("не удалось скрыть карточку", slog.String("title", row.Title), slog.String("error", err.Error()))
				continue
			}
			_ = s.imports.MarkRejected(ctx, row.ImportID)
			report.Hidden++
		}
	}
	return report, nil
}
