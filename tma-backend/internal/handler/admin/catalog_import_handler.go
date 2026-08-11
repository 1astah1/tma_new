package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type CatalogImportHandler struct {
	importRepo  *repository.CatalogImportRepo
	productRepo *repository.ProductRepo
	productSvc  *service.ProductService
	parserSvc   *service.CatalogParserService
	curationSvc *service.CatalogCurationService
	rebuildSvc  *service.CatalogRebuildService
	vitrinaSvc  *service.VitrinaService
}

func NewCatalogImportHandler(importRepo *repository.CatalogImportRepo, productRepo *repository.ProductRepo, productSvc *service.ProductService, parserSvc *service.CatalogParserService, curationSvc *service.CatalogCurationService, rebuildSvc *service.CatalogRebuildService, vitrinaSvc *service.VitrinaService) *CatalogImportHandler {
	return &CatalogImportHandler{
		importRepo:  importRepo,
		productRepo: productRepo,
		productSvc:  productSvc,
		parserSvc:   parserSvc,
		curationSvc: curationSvc,
		rebuildSvc:  rebuildSvc,
		vitrinaSvc:  vitrinaSvc,
	}
}

func (h *CatalogImportHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := repository.CatalogImportFilter{
		Search:      q.Get("search"),
		Source:      q.Get("source"),
		Platform:    q.Get("platform"),
		Status:      q.Get("status"),
		GameSection: q.Get("game_section"),
		Publisher:   q.Get("publisher"),
		Page:        1,
		Limit:       25,
	}
	if v := q.Get("page"); v != "" {
		if page, err := strconv.Atoi(v); err == nil {
			filter.Page = page
		}
	}
	if v := q.Get("limit"); v != "" {
		if limit, err := strconv.Atoi(v); err == nil {
			filter.Limit = limit
		}
	}
	if v := q.Get("release_year"); v != "" {
		if year, err := strconv.Atoi(v); err == nil {
			filter.ReleaseYear = year
		}
	}

	items, total, err := h.importRepo.List(r.Context(), filter)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	productIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		if item.ProductID != nil {
			productIDs = append(productIDs, *item.ProductID)
		}
	}
	statuses, _ := h.productRepo.StatusByIDs(r.Context(), productIDs)

	type importRow struct {
		domain.CatalogImport
		ProductStatus *string `json:"product_status"`
	}
	rows := make([]importRow, len(items))
	for i, item := range items {
		rows[i].CatalogImport = item
		if item.ProductID != nil {
			if status, ok := statuses[*item.ProductID]; ok {
				rows[i].ProductStatus = &status
			}
		}
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": rows,
		"meta": map[string]interface{}{
			"page":  filter.Page,
			"limit": filter.Limit,
			"total": total,
		},
	})
}

func (h *CatalogImportHandler) Summary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.importRepo.Summary(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusOK, summary)
}

func (h *CatalogImportHandler) FilterOptions(w http.ResponseWriter, r *http.Request) {
	backfill := r.URL.Query().Get("backfill") != "0" && r.URL.Query().Get("backfill") != "false"
	if backfill {
		_ = h.importRepo.BackfillImportMetadata(r.Context())
	}
	opts, err := h.importRepo.GetFilterOptions(r.Context(), false)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	opts.Backfilled = backfill
	handler.RespondJSON(w, http.StatusOK, opts)
}

func (h *CatalogImportHandler) BackfillMetadata(w http.ResponseWriter, r *http.Request) {
	syncMode := r.URL.Query().Get("sync") == "1" || r.URL.Query().Get("sync") == "true"
	run := func(ctx context.Context) (int, error) {
		if err := h.importRepo.BackfillImportMetadata(ctx); err != nil {
			return 0, err
		}
		return h.parserSvc.EnrichImportMetadata(ctx)
	}

	if syncMode {
		updated, err := run(r.Context())
		if err != nil {
			handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
		opts, err := h.importRepo.GetFilterOptions(r.Context(), false)
		if err != nil {
			handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
		handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"publishers":    opts.Publishers,
			"release_years": opts.ReleaseYears,
			"updated":       updated,
			"status":        "done",
		})
		return
	}

	go func() {
		ctx := context.WithoutCancel(r.Context())
		_, _ = run(ctx)
	}()
	handler.RespondJSON(w, http.StatusAccepted, map[string]string{"status": "started"})
}

func (h *CatalogImportHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid ID")
		return
	}
	item, err := h.importRepo.GetByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Catalog import not found")
		return
	}
	handler.RespondJSON(w, http.StatusOK, item)
}

