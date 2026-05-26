package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"tma-backend/internal/domain"
)

type OrderRepo struct {
	db *sqlx.DB
}

func NewOrderRepo(db *sqlx.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

type OrderFilter struct {
	Status         *string    `json:"status,omitempty"`
	Statuses       []string   `json:"statuses,omitempty"`
	PaymentMethod  *string    `json:"payment_method,omitempty"`
	DeliveryMethod *string    `json:"delivery_method,omitempty"`
	AdminID        *uuid.UUID `json:"admin_id,omitempty"`
	DateFrom       *time.Time `json:"date_from,omitempty"`
	DateTo         *time.Time `json:"date_to,omitempty"`
	Search         *string    `json:"search,omitempty"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
	Page           int        `json:"page"`
	Limit          int        `json:"limit"`
}

func (r *OrderRepo) List(ctx context.Context, f OrderFilter) ([]domain.Order, int, error) {
	args := []interface{}{}
	where := []string{}
	argIdx := 1

	if f.Status != nil && *f.Status != "" {
		where = append(where, fmt.Sprintf("o.status = $%d", argIdx))
		args = append(args, *f.Status)
		argIdx++
	}

	if len(f.Statuses) > 0 {
		placeholders := make([]string, len(f.Statuses))
		for i, s := range f.Statuses {
			placeholders[i] = fmt.Sprintf("$%d", argIdx+i)
			args = append(args, s)
		}
		where = append(where, fmt.Sprintf("o.status IN (%s)", strings.Join(placeholders, ",")))
		argIdx += len(f.Statuses)
	}

	if f.PaymentMethod != nil && *f.PaymentMethod != "" {
		where = append(where, fmt.Sprintf("o.payment_method = $%d", argIdx))
		args = append(args, *f.PaymentMethod)
		argIdx++
	}
	if f.DeliveryMethod != nil && *f.DeliveryMethod != "" {
		where = append(where, fmt.Sprintf("o.delivery_method = $%d", argIdx))
		args = append(args, *f.DeliveryMethod)
		argIdx++
	}
	if f.AdminID != nil {
		where = append(where, fmt.Sprintf("o.assigned_admin_id = $%d", argIdx))
		args = append(args, *f.AdminID)
		argIdx++
	}
	if f.DateFrom != nil {
		where = append(where, fmt.Sprintf("o.created_at >= $%d", argIdx))
		args = append(args, *f.DateFrom)
		argIdx++
	}
	if f.DateTo != nil {
		where = append(where, fmt.Sprintf("o.created_at <= $%d", argIdx))
		args = append(args, *f.DateTo)
		argIdx++
	}
	if f.UserID != nil {
		where = append(where, fmt.Sprintf("o.user_id = $%d", argIdx))
		args = append(args, *f.UserID)
		argIdx++
	}
	if f.Search != nil && *f.Search != "" {
		where = append(where, fmt.Sprintf("o.id::text ILIKE '%%' || $%d || '%%'", argIdx))
		args = append(args, *f.Search)
		argIdx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM orders o" + whereClause
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	offset := (f.Page - 1) * f.Limit

	query := `SELECT o.* FROM orders o` + whereClause + " ORDER BY o.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, f.Limit, offset)

	var orders []domain.Order
	if err := r.db.SelectContext(ctx, &orders, query, args...); err != nil {
		return nil, 0, err
	}
	if orders == nil {
		orders = []domain.Order{}
	}

	// Populate products
	r.populateProducts(ctx, orders)

	return orders, total, nil
}

func (r *OrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	var o domain.Order
	err := r.db.GetContext(ctx, &o,
		`SELECT o.* FROM orders o WHERE o.id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) GetByIDWithJoins(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	var o domain.Order
	err := r.db.GetContext(ctx, &o,
		`SELECT o.* FROM orders o WHERE o.id = $1`, id)
	if err != nil {
		return nil, err
	}

	// Fetch product
	var p domain.Product
	if err := r.db.GetContext(ctx, &p, `SELECT * FROM products WHERE id = $1`, o.ProductID); err == nil {
		o.Product = &p
		// Fix NULL payment_amount from old orders
		if o.PaymentAmount == nil && p.Price > 0 {
			amount := p.Price * float64(o.Quantity)
			if p.DiscountPercent > 0 {
				amount = amount * (1 - p.DiscountPercent/100)
			}
			o.PaymentAmount = &amount
		}
	}

	// Fetch user
	var u domain.User
	if err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = $1`, o.UserID); err == nil {
		o.User = &u
	}

	// Fetch key value if assigned
	if o.KeyID != nil {
		var k domain.ProductKey
		if err := r.db.GetContext(ctx, &k, `SELECT * FROM product_keys WHERE id = $1`, *o.KeyID); err == nil {
			o.KeyValue = &k.Key
		}
	}

	return &o, nil
}

func (r *OrderRepo) GetByUserID(ctx context.Context, userID uuid.UUID, statuses []string, page, limit int) ([]domain.Order, int, error) {
	args := []interface{}{userID}
	where := " WHERE o.user_id = $1"
	argIdx := 2

	if len(statuses) > 0 {
		placeholders := make([]string, len(statuses))
		for i, s := range statuses {
			placeholders[i] = fmt.Sprintf("$%d", argIdx+i)
			args = append(args, s)
		}
		where += fmt.Sprintf(" AND o.status IN (%s)", strings.Join(placeholders, ","))
		argIdx += len(statuses)
	}

	var total int
	if err := r.db.GetContext(ctx, &total,
		"SELECT COUNT(*) FROM orders o"+where, args...); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query := `SELECT o.* FROM orders o` + where + ` ORDER BY o.created_at DESC`
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	var orders []domain.Order
	if err := r.db.SelectContext(ctx, &orders, query, args...); err != nil {
		return nil, 0, err
	}
	if orders == nil {
		orders = []domain.Order{}
	}

	// Populate products
	r.populateProducts(ctx, orders)

	return orders, total, nil
}

func (r *OrderRepo) populateProducts(ctx context.Context, orders []domain.Order) {
	productIDs := make(map[uuid.UUID]bool)
	for i := range orders {
		if orders[i].Product == nil && orders[i].ProductID != uuid.Nil {
			productIDs[orders[i].ProductID] = true
		}
	}
	if len(productIDs) == 0 {
		return
	}

	ids := make([]uuid.UUID, 0, len(productIDs))
	for id := range productIDs {
		ids = append(ids, id)
	}

	var products []domain.Product
	query, args, _ := sqlx.In(`SELECT id, title, price, discount_percent, image_url, type, platform FROM products WHERE id IN (?)`, ids)
	query = r.db.Rebind(query)
	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		return
	}

	productMap := make(map[uuid.UUID]*domain.Product)
	for i := range products {
		productMap[products[i].ID] = &products[i]
	}

	for i := range orders {
		if p, ok := productMap[orders[i].ProductID]; ok {
			orders[i].Product = p
			// Fix NULL payment_amount from old orders
			if orders[i].PaymentAmount == nil && p.Price > 0 {
				amount := p.Price * float64(orders[i].Quantity)
				if p.DiscountPercent > 0 {
					amount = amount * (1 - p.DiscountPercent/100)
				}
				orders[i].PaymentAmount = &amount
			}
		}
	}
}

func (r *OrderRepo) Create(ctx context.Context, o *domain.Order) error {
	err := r.db.GetContext(ctx, o,
		`INSERT INTO orders (user_id, product_id, variant_id, quantity, delivery_method, status, payment_amount)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING *`, o.UserID, o.ProductID, o.VariantID, o.Quantity, o.DeliveryMethod, domain.OrderStatusNew, o.PaymentAmount)
	return err
}

func (r *OrderRepo) CreateInTx(tx *sqlx.Tx, o *domain.Order) error {
	err := tx.GetContext(context.Background(), o,
		`INSERT INTO orders (user_id, product_id, variant_id, quantity, delivery_method, status, payment_amount)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING *`, o.UserID, o.ProductID, o.VariantID, o.Quantity, o.DeliveryMethod, domain.OrderStatusNew, o.PaymentAmount)
	return err
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE orders SET status=$1, updated_at=NOW() WHERE id=$2", status, id)
	return err
}

func (r *OrderRepo) Update(ctx context.Context, o *domain.Order) error {
	_, err := r.db.NamedExecContext(ctx,
		`UPDATE orders SET 
		 status=:status, payment_method=:payment_method, 
		 payment_amount=:payment_amount, payment_receipt_url=:payment_receipt_url,
		 payment_verified_by=:payment_verified_by, key_id=:key_id,
		 assigned_admin_id=:assigned_admin_id, cancelled_reason=:cancelled_reason,
		 variant_id=:variant_id, quantity=:quantity,
		 updated_at=NOW()
		 WHERE id=:id`, o)
	return err
}

func (r *OrderRepo) AddHistory(ctx context.Context, h *domain.OrderHistory) error {
	_, err := r.db.NamedExecContext(ctx,
		`INSERT INTO order_history (order_id, old_status, new_status, changed_by_id, changed_by_type, comment)
		 VALUES (:order_id, :old_status, :new_status, :changed_by_id, :changed_by_type, :comment)`, h)
	return err
}

func (r *OrderRepo) AddHistoryInTx(tx *sqlx.Tx, h *domain.OrderHistory) error {
	_, err := tx.NamedExecContext(context.Background(),
		`INSERT INTO order_history (order_id, old_status, new_status, changed_by_id, changed_by_type, comment)
		 VALUES (:order_id, :old_status, :new_status, :changed_by_id, :changed_by_type, :comment)`, h)
	return err
}

func (r *OrderRepo) GetHistory(ctx context.Context, orderID uuid.UUID) ([]domain.OrderHistory, error) {
	var h []domain.OrderHistory
	err := r.db.SelectContext(ctx, &h,
		"SELECT * FROM order_history WHERE order_id=$1 ORDER BY created_at ASC", orderID)
	if err != nil {
		return nil, err
	}
	if h == nil {
		h = []domain.OrderHistory{}
	}
	return h, nil
}

func (r *OrderRepo) GetExpired2FA(ctx context.Context, timeout time.Duration) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.SelectContext(ctx, &orders,
		`SELECT o.* FROM orders o
		 WHERE o.status = 'AWAITING_2FA'
		 AND o.updated_at < NOW() - $1::interval`, timeout.String())
	if err != nil {
		return nil, err
	}
	if orders == nil {
		orders = []domain.Order{}
	}
	return orders, nil
}

func (r *OrderRepo) GetExpiredWaitingPayment(ctx context.Context, timeout time.Duration) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.SelectContext(ctx, &orders,
		`SELECT o.* FROM orders o
		 WHERE o.status = 'WAITING_PAYMENT'
		 AND o.created_at < NOW() - $1::interval`, timeout.String())
	if err != nil {
		return nil, err
	}
	if orders == nil {
		orders = []domain.Order{}
	}
	return orders, nil
}

func (r *OrderRepo) GetDashboardStats(ctx context.Context) (map[string]interface{}, error) {
	stats := map[string]interface{}{}

	var ordersToday int
	r.db.GetContext(ctx, &ordersToday,
		"SELECT COUNT(*) FROM orders WHERE created_at >= CURRENT_DATE")
	stats["orders_today"] = ordersToday

	var totalOrders int
	r.db.GetContext(ctx, &totalOrders, "SELECT COUNT(*) FROM orders")
	stats["total_orders"] = totalOrders

	var revenueToday float64
	r.db.GetContext(ctx, &revenueToday,
		`SELECT COALESCE(SUM(COALESCE(o.payment_amount, p.price * o.quantity * (1 - COALESCE(p.discount_percent, 0)/100))), 0) 
		 FROM orders o LEFT JOIN products p ON o.product_id = p.id 
		 WHERE o.status IN ('COMPLETED','KEY_ISSUED','ACTIVATED') AND o.updated_at >= CURRENT_DATE`)
	stats["revenue_today"] = revenueToday

	var revenueWeek float64
	r.db.GetContext(ctx, &revenueWeek,
		`SELECT COALESCE(SUM(COALESCE(o.payment_amount, p.price * o.quantity * (1 - COALESCE(p.discount_percent, 0)/100))), 0) 
		 FROM orders o LEFT JOIN products p ON o.product_id = p.id 
		 WHERE o.status IN ('COMPLETED','KEY_ISSUED','ACTIVATED') AND o.updated_at >= CURRENT_DATE - INTERVAL '7 days'`)
	stats["revenue_week"] = revenueWeek

	var revenueMonth float64
	r.db.GetContext(ctx, &revenueMonth,
		`SELECT COALESCE(SUM(COALESCE(o.payment_amount, p.price * o.quantity * (1 - COALESCE(p.discount_percent, 0)/100))), 0) 
		 FROM orders o LEFT JOIN products p ON o.product_id = p.id 
		 WHERE o.status IN ('COMPLETED','KEY_ISSUED','ACTIVATED') AND o.updated_at >= DATE_TRUNC('month', CURRENT_DATE)`)
	stats["revenue_month"] = revenueMonth

	var totalRevenue float64
	r.db.GetContext(ctx, &totalRevenue,
		`SELECT COALESCE(SUM(COALESCE(o.payment_amount, p.price * o.quantity * (1 - COALESCE(p.discount_percent, 0)/100))), 0) 
		 FROM orders o LEFT JOIN products p ON o.product_id = p.id 
		 WHERE o.status IN ('COMPLETED','KEY_ISSUED','ACTIVATED')`)
	stats["total_revenue"] = totalRevenue

	var activeTasks int
	r.db.GetContext(ctx, &activeTasks,
		"SELECT COUNT(*) FROM orders WHERE status IN ('WAITING_ACTIVATION','AWAITING_CREDENTIALS','CREDENTIALS_RECEIVED','AWAITING_2FA','ACTIVATING')")
	stats["active_tasks"] = activeTasks

	var waitingPayment int
	r.db.GetContext(ctx, &waitingPayment,
		"SELECT COUNT(*) FROM orders WHERE status = 'WAITING_PAYMENT'")
	stats["waiting_payment"] = waitingPayment

	var pendingOrders int
	r.db.GetContext(ctx, &pendingOrders,
		"SELECT COUNT(*) FROM orders WHERE status = 'PAYMENT_VERIFICATION'")
	stats["pending_orders"] = pendingOrders

	var availableKeys int
	r.db.GetContext(ctx, &availableKeys,
		"SELECT COUNT(*) FROM product_keys WHERE status = 'available'")
	stats["available_keys"] = availableKeys

	var soldKeys int
	r.db.GetContext(ctx, &soldKeys,
		"SELECT COUNT(*) FROM product_keys WHERE status = 'sold'")
	stats["sold_keys"] = soldKeys

	var totalUsers int
	r.db.GetContext(ctx, &totalUsers, "SELECT COUNT(*) FROM users")
	stats["total_users"] = totalUsers

	var pendingRevenue float64
	r.db.GetContext(ctx, &pendingRevenue,
		`SELECT COALESCE(SUM(COALESCE(o.payment_amount, p.price * o.quantity * (1 - COALESCE(p.discount_percent, 0)/100))), 0) 
		 FROM orders o LEFT JOIN products p ON o.product_id = p.id 
		 WHERE o.status IN ('WAITING_PAYMENT','PAYMENT_VERIFICATION','PAID')`)
	stats["pending_revenue"] = pendingRevenue

	var avgCheck float64
	r.db.GetContext(ctx, &avgCheck,
		`SELECT COALESCE(AVG(COALESCE(o.payment_amount, p.price * o.quantity * (1 - COALESCE(p.discount_percent, 0)/100))), 0) 
		 FROM orders o LEFT JOIN products p ON o.product_id = p.id 
		 WHERE o.status IN ('COMPLETED','KEY_ISSUED','ACTIVATED')`)
	stats["avg_check"] = avgCheck

	var conversionRate float64
	var totalOrdersForConversion int
	r.db.GetContext(ctx, &totalOrdersForConversion, "SELECT COUNT(*) FROM orders")
	var completedOrders int
	r.db.GetContext(ctx, &completedOrders, "SELECT COUNT(*) FROM orders WHERE status IN ('COMPLETED','KEY_ISSUED','ACTIVATED')")
	if totalOrdersForConversion > 0 {
		conversionRate = float64(completedOrders) / float64(totalOrdersForConversion) * 100
	}
	stats["conversion_rate"] = conversionRate

	var recentOrders []map[string]interface{}
	r.db.SelectContext(ctx, &recentOrders,
		`SELECT o.id, o.status, COALESCE(o.payment_amount, p.price * o.quantity * (1 - COALESCE(p.discount_percent, 0)/100)) as payment_amount, p.title as product_title, o.created_at
		 FROM orders o LEFT JOIN products p ON o.product_id = p.id
		 ORDER BY o.created_at DESC LIMIT 10`)
	stats["recent_orders"] = recentOrders

	return stats, nil
}
