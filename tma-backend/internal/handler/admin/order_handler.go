package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/repository"
	"tma-backend/internal/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repository.OrderFilter{
		Page:  1,
		Limit: 20,
	}

	if v := q.Get("status"); v != "" {
		f.Status = &v
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
	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"order":   order,
		"history": history,
	})
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
	case domain.OrderStatusPaid:
		err = h.svc.ConfirmPayment(r.Context(), id, adminID, "")
	case domain.OrderStatusKeyIssued:
		err = h.svc.IssueKey(r.Context(), id, adminID)
	case domain.OrderStatusAwaitingCredentials:
		err = h.svc.AssignActivation(r.Context(), id, adminID)
	case domain.OrderStatusAwaiting2FA:
		err = h.svc.Request2FA(r.Context(), id, adminID)
	case domain.OrderStatusActivated:
		err = h.svc.CompleteActivation(r.Context(), id, adminID)
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
