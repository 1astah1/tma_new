package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
)

type AdminUserHandler struct {
	userRepo    *repository.UserRepo
	adminRepo   *repository.AdminRepo
	accountRepo *repository.AccountRepo
}

func NewAdminUserHandler(userRepo *repository.UserRepo, adminRepo *repository.AdminRepo, accountRepo *repository.AccountRepo) *AdminUserHandler {
	return &AdminUserHandler{userRepo: userRepo, adminRepo: adminRepo, accountRepo: accountRepo}
}

func (h *AdminUserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := q.Get("search")
	page, _ := parseInt(q.Get("page"), 1)
	limit, _ := parseInt(q.Get("limit"), 20)

	users, total, err := h.userRepo.List(r.Context(), search, page, limit)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": users,
		"meta": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *AdminUserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid user ID")
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	accounts, _ := h.accountRepo.GetByUserID(r.Context(), id)

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"user":     user,
		"accounts": accounts,
	})
}

func (h *AdminUserHandler) ListAdmins(w http.ResponseWriter, r *http.Request) {
	admins, err := h.adminRepo.List(r.Context())
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{"data": admins})
}

func (h *AdminUserHandler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TelegramID int64    `json:"telegram_id"`
		Username   string   `json:"username"`
		Password   string   `json:"password"`
		Roles      []string `json:"roles"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	admin := &domain.Admin{
		TelegramID: req.TelegramID,
		Username:   req.Username,
		Roles:      req.Roles,
		IsActive:   true,
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to hash password")
		return
	}
	hashStr := string(hashed)
	admin.PasswordHash = &hashStr

	if err := h.adminRepo.Create(r.Context(), admin); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusCreated, admin)
}

func (h *AdminUserHandler) UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid admin ID")
		return
	}

	admin, err := h.adminRepo.GetByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Admin not found")
		return
	}

	var req struct {
		Username *string   `json:"username"`
		Roles    *[]string `json:"roles"`
		IsActive *bool     `json:"is_active"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Username != nil {
		admin.Username = *req.Username
	}
	if req.Roles != nil {
		admin.Roles = *req.Roles
	}
	if req.IsActive != nil {
		admin.IsActive = *req.IsActive
	}

	if err := h.adminRepo.Update(r.Context(), admin); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, admin)
}

func (h *AdminUserHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repository.AuditFilter{}
	f.Page, _ = parseInt(q.Get("page"), 1)
	f.Limit, _ = parseInt(q.Get("limit"), 50)

	logs, total, err := h.adminRepo.GetLogs(r.Context(), f)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": logs,
		"meta": map[string]interface{}{
			"page":  f.Page,
			"limit": f.Limit,
			"total": total,
		},
	})
}

func (h *AdminUserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid user ID")
		return
	}

	var req struct {
		IsBanned    *bool   `json:"is_banned"`
		AdminNotes  *string `json:"admin_notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	if req.IsBanned != nil {
		if err := h.userRepo.UpdateBan(r.Context(), id, *req.IsBanned); err != nil {
			handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update ban status")
			return
		}
	}

	if req.AdminNotes != nil {
		if err := h.userRepo.UpdateAdminNotes(r.Context(), id, *req.AdminNotes); err != nil {
			handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update notes")
			return
		}
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdminUserHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	_, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid user ID")
		return
	}

	q := r.URL.Query()
	page, _ := parseInt(q.Get("page"), 1)
	limit, _ := parseInt(q.Get("limit"), 20)

	orders, total, err := h.userRepo.List(r.Context(), "", page, limit)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": orders,
		"meta": map[string]interface{}{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *AdminUserHandler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid user ID")
		return
	}

	totalOrders, _ := h.userRepo.CountOrders(r.Context(), id)

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"total_orders": totalOrders,
	})
}

func parseInt(s string, fallback int) (int, error) {
	if s == "" {
		return fallback, nil
	}
	return strconv.Atoi(s)
}
