package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type OrderHandler struct {
	svc          *service.OrderService
	orderRepo    *repository.OrderRepo
	keyRepo      *repository.KeyRepo
	adminRepo    *repository.AdminRepo
	encSvc       *service.EncryptionService
	templateRepo *repository.TemplateRepo
}

func NewOrderHandler(svc *service.OrderService, orderRepo *repository.OrderRepo, keyRepo *repository.KeyRepo, adminRepo *repository.AdminRepo, encSvc *service.EncryptionService, templateRepo *repository.TemplateRepo) *OrderHandler {
	return &OrderHandler{svc: svc, orderRepo: orderRepo, keyRepo: keyRepo, adminRepo: adminRepo, encSvc: encSvc, templateRepo: templateRepo}
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repository.OrderFilter{
		Page:  1,
		Limit: 20,
	}

	if v := q.Get("status"); v != "" {
		statuses := strings.Split(v, ",")
		for i, s := range statuses {
			statuses[i] = strings.TrimSpace(s)
		}
		if len(statuses) == 1 {
			f.Status = &statuses[0]
		} else {
			f.Statuses = statuses
		}
	}
	if v := q.Get("payment_method"); v != "" {
		f.PaymentMethod = &v
	}
	if v := q.Get("delivery_method"); v != "" {
		f.DeliveryMethod = &v
	}
	if v := q.Get("admin_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.AdminID = &id
		}
	}
	if v := q.Get("date_from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.DateFrom = &t
		}
	}
	if v := q.Get("date_to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.DateTo = &t
		}
	}
	if v := q.Get("search"); v != "" {
		f.Search = &v
	}
	if v := q.Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			f.Page = p
		}
	}
	if v := q.Get("limit"); v != "" {
		if l, err := strconv.Atoi(v); err == nil {
			f.Limit = l
		}
	}

	orders, total, err := h.svc.List(r.Context(), f)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"data": orders,
		"meta": map[string]interface{}{
			"page":  f.Page,
			"limit": f.Limit,
			"total": total,
		},
	})
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	order, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Order not found")
		return
	}

	history, _ := h.svc.GetHistory(r.Context(), id)

	result := map[string]interface{}{
		"order":   order,
		"history": history,
	}

	// Fetch key if exists
	if order.KeyID != nil {
		key, err := h.keyRepo.GetByID(r.Context(), *order.KeyID)
		if err == nil {
			result["key"] = map[string]interface{}{
				"id":         key.ID,
				"product_id": key.ProductID,
				"key":        key.Key,
				"status":     key.Status,
				"order_id":   key.OrderID,
				"created_at": key.CreatedAt,
			}
		}
	}

	// Fetch assigned admin info
	if order.AssignedAdminID != nil {
		admin, err := h.adminRepo.GetByID(r.Context(), *order.AssignedAdminID)
		if err == nil {
			result["assigned_admin"] = admin
		}
	}

	// Fetch user account credentials if activation order
	if order.DeliveryMethod == domain.DeliveryMethodActivation {
		account, err := h.svc.GetAccount(r.Context(), id)
		if err == nil && account != nil {
			// Decrypt credentials for admin view
			loginBytes, _ := h.encSvc.Decrypt(account.Login)
			passwordBytes, _ := h.encSvc.Decrypt(account.Password)
			twoFactorCode := ""
			if account.TwoFactorCode != nil {
				codeBytes, err := h.encSvc.Decrypt(*account.TwoFactorCode)
				if err == nil {
					twoFactorCode = string(codeBytes)
				} else {
					// Fallback: if decryption fails, it might be stored as plain text
					twoFactorCode = *account.TwoFactorCode
				}
			}

			result["account"] = map[string]interface{}{
				"id":             account.ID,
				"user_id":        account.UserID,
				"order_id":       account.OrderID,
				"platform":       account.Platform,
				"login":          string(loginBytes),
				"password":       string(passwordBytes),
				"two_factor_code": twoFactorCode,
				"notes":          account.Notes,
				"created_at":     account.CreatedAt,
			}
		}
	}

	handler.RespondJSON(w, http.StatusOK, result)
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	var req struct {
		Status  string `json:"status"`
		Comment string `json:"comment"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	adminID := handler.GetAdminID(r.Context())

	switch domain.OrderStatus(req.Status) {
	case domain.OrderStatusPaymentVerification:
		err = h.svc.SimpleStatusChange(r.Context(), id, adminID, domain.OrderStatusPaymentVerification, req.Comment)
	case domain.OrderStatusPaid:
		err = h.svc.ConfirmPayment(r.Context(), id, adminID, "")
	case domain.OrderStatusWaitingActivation:
		err = h.svc.SimpleStatusChange(r.Context(), id, adminID, domain.OrderStatusWaitingActivation, req.Comment)
	case domain.OrderStatusKeyIssued:
		err = h.svc.IssueKey(r.Context(), id, adminID)
	case domain.OrderStatusAwaitingCredentials:
		ord, _ := h.orderRepo.GetByID(r.Context(), id)
		if ord != nil && ord.Status == domain.OrderStatusWaitingActivation {
			err = h.svc.AssignActivation(r.Context(), id, adminID)
		} else {
			err = h.svc.SimpleStatusChange(r.Context(), id, adminID, domain.OrderStatusAwaitingCredentials, req.Comment)
		}
	case domain.OrderStatusCredentialsReceived:
		err = h.svc.SimpleStatusChange(r.Context(), id, adminID, domain.OrderStatusCredentialsReceived, req.Comment)
	case domain.OrderStatusAwaiting2FA:
		ord, _ := h.orderRepo.GetByID(r.Context(), id)
		if ord != nil && ord.Status == domain.OrderStatusCredentialsReceived {
			err = h.svc.Request2FA(r.Context(), id, adminID)
		} else {
			err = h.svc.SimpleStatusChange(r.Context(), id, adminID, domain.OrderStatusAwaiting2FA, req.Comment)
		}
	case domain.OrderStatusActivating:
		err = h.svc.SimpleStatusChange(r.Context(), id, adminID, domain.OrderStatusActivating, req.Comment)
	case domain.OrderStatusActivated:
		err = h.svc.CompleteActivation(r.Context(), id, adminID)
	case domain.OrderStatusCompleted:
		err = h.svc.SimpleStatusChange(r.Context(), id, adminID, domain.OrderStatusCompleted, req.Comment)
	case domain.OrderStatusCredentialsInvalid:
		err = h.svc.SimpleStatusChange(r.Context(), id, adminID, domain.OrderStatusCredentialsInvalid, req.Comment)
	case domain.OrderStatusInvalid2FA:
		err = h.svc.SimpleStatusChange(r.Context(), id, adminID, domain.OrderStatusInvalid2FA, req.Comment)
	case domain.OrderStatusCancelled:
		err = h.svc.CancelOrder(r.Context(), id, adminID, req.Comment)
	case domain.OrderStatusRefundRequested:
		err = h.svc.RequestRefund(r.Context(), id, adminID)
	case domain.OrderStatusRefunded:
		err = h.svc.ProcessRefund(r.Context(), id, adminID)
	default:
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid status transition")
		return
	}

	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *OrderHandler) DecryptCredentials(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	adminID := handler.GetAdminID(r.Context())

	creds, err := h.svc.DecryptCredentials(r.Context(), id, adminID)
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "DECRYPT_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, creds)
}

func (h *OrderHandler) GetAvailableKeys(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	order, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Order not found")
		return
	}

	keys, _, err := h.keyRepo.GetByProductID(r.Context(), order.ProductID, strPtr("available"), 1, 100)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get keys")
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"keys": keys,
		"count": len(keys),
	})
}

func (h *OrderHandler) AssignSpecificKey(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	var req struct {
		KeyID string `json:"key_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	keyID, err := uuid.Parse(req.KeyID)
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid key ID")
		return
	}

	adminID := handler.GetAdminID(r.Context())

	order, err := h.svc.GetByID(r.Context(), orderID)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Order not found")
		return
	}

	if order.Status != domain.OrderStatusPaid {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", "Order must be in PAID status")
		return
	}

	key, err := h.keyRepo.GetByID(r.Context(), keyID)
	if err != nil {
		handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Key not found")
		return
	}

	if key.Status != domain.KeyStatusAvailable {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", "Key is not available")
		return
	}

	if key.ProductID != order.ProductID {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", "Key does not match product")
		return
	}

	err = h.keyRepo.AssignSpecificKey(r.Context(), keyID, orderID)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "ORDER_ERROR", "Failed to assign key")
		return
	}

	order.KeyID = &keyID
	if err := h.orderRepo.Update(r.Context(), order); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "ORDER_ERROR", "Failed to update order")
		return
	}

	h.svc.SendChatMessage(r.Context(), orderID, adminID, "admin", "Ваш ключ активации: "+key.Key)

	if err := h.changeStatus(r.Context(), order, domain.OrderStatusKeyIssued, adminID, domain.ChangedByAdmin, "Key issued"); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", err.Error())
		return
	}

	if err := h.changeStatus(r.Context(), order, domain.OrderStatusCompleted, uuid.Nil, domain.ChangedBySystem, "Order completed"); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "key_assigned"})
}

