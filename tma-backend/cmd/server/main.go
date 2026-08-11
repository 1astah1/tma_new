package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"tma-backend/internal/bot"
	"tma-backend/internal/config"
	"tma-backend/internal/dbmigrate"
	"tma-backend/internal/domain"
	h "tma-backend/internal/handler"
	"tma-backend/internal/handler/admin"
	"tma-backend/internal/handler/public"
	"tma-backend/internal/logger"
	"tma-backend/internal/middleware"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.App.Environment)

	// Database connection
	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	if err != nil {
		slog.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	slog.Info("Connected to database")

	if cfg.App.AutoMigrate {
		if err := dbmigrate.Run(db); err != nil {
			slog.Error("Failed to apply migrations", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	// Repositories
	userRepo := repository.NewUserRepo(db)
	productRepo := repository.NewProductRepo(db)
	orderRepo := repository.NewOrderRepo(db)
	keyRepo := repository.NewKeyRepo(db)
	adminRepo := repository.NewAdminRepo(db)
	accountRepo := repository.NewAccountRepo(db)
	settingsRepo := repository.NewSettingsRepo(db)
	if err := service.LoadPricingFormulasFromSettings(context.Background(), settingsRepo); err != nil {
		slog.Warn("pricing formulas load failed", slog.String("error", err.Error()))
	}
	promoRepo := repository.NewPromoCodeRepo(db)
	chatRepo := repository.NewChatRepo(db)
	templateRepo := repository.NewTemplateRepo(db)
	catalogImportRepo := repository.NewCatalogImportRepo(db)
	if err := catalogImportRepo.EnsureSchema(context.Background()); err != nil {
		slog.Error("Failed to ensure catalog import schema", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Services
	encSvc := service.NewEncryptionService(cfg.Telegram.EncryptKey)
	authSvc := service.NewAuthService(cfg, userRepo, adminRepo)
	notifSvc := service.NewNotificationService(cfg.Telegram.BotToken)
	notifSvc.SetTMAURL(cfg.App.TMAURL)
	auditSvc := service.NewAuditService(adminRepo)
	englishTitles := service.NewEnglishTitleResolver(func(ctx context.Context, productID uuid.UUID) (string, string, bool) {
		return productRepo.GetImportLinkByProductID(ctx, productID)
	})
	productSvc := service.NewProductService(productRepo, englishTitles)
	promoSvc := service.NewPromoService(promoRepo)
	orderSvc := service.NewOrderService(db, orderRepo, productRepo, keyRepo, accountRepo, userRepo, chatRepo, promoSvc, encSvc, notifSvc, auditSvc)
	catalogParserSvc := service.NewCatalogParserService(catalogImportRepo, settingsRepo)
	catalogCurationSvc := service.NewCatalogCurationService(catalogImportRepo, productRepo, catalogParserSvc)
	vitrinaSvc := service.NewVitrinaService(catalogImportRepo, productRepo, settingsRepo)
	catalogRebuildSvc := service.NewCatalogRebuildService(catalogImportRepo, productRepo, settingsRepo, catalogParserSvc, catalogCurationSvc, vitrinaSvc)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		updated, err := vitrinaSvc.UpdateAllScores(ctx)
		if err != nil {
			slog.Warn("vitrina score recalc failed", slog.String("error", err.Error()))
			return
		}
		slog.Info("vitrina scores recalculated", slog.Int64("updated", updated))
	}()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
		defer cancel()
		if err := catalogImportRepo.BackfillImportMetadata(ctx); err != nil {
			slog.Warn("catalog reclassify failed", slog.String("error", err.Error()))
		}
		if _, err := catalogImportRepo.ClearInvalidReleaseDates(ctx); err != nil {
			slog.Warn("clear invalid import dates failed", slog.String("error", err.Error()))
		}
		if _, err := productRepo.ClearInvalidReleaseDates(ctx); err != nil {
			slog.Warn("clear invalid product dates failed", slog.String("error", err.Error()))
		}
		result, err := catalogCurationSvc.AutoPublishFresh(ctx, 500)
		if err != nil {
			slog.Warn("catalog auto publish failed", slog.String("error", err.Error()))
			return
		}
		slog.Info("catalog curation finished",
			slog.Int64("published", result.Published),
			slog.Int64("imports_rejected", result.ImportsRejected),
			slog.Int64("products_deleted", result.ProductsDeleted),
		)
	}()

	go func() {
		ctx := context.Background()
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			runCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
			if _, err := catalogCurationSvc.AutoPublishFresh(runCtx, 300); err != nil {
				slog.Warn("scheduled catalog curation failed", slog.String("error", err.Error()))
			}
			cancel()
		}
	}()

	// Start background workers
	go func() {
		ctx := context.Background()
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			orderSvc.Expire2FACodes(ctx)
			orderSvc.ExpireUnpaidOrders(ctx)
		}
	}()

	// Init bot
	tgBot := bot.NewBot(cfg.Telegram.BotToken, db, adminRepo, orderRepo, userRepo, notifSvc, orderSvc, cfg.App.TMAURL, fmt.Sprintf("http://localhost:%s", cfg.Server.Port))

	// Ensure upload dir exists
	os.MkdirAll(cfg.UploadDir, 0755)

	// Router
	r := chi.NewRouter()

	// Metrics middleware (applied to all routes)
	r.Use(middleware.Metrics)

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Recover())
	r.Use(middleware.Logging)
	if cfg.App.Environment == "development" {
		r.Use(middleware.CORS("*"))
	} else {
		r.Use(middleware.CORS(cfg.App.TMAURL, cfg.App.AdminURL))
	}

	// Rate limiter: 10 requests per second, burst of 20
	rateLimiter := middleware.NewRateLimiter(10, 20)

	// Health checks
	healthHandler := h.NewHealthHandler(db)
	r.Get("/health", healthHandler.Liveness)
	r.Get("/health/ready", healthHandler.Readiness)

	// Prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

	// Public API (TMA)
	r.Route("/api/v1", func(r chi.Router) {
		// Auth
		r.Post("/auth/telegram", func(w http.ResponseWriter, r *http.Request) {
			rateLimiter.Middleware(http.HandlerFunc(handleTelegramAuth(authSvc, cfg))).ServeHTTP(w, r)
		})

		// Public
		productHandler := public.NewProductHandler(productSvc, productRepo)
		r.Get("/products", productHandler.List)
		r.Get("/products/{id}", productHandler.GetByID)
		r.Get("/platforms", productHandler.GetPlatforms)
		contentHandler := public.NewContentHandler(productRepo, settingsRepo, vitrinaSvc)
		r.Get("/content/popular-products", contentHandler.PopularProducts)
		r.Get("/content/home-feed", contentHandler.HomeFeed)
		r.Get("/content/home-categories", contentHandler.HomeCategories)
		r.Get("/content/home-categories/{id}", contentHandler.HomeCategoryByID)
		r.Get("/content/home-banners", contentHandler.HomeBanners)
		r.Get("/content/shop-settings", contentHandler.ShopSettings)

		// Payment details (public)
		publicPaymentHandler := public.NewProfileHandler(userRepo, orderRepo, settingsRepo, adminRepo)
		r.Get("/payments/details", publicPaymentHandler.GetPaymentDetails)

		// Promo codes
		promoHandler := public.NewPromoHandler(promoSvc)
		r.Post("/promo/validate", promoHandler.Validate)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.UserAuth(authSvc, userRepo, cfg.App.Environment))

			orderHandler := public.NewOrderHandler(orderSvc, cfg.UploadDir)
			r.Post("/orders", func(w http.ResponseWriter, r *http.Request) {
				rateLimiter.Middleware(http.HandlerFunc(orderHandler.Create)).ServeHTTP(w, r)
			})
			r.Post("/orders/batch", func(w http.ResponseWriter, r *http.Request) {
				rateLimiter.Middleware(http.HandlerFunc(orderHandler.CreateBatch)).ServeHTTP(w, r)
			})
			r.Get("/orders", orderHandler.List)
			r.Get("/orders/{id}", orderHandler.GetByID)
			r.Post("/orders/{id}/credentials", func(w http.ResponseWriter, r *http.Request) {
				rateLimiter.Middleware(http.HandlerFunc(orderHandler.SendCredentials)).ServeHTTP(w, r)
			})
			r.Post("/orders/{id}/2fa-code", func(w http.ResponseWriter, r *http.Request) {
				rateLimiter.Middleware(http.HandlerFunc(orderHandler.Send2FACode)).ServeHTTP(w, r)
			})
			r.Post("/orders/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
				rateLimiter.Middleware(http.HandlerFunc(orderHandler.CancelOrder)).ServeHTTP(w, r)
			})
			r.Post("/orders/{id}/refund", func(w http.ResponseWriter, r *http.Request) {
				rateLimiter.Middleware(http.HandlerFunc(orderHandler.RequestRefund)).ServeHTTP(w, r)
			})
			r.Get("/orders/{id}/chat", orderHandler.GetChatMessages)
			r.Post("/orders/{id}/chat", func(w http.ResponseWriter, r *http.Request) {
				rateLimiter.Middleware(http.HandlerFunc(orderHandler.SendChatMessage)).ServeHTTP(w, r)
			})
			r.Post("/orders/{id}/payment", func(w http.ResponseWriter, r *http.Request) {
				rateLimiter.Middleware(http.HandlerFunc(orderHandler.ConfirmPayment)).ServeHTTP(w, r)
			})
			r.Post("/orders/batch/payment", func(w http.ResponseWriter, r *http.Request) {
				rateLimiter.Middleware(http.HandlerFunc(orderHandler.ConfirmBatchPayment)).ServeHTTP(w, r)
			})

			profileHandler := public.NewProfileHandler(userRepo, orderRepo, settingsRepo, adminRepo)
			r.Get("/profile", profileHandler.GetProfile)
			r.Get("/payment-details", profileHandler.GetPaymentDetails)
		})

		// Админка внутри мини-аппа: те же обработчики, но вход по обычному
		// токену Telegram — владельцу удобнее вести заказы с телефона.
		r.Route("/tma-admin", func(r chi.Router) {
			r.Use(middleware.TMAAdminAuth(authSvc, userRepo, adminRepo))

			tmaOrders := admin.NewOrderHandler(orderSvc, orderRepo, adminRepo, encSvc, templateRepo)
			r.Get("/orders", tmaOrders.List)
			r.Get("/orders/{id}", tmaOrders.GetByID)
			r.Patch("/orders/{id}/status", tmaOrders.UpdateStatus)
			r.Get("/orders/{id}/chat", tmaOrders.GetChatMessages)
			r.Post("/orders/{id}/chat", tmaOrders.SendChatMessage)

			tmaDashboard := admin.NewDashboardHandler(orderSvc, catalogImportRepo)
			r.Get("/dashboard", tmaDashboard.GetStats)

			tmaProducts := admin.NewAdminProductHandler(productSvc, productRepo, catalogImportRepo)
			r.Get("/products", tmaProducts.List)
			r.Get("/products/{id}", tmaProducts.GetByID)
			r.Put("/products/{id}", tmaProducts.Update)
			r.Delete("/products/{id}", tmaProducts.Delete)

			tmaSettings := admin.NewSettingsHandler(settingsRepo)
			r.Get("/settings", tmaSettings.Get)
			r.Put("/settings", tmaSettings.Upsert)
			r.Get("/settings/pricing-preview", tmaSettings.PricingPreview)

			tmaUsers := admin.NewAdminUserHandler(userRepo, adminRepo, accountRepo)
			r.Get("/users", tmaUsers.ListUsers)
			r.Get("/users/{id}", tmaUsers.GetUser)
			r.Patch("/users/{id}", tmaUsers.UpdateUser)
			r.Get("/logs", tmaUsers.GetLogs)

			tmaPromos := admin.NewPromoHandler(promoSvc)
			r.Get("/promos", tmaPromos.List)
			r.Post("/promos", tmaPromos.Create)
			r.Put("/promos/{id}", tmaPromos.Update)
			r.Delete("/promos/{id}", tmaPromos.Delete)

			tmaCatalog := admin.NewCatalogImportHandler(catalogImportRepo, productRepo, productSvc, catalogParserSvc, catalogCurationSvc, catalogRebuildSvc, vitrinaSvc)
			r.Get("/catalog-imports/summary", tmaCatalog.Summary)
			r.Post("/catalog-imports/import-wanted", tmaCatalog.ImportWantedList)
			r.Get("/catalog-imports/wanted-report", tmaCatalog.WantedListReport)
			r.Get("/catalog-parser/status", tmaCatalog.ParserStatus)
			r.Post("/catalog-imports/auto-publish-fresh", tmaCatalog.AutoPublishFresh)
			r.Post("/catalog-imports/deduplicate", tmaCatalog.Deduplicate)

			tmaBroadcast := admin.NewBroadcastHandler(userRepo, notifSvc)
			r.Post("/broadcast", tmaBroadcast.Send)

			// Состав команды правит только владелец.
			r.Group(func(r chi.Router) {
				r.Use(middleware.SuperAdminOnly)
				r.Get("/admins", tmaUsers.ListAdmins)
				r.Post("/admins", tmaUsers.CreateAdmin)
				r.Put("/admins/{id}", tmaUsers.UpdateAdmin)
			})
		})

		// Public FAQ (no auth required)
		faqHandler := public.NewFAQHandler(templateRepo)
		r.Get("/faq", faqHandler.ListFAQ)
		r.Get("/faq/answer", faqHandler.GetFAQAnswer)
	})

	// Serve uploaded files
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

	// Admin API
	r.Route("/api/v1/admin", func(r chi.Router) {
		bruteForce := middleware.NewBruteForceProtector(5, 15*time.Minute)
		authHandler := admin.NewAuthHandler(authSvc, bruteForce)
		r.Post("/auth/login", func(w http.ResponseWriter, r *http.Request) {
			bruteForce.Middleware(http.HandlerFunc(authHandler.Login)).ServeHTTP(w, r)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.AdminAuth(authSvc, cfg.App.Environment))

			r.Post("/upload", handleFileUpload(cfg.UploadDir))
			r.Get("/auth/me", authHandler.Me)

			// Products (admin CRUD)
			adminProductHandler := admin.NewAdminProductHandler(productSvc, productRepo, catalogImportRepo)
			r.Get("/products", adminProductHandler.List)
			r.Get("/products/{id}", adminProductHandler.GetByID)
			r.Post("/products/{id}/sync-from-import", adminProductHandler.SyncFromImport)
			r.Post("/products", adminProductHandler.Create)
			r.Put("/products/{id}", adminProductHandler.Update)
			r.Delete("/products/{id}", adminProductHandler.Delete)

			// Catalog imports (admin approval queue)
			catalogImportHandler := admin.NewCatalogImportHandler(catalogImportRepo, productRepo, productSvc, catalogParserSvc, catalogCurationSvc, catalogRebuildSvc, vitrinaSvc)
			r.Get("/catalog-imports", catalogImportHandler.List)
			r.Get("/catalog-imports/summary", catalogImportHandler.Summary)
			r.Get("/catalog-imports/filter-options", catalogImportHandler.FilterOptions)
			r.Post("/catalog-imports/backfill-metadata", catalogImportHandler.BackfillMetadata)
			r.Post("/catalog-imports/deduplicate", catalogImportHandler.Deduplicate)
			r.Post("/catalog-imports/auto-publish-fresh", catalogImportHandler.AutoPublishFresh)
			r.Post("/catalog-imports/refresh-catalog", catalogImportHandler.RefreshCatalog)
			r.Post("/catalog-imports/rebuild", catalogImportHandler.RebuildCatalog)
			r.Get("/catalog-imports/rebuild-status", catalogImportHandler.RebuildStatus)
			r.Post("/catalog-imports/republish-pending", catalogImportHandler.RepublishPending)
			r.Post("/catalog-imports/import-ps", catalogImportHandler.ImportPSStore)
			r.Post("/catalog-imports/import-xbox", catalogImportHandler.ImportXbox)
			r.Post("/catalog-imports/sync-catalog", catalogImportHandler.SyncCatalog)
			r.Get("/catalog-imports/{id}", catalogImportHandler.GetByID)
			r.Post("/catalog-imports/reset-rejected", catalogImportHandler.ResetRejected)
			r.Post("/catalog-imports/{id}/approve", catalogImportHandler.Approve)
			r.Post("/catalog-imports/{id}/reject", catalogImportHandler.Reject)
			r.Get("/catalog-parser/status", catalogImportHandler.ParserStatus)
			r.Post("/catalog-parser/run", catalogImportHandler.RunParser)
			r.Post("/catalog-imports/import-wanted", catalogImportHandler.ImportWantedList)
			r.Get("/catalog-imports/wanted-report", catalogImportHandler.WantedListReport)
			r.Post("/catalog-parser/reset", catalogImportHandler.ResetCatalog)
			r.Post("/catalog-parser/activate-all-games", catalogImportHandler.ActivateAllGames)

			// Dashboard
			dashHandler := admin.NewDashboardHandler(orderSvc, catalogImportRepo)
			r.Get("/dashboard", dashHandler.GetStats)

			// Orders
			orderHandler := admin.NewOrderHandler(orderSvc, orderRepo, adminRepo, encSvc, templateRepo)
			r.Get("/orders", orderHandler.List)
			r.Get("/orders/{id}", orderHandler.GetByID)
			r.Patch("/orders/{id}/status", orderHandler.UpdateStatus)
			r.Patch("/orders/bulk-status", orderHandler.BulkUpdateStatus)
			r.Post("/orders/{id}/decrypt-credentials", orderHandler.DecryptCredentials)
			r.Get("/orders/{id}/chat", orderHandler.GetChatMessages)
			r.Post("/orders/{id}/chat", orderHandler.SendChatMessage)
			r.Post("/orders/{id}/chat/template", orderHandler.SendTemplateMessage)

			// Users & Admins
			userHandler := admin.NewAdminUserHandler(userRepo, adminRepo, accountRepo)
			r.Get("/users", userHandler.ListUsers)
			r.Get("/users/{id}", userHandler.GetUser)
			r.Get("/users/{id}/stats", userHandler.GetUserStats)
			r.Patch("/users/{id}", userHandler.UpdateUser)
			r.Get("/admins", userHandler.ListAdmins)
			r.Post("/admins", userHandler.CreateAdmin)
			r.Put("/admins/{id}", userHandler.UpdateAdmin)

			broadcastHandler := admin.NewBroadcastHandler(userRepo, notifSvc)
			r.Post("/broadcast", broadcastHandler.Send)

			// Settings
			settingsHandler := admin.NewSettingsHandler(settingsRepo)
			r.Get("/settings", settingsHandler.Get)
			r.Get("/settings/pricing-preview", settingsHandler.PricingPreview)
			r.Put("/settings", settingsHandler.Upsert)

			// Promo codes (admin)
			adminPromoHandler := admin.NewPromoHandler(promoSvc)
			r.Get("/promos", adminPromoHandler.List)
			r.Get("/promos/{id}", adminPromoHandler.GetByID)
			r.Post("/promos", adminPromoHandler.Create)
			r.Put("/promos/{id}", adminPromoHandler.Update)
			r.Delete("/promos/{id}", adminPromoHandler.Delete)

			// Templates & FAQ
			templateHandler := admin.NewTemplateHandler(templateRepo)
			r.Get("/templates", templateHandler.ListTemplates)
			r.Get("/templates/{id}", templateHandler.GetTemplate)
			r.Post("/templates", templateHandler.CreateTemplate)
			r.Put("/templates/{id}", templateHandler.UpdateTemplate)
			r.Delete("/templates/{id}", templateHandler.DeleteTemplate)
			r.Get("/faq-items", templateHandler.ListFAQ)
			r.Get("/faq-items/{id}", templateHandler.GetFAQ)
			r.Post("/faq-items", templateHandler.CreateFAQ)
			r.Put("/faq-items/{id}", templateHandler.UpdateFAQ)
			r.Delete("/faq-items/{id}", templateHandler.DeleteFAQ)

			// Logs
			r.Get("/logs", userHandler.GetLogs)
		})
	})

	// Bot webhook + polling
	if cfg.Telegram.BotToken != "" {
		tgBot.RegisterRoutes(r)
		tgBot.StartPolling(context.Background())
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 180 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		slog.Info("Shutting down server...")
		server.Close()
	}()

	slog.Info("Server starting", slog.String("addr", addr))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Server error", slog.String("error", err.Error()))
	}
}

