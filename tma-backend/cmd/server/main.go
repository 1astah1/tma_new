package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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

	"tma-backend/internal/bot"
	"tma-backend/internal/config"
	"tma-backend/internal/domain"
	"tma-backend/internal/handler/admin"
	"tma-backend/internal/handler/public"
	h "tma-backend/internal/handler"
	"tma-backend/internal/middleware"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

func main() {
	cfg := config.Load()

	// Database connection
	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Printf("Connected to database: %s", cfg.Database.URL)

	// Repositories
	userRepo := repository.NewUserRepo(db)
	productRepo := repository.NewProductRepo(db)
	orderRepo := repository.NewOrderRepo(db)
	keyRepo := repository.NewKeyRepo(db)
	adminRepo := repository.NewAdminRepo(db)
	accountRepo := repository.NewAccountRepo(db)
	settingsRepo := repository.NewSettingsRepo(db)

	// Services
	encSvc := service.NewEncryptionService(cfg.Telegram.EncryptKey)
	authSvc := service.NewAuthService(cfg, userRepo, adminRepo)
	notifSvc := service.NewNotificationService(cfg.Telegram.BotToken)
	auditSvc := service.NewAuditService(adminRepo)
	productSvc := service.NewProductService(productRepo)
	orderSvc := service.NewOrderService(db, orderRepo, productRepo, keyRepo, accountRepo, userRepo, encSvc, notifSvc, auditSvc)

	// Start background workers
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			orderSvc.Expire2FACodes(nil)
			orderSvc.ExpireUnpaidOrders(nil)
		}
	}()

	// Init bot
	tgBot := bot.NewBot(cfg.Telegram.BotToken)

	// Ensure upload dir exists
	os.MkdirAll(cfg.UploadDir, 0755)

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logging)
	r.Use(middleware.CORS("*"))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		h.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Public API (TMA)
	r.Route("/api/v1", func(r chi.Router) {
		// Auth
		r.Post("/auth/telegram", handleTelegramAuth(authSvc, userRepo))

			// Public
		productHandler := public.NewProductHandler(productSvc)
		r.Get("/products", productHandler.List)
		r.Get("/products/{id}", productHandler.GetByID)
		r.Get("/platforms", productHandler.GetPlatforms)

		// Payment details (public)
		publicPaymentHandler := public.NewProfileHandler(userRepo, orderRepo, settingsRepo)
		r.Get("/payments/details", publicPaymentHandler.GetPaymentDetails)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.UserAuth(authSvc))

			orderHandler := public.NewOrderHandler(orderSvc)
			r.Post("/orders", orderHandler.Create)
			r.Get("/orders", orderHandler.List)
			r.Get("/orders/{id}", orderHandler.GetByID)
			r.Post("/orders/{id}/confirm-payment", orderHandler.ConfirmPayment)
			r.Post("/orders/{id}/credentials", orderHandler.SendCredentials)
			r.Post("/orders/{id}/2fa-code", orderHandler.Send2FACode)

			profileHandler := public.NewProfileHandler(userRepo, orderRepo, settingsRepo)
			r.Get("/profile", profileHandler.GetProfile)
			r.Get("/payment-details", profileHandler.GetPaymentDetails)
		})
	})

	// Serve uploaded files
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))))

	// Admin API
	r.Route("/api/v1/admin", func(r chi.Router) {
		authHandler := admin.NewAuthHandler(authSvc)
		r.Post("/auth/login", authHandler.Login)

		r.Post("/upload", handleFileUpload(cfg.UploadDir))

		r.Group(func(r chi.Router) {
			r.Use(middleware.AdminAuth(authSvc))

			r.Get("/auth/me", authHandler.Me)

			// Products (admin CRUD)
			adminProductHandler := admin.NewAdminProductHandler(productSvc)
			r.Get("/products", adminProductHandler.List)
			r.Get("/products/{id}", adminProductHandler.GetByID)
			r.Post("/products", adminProductHandler.Create)
			r.Put("/products/{id}", adminProductHandler.Update)
			r.Delete("/products/{id}", adminProductHandler.Delete)

			// Dashboard
			dashHandler := admin.NewDashboardHandler(orderSvc)
			r.Get("/dashboard", dashHandler.GetStats)

			// Orders
			orderHandler := admin.NewOrderHandler(orderSvc)
			r.Get("/orders", orderHandler.List)
			r.Get("/orders/{id}", orderHandler.GetByID)
			r.Patch("/orders/{id}/status", orderHandler.UpdateStatus)
			r.Post("/orders/{id}/decrypt-credentials", orderHandler.DecryptCredentials)

			// Users & Admins
			userHandler := admin.NewAdminUserHandler(userRepo, adminRepo, accountRepo)
			r.Get("/users", userHandler.ListUsers)
			r.Get("/users/{id}", userHandler.GetUser)
			r.Get("/admins", userHandler.ListAdmins)
			r.Post("/admins", userHandler.CreateAdmin)
			r.Put("/admins/{id}", userHandler.UpdateAdmin)

			// Settings
			settingsHandler := admin.NewSettingsHandler(settingsRepo)
			r.Get("/settings", settingsHandler.Get)
			r.Put("/settings", settingsHandler.Upsert)

			// Logs
			r.Get("/logs", userHandler.GetLogs)
		})
	})

	// Bot webhook
	if cfg.Telegram.BotToken != "" {
		tgBot.RegisterRoutes(r)
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
		server.Close()
	}()

	log.Printf("Server starting on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func handleTelegramAuth(authSvc *service.AuthService, userRepo *repository.UserRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			InitData string `json:"initData"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
			return
		}

		// For development, allow test auth
		var user *domain.User
		var err error

		if req.InitData == "" || req.InitData == "test" {
			// Dev mode: auto-create test user
			tgID := int64(123456789)
			username := "test_user"
			user, err = userRepo.Upsert(r.Context(), tgID, &username, nil)
		} else {
			user, err = authSvc.AuthenticateUser(r.Context(), req.InitData)
		}

		if err != nil {
			h.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication failed")
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