func (h *CatalogImportHandler) RunParser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Full bool `json:"full"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := h.parserSvc.RunAsync(r.Context(), req.Full); err != nil {
		handler.RespondError(w, http.StatusConflict, "PARSER_RUNNING", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusAccepted, h.parserSvc.Status())
}

// ImportWantedList — импорт по списку желаемых игр: цены по гео, описания,
// картинки и раздел берутся из магазинов, список задаёт только состав.
func (h *CatalogImportHandler) ImportWantedList(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	publish := func(ctx context.Context) error {
		_, err := h.curationSvc.PublishWantedImports(ctx, 5000)
		return err
	}

	if err := h.parserSvc.RunWantedImportAsync(r.Context(), req.Path, publish); err != nil {
		handler.RespondError(w, http.StatusConflict, "IMPORT_FAILED", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusAccepted, h.parserSvc.Status())
}

// WantedListReport — итог последнего прогона: что нашлось, что нет.
func (h *CatalogImportHandler) WantedListReport(w http.ResponseWriter, r *http.Request) {
	report, errText := service.LastWantedImportReport()
	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"report": report,
		"error":  errText,
		"status": h.parserSvc.Status(),
	})
}

func (h *CatalogImportHandler) ParserStatus(w http.ResponseWriter, r *http.Request) {
	handler.RespondJSON(w, http.StatusOK, h.parserSvc.Status())
}

func (h *CatalogImportHandler) ResetCatalog(w http.ResponseWriter, r *http.Request) {
	h.parserSvc.Stop()
	result, err := h.importRepo.ResetGameCatalog(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusOK, result)
}

func (h *CatalogImportHandler) ActivateAllGames(w http.ResponseWriter, r *http.Request) {
	count, err := h.productSvc.ActivateAllGames(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{"activated": count})
}

func (h *CatalogImportHandler) Deduplicate(w http.ResponseWriter, r *http.Request) {
	result, err := h.curationSvc.Deduplicate(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusOK, result)
}

func (h *CatalogImportHandler) AutoPublishFresh(w http.ResponseWriter, r *http.Request) {
	pub, err := h.vitrinaSvc.RepublishPending(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	activated, _ := h.productSvc.ActivateAllGames(r.Context())
	synced, _ := h.productSvc.SyncMetadataFromImports(r.Context(), service.MinPaidPriceRUB())
	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"published":        pub.Published,
		"linked_existing":  pub.LinkedExisting,
		"activated":        activated,
		"products_synced":  synced,
		"imports_rejected": pub.Rejected,
		"scores_updated":   pub.ScoresUpdated,
	})
}

func (h *CatalogImportHandler) ImportPSStore(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithoutCancel(r.Context())
	slog.Info("PS Store light import starting in background")
	go func() {
		defer func() {
			if rc := recover(); rc != nil {
				slog.Error("PS Store import panic", slog.Any("panic", rc))
			}
		}()
		imported, err := h.parserSvc.RunPSImportLight(ctx)
		if err != nil {
			slog.Warn("PS Store import failed", slog.String("error", err.Error()))
			return
		}
		slog.Info("PS Store import phase 1 done", slog.Int("imported", imported))
		enriched, _ := h.parserSvc.EnrichImportMetadata(ctx)
		slog.Info("PS Store import phase 2 done", slog.Int("enriched", enriched))
		rejected, _ := h.curationSvc.RejectUnsellableImports(ctx)
		slog.Info("PS Store import phase 3 done", slog.Int64("rejected", rejected))
		pub, _ := h.vitrinaSvc.PublishAll(ctx)
		slog.Info("PS Store import phase 4 done", slog.Int64("published", pub.Published))
		synced, _ := h.productSvc.SyncMetadataFromImports(ctx, service.MinPaidPriceRUB())
		_, _ = h.vitrinaSvc.UpdateAllScores(ctx)
		slog.Info("PS Store import finished",
			slog.Int("imported", imported),
			slog.Int("enriched", enriched),
			slog.Int64("imports_rejected", rejected),
			slog.Int64("published", pub.Published),
			slog.Int64("products_synced", synced),
		)
	}()
	handler.RespondJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":          "started",
		"imported":        0,
		"published":       0,
		"products_synced": 0,
	})
}

func (h *CatalogImportHandler) ImportXbox(w http.ResponseWriter, r *http.Request) {
	imported, err := h.parserSvc.RunXboxImportSync(r.Context(), false)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	enriched, _ := h.parserSvc.EnrichXboxOnly(r.Context())
	rejected, _ := h.curationSvc.RejectUnsellableImports(r.Context())
	pub, _ := h.vitrinaSvc.PublishAll(r.Context())
	synced, _ := h.productSvc.SyncMetadataFromImports(r.Context(), service.MinPaidPriceRUB())
	_, _ = h.vitrinaSvc.UpdateAllScores(r.Context())
	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"imported":         imported,
		"enriched":         enriched,
		"imports_rejected": rejected,
		"published":        pub.Published,
		"linked_existing":  pub.LinkedExisting,
		"products_synced":  synced,
	})
}

func (h *CatalogImportHandler) RefreshCatalog(w http.ResponseWriter, r *http.Request) {
	syncMode := r.URL.Query().Get("sync") == "1" || r.URL.Query().Get("sync") == "true"
	if syncMode {
		result, err := h.curationSvc.RefreshCatalog(r.Context())
		if err != nil {
			handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
		handler.RespondJSON(w, http.StatusOK, result)
		return
	}

	go func() {
		ctx := context.WithoutCancel(r.Context())
		result, err := h.curationSvc.RefreshCatalog(ctx)
		if err != nil {
			slog.Warn("catalog refresh failed", slog.String("error", err.Error()))
			return
		}
		slog.Info("catalog refresh finished",
			slog.Int("enriched", result.Enriched),
			slog.Int64("products_synced", result.ProductsSynced),
			slog.Int64("products_hidden", result.ProductsHidden),
			slog.Int64("published", result.Published),
			slog.Int64("imports_rejected", result.ImportsRejected),
		)
	}()
	handler.RespondJSON(w, http.StatusAccepted, map[string]string{"status": "started"})
}

func (h *CatalogImportHandler) RebuildCatalog(w http.ResponseWriter, r *http.Request) {
	syncMode := r.URL.Query().Get("sync") == "1" || r.URL.Query().Get("sync") == "true"
	if syncMode {
		result, err := h.rebuildSvc.RunSync(r.Context())
		if err != nil {
			handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
		handler.RespondJSON(w, http.StatusOK, result)
		return
	}
	if err := h.rebuildSvc.RunAsync(r.Context()); err != nil {
		handler.RespondError(w, http.StatusConflict, "REBUILD_RUNNING", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusAccepted, h.rebuildSvc.Status())
}

func (h *CatalogImportHandler) RebuildStatus(w http.ResponseWriter, r *http.Request) {
	handler.RespondJSON(w, http.StatusOK, h.rebuildSvc.Status())
}

func (h *CatalogImportHandler) RepublishPending(w http.ResponseWriter, r *http.Request) {
	result, err := h.vitrinaSvc.RepublishPending(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusOK, result)
}

func (h *CatalogImportHandler) SyncCatalog(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Full bool `json:"full"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	go func() {
		ctx := context.WithoutCancel(r.Context())
		result, err := h.curationSvc.SyncCatalog(ctx, req.Full)
		if err != nil {
			slog.Warn("catalog sync failed", slog.String("error", err.Error()))
			return
		}
		slog.Info("catalog sync finished",
			slog.Int64("published", result.Published),
			slog.Int64("imports_rejected", result.ImportsRejected),
			slog.Int64("products_deleted", result.ProductsDeleted),
		)
	}()
	handler.RespondJSON(w, http.StatusAccepted, map[string]string{"status": "started"})
}

