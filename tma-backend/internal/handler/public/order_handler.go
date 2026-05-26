package public

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"tma-backend/internal/domain"
	"tma-backend/internal/handler"
	"tma-backend/internal/service"
)

type OrderHandler struct {
	svc       *service.OrderService
	uploadDir string
}

func NewOrderHandler(svc *service.OrderService, uploadDir string) *OrderHandler {
	return &OrderHandler{svc: svc, uploadDir: uploadDir}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserID(r.Context())

	var req struct {
		ProductID      string  `json:"product_id"`
		DeliveryMethod string  `json:"delivery_method"`
		VariantID      *string `json:"variant_id"`
		Quantity       *int    `json:"quantity"`
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

	qty := 1
	if req.Quantity != nil && *req.Quantity > 0 {
		qty = *req.Quantity
	}

	order, err := h.svc.CreateOrder(r.Context(), userID, productID, dm, req.VariantID, qty)
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserID(r.Context())
	q := r.URL.Query()

	var statuses []string
	if v := q.Get("status"); v != "" {
		statuses = strings.Split(v, ",")
		for i, s := range statuses {
			statuses[i] = strings.TrimSpace(s)
		}
	}

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 {
		limit = 20
	}

	orders, total, err := h.svc.GetUserOrders(r.Context(), userID, statuses, page, limit)
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

func (h *OrderHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	userID := handler.GetUserID(r.Context())

	var req struct {
		Items []struct {
			ProductID      string `json:"product_id"`
			DeliveryMethod string `json:"delivery_method"`
			VariantID      *string `json:"variant_id"`
			Quantity       int    `json:"quantity"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	if len(req.Items) == 0 {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "No items provided")
		return
	}

	var orders []*domain.Order
	var totalAmount float64

	for _, item := range req.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid product ID")
			return
		}

		dm := domain.DeliveryMethod(item.DeliveryMethod)
		if dm != domain.DeliveryMethodKey && dm != domain.DeliveryMethodActivation {
			handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid delivery method")
			return
		}

		qty := item.Quantity
		if qty < 1 {
			qty = 1
		}

		order, err := h.svc.CreateOrder(r.Context(), userID, productID, dm, item.VariantID, qty)
		if err != nil {
			handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", err.Error())
			return
		}

		if order.PaymentAmount != nil {
			totalAmount += *order.PaymentAmount
		}
		orders = append(orders, order)
	}

	handler.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"orders":       orders,
		"total_amount": totalAmount,
		"order_ids":    func() []string {
			ids := make([]string, len(orders))
			for i, o := range orders {
				ids[i] = o.ID.String()
			}
			return ids
		}(),
	})
}

func (h *OrderHandler) ConfirmBatchPayment(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(5 << 20)

	var orderIDs []string
	if ids := r.FormValue("order_ids"); ids != "" {
		if err := json.Unmarshal([]byte(ids), &orderIDs); err != nil {
			handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order_ids format")
			return
		}
	}

	if len(orderIDs) == 0 {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "No order IDs provided")
		return
	}

	paymentMethod := r.FormValue("payment_method")
	if paymentMethod == "" {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Payment method is required")
		return
	}

	file, header, err := r.FormFile("receipt")
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Receipt file is required")
		return
	}
	defer file.Close()

	if header.Size > 5<<20 {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "File too large (max 5MB)")
		return
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mimeType := http.DetectContentType(buf[:n])
	allowedTypes := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true,
		"application/pdf": true,
	}
	if !allowedTypes[mimeType] {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid file type (images and PDF only)")
		return
	}

	file.Seek(0, io.SeekStart)

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	uploadPath := filepath.Join(h.uploadDir, filename)

	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save receipt")
		return
	}

	dst, err := os.Create(uploadPath)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save receipt")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save receipt")
		return
	}

	receiptURL := "/uploads/" + filename

	var processed []string
	var failed []map[string]interface{}

	for _, idStr := range orderIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			failed = append(failed, map[string]interface{}{"order_id": idStr, "error": "Invalid ID"})
			continue
		}

		if err := h.svc.UploadReceipt(r.Context(), id, paymentMethod, receiptURL); err != nil {
			failed = append(failed, map[string]interface{}{"order_id": idStr, "error": err.Error()})
		} else {
			processed = append(processed, idStr)
		}
	}

	handler.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"processed": processed,
		"failed":    failed,
		"receipt":   receiptURL,
	})
}

func (h *OrderHandler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	// Limit upload to 5MB
	r.ParseMultipartForm(5 << 20)
	paymentMethod := r.FormValue("payment_method")

	file, header, err := r.FormFile("receipt")
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Receipt file is required")
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > 5<<20 {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "File too large (max 5MB)")
		return
	}

	// Validate MIME type
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	mimeType := http.DetectContentType(buf[:n])
	allowedTypes := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/gif": true, "image/webp": true,
		"application/pdf": true,
	}
	if !allowedTypes[mimeType] {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid file type (images and PDF only)")
		return
	}

	// Reset file reader
	file.Seek(0, io.SeekStart)

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	uploadPath := filepath.Join(h.uploadDir, filename)

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save receipt")
		return
	}

	// Save file to disk
	dst, err := os.Create(uploadPath)
	if err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save receipt")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		handler.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save receipt")
		return
	}

	receiptURL := "/uploads/" + filename

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

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.CancelOrderByUser(r.Context(), id, req.Reason); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "CANCELLED"})
}

func (h *OrderHandler) RequestRefund(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.svc.RequestRefundByUser(r.Context(), id, req.Reason); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "ORDER_ERROR", err.Error())
		return
	}

	handler.RespondJSON(w, http.StatusOK, map[string]string{"status": "REFUND_REQUESTED"})
}

func (h *OrderHandler) SendChatMessage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid order ID")
		return
	}

	userID := handler.GetUserID(r.Context())

	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.RespondError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON")
		return
	}

	if err := h.svc.SendChatMessage(r.Context(), id, userID, "user", req.Message); err != nil {
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