func (h *OrderHandler) changeStatus(ctx context.Context, order *domain.Order, newStatus domain.OrderStatus, changedByID uuid.UUID, changedByType domain.ChangedByType, comment string) error {
	oldStatus := order.Status
	order.Status = newStatus

	if err := h.svc.UpdateStatus(ctx, order.ID, newStatus); err != nil {
		return err
	}

	history := &domain.OrderHistory{
		OrderID:       order.ID,
		OldStatus:     &oldStatus,
		NewStatus:     newStatus,
		ChangedByID:   func() *uuid.UUID { if changedByID == uuid.Nil { return nil }; return &changedByID }(),
		ChangedByType: changedByType,
		Comment:       &comment,
	}
	if err := h.svc.AddHistory(ctx, history); err != nil {
		return err
	}

	return nil
}

func strPtr(s string) *string { return &s }

func (h *OrderHandler) SendChatMessage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	adminID := handler.GetAdminID(r.Context())

	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	if err := h.svc.SendChatMessage(r.Context(), id, adminID, "admin", req.Message); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (h *OrderHandler) SendTemplateMessage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	adminID := handler.GetAdminID(r.Context())

	var req struct {
		TemplateID uuid.UUID `json:"template_id"`
		Message    string    `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	message := req.Message
	if req.TemplateID != uuid.Nil {
		t, err := h.templateRepo.GetByID(r.Context(), req.TemplateID)
		if err != nil {
			handler.RespondError(w, http.StatusNotFound, "NOT_FOUND", "Template not found")
			return
		}
		message = t.Message
	}

	if message == "" {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Message is required")
		return
	}

	if err := h.svc.SendChatMessage(r.Context(), id, adminID, "admin", message); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "sent"})
}

func (h *OrderHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	messages, err := h.svc.GetChatMessages(r.Context(), id)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, messages)
}

func (h *OrderHandler) BulkUpdateStatus(w http.ResponseWriter, r *http.Request) {
	adminID := handler.GetAdminID(r.Context())

	var req struct {
		IDs     []string `json:"ids"`
		Status  string   `json:"status"`
		Comment string   `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	var ids []uuid.UUID
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "No valid order IDs")
		return
	}

	success, err := h.svc.BulkUpdateStatus(r.Context(), ids, domain.OrderStatus(req.Status), adminID, req.Comment)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"updated": success,
		"total":   len(ids),
	})
}