func (h *CatalogImportHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid ID")
		return
	}

	item, err := h.importRepo.GetByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Catalog import not found")
		return
	}
	if item.Status != domain.CatalogImportPending {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_STATUS", "Only pending imports can be approved")
		return
	}

	var req struct {
		Platform string   `json:"platform"`
		Price    *float64 `json:"price"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if !service.IsSellableCatalogItem(item) {
		handler.RespondError(w, http.StatusBadRequest, "UNSALEABLE", "Бесплатные, демо и мусорные позиции не публикуются")
		return
	}

	platform := service.PickProductPlatform(item.Platforms, req.Platform)
	price := service.EffectiveCatalogPriceRUB(item)
	if req.Price != nil && *req.Price >= service.MinPaidPriceRUB() {
		price = *req.Price
	}
	if price < service.MinPaidPriceRUB() {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_PRICE", "Нужна актуальная цена из магазина (мин. 149 ₽)")
		return
	}

	titleKey := item.TitleKey
	if titleKey == "" {
		titleKey = service.NormalizeGameTitle(item.Title)
	}

	product := &domain.Product{
		Title:           item.Title,
		TitleKey:        titleKey,
		Description:     item.Description,
		Platform:        domain.Platform(platform),
		Type:            domain.ProductTypeGame,
		GameSection:     item.GameSection,
		ReleaseDate:     item.ReleaseDate,
		Price:           price,
		DiscountPercent: 0,
		Variants:        json.RawMessage("[]"),
		ImageURL:        item.ImageURL,
		DeliveryMethods: pq.StringArray{string(domain.DeliveryMethodActivation)},
		Prices:          item.Prices,
		Status:          domain.ProductStatusActive,
	}

	if err := h.productSvc.Create(r.Context(), product); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if err := h.importRepo.MarkApproved(r.Context(), item.ID, product.ID); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"import":  item,
		"product": product,
	})
}

func (h *CatalogImportHandler) ResetRejected(w http.ResponseWriter, r *http.Request) {
	count, err := h.importRepo.ResetRejectedImports(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"count":   count,
		"message": fmt.Sprintf("Сброшено %d отклонённых импортов в статус pending", count),
	})
}

func (h *CatalogImportHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid ID")
		return
	}
	if err := h.importRepo.MarkRejected(r.Context(), id); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}
