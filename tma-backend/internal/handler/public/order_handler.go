package public

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserID(r.Context())

	var req struct {
		ProductID      string `json:"product_id"`
		DeliveryMethod string `json:"delivery_method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid product ID")
		return
	}

	dm := domain.DeliveryMethod(req.DeliveryMethod)
	if dm != domain.DeliveryMethodKey && dm != domain.DeliveryMethodActivation {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid delivery method")
		return
	}

	order, err := h.svc.CreateOrder(r.Context(), userID, productID, dm)
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserID(r.Context())
	q := r.URL.Query()

	var status *string
	if v := q.Get("status"); v != "" {
		status = &v
	}

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 {
		limit = 20
	}

	orders, total, err := h.svc.GetUserOrders(r.Context(), userID, status, page, limit)
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

	history, err := h.svc.GetHistory(r.Context(), id)
	if err == nil {
		type orderWithHistory struct {
			*domain.Order
			History []domain.OrderHistory `json:"history"`
		}
		handler.RespondJSON(w, http.StatusOK, orderWithHistory{Order: order, History: history})
		return
	}

	handler.RespondJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	// Handle multipart form for receipt
	r.ParseMultipartForm(10 << 20)
	paymentMethod := r.FormValue("payment_method")

	file, header, err := r.FormFile("receipt")
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Receipt file is required")
		return
	}
	defer file.Close()

	// In a real app, upload to S3/minio
	receiptURL := "/uploads/" + header.Filename
	data, _ := io.ReadAll(file)
	// Save file locally - simplified
	_ = data

	if err := h.svc.UploadReceipt(r.Context(), id, paymentMethod, receiptURL); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "PAYMENT_VERIFICATION"})
}

func (h *OrderHandler) SendCredentials(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	userID := handler.GetUserID(r.Context())

	var req struct {
		Platform string `json:"platform"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	if err := h.svc.ReceiveCredentials(r.Context(), id, userID,
		domain.Platform(req.Platform), req.Login, req.Password); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "CREDENTIALS_RECEIVED"})
}

func (h *OrderHandler) Send2FACode(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	if err := h.svc.Receive2FA(r.Context(), id, req.Code); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "ACTIVATING"})
}