func handleTelegramAuth(authSvc *service.AuthService, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			InitData string `json:"initData"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
			return
		}

		devBypass := cfg.App.Environment == "development"
		user, err := authSvc.AuthenticateUser(r.Context(), req.InitData, devBypass)
		if err != nil {
			if err == domain.ErrForbidden {
				h.RespondError(w, http.StatusForbidden, "FORBIDDEN", "Account is banned")
				return
			}
			msg := "Authentication failed"
			if cfg.App.Environment == "development" {
				msg = err.Error()
			}
			h.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", msg)
			return
		}

		token, err := authSvc.GenerateUserToken(user)
		if err != nil {
			h.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Token generation failed")
			return
		}

		h.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"token": token,
			"user":  user,
		})
	}
}

func handleFileUpload(uploadDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(50 << 20) // 50 MB
		file, header, err := r.FormFile("file")
		if err != nil {
			h.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "file is required")
			return
		}
		defer file.Close()

		ext := filepath.Ext(header.Filename)
		name := uuid.New().String() + ext
		dest := filepath.Join(uploadDir, name)

		out, err := os.Create(dest)
		if err != nil {
			h.RespondError(w, http.StatusInternalServerError, "UPLOAD_ERROR", "failed to save file")
			return
		}
		defer out.Close()

		io.Copy(out, file)
		url := "/uploads/" + name
		h.RespondJSON(w, http.StatusOK, map[string]string{"url": url, "filename": name})
	}
}
